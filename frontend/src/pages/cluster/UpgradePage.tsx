import { Alert, Button, Card, Descriptions, Space, Switch } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { getResource, mutateResource, refreshResource } from '../../api/resource'
import type { ResourceDTO } from '../../api/types'
import type { ApiRecord } from '../../api/client'
import { ResourceMetaBar } from '../../components/ResourceMetaBar'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useClusterContext } from '../../state/ClusterContext'
import { DraggableModal } from '../../components/DraggableModal'
import { UpgradeCheck } from './UpgradeCheck'
import { UpgradeDaemons } from './UpgradeDaemons'
import { RuntimeLogsPanel } from '../monitoring/RuntimeLogsPage'
import { message } from '../../utils/appMessage'

type UpgradeControl = 'pause' | 'resume' | 'stop'
const controlLabels = { pause: '暂停升级', resume: '恢复升级', stop: '停止升级' }

export function watchUpgradeStatus(clusterId: number, auto: boolean, onResult: (item: ResourceDTO) => void, onError: (error: unknown) => void) {
  const abort = new AbortController()
  let timer: ReturnType<typeof setTimeout> | undefined
  async function read() {
    try {
      const { item } = await getResource('/upgrade', clusterId, {}, { signal: abort.signal, suppressErrorNotification: true })
      if (!abort.signal.aborted) onResult(item)
    } catch (error) { if (!abort.signal.aborted) onError(error) }
    finally { if (auto && !abort.signal.aborted) timer = setTimeout(read, 10000) }
  }
  void read()
  return () => { abort.abort(); clearTimeout(timer) }
}

export function upgradeControlAllowed(data: ApiRecord, stale: boolean, action: UpgradeControl) {
  if (stale || data.in_progress !== true) return false
  if (action === 'stop') return true
  return action === 'pause' ? data.is_paused === false : action === 'resume' && data.is_paused === true
}

export function upgradeStatusFields(data: ApiRecord) {
  const value = (input: unknown): string => input == null ? '未提供' : typeof input === 'string' ? (input || '无') : Array.isArray(input) && input.every((entry) => typeof entry === 'string') ? (input.join('、') || '无') : '未知'
  return [
    ['升级状态', data.in_progress === false ? '未在升级' : data.in_progress === true ? (data.is_paused === true ? '已暂停' : data.is_paused === false ? '进行中' : '进行中（暂停状态未知）') : '未知'],
    ['目标镜像', value(data.target_image)], ['升级范围', value(data.which)],
    ['进度', value(data.progress)], ['已完成服务', value(data.services_complete)],
    ['状态消息', value(data.message)]
  ]
}

export function UpgradePage() {
  const { selectedClusterId } = useClusterContext()
  return <UpgradeContent key={selectedClusterId ?? 'none'} selectedClusterId={selectedClusterId} />
}

function UpgradeContent({ selectedClusterId }: { selectedClusterId?: number }) {
  const active = useRef(true)
  useEffect(() => { active.current = true; return () => { active.current = false } }, [])
  const [revision, setRevision] = useState(0)
  const [record, setRecord] = useState<ResourceDTO | null>(null)
  const [loading, setLoading] = useState(false)
  const [auto, setAuto] = useState(true)
  const [error, setError] = useState('')
  const [needsCollection, setNeedsCollection] = useState(false)
  const [pending, setPending] = useState<{ action: UpgradeControl; clusterId: number } | null>(null)
  const operation = useMutationOperation()
  useEffect(() => {
    setRecord(null); setError(''); setLoading(false)
    if (!selectedClusterId) return
    setLoading(true)
    return watchUpgradeStatus(selectedClusterId, auto,
      (item) => { setRecord(item); setError(''); setLoading(false) },
      (err) => { setRecord((previous) => previous ? { ...previous, stale: true } : null); setError(err instanceof Error ? err.message : '升级状态读取失败'); setLoading(false) })
  }, [selectedClusterId, revision, auto])
  useEffect(() => { setPending(null) }, [selectedClusterId])
  async function control() {
    if (!active.current || needsCollection || !pending || pending.clusterId !== selectedClusterId || !record || operation.loading || !upgradeControlAllowed(record.data, record.stale, pending.action)) return
    const { action, clusterId } = pending
    try {
      await operation.run(async () => {
        await mutateResource('/upgrade/action', 'POST', { cluster_id: clusterId, action }, { ifMatch: record.resource_version })
        await refreshResource({ clusterId, kinds: ['upgrade'] })
      }, false)
      if (!active.current) return
      message.success(`${controlLabels[action]}状态已核验`)
      setPending(null)
      setRevision((value) => value + 1)
    } catch (err) {
      if (!active.current) return
      setNeedsCollection(true)
      setPending(null)
      setError(err instanceof Error && err.message ? err.message : '升级控制结果未知')
    }
  }
  async function collect() {
    if (!active.current || !selectedClusterId) return
    try {
      await operation.run(() => refreshResource({ clusterId: selectedClusterId, kinds: ['upgrade', 'daemon'] }), false)
      if (!active.current) return
      setRecord(null)
      setNeedsCollection(false)
      message.success('升级状态与守护进程采集完成')
      setRevision((value) => value + 1)
    } catch (err) { if (active.current) setError(err instanceof Error ? err.message : '采集失败') }
  }
  return <><Card title="集群升级状态" loading={loading} extra={<Space wrap>
    <Switch aria-label="自动读取升级状态" checked={auto} onChange={setAuto} checkedChildren="自动读取" unCheckedChildren="已暂停读取" />
    {(['pause', 'resume', 'stop'] as const).map((action) => <Button key={action} danger={action === 'stop'} disabled={needsCollection || !selectedClusterId || loading || operation.loading || !record || !upgradeControlAllowed(record.data, record.stale, action)} onClick={() => selectedClusterId && setPending({ action, clusterId: selectedClusterId })}>{controlLabels[action]}</Button>)}
    <Button disabled={!selectedClusterId || loading || operation.loading} onClick={() => setRevision((value) => value + 1)}>重新读取</Button>
    <Button disabled={!selectedClusterId || loading || operation.loading} loading={operation.loading} onClick={() => void collect()}>重新采集</Button>
  </Space>}>
    {!selectedClusterId && <Alert type="info" message="请先选择集群" />}
    {selectedClusterId && <Alert type="info" message="自动读取每 10 秒更新库存中的升级状态，不会触发集群操作。请结合采集时间判断；需要立即采集时点击重新采集。" />}
    {error && <Alert type="error" message={error} />}
    {needsCollection && <Alert type="warning" message="上次升级控制结果未确认，命令可能已经执行。请点击重新采集以核对实际状态；读取旧库存不会解除操作锁定。" />}
    {record && <>
      <ResourceMetaBar observedAt={record.observed_at} stale={record.stale} />
      {record.stale && <Alert type="warning" message="升级状态已过期或读取失败，已禁用升级操作，请重新读取或采集。" />}
      <Descriptions column={1} items={upgradeStatusFields(record.data).map(([label, children]) => ({ key: label, label, children }))} />
    </>}
  </Card>
    {selectedClusterId && <UpgradeCheck key={selectedClusterId} clusterId={selectedClusterId} record={record} disabled={needsCollection || loading || operation.loading} onStarted={() => setRevision((value) => value + 1)} />}
    {selectedClusterId && <UpgradeDaemons key={`daemons-${selectedClusterId}`} clusterId={selectedClusterId} revision={revision} />}
    {selectedClusterId && <RuntimeLogsPanel key={`logs-${selectedClusterId}`} compact />}
    <DraggableModal title={pending ? controlLabels[pending.action] : ''} open={pending !== null} confirmLoading={operation.loading} onCancel={() => { if (!operation.loading) setPending(null) }} onOk={() => void control()} okButtonProps={{ danger: pending?.action === 'stop' }}>
      <Alert type="warning" message={pending?.action === 'stop' ? '停止后不会回滚已升级的守护进程，集群可能保留混合版本。确认停止？' : pending?.action === 'pause' ? '暂停后不再继续升级，已升级的守护进程不会回滚。确认暂停？' : '恢复后将继续升级守护进程，可能影响服务。确认恢复？'} />
    </DraggableModal>
  </>
}

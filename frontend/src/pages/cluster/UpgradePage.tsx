import { Alert, Button, Card, Descriptions, Space } from 'antd'
import { useEffect, useState } from 'react'
import { getResource, mutateResource, refreshResource } from '../../api/resource'
import type { ResourceDTO } from '../../api/types'
import type { ApiRecord } from '../../api/client'
import { ResourceMetaBar } from '../../components/ResourceMetaBar'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useClusterContext } from '../../state/ClusterContext'
import { DraggableModal } from '../../components/DraggableModal'
import { UpgradeCheck } from './UpgradeCheck'

type UpgradeControl = 'pause' | 'resume' | 'stop'
const controlLabels = { pause: '暂停升级', resume: '恢复升级', stop: '停止升级' }

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
  const [revision, setRevision] = useState(0)
  const [record, setRecord] = useState<ResourceDTO | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [pending, setPending] = useState<{ action: UpgradeControl; clusterId: number } | null>(null)
  const operation = useMutationOperation()
  useEffect(() => {
    const abort = new AbortController()
    setRecord(null); setError(''); setLoading(false)
    if (!selectedClusterId) return () => abort.abort()
    setLoading(true)
    void getResource('/upgrade', selectedClusterId, {}, { signal: abort.signal, suppressErrorNotification: true })
      .then(({ item }) => { if (!abort.signal.aborted) setRecord(item) })
      .catch((err) => { if (!abort.signal.aborted) setError(err instanceof Error ? err.message : '升级状态读取失败') })
      .finally(() => { if (!abort.signal.aborted) setLoading(false) })
    return () => abort.abort()
  }, [selectedClusterId, revision])
  useEffect(() => { setPending(null) }, [selectedClusterId])
  async function control() {
    if (!pending || pending.clusterId !== selectedClusterId || !record || operation.loading || !upgradeControlAllowed(record.data, record.stale, pending.action)) return
    const { action, clusterId } = pending
    try {
      await operation.run(async () => {
        await mutateResource('/upgrade/action', 'POST', { cluster_id: clusterId, action }, { ifMatch: record.resource_version })
        await refreshResource({ clusterId, kinds: ['upgrade'] })
      }, `${controlLabels[action]}状态已核验`)
      setPending(null)
      setRevision((value) => value + 1)
    } catch (err) { setError(err instanceof Error ? err.message : '升级控制失败') }
  }
  async function collect() {
    if (!selectedClusterId) return
    try {
      await operation.run(() => refreshResource({ clusterId: selectedClusterId, kinds: ['upgrade'] }), '升级状态采集完成')
      setRevision((value) => value + 1)
    } catch (err) { setError(err instanceof Error ? err.message : '采集失败') }
  }
  return <><Card title="集群升级状态" loading={loading} extra={<Space wrap>
    {(['pause', 'resume', 'stop'] as const).map((action) => <Button key={action} danger={action === 'stop'} disabled={!selectedClusterId || loading || operation.loading || !record || !upgradeControlAllowed(record.data, record.stale, action)} onClick={() => selectedClusterId && setPending({ action, clusterId: selectedClusterId })}>{controlLabels[action]}</Button>)}
    <Button disabled={!selectedClusterId || loading || operation.loading} onClick={() => setRevision((value) => value + 1)}>重新读取</Button>
    <Button disabled={!selectedClusterId || loading || operation.loading} loading={operation.loading} onClick={() => void collect()}>重新采集</Button>
  </Space>}>
    {!selectedClusterId && <Alert type="info" message="请先选择集群" />}
    {error && <Alert type="error" message={error} />}
    {record && <>
      <ResourceMetaBar observedAt={record.observed_at} stale={record.stale} />
      <Descriptions column={1} items={upgradeStatusFields(record.data).map(([label, children]) => ({ key: label, label, children }))} />
    </>}
  </Card>
    {selectedClusterId && <UpgradeCheck key={selectedClusterId} clusterId={selectedClusterId} />}
    <DraggableModal title={pending ? controlLabels[pending.action] : ''} open={pending !== null} confirmLoading={operation.loading} onCancel={() => { if (!operation.loading) setPending(null) }} onOk={() => void control()} okButtonProps={{ danger: pending?.action === 'stop' }}>
      <Alert type="warning" message={pending?.action === 'stop' ? '停止后不会回滚已升级的守护进程，集群可能保留混合版本。确认停止？' : pending?.action === 'pause' ? '暂停后不再继续升级，已升级的守护进程不会回滚。确认暂停？' : '恢复后将继续升级守护进程，可能影响服务。确认恢复？'} />
    </DraggableModal>
  </>
}

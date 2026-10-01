import { Alert, Button, Card, Descriptions, Space } from 'antd'
import { useEffect, useState } from 'react'
import { getResource, refreshResource } from '../../api/resource'
import type { ResourceDTO } from '../../api/types'
import type { ApiRecord } from '../../api/client'
import { ResourceMetaBar } from '../../components/ResourceMetaBar'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useClusterContext } from '../../state/ClusterContext'

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
  async function collect() {
    if (!selectedClusterId) return
    try {
      await operation.run(() => refreshResource({ clusterId: selectedClusterId, kinds: ['upgrade'] }), '升级状态采集完成')
      setRevision((value) => value + 1)
    } catch (err) { setError(err instanceof Error ? err.message : '采集失败') }
  }
  return <Card title="集群升级状态" loading={loading} extra={<Space>
    <Button disabled={!selectedClusterId || loading} onClick={() => setRevision((value) => value + 1)}>重新读取</Button>
    <Button disabled={!selectedClusterId || loading} loading={operation.loading} onClick={() => void collect()}>重新采集</Button>
  </Space>}>
    {!selectedClusterId && <Alert type="info" message="请先选择集群" />}
    {error && <Alert type="error" message={error} />}
    {record && <>
      <ResourceMetaBar observedAt={record.observed_at} stale={record.stale} />
      <Descriptions column={1} items={upgradeStatusFields(record.data).map(([label, children]) => ({ key: label, label, children }))} />
    </>}
  </Card>
}

import { Alert, Button, Card, Space, Spin } from 'antd'
import { useEffect, useState } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'
import { AppTable } from '../../components/AppTable'
import { alertField } from './alertColumns'

export function silencedAlerts(value: unknown, silenceId: string) {
  if (!Array.isArray(value) || !silenceId) throw new Error('告警列表或静默 ID 无效')
  let incomplete = false
  const items: ApiRecord[] = []
  for (const item of value) {
    if (!item || typeof item !== 'object' || Array.isArray(item)) throw new Error('告警列表包含无效记录')
    const ids = item.status?.silencedBy
    if (!Array.isArray(ids) || ids.some((id: unknown) => typeof id !== 'string')) { incomplete = true; continue }
    if (ids.includes(silenceId)) items.push(item)
  }
  return { items, incomplete }
}

export function SilencedAlerts({ clusterId, silenceId }: { clusterId: number; silenceId: unknown }) {
  if (typeof silenceId !== 'string' || !silenceId.trim()) return <Alert type="warning" message="静默 ID 缺失，无法读取关联告警" />
  return <SilencedAlertsContent key={JSON.stringify([clusterId, silenceId])} clusterId={clusterId} silenceId={silenceId} />
}

function SilencedAlertsContent({ clusterId, silenceId }: { clusterId: number; silenceId: string }) {
  const [result, setResult] = useState<ReturnType<typeof silencedAlerts> | null>(null)
  const [error, setError] = useState('')
  const [revision, setRevision] = useState(0)
  useEffect(() => {
    const controller = new AbortController()
    setResult(null); setError('')
    void request<{ items: unknown }>('/alert/alerts', jsonInit('GET', { cluster_id: clusterId }, { signal: controller.signal, suppressErrorNotification: true }))
      .then(data => { const value = silencedAlerts(data.items, silenceId); if (!controller.signal.aborted) setResult(value) })
      .catch((err: unknown) => { if (!controller.signal.aborted) setError(err instanceof Error && err.message ? err.message : '关联告警读取失败') })
    return () => controller.abort()
  }, [clusterId, silenceId, revision])
  return <Card title="当前关联告警" style={{ marginTop: 16 }} extra={<Button onClick={() => setRevision(value => value + 1)}>刷新关联</Button>}>
    <Space direction="vertical" style={{ width: '100%' }}>
      <Alert type="info" message="仅展示所选集群 FSID 的 cluster 标签匹配且当前 silencedBy 包含此静默 ID 的告警；不包含其他集群、缺少 cluster 标签或历史告警，不代表未来匹配范围或静默效果预测。" />
      {error ? <Alert type="error" message={error} /> : !result ? <Spin /> : <>
        {result.incomplete && <Alert type="warning" message="部分告警缺少有效关联字段，以下结果不完整，不能据此认定没有关联告警。" />}
        <AppTable<ApiRecord> size="small" dataSource={result.items} rowKey={(row, index) => JSON.stringify([row.fingerprint, index])} pagination={{ defaultPageSize: 10 }} columns={[
          { title: '名称', render: (_, row) => alertField(row.labels, 'alertname') },
          { title: '严重程度', render: (_, row) => alertField(row.labels, 'severity') },
          { title: '摘要', render: (_, row) => alertField(row.annotations, 'summary') },
          { title: '状态', render: (_, row) => alertField(row.status, 'state') },
          { title: '指纹', dataIndex: 'fingerprint' },
          { title: '开始时间', dataIndex: 'startsAt' }
        ]} />
      </>}
    </Space>
  </Card>
}

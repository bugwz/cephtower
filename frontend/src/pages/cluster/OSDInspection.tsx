import { Alert, Button, Card, Tabs } from 'antd'
import { useEffect, useState } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'
import { RecordDetail } from '../../components/RecordDetail'
import { OSDHistogram } from './OSDHistogram'

export function OSDInspection({ clusterId, osdId, record }: { clusterId: number; osdId: string; record: ApiRecord }) {
  return <Tabs items={[
    { key: 'map', label: 'OSD 状态', children: <RecordDetail record={record} /> },
    { key: 'metadata', label: '元数据', children: <Diagnostic clusterId={clusterId} osdId={osdId} section="metadata" /> },
    { key: 'histogram', label: '性能直方图', children: <Diagnostic clusterId={clusterId} osdId={osdId} section="histogram" /> }
  ]} />
}

function Diagnostic({ clusterId, osdId, section }: { clusterId: number; osdId: string; section: string }) {
  const [data, setData] = useState<ApiRecord | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [revision, setRevision] = useState(0)
  useEffect(() => {
    const abort = new AbortController()
    setLoading(true); setError(''); setData(null)
    void request<ApiRecord>('/osd/inspection', jsonInit('GET', { cluster_id: clusterId, osd_id: osdId, section }, { signal: abort.signal, suppressErrorNotification: true }))
      .then((value) => { if (!abort.signal.aborted) setData(value) })
      .catch((err) => { if (!abort.signal.aborted) setError(err instanceof Error ? err.message : '读取失败') })
      .finally(() => { if (!abort.signal.aborted) setLoading(false) })
    return () => abort.abort()
  }, [clusterId, osdId, section, revision])
  return <Card loading={loading} extra={<Button disabled={loading} onClick={() => setRevision((value) => value + 1)}>重新读取</Button>}>
    {error && <Alert type="error" message={error} />}
    {data && (section === 'histogram' ? <OSDHistogram data={data} /> : <RecordDetail record={data} />)}
  </Card>
}

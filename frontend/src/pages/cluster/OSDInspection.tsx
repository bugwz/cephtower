import { Alert, Button, Card, Descriptions, Tabs } from 'antd'
import { useEffect, useState } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'
import { RecordDetail } from '../../components/RecordDetail'
import { OSDHistogram } from './OSDHistogram'
import { DaemonPerf } from './DaemonPerf'

export function OSDInspection({ clusterId, osdId, record }: { clusterId: number; osdId: string; record: ApiRecord }) {
  return <Tabs items={[
    { key: 'map', label: 'OSD 状态', children: <RecordDetail record={record} /> },
    { key: 'network', label: '网络地址', children: <>
      <Alert type="info" message="原生 OSDMap 地址快照，包含协议、地址和 nonce；仅展示，不探测连通性。null 表示未取得，addrvec 空数组表示本次未发布地址。" />
      <RecordDetail record={{ public_addrs: record.public_addrs, cluster_addrs: record.cluster_addrs, heartbeat_front_addrs: record.heartbeat_front_addrs, heartbeat_back_addrs: record.heartbeat_back_addrs }} />
    </> },
    { key: 'map-history', label: '状态历史 Epoch', children: <>
      <Alert type="info" message="来自 ceph osd dump 的 OSDMap 历史版本号，不是时间戳；不据此推断具体停机时间。" />
      <Descriptions bordered column={2}>{[['last_clean_begin', '上次 clean 区间起点'], ['last_clean_end', '上次 clean 区间终点'], ['up_from', '标记 up 的 Epoch'], ['up_thru', 'Up thru'], ['down_at', 'Down at'], ['lost_at', '标记丢失的 Epoch']].map(([key, label]) => <Descriptions.Item key={key} label={label}>{osdHistoryEpoch(record[key])}</Descriptions.Item>)}</Descriptions>
    </> },
    { key: 'metadata', label: '元数据', children: <Diagnostic clusterId={clusterId} osdId={osdId} section="metadata" /> },
    { key: 'perf', label: '性能计数器', children: <DaemonPerf key={`${clusterId}:osd.${osdId}`} clusterId={clusterId} name={`osd.${osdId}`} /> },
    { key: 'histogram', label: '性能直方图', children: <Diagnostic clusterId={clusterId} osdId={osdId} section="histogram" /> }
  ]} />
}

export function osdHistoryEpoch(value: unknown): string {
  return typeof value === 'number' && Number.isInteger(value) && value >= 0 && value <= 4294967295 ? String(value) : '未采集或格式无效'
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

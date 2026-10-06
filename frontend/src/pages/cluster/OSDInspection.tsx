import { Alert, Button, Card, Descriptions, Tabs } from 'antd'
import { useEffect, useState } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'
import { RecordDetail } from '../../components/RecordDetail'
import { OSDHistogram } from './OSDHistogram'
import { DaemonPerf } from './DaemonPerf'
import { OSDNetwork } from './OSDNetwork'
import { OSDUsage } from './OSDUsage'

export function OSDInspection({ clusterId, osdId, record }: { clusterId: number; osdId: string; record: ApiRecord }) {
  return <Tabs items={[
    { key: 'map', label: 'OSD 状态', children: <RecordDetail record={record} /> },
    { key: 'network', label: '网络地址', children: <OSDNetwork record={record} /> },
    { key: 'usage', label: '容量、PG 与延迟', children: <OSDUsage record={record} /> },
    { key: 'map-history', label: '状态历史 Epoch', children: <>
      <Alert type="info" message="来自 ceph osd dump 的 OSDMap 历史版本号，不是时间戳；不据此推断具体停机时间。" />
      <Descriptions bordered column={2}>{[['last_clean_begin', '上次 clean 区间起点'], ['last_clean_end', '上次 clean 区间终点'], ['up_from', '标记 up 的 Epoch'], ['up_thru', 'Up thru'], ['down_at', 'Down at'], ['lost_at', '标记丢失的 Epoch']].map(([key, label]) => <Descriptions.Item key={key} label={label}>{osdHistoryEpoch(record[key])}</Descriptions.Item>)}</Descriptions>
    </> },
    { key: 'metadata', label: '元数据', children: <Diagnostic clusterId={clusterId} osdId={osdId} section="metadata" /> },
    { key: 'devices', label: '关联设备', children: <Diagnostic clusterId={clusterId} osdId={osdId} section="devices" /> },
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
    {data && (section === 'histogram' ? <OSDHistogram data={data} /> : section === 'devices' ? <OSDDevices data={data} /> : <RecordDetail record={data} />)}
  </Card>
}

export function osdDeviceRecords(data: ApiRecord): ApiRecord[] | null {
  const devices = data.devices
  return Array.isArray(devices) && devices.every(device => device && typeof device === 'object' && !Array.isArray(device) && typeof device.devid === 'string' && device.devid.trim().length > 0) ? devices : null
}

function OSDDevices({ data }: { data: ApiRecord }) {
  const devices = osdDeviceRecords(data)
  if (!devices) return <Alert type="warning" message="关联设备响应无效，不能判断是否存在设备" />
  return <>
    <Alert type="info" message="来自 ceph device ls-by-daemon 的设备关联记录；没有记录不代表磁盘不存在或健康。寿命预测字段仅在原生响应提供时显示。" />
    {devices.length === 0 ? <Alert type="info" message="本次未返回关联设备记录" /> : devices.map((device, index) => <Card key={`${index}-${device.devid}`} title={String(device.devid)}><RecordDetail record={device} /></Card>)}
  </>
}

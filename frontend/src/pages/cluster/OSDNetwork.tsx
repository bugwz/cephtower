import { Alert, Card, Space, Typography } from 'antd'
import type { ApiRecord } from '../../api/client'
import { AppTable } from '../../components/AppTable'

export function osdAddressRows(value: unknown): ApiRecord[] | null {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return null
  const vector = (value as ApiRecord).addrvec
  if (!Array.isArray(vector) || vector.some(row => !row || typeof row !== 'object' || Array.isArray(row))) return null
  return vector.map((row, index) => ({ ...row, index }))
}

export function osdAddressText(value: unknown): string {
  return typeof value === 'string' ? value || '空字符串（原生）' : '未采集或格式无效'
}

export function osdAddressNonce(value: unknown): string {
  return typeof value === 'number' && Number.isInteger(value) && value >= 0 && value <= 4294967295 ? String(value) : '未采集或格式无效'
}

export function OSDNetwork({ record }: { record: ApiRecord }) {
  return <Space direction="vertical" className="page-stack">
    <Alert type="info" message="原生 OSDMap 地址快照；仅展示协议、地址和 nonce，不探测连通性。" />
    {record.stale !== false && <Alert type="warning" message="库存已过期或新鲜度未知，请重新采集 OSD。" />}
    {[
      ['public_addrs', 'Public（客户端网络）'], ['cluster_addrs', 'Cluster（集群网络）'],
      ['heartbeat_front_addrs', 'Heartbeat front'], ['heartbeat_back_addrs', 'Heartbeat back']
    ].map(([field, title]) => {
      const rows = osdAddressRows(record[field])
      return <Card key={field} title={title} size="small">
        {rows === null ? <Alert type="warning" message="地址列表未采集或格式无效，不能认定没有地址。" /> : <AppTable<ApiRecord> dataSource={rows} rowKey="index" pagination={false} locale={{ emptyText: '本次采集未发布地址' }} columns={[
          { title: '协议', dataIndex: 'type', render: osdAddressText },
          { title: '原始地址', dataIndex: 'addr', render: value => <Typography.Text style={{ overflowWrap: 'anywhere' }}>{osdAddressText(value)}</Typography.Text> },
          { title: 'Nonce', dataIndex: 'nonce', render: osdAddressNonce }
        ]} />}
      </Card>
    })}
  </Space>
}

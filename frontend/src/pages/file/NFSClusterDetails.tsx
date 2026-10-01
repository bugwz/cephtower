import { Alert, Card, Descriptions } from 'antd'
import type { ApiRecord } from '../../api/client'
import { AppTable } from '../../components/AppTable'

export function NFSClusterDetails({ row }: { row: ApiRecord }) {
  if (row.info_available !== true) return <Alert type="warning" message="未获取 NFS 集群端点信息" description="集群存在，但原生 cluster info 查询未成功或未返回有效数据。" />
  const backends = Array.isArray(row.backend) ? row.backend as ApiRecord[] : []
  const text = (value: unknown) => value == null || value === '' ? '未提供' : String(value)
  return <Card size="small" title="NFS 集群端点">
    <Descriptions column={2} size="small">
      <Descriptions.Item label="VIP">{text(row.virtual_ip)}</Descriptions.Item>
      <Descriptions.Item label="入口端口">{text(row.port)}</Descriptions.Item>
      <Descriptions.Item label="入口模式">{text(row.ingress_mode)}</Descriptions.Item>
      <Descriptions.Item label="监控端口">{text(row.monitor_port)}</Descriptions.Item>
    </Descriptions>
    <AppTable<ApiRecord> rowKey={(_, index) => String(index)} dataSource={backends} pagination={false} columns={[
      { title: '后端主机', dataIndex: 'hostname', render: text },
      { title: '后端地址', dataIndex: 'ip', render: text },
      { title: '后端端口', dataIndex: 'port', render: text }
    ]} />
  </Card>
}

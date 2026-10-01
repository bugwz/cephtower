import { Alert, Card, Descriptions, Space } from 'antd'
import type { ApiRecord } from '../../api/client'
import { AppTable } from '../../components/AppTable'
import { nfsFSAL } from './nfsExportFields'

function valueText(value: unknown) {
  if (value == null || value === '') return '未知'
  if (Array.isArray(value)) return value.map(String).join('、') || '未设置'
  return String(value)
}

export function NFSExportDetails({ row }: { row: ApiRecord }) {
  const fsal = nfsFSAL(row)
  const clients = Array.isArray(row.clients) && row.clients.every((value) => Boolean(value) && typeof value === 'object' && !Array.isArray(value)) ? row.clients as ApiRecord[] : undefined
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Card size="small" title="导出访问配置">
      <Descriptions column={2} size="small">
        <Descriptions.Item label="存储后端">{valueText(fsal.name)}</Descriptions.Item>
        <Descriptions.Item label="文件系统">{valueText(fsal.fs_name)}</Descriptions.Item>
        <Descriptions.Item label="后端用户">{valueText(fsal.user_id)}</Descriptions.Item>
        <Descriptions.Item label="访问类型">{valueText(row.access_type)}</Descriptions.Item>
        <Descriptions.Item label="身份映射">{valueText(row.squash)}</Descriptions.Item>
        <Descriptions.Item label="安全标签">{row.security_label === true ? '启用' : row.security_label === false ? '禁用' : '未知'}</Descriptions.Item>
        <Descriptions.Item label="NFS 协议">{valueText(row.protocols)}</Descriptions.Item>
        <Descriptions.Item label="传输协议">{valueText(row.transports)}</Descriptions.Item>
        <Descriptions.Item label="安全类型">{valueText(row.sectype)}</Descriptions.Item>
      </Descriptions>
    </Card>
    <Card size="small" title="客户端访问规则">
      {clients === undefined ? <Alert type="info" message="未返回客户端规则" /> : clients.length === 0 ? <Alert type="info" message="未配置客户端级规则，使用导出级配置" /> : <AppTable<ApiRecord> rowKey={(_, index) => String(index)} dataSource={clients} pagination={false} columns={[
        { title: '客户端地址 / 网段', dataIndex: 'addresses', render: valueText },
        { title: '访问类型', dataIndex: 'access_type', render: (value) => value == null || value === '' ? `继承（${valueText(row.access_type)}）` : valueText(value) },
        { title: '身份映射', dataIndex: 'squash', render: (value) => value == null || value === '' ? `继承（${valueText(row.squash)}）` : valueText(value) }
      ]} />}
    </Card>
  </Space>
}

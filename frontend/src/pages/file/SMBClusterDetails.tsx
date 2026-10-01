import { Alert, Card, Descriptions } from 'antd'
import type { ApiRecord } from '../../api/client'

export function SMBClusterDetails({ row }: { row: ApiRecord }) {
  if (row.info_available !== true) return <Alert type="warning" message="未获取 SMB 集群配置" description="不能根据名称推断认证模式或部署配置，请刷新采集后重试。" />
  const value = (data: unknown) => data == null ? '未提供' : typeof data === 'object' ? JSON.stringify(data, null, 2) : String(data)
  return <Card size="small" title="SMB 原生集群配置">
    <Descriptions column={1} size="small">
      {[
        ['认证模式', row.auth_mode], ['期望状态', row.intent], ['集群模式', row.clustering],
        ['域设置 / 加域凭据引用', row.domain_settings], ['用户组资源引用', row.user_group_settings],
        ['自定义 DNS', row.custom_dns], ['公开地址', row.public_addrs], ['部署放置', row.placement]
      ].map(([label, data]) => <Descriptions.Item key={String(label)} label={String(label)}><pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>{value(data)}</pre></Descriptions.Item>)}
    </Descriptions>
  </Card>
}

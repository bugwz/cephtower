import { Alert, Descriptions, Space } from 'antd'
import { isRecord } from '../../api/client'

export function IngressSettings({ value }: { value: unknown }) {
  if (!isRecord(value)) return <Alert type="warning" message="本次未采集 Ingress 配置" />
  const text = (value: unknown) => typeof value === 'string' && value.trim() ? value : '未报告'
  const flag = (value: unknown) => value === true ? '是' : value === false ? '否' : '未报告'
  const port = (value: unknown) => typeof value === 'number' && Number.isInteger(value) && value > 0 && value <= 65535 ? String(value) : '未报告或格式无效'
  const networks = value.virtual_interface_networks
  return <Space direction="vertical" style={{ minWidth: 240 }}>
    <Alert type="info" message="以下是配置快照，不代表 VIP 已绑定、端口可达或 TLS 连接正常；证书、私钥和密码不返回此视图。" />
    <Descriptions size="small" column={1} items={[
      { key: 'backend', label: '后端服务', children: text(value.backend_service) },
      { key: 'vip', label: '配置虚拟 IP', children: text(value.virtual_ip) },
      { key: 'frontend', label: '前端端口', children: port(value.frontend_port) },
      { key: 'monitor', label: '监控端口', children: port(value.monitor_port) },
      { key: 'ssl', label: '启用 TLS', children: flag(value.ssl) },
      { key: 'keepalive', label: '仅 Keepalived', children: flag(value.keepalive_only) },
      { key: 'networks', label: '虚拟接口候选网段', children: Array.isArray(networks) && networks.every(network => typeof network === 'string' && network.trim()) ? networks.length ? networks.join('、') : '未指定候选网段' : '未报告或格式无效' }
    ]} />
  </Space>
}

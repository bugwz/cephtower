import { Alert, Card, Descriptions, Space, Typography } from 'antd'
import { isRecord, type ApiRecord } from '../../api/client'
import { hostNICCount } from './HostPage'

export function HostNativeSummary({ host }: { host: ApiRecord }) {
  const summary = isRecord(host.native_summary) ? host.native_summary : {}
  const fields = [
    ['server', '服务器型号'], ['cpu_summary', 'CPU 核 / 线程'], ['ram', '内存摘要'],
    ['hdd_summary', 'HDD 摘要'], ['ssd_summary', 'SSD 摘要'], ['os', '操作系统']
  ]
  return <Card title="编排器原生主机摘要" className="page-surface-card">
    <Space direction="vertical" style={{ width: '100%' }}>
      <Typography.Text type="secondary">来自 ceph orch host ls --detail 的库存快照。保留原始单位与舍入值，不代表精确容量、当前利用率或硬件健康。</Typography.Text>
      {host.stale !== false && <Alert type="warning" showIcon message="主机库存已过期或新鲜度未知，下方摘要不代表当前状态" />}
      <Descriptions bordered size="small" column={{ xs: 1, sm: 2 }}>
        {fields.map(([key, label]) => <Descriptions.Item key={key} label={label}>{typeof summary[key] === 'string' && summary[key] !== '' ? summary[key] as string : '未报告'}</Descriptions.Item>)}
        <Descriptions.Item label="网卡数量">{hostNICCount(host)}</Descriptions.Item>
      </Descriptions>
    </Space>
  </Card>
}

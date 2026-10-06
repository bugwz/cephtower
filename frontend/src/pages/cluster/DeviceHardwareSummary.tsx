import { Alert, Card, Descriptions, Space, Typography } from 'antd'
import { isRecord, type ApiRecord } from '../../api/client'

export function DeviceHardwareSummary({ device, stale }: { device: ApiRecord; stale?: boolean }) {
  const data = isRecord(device.lsm_data) ? device.lsm_data : {}
  const leds = isRecord(data.ledSupport) ? data.ledSupport : {}
  const fields = [
    ['serialNum', 'LSM 序列号'], ['health', '健康报告'], ['mediaType', '介质类型'],
    ['transport', '传输类型'], ['rpm', '转速（原值）'], ['linkSpeed', '链路速率（原值）']
  ]
  const lightFields = [
    ['IDENTsupport', '定位灯支持'], ['IDENTstatus', '定位灯状态'],
    ['FAILsupport', '故障灯支持'], ['FAILstatus', '故障灯状态']
  ]
  const display = (value: unknown) => typeof value === 'string' && value !== '' ? value : '未报告'
  const errors = Array.isArray(data.errors) && data.errors.every(value => typeof value === 'string') ? data.errors as string[] : null
  return <Card title="设备硬件诊断（LSM）" className="page-surface-card">
    <Space direction="vertical" style={{ width: '100%' }}>
      <Typography.Text type="secondary">来自 ceph orch device ls 库存中的 lsm_data。增强扫描或硬件支持不足时可能无报告；这些快照不是操作后的实时回读，不代表设备可安全擦除。</Typography.Text>
      {(stale !== false || device.stale !== false) && <Alert type="warning" showIcon message="库存已过期或新鲜度未知，诊断不代表当前硬件状态" />}
      <Descriptions bordered size="small" column={{ xs: 1, sm: 2 }}>
        {fields.map(([key, label]) => <Descriptions.Item key={key} label={label}>{display(data[key])}</Descriptions.Item>)}
        {lightFields.map(([key, label]) => <Descriptions.Item key={key} label={label}>{display(leds[key])}</Descriptions.Item>)}
        <Descriptions.Item label="LSM 查询错误" span={2}>
          {errors === null ? '未报告' : errors.length === 0 ? '未报告查询错误（不代表健康）' : <Space direction="vertical">{errors.map((error, index) => <Typography.Text key={index}>{error}</Typography.Text>)}</Space>}
        </Descriptions.Item>
      </Descriptions>
    </Space>
  </Card>
}

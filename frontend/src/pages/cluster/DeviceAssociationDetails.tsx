import { Alert, Descriptions, Space, Typography } from 'antd'
import { isRecord, type ApiRecord } from '../../api/client'

export function deviceWearDisplay(value: unknown): string {
  if (value === undefined || value === null) return '未报告'
  if (typeof value !== 'number' || !Number.isFinite(value) || value < 0 || !Number.isFinite(value * 100)) return '格式无效'
  return `约 ${(value * 100).toFixed(2)}%（原始比例 ${value}）`
}

export function DeviceAssociationDetails({ device }: { device: ApiRecord }) {
  const text = (value: unknown) => typeof value === 'string' && value.trim() !== '' ? value : '未返回或格式无效'
  const locations = Array.isArray(device.location) && device.location.every(isRecord) ? device.location : null
  const daemons = Array.isArray(device.daemons) && device.daemons.every(value => typeof value === 'string' && value.trim() !== '') ? device.daemons as string[] : null
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Descriptions bordered size="small" column={{ xs: 1, sm: 2 }}>
      <Descriptions.Item label="设备 ID">{text(device.devid)}</Descriptions.Item>
      <Descriptions.Item label="关联守护进程">{daemons === null ? '未返回或格式无效' : daemons.length === 0 ? '本次未报告关联进程' : daemons.join('、')}</Descriptions.Item>
      <Descriptions.Item label="寿命预测下界（原始时间）">{text(device.life_expectancy_min)}</Descriptions.Item>
      <Descriptions.Item label="寿命预测上界（原始时间）">{text(device.life_expectancy_max)}</Descriptions.Item>
      <Descriptions.Item label="预测生成时间">{text(device.life_expectancy_stamp)}</Descriptions.Item>
      <Descriptions.Item label="已使用寿命（wear_level）">{deviceWearDisplay(device.wear_level)}</Descriptions.Item>
    </Descriptions>
    <Typography.Text type="secondary">寿命预测是原生报告的预计时间范围，不是保证寿命，也不代表当前健康状态；未返回字段不按零或健康处理。</Typography.Text>
    <Typography.Text type="secondary">磨损比例来自 Ceph devicehealth 保存的 ATA/NVMe 报告，1 表示 100% 已使用寿命，可能超过 100%；不是剩余寿命或当前健康结论。</Typography.Text>
    {locations === null ? <Alert type="warning" message="设备位置未返回或格式无效" /> : locations.length === 0 ? <Alert type="info" message="本次未报告设备位置" /> : locations.map((location, index) => <Descriptions key={index} title={`设备位置 ${index + 1}`} bordered size="small" column={1}>
      <Descriptions.Item label="主机">{text(location.host)}</Descriptions.Item>
      <Descriptions.Item label="设备名称">{text(location.dev)}</Descriptions.Item>
      <Descriptions.Item label="设备路径">{text(location.path)}</Descriptions.Item>
    </Descriptions>)}
  </Space>
}

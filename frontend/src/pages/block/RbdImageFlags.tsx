import { Space, Tag, Typography } from 'antd'

export function rbdImageFlags(details: unknown) {
  const flags = details && typeof details === 'object' && !Array.isArray(details)
    ? (details as Record<string, unknown>).flags : undefined
  if (!Array.isArray(flags) || !flags.every((flag) => typeof flag === 'string' && flag.trim() !== '')) {
    return [{ label: '镜像标志未返回或无效', color: 'default', explanation: '无法确认对象映射和 fast-diff 状态' }]
  }
  if (flags.length === 0) return [{ label: '未报告异常标志', color: 'green', explanation: '仅表示本次采集的 flags 为空，不代表完整健康检查' }]
  return flags.map((flag: string) => {
    if (flag === 'object map invalid') return { label: 'object map invalid', color: 'orange', explanation: '对象映射无效' }
    if (flag === 'fast diff invalid') return { label: 'fast diff invalid', color: 'orange', explanation: 'fast-diff 映射无效，用量统计可能较慢' }
    return { label: flag, color: 'default', explanation: '未识别的原生标志，请核对 Ceph 版本说明' }
  })
}

export function RbdImageFlags({ details }: { details: unknown }) {
  return <Space direction="vertical" size={4}>{rbdImageFlags(details).map((flag, index) =>
    <div key={`${index}:${flag.label}`}><Tag color={flag.color}>{flag.label}</Tag><Typography.Text type="secondary">{flag.explanation}</Typography.Text></div>
  )}</Space>
}

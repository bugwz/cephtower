import { Alert, Descriptions, Space } from 'antd'
import { isRecord } from '../../api/client'

export function servicePlacementFields(value: unknown): Array<{ label: string; value: string }> {
  if (!isRecord(value)) return []
  return Object.entries(value).map(([key, value]) => ({
    label: ({ hosts: '指定主机', count: '实例数量', count_per_host: '每主机实例数量', label: '主机标签', host_pattern: '主机匹配规则' } as Record<string, string>)[key] ?? key,
    value: value === null || value === undefined ? '未报告' : typeof value === 'string' ? value || '空字符串' : JSON.stringify(value)
  }))
}

export function ServicePlacement({ placement, unmanaged }: { placement: unknown; unmanaged: unknown }) {
  const fields = servicePlacementFields(placement)
  return <Space direction="vertical" style={{ minWidth: 220, whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>
    {unmanaged === true && <Alert type="info" message="非托管：编排器不自动执行以下放置配置" />}
    {!isRecord(placement) ? <Alert type="warning" message="放置配置未返回或格式无效" /> : fields.length === 0 ? <span>未指定放置约束，具体默认值由 Ceph 决定</span> : <Descriptions size="small" column={1} items={fields.map((field, index) => ({ key: index, label: field.label, children: field.value }))} />}
    {isRecord(placement) && <details><summary>原始放置配置</summary><pre style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{JSON.stringify(placement, null, 2)}</pre></details>}
  </Space>
}

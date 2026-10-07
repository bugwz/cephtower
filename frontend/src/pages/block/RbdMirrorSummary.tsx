import { Alert, Descriptions, Space, Table, Tag } from 'antd'

const isRecord = (value: unknown): value is Record<string, unknown> => value !== null && typeof value === 'object' && !Array.isArray(value)
export function mirrorStateCount(value: unknown): string {
  if (typeof value === 'number' && Number.isSafeInteger(value) && value >= 0) return String(value)
  if (typeof value === 'string' && value.trim() === value && /^(0|[1-9][0-9]*)$/.test(value)) return value
  return '未返回有效数量'
}
export function mirrorHealth(value: unknown) {
  const label = typeof value === 'string' && value !== '' ? value : '未返回或无效'
  const color = label === 'OK' ? 'green' : label === 'WARNING' ? 'orange' : label === 'ERROR' ? 'red' : 'default'
  return <Tag color={color}>{label}</Tag>
}

export function RbdMirrorSummary({ value, mode }: { value: unknown; mode: unknown }) {
  if (mode === 'disabled') return <Alert type="info" message="池同步已停用，未查询运行健康状态" />
  if (!isRecord(value)) return <Alert type="warning" message="池同步健康汇总不可用或格式无效" />
  const states = isRecord(value.states) ? Object.entries(value.states).map(([state, count]) => ({ state, count: mirrorStateCount(count) })) : undefined
  return <Space direction="vertical" style={{ minWidth: 320, width: '100%' }}>
    <Descriptions size="small" column={1} items={[
      ['health', '池同步健康'], ['daemon_health', '守护进程健康'], ['image_health', '镜像健康']
    ].map(([key, label]) => ({ key, label, children: mirrorHealth(value[key]) }))} />
    <span>来源：rbd mirror pool status 原生汇总。健康与数量为采集快照，不证明当前已同步；状态数量不区分主副角色，也不是远端镜像数。</span>
    {!states ? <Alert type="warning" message="镜像状态数量不可用" /> : !states.length ? <Alert type="info" message="本次汇总未返回状态计数" /> : <Table size="small" rowKey="state" pagination={{ pageSize: 5 }} dataSource={states} columns={[
      { title: '原生镜像状态', dataIndex: 'state' }, { title: '镜像数量', dataIndex: 'count' }
    ]} />}
  </Space>
}

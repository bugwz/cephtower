import { Alert, Space, Table } from 'antd'

export function RbdMirrorLeaderCounts({ value, status }: { value: unknown; status?: unknown }) {
  if (status === 'disabled') return <Alert type="info" message="池同步已禁用，未查询 Leader 计数" />
  if (status !== 'available') return <Alert type="warning" message="Leader 计数未读取或身份未核实，请检查采集状态；不代表镜像数为零" />
  if (!value || typeof value !== 'object' || Array.isArray(value)) return <Alert type="info" message="未取得身份一致的 Leader 计数，不代表镜像数为零" />
  const data = value as Record<string, unknown>
  if (typeof data.instance_id !== 'string' || !data.instance_id || !Array.isArray(data.namespaces) || data.namespaces.some(row => !row || typeof row !== 'object' || typeof row.namespace !== 'string')) return <Alert type="warning" message="Leader 计数格式无效" />
  const count = (value: unknown) => typeof value === 'string' && /^(0|[1-9][0-9]*)$/.test(value) && value.trim() === value ? value : '未返回有效数量'
  return <Space direction="vertical">
    <span>Leader 实例：{data.instance_id}；来源：ceph service status。按命名空间报告，不汇总为池总数，也不代表同步完成。</span>
    <Table size="small" rowKey="namespace" pagination={{ pageSize: 5 }} dataSource={data.namespaces} columns={[
      { title: '命名空间', dataIndex: 'namespace', render: value => value === '' ? '默认命名空间' : value },
      { title: '本地镜像数', dataIndex: 'image_local_count', render: count },
      { title: '远端镜像数', dataIndex: 'image_remote_count', render: count }
    ]} />
  </Space>
}

export function rbdMirrorDaemonRows(value: unknown) {
  if (!Array.isArray(value) || value.some(row => !row || typeof row !== 'object' || Array.isArray(row))) return undefined
  return value.map((row, index) => {
    const text = (key: string) => typeof row[key] === 'string' && row[key] !== '' ? row[key] as string : '未返回或无效'
    return {
      index, service: text('service_id'), instance: text('instance_id'), client: text('client_id'),
      hostname: text('hostname'), version: text('ceph_version'), health: text('health'),
      leader: row.leader === true ? '是' : row.leader === false ? '否' : '未返回或无效',
      callouts: row.callouts === undefined ? '未返回提示' : Array.isArray(row.callouts) && row.callouts.every((item: unknown) => typeof item === 'string')
        ? row.callouts.length ? row.callouts.join('\n') : '空提示列表' : '提示格式无效'
    }
  })
}

export function RbdMirrorDaemons({ value }: { value: unknown }) {
  const rows = rbdMirrorDaemonRows(value)
  if (!rows) return <Alert type="warning" message="同步守护进程信息不可用或格式无效" />
  if (!rows.length) return <Alert type="info" message="本次池状态未返回同步守护进程，不代表同步健康" />
  return <Space direction="vertical" style={{ minWidth: 640, width: '100%' }}>
    <span>来源：rbd mirror pool status --verbose；Leader 与健康状态仅针对本池。库存不是实时状态，也不是跨池汇总健康。</span>
    <Table size="small" rowKey="index" pagination={{ pageSize: 5 }} scroll={{ x: 1100 }} dataSource={rows} columns={[
      { title: '实例 ID', dataIndex: 'instance' }, { title: '服务 ID', dataIndex: 'service' },
      { title: '客户端 ID', dataIndex: 'client' }, { title: '主机', dataIndex: 'hostname' },
      { title: 'Ceph 版本', dataIndex: 'version' }, { title: '本池 Leader', dataIndex: 'leader' },
      { title: '原生健康状态', dataIndex: 'health' },
      { title: '原生提示', dataIndex: 'callouts', render: (text: string) => <span style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{text}</span> }
    ]} />
  </Space>
}

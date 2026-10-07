import { Alert, Space, Table } from 'antd'

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

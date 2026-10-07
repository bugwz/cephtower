import { Alert, Descriptions, Space, Table } from 'antd'

function record(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}
function text(row: Record<string, unknown>, key: string) {
  return typeof row[key] === 'string' ? row[key] as string : '未返回或无效'
}
export function rbdMirrorImageRows(value: unknown) {
  if (!Array.isArray(value) || !value.every(record)) return undefined
  return value.map((row, index) => ({
    index, name: text(row, 'name'), globalId: text(row, 'global_id'), state: text(row, 'state'),
    description: text(row, 'description'), updated: text(row, 'last_update'),
    daemon: record(row.daemon_service) ? row.daemon_service : undefined,
    peers: Array.isArray(row.peer_sites) && row.peer_sites.every(record) ? row.peer_sites.map((peer, index) => ({
      index, name: text(peer, 'site_name'), uuid: text(peer, 'mirror_uuid'), state: text(peer, 'state'),
      description: text(peer, 'description'), updated: text(peer, 'last_update')
    })) : undefined
  }))
}
const longText = (value: string) => <span style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{value}</span>

export function RbdMirrorImages({ value }: { value: unknown }) {
  const rows = rbdMirrorImageRows(value)
  if (!rows) return <Alert type="warning" message="镜像同步详情不可用或格式无效" />
  if (!rows.length) return <Alert type="info" message="本次池状态未返回镜像，不代表同步完成" />
  return <Space direction="vertical" style={{ minWidth: 640, width: '100%' }}>
    <span>来源：rbd mirror pool status --verbose。状态与时间保留原文；本地和各远端分别报告，不推断同步已完成。展开镜像查看站点及守护进程。</span>
    <Table size="small" rowKey="index" pagination={{ pageSize: 5 }} scroll={{ x: 950 }} dataSource={rows} columns={[
      { title: '镜像', dataIndex: 'name' }, { title: '全局 ID', dataIndex: 'globalId' },
      { title: '本地原生状态', dataIndex: 'state' }, { title: '本地描述', dataIndex: 'description', render: longText },
      { title: '本地更新时间（原文）', dataIndex: 'updated' }
    ]} expandable={{ expandedRowRender: row => <Space direction="vertical" style={{ width: '100%' }}>
      {row.daemon ? <Descriptions size="small" bordered column={2} items={[
        ['service_id', '服务 ID'], ['instance_id', '实例 ID'], ['daemon_id', '守护进程 ID'], ['hostname', '主机']
      ].map(([key, label]) => ({ key, label, children: longText(text(row.daemon!, key)) }))} /> : <Alert type="info" message="未返回有效关联守护进程，不能据此判定进程停止" />}
      {!row.peers ? <Alert type="info" message="未返回有效远端站点状态，不代表未配置 Peer" /> : !row.peers.length ? <Alert type="info" message="远端站点状态列表为空" /> : <Table size="small" rowKey="index" pagination={{ pageSize: 5 }} dataSource={row.peers} columns={[
        { title: '远端站点名称', dataIndex: 'name', render: value => value === '' ? '名称未解析' : value },
        { title: 'Mirror UUID', dataIndex: 'uuid' }, { title: '原生状态', dataIndex: 'state' },
        { title: '描述', dataIndex: 'description', render: longText }, { title: '更新时间（原文）', dataIndex: 'updated' }
      ]} />}
    </Space> }} />
  </Space>
}

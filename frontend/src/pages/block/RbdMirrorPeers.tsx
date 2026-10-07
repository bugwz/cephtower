import { Alert, Space, Table } from 'antd'

function peerRows(value: unknown): Record<string, unknown>[] | undefined {
  return Array.isArray(value) && value.every(item => item && typeof item === 'object' && !Array.isArray(item)) ? value : undefined
}
export function rbdMirrorPeerOptions(row?: Record<string, unknown>) {
  const peers = peerRows(row?.peers) ?? []
  return peers.filter(peer => typeof peer.uuid === 'string' && peer.uuid.trim() === peer.uuid && /^[0-9a-f]{8}(-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i.test(peer.uuid) && peers.filter(other => typeof other.uuid === 'string' && other.uuid.toLowerCase() === String(peer.uuid).toLowerCase()).length === 1)
    .map(peer => ({ value: peer.uuid as string, label: `${typeof peer.site_name === 'string' && peer.site_name ? peer.site_name : '未命名站点'} · ${peer.uuid}` }))
}
export function rbdMirrorPeerIdentity(values: Record<string, unknown>, row?: Record<string, unknown>): string {
  if (typeof values.uuid !== 'string' || !rbdMirrorPeerOptions(row).some(option => option.value === values.uuid)) throw new Error('请选择当前池唯一有效的 Peer UUID；库存不可用时请先重新采集')
  return values.uuid
}
export function RbdMirrorPeers({ value }: { value: unknown }) {
  const peers = peerRows(value)
  if (!peers) return <Alert type="warning" message="Peer 库存不可用或格式无效" />
  if (!peers.length) return <Alert type="info" message="本次池信息未返回 Peer" />
  return <Space direction="vertical" style={{ minWidth: 500, width: '100%' }}>
    <span>Peer UUID 用于编辑/删除；Mirror UUID 标识远端站点，两者不能互换。方向为原生配置，不代表当前复制健康；库存可能已过期。</span>
    <Table size="small" rowKey="index" pagination={{ pageSize: 5 }} scroll={{ x: 850 }} dataSource={peers.map((peer,index)=>({index,...Object.fromEntries(['uuid','site_name','mirror_uuid','client_name','direction'].map(key=>[key,typeof peer[key]==='string' && peer[key]!=='' ? peer[key] : '未返回或为空']))}))} columns={[
      { title: 'Peer UUID', dataIndex: 'uuid' }, { title: '站点名称', dataIndex: 'site_name' },
      { title: '远端 Mirror UUID', dataIndex: 'mirror_uuid' }, { title: '客户端名称', dataIndex: 'client_name' },
      { title: '原生同步方向', dataIndex: 'direction' }
    ]} />
  </Space>
}

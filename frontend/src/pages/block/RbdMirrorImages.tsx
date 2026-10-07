import { Alert, Descriptions, Space, Table } from 'antd'
import type { Key } from 'react'

function record(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}
function text(row: Record<string, unknown>, key: string) {
  return typeof row[key] === 'string' ? row[key] as string : '未返回或无效'
}
export function rbdMirrorStateCategory(value: unknown): string {
  if (typeof value !== 'string' || value.trim() !== value || !/^(up|down)\+(unknown(?: \([0-9]+\))?|error|syncing|starting_replay|replaying|stopping_replay|stopped)$/.test(value)) return '未知或无效'
  if (value.startsWith('down+') || value === 'up+error') return '需关注'
  if (['up+syncing', 'up+starting_replay', 'up+replaying'].includes(value)) return '同步或重放中'
  if (['up+stopping_replay', 'up+stopped'].includes(value)) return '停止中或已停止'
  return '未知或无效'
}
const categories = ['需关注', '同步或重放中', '停止中或已停止', '未知或无效']
const categoryColumn = {
  title: '状态分类', dataIndex: 'category',
  filters: categories.map(value => ({ text: value, value })),
  onFilter: (value: Key | boolean, row: { category: string }) => row.category === value
}
export function rbdMirrorImageRows(value: unknown) {
  if (!Array.isArray(value) || !value.every(record)) return undefined
  return value.map((row, index) => ({
    index, name: text(row, 'name'), globalId: text(row, 'global_id'), state: text(row, 'state'),
    category: rbdMirrorStateCategory(row.state),
    mode: row.mirror_mode === 'journal' || row.mirror_mode === 'snapshot' ? row.mirror_mode : '未返回有效模式',
    role: row.mirror_primary === true ? '主端' : row.mirror_primary === false ? '非主端' : '未返回有效角色',
    mirrorState: typeof row.mirror_image_state === 'string' && ['enabled', 'disabling', 'creating'].includes(row.mirror_image_state) ? row.mirror_image_state : '未返回有效配置状态',
    metrics: row.replay_metrics,
    bootstrap: row.bootstrap_percent,
    description: text(row, 'description'), updated: text(row, 'last_update'),
    daemon: record(row.daemon_service) ? row.daemon_service : undefined,
    peers: Array.isArray(row.peer_sites) && row.peer_sites.every(record) ? row.peer_sites.map((peer, index) => ({
      index, name: text(peer, 'site_name'), uuid: text(peer, 'mirror_uuid'), state: text(peer, 'state'),
      category: rbdMirrorStateCategory(peer.state),
      metrics: peer.replay_metrics,
      bootstrap: peer.bootstrap_percent,
      description: text(peer, 'description'), updated: text(peer, 'last_update')
    })) : undefined
  }))
}
const longText = (value: string) => <span style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{value}</span>

export function RbdBootstrapProgress({ value }: { value: unknown }) {
  const valid = typeof value === 'string' && value.trim() === value && /^(0|[1-9][0-9]?|100)$/.test(value)
  return <span>{valid ? `${value}%（引导复制阶段，不代表整体同步完成）` : '未返回有效复制阶段进度'}</span>
}

export function RbdReplayMetrics({ value, mode }: { value: unknown; mode?: unknown }) {
  return <Descriptions size="small" column={2} title="描述中的重放统计（采集快照，非实时速率；预计时间不保证完成）" items={[
    ['bytes_per_second', '速率（B/s）'], ['seconds_until_synced', '预计剩余秒数'],
    ['syncing_percent', '同步百分比（%）'], ['entries_behind_primary', '落后主端条目数'], ['replay_state', '快照重放子状态']
  ].map(([key, label]) => ({ key, label, children: key === 'entries_behind_primary' && mode === 'snapshot' ? '不适用（快照同步模式）' : key === 'replay_state' ? rbdReplayStateText(record(value) ? value[key] : undefined) : record(value) && typeof value[key] === 'string' ? value[key] as string : '未返回有效值' }))} />
}

export function rbdReplayStateText(value: unknown): string {
  if (value === 'idle') return 'idle（快照重放空闲，不代表所有站点同步完成）'
  if (value === 'syncing') return 'syncing（快照复制中）'
  return '未返回有效值'
}

export function RbdMirrorImages({ value, mode }: { value: unknown; mode?: unknown }) {
  if (mode === 'disabled') return <Alert type="info" message="池同步已禁用，未查询镜像同步运行状态；不代表池内没有镜像" />
  const rows = rbdMirrorImageRows(value)
  if (!rows) return <Alert type="warning" message="镜像同步详情不可用或格式无效" />
  if (!rows.length) return <Alert type="info" message="本次池状态未返回镜像，不代表同步完成" />
  return <Space direction="vertical" style={{ minWidth: 640, width: '100%' }}>
    <span>来源：rbd mirror pool status --verbose；模式、角色和配置状态关联同池默认命名空间的 rbd info（名称及全局 ID 一致）。均为采集快照，不推断同步已完成。展开镜像查看站点及守护进程。</span>
    <span>本池已返回镜像的本地状态（筛选前）：{categories.map(category => `${category} ${rows.filter(row => row.category === category).length}`).join('；')}。停止不表示同步成功；远端分类单独展示。</span>
    <Table size="small" rowKey="index" pagination={{ pageSize: 5 }} scroll={{ x: 950 }} dataSource={rows} columns={[
      { title: '镜像', dataIndex: 'name' }, { title: '全局 ID', dataIndex: 'globalId' },
      { title: '镜像同步模式', dataIndex: 'mode' }, { title: '本地角色', dataIndex: 'role' }, { title: '同步配置状态', dataIndex: 'mirrorState' },
      { title: '本地原生状态', dataIndex: 'state' }, { title: '本地描述', dataIndex: 'description', render: longText },
      { title: '本地更新时间（原文）', dataIndex: 'updated' }, categoryColumn,
      { title: '本地引导复制进度', dataIndex: 'bootstrap', render: value => <RbdBootstrapProgress value={value} /> }
    ]} expandable={{ expandedRowRender: row => <Space direction="vertical" style={{ width: '100%' }}>
      {row.daemon ? <Descriptions size="small" bordered column={2} items={[
        ['service_id', '服务 ID'], ['instance_id', '实例 ID'], ['daemon_id', '守护进程 ID'], ['hostname', '主机']
      ].map(([key, label]) => ({ key, label, children: longText(text(row.daemon!, key)) }))} /> : <Alert type="info" message="未返回有效关联守护进程，不能据此判定进程停止" />}
      {!row.peers ? <Alert type="info" message="未返回有效远端站点状态，不代表未配置 Peer" /> : !row.peers.length ? <Alert type="info" message="远端站点状态列表为空" /> : <Table size="small" rowKey="index" pagination={{ pageSize: 5 }} dataSource={row.peers} columns={[
        { title: '远端站点名称', dataIndex: 'name', render: value => value === '' ? '名称未解析' : value },
        { title: 'Mirror UUID', dataIndex: 'uuid' }, { title: '原生状态', dataIndex: 'state' },
        { title: '描述', dataIndex: 'description', render: longText }, { title: '更新时间（原文）', dataIndex: 'updated' }, categoryColumn,
        { title: '远端重放统计', dataIndex: 'metrics', render: value => <RbdReplayMetrics value={value} /> },
        { title: '远端引导复制进度', dataIndex: 'bootstrap', render: value => <RbdBootstrapProgress value={value} /> }
      ]} />}
      <RbdReplayMetrics value={row.metrics} mode={row.mode} />
    </Space> }} />
  </Space>
}

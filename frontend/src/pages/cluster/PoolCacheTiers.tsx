import { Alert, Button, Card, Space, Table } from 'antd'
import { useCallback } from 'react'
import { isRecord, type ApiRecord } from '../../api/client'
import { listAllResources, type ResourceListResult } from '../../api/resource'
import { useResource } from '../../hooks'

export function cacheTierRows(poolId: number, inventory: ResourceListResult) {
  const base = inventory.items.filter((row) => row.id === poolId)
  if (base.length !== 1 || !isRecord(base[0].raw_detail)) throw new Error('当前池的缓存层关系不可用')
  const tiers = base[0].raw_detail.tiers
  if (!Array.isArray(tiers) || tiers.some((id) => !Number.isSafeInteger(id) || id < 0) || new Set(tiers).size !== tiers.length) throw new Error('原生缓存层 ID 列表缺失或无效')
  const rows = tiers.map((id) => {
    const matches = inventory.items.filter((row) => row.id === id)
    const row = matches.length === 1 ? matches[0] : undefined
    const raw = row && isRecord(row.raw_detail) ? row.raw_detail : undefined
    return { id, name: typeof row?.name === 'string' ? row.name : '未解析', available: Boolean(raw), raw: raw ?? {} }
  })
  return { rows, stale: inventory.stale !== false || base[0].stale !== false || tiers.some((id) => inventory.items.some((row) => row.id === id && row.stale !== false)) }
}

export function cacheTierCounter(value: unknown) {
  return typeof value === 'string' && /^(0|[1-9][0-9]*)$/.test(value) ? value : '未提供或无效'
}

export function PoolCacheTiers({ clusterId, poolId }: { clusterId: number; poolId: number }) {
  const loader = useCallback(async () => cacheTierRows(poolId, await listAllResources('/pools', clusterId)), [clusterId, poolId])
  const { data, loading, error, refresh } = useResource(loader)
  const fields = [
    ['cache_min_evict_age', '最短淘汰时间（秒）'], ['cache_min_flush_age', '最短刷写时间（秒）'],
    ['target_max_bytes', '目标最大字节数（B）'], ['target_max_objects', '目标最大对象数'],
  ]
  return <Card className="page-surface-card" title="缓存层详情（采集时）" extra={<Button loading={loading} onClick={() => refresh()}>重读库存</Button>}>
    <Space direction="vertical" style={{ width: '100%' }}>
      <Alert type="info" message="按当前池的 tiers ID 关联缓存池，来源为 ceph osd pool ls detail。此处只读展示，不执行缓存层创建、移除、刷写或淘汰。" />
      {error && <Alert type="error" message={error} description={data ? '下方为上次读取结果，不代表当前状态。' : undefined} />}
      {data?.stale && <Alert type="warning" message="关联库存过期或新鲜度未知，以下仅为历史关系。" />}
      {data?.rows.some((row) => !row.available) && <Alert type="warning" message="部分缓存池未在库存中唯一解析，不能确认其配置；保留原始 ID 供核实。" />}
      <Table<{ id: number; name: string; available: boolean; raw: ApiRecord }> size="small" loading={loading} rowKey="id" pagination={false} scroll={{ x: 'max-content' }} dataSource={data?.rows ?? []} locale={{ emptyText: data ? '当前采集的 tiers 列表为空' : '缓存层关系未读取' }} columns={[
        { title: '缓存池 ID', dataIndex: 'id' }, { title: '名称', dataIndex: 'name' },
        { title: '缓存模式', key: 'mode', render: (_, row) => typeof row.raw.cache_mode === 'string' && row.raw.cache_mode ? row.raw.cache_mode : '未提供' },
        ...fields.map(([key, title]) => ({ title, key, render: (_: unknown, row: { raw: ApiRecord }) => cacheTierCounter(row.raw[key]) })),
      ]} />
    </Space>
  </Card>
}

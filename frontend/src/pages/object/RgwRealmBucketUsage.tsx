import { Alert, Button, Card, Descriptions, Space, Table } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { jsonInit, request } from '../../api/client'
import { useClusterContext } from '../../state/ClusterContext'

type Row = { realm_id: string; zone_id: string; service_map_id: string; bucket_count: string; object_count: string; size_actual_bytes: string; usage_category: 'rgw.main'; started_at: string; observed_at: string; scope: 'zone'; source: 'radosgw-admin' }
type Counts = { scope: 'registered_realms'; source: 'service_map+radosgw-admin'; selection: 'lowest_zone_id_then_service_map_id'; realm_count: number; bucket_count: string; object_count: string; size_actual_bytes: string; usage_category: 'rgw.main'; started_at: string; observed_at: string; items: Row[] }
export function realmBucketUsageData(value: unknown): Counts {
  const d = value as Counts
  const decimal = (v: unknown) => typeof v === 'string' && /^(0|[1-9][0-9]*)$/.test(v)
  const interval = (a: unknown, b: unknown) => typeof a === 'string' && typeof b === 'string' && Number.isFinite(Date.parse(a)) && Number.isFinite(Date.parse(b)) && Date.parse(a) <= Date.parse(b)
  const identity = (v: unknown) => typeof v === 'string' && v.length > 0 && !/[\u0000-\u001f\u007f-\u009f]/.test(v)
  if (!d || d.scope !== 'registered_realms' || d.source !== 'service_map+radosgw-admin' || d.selection !== 'lowest_zone_id_then_service_map_id' || !Number.isSafeInteger(d.realm_count) || d.realm_count < 0 || d.usage_category !== 'rgw.main' || !decimal(d.object_count) || !decimal(d.size_actual_bytes) || !decimal(d.bucket_count) || !interval(d.started_at, d.observed_at) || !Array.isArray(d.items) || d.items.length !== d.realm_count) throw new Error('Invalid realm aggregate')
  const realms = new Set<string>(), zones = new Set<string>(), services = new Set<string>()
  let total = 0n, objects = 0n, bytes = 0n
  const items = d.items.map(row => {
    if (!row || row.scope !== 'zone' || row.source !== 'radosgw-admin' || !identity(row.realm_id) || !identity(row.zone_id) || !identity(row.service_map_id) || realms.has(row.realm_id) || zones.has(row.zone_id) || services.has(row.service_map_id) || row.usage_category !== 'rgw.main' || !decimal(row.object_count) || !decimal(row.size_actual_bytes) || !decimal(row.bucket_count) || !interval(row.started_at, row.observed_at) || Date.parse(row.started_at) < Date.parse(d.started_at) || Date.parse(row.observed_at) > Date.parse(d.observed_at)) throw new Error('Invalid realm item')
    realms.add(row.realm_id); zones.add(row.zone_id); services.add(row.service_map_id)
    if (row.bucket_count === '0' && (row.object_count !== '0' || row.size_actual_bytes !== '0')) throw new Error('Invalid empty bucket aggregate')
    total += BigInt(row.bucket_count); objects += BigInt(row.object_count); bytes += BigInt(row.size_actual_bytes)
    return { realm_id: row.realm_id, zone_id: row.zone_id, service_map_id: row.service_map_id, bucket_count: row.bucket_count, object_count: row.object_count, size_actual_bytes: row.size_actual_bytes, usage_category: row.usage_category, started_at: row.started_at, observed_at: row.observed_at, scope: row.scope, source: row.source }
  })
  if (total.toString() !== d.bucket_count || objects.toString() !== d.object_count || bytes.toString() !== d.size_actual_bytes) throw new Error('Inconsistent realm total')
  return { scope: d.scope, source: d.source, selection: d.selection, realm_count: d.realm_count, bucket_count: d.bucket_count, object_count: d.object_count, size_actual_bytes: d.size_actual_bytes, usage_category: d.usage_category, started_at: d.started_at, observed_at: d.observed_at, items }
}
export function RgwRealmBucketUsage() {
  const { selectedClusterId } = useClusterContext()
  return <RgwRealmBucketUsageView key={selectedClusterId ?? 'none'} clusterId={selectedClusterId} />
}
export function RgwRealmBucketUsageView({ clusterId }: { clusterId?: number }) {
  const current = useRef(clusterId), mounted = useRef(true), sequence = useRef(0), abort = useRef<AbortController>()
  current.current = clusterId
  const [state, setState] = useState<{ clusterId?: number; busy: boolean; data?: Counts; error?: boolean }>({ busy: false })
  useEffect(() => { mounted.current = true; setState({ clusterId, busy: false }); return () => { mounted.current = false; abort.current?.abort(); sequence.current++ } }, [clusterId])
  async function read() {
    if (!clusterId || !mounted.current || current.current !== clusterId) return
    abort.current?.abort()
    const controller = new AbortController(), ticket = ++sequence.current
    abort.current = controller
    setState({ clusterId, busy: true })
    try {
      const value = await request<unknown>('/rgw/realms/buckets/usage', jsonInit('GET', { cluster_id: clusterId }, { signal: controller.signal, cache: 'no-store', suppressErrorNotification: true }))
      if (mounted.current && current.current === clusterId && ticket === sequence.current && !controller.signal.aborted) setState({ clusterId, busy: false, data: realmBucketUsageData(value) })
    } catch {
      if (mounted.current && current.current === clusterId && ticket === sequence.current && !controller.signal.aborted) setState({ clusterId, busy: false, error: true })
    }
  }
  const scoped = state.clusterId === clusterId ? state : undefined
  return <Card title="已注册 Realm Bucket 使用量汇总"><Space direction="vertical" style={{ width: '100%' }}>
    <Alert type="info" message="每个已注册 Realm 选取一个代表 Zone" description="按 Zone ID、服务映射 ID 字典序选择。只覆盖 service map 中身份完整的 Realm，不是所有配置 Realm；不表示在线健康或 Zone 复制一致。逐项读取非原子快照，最多两分钟，任一失败不展示部分总数。不同 Realm 的同名 Bucket 分别计数；rgw.main 的 size_actual 为取整统计容量，不是物理占用。" />
    <Button disabled={!clusterId || scoped?.busy} loading={scoped?.busy} onClick={() => void read()}>读取跨 Realm Bucket 汇总</Button>
    {!clusterId && <Alert type="info" message="请先选择集群" />}
    {scoped?.error && <Alert type="error" message="汇总读取失败或响应无效，未展示旧值或部分计数" />}
    {scoped?.data && <>
      {scoped.data.realm_count === 0 && <Alert type="warning" message="没有已注册 Realm；范围内零值不代表集群没有 Bucket" />}
      <Descriptions column={1} items={[
        { key: 'realms', label: '已注册 Realm 数', children: String(scoped.data.realm_count) },
        { key: 'buckets', label: '范围内 Bucket 总数', children: scoped.data.bucket_count },
        { key: 'objects', label: '对象总数', children: scoped.data.object_count },
        { key: 'bytes', label: '统计容量（B）', children: scoped.data.size_actual_bytes },
        { key: 'start', label: '读取开始', children: scoped.data.started_at },
        { key: 'end', label: '读取完成', children: scoped.data.observed_at },
      ]} />
      <Table<Row> rowKey="realm_id" dataSource={scoped.data.items} pagination={{ pageSize: 10 }} scroll={{ x: 'max-content' }} columns={[
        { title: 'Realm ID', dataIndex: 'realm_id' }, { title: '代表 Zone ID', dataIndex: 'zone_id' },
        { title: '服务映射 ID', dataIndex: 'service_map_id' }, { title: 'Bucket 数', dataIndex: 'bucket_count' },
        { title: '对象数', dataIndex: 'object_count' }, { title: '统计容量（B）', dataIndex: 'size_actual_bytes' },
        { title: '读取开始', dataIndex: 'started_at' }, { title: '读取完成', dataIndex: 'observed_at' },
      ]} />
    </>}
  </Space></Card>
}

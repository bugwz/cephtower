import { Alert, Button, Card, Descriptions, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { jsonInit, request } from '../../api/client'
import { useClusterContext } from '../../state/ClusterContext'

type Usage = { bucket_count: number; object_count: string; size_actual_bytes: string; usage_category: 'rgw.main'; source: 'radosgw-admin'; started_at: string; observed_at: string }
export function bucketUsageData(value: unknown): Usage {
  const data = value as Usage
  const decimal = (v: unknown) => typeof v === 'string' && /^(0|[1-9][0-9]*)$/.test(v)
  if (!data || data.source !== 'radosgw-admin' || data.usage_category !== 'rgw.main' || !Number.isSafeInteger(data.bucket_count) || data.bucket_count < 0 || !decimal(data.object_count) || !decimal(data.size_actual_bytes) || typeof data.started_at !== 'string' || typeof data.observed_at !== 'string' || !Number.isFinite(Date.parse(data.started_at)) || !Number.isFinite(Date.parse(data.observed_at)) || Date.parse(data.started_at) > Date.parse(data.observed_at)) throw new Error('Invalid bucket usage')
  if (data.bucket_count === 0 && (data.object_count !== '0' || data.size_actual_bytes !== '0')) throw new Error('Invalid empty aggregate')
  return { bucket_count: data.bucket_count, object_count: data.object_count, size_actual_bytes: data.size_actual_bytes, usage_category: 'rgw.main', source: 'radosgw-admin', started_at: data.started_at, observed_at: data.observed_at }
}
export function averageBucketObjectBytes(data: Usage): string {
  const count = BigInt(data.object_count)
  if (count === 0n) return '不适用（对象数为 0）'
  // Fixed-point integer division avoids rounding large totals through Number.
  const hundredths = BigInt(data.size_actual_bytes) * 100n / count
  return `${hundredths / 100n}.${String(hundredths % 100n).padStart(2, '0')} B/对象`
}
export function RgwBucketUsage() {
  const { selectedClusterId } = useClusterContext()
  return <RgwBucketUsageView key={selectedClusterId ?? 'none'} clusterId={selectedClusterId} />
}
export function RgwBucketUsageView({ clusterId }: { clusterId?: number }) {
  const current = useRef(clusterId), mounted = useRef(true), sequence = useRef(0), abort = useRef<AbortController>()
  current.current = clusterId
  const [state, setState] = useState<{ clusterId?: number; busy: boolean; data?: Usage; error?: boolean }>({ busy: false })
  useEffect(() => { mounted.current = true; setState({ clusterId, busy: false }); return () => { mounted.current = false; abort.current?.abort(); sequence.current++ } }, [clusterId])
  async function read() {
    if (!clusterId || !mounted.current || current.current !== clusterId) return
    abort.current?.abort()
    const controller = new AbortController(), ticket = ++sequence.current
    abort.current = controller
    setState({ clusterId, busy: true })
    try {
      const value = await request<unknown>('/rgw/buckets/usage', jsonInit('GET', { cluster_id: clusterId }, { signal: controller.signal, cache: 'no-store', suppressErrorNotification: true }))
      if (mounted.current && current.current === clusterId && ticket === sequence.current && !controller.signal.aborted) setState({ clusterId, busy: false, data: bucketUsageData(value) })
    } catch {
      if (mounted.current && current.current === clusterId && ticket === sequence.current && !controller.signal.aborted) setState({ clusterId, busy: false, error: true })
    }
  }
  const scoped = state.clusterId === clusterId ? state : undefined
  return <Card title="Bucket 使用量汇总"><Space direction="vertical" style={{ width: '100%' }}>
    <Alert type="info" message="按需完整枚举并逐桶读取，不自动轮询" description="来自 radosgw-admin 的 rgw.main 统计。扫描最多两分钟，顺序读取不是原子快照；任一桶统计缺失或失败时不展示部分汇总。size_actual 是取整后的统计容量，不是物理磁盘占用。平均值按该容量除以对象数，截断到小数点后两位。" />
    <Button disabled={!clusterId || scoped?.busy} loading={scoped?.busy} onClick={() => void read()}>读取 Bucket 使用量</Button>
    {!clusterId && <Alert type="info" message="请先选择集群" />}
    {scoped?.error && <Alert type="error" message="使用量读取失败或响应无效，未展示旧值或部分汇总" />}
    {scoped?.data && <Descriptions column={1} items={[
      { key: 'buckets', label: 'Bucket 数量', children: String(scoped.data.bucket_count) },
      { key: 'objects', label: '对象数', children: scoped.data.object_count },
      { key: 'bytes', label: '统计容量（B）', children: scoped.data.size_actual_bytes },
      { key: 'average', label: '平均对象大小', children: averageBucketObjectBytes(scoped.data) },
      { key: 'start', label: '读取开始', children: scoped.data.started_at },
      { key: 'end', label: '读取完成', children: scoped.data.observed_at },
    ]} />}
  </Space></Card>
}

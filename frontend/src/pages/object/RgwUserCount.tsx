import { Alert, Button, Card, Descriptions, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { jsonInit, request } from '../../api/client'
import { useClusterContext } from '../../state/ClusterContext'

type Count = { user_count: number; scope: 'current_rgw_configuration'; source: 'radosgw-admin'; started_at: string; observed_at: string }
export function userCountData(value: unknown): Count {
  const data = value as Count
  if (!data || data.scope !== 'current_rgw_configuration' || data.source !== 'radosgw-admin' || !Number.isSafeInteger(data.user_count) || data.user_count < 0 || typeof data.started_at !== 'string' || typeof data.observed_at !== 'string' || !Number.isFinite(Date.parse(data.started_at)) || !Number.isFinite(Date.parse(data.observed_at)) || Date.parse(data.started_at) > Date.parse(data.observed_at)) throw new Error('Invalid user count')
  return { user_count: data.user_count, scope: data.scope, source: data.source, started_at: data.started_at, observed_at: data.observed_at }
}
export function RgwUserCount() {
  const { selectedClusterId } = useClusterContext()
  return <RgwUserCountView key={selectedClusterId ?? 'none'} clusterId={selectedClusterId} />
}
export function RgwUserCountView({ clusterId }: { clusterId?: number }) {
  const current = useRef(clusterId), mounted = useRef(true), sequence = useRef(0), abort = useRef<AbortController>()
  current.current = clusterId
  const [state, setState] = useState<{ clusterId?: number; busy: boolean; data?: Count; error?: boolean }>({ busy: false })
  useEffect(() => { mounted.current = true; setState({ clusterId, busy: false }); return () => { mounted.current = false; abort.current?.abort(); sequence.current++ } }, [clusterId])
  async function read() {
    if (!clusterId || !mounted.current || current.current !== clusterId) return
    abort.current?.abort()
    const controller = new AbortController(), ticket = ++sequence.current
    abort.current = controller
    setState({ clusterId, busy: true })
    try {
      const value = await request<unknown>('/rgw/users/count', jsonInit('GET', { cluster_id: clusterId }, { signal: controller.signal, cache: 'no-store', suppressErrorNotification: true }))
      if (mounted.current && current.current === clusterId && ticket === sequence.current && !controller.signal.aborted) setState({ clusterId, busy: false, data: userCountData(value) })
    } catch {
      if (mounted.current && current.current === clusterId && ticket === sequence.current && !controller.signal.aborted) setState({ clusterId, busy: false, error: true })
    }
  }
  const scoped = state.clusterId === clusterId ? state : undefined
  return <Card title="当前 RGW 配置用户数量"><Space direction="vertical" style={{ width: '100%' }}>
    <Alert type="info" message="来自 radosgw-admin user list 完整枚举" description="仅统计当前 RGW 配置范围的用户元数据身份，不是跨 Realm 总数、Account 数量或在线用户数。枚举期间可能有并发变更，不保证原子快照。按需读取，不自动轮询。" />
    <Button disabled={!clusterId || scoped?.busy} loading={scoped?.busy} onClick={() => void read()}>读取用户数量</Button>
    {!clusterId && <Alert type="info" message="请先选择集群" />}
    {scoped?.error && <Alert type="error" message="用户数量读取失败或响应无效，未展示旧值或部分计数" />}
    {scoped?.data && <Descriptions column={1} items={[
      { key: 'users', label: '用户数量', children: String(scoped.data.user_count) },
      { key: 'scope', label: '统计范围', children: '当前 RGW 配置（非跨 Realm 汇总）' },
      { key: 'start', label: '读取开始', children: scoped.data.started_at },
      { key: 'end', label: '读取完成', children: scoped.data.observed_at },
    ]} />}
  </Space></Card>
}

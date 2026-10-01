import { Alert, Button, Card, Space, Switch, Tag, Typography } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { jsonInit, request } from '../../api/client'
import { AppTable } from '../../components/AppTable'
import { performancePoints as points, type PerformancePoint as Point } from './cephfsPerformanceSeries'
import { CephFSRankStatus } from './CephFSRankStatus'

interface Counter { name: string; value: string | null }
interface Daemon {
  name: string; gid: string; rank: number; state: string
  counters: Counter[]; error?: string; observed_at: string
}
interface Performance { filesystem: string; items: Daemon[]; observed_at: string }

export function CephFSPerformance({ clusterId, filesystem }: { clusterId?: number; filesystem: string }) {
  const [data, setData] = useState<Performance | null>(null)
  const [history, setHistory] = useState<Record<string, Daemon[]>>({})
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [automatic, setAutomatic] = useState(true)
  const [revision, setRevision] = useState(0)
  const scope = useRef('')

  useEffect(() => {
    const key = `${clusterId}:${filesystem}`
    if (scope.current !== key) {
      scope.current = key
      setData(null)
      setHistory({})
      setError('')
      setLoading(false)
    }
    const controller = new AbortController()
    let timer: ReturnType<typeof setTimeout> | undefined
    async function sample() {
      if (!clusterId || !filesystem) return
      setLoading(true)
      try {
        const response = await request<Performance>('/filesystem/performance', jsonInit('GET', {
          cluster_id: clusterId, fs: filesystem
        }, { signal: controller.signal, suppressErrorNotification: true }))
        if (controller.signal.aborted) return
        setData(response)
        setError('')
        setHistory((previous) => {
          const next: Record<string, Daemon[]> = Object.create(null)
          for (const daemon of response.items) {
            const existing = Object.prototype.hasOwnProperty.call(previous, daemon.name) ? previous[daemon.name] : []
            const sameInstance = existing[existing.length - 1]?.gid === daemon.gid
            next[daemon.name] = daemon.error ? [] : [...(sameInstance ? existing : []), daemon].slice(-60)
          }
          return next
        })
      } catch (err) {
        if (!controller.signal.aborted) {
          setError(err instanceof Error ? err.message : '读取 MDS 性能计数器失败')
          setHistory({})
        }
      } finally {
        if (!controller.signal.aborted) {
          setLoading(false)
          if (automatic) timer = setTimeout(() => void sample(), 10000)
        }
      }
    }
    void sample()
    return () => { controller.abort(); if (timer) clearTimeout(timer) }
  }, [automatic, clusterId, filesystem, revision])

  return <Card title="MDS 性能计数器" extra={<Space>
    <Switch checked={automatic} onChange={setAutomatic} checkedChildren="自动采样" unCheckedChildren="手动采样" />
    <Button loading={loading} onClick={() => setRevision((value) => value + 1)}>采样</Button>
  </Space>}>
    <Alert type="info" showIcon message="实时 perf dump；每次完成后间隔 10 秒采样，保留最近 60 次" description="趋势仅来自当前页面会话，不是 ceph-mgr 历史数据。请求速率按相邻采样的累计差值 / 实际秒数计算；MDS 实例变化、计数器回退或读取失败时不衔接旧速率。缺失指标显示 —，不是 0。" />
    {error && <Alert type="error" showIcon message={error} description="下方保留最后一次成功结果，不代表当前状态。" />}
    <CephFSRankStatus clusterId={clusterId} filesystem={filesystem} automatic={automatic} revision={revision} samples={error ? [] : data?.items ?? []} history={history} />
    {data && data.items.length === 0 && <Alert type="info" showIcon message="该文件系统当前没有关联的 MDS 实例" />}
    {data?.items.map((daemon) => <Card key={daemon.name} type="inner" title={<Space>{daemon.name}<Tag>Rank {daemon.rank}</Tag><Tag>{daemon.state}</Tag></Space>}>
      <Typography.Text type="secondary">GID {daemon.gid} · 采样时间 {new Date(daemon.observed_at).toLocaleString()}</Typography.Text>
      {daemon.error ? <Alert type="warning" showIcon message="此 MDS 采样失败" description={daemon.error} /> : <>
        <Trend title="内存 inode 数 · mds_mem.ino" points={points(history[daemon.name] ?? [], 'mds_mem.ino', false)} color="#1677ff" />
        <Trend title="客户端请求 / 秒 · mds_server.handle_client_request" points={points(history[daemon.name] ?? [], 'mds_server.handle_client_request', true)} color="#52c41a" />
        <AppTable<Counter> rowKey="name" pagination={false} dataSource={daemon.counters} columns={[
          { title: '计数器', dataIndex: 'name' },
          { title: '当前值（累计计数器不等于每秒速率）', dataIndex: 'value', render: (value) => value ?? '—' }
        ]} />
      </>}
    </Card>)}
  </Card>
}

function Trend({ title, points: samples, color }: { title: string; points: Array<Point | null>; color: string }) {
  const valid = samples.filter((point): point is Point => point !== null)
  if (!valid.length) return <Typography.Paragraph>{title}：等待有效采样（速率至少需两次）</Typography.Paragraph>
  const first = valid[0].time
  const span = Math.max(1, valid[valid.length - 1].time - first)
  const max = Math.max(1, ...valid.map((point) => point.value))
  const position = (point: Point) => `${40 + (point.time - first) / span * 700},${110 - point.value / max * 80}`
  const segments: string[] = []
  let segment: string[] = []
  for (const point of samples) {
    if (point) segment.push(position(point))
    else { if (segment.length) segments.push(segment.join(' ')); segment = [] }
  }
  if (segment.length) segments.push(segment.join(' '))
  return <div>
    <Typography.Text>{title} · 最新 {valid[valid.length - 1].value.toLocaleString(undefined, { maximumFractionDigits: 2 })}</Typography.Text>
    <svg viewBox="0 0 800 140" width="100%" role="img" aria-label={title}>
      <line x1="40" y1="110" x2="740" y2="110" stroke="currentColor" opacity="0.3" />
      <text x="40" y="20" fill="currentColor" fontSize="12">{max.toLocaleString()}</text>
      <text x="40" y="130" fill="currentColor" fontSize="12">{new Date(first).toLocaleTimeString()}</text>
      <text x="740" y="130" textAnchor="end" fill="currentColor" fontSize="12">{new Date(valid[valid.length - 1].time).toLocaleTimeString()}</text>
      {segments.map((line, index) => <polyline key={index} points={line} fill="none" stroke={color} strokeWidth="2" />)}
      {valid.map((point) => { const [x, y] = position(point).split(','); return <circle key={point.time} cx={x} cy={y} r="3" fill={color}><title>{new Date(point.time).toLocaleTimeString()} · {point.value}</title></circle> })}
    </svg>
  </div>
}

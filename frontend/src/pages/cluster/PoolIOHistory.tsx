import { Alert, Button, Card, Empty, Space, Typography } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { isRecord } from '../../api/client'
import { queryMetricRange, type MetricResponse } from '../../api/external'
import { MetricNotices } from '../monitoring/MetricNotices'

interface Point { time: number; value: number }
const metrics = [
  { id: 'pool_read_bytes', title: '读取吞吐率', unit: 'B/s' },
  { id: 'pool_write_bytes', title: '写入吞吐率', unit: 'B/s' },
  { id: 'pool_read_ops', title: '读取 IOPS', unit: 'op/s' },
  { id: 'pool_write_ops', title: '写入 IOPS', unit: 'op/s' }
]

export function poolHistoryPoints(response: MetricResponse, poolId: number): Point[] {
  if (response.result_type !== 'matrix') throw new Error('历史指标响应不是时间序列')
  const series = response.series.filter((item) => isRecord(item.metric) && item.metric.pool_id === String(poolId))
  if (series.length > 1) throw new Error('同一池返回了重复历史序列')
  if (!series.length) return []
  if (!Array.isArray(series[0].values)) throw new Error('历史指标缺少样本数组')
  const result: Point[] = []
  for (const sample of series[0].values) {
    if (!Array.isArray(sample) || sample.length !== 2 || typeof sample[0] !== 'number' || !Number.isFinite(sample[0])) continue
    if (!Number.isFinite(new Date(sample[0] * 1000).getTime())) continue
    if (typeof sample[1] !== 'string' || sample[1].trim() !== sample[1] || !/^[+-]?(?:\d+(?:\.\d*)?|\.\d+)(?:[eE][+-]?\d+)?$/.test(sample[1])) continue
    const value = Number(sample[1])
    if (!Number.isFinite(value) || value < 0) continue
    if (value === 0 && /[1-9]/.test(sample[1].split(/[eE]/)[0])) continue
    result.push({ time: sample[0] * 1000, value })
  }
  result.sort((a, b) => a.time - b.time)
  if (result.some((point, index) => index > 0 && point.time === result[index - 1].time)) throw new Error('历史指标包含重复时间戳')
  return result
}

export function poolHistoryPath(points: Point[]): string {
  if (!points.length) return ''
  const first = points[0].time
  const span = Math.max(1, points[points.length - 1].time - first)
  const max = Math.max(1, ...points.map((point) => point.value))
  return points.map((point, index) => `${index === 0 || point.time - points[index - 1].time > 45000 ? 'M' : 'L'}${40 + (point.time - first) / span * 700},${110 - point.value / max * 80}`).join(' ')
}

export function PoolIOHistory({ clusterId, poolId }: { clusterId: number; poolId: number }) {
  const [data, setData] = useState<{ points: Point[]; meta?: MetricResponse['meta'] }[] | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const controller = useRef<AbortController | null>(null)
  useEffect(() => () => controller.current?.abort(), [])
  async function query() {
    if (loading) return
    controller.current?.abort()
    const active = new AbortController()
    controller.current = active
    setLoading(true)
    setData(null)
    setError('')
    const end = new Date()
    const start = new Date(end.getTime() - 3600000)
    try {
      const results = await Promise.all(metrics.map(async (metric) => {
        const result = await queryMetricRange(clusterId, {
          metricId: metric.id, start: start.toISOString(), end: end.toISOString(), step: '30s'
        }, { signal: active.signal, suppressErrorNotification: true })
        return { points: poolHistoryPoints(result, poolId), meta: result.meta }
      }))
      if (!active.signal.aborted) setData(results)
    } catch (err) {
      if (!active.signal.aborted) setError(err instanceof Error ? err.message : '历史查询失败')
    } finally {
      if (!active.signal.aborted) setLoading(false)
    }
  }
  return <Card title="池 I/O 历史（最近一小时）" extra={<Button loading={loading} onClick={() => void query()}>查询历史</Button>}>
    <Space direction="vertical" className="full-width-control">
      <Typography.Text type="secondary">需要当前集群配置 Prometheus，并采集 Ceph pool 指标。曲线为 5 分钟窗口平均速率，每 30 秒一个点；不使用 CLI 快照拼造历史，缺失样本不补零。</Typography.Text>
      {error && <Alert type="warning" showIcon message="无法读取池历史指标" description={error} />}
      {!data && !error && <Typography.Text>点击查询读取历史样本。</Typography.Text>}
      {data?.map(({ points, meta }, index) => <div key={metrics[index].id}>
        <Typography.Text>{metrics[index].title} · {metrics[index].unit}</Typography.Text>
        <MetricNotices meta={meta} source={metrics[index].title} />
        {!points.length ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="此池没有有效历史样本" /> : <>
          <Typography.Paragraph>最新样本：{new Date(points[points.length - 1].time).toLocaleString()} · {points[points.length - 1].value.toLocaleString(undefined, { maximumFractionDigits: 2 })} {metrics[index].unit}；图中最大值 {Math.max(...points.map((point) => point.value)).toLocaleString()}</Typography.Paragraph>
          <svg viewBox="0 0 800 145" width="100%" role="img" aria-label={`${metrics[index].title}历史曲线`}>
            <line x1="40" y1="110" x2="740" y2="110" stroke="currentColor" opacity="0.3" />
            <path d={poolHistoryPath(points)} fill="none" stroke="#1677ff" strokeWidth="2" />
            {points.length === 1 && <circle cx="40" cy={110 - points[0].value / Math.max(1, points[0].value) * 80} r="3" fill="#1677ff" />}
            <text x="40" y="135" fill="currentColor" fontSize="12">{new Date(points[0].time).toLocaleTimeString()}</text>
            <text x="740" y="135" textAnchor="end" fill="currentColor" fontSize="12">{new Date(points[points.length - 1].time).toLocaleTimeString()}</text>
          </svg>
        </>}
      </div>)}
    </Space>
  </Card>
}

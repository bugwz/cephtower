import { Alert } from 'antd'
import type { ApiRecord } from '../../api/client'

interface Sample { index: number; timestamp: string; value: string; utc: string; status: string }

export function metricTrend(samples: Sample[], meta?: ApiRecord) {
  const start = typeof meta?.start === 'string' ? Date.parse(meta.start) / 1000 : NaN
  const end = typeof meta?.end === 'string' ? Date.parse(meta.end) / 1000 : NaN
  const step = meta?.step_seconds
  if (!Number.isFinite(start) || !Number.isFinite(end) || end <= start || typeof step !== 'number' || !Number.isFinite(step) || step <= 0) return null
  const valid = (sample: Sample) => sample.status === '有效值' && Number.isFinite(Number(sample.value)) && Number.isFinite(Number(sample.timestamp)) && Number(sample.timestamp) >= start && Number(sample.timestamp) <= end
  const finite = samples.filter(valid)
  if (!finite.length) return null
  const min = finite.reduce((value, sample) => Math.min(value, Number(sample.value)), Infinity)
  const max = finite.reduce((value, sample) => Math.max(value, Number(sample.value)), -Infinity)
  const scale = Math.max(Math.abs(min), Math.abs(max), 1)
  const lower = min / scale, upper = max / scale
  let previous: number | undefined
  const points: { x: number; y: number; sample: Sample }[] = []
  const path: string[] = []
  for (const sample of samples) {
    if (!valid(sample)) { previous = undefined; continue }
    const time = Number(sample.timestamp)
    const x = 60 + (time - start) / (end - start) * 680
    const y = upper === lower ? 85 : 140 - (Number(sample.value) / scale - lower) / (upper - lower) * 110
    // Never interpolate across missing evaluations or non-increasing timestamps.
    const command = previous !== undefined && time > previous && time - previous <= step * 1.000001 ? 'L' : 'M'
    path.push(`${command}${x},${y}`)
    points.push({ x, y, sample })
    previous = time
  }
  return { path: path.join(' '), points, min, max, start: String(meta!.start), end: String(meta!.end) }
}

export function MetricTrend({ samples, meta, name }: { samples: Sample[]; meta?: ApiRecord; name: string }) {
  const chart = metricTrend(samples, meta)
  if (!chart) return <Alert type="info" message="无可绘制的有限样本，或查询时间范围不可用；请查看原始样本。" />
  return <div>
    <div>趋势为浮点近似，各序列独立纵轴；缺失、异常或非有限样本处断线，不补零。悬停采样点查看原值。</div>
    <svg viewBox="0 0 800 190" width="100%" role="img" aria-label={`${name}范围查询趋势`}>
      <line x1="60" y1="140" x2="740" y2="140" stroke="currentColor" />
      <path d={chart.path} fill="none" stroke="#1677ff" strokeWidth="2" />
      {chart.points.map(point => <circle key={point.sample.index} cx={point.x} cy={point.y} r="2.5" fill="#1677ff"><title>{point.sample.utc}: {point.sample.value}</title></circle>)}
      <text x="60" y="18" fill="currentColor" fontSize="12">范围：{chart.min} — {chart.max}</text>
      <text x="60" y="164" fill="currentColor" fontSize="12">{chart.start}</text>
      <text x="740" y="182" textAnchor="end" fill="currentColor" fontSize="12">{chart.end}</text>
    </svg>
  </div>
}

import { Alert, Collapse, Empty, Typography } from 'antd'
import type { ApiRecord } from '../../api/client'
import { RecordDetail } from '../../components/RecordDetail'

type Range = { min?: number; max?: number }
type Axis = { name: string; ranges: Range[] }
type Histogram = { name: string; axes: [Axis, Axis]; values: number[][] }
const object = (value: unknown): value is Record<string, unknown> => Boolean(value) && typeof value === 'object' && !Array.isArray(value)

function histograms(value: unknown, path = ''): Histogram[] {
  if (!object(value)) return []
  if (Array.isArray(value.axes)) {
    const axes = value.axes
    if (axes.length !== 2 || !axes.every((axis) => object(axis) && typeof axis.name === 'string' && Array.isArray(axis.ranges) && axis.ranges.length > 0 && axis.ranges.every((range: unknown) => object(range) && [range.min, range.max].every((bound) => bound === undefined || (typeof bound === 'number' && Number.isFinite(bound)))))) return []
    if (!Array.isArray(value.values) || value.values.length !== axes[0].ranges.length || !value.values.every((row) => Array.isArray(row) && row.length === axes[1].ranges.length && row.every((cell) => typeof cell === 'number' && Number.isFinite(cell) && cell >= 0))) return []
    return [{ name: path, axes: axes as [Axis, Axis], values: value.values as number[][] }]
  }
  return Object.entries(value).flatMap(([key, nested]) => histograms(nested, path ? `${path}.${key}` : key))
}

function label(range: Range) {
  if (range.min === undefined) return range.max === undefined ? '全部' : `≤ ${range.max}`
  return range.max === undefined ? `≥ ${range.min}` : `${range.min} ～ ${range.max}`
}

export function OSDHistogram({ data }: { data: ApiRecord }) {
  const charts = histograms(data)
  return <>
    <Alert type="info" message="每个格子表示对应两个分桶范围内的累计事件数，边界均包含端点。颜色按当前直方图的计数对数缩放；这些数值不是每秒速率。坐标单位沿用 Ceph 返回的轴名称。" />
    {!charts.length && <Empty description="未返回可绘制的二维直方图，可展开原始数据检查。" />}
    {charts.map((chart) => {
      const maximum = chart.values.reduce((max, row) => row.reduce((current, value) => Math.max(current, value), max), 0)
      return <section key={chart.name} style={{ marginTop: 20 }}>
        <Typography.Title level={5}>{chart.name}</Typography.Title>
        <Typography.Paragraph>行：{chart.axes[0].name} · 列：{chart.axes[1].name} · 最大桶计数：{maximum.toLocaleString()}</Typography.Paragraph>
        <div style={{ overflow: 'auto', maxHeight: 520 }}>
          <table style={{ borderCollapse: 'collapse', fontSize: 12 }} aria-label={chart.name}>
            <thead><tr><th scope="col">{chart.axes[0].name} / {chart.axes[1].name}</th>{chart.axes[1].ranges.map((range, index) => <th scope="col" key={index} style={{ minWidth: 85, padding: 6 }}>{label(range)}</th>)}</tr></thead>
            <tbody>{chart.values.map((row, y) => <tr key={y}>
              <th scope="row" style={{ whiteSpace: 'nowrap', padding: 6 }}>{label(chart.axes[0].ranges[y])}</th>
              {row.map((value, x) => {
                const intensity = maximum ? Math.log1p(value) / Math.log1p(maximum) : 0
                return <td key={x} title={`${chart.axes[0].name}: ${label(chart.axes[0].ranges[y])}\n${chart.axes[1].name}: ${label(chart.axes[1].ranges[x])}\n计数: ${value}`} style={{ textAlign: 'right', padding: 6, border: '1px solid #d9d9d9', background: value ? `rgba(22,119,255,${0.08 + intensity * 0.82})` : 'transparent', color: intensity > 0.65 ? '#fff' : 'inherit' }}>{value.toLocaleString()}</td>
              })}
            </tr>)}</tbody>
          </table>
        </div>
      </section>
    })}
    <Collapse style={{ marginTop: 16 }} items={[{ key: 'raw', label: '原始直方图数据', children: <RecordDetail record={data} /> }]} />
  </>
}

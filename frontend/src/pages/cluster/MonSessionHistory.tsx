import { Alert, Button, Card, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import type { ApiRecord } from '../../api/client'
import { queryMetricRange, type MetricResponse } from '../../api/external'
import { AppTable } from '../../components/AppTable'
import { metricSamples } from '../monitoring/MetricPage'
import { MetricTrend } from '../monitoring/MetricTrend'

export function monSessionSeries(data: MetricResponse, name: string): ApiRecord[] {
  const fsid = data.meta?.cluster_fsid
  if (data.result_type !== 'matrix' || typeof fsid !== 'string' || !fsid || data.meta?.mon_name !== name || !Array.isArray(data.series)) throw new Error('MON 历史响应范围异常')
  if (data.series.some(row => {
    const labels = row?.metric as ApiRecord | undefined
    return !labels || labels.cluster !== fsid || labels.ceph_daemon !== `mon.${name}` || !Array.isArray(row.values)
  })) throw new Error('MON 历史序列标签或样本数组异常')
  return data.series
}

export function monSessionSamples(row: ApiRecord) {
  return metricSamples(row).map(sample => {
    if (sample.status !== '有效值') return sample
    const [mantissa, exponent = '0'] = sample.value.toLowerCase().split('e')
    const digits = mantissa.replace(/^[+-]/, '').replace('.', '')
    const fractionLength = mantissa.includes('.') ? mantissa.length - mantissa.indexOf('.') - 1 : 0
    const decimalPlaces = fractionLength - Number(exponent)
    const trailingZeros = /0*$/.exec(digits)![0].length
    const zero = /^0+$/.test(digits)
    const invalid = !zero && (mantissa.startsWith('-') || decimalPlaces > trailingZeros)
    return invalid ? { ...sample, status: '会话数异常（应为非负整数）' } : sample
  })
}

export function MonSessionHistory({ clusterId, monName }: { clusterId: number; monName: string }) {
  const [data, setData] = useState<MetricResponse | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const controller = useRef<AbortController | null>(null)
  useEffect(() => () => controller.current?.abort(), [])
  async function query() {
    controller.current?.abort()
    const active = new AbortController()
    controller.current = active
    setLoading(true); setError(''); setData(null)
    const end = new Date()
    try {
      const result = await queryMetricRange(clusterId, { metricId: 'mon_sessions', monName, start: new Date(end.getTime() - 3600000).toISOString(), end: end.toISOString(), step: '30s' }, { signal: active.signal, suppressErrorNotification: true })
      monSessionSeries(result, monName)
      if (!active.signal.aborted) setData(result)
    } catch (err) {
      if (!active.signal.aborted) setError(err instanceof Error ? err.message : 'MON 历史读取失败')
    } finally {
      if (!active.signal.aborted) setLoading(false)
    }
  }
  return <Card title="MON 会话数历史 · 最近一小时" extra={<Button loading={loading} onClick={() => void query()}>查询历史</Button>}>
    <Space direction="vertical" className="page-stack">
      <Alert type="info" message="使用 Prometheus 的 ceph_mon_num_sessions，非 MGR 内存历史或 CLI 快照。需配置 Prometheus 并采集该指标，cluster 标签必须为集群 FSID。按 MON 精确筛选，不合并不同采集来源；30 秒评估步长可能复用最近抓取值，缺失不补零。" />
      {error && <Alert type="warning" message={error} />}
      {!data && !error && <span>点击查询读取历史。</span>}
      {data?.series.length === 0 && <Alert type="info" message="没有匹配的历史序列，不表示会话数为零。" />}
      {data?.series.map((row, index) => <div key={index}>
        <pre style={{ whiteSpace: 'pre-wrap' }}>{JSON.stringify(row.metric, null, 2)}</pre>
        <MetricTrend samples={monSessionSamples(row)} meta={data.meta} name={`mon.${monName} 会话数`} />
        <AppTable dataSource={monSessionSamples(row)} rowKey="index" size="small" pagination={{ defaultPageSize: 5 }} columns={[{ title: '评估时间 UTC', dataIndex: 'utc' }, { title: '原始值', dataIndex: 'value' }, { title: '样本状态', dataIndex: 'status' }]} />
      </div>)}
    </Space>
  </Card>
}

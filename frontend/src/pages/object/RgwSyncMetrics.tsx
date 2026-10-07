import { Alert, Button, Card, Space, Table } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { queryMetricRange, type MetricResponse } from '../../api/external'
import { metricSamples } from '../monitoring/MetricPage'
import { MetricTrend } from '../monitoring/MetricTrend'
import { MetricNotices } from '../monitoring/MetricNotices'
import { rgwSyncMetrics } from './RgwSyncMetricLinks'

type Result = { response?: MetricResponse; error?: string }

export function RgwSyncMetrics({ clusterId }: { clusterId?: number }) {
  const current = useRef(clusterId), mounted = useRef(true), sequence = useRef(0), abort = useRef<AbortController>()
  current.current = clusterId
  const [state, setState] = useState<{ clusterId?: number; busy: boolean; rows?: Result[] }>({ busy: false })
  useEffect(() => {
    mounted.current = true
    setState({ clusterId, busy: false })
    return () => { mounted.current = false; abort.current?.abort(); sequence.current++ }
  }, [clusterId])
  async function read() {
    if (!clusterId || !mounted.current || current.current !== clusterId) return
    abort.current?.abort()
    const controller = new AbortController(), ticket = ++sequence.current
    abort.current = controller
    setState({ clusterId, busy: true })
    const end = new Date(), start = new Date(end.getTime() - 3600000)
    const rows = await Promise.all(rgwSyncMetrics.map(async metric => {
      try {
        const response = await queryMetricRange(clusterId, { metricId: metric.id, start: start.toISOString(), end: end.toISOString(), step: '60s' }, { signal: controller.signal, cache: 'no-store', suppressErrorNotification: true })
        if (response.result_type !== 'matrix') throw new Error('Invalid range result')
        return { response }
      } catch { return { error: '查询失败或响应无效，请检查监控权限、端点与采集配置。' } }
    }))
    if (mounted.current && current.current === clusterId && sequence.current === ticket && !controller.signal.aborted) setState({ clusterId, busy: false, rows })
  }
  const scoped = state.clusterId === clusterId ? state : undefined
  return <Card size="small" title="RGW 同步趋势（最近一小时）">
    <Space direction="vertical" style={{ width: '100%' }}>
      <Alert type="info" message="按需读取 Prometheus，60 秒评估步长" description="四类来源指标使用一分钟速率或加权平均延迟；分片时差为原始秒数。未按当前 Zone 过滤，同名来源 Zone 可能在共享端点内合并，请核对全部标签。时差仅由增量同步更新，不证明同步完成；无数据不代表零。" />
      <Button disabled={!clusterId || scoped?.busy} loading={scoped?.busy} onClick={() => void read()}>读取同步趋势总览</Button>
      {scoped?.rows?.map((row, index) => <Card size="small" key={rgwSyncMetrics[index].id} title={`${rgwSyncMetrics[index].title} · ${rgwSyncMetrics[index].unit}`}>
        {row.error && <Alert type="warning" message={row.error} />}
        <MetricNotices source={rgwSyncMetrics[index].title} meta={row.response?.meta} />
        {row.response?.series.length === 0 && <Alert type="info" message="无样本（不代表零或同步健康）" />}
        {row.response?.series.map((series, seriesIndex) => {
          const samples = metricSamples(series)
          return <div key={seriesIndex}>
            <pre style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{JSON.stringify(series.metric)}</pre>
            <MetricTrend name={`${rgwSyncMetrics[index].title} ${seriesIndex + 1}`} samples={samples} meta={row.response?.meta} />
            <details><summary>原始评估样本（{samples.length}）</summary><Table size="small" rowKey="index" pagination={{ pageSize: 10 }} dataSource={samples} columns={[
              { title: 'UTC 评估时间', dataIndex: 'utc' }, { title: '原始数值', dataIndex: 'value' }, { title: '状态', dataIndex: 'status' }
            ]} /></details>
          </div>
        })}
      </Card>)}
    </Space>
  </Card>
}

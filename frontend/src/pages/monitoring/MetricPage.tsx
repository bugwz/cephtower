import { BarChartOutlined, LineChartOutlined, ReloadOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Form, Input, Select, Segmented, Space, Statistic, Tag, Typography } from 'antd'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { queryMetric, queryMetricRange, type MetricResponse } from '../../api/external'
import type { ApiRecord } from '../../api/client'
import { AppTable } from '../../components/AppTable'
import { Page } from '../../components/Page'
import { useFeatureRequirements } from '../../hooks/useFeatureRequirements'
import { useClusterContext } from '../../state/ClusterContext'
import { MetricTrend } from './MetricTrend'

const { Text } = Typography

type MetricMode = 'instant' | 'range'

interface MetricFormValues {
  mode: MetricMode
  metric_id: string
  time?: string
  start?: string
  end?: string
  step?: string
}

interface MetricRow extends ApiRecord {
  row_id: string
  metric_name: string
  labels: string
  latest_value: string
  points: number
  samples: MetricSample[]
}

interface MetricSample extends ApiRecord {
  index: number
  timestamp: string
  utc: string
  value: string
  status: string
}

export function metricSamples(item: ApiRecord): MetricSample[] {
  const samples = Array.isArray(item.values) ? item.values : Array.isArray(item.value) ? [item.value] : []
  return samples.map((sample, index) => {
    const tuple = Array.isArray(sample) ? sample : []
    const timestamp = tuple[0]
    const value = tuple[1]
    const date = typeof timestamp === 'number' && Number.isFinite(timestamp) ? new Date(timestamp * 1000) : null
    const validTime = date !== null && Number.isFinite(date.getTime())
    const validValue = typeof value === 'string' && (/^[+-]?(?:\d+(?:\.\d*)?|\.\d+)(?:[eE][+-]?\d+)?$/.test(value) || ['NaN', '+Inf', '-Inf', 'Inf'].includes(value))
    return {
      index, timestamp: timestamp == null ? '未提供' : String(timestamp),
      utc: validTime ? date.toISOString() : '无效时间',
      value: typeof value === 'string' ? value : '未提供或格式异常',
      status: tuple.length !== 2 || !validTime || !validValue ? '格式异常' : Number.isFinite(Number(value)) ? '有效值' : '非有限值（不代表零）'
    }
  })
}

const metricOptions = [
  { label: 'SMB2 请求累计耗时速率（µs/s）', value: 'smb_request_duration_rate', description: '各操作 5 分钟累计耗时 rate，非单请求平均延迟' },
  { label: 'SMB 指标采集状态（各实例）', value: 'smb_metrics_status', description: 'smb_metrics_status：0 Down / 1 Up' },
  { label: 'SMB 会话数（各实例）', value: 'smb_sessions', description: 'smb_sessions_total' },
  { label: 'SMB 用户数（各实例）', value: 'smb_users', description: 'smb_users_total' },
  { label: 'SMB 共享活动（各实例）', value: 'smb_share_activity', description: 'smb_share_activity' },
  { label: 'SMB2 接收速率（B/s）', value: 'smb_in_bytes_rate', description: '各序列 5 分钟 rate' },
  { label: 'SMB2 发送速率（B/s）', value: 'smb_out_bytes_rate', description: '各序列 5 分钟 rate' },
  { label: 'SMB2 请求率（请求/s）', value: 'smb_request_rate', description: '各序列 5 分钟 rate' },
	{ label: 'RGW 请求率（请求/s）', value: 'rgw_request_rate', description: 'sum rate over 1m' },
	{ label: 'RGW GET 平均延迟（ms）', value: 'rgw_get_latency_ms', description: 'weighted latency over 1m' },
	{ label: 'RGW PUT 平均延迟（ms）', value: 'rgw_put_latency_ms', description: 'weighted latency over 1m' },
	{ label: 'RGW GET 带宽（B/s）', value: 'rgw_get_bytes_rate', description: 'sum rate over 1m' },
	{ label: 'RGW PUT 带宽（B/s）', value: 'rgw_put_bytes_rate', description: 'sum rate over 1m' },
  { label: '集群健康', value: 'cluster_health', description: 'ceph_health_status' },
  { label: '容量使用率', value: 'capacity_used_percent', description: 'used / total' },
  { label: '客户端读吞吐', value: 'client_read_bytes', description: 'pool read bytes rate' },
  { label: '客户端写吞吐', value: 'client_write_bytes', description: 'pool write bytes rate' }
]

export function MetricPage() {
  const { selectedClusterId } = useClusterContext()
  const [params] = useSearchParams()
  const initialMetric = metricPreset(params.get('metric'))
  return <MetricContent key={`${selectedClusterId ?? 'none'}:${initialMetric}`} selectedClusterId={selectedClusterId} initialMetric={initialMetric} />
}

export function metricPreset(value: string | null): string {
  return metricOptions.some(option => option.value === value) ? value! : 'cluster_health'
}

function MetricContent({ selectedClusterId, initialMetric }: { selectedClusterId?: number; initialMetric: string }) {
  const active = useRef(true)
  const pending = useRef<AbortController | null>(null)
  useEffect(() => {
    active.current = true
    return () => { active.current = false; pending.current?.abort() }
  }, [])
  const [form] = Form.useForm<MetricFormValues>()
  const [mode, setMode] = useState<MetricMode>('instant')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [result, setResult] = useState<MetricResponse | null>(null)
  const featureStatus = useFeatureRequirements(selectedClusterId, { requiredEndpoints: ['prometheus'] })
  const blocked = featureStatus.loading || featureStatus.blocked || Boolean(featureStatus.error)

  const initialValues = useMemo(() => {
    const end = new Date()
    const start = new Date(end.getTime() - 60 * 60 * 1000)
    return {
      mode: 'instant' as MetricMode,
      metric_id: initialMetric,
      time: end.toISOString(),
      start: start.toISOString(),
      end: end.toISOString(),
      step: '30s'
    }
  }, [initialMetric])

  const rows = useMemo(() => normalizeSeries(result?.series ?? []), [result])

  async function submit(values: MetricFormValues) {
    if (!active.current) return
    if (!selectedClusterId) {
      setError('请先选择集群')
      return
    }
    if (blocked) {
      setError('当前集群未配置或未启用 prometheus endpoint')
      return
    }
    pending.current?.abort()
    const controller = new AbortController()
    pending.current = controller
    const current = () => active.current && pending.current === controller && !controller.signal.aborted
    const init = { signal: controller.signal, suppressErrorNotification: true }
    setLoading(true)
    setError('')
    setResult(null)
    try {
      const payload = values.mode === 'range'
        ? await queryMetricRange(selectedClusterId, {
          metricId: values.metric_id,
          start: values.start ?? '',
          end: values.end ?? '',
          step: values.step ?? '30s'
        }, init)
        : await queryMetric(selectedClusterId, {
          metricId: values.metric_id,
          time: values.time
        }, init)
      if (current()) setResult(payload)
    } catch (err) {
      if (current()) setError(err instanceof Error && err.message ? err.message : '指标查询失败')
    } finally {
      if (current()) setLoading(false)
    }
  }

  function applyPreset(metricId: string) {
    if (blocked) {
      setError('当前集群未配置或未启用 prometheus endpoint')
      return
    }
    form.setFieldsValue({ metric_id: metricId })
    void form.submit()
  }

  return (
    <Page title="性能指标">
      <Alert type="info" showIcon message="SMB 指标需要 Prometheus 采集 SMB exporter；保留原始实例和操作标签，不汇总为集群总数。共享 endpoint 可能包含多个集群，请核对结果标签；无数据不代表零。速率使用固定 5 分钟窗口。" />
      <Space direction="vertical" size={16} className="page-stack">
        <div className="metrics-grid metric-preset-grid">
          {metricOptions.map((item) => (
            <Card key={item.value} className="metric-preset-card" hoverable onClick={() => applyPreset(item.value)}>
              <Statistic title={item.label} value={item.description} prefix={<LineChartOutlined />} />
              <Text type="secondary">{item.value}</Text>
            </Card>
          ))}
        </div>

        <Card
          className="page-surface-card"
          title="Prometheus 指标查询"
          extra={<Button icon={<ReloadOutlined />} loading={loading} disabled={!selectedClusterId || blocked} onClick={() => form.submit()}>查询</Button>}
        >
          <Space direction="vertical" size={16} className="page-stack">
            <FeatureRequirementAlert status={featureStatus} />
            {error ? <Alert type="error" showIcon message="查询失败" description={error} /> : null}
            <Form
              form={form}
              layout="vertical"
              initialValues={initialValues}
              onFinish={submit}
              onValuesChange={(changed) => {
                if (changed.mode) {
                  setMode(changed.mode)
                }
              }}
            >
              <div className="metric-query-grid">
                <Form.Item name="mode" label="查询模式">
                  <Segmented
                    options={[
                      { label: '即时', value: 'instant', icon: <BarChartOutlined /> },
                      { label: '范围', value: 'range', icon: <LineChartOutlined /> }
                    ]}
                  />
                </Form.Item>
                <Form.Item name="metric_id" label="指标" rules={[{ required: true }]}>
                  <Select options={metricOptions} optionRender={(option) => (
                    <Space direction="vertical" size={0}>
                      <Text>{option.label}</Text>
                      <Text type="secondary">{option.value}</Text>
                    </Space>
                  )} />
                </Form.Item>
                {mode === 'instant' ? (
                  <Form.Item name="time" label="查询时间">
                    <Input />
                  </Form.Item>
                ) : (
                  <>
                    <Form.Item name="start" label="开始时间" rules={[{ required: true }]}>
                      <Input />
                    </Form.Item>
                    <Form.Item name="end" label="结束时间" rules={[{ required: true }]}>
                      <Input />
                    </Form.Item>
                    <Form.Item name="step" label="步长" rules={[{ required: true }]}>
                      <Input />
                    </Form.Item>
                  </>
                )}
              </div>
            </Form>

            <Alert type="info" message="RGW 指标使用一分钟速率窗口并聚合监控端点内的所有匹配序列。端点应仅包含当前集群且避免重复采集；无样本或 NaN 不代表零值。" />
            <Space wrap>
              <Tag color="blue">result_type: {result?.result_type ?? '-'}</Tag>
              <Tag>series: {rows.length}</Tag>
              <Tag>points: {rows.reduce((sum, row) => sum + row.points, 0)}</Tag>
            </Space>

            <AppTable<MetricRow>
              size="middle"
              rowKey="row_id"
              loading={loading}
              dataSource={rows}
              expandable={{ expandedRowRender: (row) => <Space direction="vertical" style={{ width: '100%' }}>
                {result?.result_type === 'matrix' && <MetricTrend samples={row.samples} meta={result.meta} name={row.metric_name} />}
                <Text style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{row.labels}</Text>
                <Text type="secondary">按接口返回顺序展示全部样本，时间以 UTC 显示；原始值不做浮点转换。非有限值不代表零。</Text>
                <AppTable<MetricSample> size="small" rowKey="index" dataSource={row.samples} pagination={{ defaultPageSize: 20, showSizeChanger: true }} columns={[
                  { title: '时间（UTC）', dataIndex: 'utc' },
                  { title: '原始时间戳（秒）', dataIndex: 'timestamp' },
                  { title: '原始值', dataIndex: 'value' },
                  { title: '样本状态', dataIndex: 'status' }
                ]} />
              </Space> }}
              pagination={{ defaultPageSize: 10, showSizeChanger: true }}
              scroll={{ x: 980 }}
              columns={[
                { title: '指标', dataIndex: 'metric_name', width: 220, ellipsis: true },
                { title: 'Labels', dataIndex: 'labels', ellipsis: true },
                { title: '最新值', dataIndex: 'latest_value', width: 160 },
                { title: '点数', dataIndex: 'points', width: 90 }
              ]}
            />
          </Space>
        </Card>
      </Space>
    </Page>
  )
}

function normalizeSeries(series: ApiRecord[]): MetricRow[] {
  return series.map((item, index) => {
    const metric = readRecord(item.metric)
    const samples = metricSamples(item)
    return {
      row_id: `${index}-${JSON.stringify(metric)}`,
      metric_name: String(metric.__name__ ?? metric.job ?? metric.instance ?? `series-${index + 1}`),
      labels: JSON.stringify(metric),
      latest_value: samples[samples.length - 1]?.value ?? '-',
      points: samples.length,
      samples
    }
  })
}

function readRecord(value: unknown): ApiRecord {
  return typeof value === 'object' && value !== null && !Array.isArray(value) ? value as ApiRecord : {}
}

function FeatureRequirementAlert({ status }: { status: ReturnType<typeof useFeatureRequirements> }) {
  if (status.loading) {
    return <Alert type="info" showIcon message="正在校验当前集群的功能依赖" />
  }
  if (status.error) {
    return <Alert type="warning" showIcon message="功能依赖检查失败" description={status.error} />
  }
  if (status.reasons.length) {
    return <Alert type="warning" showIcon message="当前集群暂不可执行该页面的查询操作" description={status.reasons.join('; ')} />
  }
  return null
}

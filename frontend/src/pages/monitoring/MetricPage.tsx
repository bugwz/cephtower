import { BarChartOutlined, LineChartOutlined, ReloadOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Form, Input, Select, Segmented, Space, Statistic, Tag, Typography } from 'antd'
import { useMemo, useState } from 'react'
import { queryMetric, queryMetricRange, type MetricResponse } from '../../api/external'
import type { ApiRecord } from '../../api/client'
import { AppTable } from '../../components/AppTable'
import { Page } from '../../components/Page'
import { useFeatureRequirements } from '../../hooks/useFeatureRequirements'
import { useClusterContext } from '../../state/ClusterContext'

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
}

const metricOptions = [
  { label: '集群健康', value: 'cluster_health', description: 'ceph_health_status' },
  { label: '容量使用率', value: 'capacity_used_percent', description: 'used / total' },
  { label: '客户端读吞吐', value: 'client_read_bytes', description: 'pool read bytes rate' },
  { label: '客户端写吞吐', value: 'client_write_bytes', description: 'pool write bytes rate' }
]

export function MetricPage() {
  const { selectedClusterId } = useClusterContext()
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
      metric_id: 'cluster_health',
      time: end.toISOString(),
      start: start.toISOString(),
      end: end.toISOString(),
      step: '30s'
    }
  }, [])

  const rows = useMemo(() => normalizeSeries(result?.series ?? []), [result])

  async function submit(values: MetricFormValues) {
    if (!selectedClusterId) {
      setError('请先选择集群')
      return
    }
    if (blocked) {
      setError('当前集群未配置或未启用 prometheus endpoint')
      return
    }
    setLoading(true)
    setError('')
    try {
      const payload = values.mode === 'range'
        ? await queryMetricRange(selectedClusterId, {
          metricId: values.metric_id,
          start: values.start ?? '',
          end: values.end ?? '',
          step: values.step ?? '30s'
        })
        : await queryMetric(selectedClusterId, {
          metricId: values.metric_id,
          time: values.time
        })
      setResult(payload)
    } catch (err) {
      setError(err instanceof Error ? err.message : '指标查询失败')
    } finally {
      setLoading(false)
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
    const values = Array.isArray(item.values) ? item.values : undefined
    const value = Array.isArray(item.value) ? item.value : undefined
    const latest = values?.[values.length - 1] ?? value
    return {
      row_id: `${index}-${JSON.stringify(metric)}`,
      metric_name: String(metric.__name__ ?? metric.job ?? metric.instance ?? `series-${index + 1}`),
      labels: JSON.stringify(metric),
      latest_value: Array.isArray(latest) ? String(latest[1] ?? latest[0] ?? '-') : '-',
      points: values?.length ?? (value ? 1 : 0)
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

import { BarChartOutlined, LineChartOutlined, ReloadOutlined } from '@ant-design/icons'
import { Alert, AutoComplete, Button, Card, Form, Input, Select, Segmented, Space, Statistic, Tag, Typography } from 'antd'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { queryMetric, queryMetricRange, type MetricResponse } from '../../api/external'
import type { ApiRecord } from '../../api/client'
import { AppTable } from '../../components/AppTable'
import { Page } from '../../components/Page'
import { useFeatureRequirements } from '../../hooks/useFeatureRequirements'
import { useClusterContext } from '../../state/ClusterContext'
import { MetricTrend } from './MetricTrend'
import { MetricNotices } from './MetricNotices'

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
  { label: 'RGW 按用户 GET 累计操作数', value: 'rgw_user_get_ops_total', description: '保留 user、tenant 与实例标签；缓存计数，非审计记录' },
  { label: 'RGW 按用户 PUT 累计操作数', value: 'rgw_user_put_ops_total', description: '保留 user、tenant 与实例标签；缓存计数，非审计记录' },
  { label: 'RGW 按用户 DELETE 累计操作数', value: 'rgw_user_delete_ops_total', description: '保留 user、tenant 与实例标签；缓存计数，非审计记录' },
  { label: 'RGW 按用户 COPY 累计操作数', value: 'rgw_user_copy_ops_total', description: '保留 user、tenant 与实例标签；缓存计数，非审计记录' },
  { label: 'RGW 按用户 列举对象 累计操作数', value: 'rgw_user_list_ops_total', description: '保留 user、tenant 与实例标签；缓存计数，非审计记录' },
  { label: 'RGW 按用户 GET 累计字节（B）', value: 'rgw_user_get_bytes_total', description: '缓存累计操作字节，非用户容量或带宽' },
  { label: 'RGW 按用户 PUT 累计字节（B）', value: 'rgw_user_put_bytes_total', description: '缓存累计操作字节，非用户容量或带宽' },
  { label: 'RGW 按用户 DELETE 累计字节（B）', value: 'rgw_user_delete_bytes_total', description: '缓存累计操作字节，非用户容量或带宽' },
  { label: 'RGW 按用户 COPY 累计字节（B）', value: 'rgw_user_copy_bytes_total', description: '缓存累计操作字节，非用户容量或带宽' },
  { label: 'RGW 按桶 GET 累计操作数', value: 'rgw_bucket_get_ops_total', description: '保留 bucket、tenant 与实例标签；缓存计数，非库存' },
  { label: 'RGW 按桶 PUT 累计操作数', value: 'rgw_bucket_put_ops_total', description: '保留 bucket、tenant 与实例标签；缓存计数，非库存' },
  { label: 'RGW 按桶 DELETE 累计操作数', value: 'rgw_bucket_delete_ops_total', description: '保留 bucket、tenant 与实例标签；缓存计数，非库存' },
  { label: 'RGW 按桶 COPY 累计操作数', value: 'rgw_bucket_copy_ops_total', description: '保留 bucket、tenant 与实例标签；缓存计数，非库存' },
  { label: 'RGW 按桶 列举对象 累计操作数', value: 'rgw_bucket_list_ops_total', description: '保留 bucket、tenant 与实例标签；缓存计数，非库存' },
  { label: 'RGW 按桶 GET 累计字节（B）', value: 'rgw_bucket_get_bytes_total', description: '缓存累计操作字节，非带宽、桶容量或回收空间' },
  { label: 'RGW 按桶 PUT 累计字节（B）', value: 'rgw_bucket_put_bytes_total', description: '缓存累计操作字节，非带宽、桶容量或回收空间' },
  { label: 'RGW 按桶 DELETE 累计字节（B）', value: 'rgw_bucket_delete_bytes_total', description: '缓存累计操作字节，非带宽、桶容量或回收空间' },
  { label: 'RGW 按桶 COPY 累计字节（B）', value: 'rgw_bucket_copy_bytes_total', description: '缓存累计操作字节，非带宽、桶容量或回收空间' },
  { label: 'RGW GET 累计字节（B）', value: 'rgw_get_bytes_total', description: '各原生序列累计操作字节，非 B/s 或桶容量' },
  { label: 'RGW PUT 累计字节（B）', value: 'rgw_put_bytes_total', description: '各原生序列累计操作字节，非 B/s 或桶容量' },
  { label: 'RGW COPY 累计字节（B）', value: 'rgw_copy_bytes_total', description: '累计复制对象字节，不代表网络流量或当前容量' },
  { label: 'RGW DELETE 对象累计字节（B）', value: 'rgw_delete_bytes_total', description: '累计删除操作对象字节，不代表实际回收物理空间' },
  { label: 'RGW DELETE 对象平均延迟（ms）', value: 'rgw_delete_latency_ms', description: '一分钟操作数加权平均；不是累计耗时' },
  { label: 'RGW COPY 平均延迟（ms）', value: 'rgw_copy_latency_ms', description: '一分钟操作数加权平均；不是累计耗时' },
  { label: 'RGW 列举对象平均延迟（ms）', value: 'rgw_list_objects_latency_ms', description: '一分钟操作数加权平均；不是累计耗时' },
  { label: 'RGW 列举桶平均延迟（ms）', value: 'rgw_list_buckets_latency_ms', description: '一分钟操作数加权平均；不是累计耗时' },
  { label: 'RGW 删除桶平均延迟（ms）', value: 'rgw_delete_buckets_latency_ms', description: '一分钟操作数加权平均；不是累计耗时' },
  { label: 'RGW GET 累计操作数', value: 'rgw_get_ops_total', description: '各原生序列累计计数，不是当前对象数或请求/s' },
  { label: 'RGW PUT 累计操作数', value: 'rgw_put_ops_total', description: '各原生序列累计计数，不是当前对象数或请求/s' },
  { label: 'RGW DELETE 对象累计操作数', value: 'rgw_delete_ops_total', description: '各原生序列累计计数，重启或计数器重置可下降' },
  { label: 'RGW COPY 累计操作数', value: 'rgw_copy_ops_total', description: '各原生序列累计计数，重启或计数器重置可下降' },
  { label: 'RGW 列举对象累计操作数', value: 'rgw_list_objects_total', description: '各原生序列累计计数，不是列出的对象数量' },
  { label: 'RGW 列举桶累计操作数', value: 'rgw_list_buckets_total', description: '各原生序列累计计数，不是列出的桶数量' },
  { label: 'RGW 删除桶累计操作数', value: 'rgw_delete_buckets_total', description: '各原生序列累计计数，重启或计数器重置可下降' },
  { label: 'RGW 分片同步时差（秒）', value: 'rgw_sync_delta_seconds', description: '原始 gauge，保留分片和源/本地 Zone 标签；不是 rate 或整体同步完成度' },
  { label: 'RGW 来源 Zone 轮询平均延迟（ms）', value: 'rgw_sync_poll_latency_ms', description: '按 source_zone：一分钟 sum rate / count rate × 1000；无请求时可能为 NaN' },
  { label: 'RGW 来源 Zone 复制吞吐率（B/s）', value: 'rgw_sync_bytes_rate', description: '按 source_zone 聚合，一分钟 rate' },
  { label: 'RGW 来源 Zone 复制对象速率（对象/s）', value: 'rgw_sync_objects_rate', description: '按 source_zone 聚合，一分钟 rate' },
  { label: 'RGW 来源 Zone 复制失败速率（次/s）', value: 'rgw_sync_errors_rate', description: '按 source_zone 聚合，一分钟 rate；非累计失败数' },
  { label: 'SMB 集群节点指标数', value: 'smb_cluster_nodes', description: '按 netbiosname 统计匹配的会话指标序列，不是编排器主机库存' },
  { label: 'SMB 集群会话均值', value: 'smb_cluster_sessions_mean', description: '按 netbiosname：会话数 × 采集状态的序列均值，非总数' },
  { label: 'SMB 集群用户均值', value: 'smb_cluster_users_mean', description: '按 netbiosname：用户数 × 采集状态的序列均值，非总数' },
  { label: 'SMB 集群共享活动均值', value: 'smb_cluster_shares_mean', description: '按 netbiosname：共享活动 × 采集状态的序列均值，非总数' },
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

export function filterMetricRows<T extends { labels: string }>(rows: T[], key: string, value: string): T[] {
  if (!key) return rows
  return rows.filter(row => {
    try {
      const labels = JSON.parse(row.labels)
      return labels !== null && typeof labels === 'object' && !Array.isArray(labels) && Object.prototype.hasOwnProperty.call(labels, key) && labels[key] === value
    } catch { return false }
  })
}

export function metricLabelOptions(rows: { labels: string }[], key?: string) {
  const values = new Set<string>()
  for (const row of rows) {
    try {
      const labels = JSON.parse(row.labels)
      if (!labels || typeof labels !== 'object' || Array.isArray(labels)) continue
      for (const [name, value] of Object.entries(labels)) {
        if (typeof value !== 'string') continue
        if (key === undefined) values.add(name)
        else if (name === key) values.add(value)
      }
    } catch { /* Malformed labels do not supply suggestions. */ }
  }
  return [...values].sort().map(value => ({ value, label: value === '' ? '空字符串' : value }))
}

function MetricContent({ selectedClusterId, initialMetric }: { selectedClusterId?: number; initialMetric: string }) {
  const active = useRef(true)
  const pending = useRef<AbortController | null>(null)
  const queriedMetric = useRef<string | undefined>(undefined)
  useEffect(() => {
    active.current = true
    return () => { active.current = false; pending.current?.abort() }
  }, [])
  const [form] = Form.useForm<MetricFormValues>()
  const [mode, setMode] = useState<MetricMode>('instant')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [result, setResult] = useState<MetricResponse | null>(null)
  const [labelKey, setLabelKey] = useState('')
  const [labelValue, setLabelValue] = useState('')
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
  const visibleRows = useMemo(() => filterMetricRows(rows, labelKey, labelValue), [rows, labelKey, labelValue])
  const labelNames = useMemo(() => metricLabelOptions(rows), [rows])
  const labelValues = useMemo(() => metricLabelOptions(rows, labelKey), [rows, labelKey])

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
    if (queriedMetric.current !== values.metric_id) {
      setLabelKey('')
      setLabelValue('')
      queriedMetric.current = values.metric_id
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
      <Alert type="info" showIcon message="SMB 指标需要 Prometheus 采集 SMB exporter。各实例指标保留原始标签；集群指标按 netbiosname 聚合，均值不是总数。关联要求每个 instance 对应唯一状态序列，重复采集会导致查询失败；共享 endpoint 的同名 SMB 集群可能被合并，请核对部署范围。无数据不代表零，速率使用固定 5 分钟窗口。" />
      <Alert type="info" showIcon message="RGW 累计操作数保留原始实例标签，未跨实例求和；它们不是库存对象数、请求速率或所选时间段内的操作总数。重启或计数器重置会使数值下降，共享端点不提供集群隔离。" />
      <Alert type="info" showIcon message="按桶指标依赖 rgw_bucket_counters_cache 与 exporter 采集。缓存淘汰、重建或重启可能重置计数，不能作为持久审计记录；无数据不代表没有请求。请同时核对 bucket、tenant 和实例标签，不能只按桶名判断归属。" />
      <Alert type="info" showIcon message="按用户指标依赖 rgw_user_counters_cache 与 exporter 采集，同样可能因缓存淘汰而重置。user 标签是原生用户 ID 部分，必须结合 tenant 与实例核对；不等同于账户统计，也不保证包含匿名请求。" />
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

            <Alert type="info" message="RGW 速率和平均延迟使用一分钟窗口，来源 Zone 指标按 source_zone 聚合；共享端点的同名来源 Zone 会被合并，不代表目的 Zone 隔离。分片同步时差保留原始 gauge 和标签，不计算 rate，仅在原生增量同步更新时产生，不能据此断言当前已同步完成。端点应仅包含当前集群且避免重复采集；无样本或 NaN 不代表零值。" />
            <Space wrap>
              <Tag color="blue">result_type: {result?.result_type ?? '-'}</Tag>
              <Tag>series: {rows.length}</Tag>
              <Tag>points: {rows.reduce((sum, row) => sum + row.points, 0)}</Tag>
              <Tag>显示序列: {visibleRows.length}</Tag>
            </Space>
            <MetricNotices meta={result?.meta} />
            <Space wrap>
              <AutoComplete aria-label="精确标签名" style={{ minWidth: 260 }} options={labelNames} placeholder="标签名，例如 instance / operation" value={labelKey} onChange={value => { setLabelKey(value); setLabelValue('') }} />
              <AutoComplete aria-label="精确标签值" style={{ minWidth: 260 }} options={labelValues} placeholder="标签值（精确匹配，区分大小写）" value={labelValue} onChange={setLabelValue} />
              <Button onClick={() => { setLabelKey(''); setLabelValue('') }}>清除标签筛选</Button>
              <Text type="secondary">仅筛选已返回的序列，不改变查询或提供集群隔离；空标签名显示全部，空标签值匹配已存在的空值。</Text>
            </Space>

            <AppTable<MetricRow>
              size="middle"
              rowKey="row_id"
              loading={loading}
              dataSource={visibleRows}
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

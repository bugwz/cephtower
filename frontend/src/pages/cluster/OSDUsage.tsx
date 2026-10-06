import { Alert, Descriptions, Space } from 'antd'
import type { ApiRecord } from '../../api/client'

export function osdUsageInteger(value: unknown): string {
  return typeof value === 'string' && /^(0|[1-9]\d*)$/.test(value) && BigInt(value) <= 18446744073709551615n ? value : '未采集或格式无效'
}

export function OSDUsage({ record }: { record: ApiRecord }) {
  const stats = record.stats && typeof record.stats === 'object' && !Array.isArray(record.stats) ? record.stats as ApiRecord : {}
  const utilization = stats.utilization
  const perf = record.perf_stats && typeof record.perf_stats === 'object' && !Array.isArray(record.perf_stats) ? record.perf_stats as ApiRecord : {}
  return <Space direction="vertical" className="page-stack">
    <Alert type="info" message="来自 ceph osd df 的单 OSD 快照。容量单位为原生 KiB，不与集群总容量混用；PG 数不是对象数。" />
    {record.stale !== false && <Alert type="warning" message="库存已过期或新鲜度未知，请重新采集 OSD。" />}
    <Descriptions bordered column={2}>
      {[
        ['kb', '总容量（KiB）'], ['kb_used', '已用（KiB）'], ['kb_avail', '可用（KiB）'],
        ['kb_used_data', '数据（KiB）'], ['kb_used_omap', 'OMAP（KiB）'], ['kb_used_meta', '元数据（KiB）'], ['pgs', 'PG 数']
      ].map(([key, label]) => <Descriptions.Item key={key} label={label}>{osdUsageInteger(stats[key])}</Descriptions.Item>)}
      <Descriptions.Item label="利用率">{typeof utilization === 'number' && Number.isFinite(utilization) && utilization >= 0 && utilization <= 100 ? `${utilization}%` : '未采集或格式无效'}</Descriptions.Item>
    </Descriptions>
    <Alert type="info" message="以下延迟来自 ceph osd perf 的原生快照，单位毫秒；不是 Prometheus 一分钟平均值，也不等同于客户端端到端延迟。" />
    <Descriptions bordered column={2}>
      <Descriptions.Item label="提交延迟（ms）">{osdLatency(perf.commit_latency_ms)}</Descriptions.Item>
      <Descriptions.Item label="应用延迟（ms）">{osdLatency(perf.apply_latency_ms)}</Descriptions.Item>
    </Descriptions>
  </Space>
}

export function osdLatency(value: unknown): string {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0 ? String(value) : '未采集或格式无效'
}

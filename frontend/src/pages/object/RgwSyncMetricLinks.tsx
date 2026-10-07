import { Alert, Button, Space } from 'antd'
import { useNavigate } from 'react-router-dom'

export const rgwSyncMetrics = [
  { id: 'rgw_sync_bytes_rate', title: '复制吞吐率', unit: 'B/s' },
  { id: 'rgw_sync_objects_rate', title: '复制对象速率', unit: '对象/s' },
  { id: 'rgw_sync_errors_rate', title: '复制失败速率', unit: '次/s' },
  { id: 'rgw_sync_poll_latency_ms', title: '轮询平均延迟', unit: 'ms' },
  { id: 'rgw_sync_delta_seconds', title: '分片同步时差', unit: '秒' }
]

export function RgwSyncMetricLinks({ clusterId }: { clusterId?: number }) {
  const navigate = useNavigate()
  return <Space direction="vertical">
    <Alert type="info" message="同步性能与历史指标" description="进入当前集群的指标查询页后手动查询，需要监控页面权限和 Prometheus 采集。入口不自动限定为此 Zone；请核对返回的来源 Zone、目的 Zone（如有）和分片标签。原生同步报告与监控指标采样时间不同，均不证明同步已完成。" />
    <Space wrap>{rgwSyncMetrics.map(metric => <Button key={metric.id} disabled={!clusterId} onClick={() => navigate(`/monitoring/metric?metric=${metric.id}`)}>{metric.title}</Button>)}</Space>
  </Space>
}

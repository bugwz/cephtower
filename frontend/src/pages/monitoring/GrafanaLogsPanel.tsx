import { Alert, Button, Card, Space } from 'antd'
import { useCallback } from 'react'
import { readExternalList } from '../../api/external'
import { isRecord } from '../../api/client'
import { useResource } from '../../hooks'
import { useClusterContext } from '../../state/ClusterContext'

export function safeGrafanaLogsURL(value: unknown): string | undefined {
  if (typeof value !== 'string' || !value) return undefined
  try {
    const url = new URL(value)
    return ['https:', 'http:'].includes(url.protocol) && !url.username && !url.password ? url.href : undefined
  } catch { return undefined }
}

export function GrafanaLogsPanel() {
  const { selectedClusterId } = useClusterContext()
  const loader = useCallback(async () => {
    if (!selectedClusterId) return undefined
    const result = await readExternalList('/grafana', selectedClusterId)
    const url = safeGrafanaLogsURL(isRecord(result.meta) ? result.meta.logs_explore_url : undefined)
    if (!url) throw new Error('Grafana 未提供有效的日志入口')
    return url
  }, [selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  return <Card title="守护进程历史日志（Grafana / Loki）">
    <Space direction="vertical">
      <Alert type="info" showIcon message="需要已部署 Loki、Promtail，并在 Grafana 配置 Loki 数据源。历史日志由外部系统保存，不来自 MON 日志缓冲区。" />
      <Alert type="warning" showIcon message="入口使用当前集群配置的 Grafana 地址，但共享 Loki 数据源不保证按集群隔离。打开后请自行核对标签和查询范围；默认不执行日志查询。浏览器可能需要单独登录 Grafana，服务端凭据不会传递。" />
      {!selectedClusterId && <Alert type="info" message="请先选择集群" />}
      {error && <Alert type="error" message={error} />}
      <Space>
        <Button disabled={!selectedClusterId} loading={loading} onClick={() => void refresh()}>重新读取入口</Button>
        <Button type="primary" href={selectedClusterId && !loading && !error ? data ?? undefined : undefined} disabled={!selectedClusterId || loading || Boolean(error) || !data} target="_blank" rel="noopener noreferrer">打开 Grafana Explore</Button>
      </Space>
    </Space>
  </Card>
}

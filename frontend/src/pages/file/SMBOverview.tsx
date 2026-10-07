import { Alert, Button, Card, Space } from 'antd'
import { useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { readExternalList } from '../../api/external'
import { isRecord } from '../../api/client'
import { useResource } from '../../hooks'
import { useClusterContext } from '../../state/ClusterContext'

export function smbOverviewURL(value: unknown): string | undefined {
  if (typeof value !== 'string') return undefined
  try {
    const url = new URL(value)
    if (['http:', 'https:'].includes(url.protocol) && !url.username && !url.password) return url.href
  } catch { /* Invalid external links remain disabled. */ }
  return undefined
}

export function SMBOverview() {
  const navigate = useNavigate()
  const { selectedClusterId } = useClusterContext()
  const loader = useCallback(async () => {
    if (!selectedClusterId) return undefined
    const result = await readExternalList('/grafana', selectedClusterId)
    const url = smbOverviewURL(isRecord(result.meta) ? result.meta.smb_overview_url : undefined)
    if (!url) throw new Error('未获取有效的 SMB 监控入口')
    return url
  }, [selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const ready = Boolean(selectedClusterId && !loading && !error && data)
  return <Card title="SMB 监控总览">
    <Space direction="vertical">
      <Alert type="info" message="使用当前集群配置的 Grafana。需要部署 Ceph SMB 看板及指标采集；此入口不表示看板已安装或指标正常。" />
      <Alert type="warning" message="共享监控服务不保证按集群隔离，请在看板中核对集群、实例和时间范围。浏览器可能需要单独登录，服务端凭据不会传递。" />
      {!selectedClusterId && <Alert type="info" message="请先选择集群" />}
      {error && <Alert type="error" message={error} />}
      <Space>
        <Button disabled={!selectedClusterId} loading={loading} onClick={() => void refresh()}>重新读取监控入口</Button>
        <Button disabled={!ready} href={ready ? data ?? undefined : undefined} target="_blank" rel="noopener noreferrer">打开 SMB 看板</Button>
        <Button disabled={!selectedClusterId} onClick={() => navigate('/monitoring/metric?metric=smb_metrics_status')}>本项目内查询 SMB 指标</Button>
      </Space>
    </Space>
  </Card>
}

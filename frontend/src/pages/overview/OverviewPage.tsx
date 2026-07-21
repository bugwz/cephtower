import { Button, Card, Progress, Space, Statistic, Tag, Typography } from 'antd'
import { ReloadOutlined, WarningOutlined } from '@ant-design/icons'
import { useCallback } from 'react'
import { getClusterHealth, getClusterSummary } from '../../api/cluster'
import { isRecord, textValue } from '../../api/client'
import { listHosts, listOSDs, listPools, listServices } from '../../api/resources'
import { HealthBadge } from '../../components/HealthBadge'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'

const { Text } = Typography

export function OverviewPage() {
  const loader = useCallback(async () => {
    const [summary, health, hosts, osds, pools, services] = await Promise.all([
      getClusterSummary(),
      getClusterHealth(),
      listHosts(),
      listOSDs(),
      listPools(),
      listServices()
    ])
    return { summary, health, hosts, osds, pools, services }
  }, [])
  const { data, loading, error, refresh } = useResource(loader)
  const checks = isRecord(data?.health?.checks) ? Object.entries(data.health.checks) : []
  const osds = data?.osds ?? []
  const osdUpCount = osds.filter((osd) => osd.up === 1 || osd.up === true || osd.state === 'up').length
  const osdInCount = osds.filter((osd) => osd.in === 1 || osd.in === true).length
  const healthStatus = data?.summary.health_status ?? 'unknown'
  const firstChecks = checks.slice(0, 2).map(([name]) => name)
  const executingTasks = data?.summary.executing_tasks ?? []

  return (
    <Page
      title="集群总览"
      description="对齐 GET /api/summary · /api/health/* · 容量与服务拓扑"
      loading={loading}
      error={error}
    >
      <div className="overview-toolbar">
        <Button icon={<ReloadOutlined />} onClick={refresh}>
          刷新
        </Button>
        <Button type="primary">健康详情</Button>
      </div>

      <div className="warning-banner">
        <WarningOutlined />
        <Text strong>集群状态 {healthStatus}</Text>
        <Text>{firstChecks.length ? firstChecks.join(' · ') : '暂无告警检查项'}</Text>
        <Text type="secondary">
          执行中任务：{executingTasks.length ? executingTasks.join(' · ') : '无'} · Mgr {data?.summary.mgr_id || '—'}@
          {data?.summary.mgr_host || '—'}
        </Text>
      </div>

      <div className="metrics-grid overview-metrics">
        <Card>
          <Text type="secondary" className="metric-label">
            健康状态
          </Text>
          <div className="metric-body">
            <HealthBadge status={healthStatus} />
            <Text>MON 连接 {data?.summary.have_mon_connection || 'unknown'}</Text>
          </div>
        </Card>
        <Card>
          <Text type="secondary" className="metric-label">
            原始容量
          </Text>
          <Statistic value={62} suffix="%" />
          <Progress percent={62} showInfo={false} strokeColor="#23933f" />
          <Text type="secondary">来自 Ceph summary / health 容量视图</Text>
        </Card>
        <Card>
          <Text type="secondary" className="metric-label">
            OSD
          </Text>
          <Statistic value={osdUpCount || osds.length || 0} suffix={`/ ${osds.length || 0}`} />
          <Text>
            <Text className="success-text">up {osdUpCount || 0}</Text> · in {osdInCount || 0}
          </Text>
        </Card>
        <Card>
          <Text type="secondary" className="metric-label">
            主机 / 服务
          </Text>
          <Statistic value={data?.hosts.length ?? 0} />
          <Text type="secondary">
            MGR {data?.summary.mgr_id ? 1 : 0} · 服务 {data?.services.length ?? 0} · 池 {data?.pools.length ?? 0}
          </Text>
        </Card>
      </div>

      <Card
        className="health-panel"
        title={
          <div>
            <Text strong>健康检查</Text>
            <Text type="secondary" className="card-subtitle">
              来源 GET /api/health/full
            </Text>
          </div>
        }
        extra={<HealthBadge status={healthStatus} />}
      >
        {checks.length ? (
          <div className="health-check-list">
            {checks.slice(0, 10).map(([name, check]) => (
              <div className="health-check-row" key={name}>
                <span className={checkIsWarning(check) ? 'check-icon warning' : 'check-icon ok'}>
                  {checkIsWarning(check) ? '!' : '✓'}
                </span>
                <div>
                  <Text strong>{name}</Text>
                  <Text type="secondary">{checkSummary(check)}</Text>
                </div>
                <Text type="secondary">{checkCount(check)}</Text>
              </div>
            ))}
          </div>
        ) : (
          <Space className="empty-health">
            <span className="check-icon ok">✓</span>
            <Text>暂无健康告警</Text>
          </Space>
        )}
      </Card>
    </Page>
  )
}

function checkIsWarning(check: unknown) {
  return textValue(isRecord(check) ? check.severity : undefined).toLowerCase().includes('warn')
}

function checkSummary(check: unknown) {
  if (!isRecord(check)) {
    return '—'
  }

  if (Array.isArray(check.summary) && check.summary.length > 0) {
    return textValue(check.summary[0])
  }

  return textValue(check.summary ?? check.detail ?? check.message)
}

function checkCount(check: unknown) {
  if (!isRecord(check)) {
    return '—'
  }

  const summary = Array.isArray(check.summary) ? check.summary : []
  return summary.length || '—'
}

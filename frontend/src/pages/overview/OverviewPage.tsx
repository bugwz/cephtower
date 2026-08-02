import {
  ApiOutlined,
  DatabaseOutlined,
  HddOutlined,
  ReloadOutlined,
  SafetyCertificateOutlined,
  ThunderboltOutlined
} from '@ant-design/icons'
import { Button, Card, Descriptions, Progress, Space, Tag, Typography } from 'antd'
import { useCallback, useMemo, useState } from 'react'
import { listClusterCapabilities, type ClusterCapability } from '../../api/cluster'
import { getOptionalResource, listResource, mutateResource, refreshResource } from '../../api/resource'
import { numberValue, textValue, type ApiRecord } from '../../api/client'
import { AppTable } from '../../components/AppTable'
import { HealthBadge } from '../../components/HealthBadge'
import { Page } from '../../components/Page'
import { TableAction } from '../../components/TableActions'
import { useResource } from '../../hooks'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useResourceTableFilters } from '../../hooks/useResourceTableFilters'
import { useClusterContext } from '../../state/ClusterContext'
import { message } from '../../utils/appMessage'

const { Text } = Typography

interface OverviewData {
  overview: ApiRecord
  observedAt?: string | null
  stale?: boolean
  staleReason?: string | null
  healthChecks: ApiRecord[]
  capabilities: ClusterCapability[]
}

export function OverviewPage() {
  const { selectedClusterId } = useClusterContext()
  const [refreshing, setRefreshing] = useState(false)
  const operationMutation = useMutationOperation()
  const healthTableFilters = useResourceTableFilters({
    path: '/health',
    fields: ['code', 'severity', 'summary', 'count'],
    clusterId: selectedClusterId
  })
  const loader = useCallback(async (): Promise<OverviewData> => {
    if (!selectedClusterId) {
      return { overview: {}, healthChecks: [], capabilities: [] }
    }
    const [overviewResult, healthResult, capabilities] = await Promise.all([
      getOptionalResource('/overview', selectedClusterId),
      listResource('/health', selectedClusterId, { filters: healthTableFilters.filters }),
      listClusterCapabilities(selectedClusterId)
    ])
    return {
      overview: overviewResult?.item.data as ApiRecord ?? {},
      observedAt: overviewResult?.item.observed_at,
      stale: overviewResult?.item.stale,
      healthChecks: healthResult.items,
      staleReason: healthResult.staleReason,
      capabilities
    }
  }, [healthTableFilters.filters, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)

  const capacity = readRecord(data?.overview.capacity)
  const services = readRecord(data?.overview.services)
  const clientIO = readRecord(data?.overview.client_io)
  const usedPercent = capacityPercent(capacity)
  const supportedCapabilities = data?.capabilities.filter((item) => item.supported).length ?? 0

  async function refreshAll() {
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    setRefreshing(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, scope: 'all' }), '刷新成功')
      await refresh()
    } finally {
      setRefreshing(false)
    }
  }

  async function toggleHealth(row: ApiRecord, muted: boolean) {
    if (!selectedClusterId) {
      return
    }
    const code = textValue(row.code ?? row.name ?? row.natural_key, '')
    await operationMutation.run(() => mutateResource(muted ? '/health/mute' : '/health/mute', muted ? 'DELETE' : 'POST', {
      cluster_id: selectedClusterId,
      code
    }), muted ? '健康检查取消静默执行成功' : '健康检查静默执行成功')
    await refresh()
  }

  return (
    <Page title="总览" loading={loading} error={error}>
      <Space direction="vertical" size={16} className="page-stack">
        <Card
          className="page-surface-card overview-surface-card"
          title="集群总览"
          extra={
            <Space>
              <HealthBadge status={textValue(data?.overview.health_status, 'UNKNOWN')} />
              <Button icon={<ReloadOutlined />} loading={refreshing} onClick={refreshAll}>刷新集群</Button>
            </Space>
          }
        >
          <Space direction="vertical" size={16} className="page-stack">
            <div className="metrics-grid">
              <MetricCard icon={<DatabaseOutlined />} label="容量使用率" value={`${usedPercent}%`} detail={`${formatBytes(capacity.used_bytes)} / ${formatBytes(capacity.total_bytes)}`} />
              <MetricCard icon={<HddOutlined />} label="OSD" value={serviceValue(services.osd, 'up', 'total')} detail={`in ${servicePart(services.osd, 'in')}`} />
              <MetricCard icon={<ApiOutlined />} label="MON" value={serviceValue(services.mon, 'in_quorum', 'total')} detail="quorum / total" />
              <MetricCard icon={<SafetyCertificateOutlined />} label="能力" value={`${supportedCapabilities}/${data?.capabilities.length ?? 0}`} detail="supported capabilities" />
              <MetricCard icon={<ThunderboltOutlined />} label="读写吞吐" value={`${formatBytes(clientIO.read_bytes_per_second)}/s`} detail={`write ${formatBytes(clientIO.write_bytes_per_second)}/s`} />
            </div>
            <Card title="容量">
              <Progress percent={usedPercent} strokeColor="#168766" />
              <Descriptions size="small" column={{ xs: 1, sm: 3 }}>
                <Descriptions.Item label="Total">{formatBytes(capacity.total_bytes)}</Descriptions.Item>
                <Descriptions.Item label="Used">{formatBytes(capacity.used_bytes)}</Descriptions.Item>
                <Descriptions.Item label="Available">{formatBytes(capacity.available_bytes)}</Descriptions.Item>
              </Descriptions>
            </Card>
          </Space>
        </Card>

        <div className="content-grid">
          <Card title="健康检查">
            <AppTable<ApiRecord>
              size="small"
              rowKey={(row) => textValue(row.code ?? row.name ?? row.natural_key)}
              dataSource={data?.healthChecks ?? []}
              pagination={{ defaultPageSize: 10, showSizeChanger: true }}
              onChange={(_pagination, filters) => healthTableFilters.handleFilterChange(tableFilters(filters))}
              columns={[
                { ...filterColumn('Code', 'code', healthTableFilters), ellipsis: true },
                { ...filterColumn('级别', 'severity', healthTableFilters), render: (value) => <HealthBadge status={textValue(value)} /> },
                { ...filterColumn('摘要', 'summary', healthTableFilters), ellipsis: true },
                { ...filterColumn('数量', 'count', healthTableFilters), width: 80, render: (value) => textValue(value, '-') },
                {
                  title: '操作',
                  width: 80,
                  render: (_, row) => {
                    const muted = Boolean(row.muted)
                    return <TableAction onClick={() => toggleHealth(row, muted)}>{muted ? '取消静默' : '静默'}</TableAction>
                  }
                }
              ]}
            />
          </Card>
        </div>
      </Space>
    </Page>
  )
}

function MetricCard({ icon, label, value, detail }: { icon: React.ReactNode; label: string; value: string; detail: string }) {
  return (
    <Card className="dashboard-metric-card">
      <div className="dashboard-metric-icon metric-tone-green">{icon}</div>
      <div>
        <Text type="secondary">{label}</Text>
        <strong>{value}</strong>
        <span>{detail}</span>
      </div>
    </Card>
  )
}

function readRecord(value: unknown): ApiRecord {
  return typeof value === 'object' && value !== null && !Array.isArray(value) ? value as ApiRecord : {}
}

function capacityPercent(capacity: ApiRecord) {
  const used = numberValue(capacity.used_bytes)
  const total = numberValue(capacity.total_bytes)
  return used !== undefined && total ? Math.round((used / total) * 100) : 0
}

function formatBytes(value: unknown) {
  const bytes = numberValue(value)
  if (bytes === undefined) {
    return '-'
  }
  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB']
  let size = bytes
  let index = 0
  while (size >= 1024 && index < units.length - 1) {
    size /= 1024
    index += 1
  }
  return `${size.toFixed(size >= 10 || index === 0 ? 0 : 1)} ${units[index]}`
}

function filterColumn(title: string, field: string, tableFilters: ReturnType<typeof useResourceTableFilters>) {
  return {
    title,
    dataIndex: field,
    key: field,
    filterMultiple: true,
    filterSearch: true,
    filters: (tableFilters.filterOptions[field] ?? []).map((value) => ({ text: value, value })),
    filteredValue: tableFilters.filters[field] ?? null
  }
}

function tableFilters(filters: Record<string, unknown>) {
  return Object.fromEntries(
    Object.entries(filters)
      .map(([field, values]) => [field, Array.isArray(values) ? values.map(String).filter(Boolean) : []] as const)
      .filter(([, values]) => values.length > 0)
  )
}

function serviceValue(value: unknown, primary: string, secondary: string) {
  const record = readRecord(value)
  return `${textValue(record[primary], '-')}/${textValue(record[secondary], '-')}`
}

function servicePart(value: unknown, key: string) {
  return textValue(readRecord(value)[key], '-')
}

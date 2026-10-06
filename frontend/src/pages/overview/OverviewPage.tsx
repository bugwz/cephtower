import {
  ApiOutlined,
  DatabaseOutlined,
  HddOutlined,
  ReloadOutlined,
  SafetyCertificateOutlined,
  SyncOutlined,
  ThunderboltOutlined
} from '@ant-design/icons'
import { Alert, Button, Card, Descriptions, Form, Input, Progress, Space, Switch, Tag, Typography } from 'antd'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { listClusterCapabilities, type ClusterCapability } from '../../api/cluster'
import { getOptionalResource, listResource, mutateResource, refreshResource } from '../../api/resource'
import { numberValue, textValue, type ApiRecord } from '../../api/client'
import { AppTable } from '../../components/AppTable'
import { HealthBadge } from '../../components/HealthBadge'
import { DraggableModal } from '../../components/DraggableModal'
import { ResourceMetaBar } from '../../components/ResourceMetaBar'
import { Page } from '../../components/Page'
import { TableAction } from '../../components/TableActions'
import { useResource } from '../../hooks'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useResourceTableFilters } from '../../hooks/useResourceTableFilters'
import { useClusterContext } from '../../state/ClusterContext'
import { pgCategory, pgSummary } from './pgCategory'
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
  return <OverviewContent key={selectedClusterId ?? 'none'} selectedClusterId={selectedClusterId} />
}

function OverviewContent({ selectedClusterId }: { selectedClusterId: number | undefined }) {
  const active = useRef(true)
  useEffect(() => { active.current = true; return () => { active.current = false } }, [])
  const [refreshing, setRefreshing] = useState(false)
  const operationMutation = useMutationOperation()
  const [muteTarget, setMuteTarget] = useState<ApiRecord | null>(null)
  const [muteForm] = Form.useForm<{ ttl?: string; sticky?: boolean }>()
  const [mutatingHealth, setMutatingHealth] = useState(false)
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
      stale: Boolean(overviewResult?.item.stale || healthResult.stale),
      healthChecks: healthResult.items,
      staleReason: healthResult.staleReason,
      capabilities
    }
  }, [healthTableFilters.filters, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)

  const capacity = readRecord(data?.overview.capacity)
  const services = readRecord(data?.overview.services)
  const clientIO = readRecord(data?.overview.client_io)
  const objectStats = readRecord(data?.overview.object_stats)
  const usedPercent = capacityPercent(capacity)
  const objectHealth = objectHealthSummary(objectStats)
  const supportedCapabilities = data?.capabilities.filter((item) => item.supported).length ?? 0

  async function refreshAll() {
    if (!active.current) return
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    setRefreshing(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, scope: 'all' }), false)
      if (!active.current) return
      message.success('刷新成功')
      await refresh()
    } finally {
      if (active.current) setRefreshing(false)
    }
  }

  async function toggleHealth(row: ApiRecord, muted: boolean, options: { ttl?: string; sticky?: boolean } = {}) {
    if (!active.current || !selectedClusterId || mutatingHealth) return
    const code = textValue(row.code ?? row.name ?? row.natural_key, '')
    setMutatingHealth(true)
    try {
      await operationMutation.run(() => mutateResource('/health/mute', muted ? 'DELETE' : 'POST', {
        cluster_id: selectedClusterId, code,
        ...(!muted ? { ...(options.ttl ? { ttl: options.ttl.trim() } : {}), sticky: Boolean(options.sticky) } : {})
      }), false)
      if (!active.current) return
      message.success(muted ? '健康检查已取消静默' : '健康检查已静默')
      setMuteTarget(null)
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kind: 'health_check' }), false)
      if (!active.current) return
      await refresh()
    } finally {
      if (active.current) setMutatingHealth(false)
    }
  }

  const pgStates = useMemo(() => Array.isArray(data?.overview.placement_groups)
    ? (data.overview.placement_groups as ApiRecord[]) : [], [data?.overview.placement_groups])
  const pgTotals = useMemo(() => pgSummary(data?.overview.placement_groups), [data?.overview.placement_groups])

  return (
    <Page title="总览" loading={loading} error={error}>
      <Space direction="vertical" size={16} className="page-stack">
        <ResourceMetaBar observedAt={data?.observedAt} stale={data?.stale} staleReason={data?.staleReason} />
        {data?.stale ? <Alert type="warning" showIcon message="采集数据已过期，请刷新后确认当前集群状态。" /> : null}
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
              <MetricCard icon={<DatabaseOutlined />} label="容量使用率" value={usedPercent === undefined ? '未知' : `${usedPercent}%`} detail={`${formatBytes(capacity.used_bytes)} / ${formatBytes(capacity.total_bytes)}`} />
              <MetricCard icon={<HddOutlined />} label="OSD" value={serviceValue(services.osd, 'up', 'total')} detail={`in ${servicePart(services.osd, 'in')}`} />
              <MetricCard icon={<ApiOutlined />} label="MON" value={serviceValue(services.mon, 'in_quorum', 'total')} detail="quorum / total" />
              <MetricCard icon={<SafetyCertificateOutlined />} label="能力" value={`${supportedCapabilities}/${data?.capabilities.length ?? 0}`} detail="supported capabilities" />
              <MetricCard icon={<ThunderboltOutlined />} label="读写吞吐" value={`${formatBytes(clientIO.read_bytes_per_second)}/s`} detail={`write ${formatBytes(clientIO.write_bytes_per_second)}/s`} />
              <MetricCard icon={<SyncOutlined />} label="恢复吞吐" value={`${formatBytes(clientIO.recovering_bytes_per_second)}/s`} detail={`scrub ${scrubStatusLabel(data?.overview.scrub_status)}`} />
            </div>
            <Card title="容量">
              {usedPercent === undefined ? <Alert type="warning" message="容量使用率不可用：数据缺失、总容量为零或数值不一致，不代表使用率为 0%。" /> : <Progress percent={usedPercent} strokeColor="#168766" />}
              <Descriptions size="small" column={{ xs: 1, sm: 3 }}>
                <Descriptions.Item label="Total">{formatBytes(capacity.total_bytes)}</Descriptions.Item>
                <Descriptions.Item label="Used">{formatBytes(capacity.used_bytes)}</Descriptions.Item>
                <Descriptions.Item label="Available">{formatBytes(capacity.available_bytes)}</Descriptions.Item>
              </Descriptions>
            </Card>
          </Space>
        </Card>

        <div className="content-grid">
          <Card title="Placement Groups">
            <Space wrap>
              <Text type="secondary">PG 总数：{pgTotals?.total ?? '—'}</Text>
              {pgTotals?.categories.map((category) => <Tag key={category.key} color={category.color}>{category.label}：{category.count}</Tag>)}
            </Space>
            {!pgTotals && <Alert type="warning" message="PG 分类统计不可用：数据缺失或计数无效，不代表没有 PG。" />}
            <AppTable<ApiRecord> size="small" rowKey="name" dataSource={pgStates} pagination={false} columns={[
              { title: '状态', dataIndex: 'name', render: (value) => <Tag>{String(value)}</Tag> },
              { key: 'category', title: '分类', dataIndex: 'name', filters: [{ text: '正常', value: 'clean' }, { text: '处理中', value: 'working' }, { text: '告警', value: 'warning' }, { text: '未知', value: 'unknown' }], onFilter: (value, row) => pgCategory(row.name).key === value, render: (value) => { const category = pgCategory(value); return <Tag color={category.color}>{category.label}</Tag> } },
              { title: '数量', dataIndex: 'count' },
              { title: '占比', render: (_, row) => pgTotals ? <Progress percent={pgTotals.total ? Math.round(Number(row.count) / pgTotals.total * 1000) / 10 : 0} /> : <Text type="secondary">—</Text> }
            ]} />
          </Card>
          <Card title="对象健康">
            {objectHealth.known && objectHealth.percent !== undefined
              ? <Progress percent={objectHealth.percent} status={objectHealth.affected > 0 ? 'exception' : 'normal'} />
              : <Text type="secondary">{objectHealth.known ? '当前没有对象副本，健康比例不适用' : '对象副本统计缺失或不一致，健康比例未知'}</Text>}
            <Descriptions column={1} size="small" bordered>
              <Descriptions.Item label="对象数">{formatCount(objectStats.objects)}</Descriptions.Item>
              <Descriptions.Item label="健康副本">{formatCount(objectHealth.healthy)}</Descriptions.Item>
              <Descriptions.Item label="降级副本">{formatCount(objectStats.degraded)}</Descriptions.Item>
              <Descriptions.Item label="错位副本">{formatCount(objectStats.misplaced)}</Descriptions.Item>
              <Descriptions.Item label="未找到副本">{formatCount(objectStats.unfound)}</Descriptions.Item>
            </Descriptions>
          </Card>
          <Card title="客户端 I/O 与服务">
            <Descriptions column={1} size="small" bordered>
              <Descriptions.Item label="读取 IOPS">{textValue(clientIO.read_ops_per_second, '—')}</Descriptions.Item>
              <Descriptions.Item label="写入 IOPS">{textValue(clientIO.write_ops_per_second, '—')}</Descriptions.Item>
              <Descriptions.Item label="读取吞吐">{formatBytes(clientIO.read_bytes_per_second)}/s</Descriptions.Item>
              <Descriptions.Item label="写入吞吐">{formatBytes(clientIO.write_bytes_per_second)}/s</Descriptions.Item>
              <Descriptions.Item label="恢复吞吐">{formatBytes(clientIO.recovering_bytes_per_second)}/s</Descriptions.Item>
              <Descriptions.Item label="Scrub 状态">{scrubStatusLabel(data?.overview.scrub_status)}</Descriptions.Item>
              <Descriptions.Item label="存储池">{formatCount(data?.overview.pool_count)}</Descriptions.Item>
              <Descriptions.Item label="平均 PG / OSD">{formatDecimal(data?.overview.pgs_per_osd)}</Descriptions.Item>
              <Descriptions.Item label="MGR active / standby">{serviceValue(services.mgr, 'active', 'standby')}</Descriptions.Item>
              <Descriptions.Item label="MDS active / standby">{serviceValue(services.mds, 'active', 'standby')}</Descriptions.Item>
            </Descriptions>
          </Card>
        </div>
        <div className="content-grid">
          <Card title="健康检查">
            <AppTable<ApiRecord>
              size="small"
              rowKey={(row) => textValue(row.code ?? row.name ?? row.natural_key)}
              dataSource={data?.healthChecks ?? []}
              pagination={{ defaultPageSize: 10, showSizeChanger: true }}
              onChange={(_pagination, filters) => healthTableFilters.handleFilterChange(tableFilters(filters))}
              expandable={{
                rowExpandable: (row) => Array.isArray(row.detail) && row.detail.length > 0,
                expandedRowRender: (row) => <Space direction="vertical">{(row.detail as string[]).map((detail, index) => <Text key={index}>{detail}</Text>)}</Space>
              }}
              columns={[
                { title: '检查状态', dataIndex: 'active', render: (value) => value === false ? '当前未触发' : '正在触发' },
                { title: '静默到期时间', dataIndex: 'mute_until', render: (value, row) => row.muted ? textValue(value, '未设置') : '—' },
                { title: '持续静默', dataIndex: 'sticky', render: (value) => value ? '是' : '否' },
                { title: '静默', dataIndex: 'muted', render: (value) => <Tag color={value ? 'warning' : 'default'}>{value ? '已静默' : '未静默'}</Tag> },
                { ...filterColumn('Code', 'code', healthTableFilters), ellipsis: true },
                { ...filterColumn('级别', 'severity', healthTableFilters), render: (value) => <HealthBadge status={textValue(value)} /> },
                { ...filterColumn('摘要', 'summary', healthTableFilters), ellipsis: true },
                { ...filterColumn('数量', 'count', healthTableFilters), width: 80, render: (value) => textValue(value, '-') },
                {
                  title: '操作',
                  width: 80,
                  render: (_, row) => {
                    const muted = Boolean(row.muted)
                    return <TableAction disabled={mutatingHealth || Boolean(row.stale)} onClick={() => {
                      if (muted) { void toggleHealth(row, true); return }
                      muteForm.setFieldsValue({ ttl: '1h', sticky: false })
                      setMuteTarget(row)
                    }}>{muted ? '取消静默' : '静默'}</TableAction>
                  }
                }
              ]}
            />
          </Card>
        </div>
      </Space>
      <DraggableModal title={`静默健康检查 ${textValue(muteTarget?.code, '')}`} open={Boolean(muteTarget)}
        onCancel={() => { if (!mutatingHealth) setMuteTarget(null) }} onOk={() => muteForm.submit()} confirmLoading={mutatingHealth} destroyOnClose>
        <Form form={muteForm} layout="vertical" onFinish={(values) => { if (muteTarget) void toggleHealth(muteTarget, false, values) }}>
          <Form.Item name="ttl" label="静默时限" extra="支持 s（秒）、m（分）、h（时）、d（天）、w（周）；留空表示不设置到期时间。"
            rules={[{ pattern: /^[1-9][0-9]{0,8}[smhdw]?$/, message: '请输入正整数时长，例如 1h 或 30m' }]}>
            <Input placeholder="1h" />
          </Form.Item>
          <Form.Item name="sticky" label="持续静默" valuePropName="checked" extra="启用后，健康检查恢复或数量变化不会自动取消静默，仍受静默时限约束。">
            <Switch />
          </Form.Item>
        </Form>
      </DraggableModal>
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
  const used = nonNegativeQuantity(capacity.used_bytes)
  const total = nonNegativeQuantity(capacity.total_bytes)
  if (used === undefined || total === undefined || total <= 0 || used > total) return undefined
  return Math.round((used / total) * 100)
}

function nonNegativeQuantity(value: unknown): number | undefined {
  if (typeof value !== 'number' && (typeof value !== 'string' || !/^(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?$/.test(value))) return undefined
  const parsed = numberValue(value)
  return parsed !== undefined && parsed >= 0 ? parsed : undefined
}

function formatBytes(value: unknown) {
  const bytes = nonNegativeQuantity(value)
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

function formatCount(value: unknown) {
  const count = nonNegativeQuantity(value)
  return count === undefined || !Number.isSafeInteger(count) ? '—' : count.toLocaleString()
}

function formatDecimal(value: unknown) {
  const number = numberValue(value)
  return number === undefined ? '—' : number.toLocaleString(undefined, { maximumFractionDigits: 2 })
}

function objectHealthSummary(stats: ApiRecord) {
  const unknown = { known: false, affected: 0, healthy: undefined, percent: undefined }
  const values = ['copies', 'degraded', 'misplaced', 'unfound'].map(key => nonNegativeQuantity(stats[key]))
  if (values.some(value => value === undefined || !Number.isSafeInteger(value))) return unknown
  const [copies, degraded, misplaced, unfound] = values as number[]
  const affected = degraded + misplaced + unfound
  if (!Number.isSafeInteger(affected) || affected > copies) return unknown
  const healthy = copies - affected
  return {
    known: true,
    affected,
    healthy,
    percent: copies > 0 ? Math.round(healthy / copies * 1000) / 10 : undefined
  }
}

function scrubStatusLabel(value: unknown) {
  switch (textValue(value, '')) {
    case 'active': return '进行中'
    case 'inactive': return '空闲'
    case 'disabled': return '已禁用'
    default: return '未知'
  }
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

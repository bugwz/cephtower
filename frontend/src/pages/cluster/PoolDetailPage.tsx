import { ArrowLeftOutlined, ReloadOutlined } from '@ant-design/icons'
import { Button, Card, Descriptions, Empty, Input, Space, Table, Tag, Typography } from 'antd'
import { useCallback, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { isRecord, numberValue, textValue, type ApiRecord } from '../../api/client'
import { getOptionalResource, refreshResource } from '../../api/resource'
import type { ResourceDTO } from '../../api/types'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useClusterContext } from '../../state/ClusterContext'
import { formatDateTime } from '../../utils/time'

const { Text } = Typography
const twoColumnDescriptions = { xs: 1, sm: 2, md: 2, lg: 2, xl: 2, xxl: 2 }
const excludedFallbackDetailKeys = new Set([
  'configuration',
  'raw_detail',
  'kind',
  'natural_key',
  'resource_version',
  'source',
  'observed_at',
  'created_at',
  'updated_at',
  'stale',
  'data_protection_display',
  'pg_status_display'
])

interface DetailRow {
  key: string
  name: string
  value: unknown
}

interface ConfigRow {
  key: string
  name: string
  configKey: string
  description: string
  source: string
  value: unknown
}

const rbdConfigMetadata: Record<string, { name: string, description: string }> = {
  rbd_qos_bps_burst: { name: 'BPS 突发', description: '所需的 IO 字节数突发上限。' },
  rbd_qos_bps_limit: { name: 'BPS 上限', description: '所需的每秒 IO 字节数上限。' },
  rbd_qos_iops_burst: { name: 'IOPS 突发', description: '所需的 IO 操作次数突发上限。' },
  rbd_qos_iops_limit: { name: 'IOPS 上限', description: '所需的每秒 IO 操作次数上限。' },
  rbd_qos_read_bps_burst: { name: '读 BPS 突发', description: '所需的读取的字节数突发上限。' },
  rbd_qos_read_bps_limit: { name: '读 BPS 上限', description: '所需的每秒内读取的字节数上限。' },
  rbd_qos_read_iops_burst: { name: '读 IOPS 突发', description: '所需的读操作次数突发上限。' },
  rbd_qos_read_iops_limit: { name: '读 IOPS 上限', description: '所需的每秒读操作次数上限。' },
  rbd_qos_write_bps_burst: { name: '写 BPS 突发', description: '所需的写入的字节数突发上限。' },
  rbd_qos_write_bps_limit: { name: '写 BPS 上限', description: '所需的每秒内写入的字节数上限。' },
  rbd_qos_write_iops_burst: { name: '写 IOPS 突发', description: '所需的写操作次数突发上限。' },
  rbd_qos_write_iops_limit: { name: '写 IOPS 上限', description: '所需的每秒写操作次数上限。' }
}

export function PoolDetailPage() {
  const navigate = useNavigate()
  const { name = '' } = useParams()
  const { selectedClusterId } = useClusterContext()
  const decodedName = name
  const [refreshing, setRefreshing] = useState(false)
  const [detailSearch, setDetailSearch] = useState('')
  const [configSearch, setConfigSearch] = useState('')
  const operationMutation = useMutationOperation()
  const loader = useCallback(async () => {
    if (!selectedClusterId || !decodedName) {
      return null
    }
    const payload = await getOptionalResource('/pool', selectedClusterId, { pool: decodedName })
    return payload ? normalizePoolDetail(resourceToRecord(payload.item)) : null
  }, [decodedName, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const detailRows = useMemo(() => filterRows(poolDetailRows(data), detailSearch), [data, detailSearch])
  const configRows = useMemo(() => filterRows(poolConfigRows(data), configSearch), [data, configSearch])

  async function refreshPoolDetail() {
    if (!selectedClusterId || !decodedName || refreshing) {
      return
    }
    setRefreshing(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kind: 'pool' }), '刷新成功')
      await refresh()
    } finally {
      setRefreshing(false)
    }
  }

  return (
    <Page title="存储池详情" loading={loading} error={error}>
      <Space direction="vertical" size={16} className="page-stack">
        <Card
          className="page-surface-card"
          title="基础信息"
          extra={
            <Space>
              <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/cluster/pool')}>返回</Button>
              <Button icon={<ReloadOutlined />} loading={refreshing || loading} onClick={refreshPoolDetail}>刷新</Button>
            </Space>
          }
        >
          {data ? renderOverview(data, decodedName) : (
            <Text type="secondary">暂无存储池详情</Text>
          )}
        </Card>

        <Card className="page-surface-card" title="详细信息">
          <Space direction="vertical" size={12} className="full-width-control">
            <Text type="secondary">来源：Ceph pool 详情，采集命令为 ceph osd pool ls detail --format json。</Text>
            <Input.Search allowClear placeholder="搜索 Key 或 Value" onChange={(event) => setDetailSearch(event.target.value)} onSearch={setDetailSearch} />
            <Table<DetailRow>
              size="small"
              rowKey="key"
              columns={[
                { title: 'Key', dataIndex: 'name', width: '36%', sorter: (left, right) => left.name.localeCompare(right.name) },
                { title: 'Value', dataIndex: 'value', render: renderDetailValue }
              ]}
              dataSource={detailRows}
              pagination={{ pageSize: 10, showSizeChanger: true }}
              locale={{ emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无详细信息" /> }}
            />
          </Space>
        </Card>

        <Card className="page-surface-card" title="配置">
          <Space direction="vertical" size={12} className="full-width-control">
            <Text type="secondary">来源：RBD pool 配置，采集命令为 rbd config pool list {textValue(data?.name, decodedName)} --format json。</Text>
            <Input.Search allowClear placeholder="搜索名称、密钥、来源或值" onChange={(event) => setConfigSearch(event.target.value)} onSearch={setConfigSearch} />
            <Table<ConfigRow>
              size="small"
              rowKey="key"
              columns={[
                { title: '名称', dataIndex: 'name', width: '18%', sorter: (left, right) => left.name.localeCompare(right.name) },
                { title: '描述', dataIndex: 'description', render: (value) => textValue(value) },
                { title: '密钥', dataIndex: 'configKey', width: '24%', sorter: (left, right) => left.configKey.localeCompare(right.configKey) },
                { title: '来源', dataIndex: 'source', width: 120, render: (value) => textValue(value) },
                { title: '值', dataIndex: 'value', width: 160, render: renderConfigValue }
              ]}
              dataSource={configRows}
              pagination={{ pageSize: 10, showSizeChanger: true }}
              locale={{ emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无配置项，刷新后仍为空通常表示当前集群不支持 rbd config pool list" /> }}
            />
          </Space>
        </Card>
      </Space>
    </Page>
  )
}

function renderOverview(data: ApiRecord, decodedName: string) {
  return (
    <Descriptions size="small" column={twoColumnDescriptions} bordered>
      <Descriptions.Item label="名称">{textValue(data.name, decodedName)}</Descriptions.Item>
      <Descriptions.Item label="状态">{textValue(data.status)}</Descriptions.Item>
      <Descriptions.Item label="Pool 类型">{textValue(data.type)}</Descriptions.Item>
      <Descriptions.Item label="数据保护">{textValue(data.data_protection_display)}</Descriptions.Item>
      <Descriptions.Item label="PG 状态">{textValue(data.pg_status_display)}</Descriptions.Item>
      <Descriptions.Item label="PG 自动伸缩">{textValue(data.pg_autoscale_mode)}</Descriptions.Item>
      <Descriptions.Item label="应用标记" span={2}>{renderApplications(poolApplications(data))}</Descriptions.Item>
      <Descriptions.Item label="CRUSH 规则集">{textValue(data.crush_rule)}</Descriptions.Item>
      <Descriptions.Item label="压缩模式">{textValue(data.compression_mode, 'none')}</Descriptions.Item>
      <Descriptions.Item label="最大字节数">{formatQuota(data.quota_max_bytes ?? data.max_bytes)}</Descriptions.Item>
      <Descriptions.Item label="最大对象数">{textValue(data.quota_max_objects ?? data.max_objects ?? 0)}</Descriptions.Item>
      <Descriptions.Item label="资源版本">{textValue(data.resource_version)}</Descriptions.Item>
      <Descriptions.Item label="采集时间">{formatDateTime(data.observed_at)}</Descriptions.Item>
      <Descriptions.Item label="更新时间">{formatDateTime(data.updated_at)}</Descriptions.Item>
    </Descriptions>
  )
}

function normalizePoolDetail(row: ApiRecord): ApiRecord {
  const type = poolType(row)
  const size = numberValue(row.size)
  const pgNum = numberValue(row.pg_num)
  const pgAutoscale = textValue(row.pg_autoscale_mode, 'on')
  return {
    ...row,
    type,
    data_protection_display: type === 'erasure' ? 'erasure' : `replica: x${size ?? 3}`,
    pg_status_display: pgNum ? `${pgNum} active+clean / ${pgAutoscale}` : `active+clean / ${pgAutoscale}`
  }
}

function resourceToRecord(item: ResourceDTO<ApiRecord>): ApiRecord {
  const data = isRecord(item.data) ? item.data : {}
  return {
    ...data,
    kind: item.kind,
    natural_key: item.natural_key,
    name: item.name ?? data.name,
    status: item.status ?? data.status,
    resource_version: item.resource_version,
    source: item.source,
    observed_at: item.observed_at,
    created_at: item.created_at,
    updated_at: item.updated_at,
    stale: item.stale
  }
}

function renderApplications(applications: string[]) {
  if (applications.length === 0) {
    return <Text type="secondary">-</Text>
  }
  return (
    <Space size={[4, 4]} wrap>
      {applications.map((application) => <Tag color="processing" key={application}>{application}</Tag>)}
    </Space>
  )
}

function poolApplications(row: ApiRecord) {
  if (Array.isArray(row.applications)) {
    return row.applications.map((item) => textValue(item, '')).filter(Boolean)
  }
  if (isRecord(row.application_metadata)) {
    return Object.keys(row.application_metadata)
  }
  return []
}

function poolType(row: ApiRecord) {
  return textValue(row.pool_type ?? row.type, 'replicated').toLowerCase() === 'erasure' ? 'erasure' : 'replicated'
}

function formatQuota(value: unknown) {
  const bytes = numberValue(value) ?? 0
  if (bytes <= 0) {
    return '0'
  }
  return textValue(bytes)
}

function poolDetailRows(row: ApiRecord | null | undefined): DetailRow[] {
  if (!row) {
    return []
  }
  const raw = isRecord(row.raw_detail) ? row.raw_detail : compactFallbackDetail(row)
  const rows: DetailRow[] = []
  flattenDetail(raw, '', rows)
  return rows.sort((left, right) => left.name.localeCompare(right.name))
}

function compactFallbackDetail(row: ApiRecord) {
  const result: ApiRecord = {}
  Object.entries(row).forEach(([key, value]) => {
    if (!excludedFallbackDetailKeys.has(key)) {
      result[key] = value
    }
  })
  return result
}

function flattenDetail(value: unknown, prefix: string, rows: DetailRow[]) {
  if (isRecord(value)) {
    const entries = Object.entries(value)
    if (entries.length === 0 && prefix) {
      rows.push({ key: prefix, name: prefix, value: '{}' })
      return
    }
    entries.forEach(([key, item]) => {
      const name = prefix ? `${prefix} ${key}` : key
      if (isRecord(item)) {
        flattenDetail(item, name, rows)
      } else {
        rows.push({ key: name, name, value: item })
      }
    })
    return
  }
  if (prefix) {
    rows.push({ key: prefix, name: prefix, value })
  }
}

function poolConfigRows(row: ApiRecord | null | undefined): ConfigRow[] {
  const raw = row?.configuration
  if (!Array.isArray(raw)) {
    return []
  }
  return raw
    .filter(isRecord)
    .map((item, index) => {
      const configKey = textValue(item.name ?? item.key ?? item.option, '')
      const metadata = rbdConfigMetadata[configKey]
      return {
        key: `${configKey || 'config'}-${index}`,
        name: metadata?.name ?? configKey,
        configKey,
        description: textValue(item.description ?? item.desc ?? item.help, metadata?.description ?? ''),
        source: formatConfigSource(item.source ?? item.level ?? item.who),
        value: item.value ?? item.val ?? ''
      }
    })
    .filter((item) => item.configKey)
}

function filterRows<T extends { name: string, value: unknown }>(rows: T[], keyword: string): T[] {
  const normalized = keyword.trim().toLowerCase()
  if (!normalized) {
    return rows
  }
  return rows.filter((row) => Object.values(row).some((value) => textValue(value, '').toLowerCase().includes(normalized)))
}

function renderDetailValue(value: unknown) {
  const display = value === null || value === undefined || value === '' ? '—' : textValue(value)
  if (display === '—') {
    return <Text type="secondary">—</Text>
  }
  return <Text style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>{display}</Text>
}

function renderConfigValue(value: unknown, row: ConfigRow) {
  const display = textValue(value, '')
  if (!display) {
    return <Text type="secondary">—</Text>
  }
  if (row.configKey.includes('_bps_') || row.configKey.endsWith('_bps_limit') || row.configKey.endsWith('_bps_burst')) {
    return renderDetailValue(`${display} B/s`)
  }
  if (row.configKey.includes('_iops_') || row.configKey.endsWith('_iops_limit') || row.configKey.endsWith('_iops_burst')) {
    return renderDetailValue(`${display} IOPS`)
  }
  return renderDetailValue(value)
}

function formatConfigSource(value: unknown) {
  const source = textValue(value, '')
  if (source === 'global') {
    return '全局'
  }
  if (source === 'pool') {
    return '存储池'
  }
  return source
}

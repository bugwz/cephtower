import { ArrowLeftOutlined, ReloadOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Descriptions, Empty, Input, Space, Table, Tag, Typography } from 'antd'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { isRecord, textValue, type ApiRecord } from '../../api/client'
import { getOptionalResource, refreshResource } from '../../api/resource'
import type { ResourceDTO } from '../../api/types'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useClusterContext } from '../../state/ClusterContext'
import { formatDateTime } from '../../utils/time'
import { PoolIOHistory } from './PoolIOHistory'
import { PoolPGDistribution } from './PoolPGDistribution'
import { PoolPGStateTags } from './PoolPGStateTags'
import { PoolCacheTiers } from './PoolCacheTiers'
import { poolPGStatus, poolCapacity, poolObjectCount, poolUsage, poolDataProtection, poolKind, poolIORate, poolPGAdjustment } from './PoolManagementPage'

const { Text } = Typography
const twoColumnDescriptions = { xs: 1, sm: 2, md: 2, lg: 2, xl: 2, xxl: 2 }
const excludedFallbackDetailKeys = new Set([
  'history_scope',
  'profile_warning',
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
  const { name = '' } = useParams()
  const { selectedClusterId } = useClusterContext()
  return <ScopedPoolDetailPage key={JSON.stringify([selectedClusterId, name])} selectedClusterId={selectedClusterId} decodedName={name} />
}

function ScopedPoolDetailPage({ selectedClusterId, decodedName }: { selectedClusterId?: number; decodedName: string }) {
  const navigate = useNavigate()
  const activeRef = useRef(true)
  useEffect(() => {
    activeRef.current = true
    return () => { activeRef.current = false }
  }, [])
  const resourceScope = useRef({ clusterId: selectedClusterId, name: decodedName })
  if (resourceScope.current.clusterId !== selectedClusterId || resourceScope.current.name !== decodedName) {
    resourceScope.current = { clusterId: selectedClusterId, name: decodedName }
  }
  const [refreshing, setRefreshing] = useState(false)
  const [detailSearch, setDetailSearch] = useState('')
  const [configSearch, setConfigSearch] = useState('')
  const operationMutation = useMutationOperation()
  const loader = useCallback(async (): Promise<ApiRecord | null> => {
    if (!selectedClusterId || !decodedName) {
      return null
    }
    return loadPoolDetail(selectedClusterId, decodedName)
  }, [decodedName, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const detailRows = useMemo(() => filterRows(poolDetailRows(data), detailSearch), [data, detailSearch])
  const configRows = useMemo(() => filterRows(poolConfigRows(data), configSearch), [data, configSearch])

  async function refreshPoolDetail() {
    if (!activeRef.current || !selectedClusterId || !decodedName || refreshing) {
      return
    }
    setRefreshing(true)
    const scope = resourceScope.current
    try {
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kind: 'pool' }), false)
      if (!activeRef.current || resourceScope.current !== scope) return
      await refresh()
    } finally {
      if (activeRef.current && resourceScope.current === scope) setRefreshing(false)
    }
  }

  return (
    <Page title="存储池详情" loading={loading} error={error}>
      <Space direction="vertical" size={16} className="page-stack">
        {data && poolDetailFreshnessWarning(data) && <Alert type="warning" showIcon message={poolDetailFreshnessWarning(data)} />}
        {typeof data?.profile_warning === 'string' && <Alert type="warning" showIcon message={data.profile_warning} />}
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

        {selectedClusterId && !loading && !error && data?.history_scope === `${selectedClusterId}/${decodedName}` && typeof data.id === 'number' && Number.isSafeInteger(data.id) && data.id >= 0 && <PoolIOHistory key={`${selectedClusterId}/${data.id}`} clusterId={selectedClusterId} poolId={data.id} />}
        <PoolPGDistribution key={`${selectedClusterId}/${decodedName}`} value={data?.pg_status} />
        {selectedClusterId && !loading && !error && typeof data?.id === 'number' && Number.isSafeInteger(data.id) && data.id >= 0 && <PoolCacheTiers key={`${selectedClusterId}/${data.id}`} clusterId={selectedClusterId} poolId={data.id} />}
        <Card className="page-surface-card" title="自动伸缩建议">
          <Text type="secondary">来源：ceph osd pool autoscale-status。以下为采集时的建议与计算依据，不代表已执行调整。</Text>
          {renderAutoscaleStatus(data?.autoscale_status)}
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

function autoscaleNumber(value: unknown) {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0 ? String(value) : '未采集'
}

function autoscaleBoolean(value: unknown) {
  return value === true ? '是' : value === false ? '否' : '未采集'
}

function renderAutoscaleStatus(value: unknown) {
  if (!isRecord(value)) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="未采集自动伸缩建议，模块不可用或命令失败时不会推断建议。" />
  return <Descriptions size="small" column={twoColumnDescriptions} bordered>
    <Descriptions.Item label="建议 PG 数量">{poolObjectCount(value.pg_num_final)}</Descriptions.Item>
    <Descriptions.Item label="建议调整">{autoscaleBoolean(value.would_adjust)}</Descriptions.Item>
    <Descriptions.Item label="目标数据量">{poolCapacity(value.target_bytes)}</Descriptions.Item>
    <Descriptions.Item label="自动伸缩逻辑用量">{poolCapacity(value.logical_used)}</Descriptions.Item>
    <Descriptions.Item label="冗余开销倍数">{autoscaleNumber(value.raw_used_rate)}</Descriptions.Item>
    <Descriptions.Item label="实际容量占比（原始比值）">{autoscaleNumber(value.actual_capacity_ratio)}</Descriptions.Item>
    <Descriptions.Item label="计算容量占比（含目标用量）">{autoscaleNumber(value.capacity_ratio)}</Descriptions.Item>
    <Descriptions.Item label="CRUSH 子树容量">{poolCapacity(value.subtree_capacity)}</Descriptions.Item>
    <Descriptions.Item label="目标比例">{autoscaleNumber(value.target_ratio)}</Descriptions.Item>
    <Descriptions.Item label="有效目标比例">{autoscaleNumber(value.effective_target_ratio)}</Descriptions.Item>
    <Descriptions.Item label="自动伸缩偏置">{autoscaleNumber(value.bias)}</Descriptions.Item>
    <Descriptions.Item label="Bulk 标志">{autoscaleBoolean(value.bulk)}</Descriptions.Item>
  </Descriptions>
}

function renderOverview(data: ApiRecord, decodedName: string) {
  return (
    <Descriptions size="small" column={twoColumnDescriptions} bordered>
      <Descriptions.Item label="名称">{textValue(data.name, decodedName)}</Descriptions.Item>
      <Descriptions.Item label="状态">{textValue(data.status)}</Descriptions.Item>
      <Descriptions.Item label="Pool 类型">{textValue(data.type)}</Descriptions.Item>
      <Descriptions.Item label="数据保护">{textValue(data.data_protection_display)}</Descriptions.Item>
      <Descriptions.Item label="I/O 所需最小副本/分片数">{poolMinimumSize(data.min_size)}</Descriptions.Item>
      <Descriptions.Item label="PG 状态"><PoolPGStateTags value={data.pg_status} /></Descriptions.Item>
      <Descriptions.Item label="PG 调整（采集时）">{poolPGAdjustment(data)}</Descriptions.Item>
      <Descriptions.Item label="PG 自动伸缩">{textValue(data.pg_autoscale_mode)}</Descriptions.Item>
      <Descriptions.Item label="当前 PG 数量">{poolObjectCount(data.pg_num)}</Descriptions.Item>
      <Descriptions.Item label="目标 PG 数量">{poolObjectCount(data.pg_num_target)}</Descriptions.Item>
      <Descriptions.Item label="当前 PGP 数量">{poolObjectCount(data.pgp_num)}</Descriptions.Item>
      <Descriptions.Item label="目标 PGP 数量">{poolObjectCount(data.pgp_num_target)}</Descriptions.Item>
      <Descriptions.Item label="使用率">{poolUsage(data)}</Descriptions.Item>
      <Descriptions.Item label="用户数据量">{poolCapacity(data.stored)}</Descriptions.Item>
      <Descriptions.Item label="实际占用">{poolCapacity(data.bytes_used)}</Descriptions.Item>
      <Descriptions.Item label="最大可用">{poolCapacity(data.max_avail)}</Descriptions.Item>
      <Descriptions.Item label="对象数量">{poolObjectCount(data.objects)}</Descriptions.Item>
      <Descriptions.Item label="读取速率（采集时）">{poolIORate(data.client_io_rate, 'read_bytes_sec')}</Descriptions.Item>
      <Descriptions.Item label="写入速率（采集时）">{poolIORate(data.client_io_rate, 'write_bytes_sec')}</Descriptions.Item>
      <Descriptions.Item label="读取 IOPS（采集时）">{poolIORate(data.client_io_rate, 'read_op_per_sec')}</Descriptions.Item>
      <Descriptions.Item label="写入 IOPS（采集时）">{poolIORate(data.client_io_rate, 'write_op_per_sec')}</Descriptions.Item>
      <Descriptions.Item label="累计读取量">{poolCapacity(data.read_bytes)}</Descriptions.Item>
      <Descriptions.Item label="累计写入量">{poolCapacity(data.write_bytes)}</Descriptions.Item>
      <Descriptions.Item label="累计读取次数">{poolObjectCount(data.read_operations)}</Descriptions.Item>
      <Descriptions.Item label="累计写入次数">{poolObjectCount(data.write_operations)}</Descriptions.Item>
      <Descriptions.Item label="压缩数据实际占用">{poolCapacity(data.compress_bytes_used)}</Descriptions.Item>
      <Descriptions.Item label="压缩数据原始大小">{poolCapacity(data.compress_under_bytes)}</Descriptions.Item>
      <Descriptions.Item label="纠删码配置">{textValue(data.erasure_code_profile, '未采集')}</Descriptions.Item>
      <Descriptions.Item label="应用标记" span={2}>{renderApplications(poolApplications(data))}</Descriptions.Item>
      <Descriptions.Item label="CRUSH 规则集">{textValue(data.crush_rule)}</Descriptions.Item>
      <Descriptions.Item label="压缩模式">{textValue(data.compression_mode, '未采集')}</Descriptions.Item>
      <Descriptions.Item label="压缩算法">{textValue(data.compression_algorithm, '未采集')}</Descriptions.Item>
      <Descriptions.Item label="最小压缩 Blob 大小">{poolCapacity(data.compression_min_blob_size)}</Descriptions.Item>
      <Descriptions.Item label="最大压缩 Blob 大小">{poolCapacity(data.compression_max_blob_size)}</Descriptions.Item>
      <Descriptions.Item label="所需压缩比（配置阈值）">{formatCompressionRatio(data.compression_required_ratio)}</Descriptions.Item>
      <Descriptions.Item label="最大字节数">{formatQuota(data.quota_max_bytes)}</Descriptions.Item>
      <Descriptions.Item label="最大对象数">{formatQuota(data.quota_max_objects)}</Descriptions.Item>
      <Descriptions.Item label="资源版本">{textValue(data.resource_version)}</Descriptions.Item>
      <Descriptions.Item label="采集时间">{formatDateTime(data.observed_at)}</Descriptions.Item>
      <Descriptions.Item label="更新时间">{formatDateTime(data.updated_at)}</Descriptions.Item>
    </Descriptions>
  )
}

function poolDetailFreshnessWarning(row: ApiRecord): string | undefined {
  if (row.stale === false) return undefined
  return row.stale === true
    ? '存储池库存已过期，以下状态和统计仅反映历史采集结果，请刷新后再判断。'
    : '存储池库存时效未知，无法确认以下状态和统计是否仍然有效，请重新采集。'
}

async function loadPoolDetail(clusterId: number, name: string): Promise<ApiRecord | null> {
  const payload = await getOptionalResource('/pool', clusterId, { pool: name })
  if (!payload) return null
  const row = resourceToRecord(payload.item)
  const profiles: ApiRecord[] = []
  let warning: string | undefined
  if (poolKind(row) === 'erasure') {
    if (typeof row.erasure_code_profile === 'string' && row.erasure_code_profile.trim()) {
      try {
        const profile = await getOptionalResource('/erasure/code/profile', clusterId, { name: row.erasure_code_profile })
        if (profile) profiles.push(resourceToRecord(profile.item))
      } catch {
        // Optional profile inventory must not hide the successfully loaded pool.
      }
    }
    if (!poolDataProtection(row, profiles).startsWith('EC: ')) {
      warning = '纠删码 profile 不可用、已过期或分片参数无效，无法确认 k+m；池详情仍展示已采集的数据。'
    }
  }
  return { ...normalizePoolDetail(row, profiles), profile_warning: warning, history_scope: `${clusterId}/${name}` }
}

function normalizePoolDetail(row: ApiRecord, profiles: ApiRecord[] = []): ApiRecord {
  const type = poolKind(row) ?? '未知'
  return {
    ...row,
    type,
    data_protection_display: poolDataProtection(row, profiles),
    pg_status_display: poolPGStatus(row.pg_status)
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

function poolMinimumSize(value: unknown) {
  return typeof value === 'number' && Number.isSafeInteger(value) && value > 0 ? String(value) : '未采集'
}

function formatCompressionRatio(value: unknown) {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0 && value <= 1
    ? String(value)
    : '未采集'
}

function formatQuota(value: unknown) {
  if (typeof value !== 'number' || !Number.isSafeInteger(value) || value < 0) {
    return '未采集'
  }
  return value === 0 ? '无限制（0）' : value.toLocaleString('zh-CN')
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

import { InfoCircleOutlined, PlusOutlined, ReloadOutlined, SaveOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, InputNumber, Select, Space, Tag, Tooltip, Typography } from 'antd'
import { useCallback, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { isRecord, numberValue, textValue, type ApiRecord } from '../../api/client'
import { listResource, mutateResource, refreshResource } from '../../api/resource'
import { DataTable } from '../../components/DataTable'
import { DraggableModal } from '../../components/DraggableModal'
import { Page } from '../../components/Page'
import { ResourceMetaBar } from '../../components/ResourceMetaBar'
import { TableAction, TableActions } from '../../components/TableActions'
import { useResource } from '../../hooks'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useResourceTableFilters } from '../../hooks/useResourceTableFilters'
import { useClusterContext } from '../../state/ClusterContext'
import { message } from '../../utils/appMessage'

const { Text } = Typography

type PoolFormMode = 'create' | 'edit'
type QuotaUnit = 'B' | 'KiB' | 'MiB' | 'GiB' | 'TiB' | 'PiB'

interface PoolFormValues {
  name: string
  pool_type: 'replicated' | 'erasure'
  pg_autoscale_mode: 'on' | 'off' | 'warn'
  size?: number
  applications: string[]
  crush_rule: string
  compression_mode: 'none' | 'passive' | 'aggressive' | 'force'
  quota_max_bytes?: number
  quota_unit: QuotaUnit
  quota_max_objects?: number
}

interface PoolPageData {
  pools: ApiRecord[]
  crushRules: string[]
  observedAt?: string | null
  stale: boolean
  staleReason?: string | null
}

const applicationDefaults = ['rbd', 'cephfs', 'rgw', 'mgr']
const quotaUnits: QuotaUnit[] = ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB']
const defaultPoolValues: PoolFormValues = {
  name: '',
  pool_type: 'replicated',
  pg_autoscale_mode: 'on',
  size: 3,
  applications: [],
  crush_rule: 'replicated_rule',
  compression_mode: 'none',
  quota_max_bytes: 0,
  quota_unit: 'GiB',
  quota_max_objects: 0
}

export function PoolManagementPage() {
  const navigate = useNavigate()
  const { selectedClusterId } = useClusterContext()
  const [form] = Form.useForm<PoolFormValues>()
  const [formMode, setFormMode] = useState<PoolFormMode>('create')
  const [formOpen, setFormOpen] = useState(false)
  const [editingPool, setEditingPool] = useState<ApiRecord | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [refreshingPools, setRefreshingPools] = useState(false)
  const poolType = Form.useWatch('pool_type', form) ?? 'replicated'
  const operationMutation = useMutationOperation()
  const poolTableFilters = useResourceTableFilters({
    path: '/pools',
    fields: ['name', 'status', 'type', 'pg_autoscale_mode', 'size', 'applications', 'crush_rule', 'compression_mode'],
    clusterId: selectedClusterId
  })
  const loader = useCallback(async (): Promise<PoolPageData> => {
    if (!selectedClusterId) {
      return { pools: [], crushRules: ['replicated_rule'], observedAt: null, stale: false, staleReason: null }
    }
    const [poolList, crushRules] = await Promise.all([
      listResource('/pools', selectedClusterId, { filters: poolTableFilters.filters }),
      listResource('/crush/rules', selectedClusterId).then((payload) => payload.items.map(resourceName).filter(Boolean)).catch(() => [])
    ])
    return {
      pools: poolList.items.map(normalizePoolRow),
      crushRules: Array.from(new Set(['replicated_rule', ...crushRules])),
      observedAt: poolList.observedAt,
      stale: poolList.stale,
      staleReason: poolList.staleReason
    }
  }, [poolTableFilters.filters, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const applicationOptions = useMemo(() => {
    const fromRows = (data?.pools ?? []).flatMap((row) => poolApplications(row))
    return Array.from(new Set([...applicationDefaults, ...fromRows])).map((value) => ({ label: value, value }))
  }, [data?.pools])
  const crushRuleOptions = useMemo(() => (data?.crushRules ?? ['replicated_rule']).map((value) => ({ label: value, value })), [data?.crushRules])

  async function refreshPoolData() {
    if (refreshingPools) {
      return
    }
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    setRefreshingPools(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kind: 'pool' }), '刷新成功')
      await refresh()
    } finally {
      setRefreshingPools(false)
    }
  }

  function openCreate() {
    setFormMode('create')
    setEditingPool(null)
    form.setFieldsValue(defaultPoolValues)
    setFormOpen(true)
  }

  function openEdit(row: ApiRecord) {
    setFormMode('edit')
    setEditingPool(row)
    form.setFieldsValue(poolInitialValues(row, data?.crushRules))
    setFormOpen(true)
  }

  async function submitPool(values: PoolFormValues) {
    if (!selectedClusterId || submitting) {
      return
    }
    setSubmitting(true)
    try {
      if (formMode === 'create') {
        await operationMutation.run(() => mutateResource('/pool', 'POST', poolCreateBody(values, selectedClusterId)), false)
        message.success('存储池创建执行成功')
      } else if (editingPool) {
        const requests = poolUpdateBodies(editingPool, values, selectedClusterId, data?.crushRules)
        if (requests.length === 0) {
          message.info('没有需要提交的更改')
          return
        }
        for (const body of requests) {
          await operationMutation.run(() => mutateResource('/pool', 'PATCH', body, { ifMatch: Number(editingPool.resource_version ?? 0) }), false)
        }
        message.success('存储池已更新')
      }
      setFormOpen(false)
      setEditingPool(null)
      void refresh({ showLoading: false })
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Page title="存储池管理" loading={loading} error={error}>
      <Card
        className="page-surface-card"
        title="存储池列表"
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} loading={refreshingPools} onClick={refreshPoolData}>刷新</Button>
            <Button type="primary" icon={<PlusOutlined />} disabled={!selectedClusterId} onClick={openCreate}>新建存储池</Button>
          </Space>
        }
      >
        <DataTable
          data={data?.pools ?? []}
          filterOptions={poolTableFilters.filterOptions}
          filteredValues={poolTableFilters.filters}
          onFilterChange={poolTableFilters.handleFilterChange}
          footer={<ResourceMetaBar observedAt={data?.observedAt} stale={data?.stale} staleReason={data?.staleReason} />}
          rowKeyCandidates={['natural_key', 'name', 'pool_name']}
          columns={[
            { key: 'name', title: '名称' },
            { key: 'data_protection_display', title: '数据保护', filterKey: 'type', render: (_, row) => <Tag color="default">{textValue(row.data_protection_display)}</Tag> },
            { key: 'applications_display', title: '应用标记', filterKey: 'applications', render: (_, row) => renderApplications(poolApplications(row)) },
            { key: 'pg_status_display', title: 'PG 状态', filterKey: 'pg_autoscale_mode' },
            { key: 'usage_display', title: '使用率', filterKey: false },
            { key: 'read_bytes_display', title: '读字节数', filterKey: false },
            { key: 'write_bytes_display', title: '写字节数', filterKey: false },
            {
              key: 'actions',
              title: '操作',
              filterKey: false,
              render: (_, row) => (
                <TableActions>
                  <TableAction onClick={() => openEdit(row)}>编辑</TableAction>
                  <TableAction onClick={() => navigate(`/cluster/pool/${encodeURIComponent(resourceName(row))}`)}>详情</TableAction>
                </TableActions>
              )
            }
          ]}
        />
      </Card>

      <DraggableModal
        className="cluster-modal"
        title={formMode === 'create' ? '新建存储池' : `编辑存储池：${resourceName(editingPool ?? {})}`}
        open={formOpen}
        width={640}
        onCancel={() => {
          if (!submitting) {
            setFormOpen(false)
            setEditingPool(null)
          }
        }}
        onOk={() => form.submit()}
        okText="保存"
        cancelText="取消"
        confirmLoading={submitting}
        okButtonProps={{ icon: <SaveOutlined />, loading: submitting }}
        cancelButtonProps={{ disabled: submitting }}
        destroyOnClose
        maskClosable={false}
      >
        <Form form={form} className="cluster-form pool-management-form" layout="vertical" onFinish={submitPool}>
          <div className="cluster-form-grid">
            <Form.Item name="name" label="存储池名称" rules={[{ required: true, message: '请输入存储池名称' }]}>
              <Input disabled={formMode === 'edit'} placeholder="请输入存储池名称" />
            </Form.Item>
            <Form.Item name="pool_type" label="存储池类型">
              <Select
                disabled={formMode === 'edit'}
                placeholder="请选择存储池类型"
                options={[
                  { label: '副本池', value: 'replicated' },
                  { label: '纠删码池', value: 'erasure' }
                ]}
              />
            </Form.Item>
            <Form.Item
              name="pg_autoscale_mode"
              label={<HelpLabel label="PG 自动伸缩" title="开启后系统会根据存储池使用情况自动调整 PG 数量。" />}
              rules={[{ required: true, message: '请选择 PG 自动伸缩模式' }]}
            >
              <Select
                placeholder="请选择自动伸缩模式"
                options={[
                  { label: '开启', value: 'on' },
                  { label: '关闭', value: 'off' },
                  { label: '仅告警', value: 'warn' }
                ]}
              />
            </Form.Item>
            <Form.Item name="size" label="副本数" rules={[{ required: true, message: '请输入副本数' }]}>
              <InputNumber
                min={1}
                precision={0}
                disabled={poolType !== 'replicated'}
                className="full-width-control"
                placeholder={poolType === 'replicated' ? '请输入副本数' : '纠删码池无需设置副本数'}
              />
            </Form.Item>
            <Form.Item className="cluster-form-full" name="applications" label="应用标记" rules={[{ required: true, type: 'array', min: 1, message: '请选择应用标记' }]}>
              <Select mode="multiple" allowClear placeholder="请选择应用标记" options={applicationOptions} />
            </Form.Item>
            <Form.Item name="crush_rule" label="数据分布规则集" rules={[{ required: true, message: '请选择数据分布规则集' }]}>
              <Select showSearch placeholder="请选择数据分布规则集" options={crushRuleOptions} />
            </Form.Item>
            <Form.Item name="compression_mode" label={<HelpLabel label="压缩策略" title="不压缩表示该存储池不会主动压缩对象数据。" />}>
              <Select
                placeholder="请选择压缩策略"
                options={[
                  { label: 'none', value: 'none' },
                  { label: 'passive', value: 'passive' },
                  { label: 'aggressive', value: 'aggressive' },
                  { label: 'force', value: 'force' }
                ]}
              />
            </Form.Item>
            <Form.Item name="quota_max_bytes" label={<HelpLabel label="最大容量" title="留空或设置为 0 表示不限制容量。" />}>
              <InputNumber
                min={0}
                precision={0}
                className="full-width-control"
                placeholder="请输入最大容量"
                addonAfter={(
                  <Form.Item name="quota_unit" noStyle>
                    <Select className="pool-quota-unit-select" options={quotaUnits.map((value) => ({ label: value, value }))} />
                  </Form.Item>
                )}
              />
            </Form.Item>
            <Form.Item name="quota_max_objects" label={<HelpLabel label="最大对象数量" title="留空或设置为 0 表示不限制对象数量。" />}>
              <InputNumber min={0} precision={0} className="full-width-control" placeholder="请输入最大对象数量" />
            </Form.Item>
          </div>
        </Form>
      </DraggableModal>
    </Page>
  )
}

function normalizePoolRow(row: ApiRecord): ApiRecord {
  const name = resourceName(row)
  const poolType = poolKind(row)
  const size = numberValue(row.size)
  const pgNum = numberValue(row.pg_num)
  const pgAutoscale = textValue(row.pg_autoscale_mode, 'on')
  return {
    ...row,
    name,
    type: poolType,
    data_protection_display: poolType === 'erasure' ? 'erasure' : `replica: x${size ?? 3}`,
    applications: poolApplications(row),
    applications_display: poolApplications(row).join(', '),
    pg_status_display: pgNum ? `${pgNum} active+clean / ${pgAutoscale}` : `active+clean / ${pgAutoscale}`,
    usage_display: poolUsage(row),
    read_bytes_display: formatBytes(numberValue(row.read_bytes ?? row.client_read_bytes)),
    write_bytes_display: formatBytes(numberValue(row.write_bytes ?? row.client_write_bytes))
  }
}

function poolInitialValues(row: ApiRecord, crushRules: string[] = []): PoolFormValues {
  const quotaBytes = numberValue(row.quota_max_bytes ?? row.max_bytes) ?? 0
  const unit: QuotaUnit = 'GiB'
  return {
    name: resourceName(row),
    pool_type: poolKind(row),
    pg_autoscale_mode: poolAutoscaleMode(row),
    size: numberValue(row.size) ?? 3,
    applications: poolApplications(row),
    crush_rule: readableCrushRule(row.crush_rule, crushRules),
    compression_mode: poolCompressionMode(row),
    quota_max_bytes: fromBytes(quotaBytes, unit),
    quota_unit: unit,
    quota_max_objects: numberValue(row.quota_max_objects ?? row.max_objects) ?? 0
  }
}

function poolCreateBody(values: PoolFormValues, clusterId: number): ApiRecord {
  return {
    cluster_id: clusterId,
    name: values.name,
    pool_type: values.pool_type,
    pg_num: 32,
    pg_autoscale_mode: values.pg_autoscale_mode,
    size: values.pool_type === 'replicated' ? Math.trunc(values.size ?? 3) : undefined,
    applications: values.applications ?? [],
    crush_rule: values.crush_rule,
    compression_mode: values.compression_mode,
    quota_max_bytes: quotaBytes(values.quota_max_bytes, values.quota_unit),
    quota_unit: values.quota_unit,
    quota_max_objects: Math.trunc(values.quota_max_objects ?? 0)
  }
}

function poolUpdateBodies(row: ApiRecord, values: PoolFormValues, clusterId: number, crushRules: string[] = []): ApiRecord[] {
  const current = poolInitialValues(row, crushRules)
  const pool = current.name
  const requests: ApiRecord[] = []
  const pushField = (field: string, value: string | number | undefined, previous: string | number | undefined) => {
    if (value === undefined || String(value) === String(previous ?? '')) {
      return
    }
    requests.push({ cluster_id: clusterId, pool, field, value: String(value) })
  }
  pushField('pg_autoscale_mode', values.pg_autoscale_mode, current.pg_autoscale_mode)
  if (values.pool_type === 'replicated') {
    pushField('size', Math.trunc(values.size ?? 3), Math.trunc(current.size ?? 3))
  }
  pushField('crush_rule', values.crush_rule, current.crush_rule)
  pushField('compression_mode', values.compression_mode, current.compression_mode)

  const nextQuotaBytes = quotaBytes(values.quota_max_bytes, values.quota_unit)
  const currentQuotaBytes = quotaBytes(current.quota_max_bytes, current.quota_unit)
  if (nextQuotaBytes !== currentQuotaBytes) {
    requests.push({ cluster_id: clusterId, pool, operation: 'quota', field: 'max_bytes', value: String(nextQuotaBytes), quota_unit: values.quota_unit })
  }
  const nextObjects = Math.trunc(values.quota_max_objects ?? 0)
  const currentObjects = Math.trunc(current.quota_max_objects ?? 0)
  if (nextObjects !== currentObjects) {
    requests.push({ cluster_id: clusterId, pool, operation: 'quota', field: 'max_objects', value: String(nextObjects) })
  }

  const nextApps = new Set(values.applications ?? [])
  const currentApps = new Set(current.applications ?? [])
  currentApps.forEach((application) => {
    if (!nextApps.has(application)) {
      requests.push({ cluster_id: clusterId, pool, operation: 'application', action: 'disable', application, applications: Array.from(nextApps) })
    }
  })
  nextApps.forEach((application) => {
    if (!currentApps.has(application)) {
      requests.push({ cluster_id: clusterId, pool, operation: 'application', action: 'enable', application, applications: Array.from(nextApps) })
    }
  })
  return requests
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

function HelpLabel({ label, title }: { label: string, title: string }) {
  return (
    <span className="pool-form-help-label">
      {label}
      <Tooltip title={title}>
        <InfoCircleOutlined className="pool-form-help-icon" />
      </Tooltip>
    </span>
  )
}

function poolApplications(row: ApiRecord) {
  if (Array.isArray(row.applications)) {
    return row.applications.map((item) => textValue(item, '')).filter(Boolean)
  }
  if (isRecord(row.application_metadata)) {
    return Object.keys(row.application_metadata)
  }
  const value = textValue(row.application_metadata, '')
  return value ? value.split(',').map((item) => item.trim()).filter(Boolean) : []
}

function resourceName(row: ApiRecord) {
  return textValue(row.name ?? row.pool_name ?? row.natural_key, '')
}

function readableCrushRule(value: unknown, crushRules: string[] = []) {
  const raw = textValue(value, 'replicated_rule')
  const index = Number(raw)
  if (Number.isInteger(index) && index >= 0 && crushRules[index]) {
    return crushRules[index]
  }
  return raw
}

function poolKind(row: ApiRecord): 'replicated' | 'erasure' {
  return textValue(row.pool_type ?? row.type, 'replicated').toLowerCase() === 'erasure' ? 'erasure' : 'replicated'
}

function poolAutoscaleMode(row: ApiRecord): PoolFormValues['pg_autoscale_mode'] {
  const mode = textValue(row.pg_autoscale_mode, 'on')
  return mode === 'off' || mode === 'warn' ? mode : 'on'
}

function poolCompressionMode(row: ApiRecord): PoolFormValues['compression_mode'] {
  const mode = textValue(row.compression_mode, 'none')
  return mode === 'passive' || mode === 'aggressive' || mode === 'force' ? mode : 'none'
}

function poolUsage(row: ApiRecord) {
  const percent = numberValue(row.used_percent ?? row.percent_used ?? row.usage_percent)
  if (percent === undefined) {
    return '0%'
  }
  return `${Math.max(0, Math.min(100, percent)).toFixed(percent >= 10 ? 0 : 1)}%`
}

function formatBytes(value?: number) {
  if (!value || value <= 0) {
    return '-'
  }
  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB']
  let size = value
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex += 1
  }
  return `${size.toFixed(size >= 10 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}

function quotaBytes(value: number | undefined, unit: QuotaUnit) {
  const amount = Math.trunc(value ?? 0)
  if (amount <= 0) {
    return 0
  }
  return amount * (1024 ** quotaUnits.indexOf(unit))
}

function fromBytes(value: number, unit: QuotaUnit) {
  if (!value || value <= 0) {
    return 0
  }
  return Math.floor(value / (1024 ** quotaUnits.indexOf(unit)))
}

import { InfoCircleOutlined, PlusOutlined, ReloadOutlined, SaveOutlined } from '@ant-design/icons'
import { Button, Card, Divider, Form, Input, InputNumber, Select, Space, Tag, Tooltip, Typography } from 'antd'
import { type ReactNode, useCallback, useMemo, useState } from 'react'
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
type CompressionAlgorithm = 'snappy' | 'zlib' | 'zstd' | 'lz4'
type RbdMirroringMode = 'disabled' | 'pool'
type PoolFlagState = 'on' | 'off'

interface PoolFormValues {
  name: string
  pool_type: 'replicated' | 'erasure'
  pg_autoscale_mode: 'on' | 'off' | 'warn'
  pg_num?: number
  size?: number
  applications: string[]
  erasure_code_profile: string
  crush_rule: string
  allow_ec_overwrites: PoolFlagState
  compression_mode: 'none' | 'passive' | 'aggressive' | 'force'
  compression_algorithm?: CompressionAlgorithm
  compression_min_blob_size?: number
  compression_min_blob_size_unit: QuotaUnit
  compression_max_blob_size?: number
  compression_max_blob_size_unit: QuotaUnit
  compression_required_ratio?: number
  quota_max_bytes?: number
  quota_unit: QuotaUnit
  quota_max_objects?: number
  rbd_mirroring: RbdMirroringMode
  configuration: Record<RbdPoolConfigurationKey, number>
}

interface CrushRuleFormValues {
  name: string
  root: string
  failure_domain: string
  device_class?: string
}

type ErasureCodePlugin = 'jerasure' | 'lrc' | 'isa' | 'shec' | 'clay'

interface ErasureCodeProfileFormValues {
  name: string
  plugin: ErasureCodePlugin
  k: number
  m: number
  technique?: string
  packetsize?: number
  l?: number
  crush_locality?: string
  c?: number
  d?: number
  scalar_mds?: 'jerasure' | 'isa' | 'shec'
  crush_failure_domain: string
  crush_num_failure_domains?: number
  crush_osds_per_failure_domain?: number
  crush_root: string
  crush_device_class?: string
  directory?: string
}

type RbdPoolConfigurationKey = typeof rbdPoolConfigurationFields[number]['key']

interface PoolPageData {
  pools: ApiRecord[]
  crushRules: string[]
  erasureCodeProfiles: string[]
  erasureCodeDirectory: string
  osds: ApiRecord[]
  observedAt?: string | null
  stale: boolean
  staleReason?: string | null
}

const applicationDefaults = ['rbd', 'cephfs', 'rgw', 'mgr']
const quotaUnits: QuotaUnit[] = ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB']
const compressionAlgorithms: CompressionAlgorithm[] = ['snappy', 'zlib', 'zstd', 'lz4']
const rbdMirroringModeOptions: Array<{ label: string, value: RbdMirroringMode }> = [
  { label: '关闭', value: 'disabled' },
  { label: '开启', value: 'pool' }
]
const allowECOverwritesPoolFlag = 'allow_ec_overwrites'
const calculationHelpURL = 'https://docs.ceph.com/en/tentacle/rados/operations/placement-groups/#choosing-number-of-placement-groups'
const rbdPoolConfigurationFields = [
  { key: 'rbd_qos_bps_limit', label: 'BPS 上限', unit: 'B/s', description: '所需的每秒 IO 字节数上限。' },
  { key: 'rbd_qos_iops_limit', label: 'IOPS 上限', unit: 'IOPS', description: '所需的每秒 IO 操作次数上限。' },
  { key: 'rbd_qos_read_bps_limit', label: '读 BPS 上限', unit: 'B/s', description: '所需的每秒内读取的字节数上限。' },
  { key: 'rbd_qos_read_iops_limit', label: '读 IOPS 上限', unit: 'IOPS', description: '所需的每秒读操作次数上限。' },
  { key: 'rbd_qos_write_bps_limit', label: '写 BPS 上限', unit: 'B/s', description: '所需的每秒内写入的字节数上限。' },
  { key: 'rbd_qos_write_iops_limit', label: '写 IOPS 上限', unit: 'IOPS', description: '所需的每秒写操作次数上限。' },
  { key: 'rbd_qos_bps_burst', label: 'BPS 突发', unit: 'B/s', description: '所需的 IO 字节数突发上限。' },
  { key: 'rbd_qos_iops_burst', label: 'IOPS 突发', unit: 'IOPS', description: '所需的 IO 操作次数突发上限。' },
  { key: 'rbd_qos_read_bps_burst', label: '读 BPS 突发', unit: 'B/s', description: '所需的读取的字节数突发上限。' },
  { key: 'rbd_qos_read_iops_burst', label: '读 IOPS 突发', unit: 'IOPS', description: '所需的读操作次数突发上限。' },
  { key: 'rbd_qos_write_bps_burst', label: '写 BPS 突发', unit: 'B/s', description: '所需的写入的字节数突发上限。' },
  { key: 'rbd_qos_write_iops_burst', label: '写 IOPS 突发', unit: 'IOPS', description: '所需的写操作次数突发上限。' }
] as const
const defaultErasureCodeProfile = 'default'
const defaultErasureCodeDirectory = '/usr/lib64/ceph/erasure-code'
const erasureCodePlugins: Array<{ label: string, value: ErasureCodePlugin }> = [
  { label: 'Jerasure', value: 'jerasure' },
  { label: 'LRC', value: 'lrc' },
  { label: 'ISA', value: 'isa' },
  { label: 'SHEC', value: 'shec' },
  { label: 'CLAY', value: 'clay' }
]
const defaultCrushRuleValues: CrushRuleFormValues = {
  name: '',
  root: 'default',
  failure_domain: 'host',
  device_class: undefined
}
const defaultErasureCodeProfileValues: ErasureCodeProfileFormValues = {
  name: '',
  plugin: 'isa',
  k: 7,
  m: 3,
  technique: 'reed_sol_van',
  packetsize: 2048,
  l: 3,
  crush_locality: 'host',
  c: 2,
  d: 9,
  scalar_mds: 'isa',
  crush_failure_domain: 'host',
  crush_num_failure_domains: 0,
  crush_osds_per_failure_domain: 0,
  crush_root: 'default',
  crush_device_class: undefined,
  directory: defaultErasureCodeDirectory
}
const defaultPoolValues: PoolFormValues = {
  name: '',
  pool_type: 'replicated',
  pg_autoscale_mode: 'on',
  pg_num: 32,
  size: 3,
  applications: [],
  erasure_code_profile: defaultErasureCodeProfile,
  crush_rule: 'replicated_rule',
  allow_ec_overwrites: 'off',
  compression_mode: 'none',
  compression_algorithm: 'snappy',
  compression_min_blob_size: undefined,
  compression_min_blob_size_unit: 'B',
  compression_max_blob_size: undefined,
  compression_max_blob_size_unit: 'MiB',
  compression_required_ratio: 0.875,
  quota_max_bytes: 0,
  quota_unit: 'B',
  quota_max_objects: 0,
  rbd_mirroring: 'disabled',
  configuration: defaultRbdPoolConfiguration()
}

export function PoolManagementPage() {
  const navigate = useNavigate()
  const { selectedClusterId } = useClusterContext()
  const [form] = Form.useForm<PoolFormValues>()
  const [crushRuleForm] = Form.useForm<CrushRuleFormValues>()
  const [erasureCodeProfileForm] = Form.useForm<ErasureCodeProfileFormValues>()
  const [formMode, setFormMode] = useState<PoolFormMode>('create')
  const [formOpen, setFormOpen] = useState(false)
  const [editingPool, setEditingPool] = useState<ApiRecord | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [crushRuleFormOpen, setCrushRuleFormOpen] = useState(false)
  const [erasureCodeProfileFormOpen, setErasureCodeProfileFormOpen] = useState(false)
  const [submittingCrushRule, setSubmittingCrushRule] = useState(false)
  const [submittingErasureCodeProfile, setSubmittingErasureCodeProfile] = useState(false)
  const [createdCrushRules, setCreatedCrushRules] = useState<string[]>([])
  const [createdErasureCodeProfiles, setCreatedErasureCodeProfiles] = useState<string[]>([])
  const [refreshingPools, setRefreshingPools] = useState(false)
  const poolType = Form.useWatch('pool_type', form) ?? 'replicated'
  const pgAutoscaleMode = Form.useWatch('pg_autoscale_mode', form) ?? 'on'
  const applications = Form.useWatch('applications', form) ?? []
  const compressionMode = Form.useWatch('compression_mode', form) ?? 'none'
  const rbdMirroringMode = Form.useWatch('rbd_mirroring', form) ?? 'disabled'
  const compressionEnabled = compressionMode !== 'none'
  const crushRuleRoot = Form.useWatch('root', crushRuleForm) ?? 'default'
  const crushRuleDeviceClass = Form.useWatch('device_class', crushRuleForm)
  const erasureCodePlugin = Form.useWatch('plugin', erasureCodeProfileForm) ?? 'isa'
  const erasureCodeRoot = Form.useWatch('crush_root', erasureCodeProfileForm) ?? 'default'
  const erasureCodeDeviceClass = Form.useWatch('crush_device_class', erasureCodeProfileForm)
  const erasureCodeScalarMDS = Form.useWatch('scalar_mds', erasureCodeProfileForm)
  const rbdApplicationEnabled = applications.includes('rbd')
  const rbdConfigurationEnabled = poolType === 'replicated' && rbdApplicationEnabled
  const operationMutation = useMutationOperation()
  const poolTableFilters = useResourceTableFilters({
    path: '/pools',
    fields: ['name', 'status', 'type', 'pg_autoscale_mode', 'size', 'applications', 'crush_rule', 'erasure_code_profile', 'compression_mode'],
    clusterId: selectedClusterId
  })
  const loader = useCallback(async (): Promise<PoolPageData> => {
    if (!selectedClusterId) {
      return { pools: [], crushRules: ['replicated_rule'], erasureCodeProfiles: [defaultErasureCodeProfile], erasureCodeDirectory: defaultErasureCodeDirectory, osds: [], observedAt: null, stale: false, staleReason: null }
    }
    const [poolList, crushRules, erasureCodeProfileRows, osds] = await Promise.all([
      listResource('/pools', selectedClusterId, { filters: poolTableFilters.filters }),
      listResource('/crush/rules', selectedClusterId).then((payload) => payload.items.map(resourceName).filter(Boolean)).catch(() => []),
      listResource('/erasure/code/profiles', selectedClusterId).then((payload) => payload.items).catch(() => []),
      listResource('/osds', selectedClusterId).then((payload) => payload.items).catch(() => [])
    ])
    const erasureCodeProfiles = erasureCodeProfileRows.map(resourceName).filter(Boolean)
    const erasureCodeDirectory = erasureCodeProfileRows
      .map((profile) => textValue(profile.directory, ''))
      .find(Boolean) ?? defaultErasureCodeDirectory
    return {
      pools: poolList.items.map(normalizePoolRow),
      crushRules: Array.from(new Set(['replicated_rule', ...crushRules])),
      erasureCodeProfiles: Array.from(new Set([defaultErasureCodeProfile, ...erasureCodeProfiles])),
      erasureCodeDirectory,
      osds,
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
  const crushRuleOptions = useMemo(
    () => Array.from(new Set([...(data?.crushRules ?? ['replicated_rule']), ...createdCrushRules])).map((value) => ({ label: value, value })),
    [createdCrushRules, data?.crushRules]
  )
  const erasureCodeProfileOptions = useMemo(
    () => Array.from(new Set([...(data?.erasureCodeProfiles ?? [defaultErasureCodeProfile]), ...createdErasureCodeProfiles])).map((value) => ({ label: value, value })),
    [createdErasureCodeProfiles, data?.erasureCodeProfiles]
  )
  const crushRootOptions = useMemo(
    () => crushRootNames(data?.osds ?? []).map((value) => ({ label: value, value })),
    [data?.osds]
  )
  const crushDeviceClassOptions = useMemo(
    () => [
      { label: '全部设备', value: '' },
      ...Array.from(new Set((data?.osds ?? []).map((osd) => textValue(osd.device_class, '')).filter(Boolean)))
        .map((value) => ({ label: value, value }))
    ],
    [data?.osds]
  )
  const crushRuleTopology = useMemo(
    () => topologyCounts(data?.osds ?? [], crushRuleRoot, crushRuleDeviceClass),
    [crushRuleDeviceClass, crushRuleRoot, data?.osds]
  )
  const erasureCodeTopology = useMemo(
    () => topologyCounts(data?.osds ?? [], erasureCodeRoot, erasureCodeDeviceClass),
    [data?.osds, erasureCodeDeviceClass, erasureCodeRoot]
  )

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
    form.setFieldsValue({ ...defaultPoolValues, applications: [], configuration: defaultRbdPoolConfiguration() })
    setFormOpen(true)
  }

  function openEdit(row: ApiRecord) {
    setFormMode('edit')
    setEditingPool(row)
    form.setFieldsValue(poolInitialValues(row, data?.crushRules))
    setFormOpen(true)
  }

  function openCrushRuleForm() {
    const root = crushRootOptions[0]?.value ?? 'default'
    const topology = topologyCounts(data?.osds ?? [], root)
    const failureDomain = topology.host > 0 ? 'host' : failureDomainOptions(topology)[0]?.value ?? 'osd'
    crushRuleForm.setFieldsValue({ ...defaultCrushRuleValues, root, failure_domain: failureDomain })
    setCrushRuleFormOpen(true)
  }

  function openErasureCodeProfileForm() {
    const root = crushRootOptions[0]?.value ?? 'default'
    const topology = topologyCounts(data?.osds ?? [], root)
    const failureDomain = topology.host > 0 ? 'host' : failureDomainOptions(topology)[0]?.value ?? 'osd'
    erasureCodeProfileForm.setFieldsValue({
      ...defaultErasureCodeProfileValues,
      crush_root: root,
      crush_failure_domain: failureDomain,
      crush_locality: failureDomain,
      directory: data?.erasureCodeDirectory ?? defaultErasureCodeDirectory
    })
    setErasureCodeProfileFormOpen(true)
  }

  function changeErasureCodePlugin(plugin: ErasureCodePlugin) {
    if (plugin === 'isa') {
      erasureCodeProfileForm.setFieldsValue({ k: 7, m: 3, technique: 'reed_sol_van' })
      return
    }
    if (plugin === 'jerasure') {
      erasureCodeProfileForm.setFieldsValue({ k: 4, m: 2, technique: 'reed_sol_van', packetsize: 2048 })
      return
    }
    if (plugin === 'lrc') {
      erasureCodeProfileForm.setFieldsValue({ k: 4, m: 2, l: 3, crush_locality: 'host' })
      return
    }
    if (plugin === 'shec') {
      erasureCodeProfileForm.setFieldsValue({ k: 4, m: 3, c: 2 })
      return
    }
    erasureCodeProfileForm.setFieldsValue({ k: 4, m: 2, d: 5, scalar_mds: 'isa', technique: 'reed_sol_van' })
  }

  async function submitCrushRule(values: CrushRuleFormValues) {
    if (!selectedClusterId || submittingCrushRule) {
      return
    }
    setSubmittingCrushRule(true)
    try {
      await operationMutation.run(() => mutateResource('/crush/rule', 'POST', {
        cluster_id: selectedClusterId,
        name: values.name,
        root: values.root,
        failure_domain: values.failure_domain,
        device_class: values.device_class || undefined
      }), false)
      setCreatedCrushRules((current) => Array.from(new Set([...current, values.name])))
      form.setFieldValue('crush_rule', values.name)
      setCrushRuleFormOpen(false)
      message.success('CRUSH 规则创建成功并已写入集群')
      void refreshResource({ clusterId: selectedClusterId, kind: 'crush_rule' })
        .then(() => refresh({ showLoading: false }))
        .catch(() => refresh({ showLoading: false }))
    } finally {
      setSubmittingCrushRule(false)
    }
  }

  async function submitErasureCodeProfile(values: ErasureCodeProfileFormValues) {
    if (!selectedClusterId || submittingErasureCodeProfile) {
      return
    }
    setSubmittingErasureCodeProfile(true)
    try {
      await operationMutation.run(() => mutateResource('/erasure/code/profile', 'POST', erasureCodeProfileBody(values, selectedClusterId)), false)
      setCreatedErasureCodeProfiles((current) => Array.from(new Set([...current, values.name])))
      form.setFieldValue('erasure_code_profile', values.name)
      setErasureCodeProfileFormOpen(false)
      message.success('EC Profile 创建成功并已写入集群')
      void refreshResource({ clusterId: selectedClusterId, kind: 'erasure_code_profile' })
        .then(() => refresh({ showLoading: false }))
        .catch(() => refresh({ showLoading: false }))
    } finally {
      setSubmittingErasureCodeProfile(false)
    }
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
          <div className="pool-form-section">
            <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入存储池名称' }]}>
              <Input placeholder="请输入存储池名称" />
            </Form.Item>
            <Form.Item
              name="applications"
              label={<HelpLabel label="应用标记" title="存储池在使用前需要关联至少一个应用。" />}
              rules={[{ required: true, type: 'array', min: 1, message: '请选择应用标记' }]}
            >
              <Select mode="multiple" allowClear placeholder="请选择应用标记" options={applicationOptions} />
            </Form.Item>
            <Form.Item name="pool_type" label="Pool 类型">
              <Select
                disabled={formMode === 'edit'}
                placeholder="请选择存储池类型"
                options={[
                  { label: '副本池', value: 'replicated' },
                  { label: '纠删码池', value: 'erasure' }
                ]}
              />
            </Form.Item>
            {poolType === 'replicated' ? (
              <div className="pool-replicated-fields">
                <Form.Item name="size" label="副本数" rules={[{ required: true, message: '请输入副本数' }]}>
                  <InputNumber disabled={formMode === 'edit'} min={1} precision={0} className="full-width-control" placeholder="副本数" />
                </Form.Item>
                <Form.Item label={<HelpLabel label="CRUSH 规则集" title="选择数据副本的放置规则，或使用右侧按钮在当前 Ceph 集群中创建。" />} required>
                  <Space.Compact block className="pool-inline-create-control">
                    <Form.Item name="crush_rule" noStyle rules={[{ required: true, message: '请选择 CRUSH 规则集' }]}>
                      <Select disabled={formMode === 'edit'} showSearch placeholder="请选择 CRUSH 规则集" options={crushRuleOptions} />
                    </Form.Item>
                    <Tooltip title="创建 CRUSH 规则">
                      <Button icon={<PlusOutlined />} disabled={formMode === 'edit'} onClick={openCrushRuleForm} aria-label="创建 CRUSH 规则" />
                    </Tooltip>
                  </Space.Compact>
                </Form.Item>
              </div>
            ) : (
              <>
                <Form.Item label={<HelpLabel label="纠删码配置" title="选择现有 EC Profile，或使用右侧按钮在当前 Ceph 集群中创建。" />} required>
                  <Space.Compact block className="pool-inline-create-control">
                    <Form.Item name="erasure_code_profile" noStyle rules={[{ required: true, message: '请选择纠删码配置' }]}>
                      <Select disabled={formMode === 'edit'} showSearch placeholder="请选择纠删码配置" options={erasureCodeProfileOptions} />
                    </Form.Item>
                    <Tooltip title="创建 EC Profile">
                      <Button icon={<PlusOutlined />} disabled={formMode === 'edit'} onClick={openErasureCodeProfileForm} aria-label="创建 EC Profile" />
                    </Tooltip>
                  </Space.Compact>
                </Form.Item>
                <Form.Item
                  name="allow_ec_overwrites"
                  label={<HelpLabel label="纠删码覆写" title="允许覆写纠删码池中的对象，对 RBD 使用纠删码池时通常需要开启。" />}
                >
                  <Select
                    options={[
                      { label: '关闭', value: 'off' },
                      { label: '开启', value: 'on' }
                    ]}
                  />
                </Form.Item>
              </>
            )}
            <Form.Item
              name="pg_autoscale_mode"
              label={<HelpLabel label="PG 自动伸缩" title={pgAutoscaleDescription(pgAutoscaleMode)} />}
              rules={[{ required: true, message: '请选择 PG 自动伸缩模式' }]}
            >
              <Select
                placeholder="请选择自动伸缩模式"
                options={[
                  { label: 'on', value: 'on' },
                  { label: 'off', value: 'off' },
                  { label: 'warn', value: 'warn' }
                ]}
              />
            </Form.Item>
            {pgAutoscaleMode !== 'on' ? (
              <Form.Item
                name="pg_num"
                label={(
                  <HelpLabel
                    label="PG 数量"
                    title={<a href={calculationHelpURL} target="_blank" rel="noreferrer">计算帮助</a>}
                  />
                )}
                rules={[{ required: true, message: '请输入 PG 数量' }]}
              >
                <InputNumber min={1} precision={0} className="full-width-control" placeholder="32" />
              </Form.Item>
            ) : null}
            <Form.Item name="compression_mode" label={<HelpLabel label="压缩模式" title={compressionModeDescription(compressionMode)} />}>
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
            {compressionEnabled ? (
              <>
                <Form.Item name="compression_algorithm" label={<HelpLabel label="算法" title="当前存储池使用的压缩算法。" />} rules={[{ required: true, message: '请选择压缩算法' }]}>
                  <Select
                    placeholder="请选择压缩算法"
                    options={compressionAlgorithms.map((value) => ({ label: value, value }))}
                  />
                </Form.Item>
                <Form.Item
                  name="compression_min_blob_size"
                  label={<HelpLabel label="最小 Blob 大小" title="小于最小 Blob 大小的数据块不会被压缩。" />}
                  dependencies={['compression_max_blob_size', 'compression_max_blob_size_unit', 'compression_min_blob_size_unit']}
                  rules={[blobSizeRule('min')]}
                >
                  <InputNumber
                    min={0}
                    precision={0}
                    className="full-width-control"
                    placeholder="留空表示使用集群默认值"
                    addonAfter={(
                      <Form.Item name="compression_min_blob_size_unit" noStyle>
                        <Select className="pool-quota-unit-select" options={quotaUnits.map((value) => ({ label: value, value }))} />
                      </Form.Item>
                    )}
                  />
                </Form.Item>
                <Form.Item
                  name="compression_max_blob_size"
                  label={<HelpLabel label="最大 Blob 大小" title="大于最大 Blob 大小的数据块会先拆分成更小的块再压缩。" />}
                  dependencies={['compression_min_blob_size', 'compression_min_blob_size_unit', 'compression_max_blob_size_unit']}
                  rules={[blobSizeRule('max')]}
                >
                  <InputNumber
                    min={0}
                    precision={0}
                    className="full-width-control"
                    placeholder="留空表示使用集群默认值"
                    addonAfter={(
                      <Form.Item name="compression_max_blob_size_unit" noStyle>
                        <Select className="pool-quota-unit-select" options={quotaUnits.map((value) => ({ label: value, value }))} />
                      </Form.Item>
                    )}
                  />
                </Form.Item>
                <Form.Item
                  name="compression_required_ratio"
                  label={<HelpLabel label="压缩比例" title="压缩后数据块相对原始大小的比例必须至少达到此值，才会保存压缩版本。" />}
                  rules={[{ required: true, type: 'number', min: 0, max: 1, message: '请输入 0 到 1 之间的压缩比例' }]}
                >
                  <InputNumber min={0} max={1} step={0.001} precision={3} className="full-width-control" placeholder="0.875" />
                </Form.Item>
              </>
            ) : null}
            {rbdApplicationEnabled ? (
              <Form.Item name="rbd_mirroring" label={<HelpLabel label="镜像模式" title={rbdMirroringDescription(rbdMirroringMode)} />}>
                <Select
                  placeholder="请选择镜像模式"
                  options={rbdMirroringModeOptions}
                />
              </Form.Item>
            ) : null}
            <Divider className="pool-form-divider" />
            <Form.Item name="quota_max_bytes" label={<HelpLabel label="最大字节数" title="留空或设置为 0 表示禁用此配额；有效配额必须大于 0。" />}>
              <InputNumber
                min={0}
                precision={0}
                className="full-width-control"
                placeholder="0"
                addonAfter={(
                  <Form.Item name="quota_unit" noStyle>
                    <Select className="pool-quota-unit-select" options={quotaUnits.map((value) => ({ label: value, value }))} />
                  </Form.Item>
                )}
              />
            </Form.Item>
            <Form.Item name="quota_max_objects" label={<HelpLabel label="最大对象数" title="留空或设置为 0 表示禁用此配额；有效配额必须大于 0。" />}>
              <InputNumber min={0} precision={0} className="full-width-control" placeholder="0" />
            </Form.Item>
            {rbdConfigurationEnabled ? (
              <>
                <Divider className="pool-form-divider" />
                {rbdPoolConfigurationFields.map((field) => (
                  <Form.Item
                    key={field.key}
                    name={['configuration', field.key]}
                    label={<HelpLabel label={field.label} title={field.description} />}
                  >
                    <InputNumber
                      min={0}
                      precision={0}
                      className="full-width-control"
                      addonAfter={<span className="pool-static-unit">{field.unit}</span>}
                    />
                  </Form.Item>
                ))}
              </>
            ) : null}
          </div>
        </Form>
      </DraggableModal>

      <DraggableModal
        className="cluster-modal nested-cluster-modal"
        title="创建 CRUSH 规则"
        open={crushRuleFormOpen}
        width={640}
        onCancel={() => !submittingCrushRule && setCrushRuleFormOpen(false)}
        onOk={() => crushRuleForm.submit()}
        okText="创建 CRUSH 规则"
        cancelText="取消"
        confirmLoading={submittingCrushRule}
        okButtonProps={{ icon: <SaveOutlined />, loading: submittingCrushRule }}
        cancelButtonProps={{ disabled: submittingCrushRule }}
        destroyOnClose
        maskClosable={false}
      >
        <Form name="crush-rule-form" form={crushRuleForm} className="cluster-form pool-management-form" layout="vertical" onFinish={submitCrushRule}>
          <div className="pool-form-section">
            <Form.Item
              name="name"
              label="名称"
              rules={[
                { required: true, message: '请输入 CRUSH 规则名称' },
                { pattern: /^[A-Za-z0-9_-]+$/, message: '名称只能包含字母、数字、短横线和下划线' },
                uniqueNameRule(crushRuleOptions.map((option) => option.value), 'CRUSH 规则名称已存在')
              ]}
            >
              <Input placeholder="请输入 CRUSH 规则名称" />
            </Form.Item>
            <Form.Item
              name="root"
              label={<HelpLabel label="根节点" title="应用于存放数据的 CRUSH 根节点名称。" />}
              rules={[{ required: true, message: '请选择根节点' }]}
            >
              <Select
                options={crushRootOptions}
                placeholder="请选择根节点"
                getPopupContainer={nestedModalPopupContainer}
              />
            </Form.Item>
            <Form.Item
              name="failure_domain"
              label={<HelpLabel label="故障域类型" title="用于分隔副本的 CRUSH 节点类型。" />}
              rules={[{ required: true, message: '请选择故障域类型' }]}
            >
              <Select
                options={failureDomainOptions(crushRuleTopology)}
                getPopupContainer={nestedModalPopupContainer}
              />
            </Form.Item>
            <Form.Item
              name="device_class"
              label={<HelpLabel label="设备类型" title="限定用于放置数据的 CRUSH 设备类型；全部设备表示不设置 class。" />}
            >
              <Select
                options={crushDeviceClassOptions}
                getPopupContainer={nestedModalPopupContainer}
              />
            </Form.Item>
          </div>
        </Form>
      </DraggableModal>

      <DraggableModal
        className="cluster-modal nested-cluster-modal"
        title="创建 EC Profile"
        open={erasureCodeProfileFormOpen}
        width={640}
        onCancel={() => !submittingErasureCodeProfile && setErasureCodeProfileFormOpen(false)}
        onOk={() => erasureCodeProfileForm.submit()}
        okText="创建 EC Profile"
        cancelText="取消"
        confirmLoading={submittingErasureCodeProfile}
        okButtonProps={{ icon: <SaveOutlined />, loading: submittingErasureCodeProfile }}
        cancelButtonProps={{ disabled: submittingErasureCodeProfile }}
        destroyOnClose
        maskClosable={false}
      >
        <Form form={erasureCodeProfileForm} className="cluster-form pool-management-form" layout="vertical" onFinish={submitErasureCodeProfile}>
          <div className="pool-form-section">
            <Form.Item
              name="name"
              label="名称"
              rules={[
                { required: true, message: '请输入 EC Profile 名称' },
                { pattern: /^[A-Za-z0-9_-]+$/, message: '名称只能包含字母、数字、短横线和下划线' },
                uniqueNameRule(erasureCodeProfileOptions.map((option) => option.value), 'EC Profile 名称已存在')
              ]}
            >
              <Input placeholder="请输入 EC Profile 名称" />
            </Form.Item>
            <Form.Item
              name="plugin"
              label={<HelpLabel label="插件" title={erasureCodePluginDescription(erasureCodePlugin)} />}
              rules={[{ required: true, message: '请选择纠删码插件' }]}
            >
              <Select
                options={erasureCodePlugins}
                onChange={changeErasureCodePlugin}
                getPopupContainer={nestedModalPopupContainer}
              />
            </Form.Item>
            <Form.Item
              name="k"
              label={<HelpLabel label="数据块数 (k)" title="每个对象被拆分的数据块数量。" />}
              dependencies={['m', 'crush_failure_domain', 'crush_num_failure_domains', 'crush_osds_per_failure_domain', 'crush_root', 'crush_device_class']}
              rules={[
                { required: true, message: '请输入数据块数' },
                { type: 'number', min: 2, message: '数据块数必须大于或等于 2' },
                ecChunkRule(erasureCodeTopology)
              ]}
            >
              <InputNumber min={2} precision={0} className="full-width-control" />
            </Form.Item>
            <Form.Item
              name="m"
              label={<HelpLabel label="编码块数 (m)" title="为每个对象计算并存储在不同 OSD 上的编码块数量，也表示可容忍同时故障的 OSD 数。" />}
              dependencies={['k', 'crush_failure_domain', 'crush_num_failure_domains', 'crush_osds_per_failure_domain', 'crush_root', 'crush_device_class']}
              rules={[
                { required: true, message: '请输入编码块数' },
                { type: 'number', min: 1, message: '编码块数必须大于或等于 1' },
                ecChunkRule(erasureCodeTopology)
              ]}
            >
              <InputNumber min={1} precision={0} className="full-width-control" />
            </Form.Item>
            <Form.Item
              name="crush_failure_domain"
              label={<HelpLabel label="CRUSH 故障域" title="确保数据块不会放置在同一个故障域中，并用于生成 chooseleaf CRUSH 规则步骤。" />}
              rules={[{ required: true, message: '请选择 CRUSH 故障域' }]}
            >
              <Select
                options={failureDomainOptions(erasureCodeTopology)}
                getPopupContainer={nestedModalPopupContainer}
              />
            </Form.Item>
            <Form.Item
              name="crush_num_failure_domains"
              label={<HelpLabel label="CRUSH 故障域数量" title="要映射的故障域数量。与每故障域 OSD 数配合使用时会创建 CRUSH MSR 规则。" />}
              dependencies={['crush_osds_per_failure_domain', 'crush_failure_domain', 'crush_root', 'crush_device_class']}
              rules={[ecFailureDomainCountRule(erasureCodeTopology)]}
            >
              <InputNumber min={0} precision={0} className="full-width-control" />
            </Form.Item>
            <Form.Item
              name="crush_osds_per_failure_domain"
              label={<HelpLabel label="每故障域 OSD 数" title="每个故障域最多放置的 OSD 数。大于 1 时会创建 CRUSH MSR 规则。" />}
              dependencies={['crush_num_failure_domains']}
              rules={[ecOSDsPerFailureDomainRule]}
            >
              <InputNumber min={0} precision={0} className="full-width-control" />
            </Form.Item>
            {(erasureCodePlugin === 'isa' || erasureCodePlugin === 'jerasure' || erasureCodePlugin === 'clay') ? (
              <Form.Item
                name="technique"
                label={<HelpLabel label="编码技术" title={erasureCodeTechniqueDescription(erasureCodePlugin)} />}
                rules={[{ required: true, message: '请选择编码技术' }]}
              >
                <Select
                  options={erasureCodeTechniqueOptions(erasureCodePlugin, erasureCodeScalarMDS)}
                  getPopupContainer={nestedModalPopupContainer}
                />
              </Form.Item>
            ) : null}
            {erasureCodePlugin === 'jerasure' ? (
              <Form.Item name="packetsize" label={<HelpLabel label="包大小" title="Jerasure 编码技术使用的数据包大小。" />}>
                <InputNumber min={1} precision={0} className="full-width-control" />
              </Form.Item>
            ) : null}
            {erasureCodePlugin === 'lrc' ? (
              <>
                <Form.Item name="l" label={<HelpLabel label="局部校验块数 (l)" title="LRC 本地恢复组使用的局部校验块数量。" />} rules={[{ required: true, min: 1, type: 'number' }]}>
                  <InputNumber min={1} precision={0} className="full-width-control" />
                </Form.Item>
                <Form.Item name="crush_locality" label={<HelpLabel label="CRUSH 局部性" title="LRC 局部恢复组使用的 CRUSH 故障域。" />}>
                  <Select
                    options={failureDomainOptions(erasureCodeTopology)}
                    getPopupContainer={nestedModalPopupContainer}
                  />
                </Form.Item>
              </>
            ) : null}
            {erasureCodePlugin === 'shec' ? (
              <Form.Item name="c" label={<HelpLabel label="耐久性估算值 (c)" title="SHEC 编码使用的耐久性估算值，不能大于编码块数 m。" />} dependencies={['m']} rules={[shecDurabilityRule]}>
                <InputNumber min={1} precision={0} className="full-width-control" />
              </Form.Item>
            ) : null}
            {erasureCodePlugin === 'clay' ? (
              <>
                <Form.Item name="scalar_mds" label={<HelpLabel label="标量 MDS" title="CLAY 用作构建块的标量 MDS 插件。" />} rules={[{ required: true }]}>
                  <Select
                    options={['jerasure', 'isa', 'shec'].map((value) => ({ label: value.toUpperCase(), value }))}
                    getPopupContainer={nestedModalPopupContainer}
                  />
                </Form.Item>
                <Form.Item name="d" label={<HelpLabel label="辅助块数 (d)" title="CLAY 修复单个数据块时读取的 Helper chunks 数量。" />} dependencies={['k', 'm']} rules={[clayHelperChunksRule]}>
                  <InputNumber min={1} precision={0} className="full-width-control" />
                </Form.Item>
              </>
            ) : null}
            <Form.Item
              name="crush_root"
              label={<HelpLabel label="CRUSH 根节点" title="CRUSH 规则第一步使用的 bucket 名称，例如 default。" />}
              rules={[{ required: true, message: '请选择 CRUSH 根节点' }]}
            >
              <Select
                options={crushRootOptions}
                getPopupContainer={nestedModalPopupContainer}
              />
            </Form.Item>
            <Form.Item
              name="crush_device_class"
              label={<HelpLabel label="CRUSH 设备类型" title={`限定放置数据的设备类型。当前选择范围内可用 OSD：${erasureCodeTopology.osd}`} />}
            >
              <Select
                options={crushDeviceClassOptions}
                getPopupContainer={nestedModalPopupContainer}
              />
            </Form.Item>
            <Form.Item
              name="directory"
              label={<HelpLabel label="目录" title="加载纠删码插件的目录名称。" />}
            >
              <Input placeholder={defaultErasureCodeDirectory} />
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
  const quota = bytesForForm(quotaBytes, 'GiB')
  const minBlobSize = bytesForForm(numberValue(row.compression_min_blob_size), 'B', true)
  const maxBlobSize = bytesForForm(numberValue(row.compression_max_blob_size), 'MiB', true)
  return {
    name: resourceName(row),
    pool_type: poolKind(row),
    pg_autoscale_mode: poolAutoscaleMode(row),
    pg_num: numberValue(row.pg_num) ?? 32,
    size: numberValue(row.size) ?? 3,
    applications: poolApplications(row),
    erasure_code_profile: textValue(row.erasure_code_profile, defaultErasureCodeProfile),
    crush_rule: readableCrushRule(row.crush_rule, crushRules),
    allow_ec_overwrites: poolHasFlag(row, allowECOverwritesPoolFlag) ? 'on' : 'off',
    compression_mode: poolCompressionMode(row),
    compression_algorithm: poolCompressionAlgorithm(row),
    compression_min_blob_size: minBlobSize.value,
    compression_min_blob_size_unit: minBlobSize.unit,
    compression_max_blob_size: maxBlobSize.value,
    compression_max_blob_size_unit: maxBlobSize.unit,
    compression_required_ratio: numberValue(row.compression_required_ratio) ?? 0.875,
    quota_max_bytes: quota.value ?? 0,
    quota_unit: quota.unit,
    quota_max_objects: numberValue(row.quota_max_objects ?? row.max_objects) ?? 0,
    rbd_mirroring: poolRbdMirroringMode(row),
    configuration: poolRbdConfiguration(row)
  }
}

function poolCreateBody(values: PoolFormValues, clusterId: number): ApiRecord {
  return {
    cluster_id: clusterId,
    name: values.name,
    pool_type: values.pool_type,
    pg_num: values.pg_autoscale_mode === 'on' ? 1 : Math.trunc(values.pg_num ?? 32),
    pg_autoscale_mode: values.pg_autoscale_mode,
    size: values.pool_type === 'replicated' ? Math.trunc(values.size ?? 3) : undefined,
    applications: values.applications ?? [],
    erasure_code_profile: values.pool_type === 'erasure' ? values.erasure_code_profile : undefined,
    crush_rule: values.pool_type === 'replicated' ? values.crush_rule : undefined,
    flags: values.pool_type === 'erasure'
      ? values.allow_ec_overwrites === 'on' ? [allowECOverwritesPoolFlag] : []
      : undefined,
    compression_mode: values.compression_mode,
    ...compressionBody(values),
    quota_max_bytes: quotaBytes(values.quota_max_bytes, values.quota_unit),
    quota_unit: values.quota_unit,
    quota_max_objects: Math.trunc(values.quota_max_objects ?? 0),
    rbd_mirroring: values.applications.includes('rbd')
      ? values.rbd_mirroring
      : undefined,
    configuration: values.pool_type === 'replicated' && values.applications.includes('rbd')
      ? normalizedRbdPoolConfiguration(values.configuration)
      : undefined
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
  if (values.pg_autoscale_mode !== 'on') {
    pushField('pg_num', Math.trunc(values.pg_num ?? 32), Math.trunc(current.pg_num ?? 32))
  }
  if (values.compression_mode === 'none') {
    if (current.compression_mode !== 'none') {
      pushField('compression_mode', 'unset', current.compression_mode)
      pushField('compression_algorithm', 'unset', current.compression_algorithm)
      pushField('compression_min_blob_size', 0, optionalBytes(current.compression_min_blob_size, current.compression_min_blob_size_unit))
      pushField('compression_max_blob_size', 0, optionalBytes(current.compression_max_blob_size, current.compression_max_blob_size_unit))
      pushField('compression_required_ratio', 0, current.compression_required_ratio)
    }
  } else {
    pushField('compression_mode', values.compression_mode, current.compression_mode)
    pushField('compression_algorithm', values.compression_algorithm, current.compression_algorithm)
    const currentMinBlobSize = optionalBytes(current.compression_min_blob_size, current.compression_min_blob_size_unit)
    const currentMaxBlobSize = optionalBytes(current.compression_max_blob_size, current.compression_max_blob_size_unit)
    pushField(
      'compression_min_blob_size',
      compressionBytesForUpdate(values.compression_min_blob_size, values.compression_min_blob_size_unit, currentMinBlobSize),
      currentMinBlobSize
    )
    pushField(
      'compression_max_blob_size',
      compressionBytesForUpdate(values.compression_max_blob_size, values.compression_max_blob_size_unit, currentMaxBlobSize),
      currentMaxBlobSize
    )
    pushField('compression_required_ratio', compressionRatioForUpdate(values.compression_required_ratio, current.compression_required_ratio), current.compression_required_ratio)
  }

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
  if (values.pool_type === 'erasure' && values.allow_ec_overwrites !== current.allow_ec_overwrites) {
    requests.push({
      cluster_id: clusterId,
      pool,
      field: allowECOverwritesPoolFlag,
      value: String(values.allow_ec_overwrites === 'on')
    })
  }
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

  if (nextApps.has('rbd')) {
    if (values.rbd_mirroring !== current.rbd_mirroring) {
      requests.push({
        cluster_id: clusterId,
        pool,
        operation: 'rbd_mirroring',
        rbd_mirroring: values.rbd_mirroring
      })
    }
    if (values.pool_type === 'replicated') {
      const nextConfiguration = normalizedRbdPoolConfiguration(values.configuration)
      rbdPoolConfigurationFields.forEach(({ key }) => {
        if (nextConfiguration[key] !== current.configuration[key]) {
          requests.push({
            cluster_id: clusterId,
            pool,
            operation: 'rbd_configuration',
            field: key,
            value: nextConfiguration[key]
          })
        }
      })
    }
  }
  if (values.name !== current.name) {
    requests.push({ cluster_id: clusterId, pool, operation: 'rename', name: values.name })
  }
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

function HelpLabel({ label, title }: { label: string, title: ReactNode }) {
  return (
    <span className="pool-form-help-label">
      {label}
      <Tooltip title={title}>
        <InfoCircleOutlined className="pool-form-help-icon" />
      </Tooltip>
    </span>
  )
}

function nestedModalPopupContainer(trigger: HTMLElement) {
  return trigger.parentElement ?? trigger.ownerDocument.body
}

function crushRootNames(osds: ApiRecord[]) {
  const roots = osds.map((osd) => crushPath(osd).root).filter(Boolean)
  const buckets = osds.flatMap((osd) => Object.values(crushPath(osd)))
  const values = Array.from(new Set([...roots, ...buckets]))
  return values.length > 0 ? values : ['default']
}

function topologyCounts(osds: ApiRecord[], root: string, deviceClass?: string): Record<string, number> {
  const matching = osds.filter((osd) => {
    const osdClass = textValue(osd.device_class, '')
    const path = crushPath(osd)
    return (root === 'default' || Object.values(path).includes(root)) && (!deviceClass || osdClass === deviceClass)
  })
  const counts: Record<string, number> = { osd: matching.length }
  matching.forEach((osd) => {
    Object.entries(crushPath(osd)).forEach(([kind]) => {
      counts[kind] = new Set(matching.map((item) => crushPath(item)[kind]).filter(Boolean)).size
    })
  })
  if (counts.host === undefined) {
    counts.host = new Set(matching.map((osd) => textValue(osd.host, '')).filter(Boolean)).size
  }
  return counts
}

function crushPath(osd: ApiRecord): Record<string, string> {
  const path: Record<string, string> = {}
  if (isRecord(osd.crush_path)) {
    Object.entries(osd.crush_path).forEach(([kind, name]) => {
      const value = textValue(name, '')
      if (value) {
        path[kind] = value
      }
    })
  }
  const host = textValue(osd.host, '')
  if (host && !path.host) {
    path.host = host
  }
  return path
}

function failureDomainOptions(counts: Record<string, number>) {
  const order = (kind: string) => kind === 'host' ? 0 : kind === 'osd' ? 1 : 2
  return Object.entries(counts)
    .filter(([kind]) => kind !== 'root')
    .sort(([left], [right]) => order(left) - order(right) || left.localeCompare(right))
    .map(([kind, count]) => ({ label: `${kind} (${count})`, value: kind }))
}

function uniqueNameRule(existingNames: string[], messageText: string) {
  const names = new Set(existingNames)
  return {
    validator(_: unknown, value: unknown) {
      if (typeof value === 'string' && value && names.has(value)) {
        return Promise.reject(new Error(messageText))
      }
      return Promise.resolve()
    }
  }
}

function ecChunkRule(counts: Record<string, number>) {
  return ({ getFieldValue }: { getFieldValue: (name: keyof ErasureCodeProfileFormValues) => unknown }) => ({
    validator() {
      const k = numberValue(getFieldValue('k')) ?? 0
      const m = numberValue(getFieldValue('m')) ?? 0
      const failureDomain = textValue(getFieldValue('crush_failure_domain'), 'host')
      const numFailureDomains = numberValue(getFieldValue('crush_num_failure_domains')) ?? 0
      const osdsPerFailureDomain = numberValue(getFieldValue('crush_osds_per_failure_domain')) ?? 0
      if (k <= 0 || m <= 0 || numFailureDomains > 0 || osdsPerFailureDomain > 0) {
        return Promise.resolve()
      }
      const available = counts[failureDomain] ?? 0
      const chunks = k + m + (failureDomain === 'host' ? 1 : 0)
      if (available > 0 && chunks > available) {
        const expression = failureDomain === 'host' ? 'k+m+1' : 'k+m'
        return Promise.reject(new Error(`数据块 (${expression}) 已超过可用 ${failureDomain} 数量 ${available}`))
      }
      return Promise.resolve()
    }
  })
}

function ecFailureDomainCountRule(counts: Record<string, number>) {
  return ({ getFieldValue }: { getFieldValue: (name: keyof ErasureCodeProfileFormValues) => unknown }) => ({
    validator(_: unknown, value: unknown) {
      const count = numberValue(value) ?? 0
      const osdsPerDomain = numberValue(getFieldValue('crush_osds_per_failure_domain')) ?? 0
      if (count < 0) {
        return Promise.reject(new Error('CRUSH 故障域数量不能小于 0'))
      }
      if (osdsPerDomain > 0 && count < 1) {
        return Promise.reject(new Error('设置每故障域 OSD 数时必须指定 CRUSH 故障域数量'))
      }
      const failureDomain = textValue(getFieldValue('crush_failure_domain'), 'host')
      const available = counts[failureDomain] ?? 0
      if (count > 0 && available > 0 && count > available) {
        return Promise.reject(new Error(`CRUSH 故障域数量不能超过当前可用数量 ${available}`))
      }
      return Promise.resolve()
    }
  })
}

const ecOSDsPerFailureDomainRule = ({ getFieldValue }: { getFieldValue: (name: keyof ErasureCodeProfileFormValues) => unknown }) => ({
  validator(_: unknown, value: unknown) {
    const count = numberValue(value) ?? 0
    const failureDomains = numberValue(getFieldValue('crush_num_failure_domains')) ?? 0
    if (count < 0) {
      return Promise.reject(new Error('每故障域 OSD 数不能小于 0'))
    }
    if (failureDomains > 0 && count < 1) {
      return Promise.reject(new Error('设置 CRUSH 故障域数量时必须指定每故障域 OSD 数'))
    }
    return Promise.resolve()
  }
})

const shecDurabilityRule = ({ getFieldValue }: { getFieldValue: (name: keyof ErasureCodeProfileFormValues) => unknown }) => ({
  validator(_: unknown, value: unknown) {
    const c = numberValue(value) ?? 0
    const m = numberValue(getFieldValue('m')) ?? 0
    if (c < 1 || c > m) {
      return Promise.reject(new Error('耐久性估算值 c 必须在 1 到编码块数 m 之间'))
    }
    return Promise.resolve()
  }
})

const clayHelperChunksRule = ({ getFieldValue }: { getFieldValue: (name: keyof ErasureCodeProfileFormValues) => unknown }) => ({
  validator(_: unknown, value: unknown) {
    const d = numberValue(value) ?? 0
    const k = numberValue(getFieldValue('k')) ?? 0
    const m = numberValue(getFieldValue('m')) ?? 0
    if (d < k + 1 || d > k + m - 1) {
      return Promise.reject(new Error(`Helper chunks 必须在 ${k + 1} 到 ${k + m - 1} 之间`))
    }
    return Promise.resolve()
  }
})

function erasureCodeProfileBody(values: ErasureCodeProfileFormValues, clusterId: number): ApiRecord {
  const body: ApiRecord = {
    cluster_id: clusterId,
    name: values.name,
    plugin: values.plugin,
    k: Math.trunc(values.k),
    m: Math.trunc(values.m),
    'crush-failure-domain': values.crush_failure_domain,
    'crush-root': values.crush_root,
    'crush-device-class': values.crush_device_class || undefined,
    directory: values.directory?.trim() || undefined
  }
  const numFailureDomains = positiveInteger(values.crush_num_failure_domains)
  const osdsPerFailureDomain = positiveInteger(values.crush_osds_per_failure_domain)
  if (numFailureDomains !== undefined) {
    body['crush-num-failure-domains'] = numFailureDomains
  }
  if (osdsPerFailureDomain !== undefined) {
    body['crush-osds-per-failure-domain'] = osdsPerFailureDomain
  }
  if (values.plugin === 'isa' || values.plugin === 'jerasure' || values.plugin === 'clay') {
    body.technique = values.technique
  }
  if (values.plugin === 'jerasure') {
    body.packetsize = positiveInteger(values.packetsize)
  } else if (values.plugin === 'lrc') {
    body.l = positiveInteger(values.l)
    body['crush-locality'] = values.crush_locality
  } else if (values.plugin === 'shec') {
    body.c = positiveInteger(values.c)
  } else if (values.plugin === 'clay') {
    body.d = positiveInteger(values.d)
    body.scalar_mds = values.scalar_mds
  }
  return body
}

function positiveInteger(value?: number) {
  const number = Math.trunc(value ?? 0)
  return number > 0 ? number : undefined
}

function erasureCodeTechniqueOptions(plugin: ErasureCodePlugin, scalarMDS?: ErasureCodeProfileFormValues['scalar_mds']) {
  if (plugin === 'isa' || (plugin === 'clay' && scalarMDS === 'isa')) {
    return ['reed_sol_van', 'cauchy'].map((value) => ({ label: value, value }))
  }
  if (plugin === 'clay' && scalarMDS === 'shec') {
    return ['single', 'multiple'].map((value) => ({ label: value, value }))
  }
  return ['reed_sol_van', 'reed_sol_r6_op', 'cauchy_orig', 'cauchy_good', 'liberation', 'blaum_roth', 'liber8tion']
    .map((value) => ({ label: value, value }))
}

function erasureCodePluginDescription(plugin: ErasureCodePlugin) {
  switch (plugin) {
    case 'jerasure': return '通用且灵活的 Jerasure 纠删码插件。'
    case 'lrc': return '使用局部校验块减少恢复单个丢失 OSD 时所需的数据读取量。'
    case 'shec': return '使用 Shingled Erasure Code 的实验性插件。'
    case 'clay': return '使用 Coupled-Layer 编码降低修复带宽的实验性插件。'
    default: return '使用 Intel Storage Acceleration Library 的 ISA 纠删码插件。'
  }
}

function erasureCodeTechniqueDescription(plugin: ErasureCodePlugin) {
  if (plugin === 'isa') {
    return 'ISA 支持 Reed-Solomon 的 Vandermonde（reed_sol_van）和 Cauchy 两种实现。'
  }
  if (plugin === 'clay') {
    return 'CLAY 底层 Scalar MDS 插件使用的编码技术。'
  }
  return '纠删码插件使用的具体编码技术。'
}

function pgAutoscaleDescription(mode: PoolFormValues['pg_autoscale_mode']) {
  const detail = 'PG 用于在 Ceph 中分布数据，自动伸缩会随每个存储池的使用情况调整 PG 数量。'
  if (mode === 'off') {
    return `禁用此存储池的自动伸缩。${detail}`
  }
  if (mode === 'warn') {
    return `当 PG 数量需要调整时触发健康检查。${detail}`
  }
  return `启用此存储池的 PG 数量自动调整。${detail}`
}

function compressionModeDescription(mode: PoolFormValues['compression_mode']) {
  switch (mode) {
    case 'passive':
      return '仅在写入操作带有可压缩提示时压缩数据。'
    case 'aggressive':
      return '除非写入操作带有不可压缩提示，否则压缩数据。'
    case 'force':
      return '无论写入提示如何，都尝试压缩数据。'
    default:
      return '从不压缩数据。'
  }
}

function rbdMirroringDescription(mode: RbdMirroringMode) {
  switch (mode) {
    case 'pool':
      return '启用基于存储池的 RBD 镜像，池内镜像默认参与同步。'
    default:
      return '禁用此 RBD 存储池的镜像同步。'
  }
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

function poolCompressionAlgorithm(row: ApiRecord): CompressionAlgorithm {
  const algorithm = textValue(row.compression_algorithm, 'snappy')
  return algorithm === 'zlib' || algorithm === 'zstd' || algorithm === 'lz4' ? algorithm : 'snappy'
}

function poolRbdMirroringMode(row: ApiRecord): RbdMirroringMode {
  const mode = textValue(row.rbd_mirroring ?? row.mirror_mode, 'disabled')
  return mode === 'pool' ? mode : 'disabled'
}

function poolHasFlag(row: ApiRecord, flag: string): boolean {
  if (!Array.isArray(row.flags)) {
    return false
  }
  return row.flags.some((value) => textValue(value, '') === flag)
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

function optionalBytes(value: number | undefined, unit: QuotaUnit) {
  if (value === undefined || value === null) {
    return undefined
  }
  return quotaBytes(value, unit)
}

function compressionBytesForUpdate(value: number | undefined, unit: QuotaUnit, previous: number | undefined) {
  if (value === undefined || value === null) {
    return previous === undefined ? undefined : 0
  }
  return quotaBytes(value, unit)
}

function compressionRatioForUpdate(value: number | undefined, previous: number | undefined) {
  if (value === undefined || value === null) {
    return previous === undefined ? undefined : 0
  }
  return value
}

function compressionBody(values: PoolFormValues): ApiRecord {
  if (values.compression_mode === 'none') {
    return {}
  }
  return {
    compression_algorithm: values.compression_algorithm,
    compression_min_blob_size: optionalBytes(values.compression_min_blob_size, values.compression_min_blob_size_unit),
    compression_max_blob_size: optionalBytes(values.compression_max_blob_size, values.compression_max_blob_size_unit),
    compression_required_ratio: values.compression_required_ratio
  }
}

function bytesForForm(value: number | undefined, preferredUnit: QuotaUnit, emptyWhenZero = false): { value?: number, unit: QuotaUnit } {
  if (value === undefined || value <= 0) {
    return { value: emptyWhenZero ? undefined : 0, unit: preferredUnit }
  }
  for (let index = quotaUnits.indexOf(preferredUnit); index >= 0; index -= 1) {
    const divisor = 1024 ** index
    if (value % divisor === 0) {
      return { value: value / divisor, unit: quotaUnits[index] }
    }
  }
  return { value, unit: 'B' }
}

function defaultRbdPoolConfiguration(): Record<RbdPoolConfigurationKey, number> {
  return Object.fromEntries(rbdPoolConfigurationFields.map(({ key }) => [key, 0])) as Record<RbdPoolConfigurationKey, number>
}

function normalizedRbdPoolConfiguration(configuration: Partial<Record<RbdPoolConfigurationKey, number>> | undefined) {
  const result = defaultRbdPoolConfiguration()
  rbdPoolConfigurationFields.forEach(({ key }) => {
    result[key] = Math.trunc(Math.max(0, numberValue(configuration?.[key]) ?? 0))
  })
  return result
}

function poolRbdConfiguration(row: ApiRecord): Record<RbdPoolConfigurationKey, number> {
  const result = defaultRbdPoolConfiguration()
  if (!Array.isArray(row.configuration)) {
    return result
  }
  row.configuration.forEach((entry) => {
    if (!isRecord(entry)) {
      return
    }
    const key = textValue(entry.name ?? entry.key ?? entry.option, '') as RbdPoolConfigurationKey
    if (!rbdPoolConfigurationFields.some((field) => field.key === key)) {
      return
    }
    result[key] = Math.trunc(Math.max(0, numberValue(entry.value ?? entry.val ?? entry.default) ?? 0))
  })
  return result
}

function blobSizeRule(kind: 'min' | 'max') {
  return ({ getFieldValue }: { getFieldValue: (name: keyof PoolFormValues) => unknown }) => ({
    validator(_: unknown, value: unknown) {
      const minValue = kind === 'min' ? value : getFieldValue('compression_min_blob_size')
      const minUnit = getFieldValue('compression_min_blob_size_unit') as QuotaUnit
      const maxValue = kind === 'max' ? value : getFieldValue('compression_max_blob_size')
      const maxUnit = getFieldValue('compression_max_blob_size_unit') as QuotaUnit
      const minBytes = optionalBytes(numberValue(minValue), minUnit)
      const maxBytes = optionalBytes(numberValue(maxValue), maxUnit)
      if (minBytes !== undefined && maxBytes !== undefined && minBytes > maxBytes) {
        return Promise.reject(new Error(kind === 'min' ? '最小 Blob 大小不能大于最大 Blob 大小' : '最大 Blob 大小不能小于最小 Blob 大小'))
      }
      return Promise.resolve()
    }
  })
}

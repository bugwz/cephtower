import { PlusOutlined, ReloadOutlined, SaveOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Dropdown, Drawer, Form, Input, InputNumber, Modal, Select, Space, Switch, Typography } from 'antd'
import type { ColumnsType, TableProps } from 'antd/es/table'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { listResource, listResourceFilterOptions, mutateResource, refreshResource, type ResourceListResult } from '../api/resource'
import { textValue, type ApiRecord } from '../api/client'
import type { OperationRisk } from '../api/types'
import type { FieldColumn } from '../components/DataTable'
import { AppTable } from '../components/AppTable'
import { DraggableModal } from '../components/DraggableModal'
import { Page } from '../components/Page'
import { RecordDetail } from '../components/RecordDetail'
import { ResourceMetaBar } from '../components/ResourceMetaBar'
import { TableAction, TableActions } from '../components/TableActions'
import { useResource } from '../hooks'
import { useFeatureRequirements, type FeatureRequirements } from '../hooks/useFeatureRequirements'
import { useMutationOperation } from '../hooks/useMutationOperation'
import { useClusterContext } from '../state/ClusterContext'
import { message } from '../utils/appMessage'

const { Text } = Typography

export interface MutationFormField {
  name: string
  label: string
  type?: 'text' | 'number' | 'boolean' | 'select' | 'textarea'
  required?: boolean
  visibleWhen?: (values: MutationFormValues) => boolean
  placeholder?: string
  options?: Array<{ label: string; value: string | number | boolean }>
  optionsLoader?: (clusterId: number) => Promise<Array<{ label: string; value: string | number | boolean }>>
  min?: number
  max?: number
  pattern?: RegExp
  patternMessage?: string
}

export type MutationFormValues = Record<string, string | number | boolean | null | undefined | ApiRecord>

export interface ResourceFormAction {
  title: string
  buttonLabel?: string
  path: string
  method: 'POST' | 'PATCH' | 'PUT'
  successMessage: string
  confirmation?: (values: MutationFormValues, row?: ApiRecord) => string | undefined
  fields: MutationFormField[]
  initialValues?: MutationFormValues | ((row?: ApiRecord) => MutationFormValues)
  buildBody: (values: MutationFormValues, clusterId: number, row?: ApiRecord) => ApiRecord
}

export interface ResourceDeleteAction {
  title: string
  path: string
  action: string
  resourceKind: string
  risk?: OperationRisk
  successMessage: string
  buildBody: (row: ApiRecord, clusterId: number) => ApiRecord
  resourceKey: (row: ApiRecord) => string
}

export interface ResourceListPageDefinition extends FeatureRequirements {
  title: string
  path: string
  columns: FieldColumn[]
  body?: ApiRecord
  rowKeyCandidates?: string[]
  createAction?: ResourceFormAction
  updateAction?: ResourceFormAction
  extraActions?: ResourceFormAction[]
  deleteAction?: ResourceDeleteAction
  detailPath?: (row: ApiRecord) => string
}

export function ResourceListPage({ definition, embedded = false }: { definition: ResourceListPageDefinition; embedded?: boolean }) {
  const navigate = useNavigate()
  const { selectedClusterId } = useClusterContext()
  const [refreshing, setRefreshing] = useState(false)
  const [formOpen, setFormOpen] = useState(false)
  const [activeAction, setActiveAction] = useState<ResourceFormAction | null>(null)
  const [activeRow, setActiveRow] = useState<ApiRecord | undefined>()
  const [detailRow, setDetailRow] = useState<ApiRecord | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [dynamicOptions, setDynamicOptions] = useState<Record<string, MutationFormField['options']>>({})
  const [columnFilters, setColumnFilters] = useState<Record<string, string[]>>({})
  const [filterOptions, setFilterOptions] = useState<Record<string, string[]>>({})
  const [form] = Form.useForm<MutationFormValues>()
  const formValues = Form.useWatch([], form) as MutationFormValues | undefined
  const operationMutation = useMutationOperation()
  const filterFields = useMemo(() => Array.from(new Set(definition.columns.map((column) => column.key))), [definition.columns])
  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return emptyResult()
    }
    return listResource(definition.path, selectedClusterId, { body: definition.body, filters: columnFilters })
  }, [columnFilters, definition.body, definition.path, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const featureStatus = useFeatureRequirements(selectedClusterId, definition)
  const mutationBlocked = featureStatus.loading || featureStatus.blocked || Boolean(featureStatus.error)

  useEffect(() => {
    let ignore = false
    async function loadOptions() {
      if (!selectedClusterId || filterFields.length === 0) {
        setFilterOptions({})
        return
      }
      const options = await listResourceFilterOptions(definition.path, filterFields, selectedClusterId, definition.body)
      if (!ignore) {
        setFilterOptions(options)
      }
    }
    void loadOptions().catch(() => {
      if (!ignore) {
        setFilterOptions({})
      }
    })
    return () => {
      ignore = true
    }
  }, [definition.body, definition.path, filterFields, selectedClusterId])

  async function reload() {
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    setRefreshing(true)
    try {
      const kinds = refreshKinds(definition)
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, ...(kinds.length === 1 ? { kind: kinds[0] } : { kinds }) }), '刷新成功')
      await refresh()
    } finally {
      setRefreshing(false)
    }
  }

  function openForm(action: ResourceFormAction, row?: ApiRecord) {
    if (mutationBlocked) {
      message.warning('当前集群未满足该功能的操作依赖')
      return
    }
    setActiveAction(action)
    setActiveRow(row)
    form.resetFields()
    const initialValues = typeof action.initialValues === 'function' ? action.initialValues(row) : action.initialValues
    form.setFieldsValue({ ...initialValues })
    setDynamicOptions({})
    setFormOpen(true)
    if (selectedClusterId) {
      action.fields.forEach((field) => {
        if (!field.optionsLoader) {
          return
        }
        void field.optionsLoader(selectedClusterId).then((options) => {
          setDynamicOptions((current) => ({ ...current, [field.name]: options }))
        }).catch(() => {
          setDynamicOptions((current) => ({ ...current, [field.name]: [] }))
        })
      })
    }
  }

  async function submitForm(values: MutationFormValues) {
    if (!selectedClusterId || !activeAction || submitting || mutationBlocked) {
      return
    }
    const action = activeAction
    setSubmitting(true)
    try {
      const confirmation = action.confirmation?.(values, activeRow)
      if (confirmation) {
        const approved = await new Promise<boolean>((resolve) => {
          Modal.confirm({ title: action.title, content: confirmation, okText: '确认执行', okType: 'danger', cancelText: '取消', onOk: () => resolve(true), onCancel: () => resolve(false), afterClose: () => resolve(false) })
        })
        if (!approved) return
      }
      await operationMutation.run(() => mutateResource(
        action.path,
        action.method,
        action.buildBody(values, selectedClusterId, activeRow),
        activeRow?.resource_version ? { ifMatch: String(activeRow.resource_version) } : undefined
      ), false)
      setFormOpen(false)
      message.success(action.successMessage)
      if (action.path.startsWith('/rbd/')) {
        await refreshResource({ clusterId: selectedClusterId, kinds: ['rbd_image', 'rbd_snapshot', 'rbd_namespace', 'rbd_trash', 'rbd_group', 'rbd_mirroring'] })
      }
      if (action.path === '/rgw/zone') await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_zone','rgw_zonegroup'] })
      if (action.path === '/rgw/zonegroup') await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_zonegroup'] })
      if (action.path === '/rgw/realm') await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_realm','rgw_status'] })
      if (action.path.startsWith('/rgw/account')) await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_account','rgw_role'] })
      if (action.path.startsWith('/rgw/role')) await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_role'] })
      if ((action.path === '/rgw/bucket/ratelimit' || action.path === '/rgw/bucket/quota')) await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_bucket'] })
      if (action.path.startsWith('/rgw/user')) await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_user'] })
      await refresh({ showLoading: false })
    } finally {
      setSubmitting(false)
    }
  }

  async function deleteRow(row: ApiRecord) {
    if (!selectedClusterId || !definition.deleteAction || mutationBlocked) {
      return
    }
    const action = definition.deleteAction
    const resourceKey = action.resourceKey(row)
    if (!resourceKey) {
      message.error('无法识别资源标识')
      return
    }
    const generation = Number(row.resource_version ?? 0)
    const parameters = action.buildBody(row, selectedClusterId)
    if (action.risk && action.risk !== 'high') {
      Modal.confirm({
        title: `${action.title} ${resourceKey}`,
        content: '确认后将直接执行该操作。',
        okText: '提交',
        okType: action.risk === 'medium' ? 'danger' : 'primary',
        cancelText: '取消',
        async onOk() {
          await operationMutation.run(() => mutateResource(action.path, 'DELETE', parameters), false)
          message.success(action.successMessage)
          if (action.path.startsWith('/rbd/')) await refreshResource({ clusterId:selectedClusterId,kinds:['rbd_image','rbd_snapshot','rbd_namespace','rbd_trash','rbd_group'] })
          if (action.path === '/rgw/zone') await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_zone','rgw_zonegroup'] })
      if (action.path === '/rgw/zonegroup') await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_zonegroup'] })
      if (action.path === '/rgw/realm') await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_realm','rgw_status'] })
      if (action.path.startsWith('/rgw/account')) await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_account','rgw_role'] })
      if (action.path.startsWith('/rgw/role')) await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_role'] })
      if ((action.path === '/rgw/bucket/ratelimit' || action.path === '/rgw/bucket/quota')) await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_bucket'] })
      if (action.path.startsWith('/rgw/user')) await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_user'] })
          await refresh({ showLoading:false })
        }
      })
      return
    }
    Modal.confirm({
      title: `${action.title} ${resourceKey}`,
      content: '该操作为高风险操作，确认后将直接执行操作。',
      okText: '提交',
      okType: 'danger',
      cancelText: '取消',
      async onOk() {
        await operationMutation.run(() => mutateResource(action.path, 'DELETE', parameters, { ifMatch: generation }), false)
        message.success(action.successMessage)
        if (action.path.startsWith('/rbd/')) await refreshResource({ clusterId:selectedClusterId,kinds:['rbd_image','rbd_snapshot','rbd_namespace','rbd_trash','rbd_group'] })
          if (action.path === '/rgw/zone') await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_zone','rgw_zonegroup'] })
      if (action.path === '/rgw/zonegroup') await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_zonegroup'] })
      if (action.path === '/rgw/realm') await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_realm','rgw_status'] })
      if (action.path.startsWith('/rgw/account')) await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_account','rgw_role'] })
      if (action.path.startsWith('/rgw/role')) await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_role'] })
      if ((action.path === '/rgw/bucket/ratelimit' || action.path === '/rgw/bucket/quota')) await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_bucket'] })
      if (action.path.startsWith('/rgw/user')) await refreshResource({ clusterId:selectedClusterId,kinds:['rgw_user'] })
        await refresh({ showLoading:false })
      }
    })
  }

  const handleTableChange: TableProps<ApiRecord>['onChange'] = (_pagination, filters) => {
    setColumnFilters(tableFilters(filters))
  }
  const tableColumns = buildColumns(
    definition,
    openForm,
    deleteRow,
    (row) => definition.detailPath ? navigate(definition.detailPath(row)) : setDetailRow(row),
    mutationBlocked,
    filterOptions,
    columnFilters
  )
  const listActions = (
    <Space>
      <Button icon={<ReloadOutlined />} loading={refreshing} onClick={reload}>
        刷新
      </Button>
      {definition.createAction ? (
        <Button
          type="primary"
          icon={<PlusOutlined />}
          disabled={!selectedClusterId || mutationBlocked}
          onClick={() => openForm(definition.createAction!)}
        >
          {definition.createAction.buttonLabel ?? '新建'}
        </Button>
      ) : null}
    </Space>
  )
  const featureRequirementAlert = <FeatureRequirementAlert status={featureStatus} />
  const resourceTable = (
    <AppTable<ApiRecord>
      size="middle"
      columns={tableColumns}
      dataSource={data?.items ?? []}
      locale={{ emptyText: '暂无数据' }}
      pagination={{ defaultPageSize: 10, showSizeChanger: true }}
      onChange={handleTableChange}
      rowKey={(row, index) => rowKey(definition, row, index)}
      scroll={{ x: true }}
      footer={() => <ResourceMetaBar observedAt={data?.observedAt} stale={data?.stale} staleReason={data?.staleReason} />}
    />
  )
  const listContent = hasFeatureRequirementAlert(featureStatus) ? (
    <Space direction="vertical" size={16} className="page-stack">
      {featureRequirementAlert}
      {resourceTable}
    </Space>
  ) : resourceTable
  const listSurface = embedded ? (
    <div className="page-embedded-list">
      <div className="page-embedded-list-head">
        <Text strong>{definition.title}</Text>
        <div className="page-embedded-list-actions">{listActions}</div>
      </div>
      <div className="page-embedded-list-body">{listContent}</div>
    </div>
  ) : (
    <Card className="page-surface-card" title={definition.title} extra={listActions}>
      {listContent}
    </Card>
  )

  return (
    <Page title={definition.title} loading={loading} error={error} stateVariant={embedded ? 'inline' : 'card'}>
      {listSurface}

      <DraggableModal
        title={activeAction?.title ?? ''}
        open={formOpen}
        onCancel={() => setFormOpen(false)}
        onOk={() => form.submit()}
        okText="提交"
        confirmLoading={submitting}
        okButtonProps={{ icon: <SaveOutlined />, disabled: mutationBlocked }}
        destroyOnClose
      >
        <Form form={form} layout="vertical" onFinish={submitForm}>
          {activeAction?.fields.filter((field) => !field.visibleWhen || field.visibleWhen(formValues ?? {})).map((field) => (
            <Form.Item
              key={field.name}
              preserve={false}
              name={field.name}
              label={field.label}
              valuePropName={field.type === 'boolean' ? 'checked' : 'value'}
              rules={[
                ...(field.required ? [{ required: true, message: `请输入${field.label}` }] : []),
                ...(field.pattern ? [{ pattern: field.pattern, message: field.patternMessage ?? `${field.label}格式不正确` }] : [])
              ]}
            >
              {renderFormControl({ ...field, options: dynamicOptions[field.name] ?? field.options })}
            </Form.Item>
          ))}
        </Form>
      </DraggableModal>
      {!definition.detailPath ? (
        <Drawer
          title={`${definition.title}详情`}
          open={Boolean(detailRow)}
          onClose={() => setDetailRow(null)}
          width={720}
          destroyOnClose
        >
          <RecordDetail record={detailRow} />
        </Drawer>
      ) : null}
    </Page>
  )
}

function buildColumns(
  definition: ResourceListPageDefinition,
  openForm: (action: ResourceFormAction, row?: ApiRecord) => void,
  deleteRow: (row: ApiRecord) => void,
  openDetail: (row: ApiRecord) => void,
  mutationBlocked: boolean,
  filterOptions: Record<string, string[]>,
  columnFilters: Record<string, string[]>
) {
  const columns: ColumnsType<ApiRecord> = definition.columns.map((column) => ({
    title: column.title,
    dataIndex: column.key,
    key: column.key,
    ellipsis: true,
    filterMultiple: true,
    filterSearch: true,
    filters: (filterOptions[column.key] ?? []).map((value) => ({ text: value, value })),
    filteredValue: columnFilters[column.key] ?? null,
    render: (value, row) => column.render?.(value, row) ?? (value !== null && typeof value === 'object' && (!Array.isArray(value) || value.some((item) => item !== null && typeof item === 'object')) ? <Button type="link" size="small" onClick={() => openDetail({ [column.key]: value })}>{Array.isArray(value) ? `查看 ${value.length} 项` : '查看详情'}</Button> : renderValue(value))
  }))

  columns.push({
    title: '操作',
    key: 'actions',
    width: 65 + (definition.updateAction ? 50 : 0) + (definition.deleteAction ? 50 : 0) + (definition.extraActions?.length ? 80 : 0),
    fixed: 'right',
    render: (_, row) => (
      <TableActions>
        <TableAction onClick={() => openDetail(row)}>详情</TableAction>
        {definition.updateAction ? (
          <TableAction disabled={mutationBlocked} onClick={() => openForm(definition.updateAction!, row)}>编辑</TableAction>
        ) : null}
        {definition.extraActions?.length ? (
          <Dropdown
            trigger={['click']}
            disabled={mutationBlocked}
            menu={{
              items: definition.extraActions.map((action, index) => ({ key: String(index), label: action.title, disabled: mutationBlocked })),
              onClick: ({ key }) => {
                const action = definition.extraActions?.[Number(key)]
                if (action && !mutationBlocked) openForm(action, row)
              }
            }}
          >
            <Button type="link" size="small" disabled={mutationBlocked}>更多操作</Button>
          </Dropdown>
        ) : null}
        {definition.deleteAction ? (
          <TableAction danger disabled={mutationBlocked} onClick={() => deleteRow(row)}>删除</TableAction>
        ) : null}
      </TableActions>
    )
  })

  return columns
}

function refreshKinds(definition: ResourceListPageDefinition) {
  if (definition.deleteAction?.resourceKind) {
    return [definition.deleteAction.resourceKind]
  }
  const pathKinds: Record<string, string[]> = {
    '/configuration/values': ['config_value'],
    '/crush/rules': ['crush_rule'],
    '/erasure/code/profiles': ['erasure_code_profile'],
    '/filesystems': ['filesystem'],
    '/filesystem/authorizations': ['cephfs_authorization'],
    '/filesystem/clients': ['cephfs_client'],
    '/filesystem/entries': ['cephfs_entry'],
    '/filesystem/subvolume/groups': ['subvolume_group'],
    '/filesystem/subvolume/snapshots': ['cephfs_snapshot'],
    '/filesystem/subvolumes': ['subvolume'],
    '/nfs/clusters': ['nfs_cluster'],
    '/nfs/exports': ['nfs_export'],
    '/rbd/groups': ['rbd_group'],
    '/rbd/mirroring': ['rbd_mirroring'],
    '/rbd/images': ['rbd_image'],
    '/rbd/namespaces': ['rbd_namespace'],
    '/rbd/image/snapshots': ['rbd_snapshot'],
    '/rbd/trash': ['rbd_trash'],
    '/rgw/accounts': ['rgw_account'],
    '/rgw/buckets': ['rgw_bucket'],
    '/rgw/realms': ['rgw_realm'],
    '/rgw/roles': ['rgw_role'],
    '/rgw/users': ['rgw_user'],
    '/rgw/zonegroups': ['rgw_zonegroup'],
    '/rgw/zones': ['rgw_zone'],
    '/smb/clusters': ['smb_cluster'],
    '/smb/shares': ['smb_share']
  }
  return pathKinds[definition.path] ?? []
}

function FeatureRequirementAlert({ status }: { status: ReturnType<typeof useFeatureRequirements> }) {
  if (status.loading) {
    return <Alert type="info" showIcon message="正在校验当前集群的功能依赖" />
  }
  if (status.error) {
    return <Alert type="warning" showIcon message="功能依赖检查失败" description={status.error} />
  }
  if (status.reasons.length) {
    return <Alert type="warning" showIcon message="当前集群暂不可执行该页面的变更操作" description={status.reasons.join('; ')} />
  }
  return null
}

function hasFeatureRequirementAlert(status: ReturnType<typeof useFeatureRequirements>) {
  return status.loading || Boolean(status.error) || status.reasons.length > 0
}

function renderFormControl(field: MutationFormField) {
  if (field.type === 'number') {
    return <InputNumber min={field.min} max={field.max} className="full-width-control" />
  }
  if (field.type === 'boolean') {
    return <Switch />
  }
  if (field.type === 'select') {
    return <Select options={field.options ?? []} />
  }
  if (field.type === 'textarea') {
    return <Input.TextArea rows={5} spellCheck={false} placeholder={field.placeholder} />
  }
  return <Input placeholder={field.placeholder} />
}

function renderValue(value: unknown) {
  if (Array.isArray(value)) {
    return value.length ? value.map((item) => textValue(item)).join(', ') : '-'
  }
  return textValue(value, '-')
}

function tableFilters(filters: Record<string, unknown>) {
  return Object.fromEntries(
    Object.entries(filters)
      .map(([field, values]) => [field, Array.isArray(values) ? values.map(String).filter(Boolean) : []] as const)
      .filter(([, values]) => values.length > 0)
  )
}

function rowKey(definition: ResourceListPageDefinition, row: ApiRecord, index?: number) {
  const candidates = definition.rowKeyCandidates ?? ['natural_key', 'name', 'id']
  return candidates.map((key) => row[key]).find(Boolean)?.toString() ?? String(index)
}

function emptyResult(): ResourceListResult<ApiRecord> {
  return {
    items: [],
    nextCursor: null,
    observedAt: null,
    stale: false,
    staleReason: null
  }
}

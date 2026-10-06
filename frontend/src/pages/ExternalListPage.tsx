import { PlusOutlined, ReloadOutlined, SaveOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Drawer, Form, Input, InputNumber, Modal, Select, Space, Switch, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import type { Key, ReactNode } from 'react'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { readExternalList } from '../api/external'
import { mutateResource } from '../api/resource'
import type { ApiRecord } from '../api/client'
import { textValue } from '../api/client'
import type { FieldColumn } from '../components/DataTable'
import { AppTable } from '../components/AppTable'
import { DraggableModal } from '../components/DraggableModal'
import { Page } from '../components/Page'
import { RecordDetail } from '../components/RecordDetail'
import { TableAction, TableActions } from '../components/TableActions'
import { useResource } from '../hooks'
import { useFeatureRequirements, type FeatureRequirements } from '../hooks/useFeatureRequirements'
import { useMutationOperation } from '../hooks/useMutationOperation'
import { useClusterContext } from '../state/ClusterContext'
import { message } from '../utils/appMessage'
import { startAutoRefresh } from './monitoring/autoRefresh'
import type { MutationFormField, MutationFormValues, ResourceDeleteAction, ResourceFormAction } from './ResourceListPage'

const { Text } = Typography

export interface ExternalListColumn extends FieldColumn {
  localFilter?: { options: string[]; value: (row: ApiRecord) => string | string[] | undefined; defaultValues?: string[] }
}

export interface ExternalListPageDefinition extends FeatureRequirements {
  title: string
  path: string
  body?: ApiRecord
  buildQuery?: (body: ApiRecord) => URLSearchParams
  filterFields?: MutationFormField[]
  columns: ExternalListColumn[]
  rowKeyCandidates?: string[]
  autoRefreshMs?: number
  createAction?: ResourceFormAction
  updateAction?: ResourceFormAction
  extraActions?: Array<ResourceFormAction & { visibleWhen?: (row: ApiRecord) => boolean }>
  deleteAction?: ResourceDeleteAction
  detailContent?: (row: ApiRecord, clusterId: number) => ReactNode
}

export function ExternalListPage({ definition, embedded = false }: { definition: ExternalListPageDefinition; embedded?: boolean }) {
  const { selectedClusterId } = useClusterContext()
  const currentClusterId = useRef(selectedClusterId)
  const clusterGeneration = useRef(0)
  if (currentClusterId.current !== selectedClusterId) clusterGeneration.current += 1
  currentClusterId.current = selectedClusterId
  useEffect(() => () => { clusterGeneration.current += 1 }, [])
  const [formClusterId, setFormClusterId] = useState<number | undefined>()
  const [formGeneration, setFormGeneration] = useState<number | undefined>()
  const [refreshing, setRefreshing] = useState(false)
  const [autoRefresh, setAutoRefresh] = useState(true)
  const [formOpen, setFormOpen] = useState(false)
  const [activeAction, setActiveAction] = useState<ResourceFormAction | null>(null)
  const [activeRow, setActiveRow] = useState<ApiRecord | undefined>()
  const [detailRow, setDetailRow] = useState<ApiRecord | null>(null)
  const [detailGeneration, setDetailGeneration] = useState<number>()
  const visibleDetail = detailGeneration === clusterGeneration.current ? detailRow : null
  const [submitting, setSubmitting] = useState(false)
  const [queryBody, setQueryBody] = useState<ApiRecord>(definition.body ?? {})
  const [form] = Form.useForm<MutationFormValues>()
  const [filterForm] = Form.useForm<ApiRecord>()
  const operationMutation = useMutationOperation()
  const missingRequiredFilters = useMemo(() => requiredFilterLabels(definition.filterFields, queryBody), [definition.filterFields, queryBody])
  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return []
    }
    if (missingRequiredFilters.length > 0) {
      return []
    }
    const payload = await readExternalList(definition.path, selectedClusterId, queryBody, definition.buildQuery?.(queryBody))
    return payload.items
  }, [definition.path, definition.buildQuery, missingRequiredFilters.length, queryBody, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const featureStatus = useFeatureRequirements(selectedClusterId, definition)
  const mutationBlocked = featureStatus.loading || featureStatus.blocked || Boolean(featureStatus.error)

  useEffect(() => {
    const delay = definition.autoRefreshMs
    if (!delay || delay < 1000 || !autoRefresh || !selectedClusterId || loading || refreshing || formOpen || visibleDetail || submitting || mutationBlocked || missingRequiredFilters.length) return
    return startAutoRefresh(() => refresh({ showLoading: false }), delay)
  }, [definition.autoRefreshMs, autoRefresh, selectedClusterId, loading, refreshing, formOpen, visibleDetail, submitting, mutationBlocked, missingRequiredFilters.length, refresh])

  useEffect(() => {
    form.resetFields()
    setFormOpen(false)
    setActiveAction(null)
    setActiveRow(undefined)
    setDetailRow(null)
    setFormClusterId(undefined)
  }, [form, selectedClusterId])

  async function reload() {
    setRefreshing(true)
    try {
      await refresh()
    } finally {
      setRefreshing(false)
    }
  }

  function applyFilters(values: ApiRecord) {
    // Invalidate pending confirmations before React commits the new query.
    clusterGeneration.current += 1
    form.resetFields()
    setFormOpen(false)
    setActiveAction(null)
    setActiveRow(undefined)
    setDetailRow(null)
    setFormClusterId(undefined)
    setQueryBody(cleanRecord(values))
  }

  function openForm(action: ResourceFormAction, row?: ApiRecord) {
    if (mutationBlocked) {
      message.warning('当前集群未满足该功能的操作依赖')
      return
    }
    const blocked = row ? action.disabledWhen?.(row) : undefined
    if (blocked) {
      message.warning(blocked)
      return
    }
    setActiveAction(action)
    setFormGeneration(clusterGeneration.current)
    setFormClusterId(selectedClusterId)
    setActiveRow(row)
    form.resetFields()
    const initialValues = typeof action.initialValues === 'function' ? action.initialValues(row) : action.initialValues
    form.setFieldsValue({ ...initialValues })
    setFormOpen(true)
  }

  async function submitForm(values: MutationFormValues) {
    if (!selectedClusterId || formClusterId !== selectedClusterId || formGeneration !== clusterGeneration.current || !activeAction || submitting || mutationBlocked) {
      return
    }
    const action = activeAction
    const generation = clusterGeneration.current
    const blocked = activeRow ? action.disabledWhen?.(activeRow) : undefined
    if (blocked) {
      message.warning(blocked)
      return
    }
    setSubmitting(true)
    try {
      let parameters: ApiRecord
      let confirmation: string | undefined
      try {
        parameters = action.buildBody(values, selectedClusterId, activeRow)
        confirmation = action.confirmation?.(values, activeRow)
      } catch (error) {
        message.error(error instanceof Error ? error.message : '表单参数无效')
        return
      }
      if (confirmation) {
        const approved = await new Promise<boolean>((resolve) => {
          Modal.confirm({ title: action.title, content: confirmation, okText: '确认执行', okType: 'danger', cancelText: '取消', onOk: () => resolve(true), onCancel: () => resolve(false), afterClose: () => resolve(false) })
        })
        if (!approved) return
      }
      if (currentClusterId.current !== formClusterId || clusterGeneration.current !== generation) return
      await operationMutation.run(() => mutateResource(
        action.path,
        action.method,
        parameters
      ), false)
      if (currentClusterId.current !== formClusterId || clusterGeneration.current !== generation) return
      form.resetFields()
      setFormOpen(false)
      message.success(action.successMessage)
      void refresh({ showLoading: false })
    } finally {
      setSubmitting(false)
    }
  }

  async function deleteRow(row: ApiRecord) {
    if (!selectedClusterId || !definition.deleteAction || mutationBlocked) {
      return
    }
    const action = definition.deleteAction
    const blocked = action.disabledWhen?.(row)
    if (blocked) {
      message.warning(blocked)
      return
    }
    const resourceKey = action.resourceKey(row)
    if (!resourceKey) {
      message.error('无法识别资源标识')
      return
    }
    const parameters = action.buildBody(row, selectedClusterId)
    const generation = clusterGeneration.current
    const isCurrentScope = () => currentClusterId.current === selectedClusterId && clusterGeneration.current === generation
    if (action.risk && action.risk !== 'high') {
      Modal.confirm({
        title: `${action.title} ${resourceKey}`,
        content: action.confirmation?.(row) ?? '确认后将直接执行该操作。',
        okText: '提交',
        okType: action.risk === 'medium' ? 'danger' : 'primary',
        cancelText: '取消',
        async onOk() {
          if (!isCurrentScope()) return
          await operationMutation.run(() => mutateResource(action.path, 'DELETE', parameters), false)
          window.setTimeout(() => {
            if (!isCurrentScope()) return
            message.success(action.successMessage)
            void refresh({ showLoading: false })
          })
        }
      })
      return
    }
    Modal.confirm({
      title: `${action.title} ${resourceKey}`,
      content: action.confirmation?.(row) ?? '该操作为高风险操作，确认后将直接执行操作。',
      okText: '提交',
      okType: 'danger',
      cancelText: '取消',
      async onOk() {
        if (!isCurrentScope()) return
        await operationMutation.run(() => mutateResource(action.path, 'DELETE', parameters), false)
        window.setTimeout(() => {
          if (!isCurrentScope()) return
          message.success(action.successMessage)
          void refresh({ showLoading: false })
        })
      }
    })
  }

  const tableColumns = buildColumns(definition, openForm, deleteRow, (row) => { setDetailGeneration(clusterGeneration.current); setDetailRow(row) }, mutationBlocked)
  const listActions = (
    <Space>
      {definition.autoRefreshMs && <Switch checked={autoRefresh} onChange={setAutoRefresh} checkedChildren="自动刷新" unCheckedChildren="自动刷新已关闭" aria-label="自动刷新（打开表单或详情时暂停）" />}
      <Button icon={<ReloadOutlined />} loading={refreshing} onClick={reload}>刷新</Button>
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
  const filterFormContent = definition.filterFields?.length ? (
    <Form
      form={filterForm}
      layout="inline"
      initialValues={definition.body}
      onFinish={applyFilters}
    >
      {definition.filterFields.map((field) => (
        <Form.Item
          key={field.name}
          name={field.name}
          label={field.label}
          rules={field.required ? [{ required: true, message: `请输入${field.label}` }] : undefined}
        >
          {renderFormControl(field)}
        </Form.Item>
      ))}
      <Form.Item>
        <Button icon={<ReloadOutlined />} htmlType="submit">应用</Button>
      </Form.Item>
    </Form>
  ) : null
  const missingFilterAlert = missingRequiredFilters.length > 0 ? (
    <Alert
      type="info"
      showIcon
      message="请先填写筛选条件"
      description={`该页面需要 ${missingRequiredFilters.join('、')} 后再读取数据。`}
    />
  ) : null
  const externalTable = (
    <AppTable<ApiRecord>
      key={JSON.stringify([selectedClusterId, definition.path])}
      size="middle"
      columns={tableColumns}
      dataSource={data ?? []}
      pagination={{ defaultPageSize: 10, showSizeChanger: true }}
      rowKey={(row, index) => rowKey(definition, row, index)}
      scroll={{ x: true }}
    />
  )
  const hasListExtras = hasFeatureRequirementAlert(featureStatus) || Boolean(filterFormContent) || Boolean(missingFilterAlert)
  const listContent = hasListExtras ? (
    <Space direction="vertical" size={16} className="page-stack">
      {featureRequirementAlert}
      {filterFormContent}
      {missingFilterAlert}
      {externalTable}
    </Space>
  ) : externalTable
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
    <Page title={definition.title} loading={loading} error={error} stateVariant={embedded ? 'inline' : 'card'}
      errorAction={<Space direction="vertical" style={{ marginTop: 16 }}>{filterFormContent}<Button icon={<ReloadOutlined />} loading={refreshing} disabled={refreshing || !selectedClusterId} onClick={reload}>重新读取</Button></Space>}>
      {listSurface}

      <DraggableModal
        title={activeAction?.title ?? ''}
        open={formOpen}
        onCancel={() => { if (!submitting) setFormOpen(false) }}
        cancelButtonProps={{ disabled: submitting }}
        closable={!submitting}
        maskClosable={!submitting}
        keyboard={!submitting}
        onOk={() => form.submit()}
        okText="提交"
        confirmLoading={submitting}
        okButtonProps={{ icon: <SaveOutlined />, disabled: mutationBlocked || formClusterId !== selectedClusterId }}
        destroyOnClose
      >
        <Form form={form} layout="vertical" onFinish={submitForm} disabled={submitting}>
          {activeAction?.fields.map((field) => (
            <Form.Item
              key={field.name}
              name={field.name}
              label={field.label}
              valuePropName={field.type === 'boolean' ? 'checked' : 'value'}
              rules={field.required ? [{ required: true, message: `请输入${field.label}` }] : undefined}
            >
              {renderFormControl(field, submitting)}
            </Form.Item>
          ))}
        </Form>
      </DraggableModal>
      <Drawer
        title={`${definition.title}详情`}
        open={Boolean(visibleDetail)}
        onClose={() => setDetailRow(null)}
        width={720}
        destroyOnClose
      >
        <RecordDetail record={visibleDetail} />
        {visibleDetail && selectedClusterId && definition.detailContent?.(visibleDetail, selectedClusterId)}
      </Drawer>
    </Page>
  )
}

function cleanRecord(values: ApiRecord) {
  return Object.fromEntries(Object.entries(values).filter(([, value]) => value !== undefined && value !== null && value !== ''))
}

function requiredFilterLabels(fields: MutationFormField[] | undefined, values: ApiRecord) {
  return (fields ?? [])
    .filter((field) => field.required && isEmptyFilterValue(values[field.name]))
    .map((field) => field.label)
}

function isEmptyFilterValue(value: unknown) {
  return value === undefined || value === null || value === ''
}

function buildColumns(
  definition: ExternalListPageDefinition,
  openForm: (action: ResourceFormAction, row?: ApiRecord) => void,
  deleteRow: (row: ApiRecord) => void,
  openDetail: (row: ApiRecord) => void,
  mutationBlocked: boolean
) {
  const columns: ColumnsType<ApiRecord> = definition.columns.map((column) => ({
    title: column.title,
    dataIndex: column.key,
    key: column.key,
    ellipsis: true,
    ...(column.localFilter ? {
      filters: column.localFilter.options.map(value => ({ text: value, value })),
      defaultFilteredValue: column.localFilter.defaultValues,
      onFilter: (value: Key | boolean, row: ApiRecord) => {
        const actual = column.localFilter!.value(row)
        return Array.isArray(actual) ? typeof value === 'string' && actual.includes(value) : actual === value
      }
    } : {}),
    render: (value, row) => column.render?.(value, row) ?? renderValue(value)
  }))

  columns.push({
    title: '操作',
    key: 'actions',
    width: definition.extraActions?.length ? 250 : definition.updateAction && definition.deleteAction ? 150 : 110,
    fixed: 'right',
    render: (_, row) => (
      <TableActions>
        <TableAction onClick={() => openDetail(row)}>详情</TableAction>
        {definition.updateAction ? (
          <TableAction disabled={mutationBlocked || Boolean(definition.updateAction.disabledWhen?.(row))} onClick={() => openForm(definition.updateAction!, row)}>编辑</TableAction>
        ) : null}
        {definition.extraActions?.filter(action => !action.visibleWhen || action.visibleWhen(row)).map(action => (
          <TableAction key={action.title} disabled={mutationBlocked || Boolean(action.disabledWhen?.(row))} onClick={() => openForm(action, row)}>{action.buttonLabel ?? action.title}</TableAction>
        ))}
        {definition.deleteAction ? (
          <TableAction danger disabled={mutationBlocked || Boolean(definition.deleteAction.disabledWhen?.(row))} onClick={() => deleteRow(row)}>{definition.deleteAction.buttonLabel ?? '删除'}</TableAction>
        ) : null}
      </TableActions>
    )
  })

  return columns
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

function renderFormControl(field: MutationFormField, disabled = false) {
  if (field.renderControl) return field.renderControl(disabled || field.readOnly)
  if (field.type === 'number') {
    return <InputNumber min={field.min} max={field.max} className="full-width-control" readOnly={field.readOnly} />
  }
  if (field.type === 'boolean') {
    return <Switch disabled={field.readOnly} />
  }
  if (field.type === 'select') {
    return <Select options={field.options ?? []} disabled={field.readOnly} />
  }
  if (field.type === 'textarea') {
    return <Input.TextArea rows={5} spellCheck={false} placeholder={field.placeholder} readOnly={field.readOnly} />
  }
  return <Input placeholder={field.placeholder} readOnly={field.readOnly} />
}

function renderValue(value: unknown) {
  if (Array.isArray(value)) {
    return value.length ? value.map((item) => textValue(item)).join(', ') : '-'
  }
  return textValue(value, '-')
}

function rowKey(definition: ExternalListPageDefinition, row: ApiRecord, index?: number) {
  const candidates = definition.rowKeyCandidates ?? ['id', 'name', 'fingerprint', 'dashboard_uri']
  return candidates.map((key) => row[key]).find(Boolean)?.toString() ?? String(index)
}

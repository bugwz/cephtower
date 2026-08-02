import { PlusOutlined, ReloadOutlined, SaveOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, InputNumber, Modal, Space } from 'antd'
import { useCallback, useState } from 'react'
import { listResource, mutateResource, refreshResource } from '../../api/resource'
import type { ApiRecord } from '../../api/client'
import type { ResourceListResult } from '../../api/resource'
import { AppTable } from '../../components/AppTable'
import { ResourceMetaBar } from '../../components/ResourceMetaBar'
import { DraggableModal } from '../../components/DraggableModal'
import { Page } from '../../components/Page'
import { RecordDetail } from '../../components/RecordDetail'
import { TableAction, TableActions } from '../../components/TableActions'
import { useResource } from '../../hooks'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useResourceTableFilters } from '../../hooks/useResourceTableFilters'
import { useClusterContext } from '../../state/ClusterContext'
import { message } from '../../utils/appMessage'

interface PoolFormValues {
  name: string
  pg_num?: number
}

export function PoolPage() {
  const { selectedClusterId } = useClusterContext()
  const [modalOpen, setModalOpen] = useState(false)
  const [detailRow, setDetailRow] = useState<ApiRecord | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [refreshingPools, setRefreshingPools] = useState(false)
  const [form] = Form.useForm<PoolFormValues>()
  const operationMutation = useMutationOperation()
  const poolTableFilters = useResourceTableFilters({
    path: '/pools',
    fields: ['name', 'status', 'type', 'size', 'min_size', 'pg_num', 'application_metadata', 'resource_version'],
    clusterId: selectedClusterId
  })
  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return emptyResult()
    }
    return listResource('/pools', selectedClusterId, { filters: poolTableFilters.filters })
  }, [poolTableFilters.filters, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)

  async function refreshPools() {
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
    form.resetFields()
    form.setFieldsValue({ pg_num: 32 })
    setModalOpen(true)
  }

  async function submitPool(values: PoolFormValues) {
    if (!selectedClusterId || submitting) {
      return
    }
    setSubmitting(true)
    try {
      await operationMutation.run(() => mutateResource('/pool', 'POST', {
        cluster_id: selectedClusterId,
        name: values.name,
        ...(values.pg_num ? { pg_num: values.pg_num } : {})
      }), false)
      setModalOpen(false)
      message.success('存储池创建执行成功')
      void refresh({ showLoading: false })
    } finally {
      setSubmitting(false)
    }
  }

  async function deletePool(row: ApiRecord) {
    if (!selectedClusterId) {
      return
    }
    const poolName = poolKey(row)
    if (!poolName) {
      message.error('无法识别存储池名称')
      return
    }
    const generation = Number(row.resource_version ?? 0)
    const parameters = { cluster_id: selectedClusterId, name: poolName }
    Modal.confirm({
      title: `删除存储池 ${poolName}`,
      content: '该操作为高风险操作，确认后将直接执行删除操作。',
      okText: '提交删除',
      okType: 'danger',
      cancelText: '取消',
      async onOk() {
        await operationMutation.run(() => mutateResource('/pool', 'DELETE', parameters, { ifMatch: generation }), false)
        window.setTimeout(() => {
          message.success('存储池删除执行成功')
          void refresh({ showLoading: false })
        })
      }
    })
  }

  return (
    <Page title="存储池" loading={loading} error={error}>
      <Card
        className="page-surface-card"
        title="存储池"
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} loading={refreshingPools} onClick={refreshPools}>刷新</Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建存储池</Button>
          </Space>
        }
      >
        <AppTable<ApiRecord>
          size="middle"
          rowKey={(row) => String(row.natural_key ?? row.name)}
          dataSource={data?.items ?? []}
          pagination={{ defaultPageSize: 10, showSizeChanger: true }}
          scroll={{ x: 1180 }}
          onChange={(_pagination, filters) => poolTableFilters.handleFilterChange(tableFilters(filters))}
          footer={() => <ResourceMetaBar observedAt={data?.observedAt} stale={data?.stale} staleReason={data?.staleReason} />}
          columns={[
            filterColumn('名称', 'name', 180, poolTableFilters),
            filterColumn('状态', 'status', 120, poolTableFilters),
            filterColumn('类型', 'type', 120, poolTableFilters),
            filterColumn('副本/大小', 'size', 120, poolTableFilters),
            filterColumn('最小副本', 'min_size', 120, poolTableFilters),
            filterColumn('PG', 'pg_num', 100, poolTableFilters),
            { ...filterColumn('应用', 'application_metadata', 220, poolTableFilters), render: renderValue },
            filterColumn('版本', 'resource_version', 90, poolTableFilters),
            {
              title: '操作',
              width: 100,
              fixed: 'right',
              render: (_, row) => (
                <TableActions>
                  <TableAction onClick={() => setDetailRow(row)}>详情</TableAction>
                  <TableAction danger onClick={() => deletePool(row)}>删除</TableAction>
                </TableActions>
              )
            }
          ]}
        />
      </Card>

      <DraggableModal
        title="新建存储池"
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={() => form.submit()}
        okText="提交"
        confirmLoading={submitting}
        okButtonProps={{ icon: <SaveOutlined /> }}
        destroyOnClose
      >
        <Form form={form} layout="vertical" onFinish={submitPool}>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入存储池名称' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="pg_num" label="PG 数">
            <InputNumber min={1} />
          </Form.Item>
        </Form>
      </DraggableModal>
      <DraggableModal
        title="存储池详情"
        open={Boolean(detailRow)}
        onCancel={() => setDetailRow(null)}
        footer={null}
        width={720}
        destroyOnClose
      >
        <RecordDetail record={detailRow} />
      </DraggableModal>
    </Page>
  )
}

function poolKey(row: ApiRecord) {
  return String(row.name ?? row.pool_name ?? row.natural_key ?? '').trim()
}

function renderValue(value: unknown) {
  if (value === null || value === undefined || value === '') {
    return '-'
  }
  if (typeof value === 'object') {
    return JSON.stringify(value)
  }
  return String(value)
}

function filterColumn(title: string, field: string, width: number, tableFilters: ReturnType<typeof useResourceTableFilters>) {
  return {
    title,
    dataIndex: field,
    key: field,
    width,
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

function emptyResult(): ResourceListResult<ApiRecord> {
  return {
    items: [],
    nextCursor: null,
    observedAt: null,
    stale: false,
    staleReason: null
  }
}

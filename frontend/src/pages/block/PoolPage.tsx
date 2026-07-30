import { DeleteOutlined, EyeOutlined, PlusOutlined, ReloadOutlined, SaveOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, InputNumber, Modal, Space, Table } from 'antd'
import { useCallback, useState } from 'react'
import { listResource, mutateResource, refreshResource } from '../../api/resource'
import type { ApiRecord } from '../../api/client'
import type { ResourceListResult } from '../../api/resource'
import { ResourceMetaBar } from '../../components/ResourceMetaBar'
import { DraggableModal } from '../../components/DraggableModal'
import { Page } from '../../components/Page'
import { RecordDetail } from '../../components/RecordDetail'
import { useResource } from '../../hooks'
import { useMutationOperation } from '../../hooks/useMutationOperation'
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
  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return emptyResult()
    }
    return listResource('/pools', selectedClusterId)
  }, [selectedClusterId])
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
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kind: 'pool' }), '存储池刷新已触发')
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
      }), '存储池创建执行成功')
      setModalOpen(false)
      await refresh()
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
        await operationMutation.run(() => mutateResource('/pool', 'DELETE', parameters, { ifMatch: generation }), '存储池删除执行成功')
        await refresh()
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
        <Space direction="vertical" size={16} className="page-stack">
          <ResourceMetaBar observedAt={data?.observedAt} stale={data?.stale} staleReason={data?.staleReason} />
          <Table<ApiRecord>
            size="middle"
            rowKey={(row) => String(row.natural_key ?? row.name)}
            dataSource={data?.items ?? []}
            pagination={{ pageSize: 10, showSizeChanger: false }}
            scroll={{ x: 1180 }}
            columns={[
              { title: '名称', dataIndex: 'name', width: 180 },
              { title: '状态', dataIndex: 'status', width: 120 },
              { title: '类型', dataIndex: 'type', width: 120 },
              { title: '副本/大小', dataIndex: 'size', width: 120 },
              { title: '最小副本', dataIndex: 'min_size', width: 120 },
              { title: 'PG', dataIndex: 'pg_num', width: 100 },
              { title: '应用', dataIndex: 'application_metadata', width: 220, render: renderValue },
              { title: '版本', dataIndex: 'resource_version', width: 90 },
              {
                title: '操作',
                width: 190,
                fixed: 'right',
                render: (_, row) => (
                  <Space>
                    <Button size="small" icon={<EyeOutlined />} onClick={() => setDetailRow(row)}>
                      详情
                    </Button>
                    <Button danger size="small" icon={<DeleteOutlined />} onClick={() => deletePool(row)}>
                      删除
                    </Button>
                  </Space>
                )
              }
            ]}
            expandable={{
              expandedRowRender: (row) => <RecordDetail record={row} />,
              rowExpandable: (row) => Object.keys(row).length > 0
            }}
          />
        </Space>
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

function emptyResult(): ResourceListResult<ApiRecord> {
  return {
    items: [],
    nextCursor: null,
    observedAt: null,
    stale: false,
    staleReason: null
  }
}

import { PlusOutlined, ReloadOutlined, SaveOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, InputNumber, Popconfirm, Select, Space, Switch, Table, Tabs, Tag } from 'antd'
import { useCallback, useState } from 'react'
import {
  createEndpoint,
  deleteCredential,
  deleteEndpoint,
  listCredentials,
  listEndpoints,
  putCredential,
  updateEndpoint,
  type CredentialView,
  type EndpointInput,
  type EndpointView
} from '../../api/endpoint'
import type { ApiRecord } from '../../api/client'
import { DraggableModal } from '../../components/DraggableModal'
import { Page } from '../../components/Page'
import { TableAction, TableActions } from '../../components/TableActions'
import { message } from '../../utils/appMessage'
import { useResource } from '../../hooks'
import { useClusterContext } from '../../state/ClusterContext'

interface EndpointFormValues extends EndpointInput {
  endpoint_id?: number
}

interface CredentialFormValues {
  kind: string
  credential_json: string
}

const endpointKinds = ['prometheus', 'alertmanager', 'grafana', 'iscsi', 'nvmeof', 's3', 'rgw', 'rgw_admin', 'ca']
const tlsModes = ['verify_system', 'verify_custom_ca']

export function DataPage() {
  const { selectedClusterId } = useClusterContext()
  const [endpointModalOpen, setEndpointModalOpen] = useState(false)
  const [credentialModalOpen, setCredentialModalOpen] = useState(false)
  const [editingEndpoint, setEditingEndpoint] = useState<EndpointView | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [endpointForm] = Form.useForm<EndpointFormValues>()
  const [credentialForm] = Form.useForm<CredentialFormValues>()

  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return { endpoints: [] as EndpointView[], credentials: [] as CredentialView[] }
    }
    const [endpoints, credentials] = await Promise.all([
      listEndpoints(selectedClusterId),
      listCredentials(selectedClusterId)
    ])
    return { endpoints, credentials }
  }, [selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)

  function openCreateEndpoint() {
    setEditingEndpoint(null)
    endpointForm.resetFields()
    endpointForm.setFieldsValue({
      kind: 'prometheus',
      name: 'default',
      tls_mode: 'verify_system',
      timeout_seconds: 10,
      enabled: true
    })
    setEndpointModalOpen(true)
  }

  function openEditEndpoint(row: EndpointView) {
    setEditingEndpoint(row)
    endpointForm.setFieldsValue({
      endpoint_id: row.endpoint_id,
      kind: row.kind,
      name: row.name,
      url: row.url,
      tls_mode: row.tls_mode,
      ca_credential_id: row.ca_credential_id ?? undefined,
      timeout_seconds: readTimeout(row),
      enabled: row.enabled
    })
    setEndpointModalOpen(true)
  }

  function openCreateCredential() {
    credentialForm.resetFields()
    credentialForm.setFieldsValue({ kind: 'prometheus', credential_json: '{\n  "token": ""\n}' })
    setCredentialModalOpen(true)
  }

  async function submitEndpoint(values: EndpointFormValues) {
    if (!selectedClusterId || submitting) {
      return
    }
    setSubmitting(true)
    try {
      if (editingEndpoint) {
        await updateEndpoint(selectedClusterId, editingEndpoint.endpoint_id, values)
        message.success('Endpoint 已更新')
      } else {
        await createEndpoint(selectedClusterId, values)
        message.success('Endpoint 已创建')
      }
      setEndpointModalOpen(false)
      await refresh()
    } finally {
      setSubmitting(false)
    }
  }

  async function submitCredential(values: CredentialFormValues) {
    if (!selectedClusterId || submitting) {
      return
    }
    setSubmitting(true)
    try {
      const credential = JSON.parse(values.credential_json) as ApiRecord
      await putCredential(selectedClusterId, values.kind, credential)
      message.success('Credential 已保存')
      setCredentialModalOpen(false)
      await refresh()
    } finally {
      setSubmitting(false)
    }
  }

  async function removeEndpoint(row: EndpointView) {
    if (!selectedClusterId) {
      return
    }
    await deleteEndpoint(selectedClusterId, row.endpoint_id)
    message.success('Endpoint 已删除')
    await refresh()
  }

  async function removeCredential(row: CredentialView) {
    if (!selectedClusterId) {
      return
    }
    await deleteCredential(selectedClusterId, row.kind)
    message.success('Credential 已删除')
    await refresh()
  }

  return (
    <Page title="配置管理" loading={loading} error={error}>
      <Card
        className="page-surface-card"
        title="Endpoint 与 Credential"
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} loading={loading} onClick={refresh}>刷新</Button>
            <Button icon={<PlusOutlined />} disabled={!selectedClusterId} onClick={openCreateCredential}>新增 Credential</Button>
            <Button type="primary" icon={<PlusOutlined />} disabled={!selectedClusterId} onClick={openCreateEndpoint}>新增 Endpoint</Button>
          </Space>
        }
      >
        <Tabs
          items={[
            {
              key: 'endpoints',
              label: 'Endpoints',
              children: (
                <Table<EndpointView>
                  size="middle"
                  rowKey="endpoint_id"
                  dataSource={data?.endpoints ?? []}
                  pagination={{ pageSize: 8, showSizeChanger: false }}
                  scroll={{ x: 980 }}
                  columns={[
                    { title: 'ID', dataIndex: 'endpoint_id', width: 80 },
                    { title: 'Kind', dataIndex: 'kind', width: 140 },
                    { title: 'Name', dataIndex: 'name', width: 140 },
                    { title: 'URL', dataIndex: 'url', ellipsis: true },
                    { title: 'TLS', dataIndex: 'tls_mode', width: 150 },
                    { title: '状态', dataIndex: 'enabled', width: 100, render: (enabled) => <Tag color={enabled ? 'success' : 'default'}>{enabled ? '启用' : '停用'}</Tag> },
                    {
                      title: '操作',
                      width: 100,
                      render: (_, row) => (
                        <TableActions>
                          <TableAction onClick={() => openEditEndpoint(row)}>编辑</TableAction>
                          <Popconfirm title="删除 Endpoint" okText="删除" cancelText="取消" onConfirm={() => removeEndpoint(row)}>
                            <TableAction danger>删除</TableAction>
                          </Popconfirm>
                        </TableActions>
                      )
                    }
                  ]}
                />
              )
            },
            {
              key: 'credentials',
              label: 'Credentials',
              children: (
                <Table<CredentialView>
                  size="middle"
                  rowKey="kind"
                  dataSource={data?.credentials ?? []}
                  pagination={{ pageSize: 8, showSizeChanger: false }}
                  columns={[
                    { title: 'Kind', dataIndex: 'kind' },
                    { title: 'Fingerprint', dataIndex: 'fingerprint', ellipsis: true },
                    { title: '创建时间', dataIndex: 'created_at', render: formatTime },
                    { title: '更新时间', dataIndex: 'updated_at', render: formatTime },
                    {
                      title: '操作',
                      width: 70,
                      render: (_, row) => (
                        <Popconfirm title="删除 Credential" okText="删除" cancelText="取消" onConfirm={() => removeCredential(row)}>
                          <TableAction danger>删除</TableAction>
                        </Popconfirm>
                      )
                    }
                  ]}
                />
              )
            }
          ]}
        />
      </Card>

      <DraggableModal
        title={editingEndpoint ? '编辑 Endpoint' : '新增 Endpoint'}
        open={endpointModalOpen}
        onCancel={() => setEndpointModalOpen(false)}
        onOk={() => endpointForm.submit()}
        okText="保存"
        confirmLoading={submitting}
        okButtonProps={{ icon: <SaveOutlined /> }}
        destroyOnClose
      >
        <Form form={endpointForm} layout="vertical" onFinish={submitEndpoint}>
          <Form.Item name="kind" label="Kind" rules={[{ required: true }]}>
            <Select options={endpointKinds.map((kind) => ({ label: kind, value: kind }))} />
          </Form.Item>
          <Form.Item name="name" label="Name" rules={[{ required: true }]}>
            <Input placeholder="default" />
          </Form.Item>
          <Form.Item name="url" label="URL" rules={[{ required: true, type: 'url', message: '请输入 http(s) URL' }]}>
            <Input placeholder="https://example:9090" />
          </Form.Item>
          <Form.Item name="tls_mode" label="TLS 模式" rules={[{ required: true }]}>
            <Select options={tlsModes.map((mode) => ({ label: mode, value: mode }))} />
          </Form.Item>
          <Form.Item name="ca_credential_id" label="CA Credential ID">
            <InputNumber min={1} controls={false} className="full-width-control" />
          </Form.Item>
          <Form.Item name="timeout_seconds" label="超时秒数" rules={[{ required: true }]}>
            <InputNumber min={1} max={120} />
          </Form.Item>
          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </DraggableModal>

      <DraggableModal
        title="新增 Credential"
        open={credentialModalOpen}
        onCancel={() => setCredentialModalOpen(false)}
        onOk={() => credentialForm.submit()}
        okText="保存"
        confirmLoading={submitting}
        okButtonProps={{ icon: <SaveOutlined /> }}
        destroyOnClose
      >
        <Form form={credentialForm} layout="vertical" onFinish={submitCredential}>
          <Form.Item name="kind" label="Kind" rules={[{ required: true }]}>
            <Select options={endpointKinds.map((kind) => ({ label: kind, value: kind }))} />
          </Form.Item>
          <Form.Item name="credential_json" label="Credential JSON" rules={[{ required: true }]}>
            <Input.TextArea rows={8} spellCheck={false} />
          </Form.Item>
        </Form>
      </DraggableModal>
    </Page>
  )
}

function readTimeout(_row: EndpointView) {
  return 10
}

function formatTime(value?: string | null) {
  if (!value) {
    return '-'
  }
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

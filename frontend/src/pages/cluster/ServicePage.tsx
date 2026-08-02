import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Modal, Select, Space, Tabs } from 'antd'
import { useCallback, useState } from 'react'
import { textValue, type ApiRecord } from '../../api/client'
import { listResource, mutateResource, refreshResource } from '../../api/resource'
import { DataTable } from '../../components/DataTable'
import { DraggableModal } from '../../components/DraggableModal'
import { Page } from '../../components/Page'
import { TableAction, TableActions } from '../../components/TableActions'
import { useResource } from '../../hooks'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useResourceTableFilters } from '../../hooks/useResourceTableFilters'
import { useClusterContext } from '../../state/ClusterContext'
import { message } from '../../utils/appMessage'

interface ServiceFormValues {
  service_type: string
  service_id?: string
  placement_json?: string
}

const serviceTypeOptions = [
  'mon',
  'mgr',
  'mds',
  'rgw',
  'nfs',
  'smb',
  'prometheus',
  'alertmanager',
  'grafana',
  'node-exporter',
  'crash'
].map((value) => ({ label: value, value }))

export function ServicePage() {
  const { selectedClusterId } = useClusterContext()
  const serviceTableFilters = useResourceTableFilters({
    path: '/services',
    fields: ['service_name', 'service_type', 'status', 'running', 'size'],
    clusterId: selectedClusterId
  })
  const daemonTableFilters = useResourceTableFilters({
    path: '/daemons',
    fields: ['daemon_name', 'daemon_type', 'hostname', 'status_desc', 'version', 'container_image_name'],
    clusterId: selectedClusterId
  })
  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return { services: [], daemons: [] }
    }
    const [services, daemons] = await Promise.all([
      listResource('/services', selectedClusterId, { filters: serviceTableFilters.filters }).then((payload) => payload.items),
      listResource('/daemons', selectedClusterId, { filters: daemonTableFilters.filters }).then((payload) => payload.items)
    ])
    return { services, daemons }
  }, [daemonTableFilters.filters, selectedClusterId, serviceTableFilters.filters])
  const { data, loading, error, refresh } = useResource(loader)
  const [form] = Form.useForm<ServiceFormValues>()
  const [formOpen, setFormOpen] = useState(false)
  const [editingService, setEditingService] = useState<ApiRecord | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [refreshingServices, setRefreshingServices] = useState(false)
  const operationMutation = useMutationOperation()

  async function refreshServiceData() {
    if (refreshingServices) {
      return
    }
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    setRefreshingServices(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kinds: ['service', 'daemon'] }), '刷新成功')
      await refresh()
    } finally {
      setRefreshingServices(false)
    }
  }

  function openCreate() {
    setEditingService(null)
    form.resetFields()
    form.setFieldsValue({ service_type: 'rgw', placement_json: '{}' })
    setFormOpen(true)
  }

  function openEdit(row: ApiRecord) {
    setEditingService(row)
    form.resetFields()
    form.setFieldsValue({
      service_type: serviceType(row),
      service_id: serviceId(row),
      placement_json: JSON.stringify(readObject(row.placement), null, 2)
    })
    setFormOpen(true)
  }

  async function submitService(values: ServiceFormValues) {
    if (!selectedClusterId || submitting) {
      return
    }
    setSubmitting(true)
    try {
      let placement: ApiRecord
      try {
        placement = parsePlacement(values.placement_json)
      } catch (err) {
        message.error(err instanceof Error ? err.message : 'Placement JSON 格式错误')
        return
      }
      const body = {
        cluster_id: selectedClusterId,
        ...(editingService ? { name: serviceName(editingService) } : {}),
        service_type: values.service_type,
        ...(values.service_id ? { service_id: values.service_id } : {}),
        placement
      }
      const successMessage = editingService ? '服务更新执行成功' : '服务创建执行成功'
      await operationMutation.run(() => mutateResource('/service', editingService ? 'PATCH' : 'POST', body, editingService ? { ifMatch: Number(editingService.resource_version ?? 0) } : undefined), false)
      setFormOpen(false)
      message.success(successMessage)
      void refresh({ showLoading: false })
    } finally {
      setSubmitting(false)
    }
  }

  async function deleteService(row: ApiRecord) {
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    const name = serviceName(row)
    if (!name) {
      message.error('无法识别服务名')
      return
    }
    const generation = Number(row.resource_version ?? 0)
    const parameters = { cluster_id: selectedClusterId, name }
    Modal.confirm({
      title: `删除服务 ${name}`,
      content: '该操作为高风险操作，确认后将直接执行删除操作。',
      okText: '提交删除',
      okType: 'danger',
      cancelText: '取消',
      async onOk() {
        await operationMutation.run(() => mutateResource('/service', 'DELETE', parameters, { ifMatch: generation }), false)
        window.setTimeout(() => {
          message.success('服务删除执行成功')
          void refresh({ showLoading: false })
        })
      }
    })
  }

  return (
    <Page
      title="服务与守护进程"
      loading={loading}
      error={error}
    >
      <Card
        className="page-surface-card"
        title="服务与守护进程"
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} loading={refreshingServices} onClick={refreshServiceData}>刷新</Button>
            <Button type="primary" icon={<PlusOutlined />} disabled={!selectedClusterId} onClick={openCreate}>新增服务</Button>
          </Space>
        }
      >
        <Tabs
          items={[
            {
              key: 'services',
              label: '服务',
              children: (
                <div className="embedded-panel">
                <DataTable
                  data={data?.services ?? []}
                  filterOptions={serviceTableFilters.filterOptions}
                  filteredValues={serviceTableFilters.filters}
                  onFilterChange={serviceTableFilters.handleFilterChange}
                  rowKeyCandidates={['service_name', 'service_id', 'name']}
                  columns={[
                    { key: 'service_name', title: '服务名' },
                    { key: 'service_type', title: '类型' },
                    { key: 'placement', title: '放置策略' },
                    { key: 'status', title: '状态' },
                    { key: 'running', title: '运行数' },
                    { key: 'size', title: '目标数' },
                    {
                      key: 'actions',
                      title: '操作',
                      filterKey: false,
                      render: (_, row) => (
                        <TableActions>
                          <TableAction onClick={() => openEdit(row)}>编辑</TableAction>
                          <TableAction danger onClick={() => deleteService(row)}>删除</TableAction>
                        </TableActions>
                      )
                    }
                  ]}
                />
                </div>
              )
            },
            {
              key: 'daemons',
              label: '守护进程',
              children: (
                <div className="embedded-panel">
                <DataTable
                  data={data?.daemons ?? []}
                  filterOptions={daemonTableFilters.filterOptions}
                  filteredValues={daemonTableFilters.filters}
                  onFilterChange={daemonTableFilters.handleFilterChange}
                  rowKeyCandidates={['daemon_name', 'name', 'hostname']}
                  columns={[
                    { key: 'daemon_name', title: 'Daemon' },
                    { key: 'daemon_type', title: '类型' },
                    { key: 'hostname', title: '主机' },
                    { key: 'status_desc', title: '状态' },
                    { key: 'version', title: '版本' },
                    { key: 'container_image_name', title: '镜像' }
                  ]}
                />
                </div>
              )
            }
          ]}
        />
      </Card>
      <DraggableModal
        title={editingService ? '编辑服务' : '新增服务'}
        open={formOpen}
        onCancel={() => setFormOpen(false)}
        onOk={() => form.submit()}
        okText="提交"
        confirmLoading={submitting}
        destroyOnClose
      >
        <Form form={form} layout="vertical" onFinish={submitService}>
          <Form.Item name="service_type" label="服务类型" rules={[{ required: true, message: '请选择服务类型' }]}>
            <Select options={serviceTypeOptions} />
          </Form.Item>
          <Form.Item name="service_id" label="Service ID">
            <Input />
          </Form.Item>
          <Form.Item name="placement_json" label="Placement JSON">
            <Input.TextArea rows={5} spellCheck={false} placeholder='{"count":1,"host_pattern":"*"}' />
          </Form.Item>
        </Form>
      </DraggableModal>
    </Page>
  )
}

function serviceName(row: ApiRecord) {
  return textValue(row.service_name ?? row.name ?? row.service_id, '')
}

function serviceType(row: ApiRecord) {
  return textValue(row.service_type ?? row.type, 'rgw')
}

function serviceId(row: ApiRecord) {
  return textValue(row.service_id ?? row.id, '')
}

function parsePlacement(value?: string): ApiRecord {
  const parsed = JSON.parse(value || '{}')
  if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) {
    throw new Error('Placement 必须是 JSON 对象')
  }
  return parsed as ApiRecord
}

function readObject(value: unknown): ApiRecord {
  return typeof value === 'object' && value !== null && !Array.isArray(value) ? value as ApiRecord : {}
}

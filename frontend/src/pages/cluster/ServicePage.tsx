import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Form, Input, Modal, Select, Space, Tabs } from 'antd'
import { useCallback, useEffect, useRef, useState } from 'react'
import { textValue, type ApiRecord } from '../../api/client'
import { listAllResources, mutateResource, refreshResource } from '../../api/resource'
import { DataTable } from '../../components/DataTable'
import { DraggableModal } from '../../components/DraggableModal'
import { Page } from '../../components/Page'
import { TableAction, TableActions } from '../../components/TableActions'
import { useResource } from '../../hooks'
import { useResourceTableFilters } from '../../hooks/useResourceTableFilters'
import { useClusterContext } from '../../state/ClusterContext'
import { message } from '../../utils/appMessage'
import { ServiceDaemons } from './ServiceDaemons'
import { ResourceMetaBar } from '../../components/ResourceMetaBar'

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
  return <ServicePageContent key={selectedClusterId ?? 'none'} />
}

function ServicePageContent() {
  const { selectedClusterId } = useClusterContext()
  const active = useRef(true)
  const running = useRef(false)
  useEffect(() => { active.current = true; return () => { active.current = false } }, [])
  const serviceTableFilters = useResourceTableFilters({
    path: '/services',
    fields: ['name', 'type', 'running', 'size'],
    clusterId: selectedClusterId
  })
  const daemonTableFilters = useResourceTableFilters({
    path: '/daemons',
    fields: ['name', 'type', 'hostname', 'status', 'version', 'container_image'],
    clusterId: selectedClusterId
  })
  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return { services: [], daemons: [], serviceMeta: null, daemonMeta: null }
    }
    const [services, daemons] = await Promise.all([
      listAllResources('/services', selectedClusterId, { filters: serviceTableFilters.filters }),
      listAllResources('/daemons', selectedClusterId, { filters: daemonTableFilters.filters })
    ])
    return { services: services.items, daemons: daemons.items, serviceMeta: services, daemonMeta: daemons }
  }, [daemonTableFilters.filters, selectedClusterId, serviceTableFilters.filters])
  const { data, loading, error, refresh } = useResource(loader)
  const [form] = Form.useForm<ServiceFormValues>()
  const [formOpen, setFormOpen] = useState(false)
  const [detail, setDetail] = useState<{ clusterId: number; name: string } | null>(null)
  const visibleDetail = detail?.clusterId === selectedClusterId ? detail : null
  const [editingService, setEditingService] = useState<ApiRecord | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [refreshingServices, setRefreshingServices] = useState(false)

  async function refreshAfterMutation() {
    if (!active.current || !selectedClusterId) return
    try { await refreshResource({ clusterId: selectedClusterId, kinds: ['service', 'daemon'] }) }
    catch { if (active.current) message.warning('修改已执行，但重新采集失败，请刷新核对，不要重复提交。') }
    if (active.current) await refresh()
  }

  async function refreshServiceData() {
    if (!active.current || running.current) {
      return
    }
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    running.current = true; setRefreshingServices(true)
    try {
      await refreshResource({ clusterId: selectedClusterId, kinds: ['service', 'daemon'] })
      if (active.current) { message.success('刷新成功'); await refresh() }
    } finally {
      running.current = false; if (active.current) setRefreshingServices(false)
    }
  }

  function openCreate() {
    if (!active.current || running.current || !selectedClusterId) return
    setEditingService(null)
    form.resetFields()
    form.setFieldsValue({ service_type: 'rgw', placement_json: '{}' })
    setFormOpen(true)
  }

  function openEdit(row: ApiRecord) {
    if (!active.current || running.current || loading || error || !serviceWritable(row)) return
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
    if (!active.current || running.current || !selectedClusterId || loading || error || (editingService && !serviceWritable(editingService))) {
      return
    }
    running.current = true; setSubmitting(true)
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
      await mutateResource('/service', editingService ? 'PATCH' : 'POST', body, editingService ? { ifMatch: String(editingService.resource_version) } : undefined)
      if (!active.current) return
      setFormOpen(false)
      message.success(successMessage)
      await refreshAfterMutation()
    } finally {
      running.current = false; if (active.current) setSubmitting(false)
    }
  }

  async function deleteService(row: ApiRecord) {
    if (!active.current || running.current || loading || error || !serviceWritable(row)) return
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    const name = serviceName(row)
    if (!name) {
      message.error('无法识别服务名')
      return
    }
    const generation = String(row.resource_version)
    const parameters = { cluster_id: selectedClusterId, name }
    Modal.confirm({
      title: `删除服务 ${name}`,
      content: '该操作为高风险操作，确认后将直接执行删除操作。',
      okText: '提交删除',
      okType: 'danger',
      cancelText: '取消',
      async onOk() {
        if (!active.current) throw new Error('集群已切换，请重新确认删除')
        if (running.current) throw new Error('已有操作正在执行')
        running.current = true; setSubmitting(true)
        try {
          await mutateResource('/service', 'DELETE', parameters, { ifMatch: generation })
          if (!active.current) return
          message.success('服务删除执行成功')
          await refreshAfterMutation()
        } finally { running.current = false; if (active.current) setSubmitting(false) }
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
                <ResourceMetaBar observedAt={data?.serviceMeta?.observedAt} stale={data?.serviceMeta?.stale} staleReason={data?.serviceMeta?.staleReason} />
                {data?.serviceMeta?.stale && <Alert type="warning" message="服务库存已过期，请刷新后核对运行数与配置。" />}
                <DataTable
                  data={data?.services ?? []}
                  filterOptions={serviceTableFilters.filterOptions}
                  filteredValues={serviceTableFilters.filters}
                  onFilterChange={serviceTableFilters.handleFilterChange}
                  rowKeyCandidates={['service_name', 'service_id', 'name']}
                  columns={[
                    { key: 'name', title: '服务名' },
                    { key: 'type', title: '类型' },
                    { key: 'placement', title: '放置策略' },
                    { key: 'unmanaged', title: '管理模式', filterKey: false, render: (value) => value === true ? '非托管' : value === false ? '编排器管理' : '未采集' },
                    { key: 'running', title: '运行数' },
                    { key: 'size', title: '目标数' },
                    { key: 'last_refresh', title: 'Ceph 最近刷新', filterKey: false },
                    { key: 'ports', title: '端口', filterKey: false },
                    { key: 'service_url', title: '服务访问地址', filterKey: false },
                    { key: 'virtual_ip', title: '虚拟 IP', filterKey: false },
                    { key: 'container_image_name', title: '容器镜像', filterKey: false },
                    { key: 'container_image_id', title: '镜像 ID', filterKey: false },
                    { key: 'ceph_created_at', title: 'Ceph 服务创建时间', filterKey: false },
                    { key: 'events', title: '服务事件', filterKey: false, render: (value) => Array.isArray(value) ? <div style={{ whiteSpace: 'pre-wrap' }}>{value.map(String).join('\n')}</div> : '本次未返回事件' },
                    {
                      key: 'actions',
                      title: '操作',
                      filterKey: false,
                      render: (_, row) => (
                        <TableActions>
                          <TableAction disabled={!selectedClusterId || !serviceName(row)} onClick={() => { if (selectedClusterId) setDetail({ clusterId: selectedClusterId, name: serviceName(row) }) }}>守护进程</TableAction>
                          <TableAction disabled={loading || Boolean(error) || submitting || refreshingServices || !serviceWritable(row)} onClick={() => openEdit(row)}>编辑</TableAction>
                          <TableAction danger disabled={loading || Boolean(error) || submitting || refreshingServices || !serviceWritable(row)} onClick={() => deleteService(row)}>删除</TableAction>
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
                <ResourceMetaBar observedAt={data?.daemonMeta?.observedAt} stale={data?.daemonMeta?.stale} staleReason={data?.daemonMeta?.staleReason} />
                {data?.daemonMeta?.stale && <Alert type="warning" message="守护进程库存已过期，请刷新后核对状态。" />}
                <DataTable
                  data={data?.daemons ?? []}
                  filterOptions={daemonTableFilters.filterOptions}
                  filteredValues={daemonTableFilters.filters}
                  onFilterChange={daemonTableFilters.handleFilterChange}
                  rowKeyCandidates={['daemon_name', 'name', 'hostname']}
                  columns={[
                    { key: 'name', title: 'Daemon' },
                    { key: 'type', title: '类型' },
                    { key: 'hostname', title: '主机' },
                    { key: 'status', title: '状态' },
                    { key: 'version', title: '版本' },
                    { key: 'container_image', title: '镜像' }
                  ]}
                />
                </div>
              )
            }
          ]}
        />
      </Card>
      <Modal title={`服务 ${visibleDetail?.name ?? ''} 的守护进程`} open={Boolean(visibleDetail)} onCancel={() => setDetail(null)} footer={null} width="95vw" destroyOnClose>
        {visibleDetail && <ServiceDaemons key={`${visibleDetail.clusterId}:${visibleDetail.name}`} clusterId={visibleDetail.clusterId} name={visibleDetail.name} />}
      </Modal>
      <DraggableModal
        title={editingService ? '编辑服务' : '新增服务'}
        open={formOpen}
        onCancel={() => { if (!submitting) setFormOpen(false) }}
        onOk={() => form.submit()}
        okText="提交"
        confirmLoading={submitting}
        destroyOnClose
      >
        <Form form={form} layout="vertical" onFinish={submitService}>
          <Form.Item name="service_type" label="服务类型" rules={[{ required: true, message: '请选择服务类型' }]}>
            <Select disabled={Boolean(editingService)} options={serviceTypeOptions} />
          </Form.Item>
          <Form.Item name="service_id" label="Service ID">
            <Input disabled={Boolean(editingService)} />
          </Form.Item>
          <Form.Item name="placement_json" label="Placement JSON" extra='支持 count、count_per_host、hosts、label 和 host_pattern。count 与 count_per_host 互斥；每主机数量必须指定主机选择条件。host_pattern 支持通配符字符串或 {"pattern":"node-[0-9]+","pattern_type":"regex"}，正则语法由 Ceph 校验。'>
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

function serviceWritable(row: ApiRecord): boolean {
  const version = row.resource_version
  return row.stale === false && Boolean(serviceName(row)) && (typeof version === 'string' ? /^[1-9][0-9]*$/.test(version) : typeof version === 'number' && Number.isSafeInteger(version) && version > 0)
}

function serviceType(row: ApiRecord) {
  return textValue(row.service_type ?? row.type, 'rgw')
}

function serviceId(row: ApiRecord) {
  const name = serviceName(row)
  const type = serviceType(row)
  return name.startsWith(`${type}.`) ? name.slice(type.length + 1) : ''
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

import { ArrowLeftOutlined, BulbOutlined, DeleteOutlined, PlusOutlined, PoweroffOutlined, ReloadOutlined } from '@ant-design/icons'
import { Button, Card, Descriptions, Form, Input, InputNumber, Modal, Space, Switch, Tabs, Tag, Typography } from 'antd'
import { useCallback, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { numberValue, textValue, type ApiRecord } from '../../api/client'
import {
  applyDaemonAction,
  listMonitors,
  listResource,
  listOSDFlags,
  markOSD,
  mutateResource,
  refreshResource,
  reweightOSD,
  scrubOSD,
  setMgrModuleEnabled
} from '../../api/resource'
import { DataTable } from '../../components/DataTable'
import { DraggableModal, draggableModalRender } from '../../components/DraggableModal'
import { Page } from '../../components/Page'
import { ResourceMetaBar } from '../../components/ResourceMetaBar'
import { TableAction, TableActions } from '../../components/TableActions'
import { useResource } from '../../hooks'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { mergeResourceFilters, useResourceTableFilters } from '../../hooks/useResourceTableFilters'
import { useClusterContext } from '../../state/ClusterContext'
import { message } from '../../utils/appMessage'
import { formatDateTime } from '../../utils/time'
import { ClusterDetailPage } from './ClusterDetailPage'
import { ClusterPage } from './ClusterPage'
import { HostDetailPage } from './HostDetailPage'
import { formatBytes, HostPage } from './HostPage'
import { MonDetailPage } from './MonDetailPage'
import { PoolDetailPage } from './PoolDetailPage'
import { PoolManagementPage } from './PoolManagementPage'
import { ServicePage } from './ServicePage'

export { ClusterDetailPage, ClusterPage, HostDetailPage, HostPage, MonDetailPage, PoolDetailPage, PoolManagementPage, ServicePage }

const { Text } = Typography
const twoColumnDescriptions = { xs: 1, sm: 2, md: 2, lg: 2, xl: 2, xxl: 2 }

type DeviceScope = 'available' | 'used' | 'unavailable'

export function MonManagementPage() {
  const navigate = useNavigate()
  const { selectedClusterId } = useClusterContext()
  const monTableFilters = useResourceTableFilters({
    path: '/monitors',
    fields: ['name', 'rank', 'address', 'status'],
    clusterId: selectedClusterId
  })
  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return []
    }
    return listMonitors(selectedClusterId, monTableFilters.filters)
  }, [monTableFilters.filters, selectedClusterId])
  const { data, loading, error } = useResource(loader)

  return (
    <Page title="MON管理" loading={loading} error={error}>
      <Card className="page-surface-card" title="MON管理">
        <DataTable
          data={data ?? []}
          filterOptions={monTableFilters.filterOptions}
          filteredValues={monTableFilters.filters}
          onFilterChange={monTableFilters.handleFilterChange}
          rowKeyCandidates={['name', 'natural_key', 'rank']}
          columns={[
            { key: 'name', title: '名称' },
            { key: 'rank', title: 'Rank' },
            { key: 'address', title: 'Public Addr' },
            {
              key: 'status',
              title: '状态',
              render: (_, row) => <Tag color={row.in_quorum === true ? 'success' : 'default'}>{row.in_quorum === true ? '仲裁中' : '未加入仲裁'}</Tag>
            },
            {
              key: 'actions',
              title: '操作',
              filterKey: false,
              render: (_, row) => {
                const name = textValue(row.name ?? row.natural_key, '')
                return (
                  <TableActions>
                    <TableAction disabled={!name} onClick={() => navigate(`/cluster/mon/${encodeURIComponent(name)}`)}>详情</TableAction>
                  </TableActions>
                )
              }
            }
          ]}
        />
      </Card>
    </Page>
  )
}

export function MgrManagementPage() {
  const { selectedClusterId } = useClusterContext()
  const moduleTableFilters = useResourceTableFilters({
    path: '/manager/modules',
    fields: ['name', 'enabled', 'always_on'],
    clusterId: selectedClusterId
  })
  const daemonTableFilters = useResourceTableFilters({
    path: '/daemons',
    fields: ['daemon_name', 'daemon_type', 'hostname', 'status_desc', 'version'],
    clusterId: selectedClusterId
  })
  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return { modules: [], daemons: [] }
    }
    const [modules, daemons] = await Promise.all([
      listResource('/manager/modules', selectedClusterId, { filters: moduleTableFilters.filters }).then((payload) => payload.items),
      listResource('/daemons', selectedClusterId, {
        filters: mergeResourceFilters({ daemon_type: ['mgr'] }, daemonTableFilters.filters)
      }).then((payload) => payload.items)
    ])
    return { modules, daemons }
  }, [daemonTableFilters.filters, moduleTableFilters.filters, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const [pendingModule, setPendingModule] = useState('')
  const operationMutation = useMutationOperation()

  async function toggleModule(row: ApiRecord, enabled: boolean) {
    const name = textValue(row.name, '')
    if (!name || pendingModule) {
      return
    }
    setPendingModule(name)
    try {
      await operationMutation.run(() => setMgrModuleEnabled(name, enabled), enabled ? 'Mgr 模块启用执行成功' : 'Mgr 模块停用执行成功')
      refresh()
    } finally {
      setPendingModule('')
    }
  }

  return (
    <Page title="MGR管理" loading={loading} error={error}>
      <Card className="page-surface-card" title="MGR管理">
        <Tabs
          items={[
            {
              key: 'modules',
              label: '模块',
              children: (
                <div className="embedded-panel">
                <DataTable
                  data={data?.modules ?? []}
                  filterOptions={moduleTableFilters.filterOptions}
                  filteredValues={moduleTableFilters.filters}
                  onFilterChange={moduleTableFilters.handleFilterChange}
                  rowKeyCandidates={['name']}
                  columns={[
                    { key: 'name', title: '模块' },
                    {
                      key: 'enabled',
                      title: '启用',
                      render: (value, row) => {
                        const name = textValue(row.name, '')
                        return (
                          <Switch
                            checked={Boolean(value)}
                            disabled={Boolean(row.always_on) || (Boolean(pendingModule) && pendingModule !== name)}
                            loading={pendingModule === name}
                            onChange={(checked) => toggleModule(row, checked)}
                          />
                        )
                      }
                    },
                    { key: 'always_on', title: '常驻', render: (value) => <Tag color={value ? 'processing' : 'default'}>{value ? '是' : '否'}</Tag> },
                    { key: 'options', title: '配置项', render: (value) => textValue(value) }
                  ]}
                />
                </div>
              )
            },
            {
              key: 'daemons',
              label: '守护进程',
              children: <DaemonTable data={data?.daemons ?? []} refresh={refresh} tableFilters={daemonTableFilters} />
            }
          ]}
        />
      </Card>
    </Page>
  )
}

export function OsdManagementPage() {
  const { selectedClusterId } = useClusterContext()
  const osdTableFilters = useResourceTableFilters({
    path: '/osds',
    fields: ['id', 'host', 'state', 'up', 'in', 'device_class'],
    clusterId: selectedClusterId
  })
  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return { osds: [], flags: [] }
    }
    const [osds, flags] = await Promise.all([
      listResource('/osds', selectedClusterId, { filters: osdTableFilters.filters }).then((payload) => payload.items),
      listOSDFlags()
    ])
    return { osds, flags }
  }, [osdTableFilters.filters, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const [pendingOSDAction, setPendingOSDAction] = useState('')
  const [deploymentOpen, setDeploymentOpen] = useState(false)
  const [refreshingOSDs, setRefreshingOSDs] = useState(false)
  const operationMutation = useMutationOperation()

  async function refreshOSDData() {
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    setRefreshingOSDs(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kinds: ['osd', 'osd_flag'] }), '刷新成功')
      await refresh()
    } finally {
      setRefreshingOSDs(false)
    }
  }

  async function runOSDAction(id: string, action: 'in' | 'out' | 'scrub' | 'deep-scrub' | 'reweight') {
    if (action === 'reweight') {
      Modal.confirm({
        title: `调整 OSD ${id} 权重`,
        content: <ReweightForm osdID={id} refresh={refresh} />,
        modalRender: draggableModalRender,
        icon: null,
        okButtonProps: { style: { display: 'none' } },
        cancelText: '关闭'
      })
      return
    }

    const pendingKey = `${id}:${action}`
    if (pendingOSDAction) {
      return
    }
    setPendingOSDAction(pendingKey)
    try {
      if (action === 'scrub' || action === 'deep-scrub') {
        await operationMutation.run(() => scrubOSD(id, action === 'deep-scrub'), 'Scrub 执行成功')
      } else {
        await operationMutation.run(() => markOSD(id, action), `OSD ${action} 执行成功`)
      }
      refresh()
    } finally {
      setPendingOSDAction('')
    }
  }

  async function deleteOSD(row: ApiRecord) {
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    const id = osdID(row)
    if (!id) {
      message.error('无法识别 OSD ID')
      return
    }
    const generation = Number(row.resource_version ?? 0)
    const parameters = { cluster_id: selectedClusterId, osd_id: id, zap: false }
    Modal.confirm({
      title: `删除 OSD ${id}`,
      content: '该操作为高风险操作，确认后将直接执行删除操作。',
      okText: '提交删除',
      okType: 'danger',
      cancelText: '取消',
      async onOk() {
        await operationMutation.run(() => mutateResource('/osd', 'DELETE', parameters, { ifMatch: generation }), false)
        window.setTimeout(() => {
          message.success('OSD 删除执行成功')
          refresh({ showLoading: false })
        })
      }
    })
  }

  return (
    <Page title="OSD管理" loading={loading} error={error}>
      <Card
        className="page-surface-card"
        title="OSD管理"
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} loading={refreshingOSDs} onClick={refreshOSDData}>刷新</Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setDeploymentOpen(true)}>OSD 部署</Button>
          </Space>
        }
      >
        <Space direction="vertical" size={16} className="page-stack">
        <section className="embedded-panel">
          <div className="embedded-panel-title">OSD Flags</div>
          {(data?.flags ?? []).length ? data?.flags.map((flag) => <Tag key={flag}>{flag}</Tag>) : <span className="muted">未设置 OSD flags</span>}
        </section>
        <section className="embedded-panel">
          <DataTable
            data={data?.osds ?? []}
            filterOptions={osdTableFilters.filterOptions}
            filteredValues={osdTableFilters.filters}
            onFilterChange={osdTableFilters.handleFilterChange}
            rowKeyCandidates={['id', 'osd', 'service_id', 'name']}
            columns={[
              { key: 'id', title: 'ID' },
              { key: 'host', title: '主机' },
              { key: 'state', title: '状态' },
              { key: 'up', title: 'Up' },
              { key: 'in', title: 'In' },
              { key: 'device_class', title: '设备类型' },
              { key: 'stats', title: '容量/统计' },
              {
                key: 'actions',
                title: '操作',
                filterKey: false,
                render: (_, row) => {
                  const id = osdID(row)
                  return (
                    <TableActions>
                      <TableAction loading={pendingOSDAction === `${id}:in`} disabled={Boolean(pendingOSDAction) && pendingOSDAction !== `${id}:in`} onClick={() => runOSDAction(id, 'in')}>In</TableAction>
                      <TableAction loading={pendingOSDAction === `${id}:out`} disabled={Boolean(pendingOSDAction) && pendingOSDAction !== `${id}:out`} onClick={() => runOSDAction(id, 'out')}>Out</TableAction>
                      <TableAction loading={pendingOSDAction === `${id}:scrub`} disabled={Boolean(pendingOSDAction) && pendingOSDAction !== `${id}:scrub`} onClick={() => runOSDAction(id, 'scrub')}>Scrub</TableAction>
                      <TableAction disabled={Boolean(pendingOSDAction)} onClick={() => runOSDAction(id, 'reweight')}>权重</TableAction>
                      <TableAction danger disabled={Boolean(pendingOSDAction)} onClick={() => deleteOSD(row)}>删除</TableAction>
                    </TableActions>
                  )
                }
              }
            ]}
          />
        </section>
        </Space>
      </Card>
      <OSDDeploymentModal open={deploymentOpen} onClose={() => setDeploymentOpen(false)} refresh={refresh} />
    </Page>
  )
}

function OSDDeploymentModal({ open, onClose, refresh }: { open: boolean; onClose: () => void; refresh: (options?: { showLoading?: boolean }) => void }) {
  const { selectedClusterId } = useClusterContext()
  const [submitting, setSubmitting] = useState(false)
  const [form] = Form.useForm<{
    service_id?: string
    host_pattern?: string
    all?: boolean
    paths?: string
    rotational?: boolean
    model?: string
    vendor?: string
    size?: string
  }>()
  const operationMutation = useMutationOperation()

  async function submit(values: {
    service_id?: string
    host_pattern?: string
    all?: boolean
    paths?: string
    rotational?: boolean
    model?: string
    vendor?: string
    size?: string
  }, mode: 'preview' | 'create') {
    if (!selectedClusterId || submitting) {
      return
    }
    const payload = osdDeploymentPayload(values, selectedClusterId)
    setSubmitting(true)
    try {
      if (mode === 'preview') {
        await operationMutation.run(() => mutateResource('/osd/deployment/preview', 'POST', payload), false)
      } else {
        Modal.confirm({
          title: '创建 OSD 部署',
          content: '该操作为高风险操作，确认后将直接执行部署。',
          okText: '提交创建',
          okType: 'danger',
          cancelText: '取消',
          async onOk() {
            await operationMutation.run(() => mutateResource('/osd/deployment', 'POST', payload), false)
            window.setTimeout(() => {
              onClose()
              message.success('OSD 部署执行成功')
              refresh({ showLoading: false })
            })
          }
        })
        return
      }
      onClose()
      message.success('OSD 部署预览执行成功')
      refresh({ showLoading: false })
    } finally {
      setSubmitting(false)
    }
  }

  function finishPreview() {
    void form.validateFields().then((values) => submit(values, 'preview'))
  }

  function finishCreate() {
    void form.validateFields().then((values) => submit(values, 'create'))
  }

  return (
    <DraggableModal
      title="OSD 部署"
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="cancel" onClick={onClose}>取消</Button>,
        <Button key="preview" loading={submitting} onClick={finishPreview}>预览</Button>,
        <Button key="create" type="primary" loading={submitting} onClick={finishCreate}>创建</Button>
      ]}
      destroyOnClose
    >
      <Form form={form} layout="vertical" initialValues={{ all: false }} preserve={false}>
        <Form.Item name="service_id" label="Service ID">
          <Input />
        </Form.Item>
        <Form.Item name="host_pattern" label="Host Pattern">
          <Input />
        </Form.Item>
        <Form.Item name="paths" label="设备路径">
          <Input placeholder="/dev/sdb,/dev/sdc" />
        </Form.Item>
        <Form.Item name="all" label="所有可用设备" valuePropName="checked">
          <Switch />
        </Form.Item>
        <Form.Item name="rotational" label="Rotational" valuePropName="checked">
          <Switch />
        </Form.Item>
        <Form.Item name="model" label="Model">
          <Input />
        </Form.Item>
        <Form.Item name="vendor" label="Vendor">
          <Input />
        </Form.Item>
        <Form.Item name="size" label="Size">
          <Input />
        </Form.Item>
      </Form>
    </DraggableModal>
  )
}

export function DeviceManagementPage() {
  const navigate = useNavigate()
  const { selectedClusterId } = useClusterContext()
  const deviceTableFilters = useResourceTableFilters({
    path: '/host/devices',
    fields: ['hostname', 'path', 'device_id', 'size_display', 'device_type', 'usage_state'],
    clusterId: selectedClusterId
  })
  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return {
        items: [],
        observedAt: null,
        stale: false,
        staleReason: null
      }
    }
    return listResource('/host/devices', selectedClusterId, {
      filters: deviceTableFilters.filters
    })
  }, [deviceTableFilters.filters, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const [refreshingDevices, setRefreshingDevices] = useState(false)
  const operationMutation = useMutationOperation()
  const deviceRows = useMemo(() => (data?.items ?? []).map(normalizeDeviceRow), [data?.items])

  async function refreshDeviceData() {
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    setRefreshingDevices(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kind: 'device' }), '刷新成功')
      await refresh()
    } finally {
      setRefreshingDevices(false)
    }
  }

  return (
    <Page title="设备列表" loading={loading} error={error}>
      <Card
        className="page-surface-card"
        title="设备列表"
        extra={<Button icon={<ReloadOutlined />} loading={refreshingDevices || loading} onClick={refreshDeviceData}>刷新</Button>}
      >
        <DataTable
          data={deviceRows}
          filterOptions={deviceTableFilters.filterOptions}
          filteredValues={deviceTableFilters.filters}
          onFilterChange={deviceTableFilters.handleFilterChange}
          footer={<ResourceMetaBar observedAt={data?.observedAt} stale={data?.stale} staleReason={data?.staleReason} />}
          rowKeyCandidates={['natural_key', 'device_id', 'path', 'name']}
          columns={[
            { key: 'hostname', title: '主机' },
            { key: 'path', title: '路径' },
            { key: 'device_id', title: '设备 ID' },
            { key: 'size_display', title: '容量', filterKey: 'size_display' },
            { key: 'device_type', title: '类型' },
            { key: 'usage_label', title: '状态', filterKey: 'usage_state', render: (_, row) => renderDeviceUsage(row) },
            {
              key: 'actions',
              title: '操作',
              filterKey: false,
              render: (_, row) => {
                const id = deviceID(row)
                return (
                  <TableActions>
                    <TableAction disabled={!id} onClick={() => navigate(deviceDetailPath(id))}>详情</TableAction>
                  </TableActions>
                )
              }
            }
          ]}
        />
      </Card>
    </Page>
  )
}

export function DeviceDetailPage() {
  const navigate = useNavigate()
  const { deviceId = '' } = useParams()
  const { selectedClusterId } = useClusterContext()
  const decodedDeviceId = safeDecodeRouteParam(deviceId)
  const loader = useCallback(async () => {
    if (!selectedClusterId || !decodedDeviceId) {
      return {
        device: null,
        observedAt: null,
        stale: false,
        staleReason: null
      }
    }
    const payload = await listResource('/host/devices', selectedClusterId, { filters: { device_id: [decodedDeviceId] } })
    const rows = payload.items.map(normalizeDeviceRow)
    return {
      device: rows.find((row) => deviceID(row) === decodedDeviceId) ?? null,
      observedAt: payload.observedAt,
      stale: payload.stale,
      staleReason: payload.staleReason
    }
  }, [decodedDeviceId, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const device = data?.device
  const currentDeviceHost = device ? deviceHost(device) : ''
  const currentDevicePath = device ? devicePath(device) : ''
  const [pendingDeviceAction, setPendingDeviceAction] = useState('')
  const [refreshingDevice, setRefreshingDevice] = useState(false)
  const operationMutation = useMutationOperation()

  async function refreshDeviceDetail() {
    if (!selectedClusterId || refreshingDevice) {
      return
    }
    setRefreshingDevice(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kind: 'device' }), '刷新成功')
      await refresh()
    } finally {
      setRefreshingDevice(false)
    }
  }

  function confirmIdentify(state: 'on' | 'off') {
    if (!device) {
      return
    }
    const isOn = state === 'on'
    Modal.confirm({
      title: `${isOn ? '点灯' : '关灯'}设备 ${currentDeviceHost}:${currentDevicePath}`,
      content: `确认后将对主机 ${currentDeviceHost} 的设备 ${currentDevicePath} 执行${isOn ? '点灯' : '关灯'}操作。`,
      okText: `确认${isOn ? '点灯' : '关灯'}`,
      cancelText: '取消',
      async onOk() {
        await identify(state)
      }
    })
  }

  async function identify(state: 'on' | 'off', light: 'ident' | 'fault' = 'ident') {
    if (!selectedClusterId || !device) {
      message.error('请先选择集群')
      return
    }
    const pendingKey = `${currentDeviceHost}:${currentDevicePath}:identify:${state}:${light}`
    if (!currentDeviceHost || !currentDevicePath || pendingDeviceAction) {
      return
    }
    setPendingDeviceAction(pendingKey)
    try {
      await operationMutation.run(() => mutateResource('/host/device/identify', 'POST', {
        cluster_id: selectedClusterId,
        host: currentDeviceHost,
        device: currentDevicePath,
        state,
        light
      }), state === 'on' ? '设备点灯执行成功' : '设备关灯执行成功')
      await refresh({ showLoading: false })
    } finally {
      setPendingDeviceAction('')
    }
  }

  async function zap() {
    if (!selectedClusterId || !device) {
      message.error('请先选择集群')
      return
    }
    const generation = Number(device.resource_version ?? 0)
    const parameters = { cluster_id: selectedClusterId, host: currentDeviceHost, device: currentDevicePath }
    const pendingKey = `${currentDeviceHost}:${currentDevicePath}:zap`
    Modal.confirm({
      title: `擦除设备 ${currentDeviceHost}:${currentDevicePath}`,
      content: `该操作会清理主机 ${currentDeviceHost} 的设备 ${currentDevicePath} 数据，为高风险操作，确认后将直接执行。`,
      okText: '提交擦除',
      okType: 'danger',
      cancelText: '取消',
      async onOk() {
        if (pendingDeviceAction) {
          return
        }
        setPendingDeviceAction(pendingKey)
        try {
          await operationMutation.run(() => mutateResource('/host/device/zap', 'POST', parameters, { ifMatch: generation }), false)
          window.setTimeout(() => {
            message.success('设备擦除执行成功')
            refresh({ showLoading: false })
          })
        } finally {
          setPendingDeviceAction('')
        }
      }
    })
  }

  return (
    <Page title="设备详情" loading={loading} error={error}>
      <Space direction="vertical" size={16} className="page-stack">
        <Card
          className="page-surface-card"
          title="基础信息"
          extra={
            <Space className="host-detail-actions">
              <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/cluster/host/device')}>返回</Button>
              <Button icon={<ReloadOutlined />} loading={refreshingDevice || loading} onClick={refreshDeviceDetail}>刷新</Button>
            </Space>
          }
        >
          {device ? (
            <Descriptions className="host-detail-descriptions" size="small" column={twoColumnDescriptions} bordered>
              <Descriptions.Item label="主机">{textValue(device.hostname)}</Descriptions.Item>
              <Descriptions.Item label="路径">{textValue(device.path)}</Descriptions.Item>
              <Descriptions.Item label="类型">{textValue(device.device_type)}</Descriptions.Item>
              <Descriptions.Item label="容量">{textValue(device.size_display)}</Descriptions.Item>
              <Descriptions.Item label="厂商">{textValue(device.vendor)}</Descriptions.Item>
              <Descriptions.Item label="设备 ID">{textValue(device.device_id ?? device.id ?? device.name)}</Descriptions.Item>
              <Descriptions.Item label="序列号">{textValue(device.serial_number ?? device.serial)}</Descriptions.Item>
              <Descriptions.Item label="状态">{renderDeviceUsage(device)}</Descriptions.Item>
              <Descriptions.Item label="说明" span={2}>{renderDeviceNotes(device.usage_notes)}</Descriptions.Item>
              <Descriptions.Item label="创建时间">{formatDateTime(device.created_at)}</Descriptions.Item>
              <Descriptions.Item label="更新时间">{formatDateTime(device.updated_at)}</Descriptions.Item>
            </Descriptions>
          ) : (
            <Text type="secondary">暂无设备详情</Text>
          )}
        </Card>

        <Card className="page-surface-card" title="设备操作">
          <Space wrap>
            <Button
              icon={<BulbOutlined />}
              loading={pendingDeviceAction === `${currentDeviceHost}:${currentDevicePath}:identify:on:ident`}
              disabled={!device || Boolean(pendingDeviceAction)}
              onClick={() => confirmIdentify('on')}
            >
              点灯
            </Button>
            <Button
              icon={<PoweroffOutlined />}
              loading={pendingDeviceAction === `${currentDeviceHost}:${currentDevicePath}:identify:off:ident`}
              disabled={!device || Boolean(pendingDeviceAction)}
              onClick={() => confirmIdentify('off')}
            >
              关灯
            </Button>
            <Button
              danger
              icon={<DeleteOutlined />}
              loading={pendingDeviceAction === `${currentDeviceHost}:${currentDevicePath}:zap`}
              disabled={!device || Boolean(pendingDeviceAction)}
              onClick={zap}
            >
              擦除
            </Button>
          </Space>
        </Card>
      </Space>
    </Page>
  )
}

export function MdsManagementPage() {
  const { selectedClusterId } = useClusterContext()
  const serviceTableFilters = useResourceTableFilters({
    path: '/services',
    fields: ['service_name', 'service_type', 'placement', 'status', 'running', 'size'],
    clusterId: selectedClusterId
  })
  const daemonTableFilters = useResourceTableFilters({
    path: '/daemons',
    fields: ['daemon_name', 'daemon_type', 'hostname', 'status_desc', 'version'],
    clusterId: selectedClusterId
  })
  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return { services: [], daemons: [] }
    }
    const [services, daemons] = await Promise.all([
      listResource('/services', selectedClusterId, {
        filters: mergeResourceFilters({ service_type: ['mds'] }, serviceTableFilters.filters)
      }).then((payload) => payload.items),
      listResource('/daemons', selectedClusterId, {
        filters: mergeResourceFilters({ daemon_type: ['mds'] }, daemonTableFilters.filters)
      }).then((payload) => payload.items)
    ])
    return {
      services,
      daemons
    }
  }, [daemonTableFilters.filters, selectedClusterId, serviceTableFilters.filters])
  const { data, loading, error, refresh } = useResource(loader)

  return (
    <Page title="MDS管理" loading={loading} error={error}>
      <Card className="page-surface-card" title="MDS管理">
        <Tabs
          items={[
            {
              key: 'services',
              label: 'MDS服务',
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
                    { key: 'placement', title: '放置策略' },
                    { key: 'status', title: '状态' },
                    { key: 'running', title: '运行数' },
                    { key: 'size', title: '目标数' }
                  ]}
                />
                </div>
              )
            },
            {
              key: 'daemons',
              label: '守护进程',
              children: <DaemonTable data={data?.daemons ?? []} refresh={refresh} tableFilters={daemonTableFilters} />
            }
          ]}
        />
      </Card>
    </Page>
  )
}

function DaemonTable({
  data,
  refresh,
  tableFilters
}: {
  data: ApiRecord[]
  refresh: () => void
  tableFilters?: ReturnType<typeof useResourceTableFilters>
}) {
  const [pendingDaemonAction, setPendingDaemonAction] = useState('')
  const operationMutation = useMutationOperation()

  async function runAction(row: ApiRecord, action: string) {
    const name = textValue(row.daemon_name || row.name, '')
    const pendingKey = `${name}:${action}`
    if (!name || pendingDaemonAction) {
      return
    }
    setPendingDaemonAction(pendingKey)
    try {
      await operationMutation.run(() => applyDaemonAction(name, action, action === 'restart'), `Daemon ${action} 执行成功`)
      refresh()
    } finally {
      setPendingDaemonAction('')
    }
  }

  return (
    <div className="embedded-panel">
      <DataTable
        data={data}
        filterOptions={tableFilters?.filterOptions}
        filteredValues={tableFilters?.filters}
        onFilterChange={tableFilters?.handleFilterChange}
        rowKeyCandidates={['daemon_name', 'name', 'hostname']}
        columns={[
          { key: 'daemon_name', title: 'Daemon' },
          { key: 'daemon_type', title: '类型' },
          { key: 'hostname', title: '主机' },
          { key: 'status_desc', title: '状态' },
          { key: 'version', title: '版本' },
          {
            key: 'actions',
            title: '操作',
            filterKey: false,
            render: (_, row) => {
              const name = textValue(row.daemon_name || row.name, '')
              return (
                <TableActions>
                  <TableAction loading={pendingDaemonAction === `${name}:restart`} disabled={Boolean(pendingDaemonAction) && pendingDaemonAction !== `${name}:restart`} onClick={() => runAction(row, 'restart')}>重启</TableAction>
                  <TableAction loading={pendingDaemonAction === `${name}:start`} disabled={Boolean(pendingDaemonAction) && pendingDaemonAction !== `${name}:start`} onClick={() => runAction(row, 'start')}>启动</TableAction>
                  <TableAction danger loading={pendingDaemonAction === `${name}:stop`} disabled={Boolean(pendingDaemonAction) && pendingDaemonAction !== `${name}:stop`} onClick={() => runAction(row, 'stop')}>停止</TableAction>
                </TableActions>
              )
            }
          }
        ]}
      />
    </div>
  )
}

function ReweightForm({ osdID, refresh }: { osdID: string; refresh: (options?: { showLoading?: boolean }) => void }) {
  const [submitting, setSubmitting] = useState(false)
  const operationMutation = useMutationOperation()

  async function submit(values: { weight: number }) {
    if (submitting) {
      return
    }
    setSubmitting(true)
    try {
      await operationMutation.run(() => reweightOSD(osdID, values.weight), false)
      Modal.destroyAll()
      window.setTimeout(() => {
        message.success('OSD 权重调整执行成功')
        refresh({ showLoading: false })
      })
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Form layout="vertical" initialValues={{ weight: 1 }} onFinish={submit}>
      <Form.Item name="weight" label="权重" rules={[{ required: true }]}>
        <InputNumber min={0} max={1} step={0.01} precision={2} />
      </Form.Item>
      <Button type="primary" htmlType="submit" loading={submitting}>
        保存
      </Button>
    </Form>
  )
}

function osdID(row: ApiRecord) {
  return textValue(row.id ?? row.osd ?? row.service_id ?? row.name, '')
}

function osdDeploymentPayload(values: {
  service_id?: string
  host_pattern?: string
  all?: boolean
  paths?: string
  rotational?: boolean
  model?: string
  vendor?: string
  size?: string
}, clusterId: number) {
  const dataDevices: ApiRecord = {}
  if (values.all) {
    dataDevices.all = true
  }
  const paths = splitCSV(values.paths)
  if (paths.length > 0) {
    dataDevices.paths = paths
  }
  if (typeof values.rotational === 'boolean') {
    dataDevices.rotational = values.rotational
  }
  if (values.model) {
    dataDevices.model = values.model
  }
  if (values.vendor) {
    dataDevices.vendor = values.vendor
  }
  if (values.size) {
    dataDevices.size = values.size
  }
  if (Object.keys(dataDevices).length === 0) {
    dataDevices.all = true
  }
  return {
    cluster_id: clusterId,
    ...(values.service_id ? { service_id: values.service_id } : {}),
    ...(values.host_pattern ? { host_pattern: values.host_pattern } : {}),
    data_devices: dataDevices
  }
}

function splitCSV(value?: string) {
  return value?.split(',').map((item) => item.trim()).filter(Boolean) ?? []
}

function deviceHost(row: ApiRecord) {
  return textValue(row.hostname ?? row.host, '')
}

function devicePath(row: ApiRecord) {
  return textValue(row.path ?? row.device ?? row.name, '')
}

function deviceID(row: ApiRecord) {
  return textValue(row.device_id ?? row.id ?? row.natural_key, '')
}

function deviceDetailPath(id: string) {
  return `/cluster/host/device/${encodeURIComponent(id)}`
}

function safeDecodeRouteParam(value: string) {
  try {
    return decodeURIComponent(value)
  } catch {
    return value
  }
}

function normalizeDeviceRow(row: ApiRecord): ApiRecord {
  const usage = deviceUsage(row)
  const type = textValue(row.device_type ?? row.type, '')
  return {
    ...row,
    hostname: textValue(row.hostname ?? row.host, ''),
    path: devicePath(row),
    device_type: type ? type.toUpperCase() : '-',
    usage_state: usage.state,
    usage_label: usage.label,
    usage_notes: usage.notes,
    size_display: formatBytes(numberValue(row.size_bytes ?? row.size))
  }
}

function deviceUsage(row: ApiRecord) {
  const reasons = deviceReasonValues(row.rejected_reasons ?? row.reject_reasons ?? row.reasons)
  if (row.available === true) {
    return { state: 'available' as DeviceScope, label: '空闲可用', notes: [] }
  }
  if (reasons.some(isUsedDeviceReason)) {
    return { state: 'used' as DeviceScope, label: '已占用', notes: reasons.map(readableDeviceReason) }
  }
  return { state: 'unavailable' as DeviceScope, label: '不可用', notes: reasons.map(readableDeviceReason) }
}

function deviceReasonValues(value: unknown) {
  if (Array.isArray(value)) {
    return value.map((item) => textValue(item, '')).filter(Boolean)
  }

  const text = textValue(value, '')
  if (!text) {
    return []
  }

  return text.split(',').map((item) => item.trim()).filter(Boolean)
}

function isUsedDeviceReason(reason: string) {
  const normalized = reason.toLowerCase()
  return [
    'filesystem',
    'file system',
    'lvm',
    'mounted',
    'partition',
    'bluestore',
    'osd',
    'in use',
    'being used'
  ].some((keyword) => normalized.includes(keyword))
}

function readableDeviceReason(reason: string) {
  const normalized = reason.toLowerCase()
  if (normalized.includes('filesystem') || normalized.includes('file system')) {
    return '已有文件系统'
  }
  if (normalized.includes('lvm')) {
    return '已有 LVM'
  }
  if (normalized.includes('insufficient space')) {
    return 'VG 可用空间不足'
  }
  if (normalized.includes('mounted')) {
    return '已挂载'
  }
  if (normalized.includes('partition')) {
    return '已有分区'
  }
  if (normalized.includes('bluestore') || normalized.includes('osd')) {
    return '已有 OSD 数据'
  }
  if (normalized.includes('locked')) {
    return '设备被锁定'
  }
  if (normalized.includes('read-only') || normalized.includes('readonly')) {
    return '只读设备'
  }
  return reason
}

function renderDeviceUsage(row: ApiRecord) {
  const state = textValue(row.usage_state, '') as DeviceScope
  const label = textValue(row.usage_label)
  const colors: Record<DeviceScope, string> = {
    available: 'success',
    used: 'processing',
    unavailable: 'default'
  }
  return <Tag color={colors[state] ?? 'default'}>{label}</Tag>
}

function renderDeviceNotes(value: unknown) {
  if (!Array.isArray(value) || value.length === 0) {
    return '-'
  }

  return value.map((item) => <Tag key={textValue(item)}>{textValue(item)}</Tag>)
}

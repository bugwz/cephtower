import { DeleteOutlined, PlusOutlined, ReloadOutlined, ThunderboltOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, InputNumber, Modal, Space, Switch, Tabs, Tag } from 'antd'
import { useCallback, useState } from 'react'
import { textValue, type ApiRecord } from '../../api/client'
import {
  applyDaemonAction,
  listDaemons,
  listMgrModules,
  listMonitors,
  listResource,
  listOSDFlags,
  listOSDs,
  listServices,
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
import { useResource } from '../../hooks'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useClusterContext } from '../../state/ClusterContext'
import { message } from '../../utils/appMessage'
import { ClusterDetailPage } from './ClusterDetailPage'
import { ClusterPage } from './ClusterPage'
import { HostPage } from './HostPage'
import { ServicePage } from './ServicePage'

export { ClusterDetailPage, ClusterPage, HostPage, ServicePage }

export function MonManagementPage() {
  const loader = useCallback(async () => {
    const [monitor, daemons] = await Promise.all([listMonitors(), listDaemons('mon')])
    return {
      monitor,
      daemons,
      inQuorum: asRecords(monitor.in_quorum),
      outQuorum: asRecords(monitor.out_quorum)
    }
  }, [])
  const { data, loading, error, refresh } = useResource(loader)

  return (
    <Page title="MON管理" loading={loading} error={error}>
      <Card className="page-surface-card" title="MON管理">
        <Tabs
          items={[
            {
              key: 'quorum',
              label: '仲裁成员',
              children: (
                <div className="embedded-panel">
                <DataTable
                  data={data?.inQuorum ?? []}
                  rowKeyCandidates={['name', 'rank', 'addr']}
                  columns={[
                    { key: 'name', title: '名称' },
                    { key: 'rank', title: 'Rank' },
                    { key: 'public_addr', title: 'Public Addr' },
                    { key: 'priority', title: '优先级' },
                    { key: 'stats', title: '会话统计', render: (value) => textValue(value) }
                  ]}
                />
                </div>
              )
            },
            {
              key: 'out',
              label: '非仲裁成员',
              children: (
                <div className="embedded-panel">
                <DataTable
                  data={data?.outQuorum ?? []}
                  rowKeyCandidates={['name', 'rank', 'addr']}
                  columns={[
                    { key: 'name', title: '名称' },
                    { key: 'rank', title: 'Rank' },
                    { key: 'public_addr', title: 'Public Addr' },
                    { key: 'addr', title: '地址' }
                  ]}
                />
                </div>
              )
            },
            {
              key: 'daemons',
              label: '守护进程',
              children: <DaemonTable data={data?.daemons ?? []} refresh={refresh} />
            }
          ]}
        />
      </Card>
    </Page>
  )
}

export function MgrManagementPage() {
  const loader = useCallback(async () => {
    const [modules, daemons] = await Promise.all([listMgrModules(), listDaemons('mgr')])
    return { modules, daemons }
  }, [])
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
              children: <DaemonTable data={data?.daemons ?? []} refresh={refresh} />
            }
          ]}
        />
      </Card>
    </Page>
  )
}

export function OsdManagementPage() {
  const { selectedClusterId } = useClusterContext()
  const loader = useCallback(async () => {
    const [osds, flags] = await Promise.all([listOSDs(), listOSDFlags()])
    return { osds, flags }
  }, [])
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
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kinds: ['osd', 'osd_flag'] }), 'OSD 数据刷新已触发')
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
        await operationMutation.run(() => mutateResource('/osd', 'DELETE', parameters, { ifMatch: generation }), 'OSD 删除执行成功')
        refresh()
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
                render: (_, row) => {
                  const id = osdID(row)
                  return (
                    <Space>
                      <Button size="small" loading={pendingOSDAction === `${id}:in`} disabled={Boolean(pendingOSDAction) && pendingOSDAction !== `${id}:in`} onClick={() => runOSDAction(id, 'in')}>In</Button>
                      <Button size="small" loading={pendingOSDAction === `${id}:out`} disabled={Boolean(pendingOSDAction) && pendingOSDAction !== `${id}:out`} onClick={() => runOSDAction(id, 'out')}>Out</Button>
                      <Button size="small" icon={<ThunderboltOutlined />} loading={pendingOSDAction === `${id}:scrub`} disabled={Boolean(pendingOSDAction) && pendingOSDAction !== `${id}:scrub`} onClick={() => runOSDAction(id, 'scrub')}>Scrub</Button>
                      <Button size="small" disabled={Boolean(pendingOSDAction)} onClick={() => runOSDAction(id, 'reweight')}>权重</Button>
                      <Button size="small" danger icon={<DeleteOutlined />} disabled={Boolean(pendingOSDAction)} onClick={() => deleteOSD(row)}>删除</Button>
                    </Space>
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

function OSDDeploymentModal({ open, onClose, refresh }: { open: boolean; onClose: () => void; refresh: () => void }) {
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
        await operationMutation.run(() => mutateResource('/osd/deployment/preview', 'POST', payload), 'OSD 部署预览执行成功')
      } else {
        Modal.confirm({
          title: '创建 OSD 部署',
          content: '该操作为高风险操作，确认后将直接执行部署。',
          okText: '提交创建',
          okType: 'danger',
          cancelText: '取消',
          async onOk() {
            await operationMutation.run(() => mutateResource('/osd/deployment', 'POST', payload), 'OSD 部署执行成功')
            onClose()
            refresh()
          }
        })
        return
      }
      onClose()
      refresh()
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
  const { selectedClusterId } = useClusterContext()
  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return {
        items: [],
        observedAt: null,
        stale: false,
        staleReason: null
      }
    }
    return listResource('/host/devices', selectedClusterId)
  }, [selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const [pendingDeviceAction, setPendingDeviceAction] = useState('')
  const [refreshingDevices, setRefreshingDevices] = useState(false)
  const operationMutation = useMutationOperation()

  async function refreshDeviceData() {
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    setRefreshingDevices(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kind: 'device' }), '设备数据刷新已触发')
      await refresh()
    } finally {
      setRefreshingDevices(false)
    }
  }

  async function identify(row: ApiRecord, state: 'on' | 'off', light: 'ident' | 'fault' = 'ident') {
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    const host = deviceHost(row)
    const path = devicePath(row)
    const pendingKey = `${host}:${path}:identify:${state}:${light}`
    if (!host || !path || pendingDeviceAction) {
      return
    }
    setPendingDeviceAction(pendingKey)
    try {
      await operationMutation.run(() => mutateResource('/host/device/identify', 'POST', {
        cluster_id: selectedClusterId,
        host,
        device: path,
        state,
        light
      }), state === 'on' ? '设备点灯执行成功' : '设备关灯执行成功')
      refresh()
    } finally {
      setPendingDeviceAction('')
    }
  }

  async function zap(row: ApiRecord) {
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    const host = deviceHost(row)
    const path = devicePath(row)
    if (!host || !path) {
      message.error('无法识别设备主机或路径')
      return
    }
    const deviceID = encodePair(host, path)
    const generation = Number(row.resource_version ?? 0)
    const parameters = { cluster_id: selectedClusterId, device_id: deviceID }
    Modal.confirm({
      title: `擦除设备 ${host}:${path}`,
      content: '该操作会清理设备数据，为高风险操作，确认后将直接执行操作。',
      okText: '提交擦除',
      okType: 'danger',
      cancelText: '取消',
      async onOk() {
        await operationMutation.run(() => mutateResource('/device/zap', 'POST', parameters, { ifMatch: generation }), '设备擦除执行成功')
        refresh()
      }
    })
  }

  return (
    <Page title="设备管理" loading={loading} error={error}>
      <Card
        className="page-surface-card"
        title="设备管理"
        extra={<Button icon={<ReloadOutlined />} loading={refreshingDevices || loading} onClick={refreshDeviceData}>刷新</Button>}
      >
        <Space direction="vertical" size={16} className="page-stack">
          <ResourceMetaBar observedAt={data?.observedAt} stale={data?.stale} staleReason={data?.staleReason} />
          <DataTable
            data={data?.items ?? []}
            rowKeyCandidates={['natural_key', 'device_id', 'path', 'name']}
            columns={[
              { key: 'hostname', title: '主机' },
              { key: 'path', title: '路径' },
              { key: 'device_id', title: '设备 ID' },
              { key: 'available', title: '可用', render: (value) => <Tag color={value ? 'success' : 'default'}>{value ? '可用' : '不可用'}</Tag> },
              { key: 'rejected_reasons', title: '拒绝原因' },
              { key: 'device_type', title: '类型' },
              { key: 'model', title: '型号' },
              { key: 'size_bytes', title: '容量 bytes' },
              {
                key: 'actions',
                title: '操作',
                render: (_, row) => {
                  const host = deviceHost(row)
                  const path = devicePath(row)
                  const identOnKey = `${host}:${path}:identify:on:ident`
                  const identOffKey = `${host}:${path}:identify:off:ident`
                  return (
                    <Space>
                      <Button size="small" loading={pendingDeviceAction === identOnKey} disabled={Boolean(pendingDeviceAction) && pendingDeviceAction !== identOnKey} onClick={() => identify(row, 'on')}>点灯</Button>
                      <Button size="small" loading={pendingDeviceAction === identOffKey} disabled={Boolean(pendingDeviceAction) && pendingDeviceAction !== identOffKey} onClick={() => identify(row, 'off')}>关灯</Button>
                      <Button size="small" danger icon={<DeleteOutlined />} disabled={Boolean(pendingDeviceAction)} onClick={() => zap(row)}>擦除</Button>
                    </Space>
                  )
                }
              }
            ]}
          />
        </Space>
      </Card>
    </Page>
  )
}

export function MdsManagementPage() {
  const loader = useCallback(async () => {
    const [services, daemons] = await Promise.all([listServices(), listDaemons('mds')])
    return {
      services: services.filter((service) => textValue(service.service_type || service.type, '').toLowerCase() === 'mds'),
      daemons
    }
  }, [])
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
              children: <DaemonTable data={data?.daemons ?? []} refresh={refresh} />
            }
          ]}
        />
      </Card>
    </Page>
  )
}

function DaemonTable({ data, refresh }: { data: ApiRecord[]; refresh: () => void }) {
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
            render: (_, row) => {
              const name = textValue(row.daemon_name || row.name, '')
              return (
                <Space>
                  <Button size="small" icon={<ReloadOutlined />} loading={pendingDaemonAction === `${name}:restart`} disabled={Boolean(pendingDaemonAction) && pendingDaemonAction !== `${name}:restart`} onClick={() => runAction(row, 'restart')}>重启</Button>
                  <Button size="small" loading={pendingDaemonAction === `${name}:start`} disabled={Boolean(pendingDaemonAction) && pendingDaemonAction !== `${name}:start`} onClick={() => runAction(row, 'start')}>启动</Button>
                  <Button size="small" danger loading={pendingDaemonAction === `${name}:stop`} disabled={Boolean(pendingDaemonAction) && pendingDaemonAction !== `${name}:stop`} onClick={() => runAction(row, 'stop')}>停止</Button>
                </Space>
              )
            }
          }
        ]}
      />
    </div>
  )
}

function ReweightForm({ osdID, refresh }: { osdID: string; refresh: () => void }) {
  const [submitting, setSubmitting] = useState(false)
  const operationMutation = useMutationOperation()

  async function submit(values: { weight: number }) {
    if (submitting) {
      return
    }
    setSubmitting(true)
    try {
      await operationMutation.run(() => reweightOSD(osdID, values.weight), 'OSD 权重调整执行成功')
      Modal.destroyAll()
      refresh()
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

function asRecords(value: unknown): ApiRecord[] {
  return Array.isArray(value) ? value.filter((item): item is ApiRecord => typeof item === 'object' && item !== null && !Array.isArray(item)) : []
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

function encodePair(left: string, right: string) {
  const bytes = new TextEncoder().encode(`${left}\u0000${right}`)
  let binary = ''
  bytes.forEach((byte) => {
    binary += String.fromCharCode(byte)
  })
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '')
}

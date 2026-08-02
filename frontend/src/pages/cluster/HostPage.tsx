import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, InputNumber, Select, Space, Tag } from 'antd'
import { useCallback, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { isRecord, numberValue, textValue, type ApiRecord } from '../../api/client'
import { getHostSSH, listDaemons, listResource, mutateResource, refreshResource, saveHostSSH, type HostSSHPayload } from '../../api/resource'
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

interface HostFormValues {
  hostname: string
  address?: string
}

interface HostSSHFormValues {
  hostname: string
  ssh_address: string
  ssh_port?: number
  ssh_user: string
  ssh_auth_method: string
  ssh_password?: string
  ssh_private_key?: string
  ssh_key_passphrase?: string
  notes?: string
}

export function HostPage() {
  const navigate = useNavigate()
  const { selectedClusterId } = useClusterContext()
  const hostTableFilters = useResourceTableFilters({
    path: '/hosts',
    fields: ['hostname', 'address', 'system', 'kernel_release', 'status'],
    clusterId: selectedClusterId
  })
  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return { hosts: [], observedAt: null, stale: false, staleReason: null }
    }
    const [hostList, daemons, devices] = await Promise.all([
      listResource('/hosts', selectedClusterId, { filters: hostTableFilters.filters }),
      listDaemons(),
      listResource('/host/devices').then((payload) => payload.items)
    ])
    return {
      hosts: hostList.items.map((host) => normalizeHostRow(host, daemons, devices)),
      observedAt: hostList.observedAt,
      stale: hostList.stale,
      staleReason: hostList.staleReason
    }
  }, [hostTableFilters.filters, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const [form] = Form.useForm<HostFormValues>()
  const [sshForm] = Form.useForm<HostSSHFormValues>()
  const [formOpen, setFormOpen] = useState(false)
  const [sshOpen, setSSHOpen] = useState(false)
  const [sshLoading, setSSHLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [sshSubmitting, setSSHSubmitting] = useState(false)
  const [editingHost, setEditingHost] = useState<ApiRecord | null>(null)
  const [refreshingHosts, setRefreshingHosts] = useState(false)
  const operationMutation = useMutationOperation()
  const sshAuthMethod = Form.useWatch('ssh_auth_method', sshForm) ?? 'password'

  async function refreshHostData() {
    if (refreshingHosts) {
      return
    }
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    setRefreshingHosts(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kinds: ['host', 'daemon', 'device'] }), '刷新成功')
      await refresh()
    } finally {
      setRefreshingHosts(false)
    }
  }

  function openCreate() {
    form.resetFields()
    setFormOpen(true)
  }

  async function openSSHEdit(row: ApiRecord) {
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    const name = hostName(row)
    setEditingHost(row)
    setSSHOpen(true)
    sshForm.resetFields()
    sshForm.setFieldsValue({
      hostname: name,
      ssh_address: textValue(row.address ?? row.addr ?? row.ip ?? row.public_addr, ''),
      ssh_port: 22,
      ssh_user: 'root',
      ssh_auth_method: 'password'
    })
    setSSHLoading(true)
    try {
      const ssh = await getHostSSH(name, selectedClusterId)
      sshForm.setFieldsValue({
        hostname: name,
        ssh_address: textValue(ssh.ssh_address, textValue(row.address ?? row.addr ?? row.ip ?? row.public_addr, '')),
        ssh_port: numberValue(ssh.ssh_port) ?? 22,
        ssh_user: textValue(ssh.ssh_user, 'root'),
        ssh_auth_method: textValue(ssh.ssh_auth_method, 'password'),
        notes: textValue(ssh.notes, '')
      })
    } finally {
      setSSHLoading(false)
    }
  }

  async function submitHost(values: HostFormValues) {
    if (!selectedClusterId || submitting) {
      return
    }
    setSubmitting(true)
    try {
      await operationMutation.run(() => mutateResource('/host', 'POST', {
        cluster_id: selectedClusterId,
        hostname: values.hostname,
        ...(values.address ? { address: values.address } : {})
      }), false)
      setFormOpen(false)
      message.success('主机添加执行成功')
      void refresh({ showLoading: false })
    } finally {
      setSubmitting(false)
    }
  }

  async function submitHostSSH(values: HostSSHFormValues) {
    if (!selectedClusterId || sshSubmitting) {
      return
    }
    setSSHSubmitting(true)
    try {
      const payload: HostSSHPayload = {
        hostname: values.hostname,
        ssh_address: values.ssh_address,
        ssh_port: values.ssh_port,
        ssh_user: values.ssh_user,
        ssh_auth_method: values.ssh_auth_method,
        ...(values.notes?.trim() ? { notes: values.notes.trim() } : {})
      }
      if (values.ssh_password?.trim()) {
        payload.ssh_password = values.ssh_password.trim()
      }
      if (values.ssh_private_key?.trim()) {
        payload.ssh_private_key = values.ssh_private_key.trim()
      }
      if (values.ssh_key_passphrase?.trim()) {
        payload.ssh_key_passphrase = values.ssh_key_passphrase.trim()
      }
      await operationMutation.run(() => saveHostSSH(payload, selectedClusterId), false)
      setSSHOpen(false)
      setEditingHost(null)
      message.success('主机 SSH 信息已保存')
      void refresh({ showLoading: false })
    } finally {
      setSSHSubmitting(false)
    }
  }

  return (
    <Page
      title="主机列表"
      loading={loading}
      error={error}
    >
      <Card
        className="page-surface-card"
        title="主机列表"
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} loading={refreshingHosts} onClick={refreshHostData}>刷新</Button>
            <Button type="primary" icon={<PlusOutlined />} disabled={!selectedClusterId} onClick={openCreate}>新增主机</Button>
          </Space>
        }
      >
        <DataTable
          data={data?.hosts ?? []}
          filterOptions={hostTableFilters.filterOptions}
          filteredValues={hostTableFilters.filters}
          onFilterChange={hostTableFilters.handleFilterChange}
          footer={<ResourceMetaBar observedAt={data?.observedAt} stale={data?.stale} staleReason={data?.staleReason} />}
          rowKeyCandidates={['hostname', 'name', 'address', 'addr']}
          columns={[
            { key: 'hostname', title: '主机名' },
            { key: 'address_display', title: '地址', filterKey: 'address' },
            { key: 'system_display', title: '系统', filterKey: 'system' },
            { key: 'kernel_display', title: '内核版本', filterKey: 'kernel_release' },
            { key: 'daemon_count_display', title: '守护进程', filterKey: false },
            { key: 'osd_count_display', title: 'OSD', filterKey: false },
            { key: 'storage_display', title: '磁盘容量', filterKey: false },
            { key: 'status_display', title: '状态', filterKey: 'status', render: (value) => <Tag color={value === '在线' ? 'success' : 'default'}>{textValue(value)}</Tag> },
            {
              key: 'actions',
              title: '操作',
              filterKey: false,
              render: (_, row) => {
                const name = hostName(row)
                return (
                  <TableActions>
                    <TableAction onClick={() => openSSHEdit(row)}>编辑</TableAction>
                    <TableAction onClick={() => navigate(`/cluster/host/${encodeURIComponent(name)}`)}>详情</TableAction>
                  </TableActions>
                )
              }
            }
          ]}
        />
      </Card>
      <DraggableModal
        title="新增主机"
        open={formOpen}
        onCancel={() => setFormOpen(false)}
        onOk={() => form.submit()}
        okText="提交"
        confirmLoading={submitting}
        destroyOnClose
      >
        <Form form={form} layout="vertical" onFinish={submitHost}>
          <Form.Item name="hostname" label="主机名" rules={[{ required: true, message: '请输入主机名' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="address" label="地址">
            <Input />
          </Form.Item>
        </Form>
      </DraggableModal>
      <DraggableModal
        title={editingHost ? `主机 SSH 信息 - ${hostName(editingHost)}` : '主机 SSH 信息'}
        open={sshOpen}
        onCancel={() => {
          setSSHOpen(false)
          setEditingHost(null)
        }}
        onOk={() => sshForm.submit()}
        okText="保存"
        confirmLoading={sshSubmitting}
        destroyOnClose
      >
        <Form form={sshForm} layout="vertical" disabled={sshLoading} onFinish={submitHostSSH}>
          <Form.Item name="hostname" label="主机名" rules={[{ required: true, message: '请输入主机名' }]}>
            <Input disabled />
          </Form.Item>
          <Form.Item name="ssh_address" label="登录地址" rules={[{ required: true, message: '请输入登录地址' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="ssh_port" label="端口" rules={[{ required: true, message: '请输入端口' }]}>
            <InputNumber min={1} max={65535} precision={0} className="full-width-control" />
          </Form.Item>
          <Form.Item name="ssh_user" label="账号" rules={[{ required: true, message: '请输入账号' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="ssh_auth_method" label="认证方式" rules={[{ required: true, message: '请选择认证方式' }]}>
            <Select
              options={[
                { label: '密码', value: 'password' },
                { label: '私钥', value: 'private_key' }
              ]}
            />
          </Form.Item>
          {sshAuthMethod === 'password' ? (
            <Form.Item name="ssh_password" label="密码">
              <Input.Password placeholder="留空表示不修改已保存的密码" />
            </Form.Item>
          ) : (
            <>
              <Form.Item name="ssh_private_key" label="私钥">
                <Input.TextArea rows={5} placeholder="留空表示不修改已保存的私钥" />
              </Form.Item>
              <Form.Item name="ssh_key_passphrase" label="私钥口令">
                <Input.Password placeholder="留空表示不修改已保存的私钥口令" />
              </Form.Item>
            </>
          )}
          <Form.Item name="notes" label="备注">
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </DraggableModal>
    </Page>
  )
}

export function hostName(row: ApiRecord) {
  return textValue(row.hostname ?? row.host ?? row.name, '')
}

export function normalizeHostRow(row: ApiRecord, daemons: ApiRecord[] = [], devices: ApiRecord[] = []): ApiRecord {
  const name = hostName(row)
  const hostDaemons = daemons.filter((daemon) => textValue(daemon.hostname ?? daemon.host, '') === name)
  const hostDevices = devices.filter((device) => textValue(device.hostname ?? device.host, '') === name)
  const daemonNames = hostDaemons
    .map((daemon) => textValue(daemon.daemon_name ?? daemon.name ?? daemon.daemon_type ?? daemon.type, ''))
    .filter(Boolean)
  const serviceInstances = serviceInstanceRows(row.service_instances)
  const osdCount = serviceInstanceCount(serviceInstances, 'osd') ?? hostDaemons.filter((daemon) => daemonType(daemon) === 'osd').length
  const storageBytes = hostDevices.reduce((total, device) => total + (numberValue(device.size_bytes) ?? 0), 0)

  return {
    ...row,
    hostname: name,
    address_display: textValue(row.address ?? row.addr ?? row.ip ?? row.public_addr),
    status_display: hostStatus(row),
    system_display: hostSystem(row),
    platform_display: textValue(row.platform ?? hostFact(row, 'platform')),
    distro_display: textValue(row.distro ?? hostFact(row, 'distro')),
    kernel_display: hostKernel(row),
    kernel_build_display: textValue(row.kernel_build ?? hostFact(row, 'kernel_build')),
    architecture_display: textValue(row.arch ?? hostFact(row, 'arch', 'machine')),
    cpu_display: hostCPU(row),
    memory_display: formatBytes(numberValue(row.memory_bytes ?? hostFact(row, 'memory_bytes', 'memory_total'))),
    daemon_count_display: hostDaemons.length || daemonNames.length || '-',
    osd_count_display: osdCount || '-',
    disk_count_display: hostDevices.length || '-',
    storage_display: storageBytes > 0 ? formatBytes(storageBytes) : '-',
    service_instances: serviceInstances.length ? serviceInstances : daemonNames
  }
}

export function hostStatus(row: ApiRecord) {
  const status = textValue(row.status ?? row.status_desc ?? row.host_status, '')
  return status || '在线'
}

export function serviceInstanceRows(value: unknown): ApiRecord[] {
  if (!Array.isArray(value)) {
    return []
  }
  return value.filter(isRecord).map((item) => ({
    type: textValue(item.type ?? item.service_type, ''),
    count: numberValue(item.count) ?? 0
  })).filter((item) => item.type)
}

export function hostSystem(row: ApiRecord) {
  return textValue(row.system ?? hostFact(row, 'system', 'os', 'distro', 'distribution'))
}

export function hostKernel(row: ApiRecord) {
  return textValue(row.kernel_release ?? hostFact(row, 'kernel_release', 'kernel'))
}

export function hostCPU(row: ApiRecord) {
  const model = textValue(row.cpu_model ?? hostFact(row, 'cpu_model', 'processor_model', 'model_name'), '')
  const cores = numberValue(row.cpu_cores ?? hostFact(row, 'cpu_cores', 'cpu_count', 'processor_count', 'cpus'))
  if (model && cores) {
    return `${model} / ${cores} 核`
  }
  if (model) {
    return model
  }
  return cores ? `${cores} 核` : '-'
}

export function formatBytes(value?: number) {
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
  const digits = size >= 10 || unitIndex === 0 ? 0 : 1
  return `${size.toFixed(digits)} ${units[unitIndex]}`
}

function hostFact(row: ApiRecord, ...keys: string[]) {
  const facts = row.facts
  if (!isRecord(facts)) {
    return undefined
  }
  for (const key of keys) {
    if (facts[key] !== undefined) {
      return facts[key]
    }
    const matchedKey = Object.keys(facts).find((item) => item.toLowerCase() === key.toLowerCase())
    if (matchedKey) {
      return facts[matchedKey]
    }
  }
  return undefined
}

function serviceInstanceCount(instances: ApiRecord[], type: string) {
  const match = instances.find((item) => textValue(item.type, '').toLowerCase() === type)
  return numberValue(match?.count)
}

function daemonType(daemon: ApiRecord) {
  return textValue(daemon.daemon_type ?? daemon.type ?? daemon.service_type, '').toLowerCase()
}

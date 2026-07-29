import { DeleteOutlined, EditOutlined, PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Modal, Select, Space, Tabs, Tag } from 'antd'
import { useCallback, useState } from 'react'
import { textValue, type ApiRecord } from '../../api/client'
import { listHosts, listOSDFlags, listOSDs, mutateResource, refreshResource } from '../../api/resource'
import { DataTable } from '../../components/DataTable'
import { DraggableModal } from '../../components/DraggableModal'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useClusterContext } from '../../state/ClusterContext'
import { message } from '../../utils/appMessage'

interface HostFormValues {
  hostname: string
  address?: string
  label?: string
  action?: 'add' | 'rm'
}

export function HostPage() {
  const { selectedClusterId } = useClusterContext()
  const loader = useCallback(async () => {
    const [hosts, osds, flags] = await Promise.all([listHosts(), listOSDs(), listOSDFlags()])
    return { hosts, osds, flags }
  }, [])
  const { data, loading, error, refresh } = useResource(loader)
  const [form] = Form.useForm<HostFormValues>()
  const [formOpen, setFormOpen] = useState(false)
  const [editingHost, setEditingHost] = useState<ApiRecord | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [pendingAction, setPendingAction] = useState('')
  const [refreshingHosts, setRefreshingHosts] = useState(false)
  const operationMutation = useMutationOperation()

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
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kinds: ['host', 'osd', 'osd_flag'] }), '主机数据刷新已触发')
      await refresh()
    } finally {
      setRefreshingHosts(false)
    }
  }

  function openCreate() {
    setEditingHost(null)
    form.resetFields()
    form.setFieldsValue({ action: 'add' })
    setFormOpen(true)
  }

  function openEdit(row: ApiRecord) {
    setEditingHost(row)
    form.resetFields()
    form.setFieldsValue({
      hostname: hostName(row),
      address: textValue(row.addr ?? row.address, ''),
      action: 'add'
    })
    setFormOpen(true)
  }

  async function submitHost(values: HostFormValues) {
    if (!selectedClusterId || submitting) {
      return
    }
    setSubmitting(true)
    try {
      if (editingHost) {
        await operationMutation.run(() => mutateResource('/host', 'PATCH', {
          cluster_id: selectedClusterId,
          host: hostName(editingHost),
          ...(values.address ? { address: values.address } : {}),
          ...(values.label ? { label: values.label, action: values.action ?? 'add' } : {})
        }, { ifMatch: Number(editingHost.resource_version ?? 0) }), '主机更新执行成功')
      } else {
        await operationMutation.run(() => mutateResource('/host', 'POST', {
          cluster_id: selectedClusterId,
          hostname: values.hostname,
          ...(values.address ? { address: values.address } : {})
        }), '主机添加执行成功')
      }
      setFormOpen(false)
      await refresh()
    } finally {
      setSubmitting(false)
    }
  }

  async function runHostAction(row: ApiRecord, action: string) {
    if (!selectedClusterId || pendingAction) {
      return
    }
    const name = hostName(row)
    if (!name) {
      message.error('无法识别主机名')
      return
    }
    const pendingKey = `${name}:${action}`
    setPendingAction(pendingKey)
    try {
      await operationMutation.run(() => mutateResource('/host/action', 'POST', {
        cluster_id: selectedClusterId,
        host: name,
        action
      }), `主机 ${action} 执行成功`)
      await refresh()
    } finally {
      setPendingAction('')
    }
  }

  async function deleteHost(row: ApiRecord) {
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    const name = hostName(row)
    if (!name) {
      message.error('无法识别主机名')
      return
    }
    const generation = Number(row.resource_version ?? 0)
    const parameters = { cluster_id: selectedClusterId, host: name }
    Modal.confirm({
      title: `删除主机 ${name}`,
      content: '该操作为高风险操作，确认后将直接执行删除操作。',
      okText: '提交删除',
      okType: 'danger',
      cancelText: '取消',
      async onOk() {
        await operationMutation.run(() => mutateResource('/host', 'DELETE', parameters, { ifMatch: generation }), '主机删除执行成功')
        await refresh()
      }
    })
  }

  return (
    <Page
      title="主机管理"
      loading={loading}
      error={error}
    >
      <Card
        className="page-surface-card"
        title="主机管理"
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} loading={refreshingHosts} onClick={refreshHostData}>刷新</Button>
            <Button type="primary" icon={<PlusOutlined />} disabled={!selectedClusterId} onClick={openCreate}>新增主机</Button>
          </Space>
        }
      >
        <Tabs
          items={[
            {
              key: 'hosts',
              label: '主机',
              children: (
                <div className="embedded-panel">
                <DataTable
                  data={data?.hosts ?? []}
                  rowKeyCandidates={['hostname', 'addr']}
                  columns={[
                    { key: 'hostname', title: '主机名' },
                    { key: 'addr', title: '地址' },
                    { key: 'status', title: '状态' },
                    { key: 'ceph_version', title: 'Ceph 版本' },
                    { key: 'labels', title: '标签' },
                    { key: 'service_instances', title: '服务实例' },
                    {
                      key: 'actions',
                      title: '操作',
                      render: (_, row) => {
                        const name = hostName(row)
                        return (
                          <Space>
                            <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(row)}>编辑</Button>
                            <Button size="small" loading={pendingAction === `${name}:maintenance_enter`} disabled={Boolean(pendingAction) && pendingAction !== `${name}:maintenance_enter`} onClick={() => runHostAction(row, 'maintenance_enter')}>维护</Button>
                            <Button size="small" loading={pendingAction === `${name}:maintenance_exit`} disabled={Boolean(pendingAction) && pendingAction !== `${name}:maintenance_exit`} onClick={() => runHostAction(row, 'maintenance_exit')}>退出维护</Button>
                            <Button size="small" loading={pendingAction === `${name}:drain`} disabled={Boolean(pendingAction) && pendingAction !== `${name}:drain`} onClick={() => runHostAction(row, 'drain')}>Drain</Button>
                            <Button size="small" loading={pendingAction === `${name}:rescan`} disabled={Boolean(pendingAction) && pendingAction !== `${name}:rescan`} onClick={() => runHostAction(row, 'rescan')}>Rescan</Button>
                            <Button size="small" danger icon={<DeleteOutlined />} disabled={Boolean(pendingAction)} onClick={() => deleteHost(row)}>删除</Button>
                          </Space>
                        )
                      }
                    }
                  ]}
                />
                </div>
              )
            },
            {
              key: 'osds',
              label: 'OSD',
              children: (
                <div className="embedded-panel">
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
                    { key: 'stats', title: '容量/统计' }
                  ]}
                />
                </div>
              )
            },
            {
              key: 'flags',
              label: 'OSD Flags',
              children: (
                <div className="embedded-panel">
                {(data?.flags ?? []).length ? (
                  data?.flags.map((flag) => <Tag key={flag}>{flag}</Tag>)
                ) : (
                  <span className="muted">未设置 OSD flags</span>
                )}
                </div>
              )
            }
          ]}
        />
      </Card>
      <DraggableModal
        title={editingHost ? '编辑主机' : '新增主机'}
        open={formOpen}
        onCancel={() => setFormOpen(false)}
        onOk={() => form.submit()}
        okText="提交"
        confirmLoading={submitting}
        destroyOnClose
      >
        <Form form={form} layout="vertical" onFinish={submitHost}>
          <Form.Item name="hostname" label="主机名" rules={[{ required: !editingHost, message: '请输入主机名' }]}>
            <Input disabled={Boolean(editingHost)} />
          </Form.Item>
          <Form.Item name="address" label="地址">
            <Input />
          </Form.Item>
          {editingHost ? (
            <>
              <Form.Item name="label" label="标签">
                <Input />
              </Form.Item>
              <Form.Item name="action" label="标签操作">
                <Select
                  options={[
                    { label: '添加', value: 'add' },
                    { label: '移除', value: 'rm' }
                  ]}
                />
              </Form.Item>
            </>
          ) : null}
        </Form>
      </DraggableModal>
    </Page>
  )
}

function hostName(row: ApiRecord) {
  return textValue(row.hostname ?? row.host ?? row.name, '')
}

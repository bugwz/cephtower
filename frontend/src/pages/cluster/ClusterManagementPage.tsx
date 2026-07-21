import { EditOutlined, PlusOutlined, ReloadOutlined, SaveOutlined } from '@ant-design/icons'
import { Button, Card, Checkbox, Form, Input, InputNumber, Modal, Space, Switch, Table, Tag, Typography, message } from 'antd'
import { useCallback, useEffect, useState } from 'react'
import {
  createCluster,
  listClusters,
  updateCluster,
  type CephCluster,
  type CephClusterPayload
} from '../../api/clusters'
import { Page } from '../../components/Page'

const { Text } = Typography

interface ClusterFormValues {
  name: string
  description?: string
  fsid?: string
  enabled: boolean
  dashboard_enabled: boolean
  dashboard_base_url?: string
  dashboard_username?: string
  dashboard_password?: string
  dashboard_clear_secret?: boolean
  dashboard_insecure_tls?: boolean
  command_enabled: boolean
  command_bin?: string
  command_cluster?: string
  command_conf?: string
  command_name?: string
  command_keyring?: string
  command_keyring_content?: string
  command_clear_secret?: boolean
  command_timeout_seconds?: number
}

export function ClusterManagementPage() {
  const [clusters, setClusters] = useState<CephCluster[]>([])
  const [clusterLoading, setClusterLoading] = useState(true)
  const [clusterError, setClusterError] = useState('')
  const [clusterModalOpen, setClusterModalOpen] = useState(false)
  const [editingCluster, setEditingCluster] = useState<CephCluster | null>(null)
  const [form] = Form.useForm<ClusterFormValues>()

  const loadClusters = useCallback(async () => {
    setClusterLoading(true)
    setClusterError('')
    try {
      setClusters(await listClusters())
    } catch (err) {
      setClusterError(err instanceof Error ? err.message : '加载集群连接失败')
    } finally {
      setClusterLoading(false)
    }
  }, [])

  useEffect(() => {
    loadClusters()
  }, [loadClusters])

  function openCreateCluster() {
    setEditingCluster(null)
    form.resetFields()
    form.setFieldsValue(defaultClusterFormValues())
    setClusterModalOpen(true)
  }

  function openEditCluster(cluster: CephCluster) {
    setEditingCluster(cluster)
    form.setFieldsValue({
      name: cluster.name,
      description: cluster.description,
      fsid: cluster.fsid,
      enabled: cluster.enabled,
      dashboard_enabled: cluster.dashboard.enabled,
      dashboard_base_url: cluster.dashboard.base_url,
      dashboard_username: cluster.dashboard.username,
      dashboard_insecure_tls: cluster.dashboard.insecure_tls,
      dashboard_password: '',
      dashboard_clear_secret: false,
      command_enabled: cluster.command.enabled,
      command_bin: cluster.command.bin || 'ceph',
      command_cluster: cluster.command.cluster,
      command_conf: cluster.command.conf,
      command_name: cluster.command.name,
      command_keyring: cluster.command.keyring,
      command_keyring_content: '',
      command_clear_secret: false,
      command_timeout_seconds: cluster.command.timeout_seconds || 15
    })
    setClusterModalOpen(true)
  }

  async function submitCluster(values: ClusterFormValues) {
    const payload = toClusterPayload(values)
    const saved = editingCluster
      ? await updateCluster(editingCluster.id, payload)
      : await createCluster(payload)

    setClusters((current) => {
      if (!editingCluster) {
        return [...current, saved]
      }
      return current.map((cluster) => (cluster.id === saved.id ? saved : cluster))
    })
    setClusterModalOpen(false)
    form.resetFields()
    message.success(editingCluster ? '集群连接已更新' : '集群连接已创建')
  }

  return (
    <Page
      title="集群管理"
      loading={clusterLoading}
      error={clusterError}
    >
      <Card
        className="page-surface-card"
        title="集群管理"
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={loadClusters}>
              刷新
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreateCluster}>
              新建集群
            </Button>
          </Space>
        }
      >
        <Table
          size="middle"
          rowKey="id"
          dataSource={clusters}
          pagination={{ pageSize: 6, showSizeChanger: false }}
          scroll={{ x: true }}
          columns={[
            {
              title: '集群',
              key: 'cluster',
              render: (_, cluster) => (
                <div className="user-cell">
                  <Text strong>{cluster.name}</Text>
                  <Text type="secondary">{cluster.description || cluster.fsid || '未填写描述'}</Text>
                </div>
              )
            },
            {
              title: '状态',
              dataIndex: 'enabled',
              render: (enabled: boolean) => <Tag color={enabled ? 'success' : 'default'}>{enabled ? '启用' : '禁用'}</Tag>
            },
            {
              title: 'Dashboard API',
              dataIndex: 'dashboard',
              render: (dashboard: CephCluster['dashboard']) => (
                <Space>
                  <Tag color={dashboard.enabled ? 'processing' : 'default'}>{dashboard.enabled ? '启用' : '关闭'}</Tag>
                  {dashboard.password_set && <Tag color="gold">密码已保存</Tag>}
                </Space>
              )
            },
            {
              title: 'Ceph 命令',
              dataIndex: 'command',
              render: (command: CephCluster['command']) => (
                <Space>
                  <Tag color={command.enabled ? 'processing' : 'default'}>{command.enabled ? command.bin : '关闭'}</Tag>
                  {command.keyring_content_set && <Tag color="gold">Key 已保存</Tag>}
                </Space>
              )
            },
            {
              title: '更新时间',
              dataIndex: 'updated_at',
              render: (value: string) => new Date(value).toLocaleString()
            },
            {
              title: '操作',
              key: 'actions',
              render: (_, cluster) => (
                <Button icon={<EditOutlined />} onClick={() => openEditCluster(cluster)}>
                  编辑
                </Button>
              )
            }
          ]}
        />
      </Card>

      <Modal
        width={860}
        className="cluster-modal"
        title={editingCluster ? `编辑集群：${editingCluster.name}` : '新建集群'}
        open={clusterModalOpen}
        onCancel={() => setClusterModalOpen(false)}
        onOk={() => form.submit()}
        okText="保存"
        okButtonProps={{ icon: <SaveOutlined /> }}
        cancelText="取消"
        destroyOnClose
        maskClosable={false}
      >
        <Form form={form} layout="vertical" initialValues={defaultClusterFormValues()} onFinish={submitCluster} className="cluster-form">
          <section className="cluster-form-section">
            <div className="form-section-title">基础信息</div>
            <div className="cluster-form-grid">
              <Form.Item name="name" label="集群名称" rules={[{ required: true, message: '请输入集群名称' }]}>
                <Input placeholder="例如：production-ceph" />
              </Form.Item>
              <Form.Item name="fsid" label="FSID">
                <Input placeholder="Ceph FSID，可选" />
              </Form.Item>
              <Form.Item name="description" label="描述" className="form-field-full">
                <Input placeholder="业务用途、地域或维护说明" />
              </Form.Item>
              <Form.Item name="enabled" label="启用集群" valuePropName="checked" className="cluster-form-switch">
                <Switch />
              </Form.Item>
            </div>
          </section>

          <section className="cluster-form-section">
            <div className="cluster-form-section-head">
              <div className="form-section-title">Dashboard API</div>
              <Form.Item name="dashboard_enabled" valuePropName="checked" className="cluster-form-switch compact-form-item">
                <Switch />
              </Form.Item>
            </div>
            <div className="cluster-form-grid">
              <Form.Item name="dashboard_base_url" label="Dashboard 地址" className="form-field-full">
                <Input placeholder="https://ceph.example.com:8443" />
              </Form.Item>
              <Form.Item name="dashboard_username" label="用户名">
                <Input placeholder="admin" />
              </Form.Item>
              <Form.Item name="dashboard_password" label="密码">
                <Input.Password placeholder={editingCluster?.dashboard.password_set ? '留空则保持已保存密码' : 'Dashboard 登录密码'} />
              </Form.Item>
              <Form.Item name="dashboard_insecure_tls" label="跳过 TLS 校验" valuePropName="checked" className="cluster-form-switch">
                <Switch />
              </Form.Item>
              {editingCluster?.dashboard.password_set && (
                <Form.Item name="dashboard_clear_secret" valuePropName="checked" className="cluster-form-checkbox">
                  <Checkbox>清除已保存密码</Checkbox>
                </Form.Item>
              )}
            </div>
          </section>

          <section className="cluster-form-section">
            <div className="cluster-form-section-head">
              <div className="form-section-title">Ceph 命令</div>
              <Form.Item name="command_enabled" valuePropName="checked" className="cluster-form-switch compact-form-item">
                <Switch />
              </Form.Item>
            </div>
            <div className="cluster-form-grid">
              <Form.Item name="command_bin" label="ceph 可执行文件">
                <Input placeholder="ceph" />
              </Form.Item>
              <Form.Item name="command_timeout_seconds" label="命令超时">
                <InputNumber min={1} addonAfter="秒" className="full-width-input" />
              </Form.Item>
              <Form.Item name="command_cluster" label="集群参数">
                <Input placeholder="ceph" />
              </Form.Item>
              <Form.Item name="command_name" label="客户端名称">
                <Input placeholder="client.admin" />
              </Form.Item>
              <Form.Item name="command_conf" label="配置文件" className="form-field-full">
                <Input placeholder="/etc/ceph/ceph.conf" />
              </Form.Item>
              <Form.Item name="command_keyring" label="Keyring 路径" className="form-field-full">
                <Input placeholder="/etc/ceph/ceph.client.admin.keyring" />
              </Form.Item>
              <Form.Item name="command_keyring_content" label="Keyring 内容" className="form-field-full">
                <Input.TextArea
                  rows={3}
                  placeholder={editingCluster?.command.keyring_content_set ? '留空则保持已保存内容' : '可选：粘贴 keyring 内容'}
                />
              </Form.Item>
              {editingCluster?.command.keyring_content_set && (
                <Form.Item name="command_clear_secret" valuePropName="checked" className="cluster-form-checkbox">
                  <Checkbox>清除已保存 Keyring 内容</Checkbox>
                </Form.Item>
              )}
            </div>
          </section>
        </Form>
      </Modal>
    </Page>
  )
}

function defaultClusterFormValues(): Partial<ClusterFormValues> {
  return {
    enabled: true,
    dashboard_enabled: true,
    dashboard_insecure_tls: false,
    command_enabled: true,
    command_bin: 'ceph',
    command_timeout_seconds: 15
  }
}

function toClusterPayload(values: ClusterFormValues): CephClusterPayload {
  return {
    name: values.name,
    description: values.description,
    fsid: values.fsid,
    enabled: values.enabled ?? true,
    dashboard: {
      enabled: values.dashboard_enabled ?? false,
      base_url: values.dashboard_base_url,
      username: values.dashboard_username,
      password: values.dashboard_password,
      clear_secret: values.dashboard_clear_secret,
      insecure_tls: values.dashboard_insecure_tls
    },
    command: {
      enabled: values.command_enabled ?? false,
      bin: values.command_bin || 'ceph',
      cluster: values.command_cluster,
      conf: values.command_conf,
      name: values.command_name,
      keyring: values.command_keyring,
      keyring_content: values.command_keyring_content,
      clear_secret: values.command_clear_secret,
      timeout_seconds: values.command_timeout_seconds || 15
    }
  }
}

import { EditOutlined, PlusOutlined, ReloadOutlined, SaveOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Space, Table, Tag, Typography, message } from 'antd'
import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  createCluster,
  listClusters,
  updateCluster,
  type CephCluster
} from '../../api/clusters'
import { Page } from '../../components/Page'
import { DraggableModal } from '../../components/DraggableModal'

const { Text } = Typography

interface ClusterFormValues {
  name: string
  dashboard_username?: string
  dashboard_password?: string
  keyring?: string
}

export function ClusterManagementPage() {
  const navigate = useNavigate()
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
      dashboard_username: cluster.dashboard.username,
      dashboard_password: '',
      keyring: ''
    })
    setClusterModalOpen(true)
  }

  async function submitCluster(values: ClusterFormValues) {
    const result = editingCluster
      ? await updateCluster(editingCluster.id, values)
      : await createCluster(values)

    setClusterModalOpen(false)
    form.resetFields()
    message.success(result.message || (editingCluster ? '集群连接已更新' : '集群连接已创建'))
    await loadClusters()
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
                <Space>
                  <Button icon={<EditOutlined />} onClick={() => openEditCluster(cluster)}>
                    编辑
                  </Button>
                  <Button onClick={() => navigate(`/cluster/clusters/${cluster.id}`)}>
                    详情
                  </Button>
                </Space>
              )
            }
          ]}
        />
      </Card>

      <DraggableModal
        width={640}
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
          <div className="cluster-form-grid">
            <Form.Item name="name" label="集群名称" rules={[{ required: true, message: '请输入集群名称' }]}>
              <Input placeholder="例如：production-ceph" />
            </Form.Item>
            <Form.Item
              name="keyring"
              label="管理员密钥"
              rules={[{ required: !editingCluster?.command.keyring_content_set, message: '请输入 client.admin 密钥信息' }]}
            >
              <Input.Password placeholder={editingCluster?.command.keyring_content_set ? '留空则保持已保存密钥' : 'client.admin key'} />
            </Form.Item>
            <Form.Item name="dashboard_username" label="Dashboard 用户名" rules={[{ required: true, message: '请输入 Dashboard 用户名' }]}>
              <Input placeholder="admin" />
            </Form.Item>
            <Form.Item
              name="dashboard_password"
              label="Dashboard 密码"
              rules={[{ required: !editingCluster?.dashboard.password_set, message: '请输入 Dashboard 密码' }]}
            >
              <Input.Password placeholder={editingCluster?.dashboard.password_set ? '留空则保持已保存密码' : 'Dashboard 登录密码'} />
            </Form.Item>
          </div>
        </Form>
      </DraggableModal>
    </Page>
  )
}

function defaultClusterFormValues(): Partial<ClusterFormValues> {
  return {
    dashboard_username: 'admin'
  }
}

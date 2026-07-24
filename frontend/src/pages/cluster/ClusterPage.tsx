import { EyeInvisibleOutlined, EyeOutlined, PlusOutlined, ReloadOutlined, SaveOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Space, Table, Typography, message } from 'antd'
import { useCallback, useEffect, useState } from 'react'
import {
  createCluster,
  getClusterDashboardPassword,
  getClusterKeyring,
  listClusters,
  updateCluster,
  type CephCluster
} from '../../api/cluster'
import { Page } from '../../components/Page'
import { DraggableModal } from '../../components/DraggableModal'

const { Text } = Typography

interface ClusterFormValues {
  name: string
  monitor_host?: string
  dashboard_username?: string
  dashboard_password?: string
  keyring?: string
}

export function ClusterPage() {
  const [clusters, setClusters] = useState<CephCluster[]>([])
  const [clusterLoading, setClusterLoading] = useState(true)
  const [clusterError, setClusterError] = useState('')
  const [clusterModalOpen, setClusterModalOpen] = useState(false)
  const [editingCluster, setEditingCluster] = useState<CephCluster | null>(null)
  const [clusterSubmitting, setClusterSubmitting] = useState(false)
  const [visibleSecrets, setVisibleSecrets] = useState<Record<number, Partial<Record<keyof ClusterSecrets, boolean>>>>({})
  const [clusterSecrets, setClusterSecrets] = useState<Record<number, Partial<ClusterSecrets>>>({})
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

  async function submitCluster(values: ClusterFormValues) {
    if (clusterSubmitting) {
      return
    }
    setClusterSubmitting(true)
    try {
      const result = editingCluster
        ? await updateCluster(editingCluster.id, values)
        : await createCluster(values)

      setClusterModalOpen(false)
      form.resetFields()
      message.success(result.message || (editingCluster ? '集群连接已更新' : '集群连接已创建'))
      await loadClusters()
    } finally {
      setClusterSubmitting(false)
    }
  }

  async function toggleSecret(clusterID: number, secret: keyof ClusterSecrets) {
    const currentVisibility = Boolean(visibleSecrets[clusterID]?.[secret])
    if (!currentVisibility && !clusterSecrets[clusterID]?.[secret]) {
      const value = secret === 'keyring'
        ? await getClusterKeyring(clusterID)
        : await getClusterDashboardPassword(clusterID)
      setClusterSecrets((current) => ({
        ...current,
        [clusterID]: {
          ...current[clusterID],
          [secret]: value
        }
      }))
    }
    setVisibleSecrets((current) => ({
      ...current,
      [clusterID]: {
        ...current[clusterID],
        [secret]: !currentVisibility
      }
    }))
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
            <Button icon={<ReloadOutlined />} loading={clusterLoading} onClick={loadClusters}>
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
              title: '集群名称',
              key: 'cluster',
              render: (_, cluster) => (
                <div className="user-cell">
                  <Text strong>{cluster.name}</Text>
                </div>
              )
            },
            {
              title: 'MON 地址',
              dataIndex: ['command', 'monitor_host']
            },
            {
              title: '密钥',
              key: 'keyring',
              render: (_, cluster) => (
                <SecretValue
                  value={clusterSecrets[cluster.id]?.keyring}
                  visible={Boolean(visibleSecrets[cluster.id]?.keyring)}
                  configured={cluster.command.keyring_content_set}
                  onToggle={() => toggleSecret(cluster.id, 'keyring')}
                />
              )
            },
            {
              title: 'Dashboard 用户',
              dataIndex: ['dashboard', 'username']
            },
            {
              title: '密码',
              key: 'dashboard_password',
              render: (_, cluster) => (
                <SecretValue
                  value={clusterSecrets[cluster.id]?.dashboard_password}
                  visible={Boolean(visibleSecrets[cluster.id]?.dashboard_password)}
                  configured={cluster.dashboard.password_set}
                  onToggle={() => toggleSecret(cluster.id, 'dashboard_password')}
                />
              )
            },
          ]}
        />
      </Card>

      <DraggableModal
        width={640}
        className="cluster-modal"
        title={editingCluster ? `编辑集群：${editingCluster.name}` : '新建集群'}
        open={clusterModalOpen}
        onCancel={() => {
          if (!clusterSubmitting) {
            setClusterModalOpen(false)
          }
        }}
        onOk={() => form.submit()}
        okText="保存"
        confirmLoading={clusterSubmitting}
        okButtonProps={{ icon: <SaveOutlined />, loading: clusterSubmitting }}
        cancelButtonProps={{ disabled: clusterSubmitting }}
        cancelText="取消"
        destroyOnClose
        maskClosable={false}
      >
        <Form form={form} layout="vertical" initialValues={defaultClusterFormValues()} onFinish={submitCluster} className="cluster-form">
          <div className="cluster-form-grid">
            <Form.Item className="cluster-form-full" name="name" label="集群名称" rules={[{ required: true, message: '请输入集群名称' }]}>
              <Input placeholder="例如：production-ceph" />
            </Form.Item>
            <Form.Item
              className="cluster-form-full"
              name="monitor_host"
              label="MON 地址"
              rules={[{ required: true, message: '请输入 MON 地址' }]}
            >
              <Input placeholder="例如：10.0.0.11:6789,10.0.0.12:6789" />
            </Form.Item>
            <Form.Item
              className="cluster-form-full"
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

interface ClusterSecrets {
  keyring: string
  dashboard_password: string
}

function SecretValue({
  value,
  visible,
  configured,
  onToggle
}: {
  value?: string
  visible: boolean
  configured: boolean
  onToggle: () => void
}) {
  if (!configured) {
    return <Text type="secondary">未配置</Text>
  }

  return (
    <Space size={4}>
      <Text>{visible ? value || '—' : '••••••••'}</Text>
      <Button
        type="text"
        size="small"
        aria-label={visible ? '隐藏内容' : '查看内容'}
        icon={visible ? <EyeInvisibleOutlined /> : <EyeOutlined />}
        onClick={onToggle}
      />
    </Space>
  )
}

function defaultClusterFormValues(): Partial<ClusterFormValues> {
  return {
    dashboard_username: 'admin'
  }
}

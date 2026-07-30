import { PlusOutlined, ReloadOutlined, SaveOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Space, Table, Typography } from 'antd'
import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  createCluster,
  listClusters,
  updateCluster,
  type CephCluster
} from '../../api/cluster'
import { Page } from '../../components/Page'
import { DraggableModal } from '../../components/DraggableModal'
import { MonitorAddressSummary } from '../../components/MonitorAddressSummary'
import { TableAction, TableActions } from '../../components/TableActions'
import { useClusterContext } from '../../state/ClusterContext'
import { message } from '../../utils/appMessage'

const { Text } = Typography

interface ClusterFormValues {
  name: string
  monitor_host?: string
  client_username?: string
  keyring?: string
}

export function ClusterPage() {
  const navigate = useNavigate()
  const [clusters, setClusters] = useState<CephCluster[]>([])
  const [clusterLoading, setClusterLoading] = useState(true)
  const [clusterError, setClusterError] = useState('')
  const [clusterModalOpen, setClusterModalOpen] = useState(false)
  const [editingCluster, setEditingCluster] = useState<CephCluster | null>(null)
  const [clusterSubmitting, setClusterSubmitting] = useState(false)
  const [form] = Form.useForm<ClusterFormValues>()
  const { refreshClusters } = useClusterContext()

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
      monitor_host: cluster.command.monitor_host,
      client_username: cluster.client_username,
      keyring: ''
    })
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
      message.success(result.message)
      await loadClusters()
      await refreshClusters()
      setClusterModalOpen(false)
      form.resetFields()
    } finally {
      setClusterSubmitting(false)
    }
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
          tableLayout="fixed"
          dataSource={clusters}
          pagination={{ pageSize: 6, showSizeChanger: false }}
          scroll={{ x: 900 }}
          columns={[
            {
              title: '集群名称',
              key: 'cluster',
              width: 180,
              render: (_, cluster) => (
                <div className="user-cell">
                  <Text strong>{cluster.name}</Text>
                </div>
              )
            },
            {
              title: 'MON 地址',
              dataIndex: 'monitor_addresses',
              width: 320,
              render: (value) => <MonitorAddressSummary value={value} />
            },
            {
              title: 'Client 用户',
              dataIndex: 'client_username',
              width: 180
            },
            {
              title: '更新时间',
              dataIndex: 'updated_at',
              width: 200,
              render: (value) => formatDateTime(value)
            },
            {
              title: '操作',
              key: 'actions',
              width: 110,
              render: (_, cluster) => (
                <TableActions>
                  <TableAction onClick={() => openEditCluster(cluster)}>编辑</TableAction>
                  <TableAction onClick={() => navigate(`/cluster/cluster/${cluster.id}`)}>详情</TableAction>
                </TableActions>
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
            <Form.Item name="client_username" label="Client 用户" rules={[{ required: true, message: '请输入 Client 用户' }]}>
              <Input placeholder="client.admin" />
            </Form.Item>
          </div>
        </Form>
      </DraggableModal>
    </Page>
  )
}

function defaultClusterFormValues(): Partial<ClusterFormValues> {
  return {
    client_username: 'client.admin'
  }
}

function formatDateTime(value: unknown) {
  if (typeof value !== 'string' || !value) {
    return '-'
  }
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

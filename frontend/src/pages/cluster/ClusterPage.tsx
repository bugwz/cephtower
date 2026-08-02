import { InfoCircleOutlined, PlusOutlined, ReloadOutlined, SaveOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Popover, Space, Typography } from 'antd'
import type { TableProps } from 'antd/es/table'
import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  createCluster,
  listClusterFilterOptions,
  listClusters,
  updateCluster,
  type CephCluster
} from '../../api/cluster'
import { AppTable } from '../../components/AppTable'
import { Page } from '../../components/Page'
import { DraggableModal } from '../../components/DraggableModal'
import { MonitorAddressSummary } from '../../components/MonitorAddressSummary'
import { TableAction, TableActions } from '../../components/TableActions'
import { useClusterContext } from '../../state/ClusterContext'
import { message } from '../../utils/appMessage'
import { formatDateTime } from '../../utils/time'

const { Text } = Typography

interface ClusterFormValues {
  name: string
  monitor_host?: string
  client_username?: string
  keyring?: string
}

interface LoadClustersOptions {
  showLoading?: boolean
}

export function ClusterPage() {
  const navigate = useNavigate()
  const [clusters, setClusters] = useState<CephCluster[]>([])
  const [clusterLoading, setClusterLoading] = useState(true)
  const [clusterError, setClusterError] = useState('')
  const [columnFilters, setColumnFilters] = useState<Record<string, string[]>>({})
  const [filterOptions, setFilterOptions] = useState<Record<string, string[]>>({})
  const [clusterModalOpen, setClusterModalOpen] = useState(false)
  const [editingCluster, setEditingCluster] = useState<CephCluster | null>(null)
  const [clusterSubmitting, setClusterSubmitting] = useState(false)
  const [form] = Form.useForm<ClusterFormValues>()
  const { refreshClusters } = useClusterContext()

  const loadClusters = useCallback(async ({ showLoading = true }: LoadClustersOptions = {}) => {
    if (showLoading) {
      setClusterLoading(true)
    }
    setClusterError('')
    try {
      setClusters(await listClusters({ filters: columnFilters }))
    } catch (err) {
      setClusterError(err instanceof Error ? err.message : '加载集群连接失败')
    } finally {
      if (showLoading) {
        setClusterLoading(false)
      }
    }
  }, [columnFilters])

  useEffect(() => {
    loadClusters()
  }, [loadClusters])

  useEffect(() => {
    let ignore = false
    void listClusterFilterOptions(['name', 'client_username'])
      .then((options) => {
        if (!ignore) {
          setFilterOptions(options)
        }
      })
      .catch(() => {
        if (!ignore) {
          setFilterOptions({})
        }
      })
    return () => {
      ignore = true
    }
  }, [])

  const handleTableChange: TableProps<CephCluster>['onChange'] = (_pagination, filters) => {
    setColumnFilters(tableFilters(filters))
  }

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
      setClusterModalOpen(false)
      form.resetFields()
      message.success(result.message)
      void Promise.all([
        loadClusters({ showLoading: false }),
        refreshClusters()
      ])
    } finally {
      setClusterSubmitting(false)
    }
  }

  return (
    <Page
      title="集群列表"
      loading={clusterLoading}
      error={clusterError}
    >
      <Card
        className="page-surface-card"
        title="集群列表"
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} loading={clusterLoading} onClick={() => loadClusters()}>
              刷新
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreateCluster}>
              新建集群
            </Button>
          </Space>
        }
      >
        <AppTable
          size="middle"
          rowKey="id"
          tableLayout="fixed"
          dataSource={clusters}
          onChange={handleTableChange}
          pagination={{ defaultPageSize: 10, showSizeChanger: true }}
          scroll={{ x: 900 }}
          columns={[
            {
              title: '集群名称',
              key: 'name',
              width: 180,
              filterMultiple: true,
              filterSearch: true,
              filters: (filterOptions.name ?? []).map((value) => ({ text: value, value })),
              filteredValue: columnFilters.name ?? null,
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
              render: (value) => <MonitorAddressSummary value={value} maxVisible={2} />
            },
            {
              title: '认证用户',
              dataIndex: 'client_username',
              key: 'client_username',
              width: 180,
              filterMultiple: true,
              filterSearch: true,
              filters: (filterOptions.client_username ?? []).map((value) => ({ text: value, value })),
              filteredValue: columnFilters.client_username ?? null
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
                  <TableAction onClick={() => navigate(`/cluster/cluster/${encodeURIComponent(cluster.name)}`)}>详情</TableAction>
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
              label={<MonitorAddressLabel />}
              rules={[{ required: true, message: '请输入 MON 地址' }]}
            >
              <Input placeholder="例如：[v2:10.0.0.11:3300/0,v1:10.0.0.11:6789/0]" />
            </Form.Item>
            <Form.Item
              className="cluster-form-full"
              name="client_username"
              label="认证用户"
              rules={[{ required: true, message: '请输入认证用户' }]}
            >
              <Input placeholder="client.admin" />
            </Form.Item>
            <Form.Item
              className="cluster-form-full"
              name="keyring"
              label="认证密钥"
              rules={[{ required: !editingCluster?.command.keyring_content_set, message: '请输入认证密钥' }]}
            >
              <Input.Password placeholder={editingCluster?.command.keyring_content_set ? '留空则保持已保存密钥' : 'client.admin key'} />
            </Form.Item>
          </div>
        </Form>
      </DraggableModal>
    </Page>
  )
}

function MonitorAddressLabel() {
  return (
    <span className="mon-address-label">
      MON 地址
      <Popover
        placement="right"
        overlayClassName="mon-address-help-popover"
        content={
          <div className="mon-address-help">
            <div>支持以下格式，可用逗号或空格分隔多个 MON：</div>
            <code>10.0.0.11:6789,10.0.0.12:6789</code>
            <code>v2:10.0.0.11:3300/0,v2:10.0.0.12:3300/0</code>
            <code>[v2:10.0.0.11:3300/0,v1:10.0.0.11:6789/0]</code>
          </div>
        }
      >
        <InfoCircleOutlined className="mon-address-help-icon" />
      </Popover>
    </span>
  )
}

function defaultClusterFormValues(): Partial<ClusterFormValues> {
  return {
    client_username: 'client.admin'
  }
}

function tableFilters(filters: Record<string, unknown>) {
  return Object.fromEntries(
    Object.entries(filters)
      .map(([field, values]) => [field, Array.isArray(values) ? values.map(String).filter(Boolean) : []] as const)
      .filter(([, values]) => values.length > 0)
  )
}

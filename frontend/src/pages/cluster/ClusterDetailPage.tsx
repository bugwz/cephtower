import { ArrowLeftOutlined, DeleteOutlined, ExclamationCircleOutlined, ReloadOutlined } from '@ant-design/icons'
import { Button, Card, Descriptions, Modal, Space, Table, Tag, Typography, message } from 'antd'
import { useCallback } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { deleteCluster, getClusterDetail, type CephDiscoveredRecord } from '../../api/clusters'
import { textValue } from '../../api/client'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'

const { Paragraph, Text } = Typography

export function ClusterDetailPage() {
  const navigate = useNavigate()
  const { id = '' } = useParams()
  const loader = useCallback(() => getClusterDetail(id), [id])
  const { data, loading, error, refresh } = useResource(loader)
  const cluster = data?.cluster

  function confirmDeleteCluster() {
    if (!cluster) {
      return
    }
    Modal.confirm({
      title: `删除集群：${cluster.name}`,
      icon: <ExclamationCircleOutlined />,
      content: '删除后会同时清理该集群已保存的资源组件信息。',
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      async onOk() {
        const result = await deleteCluster(cluster.id)
        message.success(result.message || '集群连接已删除')
        navigate('/cluster/clusters')
      }
    })
  }

  return (
    <Page title="集群详情" loading={loading} error={error}>
      <Space direction="vertical" size={16} className="page-stack">
        <Card
          className="page-surface-card"
          title={cluster?.name ?? '集群详情'}
          extra={
            <Space>
              <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/cluster/clusters')}>
                返回
              </Button>
              <Button icon={<ReloadOutlined />} loading={loading} onClick={refresh}>
                刷新
              </Button>
              <Button danger icon={<DeleteOutlined />} disabled={!cluster} onClick={confirmDeleteCluster}>
                删除
              </Button>
            </Space>
          }
        >
          <Descriptions size="small" column={{ xs: 1, sm: 2, lg: 3 }} bordered>
            <Descriptions.Item label="集群 ID">{cluster?.id}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={cluster?.enabled ? 'success' : 'default'}>{cluster?.enabled ? '启用' : '禁用'}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="FSID">{cluster?.fsid || '未发现'}</Descriptions.Item>
            <Descriptions.Item label="Dashboard URL">{cluster?.dashboard.base_url || '未发现'}</Descriptions.Item>
            <Descriptions.Item label="Dashboard 用户">{cluster?.dashboard.username || '-'}</Descriptions.Item>
            <Descriptions.Item label="Dashboard 密码">
              <Tag color={cluster?.dashboard.password_set ? 'gold' : 'default'}>{cluster?.dashboard.password_set ? '已保存' : '未保存'}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="MON 地址">{cluster?.command.monitor_host || '-'}</Descriptions.Item>
            <Descriptions.Item label="Ceph 命令">{cluster?.command.bin || 'ceph'}</Descriptions.Item>
            <Descriptions.Item label="Ceph 用户">{cluster?.command.name || 'client.admin'}</Descriptions.Item>
            <Descriptions.Item label="Keyring">
              <Tag color={cluster?.command.keyring_content_set ? 'gold' : 'default'}>{cluster?.command.keyring_content_set ? '已保存' : '未保存'}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="更新时间">{cluster?.updated_at ? new Date(cluster.updated_at).toLocaleString() : '-'}</Descriptions.Item>
          </Descriptions>
        </Card>

        <Card className="page-surface-card" title="发现信息">
          <Space direction="vertical" size={16} className="page-stack">
            <Descriptions size="small" column={{ xs: 1, sm: 2, lg: 3 }} bordered>
              <Descriptions.Item label="FSID">{cluster?.fsid || '未发现'}</Descriptions.Item>
              <Descriptions.Item label="Dashboard URL">{cluster?.dashboard.base_url || '未发现'}</Descriptions.Item>
              <Descriptions.Item label="主机">{data?.discovery.hosts.length ?? 0}</Descriptions.Item>
              <Descriptions.Item label="MON">{data?.discovery.mons.length ?? 0}</Descriptions.Item>
              <Descriptions.Item label="MGR">{data?.discovery.mgrs.length ?? 0}</Descriptions.Item>
              <Descriptions.Item label="MDS">{data?.discovery.mdss.length ?? 0}</Descriptions.Item>
            </Descriptions>
            <Table
              size="middle"
              rowKey={(row) => `${row.category}:${row.key}`}
              dataSource={discoveryRows(data?.discovery)}
              pagination={{ pageSize: 8, showSizeChanger: false }}
              expandable={{
                expandedRowRender: (row) => (
                  <Paragraph className="snapshot-payload" copyable>
                    {formatSnapshotPayload(row.payload)}
                  </Paragraph>
                )
              }}
              columns={[
                { title: '类别', dataIndex: 'category' },
                { title: '名称', dataIndex: 'key' },
                { title: '类型', dataIndex: 'type', render: (value: string) => value || '-' },
                { title: '主机', dataIndex: 'hostname', render: (value: string) => value || '-' },
                { title: '状态', dataIndex: 'status', render: (value: string) => value || '-' },
                {
                  title: '发现时间',
                  dataIndex: 'discovered_at',
                  render: (value: string) => value ? new Date(value).toLocaleString() : '-'
                },
                {
                  title: '数据预览',
                  dataIndex: 'payload',
                  render: (value: unknown) => textValue(previewSnapshotPayload(value))
                }
              ]}
            />
          </Space>
        </Card>
      </Space>
    </Page>
  )
}

interface DiscoveryTableRow extends CephDiscoveredRecord {
  category: string
}

function discoveryRows(discovery: Awaited<ReturnType<typeof getClusterDetail>>['discovery'] | undefined): DiscoveryTableRow[] {
  if (!discovery) {
    return []
  }
  return [
    ...withCategory('hosts', discovery.hosts),
    ...withCategory('osds', discovery.osds),
    ...discovery.osd_flags.map((flag) => ({
      category: 'osd_flags',
      key: flag.name,
      payload: flag,
      discovered_at: flag.discovered_at
    })),
    ...withCategory('daemons', discovery.daemons),
    ...withCategory('services', discovery.services),
    ...withCategory('mons', discovery.mons),
    ...withCategory('mgrs', discovery.mgrs),
    ...withCategory('mdss', discovery.mdss),
    ...withCategory('mgr_modules', discovery.mgr_modules),
    ...withCategory('configuration', discovery.configuration)
  ]
}

function withCategory(category: string, records: CephDiscoveredRecord[]): DiscoveryTableRow[] {
  return records.map((record) => ({ ...record, category }))
}

function previewSnapshotPayload(payload: unknown) {
  if (Array.isArray(payload)) {
    return `${payload.length} 条记录`
  }
  if (payload && typeof payload === 'object') {
    return Object.keys(payload).slice(0, 6).join(', ')
  }
  return payload
}

function formatSnapshotPayload(payload: unknown) {
  try {
    return JSON.stringify(payload, null, 2)
  } catch {
    return textValue(payload)
  }
}

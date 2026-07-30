import { ArrowLeftOutlined, DeleteOutlined, ExclamationCircleOutlined, ReloadOutlined } from '@ant-design/icons'
import { Button, Card, Descriptions, Modal, Space, Table, Tag, Typography } from 'antd'
import { useCallback, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { deleteCluster, getClusterDetail, type CephDiscoveredRecord, type CephClusterDetail } from '../../api/cluster'
import { textValue } from '../../api/client'
import { refreshResource } from '../../api/resource'
import { MonitorAddressSummary } from '../../components/MonitorAddressSummary'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { message } from '../../utils/appMessage'

const { Paragraph, Text } = Typography

export function ClusterDetailPage() {
  const navigate = useNavigate()
  const { id = '' } = useParams()
  const loader = useCallback(() => getClusterDetail(id), [id])
  const { data, loading, error, refresh } = useResource(loader)
  const [refreshing, setRefreshing] = useState(false)
  const operationMutation = useMutationOperation()
  const cluster = data?.cluster

  async function refreshClusterDetail() {
    const clusterID = Number(cluster?.id ?? id)
    if (!Number.isFinite(clusterID) || clusterID <= 0) {
      await refresh()
      return
    }
    setRefreshing(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: clusterID, scope: 'all' }), '集群刷新执行成功')
      await refresh()
    } finally {
      setRefreshing(false)
    }
  }

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
        navigate('/cluster/cluster')
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
              <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/cluster/cluster')}>
                返回
              </Button>
              <Button icon={<ReloadOutlined />} loading={refreshing || loading} onClick={refreshClusterDetail}>
                刷新
              </Button>
              <Button danger icon={<DeleteOutlined />} disabled={!cluster} onClick={confirmDeleteCluster}>
                删除
              </Button>
            </Space>
          }
        >
          <Descriptions className="cluster-detail-descriptions cluster-basic-descriptions" size="small" column={{ xs: 1, sm: 2, lg: 3 }} bordered>
            <Descriptions.Item label="集群 ID"><Text className="detail-value-fixed">{cluster?.id ?? '-'}</Text></Descriptions.Item>
            <Descriptions.Item label="状态">{cluster ? <ClusterStatusTag status={cluster.status} enabled={cluster.enabled} /> : '-'}</Descriptions.Item>
            <Descriptions.Item label="Client 用户">{cluster?.command.name || 'client.admin'}</Descriptions.Item>
            <Descriptions.Item label="MON 地址" span={3}>
              <MonitorAddressSummary value={cluster?.command.monitor_host} />
            </Descriptions.Item>
            <Descriptions.Item label="Ceph 命令">{cluster?.command.bin || 'ceph'}</Descriptions.Item>
            <Descriptions.Item label="Keyring">
              <Tag color={cluster?.command.keyring_content_set ? 'gold' : 'default'}>{cluster?.command.keyring_content_set ? '已保存' : '未保存'}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="更新时间">{cluster?.updated_at ? new Date(cluster.updated_at).toLocaleString() : '-'}</Descriptions.Item>
          </Descriptions>
        </Card>

        <Card className="page-surface-card" title="详细信息">
          <Space direction="vertical" size={16} className="page-stack">
            <Descriptions className="cluster-detail-descriptions" size="small" column={{ xs: 1, sm: 2, lg: 3 }} bordered>
              <Descriptions.Item label="FSID">{cluster?.fsid || '未发现'}</Descriptions.Item>
              <Descriptions.Item label="Ceph 版本">{cluster?.ceph_version || '未发现'}</Descriptions.Item>
              <Descriptions.Item label="Dashboard URL">{dashboardURL(data) || '未配置'}</Descriptions.Item>
              <Descriptions.Item label="最后同步">{formatDateTime(cluster?.last_seen_at || cluster?.observed_at)}</Descriptions.Item>
              <Descriptions.Item label="观测代次">{cluster?.generation || '-'}</Descriptions.Item>
              <Descriptions.Item label="Endpoint">{data?.endpoints.length ?? 0}</Descriptions.Item>
              <Descriptions.Item label="主机">{data?.discovery.hosts.length ?? 0}</Descriptions.Item>
              <Descriptions.Item label="OSD">{data?.discovery.osds.length ?? 0}</Descriptions.Item>
              <Descriptions.Item label="MON">{data?.discovery.mons.length ?? 0}</Descriptions.Item>
              <Descriptions.Item label="MGR">{data?.discovery.mgrs.length ?? 0}</Descriptions.Item>
              <Descriptions.Item label="MDS">{data?.discovery.mdss.length ?? 0}</Descriptions.Item>
              <Descriptions.Item label="Daemon">{data?.discovery.daemons.length ?? 0}</Descriptions.Item>
              <Descriptions.Item label="Service">{data?.discovery.services.length ?? 0}</Descriptions.Item>
              <Descriptions.Item label="Credential">{data?.credentials.length ?? 0}</Descriptions.Item>
              {cluster?.last_error_message && (
                <Descriptions.Item label="最近错误" span={3}>
                  <Text type="danger">{cluster.last_error_message}</Text>
                </Descriptions.Item>
              )}
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
                { title: '类别', dataIndex: 'category', width: 140, render: (value: string) => categoryLabel(value) },
                { title: '名称', dataIndex: 'key' },
                { title: '类型', dataIndex: 'type', width: 120, render: (value: string) => value || '-' },
                { title: '主机', dataIndex: 'hostname', width: 160, render: (value: string) => value || '-' },
                { title: '状态', dataIndex: 'status', width: 120, render: (value: string) => value || '-' },
                {
                  title: '发现时间',
                  dataIndex: 'discovered_at',
                  width: 190,
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

function ClusterStatusTag({ status, enabled }: { status: string, enabled: boolean }) {
  if (!enabled) {
    return <Tag color="default">禁用</Tag>
  }
  if (status === 'available') {
    return <Tag color="success">可用</Tag>
  }
  if (status === 'unavailable') {
    return <Tag color="error">不可用</Tag>
  }
  return <Tag color="default">未知</Tag>
}

function dashboardURL(data: CephClusterDetail | null | undefined) {
  const clusterURL = data?.cluster.dashboard.base_url
  if (clusterURL) {
    return clusterURL
  }
  return data?.endpoints.find((endpoint) => endpoint.kind === 'grafana' && endpoint.enabled)?.url ?? ''
}

function categoryLabel(value: string) {
  return ({
    hosts: '主机',
    osds: 'OSD',
    osd_flags: 'OSD 标记',
    daemons: 'Daemon',
    services: 'Service',
    mons: 'MON',
    mgrs: 'MGR',
    mdss: 'MDS',
    mgr_modules: 'MGR 模块',
    configuration: '配置'
  } as Record<string, string>)[value] ?? value
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

function formatDateTime(value: unknown) {
  if (typeof value !== 'string' || !value) {
    return '-'
  }
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

function formatSnapshotPayload(payload: unknown) {
  try {
    return JSON.stringify(payload, null, 2)
  } catch {
    return textValue(payload)
  }
}

import { ArrowLeftOutlined, DeleteOutlined, ExclamationCircleOutlined, ReloadOutlined } from '@ant-design/icons'
import { Button, Card, Descriptions, Modal, Space, Tag, Typography } from 'antd'
import { useCallback, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { deleteCluster, getClusterDetail, type CephClusterDetail } from '../../api/cluster'
import type { ApiRecord } from '../../api/client'
import { refreshResource } from '../../api/resource'
import { MonitorAddressSummary } from '../../components/MonitorAddressSummary'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { message } from '../../utils/appMessage'
import { formatDateTime } from '../../utils/time'

const { Text } = Typography
const twoColumnDescriptions = { xs: 1, sm: 2, md: 2, lg: 2, xl: 2, xxl: 2 }

export function ClusterDetailPage() {
  const navigate = useNavigate()
  const { name = '' } = useParams()
  const clusterName = decodeClusterNameParam(name)
  const loader = useCallback(() => getClusterDetail(clusterName), [clusterName])
  const { data, loading, error, refresh } = useResource(loader)
  const [refreshing, setRefreshing] = useState(false)
  const operationMutation = useMutationOperation()
  const cluster = data?.cluster

  async function refreshClusterDetail() {
    const clusterID = cluster?.id
    if (typeof clusterID !== 'number' || !Number.isFinite(clusterID) || clusterID <= 0) {
      await refresh()
      return
    }
    setRefreshing(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: clusterID, scope: 'all' }), '刷新成功')
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
        window.setTimeout(() => {
          message.success(result.message || '集群连接已删除')
          navigate('/cluster/cluster')
        })
      }
    })
  }

  return (
    <Page title="集群详情" loading={loading} error={error}>
      <Space direction="vertical" size={16} className="page-stack">
        <Card
          className="page-surface-card"
          title="基础信息"
          extra={
            <Space className="cluster-detail-actions">
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
          <Descriptions className="cluster-detail-descriptions cluster-basic-descriptions" size="small" column={twoColumnDescriptions} bordered>
            <Descriptions.Item label="集群名称">{cluster?.name || '-'}</Descriptions.Item>
            <Descriptions.Item label="集群状态">{cluster ? <ClusterStatusTag status={cluster.status} enabled={cluster.enabled} /> : '-'}</Descriptions.Item>
            <Descriptions.Item label="认证用户">{cluster?.command.name || 'client.admin'}</Descriptions.Item>
            <Descriptions.Item label="认证密钥">
              <Tag color={cluster?.command.keyring_content_set ? 'gold' : 'default'}>{cluster?.command.keyring_content_set ? '已保存' : '未保存'}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="MON 地址" span={2}>
              <MonitorAddressSummary value={cluster?.command.monitor_host || cluster?.monitor_addresses} />
            </Descriptions.Item>
            <Descriptions.Item label="创建时间">{formatDateTime(cluster?.created_at)}</Descriptions.Item>
            <Descriptions.Item label="更新时间">{formatDateTime(cluster?.updated_at)}</Descriptions.Item>
          </Descriptions>
        </Card>

        <Card className="page-surface-card" title="详细信息">
          <Descriptions className="cluster-detail-descriptions" size="small" column={twoColumnDescriptions} bordered>
            <Descriptions.Item label="FSID">{clusterFSID(data)}</Descriptions.Item>
            <Descriptions.Item label="Ceph 版本">{cephVersion(data)}</Descriptions.Item>
            <Descriptions.Item label="最后同步">{formatDateTime(cluster?.last_seen_at || cluster?.observed_at)}</Descriptions.Item>
            <Descriptions.Item label="主机">{data?.discovery.hosts.length ?? 0}</Descriptions.Item>
            <Descriptions.Item label="MON">{data?.discovery.mons.length ?? 0}</Descriptions.Item>
            <Descriptions.Item label="MGR">{data?.discovery.mgrs.length ?? 0}</Descriptions.Item>
            <Descriptions.Item label="OSD">{data?.discovery.osds.length ?? 0}</Descriptions.Item>
            <Descriptions.Item label="MDS">{data?.discovery.mdss.length ?? 0}</Descriptions.Item>
            <Descriptions.Item label="Daemon">{data?.discovery.daemons.length ?? 0}</Descriptions.Item>
            <Descriptions.Item label="Service">{data?.discovery.services.length ?? 0}</Descriptions.Item>
            <Descriptions.Item label="Credential">{data?.credentials.length ?? 0}</Descriptions.Item>
            {cluster?.last_error_message && (
              <Descriptions.Item label="最近错误" span={2}>
                <Text type="danger">{cluster.last_error_message}</Text>
              </Descriptions.Item>
            )}
          </Descriptions>
        </Card>
      </Space>
    </Page>
  )
}

function decodeClusterNameParam(value: string) {
  try {
    return decodeURIComponent(value)
  } catch {
    return value
  }
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

function cephVersion(data: CephClusterDetail | null | undefined) {
  const versions = [
    normalizeCephVersion(data?.cluster.ceph_version),
    normalizeCephVersion(data?.overview?.ceph_version)
  ].filter(Boolean)
  for (const candidate of data?.discovery.daemons ?? []) {
    const record = asRecord(candidate.payload)
    if (!record || !isCephDaemonType(record.type ?? record.daemon_type)) {
      continue
    }
    const version = normalizeCephVersion(record.version ?? record.ceph_version)
    if (version) {
      versions.push(version)
    }
  }
  for (const candidate of data?.discovery.hosts ?? []) {
    const version = normalizeCephVersion(asRecord(candidate.payload)?.ceph_version)
    if (version) {
      versions.push(version)
    }
  }
  return versions.reduce(richerCephVersion, '') || '未发现'
}

function clusterFSID(data: CephClusterDetail | null | undefined) {
  const fsid = String(data?.cluster.fsid || data?.overview?.fsid || '').trim()
  return fsid || '未发现'
}

function asRecord(value: unknown): ApiRecord | undefined {
  return value && typeof value === 'object' && !Array.isArray(value) ? value as ApiRecord : undefined
}

function normalizeCephVersion(value: unknown) {
  if (typeof value !== 'string') {
    return ''
  }
  const match = value.trim().match(/\b(\d+\.\d+\.\d+)\b(?:\s+\(([0-9a-f]{7,40})\))?/i)
  if (!match) {
    return ''
  }
  return match[2] ? `${match[1]} (${match[2]})` : match[1]
}

function richerCephVersion(left: string, right: string) {
  if (!left) {
    return right
  }
  if (!right) {
    return left
  }
  const leftHasCommit = /\([0-9a-f]{7,40}\)/i.test(left)
  const rightHasCommit = /\([0-9a-f]{7,40}\)/i.test(right)
  if (leftHasCommit !== rightHasCommit) {
    return leftHasCommit ? left : right
  }
  return right.length > left.length ? right : left
}

function isCephDaemonType(value: unknown) {
  return typeof value === 'string' && ['mon', 'mgr', 'osd', 'mds', 'rgw', 'rbd-mirror', 'crash'].includes(value.trim().toLowerCase())
}

import { ArrowLeftOutlined, ReloadOutlined } from '@ant-design/icons'
import { Button, Card, Descriptions, Space, Tag, Typography } from 'antd'
import { useCallback, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { textValue, type ApiRecord } from '../../api/client'
import { listResource, refreshResource } from '../../api/resource'
import { Page } from '../../components/Page'
import { ResourceMetaBar } from '../../components/ResourceMetaBar'
import { useResource } from '../../hooks'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useClusterContext } from '../../state/ClusterContext'
import { formatDateTime } from '../../utils/time'

const { Text } = Typography
const detailColumns = { xs: 1, sm: 2, md: 2, lg: 2, xl: 2, xxl: 2 }

export function FilesystemDetailPage() {
  const navigate = useNavigate()
  const { name = '' } = useParams()
  const filesystemName = decodeRouteParam(name)
  const { selectedClusterId } = useClusterContext()
  const [refreshing, setRefreshing] = useState(false)
  const operationMutation = useMutationOperation()
  const loader = useCallback(async () => {
    if (!selectedClusterId || !filesystemName) {
      return null
    }
    const payload = await listResource('/filesystems', selectedClusterId, { name: filesystemName })
    return {
      filesystem: payload.items.find((row) => resourceName(row) === filesystemName) ?? null,
      observedAt: payload.observedAt,
      stale: payload.stale,
      staleReason: payload.staleReason
    }
  }, [filesystemName, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const filesystem = data?.filesystem

  async function refreshDetail() {
    if (!selectedClusterId || refreshing) {
      return
    }
    setRefreshing(true)
    try {
      await operationMutation.run(
        () => refreshResource({ clusterId: selectedClusterId, kinds: ['filesystem', 'mds'] }),
        '刷新成功'
      )
      await refresh()
    } finally {
      setRefreshing(false)
    }
  }

  return (
    <Page title="文件系统详情" loading={loading} error={error}>
      <Card
        className="page-surface-card"
        title={`文件系统详情 · ${filesystemName}`}
        extra={
          <Space>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/file/cephfs')}>返回</Button>
            <Button icon={<ReloadOutlined />} loading={refreshing || loading} onClick={refreshDetail}>刷新</Button>
          </Space>
        }
      >
        {filesystem ? (
          <Descriptions className="host-detail-descriptions" size="small" column={detailColumns} bordered>
            <Descriptions.Item label="名称">{resourceName(filesystem)}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={filesystem.stale === true ? 'warning' : 'success'}>
                {filesystem.stale === true ? '数据已过期' : textValue(filesystem.status, '可用')}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="文件系统 ID">{textValue(filesystem.id)}</Descriptions.Item>
            <Descriptions.Item label="最大 MDS 数">{textValue(filesystem.max_mds)}</Descriptions.Item>
            <Descriptions.Item label="元数据池 ID">{textValue(filesystem.metadata_pool)}</Descriptions.Item>
            <Descriptions.Item label="数据池 ID">{listValue(filesystem.data_pools)}</Descriptions.Item>
            <Descriptions.Item label="Rank">{listValue(filesystem.in)}</Descriptions.Item>
            <Descriptions.Item label="活跃 MDS">{mapValue(filesystem.up)}</Descriptions.Item>
            <Descriptions.Item label="数据源">{textValue(filesystem.source)}</Descriptions.Item>
            <Descriptions.Item label="资源版本">{textValue(filesystem.resource_version)}</Descriptions.Item>
            <Descriptions.Item label="采集时间">{formatDateTime(filesystem.observed_at)}</Descriptions.Item>
            <Descriptions.Item label="更新时间">{formatDateTime(filesystem.updated_at)}</Descriptions.Item>
          </Descriptions>
        ) : (
          <Text type="secondary">未找到文件系统 {filesystemName}</Text>
        )}
        <ResourceMetaBar observedAt={data?.observedAt} stale={data?.stale} staleReason={data?.staleReason} />
      </Card>
    </Page>
  )
}

function resourceName(row: ApiRecord) {
  return textValue(row.name ?? row.natural_key, '')
}

function listValue(value: unknown) {
  return Array.isArray(value) && value.length ? value.map(String).join(', ') : '—'
}

function mapValue(value: unknown) {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    return '—'
  }
  const entries = Object.entries(value as Record<string, unknown>)
  return entries.length ? entries.map(([rank, gid]) => `${rank}: ${textValue(gid)}`).join(', ') : '—'
}

function decodeRouteParam(value: string) {
  try {
    return decodeURIComponent(value)
  } catch {
    return value
  }
}

import { ArrowLeftOutlined, ReloadOutlined } from '@ant-design/icons'
import { Button, Card, Descriptions, Space, Tag, Typography } from 'antd'
import { useCallback, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { isRecord, numberValue, textValue, type ApiRecord } from '../../api/client'
import { getOptionalResource, refreshResource } from '../../api/resource'
import type { ResourceDTO } from '../../api/types'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useClusterContext } from '../../state/ClusterContext'
import { formatDateTime } from '../../utils/time'

const { Text } = Typography
const twoColumnDescriptions = { xs: 1, sm: 2, md: 2, lg: 2, xl: 2, xxl: 2 }

export function PoolDetailPage() {
  const navigate = useNavigate()
  const { name = '' } = useParams()
  const { selectedClusterId } = useClusterContext()
  const decodedName = name
  const [refreshing, setRefreshing] = useState(false)
  const operationMutation = useMutationOperation()
  const loader = useCallback(async () => {
    if (!selectedClusterId || !decodedName) {
      return null
    }
    const payload = await getOptionalResource('/pool', selectedClusterId, { pool: decodedName })
    return payload ? normalizePoolDetail(resourceToRecord(payload.item)) : null
  }, [decodedName, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)

  async function refreshPoolDetail() {
    if (!selectedClusterId || !decodedName || refreshing) {
      return
    }
    setRefreshing(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kind: 'pool' }), '刷新成功')
      await refresh()
    } finally {
      setRefreshing(false)
    }
  }

  return (
    <Page title="存储池详情" loading={loading} error={error}>
      <Space direction="vertical" size={16} className="page-stack">
        <Card
          className="page-surface-card"
          title="基础信息"
          extra={
            <Space>
              <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/cluster/pool')}>返回</Button>
              <Button icon={<ReloadOutlined />} loading={refreshing || loading} onClick={refreshPoolDetail}>刷新</Button>
            </Space>
          }
        >
          {data ? (
            <Descriptions size="small" column={twoColumnDescriptions} bordered>
              <Descriptions.Item label="名称">{textValue(data.name, decodedName)}</Descriptions.Item>
              <Descriptions.Item label="状态">{textValue(data.status)}</Descriptions.Item>
              <Descriptions.Item label="Pool 类型">{textValue(data.type)}</Descriptions.Item>
              <Descriptions.Item label="数据保护">{textValue(data.data_protection_display)}</Descriptions.Item>
              <Descriptions.Item label="PG 状态">{textValue(data.pg_status_display)}</Descriptions.Item>
              <Descriptions.Item label="PG 自动伸缩">{textValue(data.pg_autoscale_mode)}</Descriptions.Item>
              <Descriptions.Item label="应用标记" span={2}>{renderApplications(poolApplications(data))}</Descriptions.Item>
              <Descriptions.Item label="CRUSH 规则集">{textValue(data.crush_rule)}</Descriptions.Item>
              <Descriptions.Item label="压缩模式">{textValue(data.compression_mode, 'none')}</Descriptions.Item>
              <Descriptions.Item label="最大字节数">{formatQuota(data.quota_max_bytes ?? data.max_bytes)}</Descriptions.Item>
              <Descriptions.Item label="最大对象数">{textValue(data.quota_max_objects ?? data.max_objects ?? 0)}</Descriptions.Item>
              <Descriptions.Item label="资源版本">{textValue(data.resource_version)}</Descriptions.Item>
              <Descriptions.Item label="采集时间">{formatDateTime(data.observed_at)}</Descriptions.Item>
              <Descriptions.Item label="更新时间">{formatDateTime(data.updated_at)}</Descriptions.Item>
            </Descriptions>
          ) : (
            <Text type="secondary">暂无存储池详情</Text>
          )}
        </Card>
      </Space>
    </Page>
  )
}

function normalizePoolDetail(row: ApiRecord): ApiRecord {
  const type = poolType(row)
  const size = numberValue(row.size)
  const pgNum = numberValue(row.pg_num)
  const pgAutoscale = textValue(row.pg_autoscale_mode, 'on')
  return {
    ...row,
    type,
    data_protection_display: type === 'erasure' ? 'erasure' : `replica: x${size ?? 3}`,
    pg_status_display: pgNum ? `${pgNum} active+clean / ${pgAutoscale}` : `active+clean / ${pgAutoscale}`
  }
}

function resourceToRecord(item: ResourceDTO<ApiRecord>): ApiRecord {
  const data = isRecord(item.data) ? item.data : {}
  return {
    ...data,
    kind: item.kind,
    natural_key: item.natural_key,
    name: item.name ?? data.name,
    status: item.status ?? data.status,
    resource_version: item.resource_version,
    source: item.source,
    observed_at: item.observed_at,
    created_at: item.created_at,
    updated_at: item.updated_at,
    stale: item.stale
  }
}

function renderApplications(applications: string[]) {
  if (applications.length === 0) {
    return <Text type="secondary">-</Text>
  }
  return (
    <Space size={[4, 4]} wrap>
      {applications.map((application) => <Tag color="processing" key={application}>{application}</Tag>)}
    </Space>
  )
}

function poolApplications(row: ApiRecord) {
  if (Array.isArray(row.applications)) {
    return row.applications.map((item) => textValue(item, '')).filter(Boolean)
  }
  if (isRecord(row.application_metadata)) {
    return Object.keys(row.application_metadata)
  }
  return []
}

function poolType(row: ApiRecord) {
  return textValue(row.pool_type ?? row.type, 'replicated').toLowerCase() === 'erasure' ? 'erasure' : 'replicated'
}

function formatQuota(value: unknown) {
  const bytes = numberValue(value) ?? 0
  if (bytes <= 0) {
    return '0'
  }
  return textValue(bytes)
}

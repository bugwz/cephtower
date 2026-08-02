import { ArrowLeftOutlined, ReloadOutlined } from '@ant-design/icons'
import { Button, Card, Descriptions, Space, Tag, Typography } from 'antd'
import { useCallback, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { textValue } from '../../api/client'
import { listResource, refreshResource } from '../../api/resource'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useClusterContext } from '../../state/ClusterContext'
import { formatDateTime } from '../../utils/time'

const { Text } = Typography
const twoColumnDescriptions = { xs: 1, sm: 2, md: 2, lg: 2, xl: 2, xxl: 2 }

export function MonDetailPage() {
  const navigate = useNavigate()
  const { name = '' } = useParams()
  const { selectedClusterId } = useClusterContext()
  const monName = decodeRouteParam(name)
  const loader = useCallback(async () => {
    if (!selectedClusterId || !monName) {
      return null
    }
    const payload = await listResource('/monitors', selectedClusterId, { name: monName })
    return payload.items.find((row) => textValue(row.name ?? row.natural_key, '') === monName) ?? null
  }, [monName, selectedClusterId])
  const { data: mon, loading, error, refresh } = useResource(loader)
  const [refreshing, setRefreshing] = useState(false)
  const operationMutation = useMutationOperation()

  async function refreshMonDetail() {
    if (!selectedClusterId || refreshing) {
      return
    }
    setRefreshing(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kind: 'mon' }), '刷新成功')
      await refresh()
    } finally {
      setRefreshing(false)
    }
  }

  return (
    <Page title="MON详情" loading={loading} error={error}>
      <Card
        className="page-surface-card"
        title="基础信息"
        extra={
          <Space className="host-detail-actions">
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/cluster/mon')}>返回</Button>
            <Button icon={<ReloadOutlined />} loading={refreshing || loading} onClick={refreshMonDetail}>刷新</Button>
          </Space>
        }
      >
        {mon ? (
          <Descriptions className="host-detail-descriptions" size="small" column={twoColumnDescriptions} bordered>
            <Descriptions.Item label="名称">{textValue(mon.name ?? mon.natural_key)}</Descriptions.Item>
            <Descriptions.Item label="Rank">{textValue(mon.rank)}</Descriptions.Item>
            <Descriptions.Item label="Public Addr">{textValue(mon.address)}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={mon.in_quorum === true ? 'success' : 'default'}>{mon.in_quorum === true ? '仲裁中' : '未加入仲裁'}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="数据源">{textValue(mon.source)}</Descriptions.Item>
            <Descriptions.Item label="资源版本">{textValue(mon.resource_version)}</Descriptions.Item>
            <Descriptions.Item label="采集时间">{formatDateTime(mon.observed_at)}</Descriptions.Item>
            <Descriptions.Item label="更新时间">{formatDateTime(mon.updated_at)}</Descriptions.Item>
            <Descriptions.Item label="创建时间">{formatDateTime(mon.created_at)}</Descriptions.Item>
            <Descriptions.Item label="数据状态">
              <Tag color={mon.stale === true ? 'warning' : 'success'}>{mon.stale === true ? '已过期' : '最新'}</Tag>
            </Descriptions.Item>
          </Descriptions>
        ) : (
          <Text type="secondary">暂无 MON 详情</Text>
        )}
      </Card>
    </Page>
  )
}

function decodeRouteParam(value: string) {
  try {
    return decodeURIComponent(value)
  } catch {
    return value
  }
}

import { ArrowLeftOutlined, ReloadOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Descriptions, Input, Space, Typography } from 'antd'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { textValue } from '../../api/client'
import { listMonitorPerfCounters, listResource, refreshResource } from '../../api/resource'
import { DataTable } from '../../components/DataTable'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useClusterContext } from '../../state/ClusterContext'
import { formatDateTime } from '../../utils/time'
import { message } from '../../utils/appMessage'
import { MonPublicAddresses } from './MonPublicAddresses'
import { MonSessionHistory } from './MonSessionHistory'
import { monSessionCount } from './monSessionCount'
import { MonQuorumState } from './MonQuorumState'
import { MonSnapshotState } from './MonSnapshotState'
import { MonCounterValue, monCounterType } from './MonCounterValue'

const { Text } = Typography
const twoColumnDescriptions = { xs: 1, sm: 2, md: 2, lg: 2, xl: 2, xxl: 2 }

export function MonDetailPage() {
  const { name = '' } = useParams()
  const { selectedClusterId } = useClusterContext()
  const monName = decodeRouteParam(name)
  return <MonDetailContent key={JSON.stringify([selectedClusterId, monName])} selectedClusterId={selectedClusterId} monName={monName} />
}

function MonDetailContent({ selectedClusterId, monName }: { selectedClusterId?: number; monName: string }) {
  const active = useRef(true)
  useEffect(() => {
    active.current = true
    return () => { active.current = false }
  }, [])
  const navigate = useNavigate()
  const loader = useCallback(async () => {
    if (!selectedClusterId || !monName) {
      return null
    }
    const [payload, counters] = await Promise.all([
      listResource('/monitors', selectedClusterId, { name: monName }),
      listMonitorPerfCounters(monName, selectedClusterId).then(
        items => ({ items, error: '' }),
        (err: unknown) => ({ items: [], error: err instanceof Error && err.message ? err.message : '性能计数器请求失败' })
      )
    ])
    return {
      mon: payload.items.find((row) => textValue(row.name ?? row.natural_key, '') === monName) ?? null,
      counters: counters.items,
      counterError: counters.error
    }
  }, [monName, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const mon = data?.mon
  const [search, setSearch] = useState('')
  const counters = useMemo(() => {
    const needle = search.trim().toLowerCase()
    if (!needle) {
      return data?.counters ?? []
    }
    return (data?.counters ?? []).filter((counter) =>
      [counter.name, counter.description, counter.value, counter.unit].some((value) => textValue(value, '').toLowerCase().includes(needle))
    )
  }, [data?.counters, search])
  const [refreshing, setRefreshing] = useState(false)
  const operationMutation = useMutationOperation()

  async function refreshMonDetail() {
    if (!active.current || !selectedClusterId || !monName || refreshing) {
      return
    }
    setRefreshing(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kinds: ['mon', 'mon_status', 'mon_perf_counter'] }), false)
      if (!active.current) return
      message.success('刷新成功')
      await refresh()
    } finally {
      if (active.current) setRefreshing(false)
    }
  }

  return (
    <Page title="MON详情" loading={loading} error={error}>
      <Space direction="vertical" size={16} className="page-stack">
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
            <Descriptions.Item label="Priority（原生）">{textValue(mon.priority)}</Descriptions.Item>
            <Descriptions.Item label="Weight（原生）">{textValue(mon.weight)}</Descriptions.Item>
            <Descriptions.Item label="Public Addr">{textValue(mon.address)}</Descriptions.Item>
            <Descriptions.Item label="全部 Public 地址（协议 / 地址）"><MonPublicAddresses value={mon.public_addresses} /></Descriptions.Item>
            <Descriptions.Item label="状态">
              <MonQuorumState value={mon.in_quorum} />
            </Descriptions.Item>
            <Descriptions.Item label="Open sessions（采集时）">{monSessionCount(mon.open_sessions)}</Descriptions.Item>
            <Descriptions.Item label="数据源">{textValue(mon.source)}</Descriptions.Item>
            <Descriptions.Item label="资源版本">{textValue(mon.resource_version)}</Descriptions.Item>
            <Descriptions.Item label="采集时间">{formatDateTime(mon.observed_at)}</Descriptions.Item>
            <Descriptions.Item label="更新时间">{formatDateTime(mon.updated_at)}</Descriptions.Item>
            <Descriptions.Item label="创建时间">{formatDateTime(mon.created_at)}</Descriptions.Item>
            <Descriptions.Item label="数据状态">
              <MonSnapshotState observedAt={mon.observed_at} stale={mon.stale} />
            </Descriptions.Item>
          </Descriptions>
        ) : (
          <Text type="secondary">暂无 MON 详情</Text>
        )}
        </Card>
        <Card
        className="page-surface-card"
        title={`性能计数器 · mon.${monName}`}
        extra={<Input.Search allowClear placeholder="搜索名称、描述或值" value={search} onChange={(event) => setSearch(event.target.value)} style={{ width: 280 }} />}
      >
        {data?.counterError ? <Alert type="warning" showIcon message="性能计数器不可用" description={data.counterError} /> : <DataTable
          data={counters}
          rowKeyCandidates={['natural_key', 'name']}
          columns={[
            { key: 'name', title: '名称', filterKey: false },
            { key: 'description', title: '描述', filterKey: false },
            { key: 'value', title: '值 / 单位', filterKey: false, render: (value, row) => <MonCounterValue value={value} unit={row.unit} /> },
            { key: 'metric_type', title: '指标类型', filterKey: false, render: (value) => monCounterType(value) }
          ]}
        />}
        </Card>
        {selectedClusterId && monName && <MonSessionHistory key={JSON.stringify([selectedClusterId, monName])} clusterId={selectedClusterId} monName={monName} />}
      </Space>
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

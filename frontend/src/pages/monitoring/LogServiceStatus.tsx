import { Alert, Button, Space, Typography } from 'antd'
import { useCallback } from 'react'
import { type ApiRecord } from '../../api/client'
import { listAllResources } from '../../api/resource'
import { AppTable } from '../../components/AppTable'
import { ResourceMetaBar } from '../../components/ResourceMetaBar'
import { useResource } from '../../hooks'
import { useClusterContext } from '../../state/ClusterContext'

export function logServiceRows(items: ApiRecord[]): ApiRecord[] {
  return items.filter(row => row.type === 'loki' || row.type === 'promtail')
}

export function LogServiceStatus() {
  const { selectedClusterId } = useClusterContext()
  const loader = useCallback(async () => {
    if (!selectedClusterId) return undefined
    return listAllResources('/daemons', selectedClusterId, { filters: { type: ['loki', 'promtail'] } })
  }, [selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const rows = logServiceRows(data?.items ?? [])
  const reliable = Boolean(data && data.stale === false && !loading && !error)
  const display = (value: unknown) => typeof value === 'string' && value.trim() ? value : '未报告'
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Typography.Title level={5}>Ceph 编排日志服务库存</Typography.Title>
    <Button disabled={!selectedClusterId} loading={loading} onClick={() => void refresh()}>重新读取服务库存</Button>
    {error && <Alert type="error" message={`日志服务库存读取失败：${error}`} />}
    {data && !reliable && <Alert type="warning" message="以下是未确认有效的库存快照，不能据此判断日志服务当前是否运行。" />}
    {data && <ResourceMetaBar observedAt={data.observedAt} stale={data.stale} staleReason={data.staleReason} />}
    {reliable && ['loki', 'promtail'].filter(kind => !rows.some(row => row.type === kind)).map(kind => <Alert key={kind} type="warning" message={`本次库存未发现 ${kind} 守护进程；外部部署不包含在 Ceph 编排库存内。`} />)}
    <AppTable<ApiRecord> size="small" loading={loading} dataSource={rows} rowKey={row => String(row.natural_key ?? row.name)} pagination={{ pageSize: 10 }} columns={[
      { title: '类型', dataIndex: 'type' },
      { title: '守护进程', dataIndex: 'name', render: display },
      { title: '主机', dataIndex: 'hostname', render: display },
      { title: '状态（库存快照）', dataIndex: 'status', render: display },
      { title: 'Ceph 刷新时间', dataIndex: 'last_refresh', render: display }
    ]} />
    <Typography.Text type="secondary">此处读取 ceph orch ps 采集的全部匹配守护进程，不只检查首个实例；running 仅表示该进程的报告状态，不证明日志采集或查询链路正常。可在服务管理中部署或检查 Loki/Promtail。</Typography.Text>
  </Space>
}

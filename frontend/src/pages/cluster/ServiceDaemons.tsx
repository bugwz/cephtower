import { Alert, Button, Space, Typography } from 'antd'
import { useCallback } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'
import { DataTable } from '../../components/DataTable'
import { useResource } from '../../hooks'

export function ServiceDaemons({ clusterId, name }: { clusterId: number; name: string }) {
  const loader = useCallback(async () => {
    const result = await request<{ items: ApiRecord[]; service_name: string; observed_at: string }>('/service/daemons', jsonInit('GET', { cluster_id: clusterId, name }))
    if (result.service_name !== name || !Array.isArray(result.items)) throw new Error('守护进程响应与所选服务不匹配')
    return result
  }, [clusterId, name])
  const { data, error, loading, refresh } = useResource(loader)
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Typography.Text>来源：ceph orch ps --service-name；按服务读取，不根据守护进程名称猜测归属。</Typography.Text>
    <Button loading={loading} onClick={() => refresh()}>重新读取</Button>
    {error && <Alert type="error" message={error} description={data ? '下方为上次成功读取的数据，不能代表当前状态。' : '无法确认此服务的守护进程。'} />}
    {data && <Typography.Text type="secondary">读取时间：{data.observed_at}</Typography.Text>}
    {!loading && !error && data?.items.length === 0 && <Alert type="info" message="本次查询未返回此服务的守护进程" />}
    <DataTable data={data?.items ?? []} rowKeyCandidates={['daemon_name']} columns={[
      { key: 'daemon_name', title: '守护进程' }, { key: 'hostname', title: '主机' }, { key: 'status_desc', title: '状态' },
      { key: 'version', title: '版本' }, { key: 'cpu_percentage', title: 'CPU' }, { key: 'memory_usage', title: '内存用量（字节）' },
      { key: 'memory_limit', title: '内存限制（字节）' }, { key: 'container_image_name', title: '容器镜像' },
      { key: 'last_refresh', title: 'Ceph 最近刷新' }, { key: 'events', title: '事件' },
    ]} />
  </Space>
}

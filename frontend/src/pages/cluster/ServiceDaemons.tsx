import { Alert, Button, Descriptions, Space, Typography } from 'antd'
import { useCallback, useState } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'
import { DataTable } from '../../components/DataTable'
import { useResource } from '../../hooks'
import { DaemonPerf } from './DaemonPerf'

const daemonDetailFields = [
  ['daemon_type', '守护进程类型'], ['daemon_id', '守护进程 ID'],
  ['container_id', '容器 ID'], ['container_image_id', '镜像 ID'],
  ['container_image_digests', '镜像摘要'], ['memory_request', '内存请求（字节）'],
  ['ip', 'IP 地址'], ['ports', '监听端口'], ['systemd_unit', 'Systemd 单元'],
  ['is_active', '活跃实例'], ['osdspec_affinity', 'OSD 规格关联'],
  ['created', '创建时间'], ['started', '启动时间'],
  ['last_deployed', '最近部署'], ['last_configured', '最近配置'],
  ['pending_daemon_config', '待应用配置'], ['events', '事件'],
] as const

function daemonDetailText(value: unknown): string {
  if (value === undefined || value === null) return '未返回'
  if (typeof value === 'boolean') return value ? '是' : '否'
  if (typeof value === 'object') return JSON.stringify(value, null, 2)
  return String(value)
}

export function ServiceDaemons({ clusterId, name }: { clusterId: number; name: string }) {
  const [perf, setPerf] = useState<{ clusterId: number; service: string; daemon: string } | null>(null)
  const visiblePerf = perf?.clusterId === clusterId && perf.service === name ? perf : null
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
      { key: 'performance', title: '性能', filterKey: false, render: (_, row) => <Button
        disabled={typeof row.daemon_name !== 'string' || !/^(mon|mgr|mds|osd|rgw|rbd-mirror)\.[A-Za-z0-9_][A-Za-z0-9_.-]*$/.test(row.daemon_name)}
        onClick={() => setPerf({ clusterId, service: name, daemon: String(row.daemon_name) })}>性能计数器</Button> },
      { key: 'runtime_details', title: '运行详情', filterKey: false, ellipsis: false, render: (_, row) => <details>
        <summary>展开运行详情</summary>
        <Descriptions bordered size="small" column={1} style={{ minWidth: 360 }} items={daemonDetailFields.map(([key, label]) => ({
          key, label, children: <div style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{daemonDetailText(row[key])}</div>
        }))} />
      </details> },
    ]} />
    {visiblePerf && <Space direction="vertical" style={{ width: '100%' }}>
      <Space><Typography.Title level={5}>{visiblePerf.daemon} 性能计数器</Typography.Title><Button onClick={() => setPerf(null)}>关闭性能详情</Button></Space>
      <DaemonPerf key={`${visiblePerf.clusterId}:${visiblePerf.service}:${visiblePerf.daemon}`} clusterId={visiblePerf.clusterId} name={visiblePerf.daemon} />
    </Space>}
  </Space>
}

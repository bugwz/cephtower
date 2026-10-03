import { Alert, Button, Select, Space, Typography } from 'antd'
import { useCallback, useState } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'
import { DataTable } from '../../components/DataTable'
import { useResource } from '../../hooks'
import { hardwareHealthCounts, hardwareHealthGroup, type HardwareHealthGroup } from './hardwareHealth'
import { HostHardwareSummary } from './HostHardwareSummary'

export function HostHardware({ clusterId, host }: { clusterId: number; host: string }) {
  const [category, setCategory] = useState('memory')
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Alert type="info" message="原生硬件健康（node-proxy）" description="按所选类别读取 ceph orch hardware status，需要集群部署并配置 node-proxy。无记录或缺少健康字段不代表硬件健康；原生详情按文本保留数值精度。" />
    <Select aria-label="硬件类别" value={category} onChange={setCategory} style={{ width: 220 }} options={[
      { value: 'summary', label: host ? '当前主机六类硬件汇总' : '已报告主机六类硬件汇总' },
      { value: 'memory', label: '内存' }, { value: 'storage', label: '存储' }, { value: 'processors', label: '处理器' },
      { value: 'network', label: '网络' }, { value: 'power', label: '电源' }, { value: 'fans', label: '风扇' },
      ...(host ? [{ value: 'firmwares', label: '固件清单（只读）' }] : []),
    ]} />
    {category === 'summary' ? <HostHardwareSummary key={`${clusterId}:${host}`} clusterId={clusterId} host={host} />
      : <HardwareCategory key={`${clusterId}:${host}:${category}`} clusterId={clusterId} host={host} category={category} />}
  </Space>
}

function HardwareCategory({ clusterId, host, category }: { clusterId: number; host: string; category: string }) {
  const [healthFilter, setHealthFilter] = useState<HardwareHealthGroup | 'all'>('all')
  const loader = useCallback(async () => {
    const result = await request<{ host: string; category: string; items: ApiRecord[]; observed_at: string }>('/host/hardware', jsonInit('GET', { cluster_id: clusterId, host, category }))
    if (result.host !== host || result.category !== category || !Array.isArray(result.items)) throw new Error('硬件响应与所选主机或类别不匹配')
    return result
  }, [clusterId, host, category])
  const { data, loading, error, refresh } = useResource(loader)
  const counts = hardwareHealthCounts(data?.items ?? [])
  const rows = (data?.items ?? []).filter((row) => healthFilter === 'all' || hardwareHealthGroup(row.health) === healthFilter)
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Button loading={loading} onClick={() => refresh()}>重新读取</Button>
    {category === 'firmwares' && <Alert type="info" message="固件版本与发布日期为 node-proxy 原始报告；不判断是否最新，不提供升级操作。固件不计入六类硬件健康汇总。" />}
    {error && <Alert type="error" message={error} description={data ? '下方为上次成功读取的结果，不代表当前状态。' : '未能读取硬件信息，请检查 node-proxy、编排器支持和 Ceph 权限。'} />}
    {data && <Typography.Text type="secondary">读取时间：{data.observed_at}；返回组件数：{data.items.length}</Typography.Text>}
    {data && <Typography.Text>本类别已返回组件：OK {counts.ok}；其他已报告状态 {counts.other}；未报告健康状态 {counts.unknown}。非 OK 状态请结合原生详情判断；此统计不代表全主机或全群健康。</Typography.Text>}
    <Select aria-label="硬件健康状态筛选" style={{ width: 250 }} value={healthFilter} onChange={setHealthFilter} options={[
      { value: 'all', label: '全部已返回组件' }, { value: 'ok', label: 'OK' },
      { value: 'other', label: '其他已报告状态（非 OK）' }, { value: 'unknown', label: '未报告健康状态' },
    ]} />
    {!loading && !error && data?.items.length === 0 && <Alert type="warning" message="当前查询范围未返回该类别的硬件组件，不能据此判断健康" />}
    <DataTable data={rows} rowKeyCandidates={['id']} columns={[
      { key: 'host', title: '报告主机' },
      ...(category === 'firmwares' ? [
        { key: 'name', title: '固件名称' }, { key: 'version', title: '版本（原值）' }, { key: 'release_date', title: '发布日期（原值）' },
      ] : [{ key: 'system', title: '系统 / 机箱' }]),
      { key: 'component', title: '组件' },
      { key: 'health', title: '原生健康状态', render: (value) => hardwareHealthGroup(value) === 'unknown' ? '未知（未返回）' : String(value) },
      { key: 'state', title: '原生运行状态', render: (value) => value == null ? '未知（未返回）' : String(value) },
      { key: 'details', title: '原生详情', ellipsis: false, render: (value) => <details><summary>展开详情</summary><pre style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{String(value ?? '')}</pre></details> },
    ]} />
  </Space>
}

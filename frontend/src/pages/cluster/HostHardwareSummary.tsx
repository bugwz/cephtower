import { Alert, Button, Space } from 'antd'
import { useCallback } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'
import { DataTable } from '../../components/DataTable'
import { useResource } from '../../hooks'
import { hardwareHealthCounts } from './hardwareHealth'

async function loadHardwareSummary(clusterId: number, host: string): Promise<ApiRecord[]> {
  return Promise.all(['memory', 'storage', 'processors', 'network', 'power', 'fans'].map(async (category) => {
    try {
      const result = await request<{ host: string; category: string; items: ApiRecord[]; observed_at: string }>('/host/hardware', jsonInit('GET', { cluster_id: clusterId, host, category }, { suppressErrorNotification: true }))
      if (result.host !== host || result.category !== category || !Array.isArray(result.items)) throw new Error('响应与所选主机或类别不匹配')
      return { category, ...hardwareHealthCounts(result.items), reported_hosts: new Set(result.items.map((row) => row.host)).size, observed_at: result.observed_at, result: result.items.length ? '已读取' : '未返回组件，无法判断', error: null }
    } catch (err) {
      return { category, total: null, ok: null, other: null, unknown: null, reported_hosts: null, observed_at: null, result: '读取失败', error: err instanceof Error ? err.message : '读取失败' }
    }
  }))
}

export function HostHardwareSummary({ clusterId, host }: { clusterId: number; host: string }) {
  const loader = useCallback(() => loadHardwareSummary(clusterId, host), [clusterId, host])
  const { data, loading, error, refresh } = useResource(loader)
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Alert type="info" message={host ? '当前主机六类硬件汇总' : '集群已报告主机六类硬件汇总'} description="同时发起六个只读查询，各类别采样时间不同。失败类别不记作零；只统计 node-proxy 返回的组件，不保证覆盖所有主机，缺失组件也不能据此判定不存在。选择具体类别可查看组件详情。" />
    <Button loading={loading} onClick={() => refresh()}>重新读取全部类别</Button>
    {loading && data && <Alert type="info" message="正在刷新，下方保留上次读取结果" />}
    {error && <Alert type="error" message={error} />}
    {data?.some((row) => row.error) && <Alert type="warning" message="部分或全部类别读取失败，请查看各行错误；不能判断完整硬件健康" />}
    <DataTable data={data ?? []} rowKeyCandidates={['category']} columns={[
      { key: 'category', title: '原生类别' }, { key: 'result', title: '读取结果' },
      { key: 'reported_hosts', title: '返回组件的主机数' },
      { key: 'total', title: '返回组件数' }, { key: 'ok', title: 'OK' },
      { key: 'other', title: '其他已报告状态' }, { key: 'unknown', title: '未报告健康状态' },
      { key: 'observed_at', title: '读取时间' }, { key: 'error', title: '错误', ellipsis: false },
    ]} />
  </Space>
}

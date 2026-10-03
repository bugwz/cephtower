import { Alert, Button, Input, Space, Typography } from 'antd'
import { useCallback, useState } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'
import { DataTable } from '../../components/DataTable'
import { useResource } from '../../hooks'

export function DaemonPerf({ clusterId, name }: { clusterId: number; name: string }) {
  const [search, setSearch] = useState('')
  const loader = useCallback(async () => {
    const result = await request<{ daemon_name: string; items: ApiRecord[]; observed_at: string }>('/daemon/perf', jsonInit('GET', { cluster_id: clusterId, name }))
    if (result.daemon_name !== name || !Array.isArray(result.items)) throw new Error('性能响应与所选守护进程不匹配')
    return result
  }, [clusterId, name])
  const { data, loading, error, refresh } = useResource(loader)
  const rows = (data?.items ?? []).filter((row) => `${row.name} ${row.description}`.toLowerCase().includes(search.toLowerCase()))
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Alert type="info" message="原生性能计数器快照" description="使用 perf schema 与 perf dump 顺序读取，非原子采样。原始值可能为累计计数或包含 avgcount/sum 的平均值结构，不表示每秒速率；空值表示本次未返回。整数按文本展示以保留精度。" />
    <Space><Button loading={loading} onClick={() => refresh()}>重新读取</Button><Input.Search allowClear placeholder="搜索计数器或说明" value={search} onChange={(event) => setSearch(event.target.value)} /></Space>
    {error && <Alert type="error" message={error} description={data ? '下方为上次成功读取的快照。' : '无法读取此守护进程的性能计数器。'} />}
    {data && <Typography.Text type="secondary">读取时间：{data.observed_at}</Typography.Text>}
    {!loading && !error && data?.items.length === 0 && <Alert type="info" message="本次未返回性能计数器定义" />}
    <DataTable data={rows} rowKeyCandidates={['name']} columns={[
      { key: 'name', title: '计数器' }, { key: 'description', title: '说明' },
      { key: 'raw_value', title: '原始值', ellipsis: false, render: (value) => <span style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{value === null || value === undefined ? '未返回' : String(value)}</span> },
      { key: 'units', title: '原生单位' }, { key: 'value_type', title: '值类型' }, { key: 'type', title: '类型位标记' }, { key: 'priority', title: '优先级' },
    ]} />
  </Space>
}

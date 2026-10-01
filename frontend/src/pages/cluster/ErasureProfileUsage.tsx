import { Alert, Button, Space, Spin, Tag } from 'antd'
import { useEffect, useState } from 'react'
import { listAllResources, type ResourceListResult } from '../../api/resource'

export function erasureProfileUsage(profile: string, inventory: ResourceListResult) {
  const stale = inventory.stale !== false || inventory.items.some((row) => row.stale !== false)
  const names = inventory.items.filter((row) => row.erasure_code_profile === profile)
    .map((row) => row.name ?? row.pool_name ?? row.natural_key)
    .filter((name): name is string => typeof name === 'string' && name.length > 0)
  return { names: Array.from(new Set(names)), stale }
}

export function ErasureProfileUsage({ clusterId, profile }: { clusterId?: number, profile: string }) {
  const [result, setResult] = useState<ReturnType<typeof erasureProfileUsage> | null>(null)
  const [error, setError] = useState('')
  const [revision, setRevision] = useState(0)
  useEffect(() => {
    let active = true
    setResult(null); setError('')
    if (!clusterId || !profile) return
    void listAllResources('/pools', clusterId).then((inventory) => {
      if (active) setResult(erasureProfileUsage(profile, inventory))
    }).catch((err) => { if (active) setError(err instanceof Error ? err.message : '引用读取失败') })
    return () => { active = false }
  }, [clusterId, profile, revision])
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Button size="small" disabled={!clusterId} onClick={() => setRevision((value) => value + 1)}>重新读取存储池引用</Button>
    {error ? <Alert type="error" message={error} /> : !clusterId ? <Alert type="info" message="请先选择集群" /> : !result ? <Spin /> : <>
      {result.stale && <Alert type="warning" message="存储池库存过期或状态未知，以下仅为历史引用，不能据此确认配置未被使用。" />}
      {result.names.length ? <Space wrap>{result.names.map((name) => <Tag key={name}>{name}</Tag>)}</Space> : <Alert type="info" message={result.stale ? '暂无可确认的引用信息' : '当前采集库存中没有存储池使用此配置'} />}
    </>}
  </Space>
}

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

export function crushRuleUsage(id: number, inventory: ResourceListResult) {
  const validId = Number.isSafeInteger(id) && id >= 0
  const matches = inventory.items.filter((row) => validId && row.crush_rule === id)
  const names = matches.map((row) => row.name ?? row.pool_name ?? row.natural_key).filter((value): value is string => typeof value === 'string' && value.length > 0)
  return { names: Array.from(new Set(names)), stale: !validId || inventory.stale !== false || inventory.items.some((row) => row.stale !== false || typeof row.crush_rule !== 'number' || !Number.isSafeInteger(row.crush_rule) || row.crush_rule < 0) }
}

export function ErasureProfileUsage({ clusterId, profile }: { clusterId?: number, profile: string }) {
  if (!profile.trim()) return <Alert type="warning" message="纠删码配置名称缺失，无法确认存储池引用，请重新采集。" />
  return <PoolPlacementUsage key={JSON.stringify([clusterId, 'profile', profile])} clusterId={clusterId} profile={profile} />
}

export function validCrushRuleID(value: unknown): value is number {
  return typeof value === 'number' && Number.isSafeInteger(value) && value >= 0
}

export function CrushRuleUsage({ clusterId, name, id }: { clusterId?: number, name: string, id: unknown }) {
  if (!validCrushRuleID(id)) return <Alert type="warning" message="规则 ID 缺失或格式异常，无法确认存储池引用，请重新采集。" />
  return <PoolPlacementUsage key={JSON.stringify([clusterId, 'rule', id])} clusterId={clusterId} profile={name} ruleId={id} />
}

function PoolPlacementUsage({ clusterId, profile, ruleId }: { clusterId?: number, profile: string, ruleId?: number }) {
  const [result, setResult] = useState<ReturnType<typeof erasureProfileUsage> | null>(null)
  const [error, setError] = useState('')
  const [revision, setRevision] = useState(0)
  useEffect(() => {
    let active = true
    setResult(null); setError('')
    if (!clusterId || (ruleId === undefined && !profile)) return
    void listAllResources('/pools', clusterId).then((inventory) => {
      if (active) setResult(ruleId === undefined ? erasureProfileUsage(profile, inventory) : crushRuleUsage(ruleId, inventory))
    }).catch((err) => { if (active) setError(err instanceof Error ? err.message : '引用读取失败') })
    return () => { active = false }
  }, [clusterId, profile, ruleId, revision])
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Button size="small" disabled={!clusterId} onClick={() => setRevision((value) => value + 1)}>重新读取存储池引用</Button>
    {error ? <Alert type="error" message={error} /> : !clusterId ? <Alert type="info" message="请先选择集群" /> : !result ? <Spin /> : <>
      {result.stale && <Alert type="warning" message="存储池库存过期、状态未知或引用字段异常，以下仅为可识别的采集引用，不能据此确认配置未被使用。" />}
      {result.names.length ? <Space wrap>{result.names.map((name) => <Tag key={name}>{name}</Tag>)}</Space> : <Alert type="info" message={result.stale ? '暂无可确认的引用信息' : '当前采集库存中没有存储池使用此配置'} />}
    </>}
  </Space>
}

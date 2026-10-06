import { Alert, Button, Checkbox, Select, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { isRecord } from '../../api/client'
import { getResource, mutateResource, refreshResource } from '../../api/resource'

export const globalOSDFlags = [
  ['noin', '阻止自动标记 in'], ['noout', '阻止自动标记 out'],
  ['noup', '阻止标记 up'], ['nodown', '忽略故障报告，阻止标记 down'],
  ['pause', '暂停客户端读写'], ['noscrub', '禁用 scrub'],
  ['nodeep-scrub', '禁用 deep scrub'], ['nobackfill', '暂停 PG 回填'],
  ['norebalance', '仅在 PG 降级时回填'], ['norecover', '暂停 PG 恢复']
]

export function globalFlagSnapshot(item: unknown): { version: string; flags: string[] } | null {
  if (!isRecord(item) || item.stale !== false || !isRecord(item.data)) return null
  const raw = item.resource_version
  const version = typeof raw === 'string' ? raw : typeof raw === 'number' && Number.isSafeInteger(raw) ? String(raw) : ''
  if (!/^[1-9][0-9]*$/.test(version) || BigInt(version) > 18446744073709551615n) return null
  const flags = item.data.flags
  if (!Array.isArray(flags) || flags.some(flag => typeof flag !== 'string' || !flag || flag.trim() !== flag) || new Set(flags).size !== flags.length) return null
  return { version, flags }
}

export function OSDGlobalFlags({ clusterId }: { clusterId: number }) {
  const [snapshot, setSnapshot] = useState<ReturnType<typeof globalFlagSnapshot>>(null)
  const [flag, setFlag] = useState('noout')
  const [action, setAction] = useState('set')
  const [accepted, setAccepted] = useState(false)
  const [busy, setBusy] = useState(false)
  const [status, setStatus] = useState('')
  const running = useRef(false)
  const scope = useRef<object | null>({})
  useEffect(() => { scope.current = {}; return () => { scope.current = null } }, [])
  async function read() {
    if (running.current || !scope.current) return
    const current = scope.current
    running.current = true
    setBusy(true); setSnapshot(null); setAccepted(false); setStatus('')
    try {
      await refreshResource({ clusterId, kinds: ['osd_flag'] })
      if (scope.current !== current) return
      const result = await getResource('/osd/flag', clusterId)
      if (scope.current !== current) return
      const parsed = globalFlagSnapshot(result.item)
      if (!parsed) throw new Error('标志快照或版本无效，不能修改')
      setSnapshot(parsed)
    } catch (err) {
      if (scope.current === current) setStatus(err instanceof Error ? err.message : '读取失败')
    } finally {
      running.current = false
      if (scope.current === current) setBusy(false)
    }
  }
  async function save() {
    if (running.current || !scope.current || !snapshot || !accepted || !globalOSDFlags.some(([name]) => name === flag) || !['set', 'unset'].includes(action)) return
    const current = scope.current
    running.current = true
    setBusy(true); setSnapshot(null); setAccepted(false); setStatus('')
    try {
      await mutateResource('/osd/flag', 'PATCH', { cluster_id: clusterId, flag, action }, { ifMatch: snapshot.version })
      if (scope.current === current) setStatus('全局标志已回读确认。请重新采集后再操作；上方列表需刷新才能更新。')
    } catch (err) {
      if (scope.current === current) setStatus(`${err instanceof Error ? err.message : '修改未确认'}。请重新采集核对，不要直接重试。`)
    } finally {
      running.current = false
      if (scope.current === current) setBusy(false)
    }
  }
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Alert type="warning" message={`集群 ${clusterId} 全局 OSD 标志：影响整个集群。pause 会暂停客户端读写，恢复相关标志可能影响数据恢复和可用性。`} />
    <Button loading={busy} disabled={busy} onClick={() => void read()}>采集全局标志以编辑</Button>
    {snapshot && <div>已采集原生标志：{snapshot.flags.join(', ') || '无'}（只修改所选项，保留其他标志）</div>}
    <Select aria-label="全局 OSD 标志" value={flag} disabled={busy} options={globalOSDFlags.map(([value, label]) => ({ value, label: `${value} — ${label}` }))} onChange={value => { setFlag(value); setAccepted(false) }} />
    <Select aria-label="全局标志操作" value={action} disabled={busy} options={[{ value: 'set', label: '设置' }, { value: 'unset', label: '取消' }]} onChange={value => { setAction(value); setAccepted(false) }} />
    <Checkbox checked={accepted} disabled={busy || !snapshot} onChange={event => setAccepted(event.target.checked)}>确认目标集群及操作，接受全局影响</Checkbox>
    <Button danger disabled={busy || !snapshot || !accepted} onClick={() => void save()}>{action === 'set' ? '设置' : '取消'} {flag}</Button>
    {status && <Alert type="info" message={status} />}
  </Space>
}

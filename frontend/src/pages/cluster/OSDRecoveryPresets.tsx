import { Alert, Button, Card, Checkbox, Select, Space, Table } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'
import { listAllResources, mutateResource, refreshResource } from '../../api/resource'
import { configurationValueError } from './configurationValue'

export const recoveryPresets = {
  low: ['1', '1', '1', '0.5'],
  default: ['1', '3', '1', '0'],
  high: ['4', '4', '4', '0']
} as const
export const recoveryPresetNames = ['osd_max_backfills', 'osd_recovery_max_active', 'osd_recovery_max_single_start', 'osd_recovery_sleep'] as const
type Preset = keyof typeof recoveryPresets
type Change = { name: string; value: string; previous: string; version?: string }

export function recoveryPresetChange(name: string, value: string, help: ApiRecord, rows: ApiRecord[]): Change {
  if (help.name !== name || !['int', 'uint', 'float', 'size'].includes(String(help.type)) || !Array.isArray(help.flags) || help.flags.includes('no_mon_update')) throw new Error(`${name} 元数据缺失或不允许修改`)
  const error = configurationValueError(help, name, value)
  if (error) throw new Error(`${name}：${error}`)
  const matches = rows.filter(row => row.who === 'osd' && row.name === name)
  if (matches.length > 1) throw new Error(`${name} 的覆盖记录重复`)
  const row = matches[0]
  if (!row) return { name, value, previous: '无 osd 显式覆盖（不代表实时默认值）' }
  const version = row.resource_version
  const exact = typeof version === 'string' ? version : typeof version === 'number' && Number.isSafeInteger(version) ? String(version) : ''
  if (row.stale !== false || !/^[1-9]\d*$/.test(exact) || BigInt(exact) > 18446744073709551615n || typeof row.value !== 'string') throw new Error(`${name} 的覆盖值或版本无效，请重新采集`)
  return { name, value, previous: row.value, version: exact }
}

export function OSDRecoveryPresets({ clusterId, onChanged }: { clusterId: number; onChanged: () => void }) {
  const [preset, setPreset] = useState<Preset>('low')
  const [plan, setPlan] = useState<Change[] | null>(null)
  const [accepted, setAccepted] = useState(false)
  const [busy, setBusy] = useState(false)
  const [status, setStatus] = useState('')
  const running = useRef(false)
  const scope = useRef<object | null>({})
  useEffect(() => { scope.current = {}; return () => { scope.current = null } }, [])
  async function preview() {
    if (running.current || !scope.current) return
    const token = scope.current
    running.current = true; setBusy(true); setPlan(null); setAccepted(false); setStatus('')
    try {
      await refreshResource({ clusterId, kinds: ['config_value', 'config_option'] })
      if (scope.current !== token) return
      const values = await listAllResources('/configuration/values', clusterId)
      if (values.stale) throw new Error('配置库存过期，不能预览写入')
      const changes: Change[] = []
      for (const [i, name] of recoveryPresetNames.entries()) {
        if (scope.current !== token) return
        const help = await request<ApiRecord>('/configuration/option', jsonInit('GET', { cluster_id: clusterId, name }))
        changes.push(recoveryPresetChange(name, recoveryPresets[preset][i], help, values.items))
      }
      if (scope.current === token) setPlan(changes)
    } catch (err) { if (scope.current === token) setStatus(err instanceof Error ? err.message : '预览失败') }
    finally { running.current = false; if (scope.current === token) setBusy(false) }
  }
  async function apply() {
    if (running.current || !scope.current || !plan || !accepted) return
    const token = scope.current, changes = plan
    running.current = true; setBusy(true); setPlan(null); setAccepted(false)
    const completed: string[] = []
    try {
      for (const change of changes) {
        if (scope.current !== token) return
        setStatus(`正在写入 ${change.name}；已确认 ${completed.length}/4 项`)
        await mutateResource('/configuration/value', 'PUT', { cluster_id: clusterId, who: 'osd', name: change.name, value: change.value }, change.version ? { ifMatch: change.version } : undefined)
        completed.push(change.name)
      }
      if (scope.current !== token) return
      setStatus('4 项配置写入已确认；不代表所有 OSD 已实时生效。请重新采集核对。')
      await refreshResource({ clusterId, kinds: ['config_value', 'config_option'] })
      if (scope.current === token) onChanged()
    } catch (err) {
      if (scope.current === token) setStatus(`已确认 ${completed.length}/4 项：${completed.join(', ') || '无'}。后续已停止；失败项结果可能不确定，请重新采集核对，不要直接重试。${err instanceof Error ? err.message : ''}`)
    } finally { running.current = false; if (scope.current === token) setBusy(false) }
  }
  return <Card title={`参考恢复优先级预设 · 集群 ${clusterId} · osd 作用域`}>
    <Space direction="vertical" style={{ width: '100%' }}>
      <Alert type="warning" message="四项独立写入，不是原子事务；失败停止，不自动回滚或重试。" description="Default 是参考界面的固定预设，不是当前 Ceph 版本默认值。mClock 可能覆盖这些参数，本操作不改变调度器或 override 设置。" />
      <Select aria-label="恢复优先级预设" value={preset} disabled={busy} onChange={value => { setPreset(value); setPlan(null); setAccepted(false); setStatus('') }} options={[{ value: 'low', label: 'Low（低）' }, { value: 'default', label: 'Default（参考预设）' }, { value: 'high', label: 'High（高）' }]} />
      <Button disabled={busy} loading={busy} onClick={() => void preview()}>重新采集并预览</Button>
      {plan && <><Table rowKey="name" pagination={false} dataSource={plan} columns={[{ title: '参数', dataIndex: 'name' }, { title: '当前 osd 覆盖', dataIndex: 'previous' }, { title: '目标值', dataIndex: 'value' }]} />
        <Checkbox checked={accepted} disabled={busy} onChange={event => setAccepted(event.target.checked)}>确认覆盖上述 osd 配置，了解部分成功和 mClock 生效限制</Checkbox>
        <Button danger disabled={busy || !accepted} onClick={() => void apply()}>应用四项预设</Button></>}
      {status && <Alert type="info" message={status} />}
    </Space>
  </Card>
}

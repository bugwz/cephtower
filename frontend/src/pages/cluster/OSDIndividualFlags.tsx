import { Alert, Button, Checkbox, Descriptions, Select, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { type ApiRecord } from '../../api/client'
import { mutateResource } from '../../api/resource'

export const individualOSDFlags = ['noout', 'noin', 'nodown', 'noup']

export function individualFlagSnapshot(record: ApiRecord): { version: string; flags: string[] } | null {
  const raw = record.resource_version
  const version = typeof raw === 'string' ? raw : typeof raw === 'number' && Number.isSafeInteger(raw) ? String(raw) : ''
  if (record.stale !== false || !/^[1-9][0-9]*$/.test(version) || BigInt(version) > 18446744073709551615n) return null
  const state = record.state
  if (!Array.isArray(state) || state.some(value => typeof value !== 'string' || !value || value.trim() !== value) || new Set(state).size !== state.length) return null
  return { version, flags: state.filter(value => individualOSDFlags.includes(value)) }
}

export function OSDIndividualFlags({ clusterId, osdId, record }: { clusterId: number; osdId: string; record: ApiRecord }) {
  const [flag, setFlag] = useState('noout')
  const [action, setAction] = useState('set')
  const [accepted, setAccepted] = useState(false)
  const [busy, setBusy] = useState(false)
  const [attempted, setAttempted] = useState(false)
  const [status, setStatus] = useState('')
  const running = useRef(false)
  const scope = useRef<object | null>({})
  useEffect(() => { scope.current = {}; return () => { scope.current = null } }, [])
  const snapshot = individualFlagSnapshot(record)
  async function save() {
    if (!scope.current || running.current || attempted || !accepted || !snapshot || !individualOSDFlags.includes(flag) || !['set', 'unset'].includes(action)) return
    const current = scope.current
    running.current = true
    setBusy(true); setAttempted(true); setStatus('')
    try {
      await mutateResource('/osd/flag/individual', 'PATCH', { cluster_id: clusterId, osd_id: osdId, flag, action }, { ifMatch: snapshot.version })
      if (scope.current === current) setStatus('单 OSD 标志已回读确认。请关闭详情并刷新 OSD 列表，重新打开后再进行其他修改。')
    } catch (err) {
      if (scope.current === current) setStatus(`${err instanceof Error ? err.message : '修改未确认'}。请刷新并核对原生状态后再操作，不要直接重试。`)
    } finally {
      running.current = false
      if (scope.current === current) setBusy(false)
    }
  }
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Alert type="warning" message={`集群 ${clusterId} / OSD ${osdId}：这些标志影响 OSD 状态转换及故障恢复行为，请确认维护计划。`} />
    <Alert type="info" message="这里只展示和修改此 OSD 的直接标志。取消直接标志不会取消全局、CRUSH 节点或设备类别上的同名限制，不能据此推断最终生效状态。" />
    <Descriptions bordered column={2}>{individualOSDFlags.map(name => <Descriptions.Item key={name} label={name}>{!snapshot ? '未知或快照过期' : snapshot.flags.includes(name) ? '已直接设置' : '未直接设置'}</Descriptions.Item>)}</Descriptions>
    {!snapshot && <Alert type="warning" message="状态或版本无效，禁止修改。请刷新库存后重新打开详情。" />}
    <Select aria-label="单 OSD 标志" value={flag} disabled={busy || attempted} options={individualOSDFlags.map(value => ({ value, label: value }))} onChange={value => { setFlag(value); setAccepted(false) }} />
    <Select aria-label="标志操作" value={action} disabled={busy || attempted} options={[{ value: 'set', label: '设置' }, { value: 'unset', label: '取消' }]} onChange={value => { setAction(value); setAccepted(false) }} />
    <Checkbox checked={accepted} disabled={busy || attempted} onChange={event => setAccepted(event.target.checked)}>确认目标及操作，理解对故障恢复的影响</Checkbox>
    <Button danger loading={busy} disabled={!snapshot || !accepted || busy || attempted} onClick={() => void save()}>{action === 'set' ? '设置' : '取消'} {flag}</Button>
    {status && <Alert type="info" message={status} />}
  </Space>
}

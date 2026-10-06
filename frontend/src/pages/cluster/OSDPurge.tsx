import { Alert, Button, Checkbox, Input, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { type ApiRecord } from '../../api/client'
import { mutateResource } from '../../api/resource'

export function osdPurgeTarget(record: ApiRecord, osdId: string): { version: string; uuid: string } | null {
  const raw = record.resource_version
  const version = typeof raw === 'string' ? raw : typeof raw === 'number' && Number.isSafeInteger(raw) ? String(raw) : ''
  if (!/^(0|[1-9][0-9]*)$/.test(osdId) || BigInt(osdId) > 2147483647n || record.stale !== false || record.up !== false || !/^[1-9][0-9]*$/.test(version) || BigInt(version) > 18446744073709551615n) return null
  if (typeof record.uuid !== 'string' || !/^[0-9a-f]{8}(-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i.test(record.uuid)) return null
  if (!Array.isArray(record.state) || record.state.length === 0 || record.state.some(state => typeof state !== 'string' || !state || state.trim() !== state) || new Set(record.state).size !== record.state.length) return null
  if (record.state.includes('destroyed') !== (record.uuid === '00000000-0000-0000-0000-000000000000')) return null
  return { version, uuid: record.uuid }
}

export function OSDPurge({ clusterId, osdId, record }: { clusterId: number; osdId: string; record: ApiRecord }) {
  const [confirmation, setConfirmation] = useState('')
  const [accepted, setAccepted] = useState(false)
  const [busy, setBusy] = useState(false)
  const [attempted, setAttempted] = useState(false)
  const [status, setStatus] = useState('')
  const running = useRef(false)
  const scope = useRef<object | null>({})
  useEffect(() => { scope.current = {}; return () => { scope.current = null } }, [])
  const target = osdPurgeTarget(record, osdId)
  const expected = `purge osd.${osdId}`
  async function purge() {
    if (!scope.current || running.current || attempted || !accepted || confirmation !== expected || !target) return
    const current = scope.current
    running.current = true
    setBusy(true); setAttempted(true); setStatus('')
    try {
      await mutateResource('/osd/purge', 'POST', { cluster_id: clusterId, osd_id: osdId, expected_uuid: target.uuid, confirmation }, { ifMatch: target.version })
      if (scope.current === current) setStatus('已回读确认 OSD map 移除及 CRUSH 名称、bucket 引用清理。请关闭详情并刷新 OSD 库存；此操作不清空磁盘或移除编排器守护进程。')
    } catch (err) {
      if (scope.current === current) setStatus(`${err instanceof Error ? err.message : 'Purge 操作未确认'}。OSD ID、密钥及 CRUSH 条目可能已被移除；请关闭详情并刷新、核对原生状态，不要直接重试。`)
    } finally {
      if (scope.current === current) setBusy(false)
      running.current = false
    }
  }
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Alert type="error" message={`集群 ${clusterId} / OSD ${osdId}：Purge 永久移除 OSD ID、认证密钥和 CRUSH 条目，数据可能永久不可读。与 Destroy 不同，不保留 ID 用于替换。不是停止进程、磁盘 Zap 或编排器移除。`} />
    <Alert type="info" message="支持 Down 的正常或已 Destroy OSD。后端重新核验身份及原生 safe-to-destroy，执行后分别核对 OSD map 和 CRUSH 清理结果。" />
    {!target && <Alert type="warning" message="需要新鲜、版本有效、UUID 与销毁状态一致的 Down OSD 快照。请刷新库存后重新打开详情。" />}
    <div>目标 UUID：{typeof record.uuid === 'string' ? record.uuid : '未知'}</div>
    <Input aria-label="OSD Purge 确认文本" placeholder={expected} value={confirmation} disabled={busy || attempted} onChange={event => { setConfirmation(event.target.value); setAccepted(false) }} />
    <Checkbox checked={accepted} disabled={busy || attempted} onChange={event => setAccepted(event.target.checked)}>我已核对集群、OSD 和 UUID，接受永久移除和数据不可读风险</Checkbox>
    <Button danger loading={busy} disabled={!target || confirmation !== expected || !accepted || busy || attempted} onClick={() => void purge()}>永久清除 OSD（Purge）</Button>
    {status && <Alert type="info" message={status} />}
  </Space>
}

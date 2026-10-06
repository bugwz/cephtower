import { Alert, Button, Checkbox, Input, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { type ApiRecord } from '../../api/client'
import { mutateResource } from '../../api/resource'

export function osdDestroyTarget(record: ApiRecord, osdId: string): { version: string; uuid: string } | null {
  const raw = record.resource_version
  const version = typeof raw === 'string' ? raw : typeof raw === 'number' && Number.isSafeInteger(raw) ? String(raw) : ''
  if (!/^(0|[1-9][0-9]*)$/.test(osdId) || BigInt(osdId) > 2147483647n || record.stale !== false || record.up !== false || !/^[1-9][0-9]*$/.test(version) || BigInt(version) > 18446744073709551615n) return null
  if (typeof record.uuid !== 'string' || !/^[0-9a-f]{8}(-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i.test(record.uuid) || record.uuid === '00000000-0000-0000-0000-000000000000') return null
  if (!Array.isArray(record.state) || record.state.length === 0 || record.state.some(state => typeof state !== 'string' || !state || state.trim() !== state) || new Set(record.state).size !== record.state.length || record.state.includes('destroyed')) return null
  return { version, uuid: record.uuid }
}

export function OSDDestroy({ clusterId, osdId, record }: { clusterId: number; osdId: string; record: ApiRecord }) {
  const [confirmation, setConfirmation] = useState('')
  const [accepted, setAccepted] = useState(false)
  const [busy, setBusy] = useState(false)
  const [attempted, setAttempted] = useState(false)
  const [status, setStatus] = useState('')
  const running = useRef(false)
  const scope = useRef<object | null>({})
  useEffect(() => { scope.current = {}; return () => { scope.current = null } }, [])
  const target = osdDestroyTarget(record, osdId)
  const expected = `destroy osd.${osdId}`
  async function destroy() {
    if (!scope.current || running.current || attempted || !accepted || confirmation !== expected || !target) return
    const current = scope.current
    running.current = true
    setBusy(true); setAttempted(true); setStatus('')
    try {
      await mutateResource('/osd/destroy', 'POST', { cluster_id: clusterId, osd_id: osdId, expected_uuid: target.uuid, confirmation }, { ifMatch: target.version })
      if (scope.current === current) setStatus('已回读确认 OSD destroyed 状态及清空后的 UUID。请关闭详情并刷新 OSD 库存。此操作不部署替换 OSD。')
    } catch (err) {
      if (scope.current === current) setStatus(`${err instanceof Error ? err.message : '销毁未确认'}。密钥可能已被移除；请关闭详情并刷新、核对原生状态，不要直接重试。`)
    } finally {
      if (scope.current === current) setBusy(false)
      running.current = false
    }
  }
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Alert type="error" message={`集群 ${clusterId} / OSD ${osdId}：Destroy 保留 OSD ID，但移除 CephX、配置及 lockbox 密钥，数据可能永久不可读。此操作不可逆，不是停止进程，也不是编排器移除或自动替换。`} />
    <Alert type="info" message="后端将重新核对 UUID、Down 状态及原生 safe-to-destroy 结果，仅明确安全时执行。安全检查不能撤销销毁后果。" />
    {!target && <Alert type="warning" message="需要新鲜、版本有效、UUID 已知且未销毁的 Down OSD 快照。请刷新库存后重新打开详情。" />}
    <div>目标 UUID：{typeof record.uuid === 'string' ? record.uuid : '未知'}</div>
    <Input aria-label="OSD 销毁确认文本" placeholder={expected} value={confirmation} disabled={busy || attempted} onChange={event => { setConfirmation(event.target.value); setAccepted(false) }} />
    <Checkbox checked={accepted} disabled={busy || attempted} onChange={event => setAccepted(event.target.checked)}>我已核对集群、OSD 和 UUID，接受密钥移除及数据永久不可读风险</Checkbox>
    <Button danger loading={busy} disabled={!target || confirmation !== expected || !accepted || busy || attempted} onClick={() => void destroy()}>永久销毁 OSD（保留 ID）</Button>
    {status && <Alert type="info" message={status} />}
  </Space>
}

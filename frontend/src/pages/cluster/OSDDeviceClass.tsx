import { Alert, Button, Checkbox, Input, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { type ApiRecord } from '../../api/client'
import { mutateResource } from '../../api/resource'

export function osdClassVersion(record: ApiRecord): string | null {
  const raw = record.resource_version
  const version = typeof raw === 'string' ? raw : typeof raw === 'number' && Number.isSafeInteger(raw) ? String(raw) : ''
  if (record.stale !== false || typeof record.device_class !== 'string' || !/^[1-9][0-9]*$/.test(version) || BigInt(version) > 18446744073709551615n) return null
  return version
}

export function OSDDeviceClass({ clusterId, osdId, record }: { clusterId: number; osdId: string; record: ApiRecord }) {
  const [value, setValue] = useState(typeof record.device_class === 'string' ? record.device_class : '')
  const [accepted, setAccepted] = useState(false)
  const [busy, setBusy] = useState(false)
  const [attempted, setAttempted] = useState(false)
  const [status, setStatus] = useState('')
  const running = useRef(false)
  const scope = useRef<object | null>({})
  useEffect(() => { scope.current = {}; return () => { scope.current = null } }, [])
  const version = osdClassVersion(record)
  const valid = /^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$/.test(value)
  async function save() {
    if (!scope.current || running.current || attempted || !accepted || !valid || !version) return
    const current = scope.current
    running.current = true
    setBusy(true); setAttempted(true); setStatus('')
    try {
      await mutateResource('/osd/device/class', 'PUT', { cluster_id: clusterId, osd_id: osdId, device_class: value, expected_class: record.device_class }, { ifMatch: version })
      if (scope.current === current) setStatus('设备类别已回读确认。请关闭详情并刷新 OSD 列表后查看最新快照。')
    } catch (err) {
      if (scope.current === current) setStatus(`${err instanceof Error ? err.message : '修改未确认'}。可能存在部分变更，请关闭详情并刷新、核对原生类别后再操作，不要直接重试。`)
    } finally {
      if (scope.current === current) setBusy(false)
      running.current = false
    }
  }
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Alert type="warning" message={`集群 ${clusterId} / OSD ${osdId}：修改设备类别可能改变 CRUSH 放置并触发数据迁移。移除旧类别与设置新类别不是原子操作。`} />
    <div>当前快照类别：{typeof record.device_class === 'string' ? record.device_class || '无类别' : '未知'}</div>
    {!version && <Alert type="warning" message="快照过期、版本无效或类别未知，禁止修改。请刷新库存后重新打开详情。" />}
    <Input aria-label="新的 OSD 设备类别" value={value} disabled={busy || attempted} onChange={event => { setValue(event.target.value); setAccepted(false) }} />
    {!valid && <Alert type="info" message="输入 1–128 位字母、数字、下划线、点或短横线，不能以点或短横线开头。" />}
    <Checkbox checked={accepted} disabled={busy || attempted} onChange={event => setAccepted(event.target.checked)}>确认目标集群与 OSD，并接受放置变化及部分失败风险</Checkbox>
    <Button danger loading={busy} disabled={!version || !valid || !accepted || busy || attempted} onClick={() => void save()}>修改设备类别</Button>
    {status && <Alert type="info" message={status} />}
  </Space>
}

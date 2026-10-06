import { Alert, Button, Popconfirm, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { type ApiRecord } from '../../api/client'
import { mutateResource, refreshResource } from '../../api/resource'

export function removalStopTarget(record: ApiRecord): { id: string; version: string } | null {
  const id = record.osd_id
  if (typeof id !== 'number' || !Number.isInteger(id) || id < 0 || id > 2147483647 || record.stale !== false) return null
  const raw = record.resource_version
  const version = typeof raw === 'string' ? raw : typeof raw === 'number' && Number.isSafeInteger(raw) ? String(raw) : ''
  if (!/^[1-9][0-9]*$/.test(version) || BigInt(version) > 18446744073709551615n) return null
  return { id: String(id), version }
}

export function OSDRemovalStop({ clusterId, record, blocked, onChanged }: { clusterId: number; record: ApiRecord; blocked: boolean; onChanged: () => void }) {
  const target = removalStopTarget(record)
  const [busy, setBusy] = useState(false)
  const [attempted, setAttempted] = useState(false)
  const [status, setStatus] = useState('')
  const running = useRef(false)
  const scope = useRef<object | null>({})
  useEffect(() => { scope.current = {}; return () => { scope.current = null } }, [])
  async function stop() {
    if (!scope.current || running.current || attempted || blocked || !target) return
    const current = scope.current
    running.current = true
    setBusy(true); setAttempted(true); setStatus('')
    try {
      await mutateResource('/osd/removal/stop', 'POST', { cluster_id: clusterId, osd_id: target.id }, { ifMatch: target.version })
      if (scope.current !== current) return
      setStatus('已确认退出移除队列，正在刷新库存。')
      try {
        await refreshResource({ clusterId, kinds: ['osd', 'osd_removal'] })
        if (scope.current === current) onChanged()
      } catch {
        if (scope.current === current) setStatus('已确认退出队列，但库存刷新失败，请手动刷新。')
      }
    } catch (err) {
      if (scope.current === current) setStatus(`${err instanceof Error ? err.message : '取消未确认'}。请刷新队列和 OSD 状态后再核对，不要直接重试。`)
    } finally {
      running.current = false
      if (scope.current === current) setBusy(false)
    }
  }
  const disabled = blocked || !target || busy || attempted
  return <Space direction="vertical">
    <Popconfirm title={`取消集群 ${clusterId} / OSD ${target?.id ?? '未知'} 的移除？`} description="可能恢复 OSD 权重并影响数据放置。不会撤销已经完成的销毁或数据清除。" disabled={disabled} onConfirm={() => stop()} okText="确认取消移除" cancelText="返回">
      <Button danger disabled={disabled} loading={busy}>取消移除</Button>
    </Popconfirm>
    {status && <Alert type="info" message={status} />}
  </Space>
}

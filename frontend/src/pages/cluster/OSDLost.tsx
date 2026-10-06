import { Alert, Button, Checkbox, Input, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { type ApiRecord } from '../../api/client'
import { mutateResource } from '../../api/resource'
import { osdDestroyTarget } from './OSDDestroy'

export function OSDLost({ clusterId, osdId, record }: { clusterId: number; osdId: string; record: ApiRecord }) {
  const [confirmation, setConfirmation] = useState('')
  const [accepted, setAccepted] = useState(false)
  const [busy, setBusy] = useState(false)
  const [attempted, setAttempted] = useState(false)
  const [status, setStatus] = useState('')
  const running = useRef(false)
  const scope = useRef<object | null>({})
  useEffect(() => { scope.current = {}; return () => { scope.current = null } }, [])
  const target = osdDestroyTarget(record, osdId)
  const expected = `mark lost osd.${osdId}`
  async function markLost() {
    if (!scope.current || running.current || attempted || !accepted || confirmation !== expected || !target) return
    const current = scope.current
    running.current = true
    setBusy(true); setAttempted(true); setStatus('')
    try {
      await mutateResource('/osd/lost', 'POST', { cluster_id: clusterId, osd_id: osdId, expected_uuid: target.uuid, confirmation }, { ifMatch: target.version })
      if (scope.current === current) setStatus('已回读确认 lost_at 与 down_at 匹配。请关闭详情并刷新 OSD 库存查看最新状态。此操作不代表数据已恢复。')
    } catch (err) {
      if (scope.current === current) setStatus(`${err instanceof Error ? err.message : 'Lost 操作未确认'}。标记可能已生效；请关闭详情并刷新、核对原生 lost_at/down_at，不要直接重试。`)
    } finally {
      if (scope.current === current) setBusy(false)
      running.current = false
    }
  }
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Alert type="error" message={`集群 ${clusterId} / OSD ${osdId}：标记 Lost 表示接受该 OSD 数据丢失，可能导致永久数据丢失。不是标记 Down、停止进程、移除 OSD 或销毁密钥。`} />
    <Alert type="info" message="后端重新核对 UUID、Down 状态和原生 safe-to-destroy 结果，再执行 Lost；以 lost_at 等于 down_at 验证，不添加虚构的 lost 状态标志。" />
    {!target && <Alert type="warning" message="需要新鲜、版本有效、UUID 已知且未销毁的 Down OSD 快照。请刷新库存后重新打开详情。" />}
    <div>目标 UUID：{typeof record.uuid === 'string' ? record.uuid : '未知'}</div>
    <Input aria-label="OSD Lost 确认文本" placeholder={expected} value={confirmation} disabled={busy || attempted} onChange={event => { setConfirmation(event.target.value); setAccepted(false) }} />
    <Checkbox checked={accepted} disabled={busy || attempted} onChange={event => setAccepted(event.target.checked)}>我已核对集群、OSD 和 UUID，接受永久数据丢失风险</Checkbox>
    <Button danger loading={busy} disabled={!target || confirmation !== expected || !accepted || busy || attempted} onClick={() => void markLost()}>标记 OSD Lost</Button>
    {status && <Alert type="info" message={status} />}
  </Space>
}

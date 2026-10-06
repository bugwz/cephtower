import { Alert, Button, Checkbox, Input, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { mutateResource } from '../../api/resource'
import { DraggableModal } from '../../components/DraggableModal'

export function HostAddressEditor({ clusterId, hostname, currentAddress, version, blocked }: { clusterId: number; hostname: string; currentAddress: string; version: string | null; blocked: boolean }) {
  const [open, setOpen] = useState(false)
  const [address, setAddress] = useState(currentAddress)
  const [accepted, setAccepted] = useState(false)
  const [busy, setBusy] = useState(false)
  const [attempted, setAttempted] = useState(false)
  const [status, setStatus] = useState('')
  const running = useRef(false)
  const scope = useRef<object | null>({})
  useEffect(() => { scope.current = {}; return () => { scope.current = null } }, [])
  const valid = address.length > 0 && address.trim() === address && !address.startsWith('-') && !/\s/.test(address)
  async function save() {
    if (!scope.current || running.current || attempted || blocked || !version || !hostname || !accepted || !valid) return
    const current = scope.current
    running.current = true
    setBusy(true); setAttempted(true); setStatus('')
    try {
      await mutateResource('/host', 'PATCH', { cluster_id: clusterId, host: hostname, address }, { ifMatch: version })
      if (scope.current === current) setStatus('编排器地址已回读确认。请关闭弹窗并刷新主机库存；本项目的 SSH 连接配置未修改。')
    } catch (err) {
      if (scope.current === current) setStatus(`${err instanceof Error ? err.message : '地址更新未确认'}。请刷新并核对原生主机地址，不要直接重试。`)
    } finally {
      running.current = false
      if (scope.current === current) setBusy(false)
    }
  }
  return <>
    <Button disabled={blocked || !version} onClick={() => setOpen(true)}>编排器地址</Button>
    <DraggableModal title={`集群 ${clusterId} / 主机 ${hostname} 编排器地址`} open={open} onCancel={() => setOpen(false)} footer={null}>
      <Space direction="vertical" style={{ width: '100%' }}>
        <Alert type="warning" message="修改 ceph orch host set-addr 使用的主机地址，可能影响 cephadm 连接、部署和管理。不会修改本项目保存的 SSH 地址、端口或凭据。" />
        <div>当前库存地址：{currentAddress || '未知'}</div>
        <Input aria-label="新的编排器主机地址" value={address} disabled={busy || attempted} onChange={event => { setAddress(event.target.value); setAccepted(false) }} />
        {!valid && <Alert type="info" message="请输入不含空白、且不以短横线开头的地址；最终有效性由 Ceph 原生检查确认。" />}
        <Checkbox checked={accepted} disabled={busy || attempted} onChange={event => setAccepted(event.target.checked)}>确认目标主机和新地址，接受连接可能中断的风险</Checkbox>
        <Button danger loading={busy} disabled={blocked || !version || !valid || !accepted || busy || attempted} onClick={() => void save()}>修改编排器地址</Button>
        {status && <Alert type="info" message={status} />}
      </Space>
    </DraggableModal>
  </>
}

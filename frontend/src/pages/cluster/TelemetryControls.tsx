import { Alert, Button, Checkbox, Modal, Space, Typography } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { mutateResource } from '../../api/resource'
import { message } from '../../utils/appMessage'

export function TelemetryControls({ clusterId, enabled, disabled, onComplete }: { clusterId: number; enabled: boolean; disabled: boolean; onComplete: () => Promise<void> }) {
  const [open, setOpen] = useState(false)
  const [accepted, setAccepted] = useState(false)
  const [busy, setBusy] = useState(false)
  const running = useRef(false)
  const active = useRef(true)
  useEffect(() => { active.current = true; return () => { active.current = false } }, [])
  async function submit() {
    if (!active.current || running.current || disabled || (!enabled && !accepted)) return
    running.current = true; setBusy(true)
    try {
      await mutateResource('/manager/telemetry', 'PATCH', { cluster_id: clusterId, enabled: !enabled, ...(!enabled ? { license: 'sharing-1-0' } : {}) })
      if (!active.current) return
      setOpen(false); setAccepted(false)
      message.success(enabled ? '遥测停用状态已核验' : '遥测启用状态已核验')
      await onComplete()
    } finally { running.current = false; if (active.current) setBusy(false) }
  }
  return <>
    <Button danger={!enabled} disabled={disabled || busy} onClick={() => { setAccepted(false); setOpen(true) }}>{enabled ? '停用遥测' : '启用遥测'}</Button>
    <Modal title={enabled ? '确认停用遥测' : '确认启用遥测及数据共享许可'} open={open} confirmLoading={busy} okText={enabled ? '确认停用' : '同意许可并启用'} okButtonProps={{ danger: !enabled, disabled: disabled || (!enabled && !accepted) }} onCancel={() => { if (!busy) { setOpen(false); setAccepted(false) } }} onOk={submit}>
      <Space direction="vertical">
        <Alert type="warning" showIcon message={enabled ? '将执行 ceph telemetry off；停止后续遥测，不会撤回已上传的数据。' : '启用会同意当前版本的全部遥测数据集合，并允许 Ceph 按已配置通道和目标地址向外部发送报告。请先查看报告样本、通道及上传地址。'} />
        {!enabled && <><Typography.Paragraph>许可：Community Data License Agreement – Sharing – Version 1.0（sharing-1-0）。</Typography.Paragraph><Typography.Link href="https://cdla.io/sharing-1-0/" target="_blank" rel="noopener noreferrer">阅读数据共享许可</Typography.Link><Checkbox checked={accepted} disabled={busy} onChange={(event) => setAccepted(event.target.checked)}>我已阅读并同意 sharing-1-0 数据共享许可，确认启用此集群的遥测上传。</Checkbox></>}
      </Space>
    </Modal>
  </>
}

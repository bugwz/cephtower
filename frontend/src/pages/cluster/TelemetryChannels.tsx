import { Alert, Button, Modal, Space, Typography } from 'antd'
import { useEffect, useRef, useState } from 'react'
import type { ApiRecord } from '../../api/client'
import { mutateResource } from '../../api/resource'
import { message } from '../../utils/appMessage'

const channels = [
  ['basic', '基本集群信息', '集群规模、版本和基础配置'],
  ['ident', '身份及联系信息', '已填写的描述、组织和联系信息'],
  ['crash', '崩溃信息', '守护进程崩溃元数据、版本和调用栈'],
  ['device', '设备健康信息', '设备健康和 SMART 数据'],
  ['perf', '性能信息', '性能计数器和统计信息'],
] as const

export function TelemetryChannels({ clusterId, status, disabled, onComplete }: { clusterId: number; status: ApiRecord; disabled: boolean; onComplete: () => Promise<void> }) {
  const [selected, setSelected] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const active = useRef(true)
  const running = useRef(false)
  useEffect(() => { active.current = true; return () => { active.current = false } }, [])
  const channel = channels.find(([name]) => name === selected)
  const current = channel ? status[`channel_${channel[0]}`] : undefined
  async function submit() {
    if (!active.current || running.current || disabled || status.enabled !== true || !channel || typeof current !== 'boolean') return
    running.current = true; setBusy(true)
    try {
      await mutateResource('/manager/telemetry/channel', 'PATCH', { cluster_id: clusterId, channel: channel[0], enabled: !current })
      if (!active.current) return
      setSelected(null)
      message.success('遥测通道状态已核验')
    } finally {
      // Refresh even on an unconfirmed write; never automatically retry it.
      try { if (active.current) { setSelected(null); await onComplete() } }
      finally { running.current = false; if (active.current) setBusy(false) }
    }
  }
  return <Space direction="vertical">
    <Typography.Text type="secondary">原生命令仅在遥测已启用时支持修改通道；不会自动启用遥测。开启通道后，后续上传可能包含对应数据，请先查看报告样本。</Typography.Text>
    <Space wrap>{channels.map(([name, label]) => <Button key={name} disabled={disabled || busy || status.enabled !== true || typeof status[`channel_${name}`] !== 'boolean'} onClick={() => setSelected(name)}>{status[`channel_${name}`] === true ? '关闭' : '开启'}{label}</Button>)}</Space>
    <Modal open={Boolean(channel)} title="确认修改遥测通道" confirmLoading={busy} okText="确认修改" okButtonProps={{ disabled: disabled || status.enabled !== true || typeof current !== 'boolean', danger: current === false }} onOk={submit} onCancel={() => { if (!busy) setSelected(null) }}>
      <Alert type="warning" showIcon message={`${current === true ? '关闭' : '开启'}${channel?.[1] ?? ''}`} description={current === true ? '停止在后续报告中共享此通道数据，不会撤回已上传内容。' : `后续报告可能共享${channel?.[2] ?? ''}。具体字段以当前 Ceph 版本及报告预览为准。`} />
    </Modal>
  </Space>
}

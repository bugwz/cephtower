import { Alert, Button, Modal } from 'antd'
import { useState } from 'react'
import { ConfigurationPage } from './ConfigurationPage'

const telemetryOptionNames = ['interval', 'proxy', 'contact', 'description', 'organization', 'leaderboard', 'leaderboard_description', 'url', 'device_url'].map((name) => `mgr/telemetry/${name}`)

export function TelemetrySettings({ disabled, onClose }: { disabled: boolean; onClose: () => Promise<void> }) {
  const [open, setOpen] = useState(false)
  return <>
    <Button disabled={disabled} onClick={() => setOpen(true)}>上传与身份配置</Button>
    <Modal title="遥测上传与身份配置" open={open} footer={null} width="95vw" destroyOnClose onCancel={() => { setOpen(false); void onClose() }}>
      <Alert type="warning" showIcon message="修改上传地址、代理或身份信息会影响后续报告共享。" description="此面板仅管理上传间隔、目标、代理、身份及排行榜配置；不启用遥测、不修改通道、不立即发送报告。先从集群刷新配置库存，再设置 mgr 作用域覆盖值。删除覆盖值会恢复继承值或默认值，实际生效状态请在关闭面板后核对。" />
      {open && <ConfigurationPage moduleName="telemetry" optionNames={telemetryOptionNames} />}
    </Modal>
  </>
}

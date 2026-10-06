import { Alert } from 'antd'
import { ConfigurationPage } from './ConfigurationPage'

export const osdRecoveryOptions = [
  'osd_max_backfills', 'osd_recovery_max_active',
  'osd_recovery_max_single_start', 'osd_recovery_sleep',
  'osd_op_queue', 'osd_mclock_profile', 'osd_mclock_override_recovery_settings'
] as const

export function OSDRecoveryConfiguration() {
  return <>
    <Alert type="info" message="OSD 恢复速度与 QoS 参数" description="包含参考 Dashboard 的四个恢复参数，以及调度器和 mClock 配置。按项设置或移除覆盖；确认作用域为 osd 或所需实例，global 会影响所有适用守护进程。" />
    <Alert type="warning" message="写入成功不等于恢复速度已改变" description="mClock 可能覆盖恢复并发参数，并禁用恢复 sleep 参数。先核对 osd_op_queue、mClock profile 与 override 设置及参数元数据；此页展示配置数据库覆盖，不代表守护进程实时生效值。不会自动应用优先级预设或批量修改。" />
    <ConfigurationPage optionNames={osdRecoveryOptions} />
  </>
}

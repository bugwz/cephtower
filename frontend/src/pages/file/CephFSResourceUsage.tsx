import { Progress, Tooltip } from 'antd'
import { cephFSPermissions, cephFSUsage } from './cephfsSubvolumeSummary'

export function CephFSUsage({ quota, used, percent }: { quota: unknown; used: unknown; percent: unknown }) {
  const usage = cephFSUsage(quota, used, percent)
  const title = quota === 'infinite' ? '未设置配额限制' : usage.percent === undefined ? '原生使用率不可用' : usage.percent > 100 ? '已超过配额' : '配额使用率'
  return <Tooltip title={title}>
    <div style={{ minWidth: 160 }}>
      {usage.barPercent !== undefined && <Progress percent={usage.barPercent} size="small" status={usage.percent! > 100 ? 'exception' : 'normal'} format={() => `${usage.percent!.toFixed(2)}%`} />}
      <span>{usage.text}</span>
    </div>
  </Tooltip>
}

export function CephFSPermissions({ mode }: { mode: unknown }) {
  const permissions = cephFSPermissions(mode)
  if (!permissions) return <>—</>
  return <Tooltip title={`所有者 ${permissions.owner}；组 ${permissions.group}；其他 ${permissions.others}；八进制 ${permissions.octal}`}>
    <code>{permissions.owner}{permissions.group}{permissions.others} ({permissions.octal})</code>
  </Tooltip>
}

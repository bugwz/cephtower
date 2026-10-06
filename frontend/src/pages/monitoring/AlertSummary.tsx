import { Alert, Descriptions } from 'antd'
import type { ApiRecord } from '../../api/client'

export function alertSnapshotCounts(rows: ApiRecord[] | undefined) {
  if (!rows) return null
  const counts = { active: 0, critical: 0, warning: 0, info: 0, otherSeverity: 0, suppressed: 0, unprocessed: 0, unknownState: 0 }
  for (const row of rows) {
    const status = row.status
    const state = status && typeof status === 'object' && !Array.isArray(status) ? (status as ApiRecord).state : undefined
    if (state === 'active') {
      counts.active++
      const labels = row.labels
      const severity = labels && typeof labels === 'object' && !Array.isArray(labels) ? (labels as ApiRecord).severity : undefined
      if (severity === 'critical' || severity === 'warning' || severity === 'info') counts[severity]++
      else counts.otherSeverity++
    } else if (state === 'suppressed' || state === 'unprocessed') counts[state]++
    else counts.unknownState++
  }
  return counts
}

export function AlertSummary({ rows }: { rows: ApiRecord[] | undefined }) {
  const counts = alertSnapshotCounts(rows)
  if (!counts) return <Alert type="info" message="告警统计暂不可用，等待成功读取所选集群快照。" />
  return <><Alert type="info" message="统计所选集群查询快照中的全部告警实例，不受下方表格筛选影响，不是历史累计数或原生路由组数。严重程度仅统计 active 实例；未提供或未知状态单独计数。" />
    <Descriptions size="small" column={{ xs: 1, sm: 2, md: 3 }} items={[
      { key: 'active', label: '活动实例', children: counts.active },
      { key: 'critical', label: '活动 / critical', children: counts.critical },
      { key: 'warning', label: '活动 / warning', children: counts.warning },
      { key: 'info', label: '活动 / info', children: counts.info },
      { key: 'other', label: '活动 / 其他或未知严重程度', children: counts.otherSeverity },
      { key: 'suppressed', label: '已抑制', children: counts.suppressed },
      { key: 'unprocessed', label: '待处理', children: counts.unprocessed },
      { key: 'unknown', label: '未知状态', children: counts.unknownState }
    ]} />
  </>
}

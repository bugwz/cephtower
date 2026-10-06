import { ExternalListPage, type ExternalListPageDefinition } from '../ExternalListPage'
import { useState } from 'react'
import { Alert, Segmented } from 'antd'
import { alertGroupDefinition } from './AlertGroups'
export { RuntimeLogsPage } from './RuntimeLogsPage'
import { MetricPage } from './MetricPage'
import { SilencedAlerts } from './SilencedAlerts'
import { RuleAlerts } from './RuleAlerts'
import { AlertRoutingDetails } from './AlertRoutingDetails'
import { alertColumns, alertRuleColumns, silenceColumns } from './alertColumns'
import { silenceCreateAction, silenceFromAlertAction, silenceRecreateAction, silenceExpireAction, silenceUpdateAction } from './silenceActions'

export function MonitorOverviewPage() {
  return <ExternalListPage definition={externalDefinitions.grafana} />
}

export function PerformanceMetricsPage() {
  return <MetricPage />
}


export function AlertListPage() {
  const [grouped, setGrouped] = useState(false)
  return <><Segmented aria-label="告警展示方式" options={[{ label: '告警实例', value: 'instances' }, { label: '原生分组', value: 'groups' }]} value={grouped ? 'groups' : 'instances'} onChange={value => setGrouped(value === 'groups')} />
    <Alert type="info" message="告警实例与分组均按所选集群的 FSID 精确匹配 cluster 标签；缺少该标签的告警不会显示。FSID 尚未探测时读取会失败，不退回共享端点的全部告警。" />
    <ExternalListPage key={grouped ? 'groups' : 'instances'} definition={grouped ? alertGroupDefinition : externalDefinitions.alerts} /></>
}

export function AlertRulesPage() {
  return <ExternalListPage definition={externalDefinitions.rules} />
}

export function AlertSilencesPage() {
  return <ExternalListPage definition={externalDefinitions.silences} />
}

const externalDefinitions: Record<'grafana' | 'alerts' | 'rules' | 'silences', ExternalListPageDefinition> = {
  grafana: {
    title: '监控总览',
    path: '/grafana',
    requiredEndpoints: ['grafana'],
    rowKeyCandidates: ['uid'],
    columns: [
      { key: 'title', title: '看板' },
      { key: 'uid', title: 'UID' },
      { key: 'folderTitle', title: '文件夹' },
      { key: 'uri', title: 'URI' },
      { key: 'url', title: 'URL' },
      { key: 'tags', title: '标签' }
    ]
  },
  alerts: {
    title: '告警列表',
    path: '/alert/alerts',
    autoRefreshMs: 5000,
    requiredEndpoints: ['alertmanager'],
    rowKeyCandidates: ['fingerprint'],
    columns: alertColumns,
    detailContent: row => <AlertRoutingDetails row={row} />,
    extraActions: [silenceFromAlertAction]
  },
  rules: {
    title: '告警规则',
    path: '/alert/rules',
    requiredEndpoints: ['prometheus'],
    rowKeyCandidates: ['rule_key'],
    detailContent: row => <RuleAlerts row={row} />,
    columns: alertRuleColumns
  },
  silences: {
    title: '告警静默',
    path: '/alert/silences',
    requiredEndpoints: ['alertmanager'],
    rowKeyCandidates: ['id', 'silence_id'],
    createAction: silenceCreateAction,
    updateAction: silenceUpdateAction,
    detailContent: (row, clusterId) => <SilencedAlerts clusterId={clusterId} silenceId={row.id} />,
    extraActions: [silenceRecreateAction],
    deleteAction: silenceExpireAction,
    columns: silenceColumns
  }
}

import { ExternalListPage, type ExternalListPageDefinition } from '../ExternalListPage'
export { RuntimeLogsPage } from './RuntimeLogsPage'
import { MetricPage } from './MetricPage'
import { alertColumns, alertRuleColumns } from './alertColumns'
import { silenceCreateAction, silenceFromAlertAction } from './silenceActions'

export function MonitorOverviewPage() {
  return <ExternalListPage definition={externalDefinitions.grafana} />
}

export function PerformanceMetricsPage() {
  return <MetricPage />
}


export function AlertListPage() {
  return <ExternalListPage definition={externalDefinitions.alerts} />
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
    columns: [
      { key: 'title', title: '看板' },
      { key: 'uid', title: 'UID' },
      { key: 'uri', title: 'URI' },
      { key: 'url', title: 'URL' },
      { key: 'tags', title: '标签' }
    ]
  },
  alerts: {
    title: '告警列表',
    path: '/alert/alerts',
    requiredEndpoints: ['alertmanager'],
    rowKeyCandidates: ['fingerprint'],
    columns: alertColumns,
    extraActions: [silenceFromAlertAction]
  },
  rules: {
    title: '告警规则',
    path: '/alert/rules',
    requiredEndpoints: ['prometheus'],
    rowKeyCandidates: ['rule_key'],
    columns: alertRuleColumns
  },
  silences: {
    title: '告警静默',
    path: '/alert/silences',
    requiredEndpoints: ['alertmanager'],
    rowKeyCandidates: ['id', 'silence_id'],
    createAction: silenceCreateAction,
    deleteAction: {
      title: '删除告警静默',
      path: '/alert/silence',
      action: 'silence.delete',
      resourceKind: 'silence',
      risk: 'medium',
      successMessage: '告警静默删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, silence_id: silenceId(row) }),
      resourceKey: (row) => `silence/${silenceId(row)}`
    },
    columns: [
      { key: 'id', title: 'ID' },
      { key: 'status', title: '状态' },
      { key: 'matchers', title: 'Matchers' },
      { key: 'startsAt', title: '开始时间' },
      { key: 'endsAt', title: '结束时间' },
      { key: 'createdBy', title: '创建人' }
    ]
  }
}

function silenceId(row?: Record<string, unknown>) {
  return String(row?.id ?? row?.silence_id ?? '').trim()
}

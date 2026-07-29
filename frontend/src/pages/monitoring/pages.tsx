import { ExternalListPage, type ExternalListPageDefinition } from '../ExternalListPage'
import { ResourceListPage, type ResourceListPageDefinition } from '../ResourceListPage'
import { MetricPage } from './MetricPage'

export function MonitorOverviewPage() {
  return <ExternalListPage definition={externalDefinitions.grafana} />
}

export function PerformanceMetricsPage() {
  return <MetricPage />
}

export function RuntimeLogsPage() {
  return <ResourceListPage definition={resourceDefinitions.runtimeLogs} />
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
    columns: [
      { key: 'labels', title: 'Labels' },
      { key: 'annotations', title: 'Annotations' },
      { key: 'state', title: '状态' },
      { key: 'startsAt', title: '开始时间' },
      { key: 'endsAt', title: '结束时间' }
    ]
  },
  rules: {
    title: '告警规则',
    path: '/alert/rules',
    requiredEndpoints: ['alertmanager'],
    columns: [
      { key: 'name', title: '名称' },
      { key: 'state', title: '状态' },
      { key: 'query', title: '查询' },
      { key: 'duration', title: '持续时间' },
      { key: 'labels', title: 'Labels' }
    ]
  },
  silences: {
    title: '告警静默',
    path: '/alert/silences',
    requiredEndpoints: ['alertmanager'],
    rowKeyCandidates: ['id', 'silence_id'],
    createAction: {
      title: '新建告警静默',
      buttonLabel: '新建静默',
      path: '/alert/silence',
      method: 'POST',
      successMessage: '告警静默创建执行成功',
      fields: [
        { name: 'matchers_json', label: 'Matchers JSON', type: 'textarea', required: true, placeholder: '[{"name":"alertname","value":"OSDNearFull","isRegex":false,"isEqual":true}]' },
        { name: 'startsAt', label: '开始时间 RFC3339', required: true },
        { name: 'endsAt', label: '结束时间 RFC3339', required: true },
        { name: 'createdBy', label: '创建人', required: true },
        { name: 'comment', label: '说明', type: 'textarea', required: true }
      ],
      initialValues: () => {
        const start = new Date()
        const end = new Date(start.getTime() + 2 * 60 * 60 * 1000)
        return {
          matchers_json: '[{"name":"alertname","value":"","isRegex":false,"isEqual":true}]',
          startsAt: start.toISOString(),
          endsAt: end.toISOString(),
          createdBy: 'cephtower',
          comment: ''
        }
      },
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        matchers: parseJSONArray(values.matchers_json),
        startsAt: String(values.startsAt ?? ''),
        endsAt: String(values.endsAt ?? ''),
        createdBy: String(values.createdBy ?? ''),
        comment: String(values.comment ?? '')
      })
    },
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

function parseJSONArray(value: unknown) {
  const parsed = JSON.parse(String(value ?? '[]'))
  if (!Array.isArray(parsed)) {
    throw new Error('JSON 字段必须是数组')
  }
  return parsed
}

const resourceDefinitions: Record<'runtimeLogs', ResourceListPageDefinition> = {
  runtimeLogs: {
    title: '运行日志',
    path: '/logs',
    columns: [
      { key: 'name', title: '名称' },
      { key: 'status', title: '状态' },
      { key: 'level', title: '级别' },
      { key: 'message', title: '消息' },
      { key: 'timestamp', title: '时间' }
    ]
  }
}

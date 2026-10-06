import type { ApiRecord } from '../../api/client'
import type { ResourceFormAction } from '../ResourceListPage'

export function silenceMatchers(value: unknown) {
  const parsed: unknown = JSON.parse(String(value ?? '[]'))
  if (!Array.isArray(parsed) || !parsed.length || parsed.some(item => !item || typeof item !== 'object' || Array.isArray(item) || typeof item.name !== 'string' || !item.name.trim() || typeof item.value !== 'string' || typeof item.isRegex !== 'boolean' || typeof item.isEqual !== 'boolean')) {
    throw new Error('Matchers 必须为非空数组，每项包含 name、value 字符串及 isRegex、isEqual 布尔值')
  }
  return parsed
}

export function alertSilenceName(row?: ApiRecord): string | undefined {
  const labels = row?.labels
  if (!labels || typeof labels !== 'object' || Array.isArray(labels)) return undefined
  const name = (labels as ApiRecord).alertname
  return typeof name === 'string' && name.trim() ? name : undefined
}

export const silenceCreateAction: ResourceFormAction = {
  title: '新建告警静默', buttonLabel: '新建静默', path: '/alert/silence', method: 'POST',
  successMessage: '告警静默创建执行成功',
  confirmation: () => '静默将暂停所选时间范围内匹配条件命中的告警通知，不会修复告警原因。仅使用 alertname 条件会影响所有同名告警，不限于当前实例；请确认匹配条件和时间范围。',
  fields: [
    { name: 'matchers_json', label: 'Matchers JSON（所有条件同时匹配）', type: 'textarea', required: true, placeholder: '[{"name":"alertname","value":"OSDNearFull","isRegex":false,"isEqual":true}]' },
    { name: 'startsAt', label: '开始时间 RFC3339', required: true },
    { name: 'endsAt', label: '结束时间 RFC3339', required: true },
    { name: 'createdBy', label: '创建人', required: true },
    { name: 'comment', label: '说明', type: 'textarea', required: true }
  ],
  initialValues: (row) => {
    const start = new Date()
    return {
      matchers_json: JSON.stringify([{ name: 'alertname', value: alertSilenceName(row) ?? '', isRegex: false, isEqual: true }], null, 2),
      startsAt: start.toISOString(), endsAt: new Date(start.getTime() + 2 * 60 * 60 * 1000).toISOString(),
      createdBy: 'cephtower', comment: ''
    }
  },
  buildBody: (values, clusterId) => ({
    cluster_id: clusterId, matchers: silenceMatchers(values.matchers_json),
    startsAt: String(values.startsAt ?? ''), endsAt: String(values.endsAt ?? ''),
    createdBy: String(values.createdBy ?? ''), comment: String(values.comment ?? '')
  })
}

export const silenceFromAlertAction: ResourceFormAction = {
  ...silenceCreateAction, title: '从告警创建静默', buttonLabel: '创建静默',
  disabledWhen: (row) => alertSilenceName(row) === undefined ? '告警缺少有效的 alertname 标签，无法预填静默条件' : undefined
}

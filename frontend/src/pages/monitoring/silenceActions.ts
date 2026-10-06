import type { ApiRecord } from '../../api/client'
import type { ResourceDeleteAction, ResourceFormAction } from '../ResourceListPage'

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

export function silenceTimeRange(startsAt: unknown, endsAt: unknown) {
  const parse = (value: unknown) => {
    const parts = typeof value === 'string' ? /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.(\d{1,9}))?(Z|[+-](\d{2}):(\d{2}))$/.exec(value) : null
    if (!parts || typeof value !== 'string') throw new Error('静默时间必须为包含时区的 RFC3339 格式')
    const [year, month, day, hour, minute, second] = parts.slice(1, 7).map(Number)
    const leap = year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0)
    const days = [31, leap ? 29 : 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31]
    if (month < 1 || month > 12 || day < 1 || day > days[month - 1] || hour > 23 || minute > 59 || second > 59 || Number(parts[9] ?? 0) > 23 || Number(parts[10] ?? 0) > 59) throw new Error('静默时间无效')
    const milliseconds = Date.parse(value)
    if (!Number.isFinite(milliseconds)) throw new Error('静默时间无效')
    return BigInt(Math.floor(milliseconds / 1000)) * BigInt(1000000000) + BigInt((parts[7] ?? '').padEnd(9, '0'))
  }
  if (parse(endsAt) <= parse(startsAt)) throw new Error('静默结束时间必须晚于开始时间')
  return { startsAt: startsAt as string, endsAt: endsAt as string }
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
    ...silenceTimeRange(values.startsAt, values.endsAt),
    createdBy: String(values.createdBy ?? ''), comment: String(values.comment ?? '')
  })
}

export const silenceFromAlertAction: ResourceFormAction = {
  ...silenceCreateAction, title: '从告警创建静默', buttonLabel: '创建静默',
  disabledWhen: (row) => {
    if (alertSilenceName(row) === undefined) return '告警缺少有效的 alertname 标签，无法预填静默条件'
    if (alertClusterName(row) === undefined) return '告警缺少有效的 cluster 标签，无法预填集群范围'
    return undefined
  },
  initialValues: row => {
    const name = alertSilenceName(row)
    const cluster = alertClusterName(row)
    if (name === undefined || cluster === undefined) throw new Error('告警名称或集群标签缺失，请刷新列表')
    const start = new Date()
    return {
      matchers_json: JSON.stringify([
        { name: 'alertname', value: name, isRegex: false, isEqual: true },
        { name: 'cluster', value: cluster, isRegex: false, isEqual: true }
      ], null, 2),
      startsAt: start.toISOString(), endsAt: new Date(start.getTime() + 2 * 60 * 60 * 1000).toISOString(),
      createdBy: 'cephtower', comment: ''
    }
  },
  confirmation: () => '默认同时匹配告警名与 cluster 标签，将影响该集群所有同名告警，不限于当前实例。表单条件可编辑，移除或放宽 cluster 条件可能影响其他集群；请确认最终匹配范围与时间。静默只暂停通知，不会修复告警原因。'
}

export function alertClusterName(row?: ApiRecord): string | undefined {
  const labels = row?.labels
  if (!labels || typeof labels !== 'object' || Array.isArray(labels)) return undefined
  const cluster = (labels as ApiRecord).cluster
  return typeof cluster === 'string' && cluster.trim() ? cluster : undefined
}

export function silenceRecreateBlocked(row: ApiRecord): string | undefined {
  const status = row.status
  if (!status || typeof status !== 'object' || Array.isArray(status) || (status as ApiRecord).state !== 'expired') return '仅可重新创建原生状态为 expired 的静默'
  try { silenceMatchers(JSON.stringify(row.matchers)) } catch { return '原静默的匹配条件缺失或格式异常，请手动新建' }
  return undefined
}

export const silenceRecreateAction: ResourceFormAction = {
  ...silenceCreateAction,
  title: '重新创建已过期静默', buttonLabel: '重新创建',
  successMessage: '新的告警静默创建执行成功，旧记录未修改',
  disabledWhen: silenceRecreateBlocked,
  initialValues: (row) => {
    const start = new Date()
    return {
      matchers_json: JSON.stringify(row?.matchers ?? [], null, 2),
      startsAt: start.toISOString(), endsAt: new Date(start.getTime() + 2 * 60 * 60 * 1000).toISOString(),
      createdBy: typeof row?.createdBy === 'string' ? row.createdBy : '',
      comment: typeof row?.comment === 'string' ? row.comment : ''
    }
  },
  confirmation: () => '将按当前表单条件与时间创建一条新的静默，不修改原已过期记录。匹配的告警通知将被暂停，请确认匹配范围、创建人、说明和新的时间范围。'
}

export function silenceTarget(row?: ApiRecord): string {
  const id = row?.id
  if (typeof id !== 'string' || !id || id.trim() !== id || /[\/\x00-\x1f\x7f]/.test(id) || id === '.' || id === '..') throw new Error('静默 ID 无效，请刷新列表')
  return id
}

export const silenceExpireAction: ResourceDeleteAction = {
  title: '结束告警静默', buttonLabel: '结束静默', path: '/alert/silence',
  action: 'silence.delete', resourceKind: 'silence', risk: 'medium',
  successMessage: '静默结束请求执行成功，请刷新确认状态',
  confirmation: () => '将使所选静默立即过期，不删除历史记录。告警若仍满足通知条件且未被其他静默或抑制规则覆盖，可能恢复通知。',
  disabledWhen: (row) => {
    try { silenceTarget(row) } catch { return '静默 ID 无效，请刷新列表' }
    const status = row.status
    const state = status && typeof status === 'object' && !Array.isArray(status) ? (status as ApiRecord).state : undefined
    return state === 'active' || state === 'pending' ? undefined : '仅可结束 active 或 pending 状态的静默'
  },
  buildBody: (row, clusterId) => ({ cluster_id: clusterId, silence_id: silenceTarget(row) }),
  resourceKey: (row) => `silence/${silenceTarget(row)}`
}

export function silenceUpdateBlocked(row: ApiRecord): string | undefined {
  const blocked = silenceExpireAction.disabledWhen?.(row)
  if (blocked) return blocked
  for (const key of ['startsAt', 'endsAt', 'updatedAt']) {
    if (typeof row[key] !== 'string' || !Number.isFinite(Date.parse(row[key]))) return '静默时间或更新时间缺失，请刷新后编辑'
  }
  try { silenceMatchers(JSON.stringify(row.matchers)) } catch { return '静默匹配条件异常，请刷新后编辑' }
  return undefined
}

export const silenceUpdateAction: ResourceFormAction = {
  ...silenceCreateAction, title: '编辑告警静默', method: 'PATCH',
  successMessage: '静默更新请求执行成功，请刷新确认状态',
  disabledWhen: silenceUpdateBlocked,
  initialValues: (row) => ({
    matchers_json: JSON.stringify(row?.matchers ?? [], null, 2),
    startsAt: typeof row?.startsAt === 'string' ? row.startsAt : '', endsAt: typeof row?.endsAt === 'string' ? row.endsAt : '',
    createdBy: typeof row?.createdBy === 'string' ? row.createdBy : '', comment: typeof row?.comment === 'string' ? row.comment : ''
  }),
  buildBody: (values, clusterId, row) => {
    if (!row) throw new Error('静默记录缺失')
    const blocked = silenceUpdateBlocked(row)
    if (blocked) throw new Error(blocked)
    return { ...silenceCreateAction.buildBody(values, clusterId), silence_id: silenceTarget(row), expected_updated_at: row.updatedAt }
  },
  confirmation: () => '将修改已有静默的匹配条件、时间与说明，可能改变告警通知范围。后端会先核对当前更新时间，但 Alertmanager 不提供原子条件更新；请避免并发编辑，并在提交后刷新确认。'
}

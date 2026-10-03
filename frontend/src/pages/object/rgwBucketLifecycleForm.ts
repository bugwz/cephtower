export type LifecycleSelector = { kind: 'Prefix' | 'Filter'; and: boolean; prefix: string | null; tags: { key: string; value: string }[]; object_size_greater_than: string | null; object_size_less_than: string | null; archive_zone: boolean }
export type LifecycleAction = { type: string; fields: Record<string, string> }
export type LifecycleRule = { id: string; status: 'Enabled' | 'Disabled'; selector: LifecycleSelector; actions: LifecycleAction[] }
export type LifecycleDraft = { rules: LifecycleRule[] }
export const lifecycleActions: Record<string, { label: string; fields: string[] }> = {
  Expiration: { label: '当前版本过期', fields: ['Days', 'Date', 'ExpiredObjectDeleteMarker'] },
  Transition: { label: '当前版本转换', fields: ['Days', 'Date', 'StorageClass'] },
  NoncurrentVersionExpiration: { label: '非当前版本过期', fields: ['NoncurrentDays', 'NewerNoncurrentVersions'] },
  NoncurrentVersionTransition: { label: '非当前版本转换', fields: ['NoncurrentDays', 'StorageClass'] },
  AbortIncompleteMultipartUpload: { label: '终止未完成分段上传', fields: ['DaysAfterInitiation'] }
}
export const lifecycleFields: Record<string, string> = { Days: '天数', Date: 'UTC 日期', ExpiredObjectDeleteMarker: '清除过期删除标记', StorageClass: '目标存储类别', NoncurrentDays: '成为非当前版本后的天数', NewerNoncurrentVersions: '保留较新非当前版本数量', DaysAfterInitiation: '上传发起后的天数' }
const own = (value: object, key: string) => Object.prototype.hasOwnProperty.call(value, key)
const record = (value: unknown): value is Record<string, unknown> => !!value && typeof value === 'object' && !Array.isArray(value)

export function lifecycleDraft(value: unknown): LifecycleDraft {
  const rules = record(value) ? value.rules : undefined
  if (!Array.isArray(rules)) throw new Error('生命周期规则不可用')
  for (const rule of rules) {
    if (!record(rule) || typeof rule.id !== 'string' || (rule.status !== 'Enabled' && rule.status !== 'Disabled') || !Array.isArray(rule.actions)) throw new Error('生命周期规则格式无效')
    const selector = rule.selector
    if (!record(selector) || (selector.kind !== 'Prefix' && selector.kind !== 'Filter') || typeof selector.and !== 'boolean' || typeof selector.archive_zone !== 'boolean' || !Array.isArray(selector.tags)
      || ['prefix', 'object_size_greater_than', 'object_size_less_than'].some(key => selector[key] !== null && typeof selector[key] !== 'string')
      || selector.tags.some(tag => !record(tag) || typeof tag.key !== 'string' || typeof tag.value !== 'string')) throw new Error('生命周期过滤条件无效')
    for (const action of rule.actions) {
      if (!record(action) || typeof action.type !== 'string' || !own(lifecycleActions, action.type) || !record(action.fields) || Object.entries(action.fields).some(([key, item]) => !lifecycleActions[action.type as string].fields.includes(key) || typeof item !== 'string')) throw new Error('生命周期动作字段无效')
    }
  }
  return { rules: (rules as LifecycleRule[]).map(rule => ({ id: rule.id, status: rule.status, selector: { ...rule.selector, tags: rule.selector.tags.map(tag => ({ ...tag })) }, actions: rule.actions.map(action => ({ type: action.type, fields: { ...action.fields } })) })) }
}

function xml(value: string) {
  for (const character of value) {
    const code = character.codePointAt(0)!
    if (!(code === 9 || code === 10 || code === 13 || (code >= 32 && code <= 0xd7ff) || (code >= 0xe000 && code <= 0xfffd) || code >= 0x10000)) throw new Error('字段包含 XML 不支持的字符')
  }
  return value.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&apos;').replace(/\r/g, '&#13;')
}
const element = (name: string, value: string) => `<${name}>${xml(value)}</${name}>`

export function lifecycleDocument(value: unknown): string {
  const { rules } = lifecycleDraft(value)
  if (!rules.length) throw new Error('至少需要一条规则；关闭生命周期请使用删除配置操作')
  const ids = new Set<string>()
  const document = '<LifecycleConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/">' + rules.map(rule => {
    if (new TextEncoder().encode(rule.id).length > 255 || (rule.id !== '' && ids.has(rule.id))) throw new Error('规则 ID 最多 255 UTF-8 字节且非空 ID 不得重复')
    if (rule.id) ids.add(rule.id)
    const selector = rule.selector
    if (selector.kind === 'Prefix' && (selector.prefix === null || selector.and || selector.tags.length || selector.archive_zone || selector.object_size_greater_than !== null || selector.object_size_less_than !== null)) throw new Error('旧式 Prefix 只能配置前缀；使用其他条件请切换为 Filter')
    const lower = selector.object_size_greater_than, upper = selector.object_size_less_than
    for (const size of [lower, upper]) if (size !== null && size !== '' && (!/^[0-9]+$/.test(size) || BigInt(size) > 18446744073709551615n)) throw new Error('对象大小必须为 uint64 十进制字节数')
    if (lower && upper && (BigInt(upper) <= BigInt(lower) || upper <= lower)) throw new Error('对象大小上下界不符合原生范围校验')
    const conditions = (selector.prefix === null ? '' : element('Prefix', selector.prefix)) + selector.tags.map(tag => `<Tag>${element('Key', tag.key)}${element('Value', tag.value)}</Tag>`).join('') + (lower === null ? '' : element('ObjectSizeGreaterThan', lower)) + (upper === null ? '' : element('ObjectSizeLessThan', upper)) + (selector.archive_zone ? '<ArchiveZone/>' : '')
    const filter = selector.kind === 'Prefix' ? element('Prefix', selector.prefix!) : `<Filter>${selector.and ? `<And>${conditions}</And>` : conditions}</Filter>`
    const counts = new Map<string, number>(), classes = new Set<string>(), timings = new Set<string>()
    let effective = false
    const actions = rule.actions.map(action => {
      const fields = action.fields, keys = Object.keys(fields)
      counts.set(action.type, (counts.get(action.type) ?? 0) + 1)
      if (!action.type.endsWith('Transition') && counts.get(action.type)! > 1) throw new Error('单值生命周期动作不能重复')
      const has = (key: string) => own(fields, key)
      if (action.type === 'Expiration' && ['Days', 'Date', 'ExpiredObjectDeleteMarker'].filter(has).length !== 1) throw new Error('过期动作必须在 Days、Date 和删除标记中三选一')
      if (action.type === 'Transition' && (['Days', 'Date'].filter(has).length !== 1 || !has('StorageClass'))) throw new Error('转换需 Days/Date 二选一及 StorageClass')
      if (action.type.startsWith('Noncurrent') && !has('NoncurrentDays')) throw new Error('非当前版本动作需要 NoncurrentDays')
      if (action.type === 'NoncurrentVersionTransition' && !has('StorageClass')) throw new Error('非当前版本转换需要 StorageClass')
      if (action.type === 'AbortIncompleteMultipartUpload' && !has('DaysAfterInitiation')) throw new Error('终止上传需要 DaysAfterInitiation')
      for (const key of keys) {
        const item = fields[key]
        if (['Days', 'NoncurrentDays', 'DaysAfterInitiation', 'NewerNoncurrentVersions'].includes(key)) {
          const minimum = action.type.endsWith('Transition') || key === 'NewerNoncurrentVersions' ? 0 : 1
          if (!/^[0-9]+$/.test(item) || BigInt(item) < BigInt(minimum) || BigInt(item) > 2147483647n) throw new Error(`${key} 必须为 ${minimum}–2147483647 的十进制整数`)
        }
        if (key === 'Date' && !item) throw new Error('日期不能为空；需有效 UTC 午夜日期，后端会校验原生格式')
        if (key === 'ExpiredObjectDeleteMarker' && item !== 'true' && item !== 'false') throw new Error('删除标记需明确 true 或 false')
        if ((action.type === 'Expiration' || action.type === 'Transition') && ['Days', 'Date'].includes(key)) timings.add(key)
      }
      if (has('StorageClass')) {
        const identity = JSON.stringify([action.type, fields.StorageClass])
        if (classes.has(identity)) throw new Error('同类转换不能重复目标存储类别')
        classes.add(identity)
      }
      if (selector.tags.length && (action.type === 'AbortIncompleteMultipartUpload' || fields.ExpiredObjectDeleteMarker === 'true')) throw new Error('标签过滤不能与终止上传或清除过期删除标记组合')
      if (!(action.type === 'Expiration' && fields.ExpiredObjectDeleteMarker === 'false')) effective = true
      return `<${action.type}>${keys.map(key => element(key, fields[key])).join('')}</${action.type}>`
    }).join('')
    if (!effective || timings.size > 1) throw new Error('需要有效动作，当前版本过期与转换不能混用 Days 和 Date')
    return `<Rule>${element('ID', rule.id)}${element('Status', rule.status)}${filter}${actions}</Rule>`
  }).join('') + '</LifecycleConfiguration>'
  if (new TextEncoder().encode(JSON.stringify({ document })).length > 1024 * 1024 - 1024) throw new Error('生命周期配置超过请求大小限制')
  return document
}

export function lifecycleFormInitial(row?: Record<string, unknown>) {
  if (!row || row.kind !== 'lifecycle' || typeof row.bucket_id !== 'string' || !/^[A-Za-z0-9_-]+$/.test(row.bucket_id) || typeof row.configured !== 'boolean') throw new Error('生命周期身份或状态不可用，请刷新')
  const draft = lifecycleDraft({ rules: row.lifecycle_rules })
  if (row.configured) lifecycleDocument(draft)
  else if (draft.rules.length) throw new Error('生命周期状态与规则不一致')
  return { bucket_id: row.bucket_id, kind: 'lifecycle', lifecycle_draft: draft }
}
export function lifecycleFormBlocked(row: Record<string, unknown>) {
  try { lifecycleFormInitial(row); return undefined } catch (error) { return error instanceof Error ? error.message : '生命周期不可用' }
}
export function lifecycleFormInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const initial = lifecycleFormInitial(row)
  if (values.bucket_id !== initial.bucket_id || values.kind !== 'lifecycle') throw new Error('不能更改 Bucket 身份或配置类型')
  return { bucket_id: initial.bucket_id, kind: 'lifecycle', document: lifecycleDocument(values.lifecycle_draft) }
}
export function lifecycleFormConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input = lifecycleFormInput(values, row)
  return `确认整体替换 Bucket ID ${input.bucket_id} 的生命周期为 ${lifecycleDraft(values.lifecycle_draft).rules.length} 条规则？未保留的规则将移除。启用的过期规则可能永久删除对象或历史版本，转换和终止上传也可能影响业务。请先备份原文，外部并发修改可能被覆盖；提交后回读核验不代表对象处理已完成。`
}

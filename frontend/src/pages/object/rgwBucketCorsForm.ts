export type CorsRule = { id: string; allowed_origins: string[]; allowed_methods: string[]; allowed_headers: string[]; expose_headers: string[]; max_age_seconds: number | null }
export type CorsDraft = { rules: CorsRule[] }
export const corsMethods = ['GET', 'PUT', 'DELETE', 'HEAD', 'POST', 'COPY']

export function corsDraft(value: unknown): CorsDraft {
  const rules = (value as CorsDraft | undefined)?.rules
  if (!Array.isArray(rules) || rules.some(rule => !rule || typeof rule.id !== 'string'
    || ['allowed_origins', 'allowed_methods', 'allowed_headers', 'expose_headers'].some(key => !Array.isArray(rule[key as keyof CorsRule]) || (rule[key as keyof CorsRule] as unknown[]).some(item => typeof item !== 'string'))
    || (rule.max_age_seconds !== null && (!Number.isInteger(rule.max_age_seconds) || rule.max_age_seconds < 0 || rule.max_age_seconds > 4294967294)))) throw new Error('CORS 规则数据不可用')
  return { rules: rules.map(rule => ({ ...rule, allowed_origins: [...rule.allowed_origins], allowed_methods: [...rule.allowed_methods], allowed_headers: [...rule.allowed_headers], expose_headers: [...rule.expose_headers] })) }
}

function xml(value: string) {
  for (const character of value) {
    const code = character.codePointAt(0)!
    if (!(code === 9 || code === 10 || code === 13 || (code >= 32 && code <= 0xd7ff) || (code >= 0xe000 && code <= 0xfffd) || code >= 0x10000)) throw new Error('字段包含 XML 不支持的字符')
  }
  return value.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&apos;').replace(/\r/g, '&#13;')
}

export function corsDocument(value: unknown) {
  const { rules } = corsDraft(value)
  if (!rules.length) throw new Error('至少需要一条规则；关闭 CORS 请使用删除配置操作')
  const document = '<CORSConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/">' + rules.map(rule => {
    if (new TextEncoder().encode(rule.id).length > 255) throw new Error('规则 ID 最多 255 UTF-8 字节')
    if (!rule.allowed_origins.length || [...rule.allowed_origins, ...rule.allowed_headers].some(item => !item || (item.match(/\*/g)?.length ?? 0) > 1)) throw new Error('每条规则需要来源；来源和允许头不能为空且最多含一个 *')
    if (rule.allowed_methods.some(method => !corsMethods.includes(method))) throw new Error('不支持的 CORS 方法，请使用原始 XML 编辑')
    const fields = (tag: string, items: string[]) => items.map(item => `<${tag}>${xml(item)}</${tag}>`).join('')
    return `<CORSRule><ID>${xml(rule.id)}</ID>${fields('AllowedOrigin', rule.allowed_origins)}${fields('AllowedMethod', rule.allowed_methods)}${fields('AllowedHeader', rule.allowed_headers)}${fields('ExposeHeader', rule.expose_headers)}${rule.max_age_seconds === null ? '' : `<MaxAgeSeconds>${rule.max_age_seconds}</MaxAgeSeconds>`}</CORSRule>`
  }).join('') + '</CORSConfiguration>'
  if (new TextEncoder().encode(JSON.stringify({ document })).length > 1024 * 1024 - 1024) throw new Error('CORS 配置超过请求大小限制')
  return document
}

export function corsFormInitial(row?: Record<string, unknown>) {
  if (!row || row.kind !== 'cors' || typeof row.bucket_id !== 'string' || !/^[A-Za-z0-9_-]+$/.test(row.bucket_id) || typeof row.configured !== 'boolean') throw new Error('CORS 身份或状态不可用，请刷新')
  const draft = corsDraft({ rules: row.cors_rules })
  if (row.configured) corsDocument(draft)
  else if (draft.rules.length) throw new Error('CORS 状态与规则不一致')
  return { bucket_id: row.bucket_id, kind: 'cors', cors_draft: draft }
}
export function corsFormBlocked(row: Record<string, unknown>) {
  try { corsFormInitial(row); return undefined } catch (error) { return error instanceof Error ? error.message : 'CORS 不可用' }
}
export function corsFormInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const initial = corsFormInitial(row)
  if (values.bucket_id !== initial.bucket_id || values.kind !== 'cors') throw new Error('不能更改 Bucket 身份或配置类型')
  return { bucket_id: initial.bucket_id, kind: 'cors', document: corsDocument(values.cors_draft) }
}
export function corsFormConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input = corsFormInput(values, row)
  return `确认整体替换 Bucket ID ${input.bucket_id} 的 CORS 为 ${corsDraft(values.cors_draft).rules.length} 条规则？规则顺序会保留，未保留的规则将移除；可能影响浏览器跨域访问，CORS 不替代访问权限策略。请先备份原文，外部并发修改可能被覆盖，提交后回读核验。`
}

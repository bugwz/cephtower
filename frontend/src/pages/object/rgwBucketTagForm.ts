export type BucketTagEntry = { key: string; value: string }
export type BucketTagFormValue = { entries: BucketTagEntry[] }

export function bucketTagFormEntries(value: unknown): BucketTagEntry[] | undefined {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return undefined
  const entries = (value as BucketTagFormValue).entries
  if (!Array.isArray(entries) || entries.some(entry => !entry || typeof entry !== 'object' || typeof entry.key !== 'string' || typeof entry.value !== 'string')) return undefined
  return entries.map(entry => ({ key: entry.key, value: entry.value }))
}

function xmlText(value: string) {
  for (const character of value) {
    const code = character.codePointAt(0)!
    if (!(code === 9 || code === 10 || code === 13 || (code >= 0x20 && code <= 0xd7ff) || (code >= 0xe000 && code <= 0xfffd) || code >= 0x10000)) throw new Error('标签包含 XML 不支持的控制字符或无效 Unicode')
  }
  return value.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&apos;').replace(/\r/g, '&#13;')
}

export function bucketTagFormDocument(value: unknown) {
  const entries = bucketTagFormEntries(value)
  if (!entries) throw new Error('标签条目不可用，请重新打开表单')
  if (entries.length > 50) throw new Error('Bucket 最多支持 50 条标签')
  const encoder = new TextEncoder()
  const tags = entries.map((entry, index) => {
    if (!entry.key || encoder.encode(entry.key).length > 128) throw new Error(`第 ${index + 1} 条标签键必须为 1–128 UTF-8 字节`)
    if (encoder.encode(entry.value).length > 256) throw new Error(`第 ${index + 1} 条标签值最多为 256 UTF-8 字节`)
    return `<Tag><Key>${xmlText(entry.key)}</Key><Value>${xmlText(entry.value)}</Value></Tag>`
  })
  return `<Tagging xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><TagSet>${tags.join('')}</TagSet></Tagging>`
}

export function bucketTagFormInitial(row?: Record<string, unknown>) {
  if (!row || row.kind !== 'tagging' || typeof row.bucket_id !== 'string' || !/^[A-Za-z0-9_-]+$/.test(row.bucket_id)) throw new Error('请选择有效的 Bucket 标签记录')
  if (row.configured !== true && row.configured !== false) throw new Error('标签配置状态未知，请刷新后重试')
  const tag_set = { entries: row.tags }
  bucketTagFormDocument(tag_set)
  const entries = bucketTagFormEntries(tag_set)!
  if (!row.configured && entries.length !== 0) throw new Error('标签状态与条目不一致，请刷新后重试')
  return { bucket_id: row.bucket_id, kind: 'tagging', tag_set: { entries } }
}

export function bucketTagFormBlocked(row: Record<string, unknown>) {
  try { bucketTagFormInitial(row); return undefined } catch (error) { return error instanceof Error ? error.message : '标签不可用' }
}

export function bucketTagFormInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const initial = bucketTagFormInitial(row)
  if (values.bucket_id !== initial.bucket_id || values.kind !== 'tagging') throw new Error('不能更改标签记录的 Bucket ID 或配置类型')
  return { bucket_id: initial.bucket_id, kind: 'tagging', document: bucketTagFormDocument(values.tag_set) }
}

export function bucketTagFormConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const { bucket_id } = bucketTagFormInput(values, row)
  const count = bucketTagFormEntries(values.tag_set)!.length
  return `确认整体替换 Bucket ID ${bucket_id} 的全部标签为 ${count} 条？${count === 0 ? '将设置空标签集合，而不是删除标签属性。' : ''}未保留的条目将移除，可能影响依赖 Bucket 标签条件的访问权限；不修改对象标签。请先备份原文，外部并发修改可能被覆盖，提交不会自动回滚。`
}

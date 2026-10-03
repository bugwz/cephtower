function xmlText(value: string) {
  for (const character of value) {
    const code = character.codePointAt(0)!
    if (!(code === 9 || code === 10 || code === 13 || (code >= 0x20 && code <= 0xd7ff) || (code >= 0xe000 && code <= 0xfffd) || code >= 0x10000)) throw new Error('Key ID 包含 XML 不支持的字符')
  }
  return value.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&apos;').replace(/\r/g, '&#13;')
}

export function bucketEncryptionFormInitial(row?: Record<string, unknown>) {
  if (!row || row.kind !== 'encryption' || typeof row.bucket_id !== 'string' || !/^[A-Za-z0-9_-]+$/.test(row.bucket_id)) throw new Error('请选择有效的 Bucket 加密记录')
  const base = { bucket_id: row.bucket_id, kind: 'encryption' }
  if (row.configured === false && row.encryption === null) return { ...base, algorithm: undefined, kms_master_key_id: '', bucket_key_enabled: 'false' }
  if (row.configured !== true || !row.encryption || typeof row.encryption !== 'object') throw new Error('加密状态不可用，请刷新')
  const config = row.encryption as Record<string, unknown>
  if (typeof config.rule_exists !== 'boolean' || typeof config.algorithm !== 'string' || typeof config.kms_master_key_id !== 'string' || typeof config.bucket_key_enabled !== 'boolean') throw new Error('加密字段不可用，请刷新')
  if (config.algorithm && config.algorithm !== 'AES256' && config.algorithm !== 'aws:kms') throw new Error('当前算法不受表单支持，请使用原始 XML 编辑')
  return { ...base, algorithm: config.algorithm || undefined, kms_master_key_id: config.kms_master_key_id, bucket_key_enabled: String(config.bucket_key_enabled) }
}

export function bucketEncryptionFormBlocked(row: Record<string, unknown>) {
  try { bucketEncryptionFormInitial(row); return undefined } catch (error) { return error instanceof Error ? error.message : '加密数据不可用' }
}

export function bucketEncryptionFormInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const initial = bucketEncryptionFormInitial(row)
  if (values.bucket_id !== initial.bucket_id || values.kind !== initial.kind) throw new Error('不能更改 Bucket 身份或配置类型')
  if (values.algorithm !== 'AES256' && values.algorithm !== 'aws:kms') throw new Error('请选择加密算法')
  if (typeof values.kms_master_key_id !== 'string') throw new Error('Key ID 必须为文本')
  if (values.algorithm === 'aws:kms' && !values.kms_master_key_id.trim()) throw new Error('KMS 加密需要 Key ID')
  if (values.algorithm === 'AES256' && values.kms_master_key_id !== '') throw new Error('AES256 不使用 KMS Key ID，请明确清空该字段')
  if (values.bucket_key_enabled !== 'true' && values.bucket_key_enabled !== 'false') throw new Error('请选择 Bucket Key 状态')
  const key = values.algorithm === 'aws:kms' ? `<KMSMasterKeyID>${xmlText(values.kms_master_key_id)}</KMSMasterKeyID>` : ''
  const document = `<ServerSideEncryptionConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><Rule><ApplyServerSideEncryptionByDefault><SSEAlgorithm>${values.algorithm}</SSEAlgorithm>${key}</ApplyServerSideEncryptionByDefault><BucketKeyEnabled>${values.bucket_key_enabled}</BucketKeyEnabled></Rule></ServerSideEncryptionConfiguration>`
  if (new TextEncoder().encode(JSON.stringify({ ...initial, document })).length > 1024 * 1024 - 128) throw new Error('加密配置超出请求大小限制')
  return { bucket_id: initial.bucket_id, kind: 'encryption', document }
}

export function bucketEncryptionFormConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input = bucketEncryptionFormInput(values, row)
  return `确认整体替换 Bucket ID ${input.bucket_id} 的默认加密规则为 ${values.algorithm}？Bucket Key 将${values.bucket_key_enabled === 'true' ? '启用' : '不启用'}。请确认 RGW 加密服务及 KMS（如使用）已配置；保存核验不验证密钥可用性，不会重新加密已有对象。外部并发修改可能被覆盖，请先备份原文。关闭默认加密请使用删除配置操作。`
}

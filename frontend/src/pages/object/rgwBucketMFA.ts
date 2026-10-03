type Versioning = { status: string; mfa_delete: string | null }
export function bucketVersioningData(value: unknown): Versioning | undefined {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return undefined
  const data = value as Versioning
  if (!['', 'Enabled', 'Suspended'].includes(data.status) || (data.mfa_delete !== null && !['Enabled', 'Disabled'].includes(data.mfa_delete)) || data.status === '' && data.mfa_delete !== null) return undefined
  return data
}
export function bucketVersioningSummary(value: unknown, row: Record<string, unknown>) {
  if (row.kind !== 'versioning') return '—'
  const data = bucketVersioningData(value)
  if (row.configured !== true || !data) return '版本控制配置不可用'
  if (data.status === '') return '尚未启用版本控制；原生未返回 MFA Delete 字段'
  return `${data.status === 'Enabled' ? '版本控制已启用' : '版本控制已暂停（不是从未启用）'}；MFA Delete ${data.mfa_delete === null ? '未返回，状态未知' : data.mfa_delete === 'Enabled' ? '已启用' : '已停用'}`
}
export function bucketMFABlocked(row: Record<string, unknown>) {
  const data = bucketVersioningData(row.versioning)
  if (row.kind !== 'versioning' || typeof row.bucket_id !== 'string' || !/^[A-Za-z0-9_-]+$/.test(row.bucket_id) || row.configured !== true || typeof row.document !== 'string' || !row.document.trim() || !data || data.status !== '' && data.mfa_delete === null) return '请先读取完整版本控制及 MFA 配置'
  return undefined
}
export function bucketMFAInitial(row?: Record<string, unknown>) {
  if (!row || bucketMFABlocked(row)) throw new Error('版本控制快照不可用')
  const data = bucketVersioningData(row.versioning)!
  return { bucket_id: row.bucket_id as string, status: data.status || undefined, mfa_delete: data.mfa_delete || undefined, mfa_serial_secret: '', mfa_token: '', confirm_mfa: undefined }
}
export function bucketMFAInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const initial = bucketMFAInitial(row)
  if (values.bucket_id !== initial.bucket_id) throw new Error('Bucket ID 不可更改')
  if (!['Enabled', 'Suspended'].includes(String(values.status)) || !['Enabled', 'Disabled'].includes(String(values.mfa_delete))) throw new Error('请选择明确的版本控制及 MFA Delete 状态')
  if (values.status === initial.status && values.mfa_delete === initial.mfa_delete) throw new Error('状态没有变化')
  if (typeof values.mfa_serial_secret !== 'string' || !/^[!-~]{1,1024}$/.test(values.mfa_serial_secret) || typeof values.mfa_token !== 'string' || !/^[0-9]{1,128}$/.test(values.mfa_token)) throw new Error('请输入无空白 ASCII 设备序列号及纯数字验证码，保持前导零')
  if (values.confirm_mfa !== 'acknowledged') throw new Error('请确认 MFA 及版本控制影响')
  const input = { bucket_id: initial.bucket_id, status: values.status as string, mfa_delete: values.mfa_delete as string, mfa_serial_secret: values.mfa_serial_secret, mfa_token: values.mfa_token, expected_document: row!.document as string }
  if (new TextEncoder().encode(JSON.stringify(input)).length > 1024 * 1024 - 128) throw new Error('配置快照超出请求上限')
  return input
}
export function bucketMFAConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input = bucketMFAInput(values, row)
  return `确认将 Bucket ID ${input.bucket_id} 的版本控制设为 ${input.status}、MFA Delete 设为 ${input.mfa_delete}？需要 HTTPS 端点及与当前 S3 用户绑定的 MFA 设备；本操作不注册设备。验证码排队时可能过期，失败请先刷新配置再获取新码，不自动重试或回滚。暂停版本控制不会删除已有版本，对象锁可能禁止暂停；停用 MFA Delete 会撤除对应删除保护，启用后部分版本操作需要验证码。快照检查不是原子锁，回读只验证配置，不证明所有删除路径均受保护。`
}

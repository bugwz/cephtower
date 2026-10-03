export const bucketAclOptions = [
  { value: 'private', label: 'private：仅所有者完全控制（ACL 层面）' },
  { value: 'public-read', label: 'public-read：所有人（包括匿名）可读 Bucket' },
  { value: 'public-read-write', label: 'public-read-write：所有人（包括匿名）可读写 Bucket' },
  { value: 'authenticated-read', label: 'authenticated-read：所有已认证用户可读（不限本账户）' }
]

export function bucketAclFormBlocked(row: Record<string, unknown>) {
  if (row.kind !== 'acl' || row.configured !== true || typeof row.bucket_id !== 'string' || !/^[A-Za-z0-9_-]+$/.test(row.bucket_id)) return '请先读取目标 Bucket ACL'
  const acl = row.acl as { owner?: { id?: unknown }; grants?: unknown } | undefined
  if (!acl || typeof acl.owner?.id !== 'string' || !acl.owner.id || !Array.isArray(acl.grants)) return 'ACL 所有者或授权列表不可用'
  return undefined
}
export function bucketAclFormInitial(row?: Record<string, unknown>) {
  if (!row) throw new Error('请选择 ACL 记录')
  const blocked = bucketAclFormBlocked(row)
  if (blocked) throw new Error(blocked)
  // Do not infer a canned preset from a possibly custom grant set.
  return { bucket_id: row.bucket_id as string, acl: undefined, confirm_replace: undefined }
}
export function bucketAclFormInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const initial = bucketAclFormInitial(row)
  if (values.bucket_id !== initial.bucket_id) throw new Error('Bucket ID 不可更改')
  if (!bucketAclOptions.some(option => option.value === values.acl)) throw new Error('请明确选择 ACL 预设')
  if (values.confirm_replace !== 'acknowledged') throw new Error('请确认整体替换 ACL 的访问风险')
  return { bucket_id: initial.bucket_id, acl: values.acl as string }
}
export function bucketAclFormConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input = bucketAclFormInput(values, row)
  const label = bucketAclOptions.find(option => option.value === input.acl)!.label
  return `确认将 Bucket ID ${input.bucket_id} 的完整 ACL 替换为 ${label}？现有自定义授权将被移除，公开读写可能允许匿名访问、上传或删除对象。此操作不修改对象 ACL，private 也不保证最终私有，实际权限还受 Bucket Policy 等规则影响。请先备份 ACL 原文；并发修改可能被覆盖，写后无法核验不代表未生效，不会自动回滚。`
}

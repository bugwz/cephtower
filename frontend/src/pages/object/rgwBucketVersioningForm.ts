export function bucketVersioningInitial(row?: Record<string, unknown>) {
  return { versioning: row?.versioning === 'enabled' || row?.versioning === 'suspended' ? row.versioning : undefined }
}

export function bucketVersioningInput(values: Record<string, unknown>, bucketId: string) {
  if (!bucketId || !/^[A-Za-z0-9_-]+$/.test(bucketId)) throw new Error('Bucket 身份不可用，请刷新后重试')
  if (values.versioning !== 'enabled' && values.versioning !== 'suspended') throw new Error('请明确选择启用或暂停版本控制')
  return { bucket_id: bucketId, versioning: values.versioning }
}

export function bucketVersioningConfirmation(values: Record<string, unknown>, bucketId: string) {
  const input = bucketVersioningInput(values, bucketId)
  const impact = input.versioning === 'enabled'
    ? '后续对象写入将使用版本控制，可能增加存储占用；启用后只能暂停，不能恢复为从未启用状态。'
    : '暂停不是删除历史版本，也不是恢复为从未启用状态；后续写入不再生成新的非空版本 ID，可能覆盖 null 版本。'
  return `确认将 Bucket ID ${input.bucket_id} 的版本控制${input.versioning === 'enabled' ? '启用' : '暂停'}？${impact}对象锁或 MFA Delete 可能限制此操作；本操作不修改 MFA Delete，也不提交 MFA 凭据。外部并发修改可能被覆盖，提交后将回读核验状态。`
}

export function bucketDeleteInput(row: Record<string, unknown>) {
  const id = row.natural_key ?? row.bucket_id
  if (typeof id !== 'string' || !/^[A-Za-z0-9_-]+$/.test(id)) throw new Error('Bucket 编码身份不可用，请刷新库存后重试')
  return { bucket_id: id }
}

export function bucketDeleteBlocked(row: Record<string, unknown>) {
  try { bucketDeleteInput(row); return undefined } catch (error) { return error instanceof Error ? error.message : 'Bucket 身份不可用' }
}

export function bucketDeleteConfirmation(row: Record<string, unknown>) {
  const { bucket_id } = bucketDeleteInput(row)
  const name = typeof row.name === 'string' ? JSON.stringify(row.name) : '未返回'
  const tenant = typeof row.tenant === 'string' ? (row.tenant === '' ? '全局租户' : JSON.stringify(row.tenant)) : '未返回（以编码 ID 为准）'
  return `确认删除 Bucket ${name}，租户 ${tenant}，编码 ID ${bucket_id}？此操作发送原生 S3 Bucket 删除请求，不会清空对象、历史版本或删除标记，也不会绕过对象锁或保留限制；非空 Bucket 将由 RGW 拒绝。库存用量可能过期，不能据此确认已空。删除后无法通过本界面撤销；若请求失败，请刷新实际状态后再操作，不要重复提交。`
}

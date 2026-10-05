export const rgwUserOperationMaskOptions = [
  { label: '只读', value: 'read' },
  { label: '只写', value: 'write' },
  { label: '仅删除', value: 'delete' },
  { label: '读、写', value: 'read,write' },
  { label: '读、删除', value: 'read,delete' },
  { label: '写、删除', value: 'write,delete' },
  { label: '读、写、删除', value: 'read,write,delete' }
]

export function rgwUserOperationMaskInput(values: Record<string, unknown>) {
  const mask = values.op_mask
  if (typeof mask !== 'string' || !rgwUserOperationMaskOptions.some(option => option.value === mask)) throw new Error('请选择明确的用户操作掩码')
  return { op_mask: mask }
}

export function rgwUserOperationMaskInitial(row?: Record<string, unknown>) {
  if (!row || row.stale === true || typeof row.op_mask !== 'string') throw new Error('当前操作掩码未返回或库存过期，请重新采集')
  // mask_to_str emits comma-space separators and <none> for a zero mask.
  const option = rgwUserOperationMaskOptions.find(option => option.value.replace(/,/g, ', ') === row.op_mask)
  if (!option && row.op_mask !== '<none>') throw new Error('当前操作掩码格式未知，请核查原生配置')
  return { current_op_mask: row.op_mask, op_mask: option?.value }
}

export function rgwUserOperationMaskBlocked(row: Record<string, unknown>) {
  try { rgwUserOperationMaskInitial(row); return undefined } catch (error) { return error instanceof Error ? error.message : '当前操作掩码不可用' }
}

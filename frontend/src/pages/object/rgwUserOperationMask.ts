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

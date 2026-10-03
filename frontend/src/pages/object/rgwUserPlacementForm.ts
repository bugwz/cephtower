export function rgwUserPlacementInput(values: Record<string, unknown>) {
  const placement = values.default_placement
  const storage = values.default_storage_class ?? ''
  if (typeof placement !== 'string' || placement.trim() === '' || /[\0\r\n]/.test(placement)) throw new Error('请输入非空单行放置规则')
  if (typeof storage !== 'string' || /[\0\r\n]/.test(storage) || (storage !== '' && storage.trim() === '')) throw new Error('存储类必须是单行文本；留空使用原生默认类')
  return { default_placement: placement, default_storage_class: storage }
}

export function rgwUserPlacementInput(values: Record<string, unknown>) {
  const placement = values.default_placement
  const storage = values.default_storage_class ?? ''
  if (typeof placement !== 'string' || placement.trim() === '' || /[\0\r\n]/.test(placement)) throw new Error('请输入非空单行放置规则')
  if (typeof storage !== 'string' || /[\0\r\n]/.test(storage) || (storage !== '' && storage.trim() === '')) throw new Error('存储类必须是单行文本；留空使用原生默认类')
  return { default_placement: placement, default_storage_class: storage }
}

export function rgwUserPlacementTagsInput(values: Record<string, unknown>) {
  const tags = values.placement_tags_csv
  if (typeof tags !== 'string' || /[\0\r\n]/.test(tags) || tags.split(',').some(tag => tag.trim() === '')) {
    throw new Error('请输入非空逗号分隔标签，不允许空项；原生命令不支持清空')
  }
  return { placement_tags_csv: tags }
}

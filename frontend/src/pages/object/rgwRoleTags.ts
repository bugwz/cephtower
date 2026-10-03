export function rgwRoleTags(value: unknown) {
  if (!Array.isArray(value)) return undefined
  if (value.some(tag => !tag || typeof tag !== 'object' || Array.isArray(tag)
    || typeof tag.Key !== 'string' || typeof tag.Value !== 'string')) return undefined
  return value.map((tag, index) => ({ id: index, key: tag.Key as string, value: tag.Value as string }))
}

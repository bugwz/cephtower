export function rgwUserTags(value: unknown) {
  if (!Array.isArray(value)) return undefined
  if (value.some(tag => !tag || typeof tag !== 'object' || Array.isArray(tag)
    || typeof tag.key !== 'string' || typeof tag.val !== 'string')) return undefined
  return value.map((tag, index) => ({ id: index, key: tag.key as string, value: tag.val as string }))
}

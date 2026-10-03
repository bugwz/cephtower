export function rgwBucketTags(value: unknown) {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return undefined
  const entries = Object.entries(value)
  if (entries.some(([, value]) => typeof value !== 'string')) return undefined
  return entries.map(([key, value]) => ({ key, value: value as string }))
}

export function dependentFormFields(fields: Array<{ name: string; optionsDependencies?: string[] }>, changed: string[]) {
  const invalidated = new Set(changed)
  const result: string[] = []
  let expanded = true
  while (expanded) {
    expanded = false
    for (const field of fields) {
      if (!invalidated.has(field.name) && field.optionsDependencies?.some((name) => invalidated.has(name))) {
        invalidated.add(field.name)
        result.push(field.name)
        expanded = true
      }
    }
  }
  return result
}

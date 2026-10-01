import type { FieldColumn } from '../components/DataTable'

export function resourceFilterFields(columns: FieldColumn[]): string[] {
  return Array.from(new Set(columns.flatMap((column) => column.filterKey === false ? [] : [column.filterKey ?? column.key])))
}

export function resourceColumnFilters(column: FieldColumn, options: Record<string, string[]>, selected: Record<string, string[]>) {
  if (column.filterKey === false) return { key: column.key }
  const field = column.filterKey ?? column.key
  return {
    key: field,
    filterMultiple: true,
    filterSearch: true,
    filters: (options[field] ?? []).map((value) => ({ text: value, value })),
    filteredValue: selected[field] ?? null
  }
}

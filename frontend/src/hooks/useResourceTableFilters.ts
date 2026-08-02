import { useCallback, useEffect, useMemo, useState } from 'react'
import { listResourceFilterOptions } from '../api/resource'
import type { ApiRecord } from '../api/client'

export interface ResourceTableFilterOptions {
  path: string
  fields: Array<string | false | undefined>
  clusterId?: number
  body?: ApiRecord
}

export function useResourceTableFilters({ path, fields, clusterId, body }: ResourceTableFilterOptions) {
  // Callers pass inline arrays, so key the memo by field contents instead of array identity.
  const filterFieldsKey = JSON.stringify(
    Array.from(new Set(fields.filter((field): field is string => Boolean(field))))
  )
  const filterFields = useMemo(
    () => JSON.parse(filterFieldsKey) as string[],
    [filterFieldsKey]
  )
  const [filters, setFilters] = useState<Record<string, string[]>>({})
  const [filterOptions, setFilterOptions] = useState<Record<string, string[]>>({})

  useEffect(() => {
    setFilters({})
  }, [path, clusterId])

  useEffect(() => {
    let ignore = false
    async function loadOptions() {
      if (!clusterId || filterFields.length === 0) {
        setFilterOptions({})
        return
      }
      const options = await listResourceFilterOptions(path, filterFields, clusterId, body)
      if (!ignore) {
        setFilterOptions(options)
      }
    }
    void loadOptions().catch(() => {
      if (!ignore) {
        setFilterOptions({})
      }
    })
    return () => {
      ignore = true
    }
  }, [body, clusterId, filterFields, path])

  const handleFilterChange = useCallback((nextFilters: Record<string, string[]>) => {
    setFilters(nextFilters)
  }, [])

  return {
    filters,
    filterOptions,
    handleFilterChange
  }
}

export function mergeResourceFilters(...filters: Array<Record<string, string[]> | undefined>) {
  const merged: Record<string, string[]> = {}
  filters.forEach((filter) => {
    Object.entries(filter ?? {}).forEach(([field, values]) => {
      if (values.length > 0) {
        merged[field] = values
      }
    })
  })
  return merged
}

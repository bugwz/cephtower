import { Tag } from 'antd'
import type { ColumnsType, TableProps } from 'antd/es/table'
import type { ReactNode } from 'react'
import { textValue, type ApiRecord } from '../api/client'
import { formatDateTime, isDateTimeField } from '../utils/time'
import { AppTable } from './AppTable'

export interface FieldColumn {
  key: string
  title: string
  filterKey?: string | false
  render?: (value: unknown, row: ApiRecord) => React.ReactNode
}

interface DataTableProps {
  columns: FieldColumn[]
  data: ApiRecord[]
  footer?: ReactNode
  rowKeyCandidates?: string[]
  filterOptions?: Record<string, string[]>
  filteredValues?: Record<string, string[]>
  onFilterChange?: (filters: Record<string, string[]>) => void
}

export function DataTable({
  columns,
  data,
  footer,
  rowKeyCandidates = ['id', 'name', 'hostname'],
  filterOptions,
  filteredValues,
  onFilterChange
}: DataTableProps) {
  const tableColumns: ColumnsType<ApiRecord> = columns.map((column) => {
    const filterKey = column.filterKey ?? column.key
    const filterable = onFilterChange && filterKey !== false && column.key !== 'actions'
    const filterField = filterKey === false ? column.key : filterKey
    const options = filterOptions?.[filterField] ?? []
    return {
      title: column.title,
      dataIndex: column.key,
      key: filterField,
      ellipsis: true,
      ...(filterable
        ? {
            filterMultiple: true,
            filterSearch: true,
            filters: options.map((value: string) => ({ text: value, value })),
            filteredValue: filteredValues?.[filterField] ?? null
          }
        : {}),
      render: (value, row) => column.render?.(value, row) ?? renderValue(value, column.key)
    }
  })

  const handleChange: TableProps<ApiRecord>['onChange'] = (_pagination, filters) => {
    if (!onFilterChange) {
      return
    }
    onFilterChange(tableFilters(filters))
  }

  return (
    <AppTable<ApiRecord>
      columns={tableColumns}
      dataSource={data}
      onChange={handleChange}
      rowKey={(row, index) => rowKeyCandidates.map((key) => row[key]).find(Boolean)?.toString() ?? String(index)}
      footer={footer ? () => footer : undefined}
    />
  )
}

function renderValue(value: unknown, key: string) {
  if (isDateTimeField(key)) {
    return formatDateTime(value)
  }

  if (Array.isArray(value)) {
    return value.length ? value.map((item) => <Tag key={textValue(item)}>{textValue(item)}</Tag>) : '—'
  }

  return textValue(value)
}

function tableFilters(filters: Record<string, unknown>) {
  return Object.fromEntries(
    Object.entries(filters)
      .map(([field, values]) => [field, Array.isArray(values) ? values.map(String).filter(Boolean) : []] as const)
      .filter(([, values]) => values.length > 0)
  )
}

import { Tag } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import type { ReactNode } from 'react'
import { textValue, type ApiRecord } from '../api/client'
import { formatDateTime, isDateTimeField } from '../utils/time'
import { AppTable } from './AppTable'

export interface FieldColumn {
  key: string
  title: string
  render?: (value: unknown, row: ApiRecord) => React.ReactNode
}

interface DataTableProps {
  columns: FieldColumn[]
  data: ApiRecord[]
  footer?: ReactNode
  rowKeyCandidates?: string[]
}

export function DataTable({ columns, data, footer, rowKeyCandidates = ['id', 'name', 'hostname'] }: DataTableProps) {
  const tableColumns: ColumnsType<ApiRecord> = columns.map((column) => ({
    title: column.title,
    dataIndex: column.key,
    key: column.key,
    ellipsis: true,
    render: (value, row) => column.render?.(value, row) ?? renderValue(value, column.key)
  }))

  return (
    <AppTable<ApiRecord>
      columns={tableColumns}
      dataSource={data}
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

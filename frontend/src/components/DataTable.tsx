import { Empty, Table, Tag } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { textValue, type ApiRecord } from '../api/client'

export interface FieldColumn {
  key: string
  title: string
  render?: (value: unknown, row: ApiRecord) => React.ReactNode
}

interface DataTableProps {
  columns: FieldColumn[]
  data: ApiRecord[]
  rowKeyCandidates?: string[]
}

export function DataTable({ columns, data, rowKeyCandidates = ['id', 'name', 'hostname'] }: DataTableProps) {
  const tableColumns: ColumnsType<ApiRecord> = columns.map((column) => ({
    title: column.title,
    dataIndex: column.key,
    key: column.key,
    ellipsis: true,
    render: (value, row) => column.render?.(value, row) ?? renderValue(value)
  }))

  return (
    <Table
      size="middle"
      columns={tableColumns}
      dataSource={data}
      locale={{ emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据" /> }}
      pagination={{ pageSize: 8, showSizeChanger: false }}
      rowKey={(row, index) => rowKeyCandidates.map((key) => row[key]).find(Boolean)?.toString() ?? String(index)}
      scroll={{ x: true }}
    />
  )
}

function renderValue(value: unknown) {
  if (Array.isArray(value)) {
    return value.length ? value.map((item) => <Tag key={textValue(item)}>{textValue(item)}</Tag>) : '—'
  }

  return textValue(value)
}


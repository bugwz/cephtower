import { Empty, Table } from 'antd'
import type { TableProps } from 'antd'

export function AppTable<RecordType extends object = object>({
  pagination,
  locale,
  scroll,
  size = 'middle',
  ...props
}: TableProps<RecordType>) {
  return (
    <Table<RecordType>
      size={size}
      locale={{ emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据" />, ...locale }}
      pagination={pagination === false ? false : { defaultPageSize: 10, showSizeChanger: true, ...pagination }}
      scroll={{ x: true, ...scroll }}
      {...props}
    />
  )
}

import { Table } from 'antd'
import { rgwRoleTags } from './rgwRoleTags'

export function RgwRoleTagsTable({ value }: { value: unknown }) {
  const tags = rgwRoleTags(value)
  if (!tags) return <span>标签未返回或格式无效</span>
  return <Table size="small" rowKey="id" dataSource={tags} pagination={tags.length > 5 ? { pageSize: 5 } : false}
    locale={{ emptyText: '标签列表为空' }} columns={[
      { title: '标签键', dataIndex: 'key' },
      { title: '标签值', dataIndex: 'value', render: (value: string) => value === '' ? '（空字符串）' : value }
    ]} />
}

import { Table } from 'antd'
import { rgwBucketTags } from './rgwBucketTags'

export function RgwBucketTagsTable({ value }: { value: unknown }) {
  if (value === undefined) return <span>标签未返回（无标签属性或解码失败时可省略）</span>
  const tags = rgwBucketTags(value)
  if (!tags) return <span>标签格式无效</span>
  const display = (value: string) => value === '' ? '（空字符串）' : value
  return <Table size="small" rowKey="key" dataSource={tags} pagination={tags.length > 5 ? { pageSize: 5 } : false}
    locale={{ emptyText: '标签集合为空' }} columns={[
      { title: '标签键', dataIndex: 'key', render: display },
      { title: '标签值', dataIndex: 'value', render: display }
    ]} />
}

import { Table } from 'antd'

export function rgwBucketTagEntries(value: unknown) {
  if (!Array.isArray(value) || value.some(entry => !entry || typeof entry !== 'object' || typeof entry.key !== 'string' || typeof entry.value !== 'string')) return undefined
  return value.map((entry, index) => ({ index, key: entry.key as string, value: entry.value as string }))
}

export function RgwBucketTagEntries({ value }: { value: unknown }) {
  const tags = rgwBucketTagEntries(value)
  if (!tags) return <span>标签数据不可用</span>
  return <Table size="small" rowKey="index" dataSource={tags} pagination={tags.length > 5 ? { pageSize: 5 } : false}
    locale={{ emptyText: '标签集合为空（是否设置属性请查看配置状态）' }} columns={[
      { title: '键', dataIndex: 'key', render: (value: string) => <span style={{ whiteSpace: 'pre-wrap' }}>{value}</span> },
      { title: '值', dataIndex: 'value', render: (value: string) => <span style={{ whiteSpace: 'pre-wrap' }}>{value === '' ? '（空字符串）' : value}</span> }
    ]} />
}

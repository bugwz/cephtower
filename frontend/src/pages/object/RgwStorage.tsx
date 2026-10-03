import { Table } from 'antd'
import { rgwStorageRows } from './rgwStorageDetails'

export function RgwStorage({ value, categorized = false }: { value: unknown; categorized?: boolean }) {
  const rows = rgwStorageRows(value, categorized)
  if (!rows) return <span>容量与对象统计不可用</span>
  return <Table size="small" rowKey="category" dataSource={rows} pagination={rows.length > 5 ? { pageSize: 5 } : false}
    locale={{ emptyText: '本次采集未返回用量分类' }} columns={[
      { title: '统计分类', dataIndex: 'category' },
      { title: '逻辑容量（bytes）', dataIndex: 'size' },
      { title: '取整容量（bytes）', dataIndex: 'actual' },
      { title: '压缩/加密后容量（bytes）', dataIndex: 'utilized' },
      { title: '对象数', dataIndex: 'objects' }
    ]} />
}

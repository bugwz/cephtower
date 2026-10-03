import { Descriptions, Table } from 'antd'
import { rgwStorageRows, rgwStorageTimes } from './rgwStorageDetails'

export function RgwStorage({ value, categorized = false, account = false }: { value: unknown; categorized?: boolean; account?: boolean }) {
  const rows = rgwStorageRows(value, categorized)
  const times = rgwStorageTimes(value, account)
  if (!rows) return <span>{categorized
    ? 'Bucket 用量未返回或格式无效；原生命令仅在本地 Zonegroup 的普通索引可读取时返回分类用量，缺失不代表零用量'
    : '容量与对象统计不可用'}</span>
  return <><Table size="small" rowKey="category" dataSource={rows} pagination={rows.length > 5 ? { pageSize: 5 } : false}
    locale={{ emptyText: '本次采集未返回用量分类' }} columns={[
      { title: '统计分类', dataIndex: 'category' },
      { title: '逻辑容量（bytes）', dataIndex: 'size' },
      { title: '取整容量（bytes）', dataIndex: 'actual' },
      { title: '压缩/加密后容量（bytes）', dataIndex: 'utilized' },
      { title: '对象数', dataIndex: 'objects' }
    ]} />
    {!categorized && <Descriptions size="small" column={1} items={[
      { key: 'sync', label: '统计同步时间（命令原值）', children: times.synced },
      { key: 'update', label: '统计更新时间（命令原值）', children: times.updated }
    ]} />}
  </>
}

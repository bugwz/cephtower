import { Descriptions } from 'antd'
import type { ApiRecord } from '../../api/client'
import { rgwBucketIndexCount, rgwBucketIndexText } from './rgwBucketIndex'

export function RgwBucketIndexDetails({ row }: { row: ApiRecord }) {
  return <div>
    <Descriptions size="small" column={1} items={[
      { key: 'type', label: '索引类型（原值）', children: rgwBucketIndexText(row.index_type) },
      { key: 'generation', label: '索引代次', children: rgwBucketIndexCount(row.index_generation) },
      { key: 'shards', label: '索引分片数', children: rgwBucketIndexCount(row.num_shards) },
      { key: 'version', label: '索引版本', children: rgwBucketIndexText(row.ver) },
      { key: 'master', label: '主索引版本', children: rgwBucketIndexText(row.master_ver) },
      { key: 'marker', label: 'Bucket 标记', children: rgwBucketIndexText(row.marker) },
      { key: 'maximum', label: '最大索引标记', children: rgwBucketIndexText(row.max_marker) }
    ]} />
    <small>分片数及索引版本/最大标记仅在本地 Zonegroup 的普通索引可读取时返回；缺失不代表零。</small>
  </div>
}

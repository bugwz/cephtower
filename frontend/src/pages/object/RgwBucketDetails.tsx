import { Tabs } from 'antd'
import type { ReactNode } from 'react'
import type { ApiRecord } from '../../api/client'
import { RgwBucketIndexDetails } from './RgwBucketIndexDetails'
import { RgwBucketPlacementDetails } from './RgwBucketPlacementDetails'
import { RgwBucketTagsTable } from './RgwBucketTagsTable'
import { RgwQuota } from './RgwQuota'
import { RgwRateLimit } from './RgwRateLimit'
import { RgwStorage } from './RgwStorage'

export function RgwBucketDetails({ row, configuration }: { row: ApiRecord; configuration?: ReactNode }) {
  return <Tabs items={[
    { key: 'index', label: '索引详情', children: <RgwBucketIndexDetails row={row} /> },
    { key: 'placement', label: '显式存储池', children: <RgwBucketPlacementDetails value={row.explicit_placement} /> },
    { key: 'tags', label: 'Bucket 标签', children: <RgwBucketTagsTable value={row.tagset} /> },
    { key: 'quota', label: 'Bucket 配额', children: <RgwQuota value={row.bucket_quota} /> },
    { key: 'rate-limit', label: 'Bucket 限流（每 RGW）', children: <RgwRateLimit value={row.rate_limit} /> },
    { key: 'usage', label: '容量与对象统计', children: <RgwStorage value={row.usage} categorized /> },
    ...(configuration ? [{ key: 'configuration', label: '实时配置与操作', children: configuration }] : [])
  ]} />
}

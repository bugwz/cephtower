import { Tabs } from 'antd'
import type { ApiRecord } from '../../api/client'
import { RgwQuota } from './RgwQuota'
import { RgwStorage } from './RgwStorage'

export function RgwAccountDetails({ row }: { row: ApiRecord }) {
  return <Tabs items={[
    { key: 'quota', label: '账户总配额', children: <RgwQuota value={row.quota} /> },
    { key: 'bucket-quota', label: '默认 Bucket 配额', children: <RgwQuota value={row.bucket_quota} /> },
    { key: 'usage', label: '账户容量与对象统计', children: <RgwStorage value={row.storage_stats} account /> }
  ]} />
}

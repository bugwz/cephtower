import { Tabs } from 'antd'
import type { ApiRecord } from '../../api/client'
import { RgwUserIdentityDetails, RgwUserPlacementDetails } from './RgwUserIdentityDetails'
import { RgwPermissions } from './RgwPermissions'
import { RgwQuota } from './RgwQuota'
import { RgwRateLimit } from './RgwRateLimit'
import { RgwStorage } from './RgwStorage'
import { rgwStorageScope } from './rgwStorageDetails'

export function RgwUserDetails({ row }: { row: ApiRecord }) {
  return <Tabs items={[
    { key: 'identity', label: '身份与归属', children: <RgwUserIdentityDetails row={row} /> },
    { key: 'placement', label: '用户放置配置', children: <RgwUserPlacementDetails row={row} /> },
    { key: 'caps', label: '管理权限', children: <RgwPermissions value={row.caps} /> },
    { key: 'subusers', label: '子用户', children: <RgwPermissions value={row.subusers} subusers /> },
    { key: 'quota', label: '用户总配额', children: <RgwQuota value={row.user_quota} /> },
    { key: 'bucket-quota', label: '默认 Bucket 配额', children: <RgwQuota value={row.bucket_quota} /> },
    { key: 'rate-limit', label: '用户限流（每 RGW）', children: <RgwRateLimit value={row.rate_limit} /> },
    { key: 'usage', label: '容量与对象统计', children: <>
      <p>{rgwStorageScope(row.stats_scope, row.account_id)}</p>
      <RgwStorage value={row.storage_stats} />
    </> }
  ]} />
}

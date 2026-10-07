import { Alert, Tabs } from 'antd'
import { RgwQuota } from './RgwQuota'
import { RgwRateLimit } from './RgwRateLimit'

export function RgwPeriodConfig({ value }: { value: unknown }) {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    return <Alert type="warning" showIcon message="Period 配额与限流配置不可用（不推断未启用）" />
  }
  const config = value as Record<string, unknown>
  return <div>
    <Alert type="info" showIcon message="当前 Realm 的 Period 配置快照" description="不是默认 Realm 的本地全局配置，不代表用户或 Bucket 的最终有效配额，也不代表远端已同步。限流按每个 RGW 每分钟计，不是集群总吞吐量。" />
    <Tabs items={[
      { key: 'user_quota', label: '全局用户配额', children: <RgwQuota value={config.user_quota} /> },
      { key: 'bucket_quota', label: '全局 Bucket 配额', children: <RgwQuota value={config.bucket_quota} /> },
      { key: 'user_ratelimit', label: '全局用户限流', children: <RgwRateLimit value={config.user_ratelimit} /> },
      { key: 'bucket_ratelimit', label: '全局 Bucket 限流', children: <RgwRateLimit value={config.bucket_ratelimit} /> },
      { key: 'anonymous_ratelimit', label: '匿名请求限流', children: <RgwRateLimit value={config.anonymous_ratelimit} /> }
    ]} />
  </div>
}

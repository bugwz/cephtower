import { Alert, Tabs } from 'antd'
import { RgwRateLimit } from './RgwRateLimit'

export function RgwGlobalRateLimit({ value }: { value: unknown }) {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    return <Alert type="warning" showIcon message="全局限流信息不可用" />
  }
  const limits = value as Record<string, unknown>
  return <div>
    <Alert type="info" showIcon message="默认 Realm 配置；每 RGW 每分钟，不是集群总限额" description="这是采集到的全局配置，不代表特定用户或 Bucket 的最终有效限流。未返回的作用域不视为未启用。" />
    <Tabs items={[
      { key: 'user_ratelimit', label: '用户' },
      { key: 'bucket_ratelimit', label: 'Bucket' },
      { key: 'anonymous_ratelimit', label: '匿名请求' }
    ].map(({ key, label }) => ({ key, label, children: <RgwRateLimit value={limits[key]} /> }))} />
  </div>
}

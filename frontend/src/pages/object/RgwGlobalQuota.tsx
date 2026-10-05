import { Alert, Tabs } from 'antd'
import { RgwQuota } from './RgwQuota'

export function RgwGlobalQuota({ value }: { value: unknown }) {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    return <Alert type="warning" showIcon message="全局配额信息不可用" />
  }
  const quotas = value as Record<string, unknown>
  return <div>
    <Alert type="info" showIcon message="原生命令默认 Realm 范围的全局配额" description="未指定 Realm；无默认 Realm 时由 Ceph 使用空 Realm 配置。不是所有 Realm 的汇总，也不代表特定用户或 Bucket 的最终有效配额。" />
    <Tabs items={[
      { key: 'user_quota', label: '用户配额' },
      { key: 'bucket_quota', label: 'Bucket 配额' }
    ].map(({ key, label }) => ({ key, label, children: <RgwQuota value={quotas[key]} /> }))} />
  </div>
}

import { Alert, Button, Descriptions, Space, Spin } from 'antd'
import { useCallback } from 'react'
import { useResource } from '../../hooks'
import { rgwIdentityText } from './rgwUserIdentity'
import { loadRgwUserAccount } from './rgwUserAccountLookup'

export function RgwUserAccountDetails({ clusterId, accountId }: { clusterId?: number; accountId: unknown }) {
  if (accountId === '') return <Alert type="info" message="此用户未关联账户" />
  if (typeof accountId !== 'string' || !accountId || accountId !== accountId.trim()) return <Alert type="warning" message="账户 ID 未返回或格式无效，无法查询归属" />
  if (!clusterId || !Number.isSafeInteger(clusterId) || clusterId < 1) return <Alert type="info" message="请先选择集群" />
  return <AccountInventory key={JSON.stringify([clusterId, accountId])} clusterId={clusterId} accountId={accountId} />
}

export function AccountInventory({ clusterId, accountId }: { clusterId: number; accountId: string }) {
  const loader = useCallback(() => loadRgwUserAccount(clusterId, accountId), [clusterId, accountId])
  const { data, loading, error, refresh } = useResource(loader)
  const account = data?.account
  return <Space direction="vertical" style={{ width: '100%' }}>
    <p>关联依据为用户的账户 ID；以下来自 account get 的采集库存，并非实时查询。</p>
    <Button loading={loading} onClick={() => refresh()}>重新读取账户库存</Button>
    {loading && <Spin />}
    {error && <Alert type="error" message={error} description={data ? '下方为上次读取的库存，不能代表当前状态。' : undefined} />}
    {data?.stale && <Alert type="warning" message="账户库存过期或新鲜度未知，请重新采集后确认" />}
    {data && !account && <Alert type="info" message="账户库存中未找到此 ID，不能据此断定账户不存在" />}
    {account && <Descriptions size="small" column={1} items={[
      { key: 'id', label: '账户 ID', children: accountId },
      { key: 'name', label: '账户名称', children: rgwIdentityText(account.account_name, '空账户名称') },
      { key: 'tenant', label: '账户租户', children: rgwIdentityText(account.tenant, '默认租户') },
      { key: 'email', label: '账户邮箱', children: rgwIdentityText(account.email, '未设置邮箱') },
      { key: 'observed', label: '采集时间', children: rgwIdentityText(data?.observedAt, '未返回') }
    ]} />}
  </Space>
}

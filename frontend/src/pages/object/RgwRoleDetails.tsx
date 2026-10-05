import { Tabs } from 'antd'
import type { ApiRecord } from '../../api/client'
import { RgwPolicyDocument, RgwRoleManagedPolicies, RgwRolePolicyDetails } from './RgwRolePolicyDetails'
import { RgwRoleTagsTable } from './RgwRoleTagsTable'

export function RgwRoleDetails({ row }: { row: ApiRecord }) {
  return <div>
    <p>以下为采集时的角色配置，不能单独据此判断完整有效权限。</p>
    <Tabs items={[
      { key: 'trust', label: '信任策略', children: <RgwPolicyDocument value={row.AssumeRolePolicyDocument} /> },
      { key: 'inline', label: '内联权限策略', children: <RgwRolePolicyDetails value={row.PermissionPolicies} /> },
      { key: 'managed', label: '直接关联的托管策略', children: <RgwRoleManagedPolicies value={row.ManagedPermissionPolicies} /> },
      { key: 'tags', label: '角色标签', children: <RgwRoleTagsTable value={row.Tags} /> }
    ]} />
  </div>
}

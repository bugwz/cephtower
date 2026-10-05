import type { ApiRecord } from '../../api/client'
import { rgwRoleMutationBlocked, rgwRoleMutationIdentity } from './rgwRoleIdentity'

export function rgwRoleDeleteBlocked(row: ApiRecord) {
  return rgwRoleMutationBlocked(row)
}

export function rgwRoleDeleteInput(row: ApiRecord) {
  return rgwRoleMutationIdentity(row)
}

export function rgwRoleDeleteConfirmation(row: ApiRecord) {
  const input = rgwRoleDeleteInput(row)
  const scope = input.account_id ? `账户 ${JSON.stringify(input.account_id)}` : '租户作用域（角色名中的租户前缀会保留）'
  return `删除 ${scope} 的角色 ${JSON.stringify(input.name)}？此操作不可恢复，会删除角色配置及其内联策略和托管策略关联，不会删除托管策略本身。请先确认依赖此角色的应用；不能据此保证已签发的临时凭据立即失效。`
}

export function rgwRoleDeleteResourceKey(row: ApiRecord) {
  const input = rgwRoleDeleteInput(row)
  return `rgw/role/${input.account_id ? `${input.account_id}/` : ''}${input.name}`
}

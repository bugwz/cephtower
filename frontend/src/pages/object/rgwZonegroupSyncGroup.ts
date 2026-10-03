const record = (v: unknown): v is Record<string, unknown> => !!v && typeof v === 'object' && !Array.isArray(v)
const token = (v: unknown): v is string => typeof v === 'string' && !!v && !v.startsWith('-') && new TextEncoder().encode(v).length <= 512 && !/\p{Cc}/u.test(v) && ![...v].some(c => { const n = c.codePointAt(0)!; return n >= 0xd800 && n <= 0xdfff })
function snapshot(row?: Record<string, unknown>, allowEmpty = false) {
  if (!row || row.stale === true || !token(row.id) || !token(row.name) || typeof row.realm_id !== 'string' || (row.realm_id !== '' && !token(row.realm_id))) throw new Error('Zonegroup 身份或 Realm 归属不可用，请刷新')
  const groups = record(row.sync_policy) ? row.sync_policy.groups : undefined
  if (!Array.isArray(groups) || (!allowEmpty && !groups.length) || groups.some(g => !record(g) || !token(g.id) || typeof g.status !== 'string' || !g.status || !Array.isArray(g.pipes) || !record(g.data_flow)) || new Set(groups.map(g => g.id)).size !== groups.length) throw new Error('同步组数据不可用或为空')
  return { name: row.name, zonegroup_id: row.id, realm_id: row.realm_id, groups: groups as Record<string, unknown>[] }
}
export function zonegroupSyncInitial(row?: Record<string, unknown>) { const { groups: _, ...identity } = snapshot(row); return { ...identity, group_id: undefined, status: undefined, confirm_change: undefined } }
export function zonegroupSyncBlocked(row: Record<string, unknown>) { try { snapshot(row); return undefined } catch (error) { return (error as Error).message } }
export function zonegroupSyncDeleteInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const { groups, ...identity } = snapshot(row)
  if (values.name !== identity.name || values.zonegroup_id !== identity.zonegroup_id || values.realm_id !== identity.realm_id) throw new Error('Zonegroup 和 Realm 身份不可修改')
  const group = groups.find(g => g.id === values.group_id)
  if (!group) throw new Error('请输入准确的已有同步组 ID')
  if (values.confirm_delete !== 'acknowledged') throw new Error('请确认删除整组及发布风险')
  return { ...identity, group_id: group.id as string, expected_group: JSON.stringify(group) }
}
export function zonegroupSyncDeleteConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const p = zonegroupSyncDeleteInput(values,row)
  return `确认删除 Zonegroup ${JSON.stringify(p.name)}（ID ${JSON.stringify(p.zonegroup_id)}）中的整个同步组 ${JSON.stringify(p.group_id)}？组内全部数据流和管道将移除，删除 forbidden 组可能解除复制限制；不删除已有对象副本，不保证全部复制停止。${p.realm_id ? `随后提交 Realm ${JSON.stringify(p.realm_id)} 的 Period，可能同时发布其他待提交变更。` : '无 Realm，不提交 Period。'}请备份完整策略并避免外部或其他页面并发；步骤非事务，失败可能部分生效，不自动回滚或重试。`
}
export function zonegroupSyncCreateInitial(row?: Record<string, unknown>) { const { groups: _, ...identity } = snapshot(row,true); return { ...identity, group_id: '', status: undefined, confirm_create: undefined } }
export function zonegroupSyncCreateBlocked(row: Record<string, unknown>) { try { snapshot(row,true); return undefined } catch (error) { return (error as Error).message } }
export function zonegroupSyncCreateInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const { groups, ...identity } = snapshot(row,true)
  if (values.name !== identity.name || values.zonegroup_id !== identity.zonegroup_id || values.realm_id !== identity.realm_id) throw new Error('Zonegroup 和 Realm 身份不可修改')
  if (!token(values.group_id) || groups.some(g => g.id === values.group_id)) throw new Error('请输入合法且不存在的新组 ID')
  if (!['enabled','allowed','forbidden'].includes(values.status as string)) throw new Error('请明确选择初始状态')
  if (values.confirm_create !== 'acknowledged') throw new Error('请确认创建范围和发布风险')
  return { ...identity, group_id: values.group_id, status: values.status as string, expected_policy: JSON.stringify(row!.sync_policy) }
}
export function zonegroupSyncCreateConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const p = zonegroupSyncCreateInput(values,row)
  return `确认在 Zonegroup ${JSON.stringify(p.name)}（ID ${JSON.stringify(p.zonegroup_id)}）创建空同步组 ${JSON.stringify(p.group_id)}，初始状态 ${p.status}？无数据流或管道，不代表建立复制链路。${p.realm_id ? `随后提交 Realm ${JSON.stringify(p.realm_id)} 的 Period，可能发布该 Realm 的其他待提交变更。` : '无 Realm，不提交 Period。'}原生命令可覆盖同名组，请备份并避免外部并发。步骤非事务，失败可能部分生效，不自动回滚或重试；成功不代表远端同步完成。`
}
export function zonegroupSyncInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const { groups, ...identity } = snapshot(row)
  if (values.name !== identity.name || values.zonegroup_id !== identity.zonegroup_id || values.realm_id !== identity.realm_id) throw new Error('Zonegroup 和 Realm 身份不可修改')
  const group = groups.find(g => g.id === values.group_id)
  if (!group) throw new Error('请输入准确的已有同步组 ID')
  if (!['enabled','allowed','forbidden'].includes(values.status as string) || group.status === values.status) throw new Error('请选择不同的目标状态')
  if (values.confirm_change !== 'acknowledged') throw new Error('请确认策略修改与 Realm 发布范围')
  return { ...identity, group_id: group.id as string, expected_group: JSON.stringify(group), status: values.status as string }
}
export function zonegroupSyncConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const p = zonegroupSyncInput(values,row)
  return `确认修改 Zonegroup ${JSON.stringify(p.name)}（ID ${JSON.stringify(p.zonegroup_id)}）的组 ${JSON.stringify(p.group_id)} 为 ${p.status}？保留数据流和管道，可能改变所有匹配桶的复制行为，不删除已有对象副本。${p.realm_id ? `随后提交 Realm ${JSON.stringify(p.realm_id)} 的 Period，可能同时发布该 Realm 的其他待提交变更。` : '该 Zonegroup 无 Realm，不提交 Period。'}步骤非事务，失败可能部分生效，不自动回滚或重试。请备份并核对全部待发布变更，避免外部或其他页面并发；成功不代表远端同步完成。`
}

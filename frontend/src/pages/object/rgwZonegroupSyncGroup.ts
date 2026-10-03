const record = (v: unknown): v is Record<string, unknown> => !!v && typeof v === 'object' && !Array.isArray(v)
const token = (v: unknown): v is string => typeof v === 'string' && !!v && !v.startsWith('-') && new TextEncoder().encode(v).length <= 512 && !/\p{Cc}/u.test(v) && ![...v].some(c => { const n = c.codePointAt(0)!; return n >= 0xd800 && n <= 0xdfff })
function snapshot(row?: Record<string, unknown>) {
  if (!row || row.stale === true || !token(row.id) || !token(row.name) || typeof row.realm_id !== 'string' || (row.realm_id !== '' && !token(row.realm_id))) throw new Error('Zonegroup 身份或 Realm 归属不可用，请刷新')
  const groups = record(row.sync_policy) ? row.sync_policy.groups : undefined
  if (!Array.isArray(groups) || !groups.length || groups.some(g => !record(g) || !token(g.id) || typeof g.status !== 'string' || !g.status || !Array.isArray(g.pipes) || !record(g.data_flow)) || new Set(groups.map(g => g.id)).size !== groups.length) throw new Error('同步组数据不可用或为空')
  return { name: row.name, zonegroup_id: row.id, realm_id: row.realm_id, groups: groups as Record<string, unknown>[] }
}
export function zonegroupSyncInitial(row?: Record<string, unknown>) { const { groups: _, ...identity } = snapshot(row); return { ...identity, group_id: undefined, status: undefined, confirm_change: undefined } }
export function zonegroupSyncBlocked(row: Record<string, unknown>) { try { snapshot(row); return undefined } catch (error) { return (error as Error).message } }
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

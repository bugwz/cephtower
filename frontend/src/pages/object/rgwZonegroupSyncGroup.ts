import { syncPipeCreateFields, syncPipePriorityInput, syncPipeStorageClassInput, syncPipePrefixInput, syncPipePrefixWarning, syncPipeTagsInput, syncPipeTagsWarning, syncPipeACLInput, syncPipeACLWarning } from './rgwBucketSyncGroupForm'
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
function replicationPreparationSnapshot(row?: Record<string, unknown>) {
  const state = snapshot(row,true)
  if (!state.realm_id) throw new Error('需要明确的 Realm 归属以发布准备策略')
  if (state.groups.some(g => g.id === 'dashboard_admin_group')) throw new Error('dashboard_admin_group 已存在，请检查已有策略；此操作不会覆盖')
  const zones = row?.zones
  if (!Array.isArray(zones) || !zones.length || zones.some(z => !record(z) || !token(z.id) || /[\s,;=*]/u.test(z.id) || !token(z.name)) || new Set(zones.map(z => z.id)).size !== zones.length || new Set(zones.map(z => z.name)).size !== zones.length) throw new Error('当前 Zone 成员不可用或有歧义')
  return { ...state, zones: zones.map(z => z.id as string) }
}
export function zonegroupReplicationPrepareBlocked(row: Record<string, unknown>) { try { replicationPreparationSnapshot(row); return undefined } catch (error) { return (error as Error).message } }
export function zonegroupReplicationPrepareInitial(row?: Record<string, unknown>) {
  const { groups: _, zones: __, ...identity } = replicationPreparationSnapshot(row)
  return { ...identity, confirm_replication_prepare: undefined }
}
export function zonegroupReplicationPrepareInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const { groups: _, zones, ...identity } = replicationPreparationSnapshot(row)
  if (values.name !== identity.name || values.zonegroup_id !== identity.zonegroup_id || values.realm_id !== identity.realm_id) throw new Error('Zonegroup 和 Realm 身份不可修改')
  if (values.confirm_replication_prepare !== 'acknowledged') throw new Error('请确认全范围 allowed 策略及发布影响')
  return { ...identity, expected_policy: JSON.stringify(row!.sync_policy), expected_zones: zones }
}
export function zonegroupReplicationPrepareConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const p = zonegroupReplicationPrepareInput(values,row)
  const state = replicationPreparationSnapshot(row)
  return `确认在 Zonegroup ${JSON.stringify(p.name)}（${p.zonegroup_id}）准备桶复制上层策略？创建 allowed 组 dashboard_admin_group、覆盖当前 Zone IDs ${JSON.stringify(state.zones)} 的对称流 dashboard_admin_flow，以及全部 Zone/租户/桶/实例的 system 通配管道 dashboard_admin_pipe。只提供上层许可，可能让已有桶本地策略获得许可；不写 S3 规则，不代表桶复制已启用。随后提交 Realm ${JSON.stringify(p.realm_id)} 的 Period，可能发布其他待提交变更。新增 Zone 不会自动加入该对称流。请备份并避免外部或其他页面并发；分步非事务，失败可能部分生效，不自动回滚或重试。`
}
export function zonegroupPipeZonesInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const selected = zonegroupPipeDeleteInput({ ...values, confirm_pipe_delete: 'acknowledged' }, row)
  const available = row?.zones
  const validID = (v: unknown): v is string => token(v) && !/[\s,;=*]/u.test(v)
  if (!Array.isArray(available) || available.some(z => !record(z) || !validID(z.id)) || new Set(available.map(z => z.id)).size !== available.length) throw new Error('当前 Zone ID 列表不可用或有歧义')
  const zones = (side: string) => {
    let ids: unknown
    try { ids = JSON.parse(String(values[side + '_zones_json'])) } catch { throw new Error('Zone ID 必须为 JSON 数组') }
    if (!Array.isArray(ids) || !ids.length || new Set(ids).size !== ids.length || !(ids.length === 1 && ids[0] === '*') && ids.some(id => !validID(id) || !available.some(z => z.id === id))) throw new Error('最终集合必须非空、无重复，且为当前 Zone ID 或单独 ["*"]')
    return ids as string[]
  }
  const source = zones('source'), dest = zones('dest')
  if (values.confirm_pipe_zones !== 'acknowledged') throw new Error('请确认成员与发布范围')
  return { ...selected, source_zones: source, dest_zones: dest }
}
export function zonegroupPipeZonesConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const p = zonegroupPipeZonesInput(values,row)
  return `确认修改 Zonegroup ${JSON.stringify(p.name)}（${p.zonegroup_id}）中同步组 ${JSON.stringify(p.group_id)} 的管道 ${JSON.stringify(p.pipe_id)}？最终源 Zone IDs ${JSON.stringify(p.source_zones)}，目标 ${JSON.stringify(p.dest_zones)}。* 匹配全部 Zone；明确集合先增后删，通配直接切换，中间复制范围可能扩大。保留桶选择器、执行身份和高级参数。${p.realm_id ? `随后提交 Realm ${JSON.stringify(p.realm_id)} 的 Period，可能发布其他待提交变更。` : '无 Realm，不提交 Period。'}请备份并避免外部或其他页面并发；非事务，失败可能部分生效，不自动回滚或重试；成功不代表远端复制完成。`
}
export function zonegroupPipeUpdateInput(values: Record<string, unknown>, row?: Record<string, unknown>): Record<string, unknown> {
  const selected = zonegroupPipeDeleteInput({ ...values, confirm_pipe_delete: 'acknowledged' }, row)
  const group = JSON.parse(selected.expected_group)
  group.pipes = group.pipes.filter((p: { id: string }) => p.id !== selected.pipe_id)
  const input = syncPipeCreateFields({ ...values, source_zones_json: '["*"]', dest_zones_json: '["*"]', confirm_pipe_create: 'acknowledged' }, group)
  if (values.confirm_pipe_update !== 'acknowledged') throw new Error('请确认选择器、权限及发布影响')
  delete input.source_zones
  delete input.dest_zones
  return { ...input, ...selected, ...syncPipePriorityInput(values), ...syncPipeStorageClassInput(values), ...syncPipePrefixInput(values), ...syncPipeTagsInput(values), ...syncPipeACLInput(values) }
}
export function zonegroupPipeUpdateConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const p = zonegroupPipeUpdateInput(values,row)
  return `确认修改 Zonegroup ${JSON.stringify(p.name)}（${p.zonegroup_id}）中同步组 ${JSON.stringify(p.group_id)} 的管道 ${JSON.stringify(p.pipe_id)}？源租户/桶/实例 ${JSON.stringify([p.source_tenant,p.source_bucket,p.source_bucket_id])}；目标 ${JSON.stringify([p.dest_tenant,p.dest_bucket,p.dest_bucket_id])}；模式 ${p.mode}，用户 ${JSON.stringify(p.user)}。* 为通配，空租户不限租户；system 模式保留已存储 UID，不删除用户或凭据。优先级：${p.priority === undefined ? '保持原值' : p.priority}，可能改变匹配管道的选择。目标存储类：${p.storage_class === undefined ? '保持原值' : JSON.stringify(p.storage_class)}；空字符串仍是显式覆盖，不是移除字段。${syncPipePrefixWarning(p)}${syncPipeTagsWarning(p)}${syncPipeACLWarning(p)}保留 Zone 成员。不验证目标放置配置或已有对象迁移。${p.realm_id ? `随后提交 Realm ${JSON.stringify(p.realm_id)} 的 Period，可能发布其他待提交变更。` : '无 Realm，不提交 Period。'}可能改变复制范围或权限，请备份并避免外部或其他页面并发；非事务，失败可能部分生效，不自动回滚或重试；成功不代表远端复制完成。`
}
export function zonegroupPipeDeleteInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const { groups, ...identity } = snapshot(row)
  if (values.name !== identity.name || values.zonegroup_id !== identity.zonegroup_id || values.realm_id !== identity.realm_id) throw new Error('Zonegroup 和 Realm 身份不可修改')
  const group = groups.find(g => g.id === values.group_id)
  const pipes = group?.pipes
  if (!group || !Array.isArray(pipes) || pipes.some(p => !record(p) || typeof p.id !== 'string') || !token(values.pipe_id) || pipes.filter(p => p.id === values.pipe_id).length !== 1) throw new Error('请输入当前策略中唯一的完整管道 ID')
  if (values.confirm_pipe_delete !== 'acknowledged') throw new Error('请确认整个管道删除与发布风险')
  return { ...identity, group_id: group.id as string, pipe_id: values.pipe_id, expected_group: JSON.stringify(group) }
}
export function zonegroupPipeDeleteConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const p = zonegroupPipeDeleteInput(values,row)
  return `确认删除 Zonegroup ${JSON.stringify(p.name)}（${p.zonegroup_id}）中同步组 ${JSON.stringify(p.group_id)} 的整个管道 ${JSON.stringify(p.pipe_id)}？全部源/目标选择器、权限模式和高级参数将移除；保留组状态和流，不删除已有对象副本，也不保证其他策略的复制停止。${p.realm_id ? `随后提交 Realm ${JSON.stringify(p.realm_id)} 的 Period，可能发布其他待提交变更。` : '无 Realm，不提交 Period。'}请备份并避免外部或其他页面并发；非事务，失败可能部分生效，不自动回滚或重试。`
}
export function zonegroupPipeCreateInput(values: Record<string, unknown>, row?: Record<string, unknown>): Record<string, unknown> {
  const { groups, ...identity } = snapshot(row)
  if (values.name !== identity.name || values.zonegroup_id !== identity.zonegroup_id || values.realm_id !== identity.realm_id) throw new Error('Zonegroup 和 Realm 身份不可修改')
  const group = groups.find(g => g.id === values.group_id)
  if (!group) throw new Error('请输入准确的已有同步组 ID')
  const input = syncPipeCreateFields(values, group as { id: string; status: string })
  for (const side of ['source','dest']) {
    const ids = input[side + '_zones'] as string[]
    if (ids.length === 1 && ids[0] === '*') continue
    const available = row?.zones
    if (!Array.isArray(available) || available.some(z => !record(z) || !token(z.id) || !token(z.name)) || new Set(available.map(z => z.id)).size !== available.length || new Set(available.map(z => z.name)).size !== available.length || ids.some(id => !available.some(z => z.id === id))) throw new Error('明确的 Zone ID 必须属于当前 Zonegroup，且成员列表不能有歧义')
  }
  return { ...identity, ...input }
}
export function zonegroupPipeCreateConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const p = zonegroupPipeCreateInput(values,row)
  return `确认在 Zonegroup ${JSON.stringify(p.name)}（${p.zonegroup_id}）的组 ${JSON.stringify(p.group_id)} 创建管道 ${JSON.stringify(p.pipe_id)}？源 Zone IDs ${JSON.stringify(p.source_zones)}，租户/桶/实例 ${JSON.stringify([p.source_tenant,p.source_bucket,p.source_bucket_id])}；目标 ${JSON.stringify(p.dest_zones)}，${JSON.stringify([p.dest_tenant,p.dest_bucket,p.dest_bucket_id])}。模式 ${p.mode}，用户 ${JSON.stringify(p.user)}。* 为通配，空租户不限租户，不代表仅全局租户。优先级 0，无前缀/标签过滤或目标 ACL/存储类覆盖；可能改变复制范围，不更改组状态和流。${p.realm_id ? `随后提交 Realm ${JSON.stringify(p.realm_id)} 的 Period，可能发布其他待提交变更。` : '无 Realm，不提交 Period。'}请备份并避免外部或其他页面并发；非事务，失败可能部分生效，不自动回滚或重试；成功不代表远端复制完成。`
}
export function zonegroupFlowUpdateInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const { groups, ...identity } = snapshot(row)
  if (values.name !== identity.name || values.zonegroup_id !== identity.zonegroup_id || values.realm_id !== identity.realm_id) throw new Error('Zonegroup 和 Realm 身份不可修改')
  const group = groups.find(g => g.id === values.group_id)
  const entries = group && (group.data_flow as Record<string, unknown>).symmetrical
  if (!Array.isArray(entries) || entries.some(e => !record(e)) || !token(values.flow_id)) throw new Error('对称流数据不可用')
  const matches = entries.filter(e => e.id === values.flow_id)
  if (matches.length !== 1) throw new Error('对称流不存在或有歧义')
  const validID = (v: unknown): v is string => token(v) && !/[,;=*\s]/u.test(v)
  const old = matches[0].zones
  if (!Array.isArray(old) || !old.length || !old.every(validID) || new Set(old).size !== old.length) throw new Error('原成员集合不可用')
  const available = row?.zones
  if (!Array.isArray(available) || !available.length || available.some(z => !record(z) || !token(z.id) || !token(z.name)) || new Set(available.map(z => z.id)).size !== available.length || new Set(available.map(z => z.name)).size !== available.length) throw new Error('Zone 列表不可用或有歧义')
  const zones = String(values.zones ?? '').split(',').map(z => z.trim())
  if (!zones.every(z => validID(z) && available.some(a => a.id === z)) || new Set(zones).size !== zones.length) throw new Error('最终成员必须是当前 Zonegroup 内非空、不重复的 Zone ID 集合')
  if (zones.length === old.length && zones.every(z => old.includes(z))) throw new Error('成员集合未改变')
  if (values.confirm_flow_update !== 'acknowledged') throw new Error('请确认分步修改与发布风险')
  return { ...identity, group_id: group!.id as string, flow_id: values.flow_id, zones, expected_group: JSON.stringify(group) }
}
export function zonegroupFlowUpdateConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const p = zonegroupFlowUpdateInput(values,row)
  return `确认修改 Zonegroup ${JSON.stringify(p.name)}（${p.zonegroup_id}）中同步组 ${JSON.stringify(p.group_id)} 的对称流 ${JSON.stringify(p.flow_id)}，最终 Zone ID 为 ${JSON.stringify(p.zones)}？先添加后移除，中间成员集合可能扩大；保留组状态和管道。${p.realm_id ? `随后提交 Realm ${JSON.stringify(p.realm_id)} 的 Period，可能发布其他待提交变更。` : '无 Realm，不提交 Period。'}请备份并避免外部或其他页面并发；非事务，失败可能部分生效，不自动回滚或重试；不删除对象副本，成功不代表远端复制完成。`
}
export function zonegroupFlowDeleteInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const { groups, ...identity } = snapshot(row)
  if (values.name !== identity.name || values.zonegroup_id !== identity.zonegroup_id || values.realm_id !== identity.realm_id) throw new Error('Zonegroup 和 Realm 身份不可修改')
  const group = groups.find(g => g.id === values.group_id)
  if (!group) throw new Error('请输入准确的已有同步组 ID')
  if (values.confirm_flow_delete !== 'acknowledged') throw new Error('请确认删除与发布风险')
  const kind = values.flow_type
  if (kind !== 'symmetrical' && kind !== 'directional') throw new Error('请选择数据流类型')
  const entries = (group.data_flow as Record<string, unknown>)[kind]
  if (!Array.isArray(entries) || entries.some(e => !record(e))) throw new Error('已有数据流不可用')
  if (values.zones) throw new Error('整流删除不接受 Zone 成员列表')
  const base = { ...identity, group_id: group.id as string, expected_group: JSON.stringify(group), flow_type: kind }
  if (kind === 'symmetrical') {
    if (!token(values.flow_id) || entries.filter(e => e.id === values.flow_id).length !== 1) throw new Error('对称流不存在或有歧义')
    if (values.source_zone || values.dest_zone) throw new Error('对称流不接受定向字段')
    return { ...base, flow_id: values.flow_id }
  }
  const validZone = (v: unknown): v is string => token(v) && !/[,;=*\s]/u.test(v)
  if (values.flow_id || !validZone(values.source_zone) || !validZone(values.dest_zone) || values.source_zone === values.dest_zone || entries.filter(e => e.source_zone === values.source_zone && e.dest_zone === values.dest_zone).length !== 1) throw new Error('请输入策略内唯一的源/目标 Zone ID 对，不填写流 ID')
  return { ...base, source_zone: values.source_zone, dest_zone: values.dest_zone }
}
export function zonegroupFlowDeleteConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const p = zonegroupFlowDeleteInput(values,row)
  const target = 'flow_id' in p ? `整条对称流 ${JSON.stringify(p.flow_id)}（全部成员）` : `定向流 ${JSON.stringify(p.source_zone)} → ${JSON.stringify(p.dest_zone)}`
  return `确认删除 Zonegroup ${JSON.stringify(p.name)}（${p.zonegroup_id}）中同步组 ${JSON.stringify(p.group_id)} 的${target}？保留组状态和管道，不删除对象副本，不保证其他策略的复制停止。${p.realm_id ? `随后提交 Realm ${JSON.stringify(p.realm_id)} 的 Period，可能发布其他待提交变更。` : '无 Realm，不提交 Period。'}请备份并避免外部或其他页面并发；非事务，失败可能部分生效，不自动回滚或重试。`
}
export function zonegroupFlowCreateInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const { groups, ...identity } = snapshot(row)
  if (values.name !== identity.name || values.zonegroup_id !== identity.zonegroup_id || values.realm_id !== identity.realm_id) throw new Error('Zonegroup 和 Realm 身份不可修改')
  const group = groups.find(g => g.id === values.group_id)
  if (!group) throw new Error('请输入准确的已有同步组 ID')
  if (values.confirm_flow !== 'acknowledged') throw new Error('请确认数据流创建与发布风险')
  const available = row?.zones
  if (!Array.isArray(available) || !available.length || available.some(z => !record(z) || !token(z.id) || !token(z.name)) || new Set(available.map(z => z.id)).size !== available.length || new Set(available.map(z => z.name)).size !== available.length) throw new Error('Zone 列表不可用或有歧义')
  const validZone = (v: unknown): v is string => token(v) && !/[,;=*\s]/u.test(v) && available.some(z => z.id === v)
  const kind = values.flow_type
  if (kind !== 'symmetrical' && kind !== 'directional') throw new Error('请选择数据流类型')
  const data = group.data_flow as Record<string, unknown>
  const existing = data[kind] === undefined ? [] : data[kind]
  if (!Array.isArray(existing) || existing.some(e => !record(e))) throw new Error('已有数据流不可用')
  const base = { ...identity, group_id: group.id as string, expected_group: JSON.stringify(group), flow_type: kind }
  if (kind === 'symmetrical') {
    if (!token(values.flow_id) || existing.some(e => e.id === values.flow_id)) throw new Error('请输入不存在的对称流 ID')
    if (values.source_zone || values.dest_zone) throw new Error('对称流不接受定向字段')
    const zones = String(values.zones ?? '').split(',').map(z => z.trim())
    if (!zones.every(validZone) || new Set(zones).size !== zones.length) throw new Error('请填写当前 Zonegroup 内不重复的 Zone ID，逗号分隔')
    return { ...base, flow_id: values.flow_id, zones }
  }
  if (values.flow_id || values.zones) throw new Error('定向流不接受流 ID 或对称 Zone 列表')
  if (!validZone(values.source_zone) || !validZone(values.dest_zone) || values.source_zone === values.dest_zone || existing.some(e => e.source_zone === values.source_zone && e.dest_zone === values.dest_zone)) throw new Error('请选择不同且尚未配置的源、目标 Zone ID')
  return { ...base, source_zone: values.source_zone, dest_zone: values.dest_zone }
}
export function zonegroupFlowCreateConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const p = zonegroupFlowCreateInput(values,row)
  return `确认在 Zonegroup ${JSON.stringify(p.name)}（${p.zonegroup_id}）的同步组 ${JSON.stringify(p.group_id)} 创建 ${p.flow_type} 数据流？可能改变匹配管道的复制路径，不更改组状态或管道。${p.realm_id ? `随后提交 Realm ${JSON.stringify(p.realm_id)} 的 Period，可能发布其他待提交变更。` : '无 Realm，不提交 Period。'}请备份并避免外部或其他页面并发；非事务，失败可能部分生效，不自动回滚或重试；成功不代表远端复制完成。`
}
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

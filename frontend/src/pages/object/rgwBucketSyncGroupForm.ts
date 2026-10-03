function groups(row?: Record<string, unknown>, allowEmpty = false) {
  const policy = row?.bucket_sync_policy as { groups?: unknown } | undefined
  if (!row || row.stale === true || typeof row.natural_key !== 'string' || !/^[A-Za-z0-9_-]+$/.test(row.natural_key) || !policy || !Array.isArray(policy.groups) || (!allowEmpty && policy.groups.length === 0)) throw new Error('请刷新并选择同步策略可用的 Bucket')
  const seen = new Set<string>()
  for (const group of policy.groups) {
    if (!group || typeof group !== 'object' || typeof group.id !== 'string' || !group.id || typeof group.status !== 'string' || !group.status || seen.has(group.id)) throw new Error('同步组数据不可用')
    seen.add(group.id)
  }
  return policy.groups as { id: string; status: string }[]
}
export function bucketSyncGroupBlocked(row: Record<string, unknown>) {
  try { groups(row); return undefined } catch (error) { return (error as Error).message }
}
export function bucketSyncGroupInitial(row?: Record<string, unknown>) {
  groups(row)
  return { bucket_id: row!.natural_key as string, group_id: undefined, status: undefined, confirm_change: undefined }
}
export function bucketSyncGroupInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const available = groups(row)
  if (values.bucket_id !== row!.natural_key) throw new Error('Bucket ID 不可更改')
  const group = available.find(group => group.id === values.group_id)
  if (!group) throw new Error('请输入当前策略中准确的同步组 ID')
  if (!['enabled', 'allowed', 'forbidden'].includes(values.status as string)) throw new Error('请明确选择目标状态')
  if (values.status === group.status) throw new Error('同步组状态未变化')
  if (values.confirm_change !== 'acknowledged') throw new Error('请确认复制影响')
  return { bucket_id: row!.natural_key as string, group_id: group.id, expected_status: group.status, status: values.status as string }
}
export function bucketSyncGroupConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input = bucketSyncGroupInput(values, row)
  return `确认修改 Bucket ID ${input.bucket_id} 的同步组 ${JSON.stringify(input.group_id)}：${input.expected_status} → ${input.status}？enabled 启用、allowed 仅允许但不启用、forbidden 禁止；可能改变后续复制行为。不会新建数据流或管道、不修改 Zonegroup 策略或提交 period。最终同步效果仍取决于上层策略和拓扑；已有副本不会删除。请备份策略，并避免外部并发修改；核验失败不代表未生效，操作不会自动回滚。`
}

export function bucketSyncGroupDeleteInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const available = groups(row)
  if (values.bucket_id !== row!.natural_key) throw new Error('Bucket ID 不可更改')
  const group = available.find(group => group.id === values.group_id)
  if (!group) throw new Error('请输入当前策略中准确的同步组 ID')
  if (values.confirm_delete !== 'acknowledged') throw new Error('请确认删除整个组及其数据流和管道')
  return { bucket_id: row!.natural_key as string, group_id: group.id, expected_group: JSON.stringify(group) }
}
export function bucketSyncGroupDeleteConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input = bucketSyncGroupDeleteInput(values, row)
  return `确认删除 Bucket ID ${input.bucket_id} 的同步组 ${JSON.stringify(input.group_id)}？整个组及其全部数据流、管道将被移除，可能改变复制行为（包括移除 forbidden 限制）；不删除已有对象副本，不保证所有复制停止。不修改 Zonegroup 或提交 period。请先备份策略并避免外部并发修改，核验失败不代表未生效，不自动回滚。`
}

export function bucketSyncFlowInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const group = groups(row).find(group => group.id === values.group_id)
  if (values.bucket_id !== row!.natural_key) throw new Error('Bucket ID 不可更改')
  if (!group) throw new Error('请输入当前策略中准确的同步组 ID')
  if (values.confirm_flow !== 'acknowledged') throw new Error('请确认数据流影响')
  const token = (v: unknown): v is string => typeof v === 'string' && !!v && !v.startsWith('-') && new TextEncoder().encode(v).length <= 512 && !/\p{Cc}/u.test(v) && ![...v].some(c => { const n = c.codePointAt(0)!; return n >= 0xd800 && n <= 0xdfff })
  const zone = (v: unknown): v is string => token(v) && !/[\s,;=*]/u.test(v)
  const flow = (group as unknown as { data_flow: Record<string, unknown> }).data_flow
  if (!flow || typeof flow !== 'object' || Array.isArray(flow)) throw new Error('数据流配置不可用')
  const kind = values.flow_type
  if (kind !== 'symmetrical' && kind !== 'directional') throw new Error('请选择数据流类型')
  const existing = flow[kind] === undefined ? [] : flow[kind]
  if (!Array.isArray(existing) || existing.some(entry => !entry || typeof entry !== 'object')) throw new Error('数据流配置不可用')
  const base = { bucket_id: row!.natural_key as string, group_id: group.id, expected_group: JSON.stringify(group), flow_type: kind }
  if (kind === 'symmetrical') {
    if (!token(values.flow_id)) throw new Error('请输入合法的对称流 ID')
    if (values.source_zone || values.dest_zone) throw new Error('对称流请清空源/目标 Zone 字段')
    let zones: unknown
    try { zones = JSON.parse(String(values.zones_json)) } catch { throw new Error('Zone ID 列表必须为 JSON 数组') }
    if (!Array.isArray(zones) || !zones.length || zones.some(v => !zone(v)) || new Set(zones).size !== zones.length) throw new Error('请输入非空、无重复的 Zone ID 数组，不支持通配符')
    if (existing.some(entry => entry.id === values.flow_id)) throw new Error('该对称流已存在')
    return { ...base, flow_id: values.flow_id, zones }
  }
  if (!zone(values.source_zone) || !zone(values.dest_zone) || values.source_zone === values.dest_zone) throw new Error('请输入不同的源/目标 Zone ID')
  if (values.flow_id || values.zones_json) throw new Error('定向流请清空对称流 ID 与 Zone 列表')
  if (existing.some(entry => entry.source_zone === values.source_zone && entry.dest_zone === values.dest_zone)) throw new Error('该定向流已存在')
  return { ...base, source_zone: values.source_zone, dest_zone: values.dest_zone }
}
export function bucketSyncFlowConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input = bucketSyncFlowInput(values, row)
  const target = 'zones' in input ? `ID ${JSON.stringify(input.flow_id)}，Zone IDs ${JSON.stringify(input.zones)}` : `${JSON.stringify(input.source_zone)} → ${JSON.stringify(input.dest_zone)}`
  return `确认在 Bucket ID ${input.bucket_id} 的同步组 ${JSON.stringify(input.group_id)} 创建 ${input.flow_type} 数据流：${target}？可能影响现有管道的复制行为。不创建管道、不修改组状态、Zonegroup 或提交 period；不代表已建立有效复制链路或同步完成。请备份策略并避免外部并发；核验失败不代表未生效，不自动回滚。`
}

export function bucketSyncFlowUpdateInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const selected = bucketSyncFlowDeleteInput({ ...values, flow_type: 'symmetrical', confirm_flow_delete: 'acknowledged' }, row)
  const group = JSON.parse(selected.expected_group)
  const flow = group.data_flow.symmetrical.find((entry: { id: string }) => entry.id === values.flow_id)
  if (!Array.isArray(flow.zones) || flow.zones.some((zone: unknown) => typeof zone !== 'string')) throw new Error('已有 Zone 列表不可用')
  // Reuse creation validation on a copy without the selected flow; preserve the real snapshot below.
  group.data_flow.symmetrical = group.data_flow.symmetrical.filter((entry: { id: string }) => entry.id !== values.flow_id)
  const input = bucketSyncFlowInput({ ...values, flow_type: 'symmetrical', confirm_flow: 'acknowledged' }, { ...row, bucket_sync_policy: { groups: [group] } })
  if (values.confirm_flow_update !== 'acknowledged') throw new Error('请确认非事务分步修改及部分生效风险')
  if (!('zones' in input)) throw new Error('仅支持编辑对称流')
  return { bucket_id: selected.bucket_id, group_id: selected.group_id, flow_id: values.flow_id as string, expected_group: selected.expected_group, zones: input.zones }
}
export function bucketSyncFlowUpdateConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input = bucketSyncFlowUpdateInput(values, row)
  return `确认修改 Bucket ID ${input.bucket_id} 的组 ${JSON.stringify(input.group_id)} 中对称流 ${JSON.stringify(input.flow_id)}，完整目标 Zone ID 列表为 ${JSON.stringify(input.zones)}？先添加再移除，非事务操作，中间并集可能临时扩大复制范围；失败可能部分生效，不自动回滚或重试。空列表请使用删除流操作。不修改管道、组状态、Zonegroup 或 period。请备份策略并避免外部并发；核验成功不代表同步完成。`
}

export function bucketSyncFlowDeleteInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const group = groups(row).find(group => group.id === values.group_id)
  if (values.bucket_id !== row!.natural_key) throw new Error('Bucket ID 不可更改')
  if (!group) throw new Error('请输入当前策略中准确的同步组 ID')
  if (values.confirm_flow_delete !== 'acknowledged') throw new Error('请确认删除整个数据流')
  const kind = values.flow_type
  if (kind !== 'symmetrical' && kind !== 'directional') throw new Error('请选择数据流类型')
  const flow = (group as unknown as { data_flow: Record<string, unknown> }).data_flow
  const entries = flow?.[kind]
  if (!Array.isArray(entries) || !entries.length || entries.some(entry => !entry || typeof entry !== 'object')) throw new Error('所选类型的数据流不存在或不可用')
  const base = { bucket_id: row!.natural_key as string, group_id: group.id, flow_type: kind, expected_group: JSON.stringify(group) }
  if (kind === 'symmetrical') {
    if (entries.filter(entry => entry.id === values.flow_id).length !== 1 || typeof values.flow_id !== 'string' || !values.flow_id) throw new Error('请输入唯一存在的对称流 ID')
    if (values.source_zone || values.dest_zone) throw new Error('对称流请清空源/目标 Zone 字段')
    return { ...base, flow_id: values.flow_id }
  }
  const zone = (v: unknown): v is string => typeof v === 'string' && !!v && !v.startsWith('-') && new TextEncoder().encode(v).length <= 512 && !/[\s,;=*\p{Cc}]/u.test(v) && ![...v].some(c => { const n = c.codePointAt(0)!; return n >= 0xd800 && n <= 0xdfff })
  if (!zone(values.source_zone) || !zone(values.dest_zone) || values.source_zone === values.dest_zone) throw new Error('请输入不同的源/目标 Zone ID，由后端核验对应数据流')
  if (values.flow_id) throw new Error('定向流没有存储的流 ID，请留空')
  return { ...base, source_zone: values.source_zone, dest_zone: values.dest_zone }
}
export function bucketSyncFlowDeleteConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input = bucketSyncFlowDeleteInput(values, row)
  const target = 'flow_id' in input ? `对称流 ID ${JSON.stringify(input.flow_id)}（包含全部 Zone）` : `定向 Zone ID ${JSON.stringify(input.source_zone)} → ${JSON.stringify(input.dest_zone)}`
  return `确认从 Bucket ID ${input.bucket_id} 的组 ${JSON.stringify(input.group_id)} 删除${target}？只移除此数据流，保留组状态、其他流和管道，不删除已有对象副本；不保证所有复制停止。不修改 Zonegroup 或提交 period。请备份策略并避免外部并发修改，核验失败不代表未生效，不自动回滚。`
}

export function bucketSyncPipeDeleteInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const group = groups(row).find(group => group.id === values.group_id)
  if (values.bucket_id !== row!.natural_key) throw new Error('Bucket ID 不可更改')
  if (!group) throw new Error('请输入当前策略中准确的同步组 ID')
  const pipes = (group as unknown as { pipes: unknown }).pipes
  if (!Array.isArray(pipes) || pipes.some(pipe => !pipe || typeof pipe !== 'object' || typeof pipe.id !== 'string')) throw new Error('管道数据不可用')
  if (typeof values.pipe_id !== 'string' || !values.pipe_id || pipes.filter(pipe => pipe.id === values.pipe_id).length !== 1) throw new Error('请输入唯一存在的完整管道 ID')
  if (values.confirm_pipe_delete !== 'acknowledged') throw new Error('请确认删除整个管道及其全部选择器和参数')
  return { bucket_id: row!.natural_key as string, group_id: group.id, pipe_id: values.pipe_id, expected_group: JSON.stringify(group) }
}
export function bucketSyncPipeDeleteConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input = bucketSyncPipeDeleteInput(values, row)
  const group = JSON.parse(input.expected_group)
  const pipe = group.pipes.find((pipe: { id: string }) => pipe.id === input.pipe_id)
  return `确认删除 Bucket ID ${input.bucket_id} 的组 ${JSON.stringify(input.group_id)} 中的整个管道 ${JSON.stringify(input.pipe_id)}？源选择：${JSON.stringify(pipe.source)}；目标选择：${JSON.stringify(pipe.dest)}。该管道全部选择器、过滤和权限参数将被移除；保留组状态、数据流及其他管道，不删除已有对象副本，不保证所有复制停止。不修改 Zonegroup 或提交 period。请备份策略并避免外部并发，核验失败不代表未生效，不自动回滚。`
}

export function bucketSyncPipeZonesInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const selected = bucketSyncPipeDeleteInput({ ...values, confirm_pipe_delete: 'acknowledged' }, row)
  const zones = (key: string) => {
    let result: unknown
    try { result = JSON.parse(String(values[key])) } catch { throw new Error('Zone ID 列表必须为 JSON 数组') }
    if (!Array.isArray(result) || !result.length || new Set(result).size !== result.length || result.some(id => typeof id !== 'string' || !id || id.startsWith('-') || new TextEncoder().encode(id).length > 512 || /[\s,;=\p{Cc}]/u.test(id) || (id.includes('*') && (id !== '*' || result.length !== 1)) || [...id].some(c => { const n = c.codePointAt(0)!; return n >= 0xd800 && n <= 0xdfff }))) throw new Error('请输入非空、无重复 Zone ID 数组；通配仅允许单独 ["*"]')
    return result as string[]
  }
  if (values.confirm_pipe_zones !== 'acknowledged') throw new Error('请确认 Zone 范围及非事务修改风险')
  return { ...selected, source_zones: zones('source_zones_json'), dest_zones: zones('dest_zones_json') }
}
export function bucketSyncPipeZonesConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const p = bucketSyncPipeZonesInput(values, row)
  return `确认修改 Bucket ID ${p.bucket_id} 的组 ${JSON.stringify(p.group_id)} 中管道 ${JSON.stringify(p.pipe_id)}？完整源 Zone IDs ${JSON.stringify(p.source_zones)}；完整目标 Zone IDs ${JSON.stringify(p.dest_zones)}。* 匹配全部 Zone，可能扩大复制范围。明确集合先增后删，通配与明确集合直接切换；非事务操作，中间范围可能变化，失败可能部分生效，不自动回滚或重试。保留桶选择器、执行身份和高级参数，不修改组状态、数据流或 Zonegroup/period。请备份并避免外部并发，核验成功不代表同步完成。`
}

export function syncPipePriorityInput(values: Record<string, unknown>) {
  if (values.priority === undefined || values.priority === null || values.priority === '') return {}
  if (typeof values.priority !== 'number' || !Number.isInteger(values.priority) || values.priority < -2147483648 || values.priority > 2147483647) throw new Error('优先级必须为有符号 32 位整数')
  return {priority:values.priority}
}
export function syncPipeStorageClassInput(values: Record<string, unknown>) {
  if (values.storage_class_mode === undefined || values.storage_class_mode === 'preserve') return {}
  if (values.storage_class_mode === 'empty') return {storage_class:''}
  if (values.storage_class_mode !== 'set' || typeof values.storage_class !== 'string' || !values.storage_class || values.storage_class.startsWith('-') || new TextEncoder().encode(values.storage_class).length > 512 || /\p{Cc}/u.test(values.storage_class) || [...values.storage_class].some(c => {const n=c.codePointAt(0)!;return n>=0xd800&&n<=0xdfff})) throw new Error('请输入完整目标存储类；空字符串请显式选择')
  return {storage_class:values.storage_class}
}
export function bucketSyncPipeUpdateInput(values: Record<string, unknown>, row?: Record<string, unknown>): Record<string, unknown> {
  const selected = bucketSyncPipeDeleteInput({ ...values, confirm_pipe_delete: 'acknowledged' }, row)
  const group = JSON.parse(selected.expected_group)
  group.pipes = group.pipes.filter((pipe: { id: string }) => pipe.id !== values.pipe_id)
  const input = bucketSyncPipeCreateInput({ ...values, source_zones_json: '["*"]', dest_zones_json: '["*"]', confirm_pipe_create: 'acknowledged' }, { ...row, bucket_sync_policy: { groups: [group] } })
  if (values.confirm_pipe_update !== 'acknowledged') throw new Error('请确认修改桶选择器及权限模式的影响')
  delete input.source_zones
  delete input.dest_zones
  input.expected_group = selected.expected_group
  return {...input,...syncPipePriorityInput(values),...syncPipeStorageClassInput(values)}
}
export function bucketSyncPipeUpdateConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input = bucketSyncPipeUpdateInput(values, row)
  return `确认修改 Bucket ID ${input.bucket_id} 的组 ${JSON.stringify(input.group_id)} 中管道 ${JSON.stringify(input.pipe_id)}？源租户/桶/实例 ${JSON.stringify([input.source_tenant, input.source_bucket, input.source_bucket_id])}；目标 ${JSON.stringify([input.dest_tenant, input.dest_bucket, input.dest_bucket_id])}；模式 ${input.mode}，用户 ${JSON.stringify(input.user)}。* 为通配，空租户不限定租户。system 模式保留已存储 UID（不使用其权限检查），不会删除用户或凭据。优先级：${input.priority === undefined ? '保持原值' : input.priority}，可能改变匹配管道的选择。目标存储类：${input.storage_class === undefined ? '保持原值' : JSON.stringify(input.storage_class)}；空字符串仍是显式覆盖，不是移除字段。保留 Zone 成员、过滤器和目标 ACL。不验证目标放置配置或已有对象迁移。可能改变复制范围或权限；请备份并避免外部并发，失败不代表未生效，不自动回滚。仅修改桶本地管道，不提交 period，不代表同步完成。`
}

export function bucketSyncPipeCreateInput(values: Record<string, unknown>, row?: Record<string, unknown>): Record<string, unknown> {
  const group = groups(row).find(group => group.id === values.group_id)
  if (values.bucket_id !== row!.natural_key) throw new Error('Bucket ID 不可更改')
  if (!group) throw new Error('请输入当前策略中准确的同步组 ID')
  return { bucket_id: row!.natural_key, ...syncPipeCreateFields(values, group) }
}
export function syncPipeCreateFields(values: Record<string, unknown>, group: { id: string; status: string }) {
  const token = (v: unknown): v is string => typeof v === 'string' && !!v && !v.startsWith('-') && new TextEncoder().encode(v).length <= 512 && !/\p{Cc}/u.test(v) && ![...v].some(c => { const n = c.codePointAt(0)!; return n >= 0xd800 && n <= 0xdfff })
  if (!token(values.pipe_id)) throw new Error('请输入合法的新管道 ID')
  const pipes = (group as unknown as { pipes: unknown }).pipes
  if (!Array.isArray(pipes) || pipes.some(pipe => !pipe || typeof pipe !== 'object' || typeof pipe.id !== 'string')) throw new Error('管道数据不可用')
  if (pipes.some(pipe => pipe.id === values.pipe_id)) throw new Error('管道 ID 已存在，创建不会修改已有管道')
  if (values.mode !== 'system' && values.mode !== 'user') throw new Error('请选择 system 或 user 模式')
  const user = values.user === undefined ? '' : values.user
  if (values.mode === 'system' && user !== '') throw new Error('system 模式请清空用户')
  if (values.mode === 'user') {
    if (!token(user) || /\s/u.test(user)) throw new Error('请输入完整用户 UID')
    const parts = user.split('$')
    if (!(parts.length === 1 || (parts.length === 2 && parts.every(Boolean)) || (parts.length === 3 && parts[1] && parts[2]))) throw new Error('用户 UID 格式无效')
  }
  const result: Record<string, unknown> = { group_id: group.id, pipe_id: values.pipe_id, expected_group: JSON.stringify(group), mode: values.mode }
  if (values.mode === 'user') result.user = user
  for (const side of ['source', 'dest']) {
    let zones: unknown
    try { zones = JSON.parse(String(values[side + '_zones_json'])) } catch { throw new Error('Zone ID 必须为 JSON 数组') }
    if (!Array.isArray(zones) || !zones.length || zones.some(id => !token(id) || /[\s,;=]/u.test(id) || (id.includes('*') && (id !== '*' || zones.length !== 1))) || new Set(zones).size !== zones.length) throw new Error('Zone ID 列表不合法；通配符仅允许单独 ["*"]')
    result[side + '_zones'] = zones
    for (const suffix of ['tenant', 'bucket', 'bucket_id']) {
      const key = side + '_' + suffix
      const v = values[key] === undefined || values[key] === '' ? (suffix === 'bucket_id' ? '*' : '') : values[key]
      if ((v !== '' && (!token(v) || /[\s/:\\]/u.test(v) || (v.includes('*') && v !== '*'))) || (suffix === 'bucket' && !v)) throw new Error('请明确填写桶选择器；各字段不可含分隔符，* 必须独立使用')
      result[key] = v
    }
  }
  if (values.confirm_pipe_create !== 'acknowledged') throw new Error('请确认管道的匹配范围及复制影响')
  return result
}
export function bucketSyncPipeCreateConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const p = bucketSyncPipeCreateInput(values, row)
  return `确认在 Bucket ID ${p.bucket_id} 的组 ${JSON.stringify(p.group_id)} 创建管道 ${JSON.stringify(p.pipe_id)}？源：Zone IDs ${JSON.stringify(p.source_zones)}，租户/桶/实例 ${JSON.stringify([p.source_tenant,p.source_bucket,p.source_bucket_id])}；目标：${JSON.stringify(p.dest_zones)}，${JSON.stringify([p.dest_tenant,p.dest_bucket,p.dest_bucket_id])}。模式 ${p.mode}，用户 ${JSON.stringify(p.user)}。* 为通配；空租户不限定租户，不代表仅全局租户。新管道优先级 0、无前缀/标签过滤或目标 ACL/存储类覆盖，可能影响复制范围；不代表有效链路或同步完成。不修改组状态、数据流或 Zonegroup/period。请备份策略并避免外部并发，核验失败不代表未生效，不自动回滚。`
}

export function bucketSyncGroupCreateBlocked(row: Record<string, unknown>) {
  try { groups(row, true); return undefined } catch (error) { return (error as Error).message }
}
export function bucketSyncGroupCreateInitial(row?: Record<string, unknown>) {
  groups(row, true)
  return { bucket_id: row!.natural_key as string, group_id: '', status: undefined, confirm_create: undefined }
}
export function bucketSyncGroupCreateInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const available = groups(row, true)
  if (values.bucket_id !== row!.natural_key) throw new Error('Bucket ID 不可更改')
  const id = values.group_id
  if (typeof id !== 'string' || !id || id.startsWith('-') || new TextEncoder().encode(id).length > 512 || /\p{Cc}/u.test(id) || [...id].some(char => { const code = char.codePointAt(0)!; return code >= 0xd800 && code <= 0xdfff })) throw new Error('请输入合法的新同步组 ID（最多 512 字节，不以 - 开头，不含控制字符）')
  if (available.some(group => group.id === id)) throw new Error('同步组已存在，请使用状态修改操作')
  if (!['enabled', 'allowed', 'forbidden'].includes(values.status as string)) throw new Error('请明确选择初始状态')
  if (values.confirm_create !== 'acknowledged') throw new Error('请确认同步组创建范围')
  return { bucket_id: row!.natural_key as string, group_id: id, status: values.status as string }
}
export function bucketSyncGroupCreateConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input = bucketSyncGroupCreateInput(values, row)
  return `确认在 Bucket ID ${input.bucket_id} 创建同步组 ${JSON.stringify(input.group_id)}，状态为 ${input.status}？仅创建空数据流和空管道的桶本地组，不建立可工作的复制链路、不修改 Zonegroup 或提交 period。后续数据流和管道仍需配置；外部并发创建相同 ID 可能被原生命令修改，请避免并发。核验失败不代表未生效，不会自动回滚。`
}

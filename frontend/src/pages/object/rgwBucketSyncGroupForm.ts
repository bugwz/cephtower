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

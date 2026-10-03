export function objectLockFormInitial(row?: Record<string, unknown>) {
  if (!row || row.kind !== 'object-lock' || typeof row.bucket_id !== 'string' || !/^[A-Za-z0-9_-]+$/.test(row.bucket_id)) throw new Error('对象锁身份不可用')
  const base = { bucket_id: row.bucket_id, kind: 'object-lock', retention_action: undefined, confirm_lock: undefined }
  if (row.configured === false && row.object_lock === null) return { ...base, mode: undefined, unit: undefined, period: '' }
  if (row.configured !== true || !row.object_lock || typeof row.object_lock !== 'object') throw new Error('对象锁状态不可用，请刷新')
  const configuration = row.object_lock as Record<string, unknown>
  if (configuration.enabled !== true) throw new Error('对象锁启用状态不可用')
  if (configuration.default_retention === null) return { ...base, mode: undefined, unit: undefined, period: '' }
  const retention = configuration.default_retention as Record<string, unknown> | undefined
  if (!retention || (retention.mode !== 'GOVERNANCE' && retention.mode !== 'COMPLIANCE') || (retention.days === null) === (retention.years === null)
    || (retention.days !== null && typeof retention.days !== 'string') || (retention.years !== null && typeof retention.years !== 'string')) throw new Error('当前默认保留策略不受编辑器支持')
  return { ...base, mode: retention.mode, unit: retention.days !== null ? 'Days' : 'Years', period: retention.days ?? retention.years }
}
export function objectLockFormBlocked(row: Record<string, unknown>) {
  try { objectLockFormInitial(row); return undefined } catch (error) { return error instanceof Error ? error.message : '对象锁不可用' }
}
export function objectLockFormInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const initial = objectLockFormInitial(row)
  if (values.bucket_id !== initial.bucket_id || values.kind !== initial.kind) throw new Error('不能更改 Bucket 身份或配置类型')
  if (values.confirm_lock !== 'acknowledged') throw new Error('请明确确认对象锁无法关闭及保留影响')
  let rule = ''
  if (values.retention_action === 'set') {
    if (values.mode !== 'GOVERNANCE' && values.mode !== 'COMPLIANCE') throw new Error('请选择保留模式')
    if (values.unit !== 'Days' && values.unit !== 'Years') throw new Error('请选择天或年')
    if (typeof values.period !== 'string' || !/^[0-9]+$/.test(values.period) || BigInt(values.period) < 1n || BigInt(values.period) > 2147483647n) throw new Error('保留期必须为 1–2147483647 的整数')
    rule = `<Rule><DefaultRetention><Mode>${values.mode}</Mode><${values.unit}>${BigInt(values.period).toString()}</${values.unit}></DefaultRetention></Rule>`
  } else if (values.retention_action !== 'clear') throw new Error('请选择设置或清除默认保留策略')
  return { bucket_id: initial.bucket_id, kind: 'object-lock', document: `<ObjectLockConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><ObjectLockEnabled>Enabled</ObjectLockEnabled>${rule}</ObjectLockConfiguration>` }
}
export function objectLockFormConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input = objectLockFormInput(values, row)
  const effect = values.retention_action === 'clear' ? '清除默认保留期（保留对象锁启用状态）' : `设置默认 ${values.mode} 保留 ${values.period} ${values.unit === 'Days' ? '天' : '年'}`
  return `确认对 Bucket ID ${input.bucket_id} ${effect}？对象锁一旦启用无法关闭，首次启用要求版本控制为 Enabled，目标 Ceph 版本也必须支持。默认策略影响后续对象写入，可能长期阻止删除；不会解除已有对象的独立保留或 Legal Hold。请先备份配置，外部并发修改可能被覆盖；提交后回读核验。`
}

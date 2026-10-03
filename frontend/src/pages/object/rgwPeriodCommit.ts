function identity(row?: Record<string, unknown>) {
  const valid = (v: unknown): v is string => typeof v === 'string' && !!v && !v.startsWith('-') && new TextEncoder().encode(v).length <= 512 && !/\p{Cc}/u.test(v) && ![...v].some(c => { const n = c.codePointAt(0)!; return n >= 0xd800 && n <= 0xdfff })
  if (!row || row.stale === true || !valid(row.id) || !valid(row.current_period)) throw new Error('Realm 或当前 Period 数据不可用，请刷新库存')
  return { realm_id: row.id, expected_current_period: row.current_period }
}
export function periodCommitInitial(row?: Record<string, unknown>) { return { ...identity(row), confirm_commit: undefined } }
export function periodCommitBlocked(row: Record<string, unknown>) { try { identity(row); return undefined } catch (error) { return (error as Error).message } }
export function periodCommitInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const target = identity(row)
  if (values.realm_id !== target.realm_id || values.expected_current_period !== target.expected_current_period) throw new Error('Realm 和当前 Period 不可更改')
  if (values.confirm_commit !== 'acknowledged') throw new Error('请确认 Realm 范围发布影响')
  return target
}
export function periodCommitConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const p = periodCommitInput(values, row)
  return `确认提交 Realm ${JSON.stringify(p.realm_id)} 的 Period（采集时 ${JSON.stringify(p.expected_current_period)}）？将执行 period update --commit，可能发布该 Realm 的全部待提交 Zonegroup/Zone 变更，不仅是最近编辑的一项。请先备份并核对所有变更，避免外部并发；失败可能部分生效，不自动回滚或重试。回读核验不代表远端已同步。`
}

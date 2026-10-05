export type SyncReportSections = { identity: string; metadata: string; sources: string[] }

// Classify only native status lines, never zone names or diagnostic substrings.
// Even "caught up" describes this sample/source, not overall replication health.
export function rgwSyncNotice(section: string): { type: 'error' | 'warning' | 'info'; message: string; details: string[] } | undefined {
  const lines = section.split('\n').map(line => line.trim())
  const errors = lines.filter(line => line === 'zone not found' || /^(?:failed to retrieve sync info|failed to read sync status|failed read sync status|failed read recovering shards|failed to fetch source sync status|failed to fetch master sync status): .+$/.test(line))
  if (errors.length) return { type: 'error', message: '原生报告包含同步读取错误，状态可能不完整', details: errors }
  const warnings = lines.filter(line => /^(?:metadata|data) is behind on [1-9][0-9]* shards$/.test(line) || /^[1-9][0-9]* shards are recovering$/.test(line) || /^master is on a different period: master_period=.* local_period=.*$/.test(line))
  if (warnings.length) return { type: 'warning', message: '原生报告存在同步落后、恢复或 Period 不一致', details: warnings }
  const states = lines.filter(line => ['init', 'preparing for full sync', 'syncing', 'unknown', 'no sync (zone is master)', 'not syncing from zone', 'metadata is caught up with master', 'data is caught up with source'].includes(line))
  return states.length ? { type: 'info', message: '原生采样状态（不代表整体同步健康）', details: states } : undefined
}

export type SyncCounters = { full?: string; incremental?: string; total?: string; remaining?: string; remainingUnit?: 'entries' | 'buckets'; behind?: string; recovering?: string; oldestChange?: string; oldestShard?: string; behindShards?: string[]; recoveringShards?: string[] }
// These are native phase counts, not completed-work percentages. Parse only
// whole known lines and retain the original section for all diagnostics.
export function rgwSyncCounters(section: string): SyncCounters | undefined {
  const result: SyncCounters = {}
  const seen = new Set<string>()
  const integer = '(0|[1-9][0-9]*)'
  // ceph::real_time streams local time with six fractional digits and %z.
  // Keep the native timestamp intact; Date would discard microseconds.
  const timestamp = '[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\\.[0-9]{6}[+-][0-9]{4}'
  for (const line of section.split('\n').map(value => value.trim())) {
    const shards = line.match(/^(behind|recovering) shards: \[((?:(?:0|[1-9][0-9]*)(?:,(?:0|[1-9][0-9]*))*)?)\]$/)
    if (shards) {
      const field = shards[1] === 'behind' ? 'behindShards' : 'recoveringShards'
      const ids = shards[2] ? shards[2].split(',') : []
      // Native std::set output is unique and numerically sorted, not lexical.
      if (seen.has(field) || ids.some((id, index) => index > 0 && BigInt(ids[index - 1]) >= BigInt(id))) return undefined
      seen.add(field); result[field] = ids
      continue
    }
    const oldest = line.match(new RegExp(`^oldest incremental change not applied: (${timestamp}) \\[${integer}\\]$`))
    if (oldest) {
      if (seen.has('oldestChange')) return undefined
      seen.add('oldestChange'); result.oldestChange = oldest[1]; result.oldestShard = oldest[2]
      continue
    }
    let match = line.match(new RegExp(`^(full|incremental) sync: ${integer}/${integer} shards$`))
    if (match) {
      const field = match[1] === 'full' ? 'full' : 'incremental'
      if (seen.has(field) || BigInt(match[2]) > BigInt(match[3]) || (result.total !== undefined && result.total !== match[3])) return undefined
      seen.add(field); result[field] = match[2]; result.total = match[3]
      continue
    }
    match = line.match(new RegExp(`^full sync: ${integer} (entries|buckets) to sync$`))
    if (match) {
      if (seen.has('remaining')) return undefined
      seen.add('remaining'); result.remaining = match[1]; result.remainingUnit = match[2] as 'entries' | 'buckets'
      continue
    }
    match = line.match(new RegExp(`^(?:metadata|data) is behind on ${integer} shards$`))
    if (match) {
      if (seen.has('behind')) return undefined
      seen.add('behind'); result.behind = match[1]
      continue
    }
    match = line.match(new RegExp(`^${integer} shards are recovering$`))
    if (match) {
      if (seen.has('recovering')) return undefined
      seen.add('recovering'); result.recovering = match[1]
    }
  }
  if (result.full !== undefined && result.incremental !== undefined && BigInt(result.full) + BigInt(result.incremental) > BigInt(result.total!)) return undefined
  return seen.size ? result : undefined
}

// Native tab_dump uses a 15-character label and one separating space. Do not
// split on arbitrary occurrences of "source:" inside diagnostic messages.
export function rgwSyncReportSections(report: string): SyncReportSections | undefined {
  const lines = report.split('\n')
  const metadata = lines.findIndex(line => line.startsWith('  metadata sync '))
  if (metadata < 3) return undefined
  const data = lines.findIndex((line, index) => index > metadata && line.startsWith('      data sync '))
  const sources: string[] = []
  if (data >= 0) {
    let source: string[] = []
    for (const [index, line] of lines.slice(data).entries()) {
      const content = index === 0 ? line.slice(16) : line.startsWith('                ') ? line.slice(16) : line
      if (content.startsWith('source: ') && source.length) {
        sources.push(source.join('\n').trimEnd())
        source = []
      }
      source.push(content)
    }
    if (source.some(line => line.trim())) sources.push(source.join('\n').trimEnd())
  }
  return {
    identity: lines.slice(0, metadata).join('\n').trimEnd(),
    metadata: lines.slice(metadata, data < 0 ? undefined : data).map((line, index) => index === 0 ? line.slice(16) : line.startsWith('                ') ? line.slice(16) : line).join('\n').trimEnd(),
    sources
  }
}

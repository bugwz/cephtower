export type SyncReportSections = { identity: string; metadata: string; sources: string[] }

export type SyncCounters = { full?: string; incremental?: string; total?: string; remaining?: string; remainingUnit?: 'entries' | 'buckets'; behind?: string; recovering?: string }
// These are native phase counts, not completed-work percentages. Parse only
// whole known lines and retain the original section for all diagnostics.
export function rgwSyncCounters(section: string): SyncCounters | undefined {
  const result: SyncCounters = {}
  const seen = new Set<string>()
  const integer = '(0|[1-9][0-9]*)'
  for (const line of section.split('\n').map(value => value.trim())) {
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

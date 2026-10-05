export type SyncReportSections = { identity: string; metadata: string; sources: string[] }

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

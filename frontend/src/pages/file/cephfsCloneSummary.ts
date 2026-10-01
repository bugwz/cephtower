function record(value: unknown): Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value) ? value as Record<string, unknown> : {}
}

export function cloneSource(value: unknown): string {
  if (value === 'N/A') return '来源不可用'
  const source = record(value)
  if (typeof source.volume !== 'string' || !source.volume || typeof source.subvolume !== 'string' || !source.subvolume || typeof source.snapshot !== 'string' || !source.snapshot) return '—'
  const group = typeof source.group === 'string' && source.group ? source.group : '_nogroup'
  return `${source.volume}/${group}/${source.subvolume}@${source.snapshot}`
}

export function cloneProgress(value: unknown) {
  const report = record(value)
  const raw = report['percentage cloned']
  const percent = typeof raw === 'string' && /^\d+(?:\.\d+)?%$/.test(raw) ? Number(raw.slice(0, -1)) : undefined
  return {
    percent: percent !== undefined && Number.isFinite(percent) && percent >= 0 && percent <= 100 ? percent : undefined,
    amount: typeof report['amount cloned'] === 'string' ? report['amount cloned'] : undefined,
    files: typeof report['files cloned'] === 'string' ? report['files cloned'] : undefined
  }
}

export function cloneFailure(value: unknown): string {
  const failure = record(value)
  const message = typeof failure.error_msg === 'string' && failure.error_msg ? failure.error_msg : undefined
  const errno = typeof failure.errno === 'string' && /^-?\d+$/.test(failure.errno) ? failure.errno : typeof failure.errno === 'number' && Number.isSafeInteger(failure.errno) ? String(failure.errno) : undefined
  return message ? `${message}${errno === undefined ? '' : ` (errno ${errno})`}` : errno === undefined ? '—' : `errno ${errno}`
}

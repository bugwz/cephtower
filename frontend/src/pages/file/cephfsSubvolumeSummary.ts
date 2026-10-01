import { groupPermissionMode } from './cephfsGroupForm'
import { formatPoolBytes } from './cephfsPoolCapacity'

export function cephFSBytes(value: unknown): string {
  if (typeof value === 'string' && /^\d+$/.test(value)) return formatPoolBytes(value)
  if (typeof value === 'number' && Number.isSafeInteger(value) && value >= 0) return formatPoolBytes(String(value))
  return '—'
}

export function cephFSQuota(value: unknown): string {
  return value === 'infinite' ? '无限制' : cephFSBytes(value)
}

export function cephFSUsage(quota: unknown, used: unknown, percent: unknown) {
  const usedText = cephFSBytes(used)
  const quotaText = cephFSQuota(quota)
  if (quota === 'infinite') return { text: `${usedText}（无限制）`, percent: undefined, barPercent: undefined }
  const numeric = typeof percent === 'number' ? percent : typeof percent === 'string' && /^\d+(?:\.\d+)?$/.test(percent) ? Number(percent) : undefined
  const valid = numeric !== undefined && Number.isFinite(numeric) && numeric >= 0 ? numeric : undefined
  return {
    text: `${usedText} / ${quotaText}`,
    percent: valid,
    barPercent: valid === undefined ? undefined : Math.min(100, valid)
  }
}

export function cephFSPermissions(value: unknown): { octal: string; owner: string; group: string; others: string } | undefined {
  const octal = groupPermissionMode(value)
  if (octal === undefined) return undefined
  const mode = parseInt(octal, 8)
  const bits = ['---', '--x', '-w-', '-wx', 'r--', 'r-x', 'rw-', 'rwx']
  const special = (text: string, bit: number, lower: string, upper: string) => mode & bit ? text.slice(0, 2) + (text[2] === 'x' ? lower : upper) : text
  return {
    octal,
    owner: special(bits[(mode >> 6) & 7], 0o4000, 's', 'S'),
    group: special(bits[(mode >> 3) & 7], 0o2000, 's', 'S'),
    others: special(bits[mode & 7], 0o1000, 't', 'T')
  }
}

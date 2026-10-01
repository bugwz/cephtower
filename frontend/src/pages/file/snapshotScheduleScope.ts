import type { ApiRecord } from '../../api/client'

export function scheduleScope(row: ApiRecord): ApiRecord {
  return {
    fs: row.fs,
    path: row.path,
    ...(row.subvol ? { subvol: row.subvol } : {}),
    ...(row.group ? { group: row.group } : {})
  }
}

export function scheduleFormScope(row: ApiRecord): ApiRecord {
  return { subvol: undefined, group: undefined, ...scheduleScope(row) }
}

export function sameScheduleScope(left: ApiRecord, right: ApiRecord): boolean {
  return Boolean(left.fs && right.fs && left.path && right.path)
    && left.fs === right.fs && left.path === right.path
    && (left.subvol || '') === (right.subvol || '')
    && (left.group || '_nogroup') === (right.group || '_nogroup')
}

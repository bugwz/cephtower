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

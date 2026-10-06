import type { ApiRecord } from '../../api/client'

export function monQuorumGroups(rows: ApiRecord[]) {
  return [
    { key: 'in', label: '仲裁中', rows: rows.filter(row => row.in_quorum === true) },
    { key: 'out', label: '未加入仲裁', rows: rows.filter(row => row.in_quorum === false) },
    { key: 'unknown', label: '仲裁状态未知', rows: rows.filter(row => row.in_quorum !== true && row.in_quorum !== false) }
  ]
}

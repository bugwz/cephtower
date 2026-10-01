import type { ApiRecord } from '../../api/client'
import type { MutationFormValues } from '../ResourceListPage'
import { cephFSQuotaUpdateValues, cephFSQuotaDecimal } from './cephfsGroupForm'

export function subvolumeUpdateInitialValues(row?: ApiRecord): MutationFormValues {
  const size = cephFSQuotaDecimal(row?.bytes_quota)
  return { size, unlimited: row?.bytes_quota === 'infinite', no_shrink: false }
}

export function subvolumeUpdateBody(values: MutationFormValues, clusterId: number, filesystem: string, subvolume: string, group: string): ApiRecord {
  return { cluster_id: clusterId, fs: filesystem, subvolume, group, ...cephFSQuotaUpdateValues(values) }
}

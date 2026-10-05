import type { CephCluster } from '../../api/cluster'
import type { ApiRecord } from '../../api/client'
import { rgwRealmImportAction } from './rgwRealmImport'

export type TransferCluster = Pick<CephCluster, 'id' | 'fsid' | 'generation' | 'enabled'>
export function validTransferPair(source?: TransferCluster, target?: TransferCluster) {
  return !!source && !!target && source.enabled && target.enabled && source.id !== target.id &&
    !!source.fsid && !!target.fsid && source.fsid.toLowerCase() !== target.fsid.toLowerCase()
}

// Token stays local to this invocation; neither form state nor browser storage
// receives it. Revalidate both connection identities before the target write.
export async function transferRealm(input: {
  source: TransferCluster; target: TransferCluster; realmId: string; realmName: string; values: ApiRecord
}, deps: {
  current: () => boolean
  clusters: () => Promise<TransferCluster[]>
  token: (body: ApiRecord) => Promise<{ token: string }>
  importZone: (body: ApiRecord) => Promise<unknown>
}) {
  const { source, target, realmId, realmName } = input
  if (!validTransferPair(source, target) || !realmId || !realmName || !deps.current()) throw new Error('invalid transfer scope')
  const values = { ...input.values }
  const build = (token: string) => rgwRealmImportAction.buildBody!({ ...values, realm_token: token }, target.id)
  build('cHJldmlldw==') // Validate deployment inputs before reading credentials.
  const verify = async () => {
    const rows = await deps.clusters()
    if (!deps.current()) throw new Error('transfer scope changed')
    for (const expected of [source, target]) {
      const matches = rows.filter(row => row.id === expected.id)
      if (matches.length !== 1 || !matches[0].enabled || matches[0].fsid !== expected.fsid || matches[0].generation !== expected.generation) throw new Error('cluster identity changed')
    }
  }
  await verify()
  const result = await deps.token({ cluster_id: source.id, realm_id: realmId, name: realmName })
  try {
    if (!deps.current()) throw new Error('transfer scope changed')
    const body = build(result.token)
    try {
      await verify()
      return await deps.importZone(body)
    } finally { delete body.realm_token }
  } finally { result.token = '' }
}

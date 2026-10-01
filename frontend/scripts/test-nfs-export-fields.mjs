import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/file/nfsExportFields.ts', import.meta.url), 'utf8')
const exports = {}
new Function('exports', ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const row = { cluster_id: 'nfs-a', pseudo: '/share', path: '/data', access_type: 'RO', fsal: { name: 'CEPH', fs_name: 'cephfs-a', user_id: 'nfs.user' } }
assert.deepEqual(exports.nfsExportInitialValues(row), { cluster: 'nfs-a', pseudo: '/share', path: '/data', filesystem: 'cephfs-a', read_only: true })
assert.equal(exports.nfsExportEditReason(row), undefined)
assert.equal(exports.nfsExportInitialValues({ ...row, access_type: 'RW' }).read_only, false)
assert.ok(exports.nfsExportEditReason({ ...row, fsal: { name: 'RGW' } }))
assert.ok(exports.nfsExportEditReason({ ...row, access_type: 'NONE' }))
for (const fsal of [null, undefined, [], 'CEPH']) assert.deepEqual(exports.nfsFSAL({ fsal }), {})
console.log('NFS native export field checks passed')

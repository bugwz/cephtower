import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/file/nfsExportFields.ts', import.meta.url), 'utf8')
const exports = {}
new Function('exports', ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const row = { cluster_id: 'nfs-a', pseudo: '/share', path: '/data', access_type: 'RO', fsal: { name: 'CEPH', fs_name: 'cephfs-a', user_id: 'nfs.user' } }
assert.deepEqual(exports.nfsExportInitialValues(row), { cluster: 'nfs-a', pseudo: '/share', path: '/data', filesystem: 'cephfs-a', read_only: true, squash: undefined })
assert.equal(exports.nfsExportInitialValues({ ...row, squash: 'all_squash' }).squash, 'all_squash')
assert.equal(exports.nfsExportInitialValues({ ...row, squash: 'native-alias' }).squash, undefined)
assert.equal(exports.nfsExportEditReason(row), undefined)
assert.equal(exports.nfsExportInitialValues({ ...row, access_type: 'RW' }).read_only, false)
assert.ok(exports.nfsExportEditReason({ ...row, fsal: { name: 'RGW' } }))
assert.ok(exports.nfsExportEditReason({ ...row, access_type: 'NONE' }))
for (const fsal of [null, undefined, [], 'CEPH']) assert.deepEqual(exports.nfsFSAL({ fsal }), {})
console.log('NFS native export field checks passed')
const pagesSource = readFileSync(new URL('../src/pages/file/pages.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('pages.tsx', pagesSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const identity = tree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name?.text === 'exportId')
assert.ok(identity)
const code = ts.transpileModule(identity.getText(tree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
const exportId = new Function(`${code}; return exportId`)()
assert.equal(exportId({ export_id: 2, natural_key: 'bmZzLWEAMg' }), 'bmZzLWEAMg')
assert.equal(exportId({ export_id: 2 }), '')

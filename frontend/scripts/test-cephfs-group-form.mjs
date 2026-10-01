import assert from 'node:assert/strict'
import fs from 'node:fs'
import ts from 'typescript'

const source = fs.readFileSync(new URL('../src/pages/file/cephfsGroupForm.ts', import.meta.url), 'utf8')
const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText
const module = { exports: {} }
new Function('module', 'exports', compiled)(module, module.exports)
const { groupPermissionMode, groupUpdateInitialValues, groupUpdateBody } = module.exports

assert.equal(groupPermissionMode(16877), '0755')
assert.equal(groupPermissionMode(0o42750), '2750')
assert.equal(groupPermissionMode(0), '0000')
for (const value of [null, '', true, 'rwx', -1, 1.5, 999999]) assert.equal(groupPermissionMode(value), undefined)
const initial = groupUpdateInitialValues({ bytes_quota: 'infinite', data_pool: 'cephfs.hot', uid: 1000, gid: 1001, mode: 16832 })
assert.equal(initial.unlimited, true)
assert.equal(initial.size, undefined)
assert.equal(initial.mode, '0700')
assert.equal(initial.edit_attributes, false)
assert.equal(groupUpdateInitialValues({ bytes_quota: '9007199254740993' }).size, undefined)
assert.equal(groupUpdateInitialValues({ bytes_quota: 1024 }).size, 1024)
assert.deepEqual(groupUpdateBody({ ...initial, no_shrink: true }, 3, 'cephfs', 'users'), { cluster_id: 3, fs: 'cephfs', group: 'users', unlimited: true, no_shrink: false })
assert.deepEqual(groupUpdateBody({ ...initial, unlimited: false, size: 2048, no_shrink: true, edit_attributes: true }, 3, 'cephfs', 'users'), {
  cluster_id: 3, fs: 'cephfs', group: 'users', size: 2048, no_shrink: true, pool: 'cephfs.hot', uid: 1000, gid: 1001, mode: '0700'
})
console.log('CephFS subvolume group form checks passed')

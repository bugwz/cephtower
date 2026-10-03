import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const exports = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwPermissionRows.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const rows = exports.rgwPermissionRows
assert.deepEqual(rows([{ type: 'users', perm: '*' }, { type: 'future', perm: 'read, write' }]), [
  { key: 0, identity: 'users', permission: '*' }, { key: 1, identity: 'future', permission: 'read, write' }
])
assert.deepEqual(rows([{ id: 'tenant$user:sub', permissions: 'full-control' }], true), [{ key: 0, identity: 'tenant$user:sub', permission: 'full-control' }])
assert.equal(rows([{ type: 'users', perm: '' }])[0].permission, '未指定权限')
assert.deepEqual(rows([]), [])
for (const value of [undefined, null, {}, [null], [[]], [{}], [{ type: '', perm: '*' }], [{ type: 'users', perm: null }]]) assert.equal(rows(value), undefined)
assert.equal(rows([{ type: 'users', perm: '*' }], true), undefined)
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.ok(pages.includes('<RgwPermissions value={value} />'))
assert.ok(pages.includes('<RgwPermissions value={value} subusers />'))
console.log('RGW capabilities and subusers preserve native identity and permission strings')

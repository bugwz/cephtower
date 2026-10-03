import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const exports = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwAccountLimit.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const limit = exports.rgwAccountLimit
const patch = exports.rgwAccountLimitPatch
assert.deepEqual(patch({}), {})
assert.deepEqual(patch({ max_users: 5, max_buckets: null }, { max_users: 5 }), {})
assert.deepEqual(patch({ max_users: '', max_roles: undefined }), {})
for (const key of ['max_users', 'max_roles', 'max_groups', 'max_buckets', 'max_access_keys']) {
  for (const value of [-1, 0, 1, 2147483647]) assert.deepEqual(patch({ [key]: value }), { [key]: value })
  for (const value of [-2, 0.5, '0', true, NaN, Infinity, 2147483648]) assert.throws(() => patch({ [key]: value }))
}
assert.deepEqual(patch({ max_users: 5, max_roles: 6 }, { max_users: 5, max_roles: 2 }), { max_roles: 6 })
for (const value of [-1, -2, -2147483648]) assert.equal(limit(value), '无限制')
assert.equal(limit(0), '0（禁止新增）')
assert.equal(limit(2), '2')
assert.equal(limit(2147483647), '2147483647')
for (const value of [undefined, null, '0', NaN, Infinity, 0.5, 2147483648, -2147483649]) assert.equal(limit(value), '资源上限未返回或无效')
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.equal(pages.match(/render: rgwAccountLimit/g).length, 4)
assert.ok(pages.includes("title: '每用户访问密钥上限'"))
assert.ok(pages.includes('...rgwAccountLimitPatch(values, row)'))
const source = ts.createSourceFile('pages.tsx', pages, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const initializer = source.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === 'numberOrUndefined')
const initialExports = {}
new Function('exports', ts.transpileModule('export ' + initializer.getText(source), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(initialExports)
for (const value of [null, undefined, '', '0', false, 0.5, 2147483648]) assert.equal(initialExports.numberOrUndefined(value), undefined)
for (const value of [-1, 0, 2147483647]) assert.equal(initialExports.numberOrUndefined(value), value)
console.log('RGW IAM limits preserve zero creation bans and negative unlimited values')

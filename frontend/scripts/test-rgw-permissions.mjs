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
const pages = readFileSync(new URL('../src/pages/object/RgwUserDetails.tsx', import.meta.url), 'utf8')
assert.ok(pages.includes('<RgwPermissions value={row.caps} />'))
assert.ok(pages.includes('<RgwPermissions value={row.subusers} subusers />'))
console.log('RGW capabilities and subusers preserve native identity and permission strings')

const actionSource = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const source = ts.createSourceFile('pages.tsx', actionSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let confirmation
function visit(node) {
  if (ts.isObjectLiteralExpression(node)) {
    const property = name => node.properties.find(item => ts.isPropertyAssignment(item) && item.name.getText(source) === name)?.initializer
    if (property('title')?.text === '管理用户权限（caps）') {
      const expression = property('confirmation').getText(source)
      const code = ts.transpileModule(`const confirmation = ${expression}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
      confirmation = new Function('userId', `${code}; return confirmation`)(row => row.uid)
    }
  }
  ts.forEachChild(node, visit)
}
visit(source)
assert.equal(typeof confirmation, 'function')
for (const [action, wording] of [['add', '合并添加'], ['rm', '仅移除']]) {
  const text = confirmation({ action, type: 'users', permission: 'write' }, { uid: 'tenant$ns$user' })
  for (const expected of [wording, 'tenant$ns$user', 'users', 'write', '不是整体替换', '其它权限保留', '其他用户的数据']) assert.ok(text.includes(expected))
}
console.log('RGW capability confirmation identifies scoped additive and subtractive semantics')

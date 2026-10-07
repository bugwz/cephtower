import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
import './test-rgw-bucket-configuration.mjs'
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
const view = {}
const jsx = (type, props) => ({ type, props })
new Function('exports', 'require', ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwPermissions.tsx', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText)(view, name => {
  if (name === 'antd') return { Table: 'Table' }
  if (name === './rgwPermissionRows') return exports
  if (name === 'react/jsx-runtime') return { jsx, jsxs: jsx }
  throw new Error(name)
})
for (const subusers of [false, true]) {
  const value = subusers ? [{ id: 'z', permissions: 'future' }, { id: 'a', permissions: 'read' }, { id: 'b', permissions: 'read' }]
    : [{ type: 'z', perm: 'future' }, { type: 'a', perm: 'read' }, { type: 'b', perm: 'read' }]
  const snapshot = JSON.stringify(value)
  const table = view.RgwPermissions({ value, subusers })
  const [identity, permission] = table.props.columns
  assert.equal(identity.defaultSortOrder, 'ascend')
  assert.ok(identity.sorter(table.props.dataSource[0], table.props.dataSource[1]) > 0)
  assert.equal(identity.sorter(table.props.dataSource[0], table.props.dataSource[0]), 0)
  assert.deepEqual(permission.filters, [{ text: 'future', value: 'future' }, { text: 'read', value: 'read' }])
  assert.equal(permission.onFilter('future', table.props.dataSource[0]), true)
  assert.equal(permission.onFilter('read', table.props.dataSource[0]), false)
  assert.equal(permission.onFilter('READ', table.props.dataSource[1]), false)
  assert.equal(JSON.stringify(value), snapshot)
  assert.deepEqual(view.RgwPermissions({ value: [], subusers }).props.columns[1].filters, [])
}
const pages = readFileSync(new URL('../src/pages/object/RgwUserDetails.tsx', import.meta.url), 'utf8')
assert.ok(pages.includes('<RgwPermissions value={row.caps} />'))
assert.ok(pages.includes('<RgwPermissions value={row.subusers} subusers />'))
console.log('RGW capabilities and subusers preserve native identity and permission strings')

const actionSource = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const caps = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwUserCapsForm.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(caps)
const capRow = { uid: 'tenant$ns$user', caps: [{ type: 'users', perm: '*' }, { type: 'buckets', perm: 'read' }] }
assert.equal(caps.rgwCapabilityOptions(undefined, 'add').length, 17)
assert.deepEqual(caps.rgwCapabilityOptions(capRow, 'replace').map(item => item.value), ['users', 'buckets'])
for (const value of [undefined, null, {}, [null], [{ type: 'users', perm: 'unknown' }], [...capRow.caps, capRow.caps[0]]]) assert.deepEqual(caps.rgwCapabilityOptions({ caps: value }, 'replace'), [])
assert.throws(() => caps.rgwCapabilityInput({ action: 'replace', type: 'usage', permission: 'read' }, capRow))
assert.throws(() => caps.rgwCapabilityInput({ action: 'unknown', type: 'users', permission: 'read' }, capRow))
assert.throws(() => caps.rgwCapabilityInput({ action: 'add', type: 'users', permission: 'unknown' }, capRow))
const source = ts.createSourceFile('pages.tsx', actionSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let formAction
function visit(node) {
  if (ts.isObjectLiteralExpression(node)) {
    const property = name => node.properties.find(item => ts.isPropertyAssignment(item) && item.name.getText(source) === name)?.initializer
    if (property('title')?.text === '管理用户权限（caps）') {
      const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
      formAction = new Function('userId', ...Object.keys(caps), `${code}; return action`)(row => row.uid, ...Object.values(caps))
    }
  }
  ts.forEachChild(node, visit)
}
visit(source)
const confirmation = formAction.confirmation
assert.equal(formAction.path, '/rgw/user/caps')
assert.equal(formAction.method, 'POST')
assert.deepEqual(formAction.fields.find(field => field.name === 'type').optionsDependencies, ['action'])
assert.deepEqual(await formAction.fields.find(field => field.name === 'type').optionsLoader(7, capRow, { action: 'replace' }), caps.rgwCapabilityOptions(capRow, 'replace'))
assert.deepEqual(formAction.buildBody({ action: 'replace', type: 'users', permission: 'write' }, 7, capRow), { cluster_id: 7, uid: capRow.uid, action: 'replace', type: 'users', permission: 'write' })
const replacement = confirmation({ action: 'replace', type: 'users', permission: 'write' }, capRow)
for (const text of ['整体替换', '先移除', '非原子', '不要盲目重试', 'tenant$ns$user', 'users', 'write']) assert.ok(replacement.includes(text))
for (const [action, wording] of [['add', '合并添加'], ['rm', '仅移除']]) {
  const text = confirmation({ action, type: 'users', permission: 'write' }, capRow)
  for (const expected of [wording, 'tenant$ns$user', 'users', 'write', '不是整体替换', '其它权限保留', '其他用户的数据']) assert.ok(text.includes(expected))
}
console.log('RGW capability confirmation identifies scoped additive and subtractive semantics')

let deletion
function findDeletion(node) {
  if (ts.isObjectLiteralExpression(node) && node.properties.some(item => ts.isPropertyAssignment(item) && item.name.getText(source) === 'title' && item.initializer.text === '删除 RGW 用户')) {
    const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    deletion = new Function('userId', `${code}; return action`)(row => row.uid)
  }
  ts.forEachChild(node, findDeletion)
}
findDeletion(source)
assert.equal(deletion.path, '/rgw/user')
assert.deepEqual(deletion.buildBody(capRow, 7), { cluster_id: 7, uid: capRow.uid })
assert.equal(deletion.resourceKey(capRow), `rgw/user/${capRow.uid}`)
for (const text of [capRow.uid, '凭据将失效', '不可恢复', '不清理 Bucket', 'Ceph 将拒绝删除']) assert.ok(deletion.confirmation(capRow).includes(text))
console.log('RGW user deletion confirms exact UID and native non-purging behavior')

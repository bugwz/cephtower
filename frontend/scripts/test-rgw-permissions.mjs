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

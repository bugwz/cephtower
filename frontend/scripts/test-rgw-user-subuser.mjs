import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const helpers = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwUserSubuser.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(helpers)
const row = { uid: 'tenant$ns$user', subusers: [{ id: 'tenant$ns$user:swift' }, { id: 'other:swift' }, { id: 'tenant$ns$user:bad:name' }, { id: 'tenant$ns$user:dup' }, { id: 'tenant$ns$user:dup' }] }
assert.deepEqual(helpers.rgwSubuserOptions(row), [{ label: 'tenant$ns$user:swift', value: 'swift' }])
for (const value of [undefined, {}, { uid: 'user', subusers: null }, { uid: 'user', subusers: [null, {}, 1] }]) assert.deepEqual(helpers.rgwSubuserOptions(value), [])
const values = { action: 'modify', subuser: 'swift', subuser_permission: 'read', confirm_subuser: row.uid + ':swift' }
for (const subuser_permission of ['none', 'read', 'write', 'readwrite', 'full']) assert.deepEqual(helpers.rgwSubuserInput({ ...values, subuser_permission }, row), { ...values, subuser_permission })
assert.deepEqual(helpers.rgwSubuserInput({ ...values, action: 'rm' }, row), { action: 'rm', subuser: 'swift', confirm_subuser: values.confirm_subuser })
for (const change of [{ action: 'create' }, { subuser: 'other:swift' }, { subuser: 'dup' }, { confirm_subuser: '' }, { confirm_subuser: 'other:swift' }, { subuser_permission: 'read-write' }, { subuser_permission: undefined }]) assert.throws(() => helpers.rgwSubuserInput({ ...values, ...change }, row))
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const source = ts.createSourceFile('pages.tsx', pages, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let actionNode
function visit(node) {
  if (ts.isObjectLiteralExpression(node) && node.properties.some(prop => ts.isPropertyAssignment(prop) && prop.name.getText(source) === 'title' && ts.isStringLiteral(prop.initializer) && prop.initializer.text === '管理已有子用户')) actionNode = node
  ts.forEachChild(node, visit)
}
visit(source)
assert.ok(actionNode)
const code = ts.transpileModule(`const action = ${actionNode.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const action = new Function(...Object.keys(helpers), 'userId', `${code}; return action`)(...Object.values(helpers), row => row.uid)
assert.equal(action.path, '/rgw/user/subuser')
assert.equal(action.method, 'POST')
assert.equal(action.disabledWhen(row), undefined)
assert.ok(action.disabledWhen({}))
assert.deepEqual(action.buildBody(values, 7, row), { cluster_id: 7, uid: row.uid, ...values })
assert.deepEqual(action.buildBody({ ...values, action: 'rm' }, 7, row), { cluster_id: 7, uid: row.uid, action: 'rm', subuser: 'swift', confirm_subuser: values.confirm_subuser })
assert.ok(action.confirmation(values, row).includes(values.confirm_subuser))
assert.match(action.confirmation({ ...values, action: 'rm' }, row), /S3、Swift.*不可恢复/)
for (const field of ['subuser', 'action']) assert.deepEqual(action.changedValues({ [field]: 'x' }), { confirm_subuser: undefined, subuser_permission: undefined })
assert.deepEqual(action.changedValues({ subuser_permission: 'read' }), {})
assert.deepEqual(await action.fields.find(field => field.name === 'subuser').optionsLoader(7, row), helpers.rgwSubuserOptions(row))
assert.equal(action.fields.find(field => field.name === 'subuser_permission').visibleWhen({ action: 'rm' }), false)
console.log('RGW subuser scope, subuser_permission, confirmation and action binding checks passed')

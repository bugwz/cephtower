import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const load = (path, require) => {
  const exports = {}
  new Function('exports', 'require', ts.transpileModule(readFileSync(new URL(path, import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(exports, require)
  return exports
}
const subusers = load('../src/pages/object/rgwUserSubuser.ts')
const helpers = load('../src/pages/object/rgwSubuserCreate.ts', () => subusers)
const row = { uid: 'tenant$ns$user', subusers: [] }
const values = { subuser: 'swift', key_type: 's3', subuser_permission: 'read', access_key: 'ACCESS123', secret_key: 'Private+/=secret', credentials_saved: 'saved', confirm_subuser: row.uid + ':swift' }
for (const kind of ['s3', 'swift']) {
  const input = helpers.rgwSubuserCreateInput({ ...values, key_type: kind }, row)
  assert.equal(input.action, 'create')
  assert.equal(input.secret_key, values.secret_key)
  assert.equal(input.access_key, kind === 's3' ? values.access_key : undefined)
  assert.equal(input.credentials_saved, undefined)
}
for (const change of [{ subuser: 'user:swift' }, { subuser: '-option' }, { subuser: 'bad/name' }, { secret_key: '' }, { secret_key: ' a' }, { secret_key: 'a\tb' }, { secret_key: 's'.repeat(257) }, { credentials_saved: undefined }, { access_key: 'bad/key' }, { key_type: 'other' }, { confirm_subuser: 'other:swift' }, { subuser_permission: 'full-control' }]) assert.throws(() => helpers.rgwSubuserCreateInput({ ...values, ...change }, row))
for (const invalidRow of [{ uid: row.uid }, { uid: row.uid, subusers: [{ id: values.confirm_subuser }] }, undefined]) assert.throws(() => helpers.rgwSubuserCreateInput(values, invalidRow))
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const source = ts.createSourceFile('pages.tsx', pages, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let node
function visit(item) {
  if (ts.isObjectLiteralExpression(item) && item.properties.some(prop => ts.isPropertyAssignment(prop) && prop.name.getText(source) === 'title' && ts.isStringLiteral(prop.initializer) && prop.initializer.text === '创建子用户')) node = item
  ts.forEachChild(item, visit)
}
visit(source)
assert.ok(node)
const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.React } }).outputText
const action = new Function('rgwSubuserCreateInput', 'rgwSubuserPermissionOptions', 'userId', `${code}; return action`)(helpers.rgwSubuserCreateInput, subusers.rgwSubuserPermissionOptions, row => row.uid)
assert.equal(action.path, '/rgw/user/subuser')
assert.equal(action.method, 'POST')
assert.deepEqual(action.buildBody(values, 7, row), { cluster_id: 7, uid: row.uid, ...helpers.rgwSubuserCreateInput(values, row) })
const confirmation = action.confirmation(values, row)
assert.ok(confirmation.includes(values.confirm_subuser))
assert.ok(!confirmation.includes(values.secret_key) && !confirmation.includes(values.access_key))
for (const field of ['access_key', 'secret_key']) assert.equal(action.fields.find(item => item.name === field).type, 'password')
for (const field of ['subuser', 'key_type', 'access_key', 'secret_key']) assert.deepEqual(action.changedValues({ [field]: 'new' }), { credentials_saved: undefined, confirm_subuser: undefined })
assert.equal(action.fields.find(item => item.name === 'access_key').visibleWhen({ key_type: 'swift' }), false)
console.log('RGW subuser creation preserves protocol, explicit saved credentials and secret-free confirmation')

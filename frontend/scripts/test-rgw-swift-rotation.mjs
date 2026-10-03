import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const load = (path, require) => {
  const exports = {}
  new Function('exports', 'require', ts.transpileModule(readFileSync(new URL(path, import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(exports, require)
  return exports
}
const subusers = load('../src/pages/object/rgwUserSubuser.ts')
const helpers = load('../src/pages/object/rgwSwiftKeyRotation.ts', () => subusers)
const row = { uid: 'tenant$ns$user', subusers: [{ id: 'tenant$ns$user:sub' }], swift_keys: [{ user: 'tenant$ns$user:sub', active: false }] }
const values = { subuser: 'sub', confirm_subuser: row.uid + ':sub', secret_key: 'NEWsecret', credentials_saved: 'saved' }
for (const active of [false, true]) {
  const scoped = { ...row, swift_keys: [{ user: values.confirm_subuser, active }] }
  assert.equal(helpers.rgwSwiftRotationOptions(scoped)[0].active, active)
  assert.deepEqual(helpers.rgwSwiftRotationInput(values, scoped), { action: 'rotate-swift-key', subuser: 'sub', confirm_subuser: values.confirm_subuser, expected_key_active: active, secret_key: 'NEWsecret' })
}
for (const keys of [undefined, null, [], [{ user: 'other:sub', active: true }], [{ user: values.confirm_subuser }], [{ user: values.confirm_subuser, active: 'false' }], [...row.swift_keys, ...row.swift_keys]]) {
  assert.deepEqual(helpers.rgwSwiftRotationOptions({ ...row, swift_keys: keys }), [])
  assert.throws(() => helpers.rgwSwiftRotationInput(values, { ...row, swift_keys: keys }))
}
for (const change of [{ subuser: 'other' }, { confirm_subuser: '' }, { credentials_saved: undefined }, { secret_key: '' }, { secret_key: ' new' }, { secret_key: 's'.repeat(257) }]) assert.throws(() => helpers.rgwSwiftRotationInput({ ...values, ...change }, row))
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const source = ts.createSourceFile('pages.tsx', pages, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let node
function visit(item) {
  if (ts.isObjectLiteralExpression(item) && item.properties.some(prop => ts.isPropertyAssignment(prop) && prop.name.getText(source) === 'title' && ts.isStringLiteral(prop.initializer) && prop.initializer.text === '轮换 Swift 子用户密钥')) node = item
  ts.forEachChild(item, visit)
}
visit(source)
assert.ok(node)
const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.React } }).outputText
const action = new Function(...Object.keys(helpers), 'userId', `${code}; return action`)(...Object.values(helpers), row => row.uid)
assert.equal(action.path, '/rgw/user/subuser')
assert.equal(action.method, 'POST')
assert.equal(action.disabledWhen(row), undefined)
assert.ok(action.disabledWhen({}))
assert.deepEqual(action.buildBody(values, 7, row), { cluster_id: 7, uid: row.uid, ...helpers.rgwSwiftRotationInput(values, row) })
assert.equal(action.fields.find(item => item.name === 'secret_key').type, 'password')
const text = action.confirmation(values, row)
assert.ok(text.includes(values.confirm_subuser) && text.includes('旧凭据将失效') && text.includes('已停用') && !text.includes(values.secret_key))
assert.deepEqual(await action.fields[0].optionsLoader(7, row), helpers.rgwSwiftRotationOptions(row))
for (const field of ['subuser', 'secret_key']) assert.deepEqual(action.changedValues({ [field]: 'new' }), { credentials_saved: undefined, confirm_subuser: undefined })
console.log('Swift rotation preserves active state, exact subuser scope and secret-free confirmation')

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const helpers = {}, calls = []
const credentials = {}
const flags = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwUserCreateFlags.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(flags)
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwUserCreateCredentials.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(credentials)
new Function('exports', 'require', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwUserCreateAccount.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(helpers, () => ({ loadRgwMigrationAccountOptions: async (...args) => { calls.push(args); return [] } }))
const values = { uid: 'tenant$user', account_mode: 'account', account_id: 'RGW12345678901234567', account_root: 'disable', display_name: 'valid-name', credential_mode: 'none', system: 'disable', suspended: 'disable' }
assert.deepEqual(helpers.rgwUserCreateAccountInput({}), {})
assert.deepEqual(helpers.rgwUserCreateAccountInput({ ...values, account_mode: 'independent' }), {})
for (const root of ['enable', 'disable']) assert.deepEqual(helpers.rgwUserCreateAccountInput({ ...values, account_root: root }), { account_id: values.account_id, account_root: root === 'enable' })
for (const patch of [{ account_mode: null }, { account_id: '' }, { account_id: values.account_id + '\n' }, { account_root: false }, { display_name: 'invalid name' }, { display_name: 'valid\n' }]) assert.throws(() => helpers.rgwUserCreateAccountInput({ ...values, ...patch }))
for (const [uid, tenant] of [['user', ''], ['tenant$user', 'tenant'], ['tenant$ns$user', 'tenant'], ['$ns$user', '']]) {
  await helpers.loadRgwCreateAccountOptions(7, { ...values, uid })
  assert.deepEqual(calls.pop(), [7, { tenant }])
}
await helpers.loadRgwCreateAccountOptions(7, {})
assert.equal(calls.length, 0)
const source = ts.createSourceFile('pages.tsx', readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8'), ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let node
function visit(value) {
  if (ts.isObjectLiteralExpression(value) && value.properties.some(prop => ts.isPropertyAssignment(prop) && prop.name.getText(source) === 'title' && ts.isStringLiteral(prop.initializer) && prop.initializer.text === '新建 RGW 用户')) node = value
  ts.forEachChild(value, visit)
}
visit(source)
assert.ok(node)
const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.React } }).outputText
const action = new Function('loadRgwCreateAccountOptions', 'rgwUserCreateAccountInput', 'rgwBucketLimitInput', 'rgwUserCreateCredentials', 'rgwUserCreateFlags', `${code}; return action`)(helpers.loadRgwCreateAccountOptions, helpers.rgwUserCreateAccountInput, () => ({}), credentials.rgwUserCreateCredentials, flags.rgwUserCreateFlags)
assert.deepEqual(action.buildBody(values, 7), { cluster_id: 7, uid: values.uid, display_name: values.display_name, account_id: values.account_id, account_root: false, system: false, suspended: false })
assert.ok(action.confirmation(values).includes(values.account_id))
assert.ok(action.confirmation({ ...values, account_root: 'enable' }).includes('根用户权限'))
assert.ok(action.confirmation({ ...values, account_mode: 'independent' }).includes('不创建访问密钥'))
assert.deepEqual(action.changedValues({ uid: 'changed' }), { account_id: undefined, account_root: undefined, credentials_saved: undefined })
assert.deepEqual(action.changedValues({ account_mode: 'changed' }), { account_id: undefined, account_root: undefined })
assert.deepEqual(action.changedValues({ email: 'new' }), {})
const accountField = action.fields.find(field => field.name === 'account_id')
assert.deepEqual(accountField.optionsDependencies, ['uid', 'account_mode'])
assert.equal(accountField.visibleWhen(values), true)
assert.equal(accountField.visibleWhen({ account_mode: 'independent' }), false)
await accountField.optionsLoader(9, undefined, values)
assert.deepEqual(calls.pop(), [9, { tenant: 'tenant' }])
console.log('RGW account user creation preserves explicit root choice and tenant-scoped options')
const pair = { ...values, credential_mode: 's3', access_key: 'ACCESS123', secret_key: 'SavedSecret', credentials_saved: 'saved' }
assert.deepEqual(credentials.rgwUserCreateCredentials(pair), { access_key: pair.access_key, secret_key: pair.secret_key })
assert.deepEqual(credentials.rgwUserCreateCredentials({ ...pair, credential_mode: 'none' }), {})
for (const patch of [{ credential_mode: undefined }, { access_key: '' }, { secret_key: '' }, { secret_key: ' padded' }, { secret_key: 's'.repeat(257) }, { credentials_saved: undefined }]) assert.throws(() => credentials.rgwUserCreateCredentials({ ...pair, ...patch }))
assert.equal(action.buildBody(pair, 7).secret_key, pair.secret_key)
assert.ok(!action.confirmation(pair).includes(pair.secret_key) && !action.confirmation(pair).includes(pair.access_key))
for (const field of ['credential_mode', 'access_key', 'secret_key']) assert.deepEqual(action.changedValues({ [field]: 'changed' }), { credentials_saved: undefined })
for (const field of ['access_key', 'secret_key', 'credentials_saved']) {
  const definition = action.fields.find(item => item.name === field)
  assert.equal(definition.visibleWhen(pair), true)
  assert.equal(definition.visibleWhen(values), false)
}
console.log('RGW user creation requires explicit key mode and saved credentials')
for (const system of ['enable', 'disable']) for (const suspended of ['enable', 'disable']) {
  const input = { ...values, system, suspended }
  assert.deepEqual(flags.rgwUserCreateFlags(input), { system: system === 'enable', suspended: suspended === 'enable' })
  assert.equal(action.buildBody(input, 7).system, system === 'enable')
  assert.equal(action.buildBody(input, 7).suspended, suspended === 'enable')
  const text = action.confirmation(input)
  assert.ok(text.includes(system === 'enable' ? '授予 RGW 内部系统操作能力' : '不启用系统用户标志'))
  assert.ok(text.includes(suspended === 'enable' ? '第二步失败' : '保持启用'))
}
for (const field of ['system', 'suspended']) for (const value of [undefined, null, true, 'keep']) assert.throws(() => flags.rgwUserCreateFlags({ ...values, [field]: value }))
console.log('RGW initial flags preserve explicit false and explain non-atomic suspension')

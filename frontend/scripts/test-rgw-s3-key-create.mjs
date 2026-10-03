import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const load = (path, require) => {
  const exports = {}
  new Function('exports', 'require', ts.transpileModule(readFileSync(new URL(path, import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(exports, require)
  return exports
}
const subusers = load('../src/pages/object/rgwUserSubuser.ts')
const helpers = load('../src/pages/object/rgwS3KeyCreate.ts', () => subusers)
const row = { uid: 'tenant$ns$user', subusers: [{ id: 'tenant$ns$user:sub' }, { id: 'other:sub' }] }
assert.deepEqual(helpers.rgwS3KeyOwnerOptions(row).map(option => option.value), [row.uid, row.uid + ':sub'])
assert.deepEqual(helpers.rgwS3KeyOwnerOptions(undefined), [])
const base = { access_key: 'ACCESS123', secret_key: 'Private+/=secret', credentials_saved: 'saved' }
for (const owner of [row.uid, row.uid + ':sub']) {
  const input = helpers.rgwS3KeyCreateInput({ ...base, owner, confirm_owner: owner }, row)
  assert.equal(input.subuser, owner === row.uid ? undefined : 'sub')
  assert.equal(input.confirm_owner, owner)
  assert.equal(input.secret_key, base.secret_key)
  assert.equal(input.credentials_saved, undefined)
}
const values = { ...base, owner: row.uid, confirm_owner: row.uid }
for (const change of [{ owner: 'other' }, { confirm_owner: '' }, { access_key: 'bad/key' }, { secret_key: '' }, { secret_key: ' bad' }, { secret_key: 's'.repeat(257) }, { credentials_saved: undefined }]) assert.throws(() => helpers.rgwS3KeyCreateInput({ ...values, ...change }, row))
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const source = ts.createSourceFile('pages.tsx', pages, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let node
function visit(item) {
  if (ts.isObjectLiteralExpression(item) && item.properties.some(prop => ts.isPropertyAssignment(prop) && prop.name.getText(source) === 'title' && ts.isStringLiteral(prop.initializer) && prop.initializer.text === '创建 S3 访问密钥')) node = item
  ts.forEachChild(item, visit)
}
visit(source)
assert.ok(node)
const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const action = new Function(...Object.keys(helpers), 'userId', `${code}; return action`)(...Object.values(helpers), row => row.uid)
assert.equal(action.path, '/rgw/user/key')
assert.equal(action.method, 'POST')
for (const owner of [row.uid, row.uid + ':sub']) {
  const input = { ...base, owner, confirm_owner: owner }
  assert.deepEqual(action.buildBody(input, 7, row), { cluster_id: 7, uid: row.uid, ...helpers.rgwS3KeyCreateInput(input, row) })
  const text = action.confirmation(input, row)
  assert.ok(text.includes(owner) && !text.includes(base.secret_key) && !text.includes(base.access_key))
}
for (const field of ['access_key', 'secret_key']) assert.equal(action.fields.find(item => item.name === field).type, 'password')
assert.deepEqual(await action.fields[0].optionsLoader(7, row), helpers.rgwS3KeyOwnerOptions(row))
for (const field of ['owner', 'access_key', 'secret_key']) assert.deepEqual(action.changedValues({ [field]: 'new' }), { credentials_saved: undefined, confirm_owner: undefined })
console.log('S3 key creation validates exact owner selection and saved secret-free confirmation')

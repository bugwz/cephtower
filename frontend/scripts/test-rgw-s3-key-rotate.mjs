import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const load = (path, require) => {
  const exports = {}
  new Function('exports', 'require', ts.transpileModule(readFileSync(new URL(path, import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(exports, require)
  return exports
}
const subusers = load('../src/pages/object/rgwUserSubuser.ts')
const create = load('../src/pages/object/rgwS3KeyCreate.ts', () => subusers)
const remove = load('../src/pages/object/rgwS3KeyDelete.ts', () => create)
const helpers = load('../src/pages/object/rgwS3KeyRotate.ts', name => name.endsWith('Create') ? create : remove)
const row = { uid: 'tenant$ns$user', subusers: [{ id: 'tenant$ns$user:sub' }], keys: [{ user: 'tenant$ns$user' }, { user: 'tenant$ns$user:sub' }] }
const values = { access_key: 'ACCESS123', secret_key: 'NEWsecret', owner: row.uid, confirm_owner: row.uid, credentials_saved: 'saved' }
for (const change of [{ owner: 'other' }, { confirm_owner: '' }, { access_key: ' key' }, { secret_key: '' }, { secret_key: 'secret\n' }, { credentials_saved: undefined }]) assert.throws(() => helpers.rgwS3KeyRotateInput({ ...values, ...change }, row))
assert.throws(() => helpers.rgwS3KeyRotateInput(values, { ...row, keys: [] }))
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const source = ts.createSourceFile('pages.tsx', pages, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let node
function visit(item) {
  if (ts.isObjectLiteralExpression(item) && item.properties.some(prop => ts.isPropertyAssignment(prop) && prop.name.getText(source) === 'title' && ts.isStringLiteral(prop.initializer) && prop.initializer.text === '轮换 S3 访问密钥')) node = item
  ts.forEachChild(item, visit)
}
visit(source)
assert.ok(node)
const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const action = new Function('rgwS3KeyRotateInput', 'rgwS3KeyDeleteOptions', 'userId', `${code}; return action`)(helpers.rgwS3KeyRotateInput, remove.rgwS3KeyDeleteOptions, row => row.uid)
assert.equal(action.method, 'PATCH')
assert.equal(action.path, '/rgw/user/key')
assert.ok(action.disabledWhen({}))
for (const owner of [row.uid, row.uid + ':sub']) {
  const input = { ...values, owner, confirm_owner: owner }
  const body = action.buildBody(input, 7, row)
  assert.deepEqual(body, { cluster_id: 7, uid: row.uid, access_key: values.access_key, secret_key: values.secret_key, confirm_owner: owner, ...(owner === row.uid ? {} : { subuser: 'sub' }) })
  const text = action.confirmation(input, row)
  assert.ok(text.includes(owner) && text.includes('原激活状态保持不变') && !text.includes(values.secret_key) && !text.includes(values.access_key))
}
for (const field of ['access_key', 'secret_key']) assert.equal(action.fields.find(item => item.name === field).type, 'password')
assert.deepEqual(await action.fields[0].optionsLoader(7, row), remove.rgwS3KeyDeleteOptions(row))
for (const field of ['owner', 'access_key', 'secret_key']) assert.deepEqual(action.changedValues({ [field]: 'new' }), { credentials_saved: undefined, confirm_owner: undefined })
console.log('S3 key rotation requires existing owner metadata and explicit saved credentials')

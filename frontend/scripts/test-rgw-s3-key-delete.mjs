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
const helpers = load('../src/pages/object/rgwS3KeyDelete.ts', () => create)
const row = { uid: 'tenant$ns$user', subusers: [{ id: 'tenant$ns$user:sub' }], keys: [{ user: 'tenant$ns$user' }, { user: 'tenant$ns$user:sub' }] }
assert.deepEqual(helpers.rgwS3KeyDeleteOptions(row).map(option => option.value), [row.uid, row.uid + ':sub'])
for (const keys of [undefined, null, [], [null, {}], [{ user: 'other:sub' }]]) assert.deepEqual(helpers.rgwS3KeyDeleteOptions({ ...row, keys }), [])
const values = { access_key: 'ACCESS123', owner: row.uid, confirm_owner: row.uid }
for (const change of [{ owner: 'other' }, { confirm_owner: '' }, { access_key: ' key' }, { access_key: 'bad/key' }, { access_key: undefined }]) assert.throws(() => helpers.rgwS3KeyDeleteInput({ ...values, ...change }, row))
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const source = ts.createSourceFile('pages.tsx', pages, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let node
function visit(item) {
  if (ts.isObjectLiteralExpression(item) && item.properties.some(prop => ts.isPropertyAssignment(prop) && prop.name.getText(source) === 'title' && ts.isStringLiteral(prop.initializer) && prop.initializer.text === '删除 S3 访问密钥')) node = item
  ts.forEachChild(item, visit)
}
visit(source)
assert.ok(node)
const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const action = new Function(...Object.keys(helpers), 'userId', `${code}; return action`)(...Object.values(helpers), row => row.uid)
assert.equal(action.path, '/rgw/user/key')
assert.equal(action.method, 'DELETE')
assert.equal(action.disabledWhen(row), undefined)
assert.ok(action.disabledWhen({}))
for (const owner of [row.uid, row.uid + ':sub']) {
  const input = { ...values, owner, confirm_owner: owner }
  const body = action.buildBody(input, 7, row)
  assert.deepEqual(body, { cluster_id: 7, uid: row.uid, access_key: values.access_key, confirm_owner: owner, ...(owner === row.uid ? {} : { subuser: 'sub' }) })
  const text = action.confirmation(input, row)
  assert.ok(text.includes(owner) && text.includes('不可恢复') && !text.includes(values.access_key))
}
assert.equal(action.fields.find(item => item.name === 'access_key').type, 'password')
assert.deepEqual(await action.fields[0].optionsLoader(7, row), helpers.rgwS3KeyDeleteOptions(row))
for (const field of ['owner', 'access_key']) assert.deepEqual(action.changedValues({ [field]: 'new' }), { confirm_owner: undefined })
console.log('S3 key deletion preserves scoped owner selection and hidden Access Key confirmation')

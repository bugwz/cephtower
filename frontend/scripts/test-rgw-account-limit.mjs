import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const exports = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwAccountLimit.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const limit = exports.rgwAccountLimit
const patch = exports.rgwAccountLimitPatch
assert.deepEqual(patch({}), {})
assert.deepEqual(patch({ max_users: 5, max_buckets: null }, { max_users: 5 }), {})
assert.deepEqual(patch({ max_users: '', max_roles: undefined }), {})
for (const key of ['max_users', 'max_roles', 'max_groups', 'max_buckets', 'max_access_keys']) {
  for (const value of [-1, 0, 1, 2147483647]) assert.deepEqual(patch({ [key]: value }), { [key]: value })
  for (const value of [-2, 0.5, '0', true, NaN, Infinity, 2147483648]) assert.throws(() => patch({ [key]: value }))
}
assert.deepEqual(patch({ max_users: 5, max_roles: 6 }, { max_users: 5, max_roles: 2 }), { max_roles: 6 })
for (const value of [-1, -2, -2147483648]) assert.equal(limit(value), '无限制')
assert.equal(limit(0), '0（禁止新增）')
assert.equal(limit(2), '2')
assert.equal(limit(2147483647), '2147483647')
for (const value of [undefined, null, '0', NaN, Infinity, 0.5, 2147483648, -2147483649]) assert.equal(limit(value), '资源上限未返回或无效')
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.equal(pages.match(/render: rgwAccountLimit/g).length, 4)
assert.ok(pages.includes("title: '每用户访问密钥上限'"))
assert.ok(pages.includes('...rgwAccountLimitPatch(values, row)'))
const source = ts.createSourceFile('pages.tsx', pages, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let fieldsExpression
let bodyExpression
let createExpression
let createFields
function visit(node) {
  if (ts.isObjectLiteralExpression(node) && node.properties.some(property => ts.isPropertyAssignment(property) && property.name.getText(source) === 'title' && property.initializer.getText(source) === "'新建 RGW Account'")) {
    createExpression = node.properties.find(property => ts.isPropertyAssignment(property) && property.name.getText(source) === 'buildBody').initializer.getText(source)
    createFields = node.properties.find(property => ts.isPropertyAssignment(property) && property.name.getText(source) === 'fields').initializer.getText(source)
  }
  if (ts.isObjectLiteralExpression(node) && node.properties.some(property => ts.isPropertyAssignment(property) && property.name.getText(source) === 'title' && property.initializer.getText(source) === "'编辑 RGW Account'")) {
    fieldsExpression = node.properties.find(property => ts.isPropertyAssignment(property) && property.name.getText(source) === 'fields').initializer.getText(source)
    bodyExpression = node.properties.find(property => ts.isPropertyAssignment(property) && property.name.getText(source) === 'buildBody').initializer.getText(source)
  }
  ts.forEachChild(node, visit)
}
visit(source)
assert.ok(createExpression)
const createBody = new Function('rgwAccountLimitPatch', `return (${createExpression})`)(patch)
assert.deepEqual(createBody({ account_id: 'RGW123' }, 'cluster'), { cluster_id: 'cluster', account_id: 'RGW123' })
const createFieldExports = {}
new Function('exports', ts.transpileModule('export const fields = ' + createFields, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(createFieldExports)
for (const key of ['max_users', 'max_roles', 'max_groups', 'max_buckets', 'max_access_keys']) {
  const field = createFieldExports.fields.find(field => field.name === key)
  assert.equal(field.type, 'number')
  assert.equal(field.min, -1)
  assert.equal(field.max, 2147483647)
  assert.ok(field.label.includes('留空使用默认值'))
  for (const value of [-1, 0, 2147483647]) assert.deepEqual(createBody({ account_id: 'RGW123', [key]: value }, 'cluster'), { cluster_id: 'cluster', account_id: 'RGW123', [key]: value })
  for (const value of [undefined, null, '']) assert.deepEqual(createBody({ account_id: 'RGW123', [key]: value }, 'cluster'), { cluster_id: 'cluster', account_id: 'RGW123' })
  for (const value of [-2, 0.5, 2147483648, false, '0']) assert.throws(() => createBody({ account_id: 'RGW123', [key]: value }, 'cluster'))
}
const editExports = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwAccountEdit.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(editExports)
const buildBody = new Function('rgwAccountTextPatch', 'rgwAccountLimitPatch', `return (${bodyExpression})`)(editExports.rgwAccountTextPatch, patch)
const row = { account_id: 'RGW123', account_name: 'Original', email: 'old@example.test', max_users: 10 }
for (const values of [{}, { account_name: row.account_name, email: row.email, max_users: 10 }]) assert.throws(() => buildBody(values, 'cluster', row), /没有需要提交/)
assert.deepEqual(buildBody({ max_users: 0 }, 'cluster', row), { cluster_id: 'cluster', account_id: 'RGW123', max_users: 0 })
for (const key of ['account_name', 'email']) {
  assert.deepEqual(buildBody({ [key]: 'New' }, 'cluster', row), { cluster_id: 'cluster', account_id: 'RGW123', [key]: 'New' })
  for (const value of ['', null, false, 1, {}, '  ', 'bad\ntext', 'bad\0text']) assert.throws(() => buildBody({ [key]: value, max_users: 0 }, 'cluster', row), /非空单行文本/)
}
assert.deepEqual(buildBody({ account_name: '', email: '', max_users: -1 }, 'cluster', { natural_key: 'RGW456' }), { cluster_id: 'cluster', account_id: 'RGW456', max_users: -1 })
assert.ok(fieldsExpression)
const fieldExports = {}
new Function('exports', ts.transpileModule('export const fields = ' + fieldsExpression, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(fieldExports)
for (const name of ['max_users', 'max_roles', 'max_groups', 'max_buckets', 'max_access_keys']) {
  const field = fieldExports.fields.find(field => field.name === name)
  assert.equal(field.type, 'number')
  assert.equal(field.min, -1)
  assert.equal(field.max, 2147483647)
}
const initializer = source.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === 'numberOrUndefined')
const initialExports = {}
new Function('exports', ts.transpileModule('export ' + initializer.getText(source), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(initialExports)
for (const value of [null, undefined, '', '0', false, 0.5, 2147483648]) assert.equal(initialExports.numberOrUndefined(value), undefined)
for (const value of [-1, 0, 2147483647]) assert.equal(initialExports.numberOrUndefined(value), value)
console.log('RGW IAM limits preserve zero creation bans and negative unlimited values')

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
import './test-external-form-confirmation.mjs'
const helpers = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketConfiguration.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(helpers)
for (const [kind, document] of [['policy', '{"Statement":[],"large":9007199254740993}'], ['cors', '<CORSConfiguration/>'], ['lifecycle', '<LifecycleConfiguration/>'], ['encryption', '<ServerSideEncryptionConfiguration/>']]) {
  const input = { bucket_id: 'AGJ1Y2tldA', kind, document }
  assert.deepEqual(helpers.rgwBucketConfigurationInput(input), input)
}
for (const values of [{ kind: 'cors', document: '{}' }, { kind: 'policy', document: 'null' }, { kind: 'policy', document: '[]' }, { kind: 'policy', document: '{' }, { kind: 'unknown', document: '{}' }, { kind: 'policy', document: '' }]) assert.throws(() => helpers.rgwBucketConfigurationInput({ bucket_id: 'id', ...values }))
assert.throws(() => helpers.rgwBucketConfigurationInput({ bucket_id: 'id', kind: 'cors', document: '<CORSConfiguration>' + 'x'.repeat(1024 * 1024) + '</CORSConfiguration>' }))
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const source = ts.createSourceFile('pages.tsx', pages, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let definition
function visit(node) {
  if (ts.isObjectLiteralExpression(node) && node.properties.some(item => ts.isPropertyAssignment(item) && item.name.getText(source) === 'title' && item.initializer.text === 'Bucket 配置文档')) {
    const code = ts.transpileModule(`const definition = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.React } }).outputText
    definition = new Function(...Object.keys(helpers), `${code}; return definition`)(...Object.values(helpers))
  }
  ts.forEachChild(node, visit)
}
visit(source)
assert.ok(definition)
assert.equal(definition.buildQuery({ kind: 'cors' }).toString(), 'kind=cors')
assert.deepEqual(definition.filterFields.find(field => field.name === 'kind').options, helpers.rgwBucketConfigurationOptions)
assert.deepEqual(definition.columns.map(column => column.key), ['bucket_id', 'kind', 'configured', 'content_type', 'document'])
const status = definition.columns.find(column => column.key === 'configured').render
assert.equal(status(true), '已配置')
assert.equal(status(false), '未配置')
for (const value of [undefined, null, 0, 1, 'false']) assert.equal(status(value), '状态不可用')
const input = { bucket_id: 'id', kind: 'cors', document: '<CORSConfiguration/>' }
assert.deepEqual(definition.createAction.buildBody(input, 7), { cluster_id: 7, ...input })
assert.equal(definition.createAction.method, 'PATCH')
assert.equal(definition.createAction.path, '/rgw/bucket/policy')
assert.equal(definition.updateAction.path, '/rgw/bucket/policy')
assert.equal(definition.updateAction.method, 'PATCH')
for (const kind of ['policy', 'cors', 'lifecycle', 'encryption']) {
  const document = kind === 'policy' ? '{ "Statement":[], "large":9007199254740993 }\n' : '<NativeXML>原文\n  保留格式</NativeXML>'
  const row = { bucket_id: 'AGJ1Y2tldA', kind, configured: true, document }
  assert.equal(definition.updateAction.disabledWhen(row), undefined)
  const values = definition.updateAction.initialValues(row)
  assert.deepEqual(values, { bucket_id: row.bucket_id, kind, document })
  assert.deepEqual(definition.updateAction.buildBody(values, 7, row), { cluster_id: 7, ...values })
  assert.throws(() => definition.updateAction.buildBody({ ...values, bucket_id: 'other' }, 7, row))
  assert.throws(() => definition.updateAction.buildBody({ ...values, kind: kind === 'policy' ? 'cors' : 'policy' }, 7, row))
  for (const action of [definition.createAction, definition.updateAction]) {
    const confirmation = action.confirmation(values, action === definition.updateAction ? row : undefined)
    assert.ok(confirmation.includes(row.bucket_id) && confirmation.includes(kind) && confirmation.includes('整体替换'))
    assert.ok(confirmation.includes('外部并发') && !confirmation.includes(document))
  }
  const absent = { ...row, configured: false, document: null }
  assert.equal(definition.updateAction.disabledWhen(absent), undefined)
  assert.equal(definition.updateAction.initialValues(absent).document, '')
  assert.throws(() => definition.updateAction.buildBody(definition.updateAction.initialValues(absent), 7, absent))
}
for (const row of [{}, { bucket_id: 'id', kind: 'policy' }, { bucket_id: 'id', kind: 'policy', configured: true, document: null }, { bucket_id: 'id', kind: 'policy', configured: false, document: '{}' }, { bucket_id: 'id', kind: 'versioning', configured: false, document: null }]) {
  assert.ok(definition.updateAction.disabledWhen(row))
  assert.throws(() => definition.updateAction.initialValues(row))
}
for (const name of ['bucket_id', 'kind']) assert.equal(definition.updateAction.fields.find(field => field.name === name).readOnly, true)
assert.equal(definition.deleteAction.path, '/rgw/bucket/policy')
assert.equal(definition.deleteAction.action, 'rgw_bucket_policy.delete')
assert.equal(definition.deleteAction.risk, 'high')
for (const kind of ['policy', 'cors', 'lifecycle', 'encryption']) {
  const row = { bucket_id: 'dGVhbQBiaWc', kind, configured: true, document: 'private-document' }
  assert.equal(definition.deleteAction.disabledWhen(row), undefined)
  assert.deepEqual(definition.deleteAction.buildBody(row, 7), { cluster_id: 7, bucket_id: row.bucket_id, kind })
  assert.equal(definition.deleteAction.resourceKey(row), `${row.bucket_id} / ${kind}`)
  const confirmation = definition.deleteAction.confirmation(row)
  assert.ok(confirmation.includes(row.bucket_id) && confirmation.includes(kind))
  assert.ok(confirmation.includes('不会删除 Bucket 或对象') && confirmation.includes('外部并发'))
  assert.ok(!confirmation.includes(row.document))
}
for (const row of [{}, { configured: false }, { configured: 'true' }, { configured: true, bucket_id: 'id', kind: 'versioning' }, { configured: true, bucket_id: ' id', kind: 'policy' }]) {
  assert.ok(definition.deleteAction.disabledWhen(row))
  assert.throws(() => definition.deleteAction.buildBody(row, 7))
}
const externalPage = readFileSync(new URL('../src/pages/ExternalListPage.tsx', import.meta.url), 'utf8')
assert.ok(externalPage.includes('const blocked = action.disabledWhen?.(row)'))
assert.ok(externalPage.includes('Boolean(definition.deleteAction.disabledWhen?.(row))'))
assert.equal((externalPage.match(/content: action\.confirmation\?\.\(row\)/g) ?? []).length, 2)
assert.ok(readFileSync(new URL('../src/pages/ExternalListPage.tsx', import.meta.url), 'utf8').includes('definition.buildQuery?.(queryBody)'))
console.log('Bucket configuration preserves raw JSON/XML and query-scoped reads')

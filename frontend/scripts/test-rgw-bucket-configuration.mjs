import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
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
assert.ok(readFileSync(new URL('../src/pages/ExternalListPage.tsx', import.meta.url), 'utf8').includes('definition.buildQuery?.(queryBody)'))
console.log('Bucket configuration preserves raw JSON/XML and query-scoped reads')

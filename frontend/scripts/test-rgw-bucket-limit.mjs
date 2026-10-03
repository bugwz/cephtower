import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const exports = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketLimit.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const limit = exports.rgwBucketLimit
const input = exports.rgwBucketLimitInput
const patch = exports.rgwBucketLimitPatch
for (const current of [-1, 0, 1, 2147483647]) {
  assert.deepEqual(patch(current, current), {})
  for (const omitted of [undefined, null, '']) assert.deepEqual(patch(omitted, current), {})
}
assert.deepEqual(patch(0, 100), { max_buckets: 0 })
assert.deepEqual(patch(-1, 0), { max_buckets: -1 })
assert.deepEqual(patch(10, undefined), { max_buckets: 10 })
assert.deepEqual(patch(10, 0), { max_buckets: 10 })
for (const value of [-2, 0.5, 2147483648, NaN, Infinity, '0', false]) assert.throws(() => patch(value, 100))
for (const value of [undefined, null, '']) assert.deepEqual(input(value), {})
for (const value of [-1, 0, 1, 2147483647]) assert.deepEqual(input(value), { max_buckets: value })
for (const value of [-2, 0.5, 2147483648, NaN, Infinity, '0', false]) assert.throws(() => input(value))
assert.equal(limit(-1), '禁止创建 Bucket')
assert.equal(limit(-2), '禁止创建 Bucket')
assert.equal(limit(0), '无限制')
assert.equal(limit(1000), '1000')
assert.equal(limit(2147483647), '2147483647')
for (const value of [undefined, null, '0', NaN, Infinity, 0.5, 2147483648, -2147483649]) assert.equal(limit(value), 'Bucket 上限未返回或无效')
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.equal(pages.match(/render: rgwBucketLimit/g).length, 2)
assert.ok(!pages.includes('最大 Bucket 数（-1 为无限制）'))
assert.equal(pages.match(/-1 禁止创建，0 无限制/g).length, 3)
assert.equal(pages.match(/\.\.\.rgwBucketLimitInput\(values.max_buckets\)/g).length, 1)
assert.ok(pages.includes('...rgwBucketLimitPatch(values.max_buckets, row?.max_buckets)'))
assert.ok(!pages.includes('values.max_buckets !== undefined ?'))
const source = ts.createSourceFile('pages.tsx', pages, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let updateBody
function visit(node) {
  if (ts.isObjectLiteralExpression(node) && node.properties.some((property) => ts.isPropertyAssignment(property) && property.name.getText(source) === 'title' && property.initializer.getText(source) === "'更新 RGW 用户'")) {
    updateBody = node.properties.find((property) => ts.isPropertyAssignment(property) && property.name.getText(source) === 'buildBody').initializer.getText(source)
  }
  ts.forEachChild(node, visit)
}
visit(source)
assert.ok(updateBody)
const buildBody = new Function('rgwBucketLimitPatch', 'rgwUserEmailPatch', 'rgwUserFlagPatch', 'userId', `return (${updateBody})`)(patch, () => ({}), () => ({}), (row) => row.uid)
const row = { uid: 'tenant$user', max_buckets: 100 }
assert.deepEqual(buildBody({ max_buckets: 100 }, 'cluster', row), { cluster_id: 'cluster', uid: 'tenant$user' })
assert.deepEqual(buildBody({ max_buckets: 0 }, 'cluster', row), { cluster_id: 'cluster', uid: 'tenant$user', max_buckets: 0 })
assert.deepEqual(buildBody({ max_buckets: null }, 'cluster', row), { cluster_id: 'cluster', uid: 'tenant$user' })
console.log('RGW bucket count limits preserve native disabled and unlimited semantics')

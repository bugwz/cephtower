import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const exports = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketLimit.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const limit = exports.rgwBucketLimit
const input = exports.rgwBucketLimitInput
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
assert.equal(pages.match(/\.\.\.rgwBucketLimitInput\(values.max_buckets\)/g).length, 2)
assert.ok(!pages.includes('values.max_buckets !== undefined ?'))
console.log('RGW bucket count limits preserve native disabled and unlimited semantics')

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const exports = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwRateLimitDetails.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const details = exports.rgwRateLimitDetails
assert.deepEqual(details({ enabled: true, max_read_ops: 0, max_write_ops: -1, max_read_bytes: 1024, max_write_bytes: 7 }), { state: '已启用', limits: ['无限制', '无限制', '1024', '7'] })
assert.deepEqual(details({ enabled: false, max_read_ops: 10 }), { state: '未启用' })
for (const value of [null, undefined, [], '']) assert.equal(details(value).state, '限流信息不可用')
for (const enabled of [null, undefined, 'true', 1]) assert.equal(details({ enabled }).state, '限流启用状态未知')
for (const value of [null, undefined, '0', NaN, Infinity, 0.5, Number.MAX_SAFE_INTEGER + 1]) assert.equal(details({ enabled: true, max_read_ops: value }).limits[0], '未返回或超出精确显示范围')
assert.equal(readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8').match(/<RgwRateLimit value=\{value\} \/>/g).length, 2)
console.log('RGW rate limits distinguish disabled, unlimited and unavailable states')
const form = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwRateLimitForm.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(form)
const values = { max_read_ops: 0, max_write_ops: 1, max_read_bytes: 1024, max_write_bytes: Number.MAX_SAFE_INTEGER }
assert.deepEqual(form.rgwRateLimitInput({ ...values, enabled: 'enable' }), { ...values, enabled: true })
assert.deepEqual(form.rgwRateLimitInput({ ...values, enabled: 'disable' }), { ...values, enabled: false })
assert.deepEqual(form.rgwRateLimitInitial({ ...values, enabled: true }), { ...values, enabled: 'enable' })
assert.equal(form.rgwRateLimitInitial({ enabled: false }).enabled, 'disable')
for (const enabled of [undefined, null, 0, 1, true, false, 'true', 'false', '']) assert.throws(() => form.rgwRateLimitInput({ ...values, enabled }))
for (const enabled of [undefined, null, 0, 1, 'true', 'false']) assert.equal(form.rgwRateLimitInitial({ enabled }).enabled, undefined)
for (const value of [undefined, null, '', '0', false, -1, 0.5, NaN, Infinity, Number.MAX_SAFE_INTEGER + 1]) {
  for (const key of Object.keys(values)) {
    assert.throws(() => form.rgwRateLimitInput({ ...values, enabled: 'enable', [key]: value }))
    assert.equal(form.rgwRateLimitInitial({ ...values, [key]: value })[key], undefined)
  }
}
assert.equal(form.rgwRateLimitInitial(null).max_read_ops, undefined)
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.equal(pages.match(/\.\.\.rgwRateLimitInput\(values\)/g).length, 2)
assert.equal(pages.match(/rgwRateLimitInitial\(row\?\.rate_limit\)/g).length, 2)

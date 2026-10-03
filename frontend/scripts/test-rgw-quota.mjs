import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const exports = {}
const code = ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwQuotaDetails.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText
new Function('exports', code)(exports)
const details = exports.rgwQuotaDetails
assert.deepEqual(details({ enabled: true, max_size: 0, max_objects: 0 }), { state: '已启用', size: '0', objects: '0' })
assert.deepEqual(details({ enabled: true, max_size: -1, max_objects: -2 }), { state: '已启用', size: '无限制', objects: '无限制' })
assert.equal(details({ enabled: true, max_size: 4096, max_objects: 20 }).size, '4096')
assert.deepEqual(details({ enabled: false, max_size: 4096 }), { state: '未启用' })
for (const value of [null, undefined, [], '']) assert.equal(details(value).state, '配额信息不可用')
for (const enabled of [undefined, null, 'false', 0, 1]) assert.equal(details({ enabled }).state, '配额启用状态未知')
for (const value of [undefined, null, '1', NaN, Infinity, 0.5, Number.MAX_SAFE_INTEGER + 1]) assert.equal(details({ enabled: true, max_size: value }).size, '未返回或超出精确显示范围')
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.equal(pages.match(/<RgwQuota value=\{value\} \/>/g).length, 5)
console.log('RGW quota display preserves enabled, unlimited, zero and unavailable states')

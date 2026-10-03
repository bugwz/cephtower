import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const exports = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwAccountLimit.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const limit = exports.rgwAccountLimit
for (const value of [-1, -2, -2147483648]) assert.equal(limit(value), '无限制')
assert.equal(limit(0), '0（禁止新增）')
assert.equal(limit(2), '2')
assert.equal(limit(2147483647), '2147483647')
for (const value of [undefined, null, '0', NaN, Infinity, 0.5, 2147483648, -2147483649]) assert.equal(limit(value), '资源上限未返回或无效')
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.equal(pages.match(/render: rgwAccountLimit/g).length, 4)
assert.ok(pages.includes("title: '每用户访问密钥上限'"))
console.log('RGW IAM limits preserve zero creation bans and negative unlimited values')

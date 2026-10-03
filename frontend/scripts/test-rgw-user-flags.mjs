import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const exports = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwUserFlags.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
assert.equal(exports.rgwUserSuspension(0), '未暂停')
for (const value of [1, 2, 255]) assert.equal(exports.rgwUserSuspension(value), '已暂停')
for (const value of [null, undefined, true, false, '1', -1, 256, 0.5, NaN]) assert.equal(exports.rgwUserSuspension(value), '暂停状态未知')
assert.equal(exports.rgwUserBooleanFlag(true), '是')
assert.equal(exports.rgwUserBooleanFlag(false), '否')
for (const value of [undefined, null, 0, 1, 'true']) assert.equal(exports.rgwUserBooleanFlag(value), '未知')
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.equal(pages.match(/render: rgwUserSuspension/g).length, 1)
assert.equal(pages.match(/render: rgwUserBooleanFlag/g).length, 2)
console.log('RGW user flags preserve native integer and boolean representations')

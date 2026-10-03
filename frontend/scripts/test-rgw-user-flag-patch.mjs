import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const exports = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwUserFlagPatch.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const patch = exports.rgwUserFlagPatch
assert.deepEqual(patch({}), {})
assert.deepEqual(patch({ suspended: 'keep', system: 'keep' }), {})
assert.deepEqual(patch({ suspended: 'disable', system: 'enable' }), { suspended: false, system: true })
assert.deepEqual(patch({ suspended: 'enable', system: null }), { suspended: true })
for (const value of [false, true, 0, 1, '', 'false', 'unexpected']) assert.throws(() => patch({ system: value }))
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.ok(pages.includes('...rgwUserFlagPatch(values)'))
assert.ok(!pages.includes('suspended: Boolean(values.suspended)'))
assert.ok(pages.includes("suspended: 'keep'"))
assert.ok(pages.includes("system: 'keep'"))
console.log('RGW user flag updates omit unchanged values and preserve explicit false')

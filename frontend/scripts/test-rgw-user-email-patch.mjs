import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const exports = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwUserEmailPatch.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const patch = exports.rgwUserEmailPatch
assert.deepEqual(patch({}), {})
assert.deepEqual(patch({ email_action: 'keep', email: 'stale@example.com' }), {})
assert.deepEqual(patch({ email_action: 'clear', email: 'stale@example.com' }), { email: '' })
assert.deepEqual(patch({ email_action: 'set', email: 'user@example.com' }), { email: 'user@example.com' })
for (const email of [null, undefined, 1, '', ' ', 'user\n@example.com', '\0', '\r']) assert.throws(() => patch({ email_action: 'set', email }))
assert.throws(() => patch({ email_action: 'invalid' }))
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.ok(pages.includes('...rgwUserEmailPatch(values)'))
assert.ok(pages.includes("email_action: 'keep'"))
assert.ok(pages.includes("visibleWhen: (values) => values.email_action === 'set'"))
console.log('RGW email updates require explicit set or clear actions')

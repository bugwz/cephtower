import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const exports = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/monitoring/autoRefresh.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const flush = async () => { for (let i = 0; i < 10; i++) await Promise.resolve() }
let pending = null
const clock = { setTimeout(callback, delay) { assert.equal(delay, 5000); assert.equal(pending, null); pending = callback; return 1 }, clearTimeout() { pending = null } }
const tick = () => { const callback = pending; pending = null; callback() }
let calls = 0, resolve
const stop = exports.startAutoRefresh(() => { calls++; return new Promise(done => { resolve = done }) }, 5000, clock)
tick(); await flush()
assert.equal(calls, 1)
assert.equal(pending, null)
resolve(); await flush()
assert.equal(typeof pending, 'function')
tick(); await flush()
stop(); resolve(); await flush()
assert.equal(pending, null)
assert.equal(calls, 2)
const stopFailure = exports.startAutoRefresh(async () => { throw new Error('offline') }, 5000, clock)
tick(); await flush()
assert.equal(typeof pending, 'function')
stopFailure()
const stopBeforeTick = exports.startAutoRefresh(async () => { calls++ }, 5000, clock)
stopBeforeTick(); await flush()
assert.equal(calls, 2)
const page = readFileSync(new URL('../src/pages/ExternalListPage.tsx', import.meta.url), 'utf8')
assert.ok(page.includes('return startAutoRefresh(() => refresh({ showLoading: false }), delay)'))
for (const guard of ['!selectedClusterId', 'loading', 'refreshing', 'formOpen', 'visibleDetail', 'submitting', 'mutationBlocked']) assert.ok(page.includes(guard))
console.log('Alert polling waits for completion and stops after cleanup')

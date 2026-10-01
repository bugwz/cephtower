import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/file/cephfsPerformanceSeries.ts', import.meta.url), 'utf8')
const compiled = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS } }).outputText
const exports = {}
new Function('exports', compiled)(exports)
const { performancePoints } = exports
const sample = (value, seconds = 0, gid = '1') => ({ gid, observed_at: new Date(seconds * 1000).toISOString(), counters: [{ name: 'requests', value }] })
const rate = (...values) => performancePoints(values, 'requests', true).map((point) => point?.value ?? null)

assert.deepEqual(rate(sample('9007199254740993'), sample('9007199254741013', 10)), [null, 2])
assert.deepEqual(rate(sample('100'), sample('50', 10), sample('70', 20)), [null, null, 2])
assert.deepEqual(rate(sample('100'), sample('110', 10, '2')), [null, null])
assert.deepEqual(rate(sample('100'), sample('110')), [null, null])
assert.deepEqual(rate(sample(null), sample('110', 10)), [null, null])
assert.deepEqual(rate(sample('100'), sample('110', -10)), [null, null])
assert.deepEqual(rate(sample('invalid'), sample('110', 10)), [null, null])
assert.deepEqual(rate(sample('0'), sample('18446744073709551615', 10)), [null, null])
assert.deepEqual(performancePoints([sample('0'), sample('9007199254740993')], 'requests', false).map((point) => point?.value ?? null), [0, null])
assert.deepEqual(performancePoints([{ ...sample('0'), observed_at: 'invalid' }], 'requests', false), [null])
console.log('CephFS performance series checks passed')

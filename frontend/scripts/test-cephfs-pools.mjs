import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/file/cephfsPoolCapacity.ts', import.meta.url), 'utf8')
const compiled = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS } }).outputText
const exports = {}
new Function('exports', compiled)(exports)
const { poolUsagePercent, formatPoolBytes } = exports

assert.equal(poolUsagePercent('0', '100'), 0)
assert.equal(poolUsagePercent('1', '3'), 33.33)
assert.equal(poolUsagePercent('18446744073709551615', '36893488147419103230'), 50)
assert.equal(poolUsagePercent('0', '0'), undefined)
assert.equal(poolUsagePercent(null, '100'), undefined)
assert.equal(poolUsagePercent('10', '1'), undefined)
assert.equal(poolUsagePercent('-1', '100'), undefined)
assert.equal(formatPoolBytes(null), '—')
assert.equal(formatPoolBytes('0'), '0 B')
assert.equal(formatPoolBytes('1024'), '1.00 KiB')
assert.equal(formatPoolBytes('1536'), '1.50 KiB')
assert.equal(formatPoolBytes('18446744073709551615'), '15.99 EiB')
assert.equal(formatPoolBytes('bad'), '—')
console.log('CephFS pool usage checks passed')

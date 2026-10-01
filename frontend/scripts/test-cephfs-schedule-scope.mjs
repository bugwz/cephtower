import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/file/snapshotScheduleScope.ts', import.meta.url), 'utf8')
const exports = {}
new Function('exports', ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const first = { fs: 'a', path: '/dir', subvol: 'same', group: 'team', schedule: '1h', start: '2026-10-01', retention: { n: 3 } }
assert.deepEqual(exports.scheduleScope(first), { fs: 'a', path: '/dir', subvol: 'same', group: 'team' })
// Ant Design merges setFieldsValue; the next path must explicitly clear old scope fields.
const switched = { ...exports.scheduleFormScope(first), ...exports.scheduleFormScope({ fs: 'b', path: '/' }) }
assert.deepEqual(switched, { fs: 'b', path: '/', subvol: undefined, group: undefined })
assert.deepEqual(exports.scheduleScope(switched), { fs: 'b', path: '/' })
assert.deepEqual(exports.scheduleScope({ fs: 'a', path: '/', subvol: 'same', group: '_nogroup' }), { fs: 'a', path: '/', subvol: 'same', group: '_nogroup' })
console.log('CephFS snapshot schedule scope checks passed')

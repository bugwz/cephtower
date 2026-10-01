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

const textSource = readFileSync(new URL('../src/pages/file/snapshotScheduleText.ts', import.meta.url), 'utf8')
const textExports = {}
new Function('exports', ts.transpileModule(textSource, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(textExports)
assert.equal(textExports.scheduleIntervalText('2h'), '每 2 小时（2h）')
assert.equal(textExports.scheduleIntervalText('1M'), '每 1 月（1M）')
assert.equal(textExports.scheduleIntervalText('1m'), '每 1 分钟（1m）')
assert.equal(textExports.scheduleIntervalText('1y'), '每 1 年（1y）')
assert.equal(textExports.buildScheduleInterval('2', 'y'), '2y')
assert.equal(textExports.buildRetentionRules({ h: '24', d: '7', n: '3' }), '24h7d3n')
assert.equal(textExports.buildRetentionRules({}), undefined)
assert.equal(textExports.buildRetentionRules({ h: '', M: '12', m: '5' }), '5m12M')
assert.equal(textExports.buildRetentionRules({ n: '9007199254740993' }), '9007199254740993n')
assert.deepEqual(textExports.retentionRuleCounts({ h: 24, d: '7' }), { h: '24', d: '7' })
assert.deepEqual(textExports.retentionRuleCounts('24h7d'), { h: '24', d: '7' })
assert.deepEqual(textExports.retentionRuleCounts({}), {})
for (const value of [undefined, null, '1h2h', { h: 9007199254740992 }, { bad: 1 }, { h: 0 }]) assert.equal(textExports.retentionRuleCounts(value), undefined)
for (const count of ['0', '-1', '1.5', '01', '1e3']) assert.equal(textExports.buildRetentionRules({ h: '24', d: count }), undefined)
for (const [interval, unit] of [['0', 'd'], ['-1', 'h'], ['1.5', 'd'], ['1', 'Y'], ['1', 'm'], [1, 'd']]) assert.throws(() => textExports.buildScheduleInterval(interval, unit))
assert.equal(textExports.scheduleIntervalText('future'), 'future')
assert.equal(textExports.scheduleRetentionText({ h: 24, d: 7, n: 3 }), '每小时保留 24 个快照；每日保留 7 个快照；最近保留 3 个快照')
assert.equal(textExports.scheduleRetentionText('24h7d3n'), textExports.scheduleRetentionText({ h: 24, d: 7, n: 3 }))
assert.equal(textExports.scheduleRetentionText({}), '未设置')
assert.equal(textExports.scheduleRetentionText(undefined), '未知')
assert.equal(textExports.scheduleRetentionText({ h: 9007199254740992 }), 'h：未知规则')
assert.equal(textExports.scheduleRetentionText({ M: '9007199254740993' }), '每月保留 9007199254740993 个快照')
assert.equal(textExports.scheduleRetentionText({ toString: 3 }), 'toString：未知规则')
console.log('CephFS schedule text checks passed')
assert.equal(textExports.scheduleActiveText(true), '启用')
assert.equal(textExports.scheduleActiveText(false), '停用')
assert.equal(textExports.scheduleToggleAction(true), 'deactivate')
assert.equal(textExports.scheduleToggleAction(false), 'activate')
for (const value of [undefined, null, 'false', 'true', 0, 1, {}, []]) {
  assert.equal(textExports.scheduleActiveText(value), '未知')
  assert.equal(textExports.scheduleToggleAction(value), undefined)
}

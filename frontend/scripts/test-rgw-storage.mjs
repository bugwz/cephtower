import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const exports = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwStorageDetails.ts', import.meta.url), 'utf8'), {
  compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 }
}).outputText)(exports)
const rows = exports.rgwStorageRows
const stats = { size: 0, size_actual: 4096, size_utilized: 7, num_objects: 2 }
assert.deepEqual(rows({ stats }), [{ category: '汇总', size: '0', actual: '4096', utilized: '7', objects: '2' }])
assert.deepEqual(rows({ 'rgw.main': stats, future: stats }, true).map(row => row.category), ['rgw.main', 'future'])
assert.deepEqual(rows({}, true), [])
for (const value of [undefined, null, [], 'bad']) assert.equal(rows(value), undefined)
for (const value of [{}, { stats: [] }, { stats: null }]) assert.equal(rows(value), undefined)
assert.equal(rows({ bad: null }, true), undefined)
for (const size of [undefined, null, -1, 0.5, NaN, Infinity, '1', Number.MAX_SAFE_INTEGER + 1]) {
  assert.equal(rows({ stats: { ...stats, size } })[0].size, '未返回或超出精确显示范围')
}
assert.equal(rows({ stats: { num_objects: 0 } })[0].objects, '0')
assert.equal(rows({ stats: { size_kb: 1 } })[0].size, '未返回或超出精确显示范围')
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.equal(pages.match(/<RgwStorage value=\{value\} \/>/g).length, 1)
assert.equal(pages.match(/<RgwStorage value=\{value\} account \/>/g).length, 1)
assert.equal(pages.match(/<RgwStorage value=\{value\} categorized \/>/g).length, 1)
console.log('RGW storage statistics preserve category, units, zero and unavailable values')
const times = exports.rgwStorageTimes
const state = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketState.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(state)
assert.equal(state.rgwBucketVersioning('enabled'), '已启用（enabled）')
assert.equal(state.rgwBucketVersioning('off'), '未启用（off）')
assert.ok(state.rgwBucketVersioning('suspended').includes('已暂停'))
for (const value of ['future', 'toString', '__proto__']) assert.equal(state.rgwBucketVersioning(value), `未知状态：${value}`)
for (const value of [undefined, null, '', 0, false, {}]) assert.equal(state.rgwBucketVersioning(value), '版本控制状态未返回或无效')
assert.equal(state.rgwBucketBooleanState(true), '已启用')
assert.equal(state.rgwBucketBooleanState(false), '未启用')
for (const value of [undefined, null, '', 0, 1, 'false', {}]) assert.equal(state.rgwBucketBooleanState(value), '状态未返回或无效')
assert.equal(pages.match(/render: rgwBucketBooleanState/g).length, 2)
assert.ok(pages.includes('render: rgwBucketVersioning'))
assert.ok(pages.includes("title: 'MFA Delete'"))
const index = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketIndex.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(index)
for (const value of [0, 1, Number.MAX_SAFE_INTEGER]) assert.equal(index.rgwBucketIndexCount(value), String(value))
for (const value of [undefined, null, '', '0', -1, 0.5, NaN, Infinity, Number.MAX_SAFE_INTEGER + 1]) assert.equal(index.rgwBucketIndexCount(value), '未返回或超出安全整数范围')
for (const value of ['Normal', 'Indexless', 'future', '0#opaque,version']) assert.equal(index.rgwBucketIndexText(value), value)
assert.equal(index.rgwBucketIndexText(''), '（空字符串）')
for (const value of [null, undefined, 0, {}]) assert.equal(index.rgwBucketIndexText(value), '未返回或格式无效')
assert.ok(pages.includes('<RgwBucketIndexDetails row={row} />'))
const indexView = readFileSync(new URL('../src/pages/object/RgwBucketIndexDetails.tsx', import.meta.url), 'utf8')
for (const field of ['index_type', 'index_generation', 'num_shards', 'ver', 'master_ver', 'marker', 'max_marker']) assert.ok(indexView.includes(`row.${field}`))
assert.equal(exports.rgwStorageScope('user', ''), '用户汇总')
assert.ok(exports.rgwStorageScope('account', 'RGW123').includes('RGW123'))
assert.ok(exports.rgwStorageScope('account', 'RGW123').includes('不可相加'))
assert.ok(exports.rgwStorageScope('account', null).includes('账户 ID 未返回'))
for (const scope of [undefined, null, '', 'future', 0]) assert.equal(exports.rgwStorageScope(scope, 'RGW123'), '统计范围未知，请重新采集')
assert.ok(pages.includes('rgwStorageScope(value, row.account_id)'))
assert.deepEqual(times({ last_stats_sync: '2026-10-03 12:00:00.123456Z', last_stats_update: '2026-10-03 12:01:00Z' }), { synced: '2026-10-03 12:00:00.123456Z', updated: '2026-10-03 12:01:00Z' })
assert.deepEqual(times({ last_synced: 'sync', last_updated: 'update' }, true), { synced: 'sync', updated: 'update' })
assert.equal(times({ last_synced: 'wrong scope' }).synced, '未返回或格式无效')
for (const value of [null, undefined, '', 0, false, []]) assert.equal(times({ last_stats_sync: value }).synced, '未返回或格式无效')

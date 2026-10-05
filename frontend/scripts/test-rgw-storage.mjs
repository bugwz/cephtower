import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const exports = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwStorageDetails.ts', import.meta.url), 'utf8'), {
  compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 }
}).outputText)(exports)
const rows = exports.rgwStorageRows
const stats = { size: '0', size_actual: '4096', size_utilized: '7', num_objects: '2' }
assert.deepEqual(rows({ stats }), [{ category: '汇总', size: '0', actual: '4096', utilized: '7', objects: '2' }])
assert.deepEqual(rows({ 'rgw.main': stats, future: stats }, true).map(row => row.category), ['rgw.main', 'future'])
assert.deepEqual(rows({}, true), [])
for (const value of [undefined, null, [], 'bad']) assert.equal(rows(value), undefined)
for (const value of [{}, { stats: [] }, { stats: null }]) assert.equal(rows(value), undefined)
assert.equal(rows({ bad: null }, true), undefined)
for (const size of [undefined, null, -1, 0, 1, 0.5, NaN, Infinity, Number.MAX_SAFE_INTEGER + 1]) {
  assert.equal(rows({ stats: { ...stats, size } })[0].size, '未返回或超出精确显示范围')
}
assert.equal(rows({ stats: { num_objects: '0' } })[0].objects, '0')
assert.equal(rows({ stats: { size_kb: 1 } })[0].size, '未返回或超出精确显示范围')
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const userDetails = readFileSync(new URL('../src/pages/object/RgwUserDetails.tsx', import.meta.url), 'utf8')
assert.ok(userDetails.includes('<RgwStorage value={row.storage_stats} />'))
const accountDetails = readFileSync(new URL('../src/pages/object/RgwAccountDetails.tsx', import.meta.url), 'utf8')
assert.ok(accountDetails.includes('<RgwStorage value={row.storage_stats} account />'))
assert.equal(pages.match(/<RgwStorage value=\{value\} categorized \/>/g).length, 1)
console.log('RGW storage statistics preserve category, units, zero and unavailable values')
const times = exports.rgwStorageTimes
const storageView = {}
const jsx = (type, props) => ({ type, props })
new Function('exports', 'require', ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwStorage.tsx', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX } }).outputText)(storageView, (name) => name === 'react/jsx-runtime' ? { jsx, jsxs: jsx } : name === './rgwStorageDetails' ? exports : { Table: 'Table', Descriptions: 'Descriptions' })
for (const value of [undefined, null, [], { main: null }]) {
  const result = storageView.RgwStorage({ value, categorized: true })
  assert.equal(result.type, 'span')
  assert.ok(result.props.children.includes('缺失不代表零用量'))
}
assert.equal(storageView.RgwStorage({ value: undefined }).props.children, '容量与对象统计不可用')
const emptyUsage = storageView.RgwStorage({ value: {}, categorized: true })
assert.deepEqual(emptyUsage.props.children[0].props.dataSource, [])
assert.equal(emptyUsage.props.children[0].props.locale.emptyText, '本次采集未返回用量分类')
const zeroUsage = storageView.RgwStorage({ value: { main: { size: '0', size_actual: '0', size_utilized: '0', num_objects: '0' } }, categorized: true })
assert.equal(zeroUsage.props.children[0].props.dataSource[0].objects, '0')
const preciseBucket = { 'rgw.main': { size: '18446744073709551615', size_actual: '9007199254740993', size_utilized: '0', num_objects: '9007199254740995' } }
for (const account of [false, true]) {
  const table = storageView.RgwStorage({ value: { stats: preciseBucket['rgw.main'] }, account }).props.children[0]
  assert.deepEqual(table.props.dataSource, [{ category: '汇总', size: '18446744073709551615', actual: '9007199254740993', utilized: '0', objects: '9007199254740995' }])
}
assert.deepEqual(storageView.RgwStorage({ value: preciseBucket, categorized: true }).props.children[0].props.dataSource, [
  { category: 'rgw.main', size: '18446744073709551615', actual: '9007199254740993', utilized: '0', objects: '9007199254740995' }
])
for (const value of [undefined, null, 0, 1, false, {}, '', '-1', '01', '1.0', '1e3', '+1', ' 1', '18446744073709551616', '9'.repeat(100)]) {
  const bad = rows({ main: { size: value, size_actual: value, size_utilized: value, num_objects: value } }, true)[0]
  for (const key of ['size', 'actual', 'utilized', 'objects']) assert.equal(bad[key], '未返回或超出精确显示范围')
}
const tags = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketTags.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(tags)
assert.deepEqual(tags.rgwBucketTags({}), [])
assert.deepEqual(tags.rgwBucketTags({ team: 'storage', empty: '' }), [{ key: 'team', value: 'storage' }, { key: 'empty', value: '' }])
assert.deepEqual(tags.rgwBucketTags(JSON.parse('{"__proto__":"literal","constructor":"tag"}')), [{ key: '__proto__', value: 'literal' }, { key: 'constructor', value: 'tag' }])
for (const value of [undefined, null, [], [{ Key: 'wrong', Value: 'shape' }], { key: null }, { key: 0 }]) assert.equal(tags.rgwBucketTags(value), undefined)
const bucketDetails = readFileSync(new URL('../src/pages/object/RgwBucketDetails.tsx', import.meta.url), 'utf8')
assert.ok(bucketDetails.includes('<RgwBucketTagsTable value={row.tagset} />'))
const placement = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketPlacement.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(placement)
assert.deepEqual(placement.rgwBucketPlacement({ data_pool: 'zone.data:ns', data_extra_pool: '', index_pool: 'zone.index' }), { data: 'zone.data:ns', extra: '未显式指定', index: 'zone.index' })
assert.deepEqual(placement.rgwBucketPlacement({ data_pool: '', data_extra_pool: '', index_pool: '' }), { data: '未显式指定', extra: '未显式指定', index: '未显式指定' })
for (const value of [undefined, null, [], '', {}, { data_pool: 0, data_extra_pool: false, index_pool: {} }]) assert.deepEqual(placement.rgwBucketPlacement(value), { data: '未返回或格式无效', extra: '未返回或格式无效', index: '未返回或格式无效' })
assert.equal(placement.rgwBucketPlacement({ data_pool: 'one' }).data, 'one')
assert.equal(placement.rgwBucketPlacement({ data_pool: 'one' }).index, '未返回或格式无效')
assert.ok(bucketDetails.includes('<RgwBucketPlacementDetails value={row.explicit_placement} />'))
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
assert.equal(state.rgwBucketReshardState('None'), '当前未处于重新分片阶段（None）')
assert.equal(state.rgwBucketReshardState('InLogrecord'), '日志记录阶段（InLogrecord）')
assert.equal(state.rgwBucketReshardState('InProgress'), '重新分片进行中（InProgress）')
for (const value of ['future', 'none', 'toString', '__proto__']) assert.equal(state.rgwBucketReshardState(value), `未知重新分片状态：${value}`)
for (const value of [undefined, null, '', 0, false, {}]) assert.equal(state.rgwBucketReshardState(value), '重新分片状态未返回或无效')
assert.ok(pages.includes("title: '重新分片状态（采集时）', ellipsis: false, render: rgwBucketReshardState"))
const index = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketIndex.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(index)
for (const value of [0, 1, Number.MAX_SAFE_INTEGER]) assert.equal(index.rgwBucketIndexCount(value), String(value))
for (const value of [undefined, null, '', '0', -1, 0.5, NaN, Infinity, Number.MAX_SAFE_INTEGER + 1]) assert.equal(index.rgwBucketIndexCount(value), '未返回或超出安全整数范围')
for (const value of ['Normal', 'Indexless', 'future', '0#opaque,version']) assert.equal(index.rgwBucketIndexText(value), value)
assert.equal(index.rgwBucketIndexText(''), '（空字符串）')
for (const value of [null, undefined, 0, {}]) assert.equal(index.rgwBucketIndexText(value), '未返回或格式无效')
assert.ok(bucketDetails.includes('<RgwBucketIndexDetails row={row} />'))
assert.ok(pages.includes('detailContent: (row, clusterId) => <RgwBucketDetails row={row} configuration={<RgwBucketConfigurationPanel row={row} clusterId={clusterId} definition={externalDefinitions.bucketPolicy} />} />'))
assert.ok(pages.includes("key: 'index_type', title: '索引类型', render: rgwBucketIndexText"))
assert.ok(pages.includes("key: 'num_shards', title: '索引分片数', render: rgwBucketIndexCount"))
const detailsView = {}
new Function('exports', 'require', ts.transpileModule(bucketDetails, { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX } }).outputText)(detailsView, (name) => name === 'react/jsx-runtime' ? { jsx, jsxs: jsx } : name === 'antd' ? { Tabs: 'Tabs' } : { RgwBucketIndexDetails: 'Index', RgwBucketPlacementDetails: 'Placement', RgwBucketTagsTable: 'Tags', RgwQuota: 'Quota', RgwRateLimit: 'RateLimit', RgwStorage: 'Storage' })
const bucketRow = { index_type: 'Normal', explicit_placement: { data_pool: 'data' }, tagset: { team: 'storage' } }
const tabs = detailsView.RgwBucketDetails({ row: bucketRow })
assert.equal(tabs.type, 'Tabs')
assert.deepEqual(tabs.props.items.map(item => item.key), ['index', 'placement', 'tags', 'quota', 'rate-limit', 'usage', 'summary'])
assert.equal(tabs.props.items.find(item => item.key === 'summary').children.props.row, bucketRow)
assert.equal(tabs.props.items[0].children.props.row, bucketRow)
assert.equal(tabs.props.items[1].children.props.value, bucketRow.explicit_placement)
assert.equal(tabs.props.items[2].children.props.value, bucketRow.tagset)
for (const row of [{}, { bucket_quota: null, rate_limit: null, usage: null }, {
  bucket_quota: { enabled: false, max_size: 0, max_objects: 0 },
  rate_limit: { enabled: true, max_read_ops: 0 },
  usage: { 'rgw.main': { size: 0, num_objects: 0 } }
}]) {
  const items = detailsView.RgwBucketDetails({ row }).props.items
  assert.equal(items.find(item => item.key === 'quota').children.type, 'Quota')
  assert.equal(items.find(item => item.key === 'quota').children.props.value, row.bucket_quota)
  const rate = items.find(item => item.key === 'rate-limit')
  assert.equal(rate.children.type, 'RateLimit')
  assert.ok(rate.label.includes('每 RGW'))
  assert.equal(rate.children.props.value, row.rate_limit)
  const usage = items.find(item => item.key === 'usage').children
  assert.equal(usage.type, 'Storage')
  assert.equal(usage.props.value, row.usage)
  assert.equal(usage.props.categorized, true)
}
const configuration={type:'BucketConfiguration'}
assert.equal(detailsView.RgwBucketDetails({row:bucketRow,configuration}).props.items.find(item=>item.key==='configuration').children,configuration)
const indexView = readFileSync(new URL('../src/pages/object/RgwBucketIndexDetails.tsx', import.meta.url), 'utf8')
for (const field of ['index_type', 'index_generation', 'num_shards', 'ver', 'master_ver', 'marker', 'max_marker']) assert.ok(indexView.includes(`row.${field}`))
const indexComponent = {}
new Function('exports', 'require', ts.transpileModule(indexView, { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX } }).outputText)(indexComponent, name => {
  if (name === 'react/jsx-runtime') return { jsx, jsxs: jsx }
  if (name === 'antd') return { Descriptions: 'Descriptions' }
  if (name === './rgwBucketIndex') return index
  if (name === './rgwBucketState') return state
  throw new Error(name)
})
for (const status of ['None', 'InLogrecord', 'InProgress', 'future', undefined, null]) {
  const result = indexComponent.RgwBucketIndexDetails({ row: { reshard_status: status, judge_reshard_lock_time: '2026-10-06 12:00:00.123456Z' } })
  const items = result.props.children[0].props.items
  assert.equal(items.find(item => item.key === 'reshard').children, state.rgwBucketReshardState(status))
  assert.equal(items.find(item => item.key === 'judge-time').children, '2026-10-06 12:00:00.123456Z')
  assert.ok(result.props.children[2].props.children.props.children.includes('不代表任务开始或完成时间'))
}
for (const value of [undefined, null, 0, false, {}]) {
  const items = indexComponent.RgwBucketIndexDetails({ row: { judge_reshard_lock_time: value } }).props.children[0].props.items
  assert.equal(items.find(item => item.key === 'judge-time').children, '未返回或格式无效')
}
assert.equal(exports.rgwStorageScope('user', ''), '用户汇总')
assert.ok(exports.rgwStorageScope('account', 'RGW123').includes('RGW123'))
assert.ok(exports.rgwStorageScope('account', 'RGW123').includes('不可相加'))
assert.ok(exports.rgwStorageScope('account', null).includes('账户 ID 未返回'))
for (const scope of [undefined, null, '', 'future', 0]) assert.equal(exports.rgwStorageScope(scope, 'RGW123'), '统计范围未知，请重新采集')
assert.ok(userDetails.includes('rgwStorageScope(row.stats_scope, row.account_id)'))
assert.deepEqual(times({ last_stats_sync: '2026-10-03 12:00:00.123456Z', last_stats_update: '2026-10-03 12:01:00Z' }), { synced: '2026-10-03 12:00:00.123456Z', updated: '2026-10-03 12:01:00Z' })
assert.deepEqual(times({ last_synced: 'sync', last_updated: 'update' }, true), { synced: 'sync', updated: 'update' })
assert.equal(times({ last_synced: 'wrong scope' }).synced, '未返回或格式无效')
for (const value of [null, undefined, '', 0, false, []]) assert.equal(times({ last_stats_sync: value }).synced, '未返回或格式无效')
import './test-rgw-lifecycle-progress.mjs'
import './test-rgw-bucket-summary.mjs'

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
import './test-mon-counter-pagination.mjs'
import './test-mon-detail-scope.mjs'
import './test-mon-quorum-groups.mjs'

const exports = {}
const jsx = (type, props) => ({ type, props })
const snapshot = {}
new Function('exports', 'require', ts.transpileModule(readFileSync(new URL('../src/pages/cluster/MonSnapshotState.tsx', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX } }).outputText)(snapshot, name => name === 'antd' ? { Space: 'Space', Tag: 'Tag', Typography: { Text: 'Text' } } : name.includes('utils/time') ? { formatDateTime: value => value } : { jsx, jsxs: jsx })
for (const [stale, observedAt, expected] of [
  [true, '2026-10-06T10:00:00Z', '历史快照，请重新采集'],
  [false, '2026-10-06T10:00:00Z', '采集快照（非实时）'],
  [undefined, '2026-10-06T10:00:00Z', '采集状态未知'],
  ['false', '2026-10-06T10:00:00Z', '采集状态未知'],
  [false, null, '采集状态未知'], [false, 'invalid', '采集状态未知']
]) assert.equal(snapshot.MonSnapshotState({ stale, observedAt }).props.children[0].props.children, expected)
for (const file of ['pages.tsx', 'MonDetailPage.tsx']) assert.ok(readFileSync(new URL(`../src/pages/cluster/${file}`, import.meta.url), 'utf8').includes('<MonSnapshotState observedAt='))
new Function('exports', 'require', ts.transpileModule(readFileSync(new URL('../src/pages/cluster/MonPublicAddresses.tsx', import.meta.url), 'utf8'), {
  compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX }
}).outputText)(exports, () => ({ jsx, jsxs: jsx }))
const render = value => exports.MonPublicAddresses({ value })
for (const value of [undefined, null, {}, 0, 'v2:address']) assert.ok(render(value).props.children.includes('未返回或格式无效'))
assert.equal(render([]).props.children, '原生地址列表为空')
for (const value of [[null], [{}], [{ type: 'v2', addr: 123 }], [{ type: '', addr: 'x' }]]) assert.equal(render(value).props.children, '地址列表格式无效')
const result = render([{ type: 'v2', addr: '[2001:db8::1]:3300' }, { type: 'v1', addr: '192.0.2.1:6789' }])
assert.deepEqual(result.props.children.map(item => item.props.children.props.children), [['v2', ': ', '[2001:db8::1]:3300'], ['v1', ': ', '192.0.2.1:6789']])
for (const file of ['pages.tsx', 'MonDetailPage.tsx']) {
  assert.ok(readFileSync(new URL(`../src/pages/cluster/${file}`, import.meta.url), 'utf8').includes('MonPublicAddresses'))
}
console.log('MON public address list preserves native protocols and IPv6 text')
const summary = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/cluster/monQuorumSummary.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(summary)
assert.equal(summary.monQuorumMembers([]), '无仲裁成员')
assert.equal(summary.monQuorumMembers(['a', 'b']), 'a、b')
assert.equal(summary.monQuorumLeader(''), '无 Leader')
assert.equal(summary.monQuorumLeader('a'), 'a')
assert.equal(summary.monQuorumInteger('0', true), '0 秒')
assert.equal(summary.monQuorumInteger('18446744073709551615'), '18446744073709551615')
for (const value of [undefined, null, {}, false, 0]) {
  assert.equal(summary.monQuorumMembers(value), '未返回或格式无效')
  assert.equal(summary.monQuorumLeader(value), '未返回或格式无效')
  assert.equal(summary.monQuorumInteger(value), '未返回或格式无效')
}
for (const value of ['-1', '1.5', '01', '18446744073709551616']) assert.equal(summary.monQuorumInteger(value), '未返回或格式无效')
for (const field of ['quorum_names', 'quorum_leader_name', 'election_epoch', 'quorum_age']) assert.ok(readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8').includes(`data?.status?.${field}`))
const counter = {}
new Function('exports', 'require', ts.transpileModule(readFileSync(new URL('../src/pages/cluster/MonCounterValue.tsx', import.meta.url), 'utf8'), {
  compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX }
}).outputText)(counter, () => ({ jsx, jsxs: jsx }))
assert.deepEqual(counter.MonCounterValue({ value: 12.5, unit: 'B/s' }).props.children, ['12.5', ' B/s'])
assert.deepEqual(counter.MonCounterValue({ value: 0, unit: '/s' }).props.children, ['0', ' /s'])
assert.deepEqual(counter.MonCounterValue({ value: '18446744073709551615', unit: undefined }).props.children, ['18446744073709551615', ''])
for (const value of [null, undefined, false, {}, [], NaN, Infinity, '']) assert.equal(counter.MonCounterValue({ value, unit: 'B/s' }).props.children, '未返回或格式无效')
assert.equal(counter.monCounterType('counter'), '采样间隔速率')
assert.equal(counter.monCounterType('gauge'), '采集值')
assert.equal(counter.monCounterType('histogram'), '直方图')
assert.equal(counter.monCounterType('future'), '类型未知')
assert.equal(counter.monCounterType(undefined), '类型未知')
assert.ok(readFileSync(new URL('../src/pages/cluster/MonDetailPage.tsx', import.meta.url), 'utf8').includes('<MonCounterValue value={value} unit={row.unit} />'))
const quorum = {}
new Function('exports', 'require', ts.transpileModule(readFileSync(new URL('../src/pages/cluster/MonQuorumState.tsx', import.meta.url), 'utf8'), {
  compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX }
}).outputText)(quorum, name => name === 'antd' ? { Tag: 'Tag' } : { jsx, jsxs: jsx })
assert.equal(quorum.MonQuorumState({ value: true }).props.children, '仲裁中（采集时）')
assert.equal(quorum.MonQuorumState({ value: false }).props.children, '未加入仲裁（采集时）')
for (const value of [undefined, null, 0, 1, '', 'false', 'true', {}, []]) {
  const state = quorum.MonQuorumState({ value })
  assert.equal(state.props.color, 'warning')
  assert.equal(state.props.children, '仲裁状态未知，请重新采集')
}
for (const file of ['pages.tsx', 'MonDetailPage.tsx']) {
  const source = readFileSync(new URL(`../src/pages/cluster/${file}`, import.meta.url), 'utf8')
  assert.ok(source.includes('<MonQuorumState value={'))
  assert.equal(source.includes('in_quorum === true'), false)
}
const counts = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/cluster/monSessionCount.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(counts)
for (const value of ['0', '15', '9007199254740993', '18446744073709551615']) assert.equal(counts.monSessionCount(value), value)
for (const value of [undefined, null, 0, false, {}, [], '01', '-1', '1.5', '1e3', '18446744073709551616']) assert.equal(counts.monSessionCount(value), '未返回或格式无效')

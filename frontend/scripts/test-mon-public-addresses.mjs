import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const exports = {}
const jsx = (type, props) => ({ type, props })
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

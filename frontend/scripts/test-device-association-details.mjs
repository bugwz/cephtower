import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/cluster/DeviceAssociationDetails.tsx', import.meta.url), 'utf8')
const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.React } }).outputText
const exports = {}
new Function('exports', 'require', 'React', code)(exports, name => name === 'antd' ? { Alert: 'Alert', Descriptions: { Item: 'Item' }, Space: 'Space', Typography: { Text: 'Text' } } : { isRecord: value => value !== null && typeof value === 'object' && !Array.isArray(value) }, { createElement: (type, props, ...children) => ({ type, props, children }) })
const render = device => JSON.stringify(exports.DeviceAssociationDetails({ device }))
const full = render({ devid: 'serial', daemons: ['osd.1', 'osd.10'], location: [{ host: 'node1', dev: 'sda', path: '/dev/disk/by-id/serial' }, { host: 'node2', dev: 'sdb', path: '/dev/sdb' }], life_expectancy_min: '2026-10-07 00:00:00', life_expectancy_max: '2026-12-07 00:00:00', life_expectancy_stamp: '2026-10-06 00:00:00' })
for (const text of ['serial', 'osd.1、osd.10', 'node1', 'node2', '/dev/disk/by-id/serial', '/dev/sdb', '2026-10-07 00:00:00', '2026-12-07 00:00:00', '2026-10-06 00:00:00', '不是保证寿命']) assert.ok(full.includes(text), text)
for (const value of [undefined, null, {}, 'invalid', [null], [1]]) {
  const result = render({ location: value, daemons: value, life_expectancy_min: false })
  assert.ok(result.includes('设备位置未返回或格式无效'))
  assert.ok(result.includes('未返回或格式无效'))
}
const empty = render({ location: [], daemons: [] })
assert.ok(empty.includes('本次未报告设备位置'))
assert.ok(empty.includes('本次未报告关联进程'))
assert.ok(!empty.includes('设备位置未返回或格式无效'))
const page = readFileSync(new URL('../src/pages/cluster/OSDInspection.tsx', import.meta.url), 'utf8')
assert.ok(page.includes('<DeviceAssociationDetails device={device} />'))
assert.ok(page.includes('<summary>原始关联记录</summary><RecordDetail record={device} />'))
console.log('OSD device associations preserve native locations, daemons and prediction bounds')

const hostSource = readFileSync(new URL('../src/pages/cluster/HostDetailPage.tsx', import.meta.url), 'utf8')
const hostTree = ts.createSourceFile('HostDetailPage.tsx', hostSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const tableNode = hostTree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'HostDeviceInfoTable')
const tableCode = ts.transpileModule(tableNode.getText(hostTree), { compilerOptions: { target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.React } }).outputText
const react = { createElement: (type, props, ...children) => typeof type === 'function' ? type(props) : ({ type, props, children }) }
const table = new Function('React', 'DataTable', 'DeviceAssociationDetails', 'renderDeviceHealth', 'formatDateTime', `${tableCode}; return HostDeviceInfoTable`)(react, 'DataTable', exports.DeviceAssociationDetails, value => value, value => value)
const target = { devid: 'host-device', location: [{ host: 'node1', dev: 'sda', path: '/dev/disk/by-id/host-device' }, { host: 'node2', dev: 'sdb', path: '/dev/sdb' }], daemons: ['osd.0'], life_expectancy_min: '2026-10-07', life_expectancy_max: '2027-01-01' }
const column = table({ devices: [target] }).props.columns.find(column => column.key === 'association_details')
assert.equal(column.filterKey, false)
assert.equal(column.ellipsis, false)
const detail = JSON.stringify(column.render(undefined, target))
for (const value of ['host-device', '/dev/disk/by-id/host-device', 'node2', '/dev/sdb', '2026-10-07', '2027-01-01']) assert.ok(detail.includes(value), value)
assert.ok(JSON.stringify(column.render(undefined, {})).includes('未返回或格式无效'))
console.log('Host device table exposes full association details without truncating locations')

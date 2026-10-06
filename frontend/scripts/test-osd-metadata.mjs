import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/OSDMetadata.tsx', import.meta.url), 'utf8')
const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.React } }).outputText
const exports = {}
new Function('exports', 'require', 'React', code)(exports, name => name === 'antd' ? { Alert: 'Alert', Descriptions: { Item: 'Item' }, Space: 'Space' } : { RecordDetail: 'RecordDetail' }, { createElement: (type, props, ...children) => ({ type, props, children }) })
const render = data => exports.OSDMetadata({ data })
const data = { hostname: 'node1', ceph_version: 'ceph version test', osd_objectstore: 'bluestore', os: 'Linux', kernel_version: 'test-kernel', arch: 'x86_64', cpu: 'test-cpu', mem_total_kb: '18446744073709551615', bluefs: '1', bluefs_dedicated_db: '0', bluefs_dedicated_wal: '1', custom_field: 'preserved' }
const tree = render(data)
const summary = tree.children.find(child => child?.type && typeof child.type === 'object')
const rows = summary.children.flat()
for (const key of ['hostname', 'ceph_version', 'osd_objectstore', 'os', 'kernel_version', 'arch', 'cpu', 'mem_total_kb']) assert.equal(rows.find(row => row.props.key === key).children[0], data[key])
for (const [key, value] of [['bluefs', '是'], ['bluefs_dedicated_db', '否'], ['bluefs_dedicated_wal', '是']]) assert.equal(rows.find(row => row.props.key === key).children[0], value)
assert.equal(tree.children.find(child => child?.type === 'details').children[1].props.record, data)
for (const value of [undefined, null, false, 0, 1, 'true', '']) {
  const unknown = render({ bluefs: value }).children.find(child => child?.type && typeof child.type === 'object').children.flat().find(row => row.props.key === 'bluefs')
  assert.equal(unknown.children[0], '未返回或格式无效')
}
assert.ok(JSON.stringify(render({})).includes('本次未返回 OSD 元数据'))
const page = readFileSync(new URL('../src/pages/cluster/OSDInspection.tsx', import.meta.url), 'utf8')
assert.ok(page.includes("section === 'metadata' ? <OSDMetadata data={data} />"))
console.log('OSD metadata summary preserves native strings, layout flags and raw fields')

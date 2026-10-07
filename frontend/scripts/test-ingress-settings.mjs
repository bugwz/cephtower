import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/IngressSettings.tsx', import.meta.url), 'utf8')
const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.React } }).outputText
const exports = {}
new Function('exports', 'require', 'React', code)(exports, name => name === 'antd' ? { Alert: 'Alert', Descriptions: 'Descriptions', Space: 'Space' } : { isRecord: value => value !== null && typeof value === 'object' && !Array.isArray(value) }, { createElement: (type, props, ...children) => ({ type, props, children }) })
const value = { backend_service: 'rgw.a', virtual_ip: '192.0.2.10/24', frontend_port: 443, monitor_port: 9000, ssl: false, keepalive_only: true, virtual_interface_networks: ['192.0.2.0/24', '2001:db8::/64'], ssl_key: 'never-render-this' }
const tree = exports.IngressSettings({ value })
const items = tree.children.find(child => child.type === 'Descriptions').props.items
assert.equal(items.find(item => item.key === 'ssl').children, '否')
assert.equal(items.find(item => item.key === 'keepalive').children, '是')
assert.equal(items.find(item => item.key === 'networks').children, '192.0.2.0/24、2001:db8::/64')
assert.ok(!JSON.stringify(tree).includes('never-render-this'))
for (const value of [null, undefined, []]) assert.ok(JSON.stringify(exports.IngressSettings({ value })).includes('本次未采集'))
assert.ok(JSON.stringify(exports.IngressSettings({ value: {} })).includes('未报告'))
assert.ok(JSON.stringify(exports.IngressSettings({ value: { virtual_interface_networks: [] } })).includes('未指定候选网段'))
console.log('Ingress configuration view preserves explicit flags and excludes credential fields')

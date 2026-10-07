import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/block/RbdMirrorDaemons.tsx', import.meta.url), 'utf8')
const ui = {}, jsx = (type, props) => ({ type, props })
new Function('exports', 'require', ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText)(ui, name => name === 'antd' ? { Alert: 'Alert', Space: 'Space', Table: 'Table' } : { jsx, jsxs: jsx })
const sample = { service_id: 'svc', instance_id: '9007199254740993', client_id: 'client.mirror', hostname: 'host', ceph_version: 'ceph version 20', leader: false, health: 'WARNING', callouts: ['<script>bad</script>', 'lagging'] }
assert.deepEqual(ui.rbdMirrorDaemonRows([sample])[0], { index: 0, service: 'svc', instance: '9007199254740993', client: 'client.mirror', hostname: 'host', version: 'ceph version 20', leader: '否', health: 'WARNING', callouts: '<script>bad</script>\nlagging' })
for (const bad of [undefined, null, {}, [null], [[]], ['bad']]) {
  assert.equal(ui.rbdMirrorDaemonRows(bad), undefined)
  assert.equal(ui.RbdMirrorDaemons({ value: bad }).props.type, 'warning')
}
assert.equal(ui.RbdMirrorDaemons({ value: [] }).props.type, 'info')
const rows = ui.rbdMirrorDaemonRows([{}, { leader: true, health: 'FUTURE', callouts: [] }, { leader: 'false', callouts: [1] }])
assert.equal(rows[0].health, '未返回或无效')
assert.equal(rows[0].callouts, '未返回提示')
assert.equal(rows[1].leader, '是')
assert.equal(rows[1].health, 'FUTURE')
assert.equal(rows[1].callouts, '空提示列表')
assert.equal(rows[2].leader, '未返回或无效')
assert.equal(rows[2].callouts, '提示格式无效')
const rendered = ui.RbdMirrorDaemons({ value: [sample, sample] })
const table = rendered.props.children.find(node => node.type === 'Table')
assert.deepEqual(table.props.dataSource.map(row => row.index), [0, 1])
assert.equal(table.props.columns.length, 8)
assert.equal(table.props.columns.at(-1).render('<img>').props.children, '<img>')
assert.ok(rendered.props.children[0].props.children.includes('仅针对本池'))
assert.ok(readFileSync(new URL('../src/pages/block/pages.tsx', import.meta.url), 'utf8').includes('<RbdMirrorDaemons value={value} />'))
console.log('RBD mirror daemon table preserves native identity, scope, and unknown states')

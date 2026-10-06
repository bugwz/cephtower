import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/ServiceDaemons.tsx', import.meta.url), 'utf8')
const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.React } }).outputText
const exports = {}
new Function('exports', 'require', 'React', code)(exports, name => name === 'antd' ? { Descriptions: 'Descriptions' } : {}, { createElement: (type, props, ...children) => ({ type, props, children }) })
for (const type of ['mgr', 'mds']) {
  const result = exports.DaemonRuntimeDetails({ row: { type, memory_usage: '18446744073709551615', rank: '0', rank_generation: '9007199254740993', is_active: false, pending_daemon_config: true, events: ['daemon failed', 'deployment scheduled'] }, daemonType: type })
  const items = result.children[1].props.items
  const value = key => items.find(item => item.key === key).children.children[0]
  assert.equal(value('daemon_type'), type)
  assert.equal(value('memory_usage'), '18446744073709551615')
  assert.equal(value('rank'), '0')
  assert.equal(value('rank_generation'), '9007199254740993')
  assert.equal(value('is_active'), '否')
  assert.equal(value('pending_daemon_config'), '是')
  assert.equal(value('container_id'), '未返回')
  assert.match(value('events'), /daemon failed/)
  assert.match(value('events'), /deployment scheduled/)
}
const native = exports.DaemonRuntimeDetails({ row: { daemon_type: 'osd' } })
assert.equal(native.children[1].props.items[0].children.children[0], 'osd')
const page = readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('page.tsx', page, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const table = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'DaemonTable').getText(tree)
assert.ok(table.includes('<DaemonRuntimeDetails row={row} daemonType={row.type} />'))
for (const key of ['cpu_percentage', 'memory_usage', 'container_image', 'last_refresh', 'observed_at']) assert.ok(table.includes(`key: '${key}'`))
assert.ok(table.includes('row.stale !== false &&'))
console.log('MGR and MDS runtime details retain exact values, native events and unknown fields')

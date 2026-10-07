import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/monitoring/LogServiceStatus.tsx', import.meta.url), 'utf8')
const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.React } }).outputText
let clusterId = 7
let state = { data: { items: [], stale: false }, loading: false, error: '' }
let loader
const requests = []
const exports = {}
new Function('exports', 'require', 'React', code)(exports, name => ({
  antd: { Alert: 'Alert', Button: 'Button', Space: 'Space', Typography: { Title: 'Title', Text: 'Text' } },
  react: { useCallback: fn => fn },
  '../../components/AppTable': { AppTable: 'AppTable' },
  '../../components/ResourceMetaBar': { ResourceMetaBar: 'ResourceMetaBar' },
  '../../hooks': { useResource: fn => { loader = fn; return { ...state, refresh: async () => {} } } },
  '../../state/ClusterContext': { useClusterContext: () => ({ selectedClusterId: clusterId }) },
  '../../api/resource': { listAllResources: async (...args) => { requests.push(args); return state.data } }
}[name]), { createElement: (type, props, ...children) => ({ type, props, children }) })
let tree = exports.LogServiceStatus()
await loader()
assert.deepEqual(requests, [['/daemons', 7, { filters: { type: ['loki', 'promtail'] } }]])
assert.ok(JSON.stringify(tree).includes('本次库存未发现 loki'))
assert.ok(JSON.stringify(tree).includes('本次库存未发现 promtail'))
const items = [{ type: 'loki', name: 'loki.a', status: 'stopped' }, { type: 'loki', name: 'loki.b', status: 'running' }, { type: 'promtail', name: 'promtail.a', status: null }, { type: 'mgr', name: 'mgr.a' }]
assert.deepEqual(exports.logServiceRows(items), items.slice(0, 3))
state.data.items = items
tree = exports.LogServiceStatus()
const table = tree.children.find(child => child?.type === 'AppTable')
assert.equal(table.props.dataSource.length, 3)
const status = table.props.columns.find(column => column.dataIndex === 'status')
assert.equal(status.render('stopped'), 'stopped')
assert.equal(status.render(null), '未报告')
assert.equal(status.render(1), '未报告')
assert.ok(!JSON.stringify(tree).includes('本次库存未发现'))
for (const stale of [true, undefined]) {
  state.data = { items: [], stale }
  const view = JSON.stringify(exports.LogServiceStatus())
  assert.ok(view.includes('未确认有效的库存快照'))
  assert.ok(!view.includes('本次库存未发现'))
}
state = { data: { items: [], stale: false }, error: 'failed', loading: false }
assert.ok(!JSON.stringify(exports.LogServiceStatus()).includes('本次库存未发现'))
clusterId = undefined
exports.LogServiceStatus()
assert.equal(await loader(), undefined)
assert.equal(requests.length, 1)
console.log('Log services use complete scoped inventory and distinguish stale, missing and stopped instances')

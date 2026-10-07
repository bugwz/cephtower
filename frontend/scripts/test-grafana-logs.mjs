import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/monitoring/GrafanaLogsPanel.tsx', import.meta.url), 'utf8')
const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.React } }).outputText
let state = { data: 'https://grafana.test/explore', loading: false, error: '' }
let clusterId = 7
let loader
const requests = []
const exports = {}
new Function('exports', 'require', 'React', code)(exports, name => ({
  antd: { Alert: 'Alert', Button: 'Button', Card: 'Card', Space: 'Space' },
  react: { useCallback: fn => fn },
  '../../api/client': { isRecord: value => value !== null && typeof value === 'object' && !Array.isArray(value) },
  '../../hooks': { useResource: fn => { loader = fn; return { ...state, refresh: async () => {} } } },
  '../../state/ClusterContext': { useClusterContext: () => ({ selectedClusterId: clusterId }) },
  '../../api/external': { readExternalList: async (...args) => { requests.push(args); return { meta: { logs_explore_url: state.data } } } }
}[name]), { createElement: (type, props, ...children) => ({ type, props, children }) })
for (const value of [null, undefined, '', '/relative', 'javascript:alert(1)', 'https://user:secret@grafana.test', 'ftp://grafana.test']) assert.equal(exports.safeGrafanaLogsURL(value), undefined)
assert.equal(exports.safeGrafanaLogsURL('https://grafana.test/prefix/explore'), 'https://grafana.test/prefix/explore')
const buttons = tree => tree.children.flatMap(child => !child || typeof child !== 'object' ? [] : child.type === 'Button' ? [child] : buttons(child))
let tree = exports.GrafanaLogsPanel()
assert.equal(await loader(), state.data)
assert.deepEqual(requests, [['/grafana', 7]])
const link = buttons(tree).at(-1)
assert.equal(link.props.rel, 'noopener noreferrer')
assert.equal(link.props.href, state.data)
for (const patch of [{ loading: true }, { error: 'failed' }]) {
  state = { data: 'https://grafana.test/explore', loading: false, error: '', ...patch }
  const button = buttons(exports.GrafanaLogsPanel()).at(-1)
  assert.equal(button.props.disabled, true)
  assert.equal(button.props.href, undefined)
}
clusterId = undefined
const noCluster = buttons(exports.GrafanaLogsPanel()).at(-1)
assert.equal(noCluster.props.href, undefined)
assert.equal(noCluster.props.disabled, true)
assert.equal(await loader(), undefined)
assert.equal(requests.length, 1)
clusterId = 8
state.data = 'javascript:alert(1)'
exports.GrafanaLogsPanel()
await assert.rejects(loader(), /有效的日志入口/)
console.log('Grafana logs entry validates URLs, preserves cluster selection and hides failed links')

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/file/SMBOverview.tsx', import.meta.url), 'utf8')
const code = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.React, target: ts.ScriptTarget.ES2022 } }).outputText
let state = { data: 'https://grafana.test/d/feem6ehrmi2o0b/smb-overview', loading: false, error: '' }, cluster = 7, loader
const calls = [], exports = {}
new Function('exports', 'require', 'React', code)(exports, name => ({
  antd: { Alert: 'Alert', Button: 'Button', Card: 'Card', Space: 'Space' }, react: { useCallback: fn => fn },
  '../../api/client': { isRecord: v => v !== null && typeof v === 'object' && !Array.isArray(v) },
  '../../api/external': { readExternalList: async (...args) => { calls.push(args); return { meta: { smb_overview_url: state.data } } } },
  '../../hooks': { useResource: fn => { loader = fn; return { ...state, refresh: async () => {} } } },
  '../../state/ClusterContext': { useClusterContext: () => ({ selectedClusterId: cluster }) }
}[name]), { createElement: (type, props, ...children) => ({ type, props, children }) })
const buttons = node => node?.children?.flatMap(child => child?.type === 'Button' ? [child] : buttons(child)) ?? []
let tree = exports.SMBOverview()
assert.equal(await loader(), state.data)
assert.deepEqual(calls, [['/grafana', 7]])
assert.equal(buttons(tree)[1].props.href, state.data)
assert.equal(buttons(tree)[1].props.rel, 'noopener noreferrer')
for (const value of ['', null, '/relative', 'javascript:alert(1)', 'https://user:pass@grafana.test', 'ftp://grafana.test']) assert.equal(exports.smbOverviewURL(value), undefined)
for (const patch of [{ loading: true }, { error: 'failure' }, { data: undefined }]) {
  const saved = state; state = { ...state, ...patch }
  tree = exports.SMBOverview(); assert.equal(buttons(tree)[1].props.disabled, true); assert.equal(buttons(tree)[1].props.href, undefined)
  state = saved
}
cluster = undefined; exports.SMBOverview(); assert.equal(await loader(), undefined); assert.equal(calls.length, 1)
cluster = 8; state.data = 'javascript:alert(1)'; exports.SMBOverview(); await assert.rejects(loader())
console.log('SMB overview endpoint scope and safe link checks passed')

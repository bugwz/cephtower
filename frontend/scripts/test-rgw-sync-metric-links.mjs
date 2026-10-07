import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/object/RgwSyncMetricLinks.tsx', import.meta.url), 'utf8')
const exports = {}, visits = []
new Function('exports', 'require', 'React', ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.React, target: ts.ScriptTarget.ES2022 } }).outputText)(exports, name => ({
  antd: { Alert: 'Alert', Button: 'Button', Space: 'Space' },
  'react-router-dom': { useNavigate: () => path => visits.push(path) }
}[name]), { createElement: (type, props, ...children) => ({ type, props, children }) })
const nodes = node => node && typeof node === 'object' ? [node, ...(node.children ?? []).flat(Infinity).flatMap(nodes)] : []
const ids = ['rgw_sync_bytes_rate', 'rgw_sync_objects_rate', 'rgw_sync_errors_rate', 'rgw_sync_poll_latency_ms', 'rgw_sync_delta_seconds']
for (const clusterId of [undefined, 7]) {
  const rendered = nodes(exports.RgwSyncMetricLinks({ clusterId }))
  const buttons = rendered.filter(node => node.type === 'Button')
  assert.equal(buttons.length, 5)
  assert.ok(buttons.every(node => node.props.disabled === !clusterId))
  assert.ok(rendered.find(node => node.type === 'Alert').props.description.includes('不自动限定为此 Zone'))
  if (clusterId) buttons.forEach(button => button.props.onClick())
}
assert.deepEqual(visits, ids.map(id => `/monitoring/metric?metric=${id}`))
assert.ok(readFileSync(new URL('../src/pages/object/RgwSyncStatus.tsx', import.meta.url), 'utf8').includes('<RgwSyncMetricLinks clusterId={clusterId} />'))
console.log('RGW sync status links expose all sync metrics without claiming zone filtering')

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/cluster/HostNativeSummary.tsx', import.meta.url), 'utf8')
const output = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.React } }).outputText
const exports = {}
new Function('exports', 'require', 'React', output)(exports, name => name === 'antd' ? { Alert: 'Alert', Card: 'Card', Descriptions: { Item: 'Item' }, Space: 'Space', Typography: { Text: 'Text' } } : name.endsWith('HostPage') ? { hostNICCount: () => '0' } : { isRecord: value => value !== null && typeof value === 'object' && !Array.isArray(value) }, { createElement: (type, props, ...children) => ({ type, props, children }) })
for (const stale of [false, true, undefined]) {
  const tree = exports.HostNativeSummary({ host: { stale, native_summary: { server: 'Vendor Model', cpu_summary: '64C/128T', ram: '256 GiB', hdd_summary: '-', ssd_summary: 'N/A', os: 'Linux' } } })
  const text = JSON.stringify(tree)
  for (const expected of ['Vendor Model', '64C/128T', '256 GiB', 'N/A', 'Linux', '不代表精确容量']) assert.ok(text.includes(expected))
  assert.equal(text.includes('主机库存已过期'), stale !== false)
}
assert.ok(JSON.stringify(exports.HostNativeSummary({ host: {} })).includes('未报告'))
const details = readFileSync(new URL('../src/pages/cluster/HostDetailPage.tsx', import.meta.url), 'utf8')
assert.ok(details.includes('host && <HostNativeSummary host={host} />'))
console.log('Host detail renders native summaries with unit and freshness caveats')

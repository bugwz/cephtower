import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/TelemetryPage.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('telemetry.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const node = tree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'telemetryStatusValue')
const code = ts.transpileModule(node.getText(tree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const display = new Function(`${code}; return telemetryStatusValue`)()
for (const [value, expected] of [[false, '关闭'], [true, '开启'], [undefined, '未提供'], [null, '未提供'], ['', '空字符串'], [0, '0'], ['24', '24'], [{}, '未知']]) assert.equal(display(value), expected)
assert.ok(source.includes("request<{ status: ApiRecord, observed_at: string }>('/manager/telemetry/status'"))
assert.ok(!source.includes('mutateResource'))
console.log('Telemetry status display checks passed')

const reportSource = readFileSync(new URL('../src/pages/cluster/TelemetryReportPanel.tsx', import.meta.url), 'utf8')
const reportTree = ts.createSourceFile('report.tsx', reportSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const panel = reportTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'TelemetryReportPanel')
const load = panel.body.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'load')
const loadCode = ts.transpileModule(load.getText(reportTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const scenario of ['ok', 'abort', 'wrong-mode', 'error']) {
  const events = [], pending = { current: null }
  let resolve, reject, calls = 0
  const env = {
    pending, mode: 'preview', clusterId: 3,
    setLoading: (value) => events.push(['loading', value]), setReport: (value) => events.push(['report', value]), setError: (value) => events.push(['error', value]),
    jsonInit: (method, body, options) => ({ method, body, ...options }),
    request: (path, args) => { calls++; assert.equal(path, '/manager/telemetry/report'); assert.deepEqual(args.body, { cluster_id: 3, mode: 'preview' }); return new Promise((yes, no) => { resolve = yes; reject = no }) }
  }
  const invoke = new Function(...Object.keys(env), `${loadCode}; return load`)(...Object.values(env))
  const result = invoke(); await invoke(); assert.equal(calls, 1)
  if (scenario === 'abort') pending.current.abort()
  if (scenario === 'error') reject(new Error('offline'))
  else resolve({ mode: scenario === 'wrong-mode' ? 'current' : 'preview', report_json: '{"counter":9007199254740993}' })
  await result
  assert.equal(events.some(([kind, value]) => kind === 'report' && value?.report_json), scenario === 'ok')
  assert.equal(events.some(([kind, value]) => kind === 'error' && value !== ''), scenario === 'wrong-mode' || scenario === 'error')
  if (scenario !== 'abort') assert.equal(pending.current, null)
}
assert.ok(reportSource.includes('pending.current?.abort()'))
assert.ok(!reportSource.includes('JSON.parse'))
assert.ok(source.includes('<TelemetryReportPanel key={selectedClusterId}'))
console.log('Telemetry report manual loading and cancellation checks passed')

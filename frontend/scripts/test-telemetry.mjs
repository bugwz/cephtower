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

const controlSource = readFileSync(new URL('../src/pages/cluster/TelemetryControls.tsx', import.meta.url), 'utf8')
const controlTree = ts.createSourceFile('controls.tsx', controlSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const control = controlTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'TelemetryControls')
const submit = control.body.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'submit')
const submitCode = ts.transpileModule(submit.getText(controlTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const enabled of [true, false]) for (const accepted of [true, false]) for (const disabled of [true, false]) {
  const calls = [], env = {
    active: { current: true }, running: { current: false }, enabled, accepted, disabled, clusterId: 8,
    setBusy: () => {}, setOpen: () => {}, setAccepted: () => {}, message: { success: () => calls.push('success') }, onComplete: async () => calls.push('refresh'),
    mutateResource: async (path, method, body) => calls.push({ path, method, body })
  }
  const invoke = new Function(...Object.keys(env), `${submitCode}; return submit`)(...Object.values(env))
  await invoke()
  if (disabled || (!enabled && !accepted)) assert.deepEqual(calls, [])
  else assert.deepEqual(calls, [{ path: '/manager/telemetry', method: 'PATCH', body: { cluster_id: 8, enabled: !enabled, ...(!enabled ? { license: 'sharing-1-0' } : {}) } }, 'success', 'refresh'])
  assert.equal(env.running.current, false)
}
assert.ok(controlSource.includes('setAccepted(false); setOpen(true)'))
assert.ok(source.includes('key={`${selectedClusterId}:${status.enabled}`}'))
console.log('Telemetry explicit license consent checks passed')

const channelSource = readFileSync(new URL('../src/pages/cluster/TelemetryChannels.tsx', import.meta.url), 'utf8')
const channelTree = ts.createSourceFile('channels.tsx', channelSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const channelComponent = channelTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'TelemetryChannels')
const channelSubmit = channelComponent.body.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'submit')
const channelCode = ts.transpileModule(channelSubmit.getText(channelTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const scenario of ['ok', 'off', 'unknown', 'disabled', 'unmount', 'failure', 'refresh-failure']) {
  const calls = []
  let resolve, reject
  const env = {
    active: { current: true }, running: { current: false }, disabled: scenario === 'disabled', status: { enabled: scenario !== 'off' },
    channel: ['ident'], current: scenario === 'unknown' ? undefined : false, clusterId: 7,
    setBusy: () => {}, setSelected: () => {}, message: { success: () => calls.push('success') },
    onComplete: async () => { calls.push('refresh'); if (scenario === 'refresh-failure') throw new Error('refresh') },
    mutateResource: (path, method, body) => { calls.push({ path, method, body }); return new Promise((yes, no) => { resolve = yes; reject = no }) },
  }
  const invoke = new Function(...Object.keys(env), `${channelCode}; return submit`)(...Object.values(env))
  const result = invoke(); await invoke()
  if (['off', 'unknown', 'disabled'].includes(scenario)) { await result; assert.deepEqual(calls, []); continue }
  assert.deepEqual(calls, [{ path: '/manager/telemetry/channel', method: 'PATCH', body: { cluster_id: 7, channel: 'ident', enabled: true } }])
  if (scenario === 'unmount') env.active.current = false
  if (scenario === 'failure') reject(new Error('unverified')); else resolve()
  if (['failure', 'refresh-failure'].includes(scenario)) await assert.rejects(result); else await result
  assert.equal(calls.includes('success'), !['failure', 'unmount'].includes(scenario))
  assert.equal(calls.includes('refresh'), scenario !== 'unmount')
  assert.equal(env.running.current, false)
}
assert.ok(source.includes('key={`${selectedClusterId}:${data.observed_at}`}'))
console.log('Telemetry channel mutation, refresh and scope checks passed')

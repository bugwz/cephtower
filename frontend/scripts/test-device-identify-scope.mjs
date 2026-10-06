import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('pages.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const page = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'DeviceDetailContent')
const code = ts.transpileModule(page.body.statements.filter(n => ts.isFunctionDeclaration(n) && ['identify', 'confirmIdentify'].includes(n.name.text)).map(n => n.getText(tree)).join('\n'), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const scenario of ['success', 'stale', 'busy', 'before', 'during', 'failure']) {
  let modal
  const calls = [], messages = [], refreshes = []
  const env = {
    selectedClusterId: 7, device: { stale: scenario === 'stale' }, active: { current: true }, loading: false, error: '', pendingDeviceAction: '',
    zapRunning: { current: scenario === 'busy' }, identifyRunning: { current: false }, zapConfirmation: { current: null }, identifyConfirmation: { current: null },
    currentDeviceHost: 'node1', currentDevicePath: '/dev/sda', setPendingDeviceAction() {},
    Modal: { confirm: value => { modal = value; return { destroy() {} } } },
    operationMutation: { run: fn => fn() }, message: { success: value => messages.push(value) },
    mutateResource: async (...args) => { calls.push(args); if (scenario === 'during') env.active.current = false; if (scenario === 'failure') throw new Error('unconfirmed') },
    refresh: async () => refreshes.push(true)
  }
  const confirm = new Function(...Object.keys(env), `${code}; return confirmIdentify`)(...Object.values(env))
  confirm('on')
  if (['stale', 'busy'].includes(scenario)) { assert.equal(modal, undefined); continue }
  if (scenario === 'before') env.active.current = false
  try { await modal.onOk() } catch { assert.equal(scenario, 'failure') }
  await modal.onOk()
  assert.equal(calls.length, scenario === 'before' ? 0 : 1)
  if (calls.length) assert.deepEqual(calls[0], ['/device/identify', 'POST', { cluster_id: 7, host: 'node1', device: '/dev/sda', state: 'on', light: 'ident' }])
  assert.equal(refreshes.length, scenario === 'success' ? 1 : 0)
  assert.equal(messages.length, scenario === 'success' ? 1 : 0)
  assert.equal(env.identifyRunning.current, false)
}
console.log('Device identify confirms once and respects target lifetime and zap exclusion')

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('pages.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const page = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'DeviceDetailContent')
const node = page.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'zap')
const code = ts.transpileModule(node.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const expiryEffect = page.body.statements.find(node => ts.isExpressionStatement(node) && ts.isCallExpression(node.expression) && node.expression.expression.getText(tree) === 'useEffect' && node.getText(tree).includes('deviceConfirmationEpoch')).expression
assert.equal(expiryEffect.arguments[1].getText(tree), '[data, loading, error]')
for (const scenario of ['success', 'stale', 'unsafe', 'before', 'during', 'failure', 'cancelled', 'replaced', 'inventory-changed']) {
  let confirmation
  const calls = [], messages = [], refreshes = []
  const env = {
    deviceConfirmationEpoch: { current: 0 },
    selectedClusterId: 7, device: { stale: scenario === 'stale', resource_version: scenario === 'unsafe' ? Number.MAX_SAFE_INTEGER + 1 : '18446744073709551615' },
    active: { current: true }, loading: false, error: '', pendingDeviceAction: '', zapRunning: { current: false }, zapConfirmation: { current: null }, identifyRunning: { current: false }, identifyConfirmation: { current: null },
    currentDeviceHost: 'node1', currentDevicePath: '/dev/sda', setPendingDeviceAction() {},
    Modal: { confirm: options => { confirmation = options; return { destroy() {} } } },
    message: { error: value => messages.push(value), success: value => messages.push(value) },
    operationMutation: { run: fn => fn() },
    mutateResource: async (...args) => { calls.push(args); if (scenario === 'during') env.active.current = false; if (scenario === 'failure') throw new Error('unconfirmed') },
    refresh: async () => refreshes.push(true)
  }
  const zap = new Function(...Object.keys(env), `${code}; return zap`)(...Object.values(env))
  await zap()
  if (['stale', 'unsafe'].includes(scenario)) { assert.equal(confirmation, undefined); continue }
  if (scenario === 'before') env.active.current = false
  if (scenario === 'cancelled') confirmation.onCancel()
  if (scenario === 'replaced') {
    const old = confirmation
    await zap()
    confirmation = old
    old.onCancel() // Cancelling an old dialog must not invalidate the new one.
    assert.equal(env.deviceConfirmationEpoch.current, 2)
  }
  if (scenario === 'inventory-changed') {
    const effect = new Function(...Object.keys(env), `return (${expiryEffect.arguments[0].getText(tree)})`)(...Object.values(env))
    effect()()
    assert.equal(env.deviceConfirmationEpoch.current, 2)
  }
  try { await confirmation.onOk() } catch (error) { assert.equal(scenario, 'failure') }
  await confirmation.onOk()
  assert.equal(calls.length, ['before', 'cancelled', 'replaced', 'inventory-changed'].includes(scenario) ? 0 : 1)
  if (calls.length) assert.equal(calls[0][3].ifMatch, '18446744073709551615')
  assert.equal(refreshes.length, scenario === 'success' ? 1 : 0)
}
assert.ok(source.includes('key={JSON.stringify([selectedClusterId, deviceId, hostname, path])}'))
assert.ok(source.includes('zapConfirmation.current?.destroy()'))
console.log('Device zap preserves exact versions, target lifetime and single submission')

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/HostAddressEditor.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('address.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const component = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'HostAddressEditor')
const save = component.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'save')
const code = ts.transpileModule(save.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const scenario of ['success', 'failure', 'unmounted', 'blocked', 'invalid', 'unconfirmed', 'stale', 'busy', 'attempted']) {
  const calls = [], updates = [], scope = { current: {} }
  const env = { scope, running: { current: scenario === 'busy' }, attempted: scenario === 'attempted', blocked: scenario === 'blocked', valid: scenario !== 'invalid', accepted: scenario !== 'unconfirmed', version: scenario === 'stale' ? null : '18446744073709551615', clusterId: 3, hostname: 'node1', address: '2001:db8::1',
    setBusy: v => updates.push(['busy', v]), setAttempted: v => updates.push(['attempted', v]), setStatus: v => updates.push(['status', v]),
    mutateResource: async (...args) => { calls.push(args); if (scenario === 'unmounted') scope.current = null; if (scenario === 'failure') throw new Error('offline') } }
  await new Function(...Object.keys(env), `${code}; return save`)(...Object.values(env))()
  if (['blocked', 'invalid', 'unconfirmed', 'stale', 'busy', 'attempted'].includes(scenario)) { assert.deepEqual(calls, []); continue }
  assert.deepEqual(calls, [['/host', 'PATCH', { cluster_id: 3, host: 'node1', address: '2001:db8::1' }, { ifMatch: '18446744073709551615' }]])
  if (scenario === 'unmounted') assert.equal(updates.length, 3)
  if (scenario === 'failure') assert.ok(updates.some(([k,v]) => k === 'status' && v.includes('不要直接重试')))
}

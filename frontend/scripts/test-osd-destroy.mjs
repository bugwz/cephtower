import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/OSDDestroy.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('destroy.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const compile = node => ts.transpileModule(node.getText(tree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const helper = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'osdDestroyTarget')
const target = new Function(`${compile(helper)}; return osdDestroyTarget`)()
const record = { stale: false, up: false, state: ['exists'], uuid: '12345678-1234-1234-1234-123456789abc', resource_version: '18446744073709551615' }
assert.deepEqual(target(record, '0'), { version: record.resource_version, uuid: record.uuid })
for (const id of ['all', '01', '-1', '2147483648']) assert.equal(target(record, id), null)
for (const override of [{ stale: true }, { up: true }, { up: null }, { uuid: null }, { uuid: '' }, { uuid: '00000000-0000-0000-0000-000000000000' }, { state: [] }, { state: null }, { state: [null] }, { state: ['exists', 'exists'] }, { state: ['exists', 'destroyed'] }, { resource_version: '01' }, { resource_version: '18446744073709551616' }, { resource_version: 9007199254740992 }]) assert.equal(target({ ...record, ...override }, '0'), null)
for (const action of ['destroy', 'lost', 'purge']) {
const actionSource = action === 'destroy' ? source : readFileSync(new URL(`../src/pages/cluster/${action === 'lost' ? 'OSDLost' : 'OSDPurge'}.tsx`, import.meta.url), 'utf8')
const actionTree = ts.createSourceFile('action.tsx', actionSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const component = actionTree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === (action === 'destroy' ? 'OSDDestroy' : action === 'lost' ? 'OSDLost' : 'OSDPurge'))
const actionFunction = action === 'lost' ? 'markLost' : action
const destroy = component.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === actionFunction)
const actionCode = ts.transpileModule(destroy.getText(actionTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const expected = action === 'lost' ? 'mark lost osd.0' : `${action} osd.0`
assert.ok(actionSource.includes(`const target = ${action === 'purge' ? 'osdPurgeTarget' : 'osdDestroyTarget'}(record, osdId)`))
let actionTarget = target
if (action === 'purge') {
  const helper = actionTree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'osdPurgeTarget')
  const code = ts.transpileModule(helper.getText(actionTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
  actionTarget = new Function(`${code}; return osdPurgeTarget`)()
  assert.deepEqual(actionTarget(record, '0'), target(record, '0'))
  const destroyed = { ...record, state: ['exists', 'destroyed'], uuid: '00000000-0000-0000-0000-000000000000' }
  assert.deepEqual(actionTarget(destroyed, '0'), { uuid: destroyed.uuid, version: record.resource_version })
  for (const invalid of [{ ...destroyed, state: ['exists'] }, { ...record, state: ['destroyed'] }, { ...record, stale: true }, { ...record, up: true }, { ...record, uuid: null }, { ...record, state: [] }, { ...record, resource_version: '18446744073709551616' }]) assert.equal(actionTarget(invalid, '0'), null)
}
for (const scenario of ['success', 'failure', 'unmounted', 'busy', 'attempted', 'unconfirmed', 'wrong-text', 'stale', 'already-unmounted', 'double-click']) {
  const calls = [], updates = [], scope = { current: scenario === 'already-unmounted' ? null : {} }
  const env = { scope, running: { current: scenario === 'busy' }, attempted: scenario === 'attempted', accepted: scenario !== 'unconfirmed', confirmation: scenario === 'wrong-text' ? 'wrong osd.1' : expected, expected, target: scenario === 'stale' ? null : actionTarget(record, '0'), clusterId: 17, osdId: '0',
    setBusy: v => updates.push(['busy', v]), setAttempted: v => updates.push(['attempted', v]), setStatus: v => updates.push(['status', v]),
    mutateResource: async (...args) => { calls.push(args); if (scenario === 'unmounted') scope.current = null; if (scenario === 'failure') throw new Error('offline') } }
  const run = new Function(...Object.keys(env), `${actionCode}; return ${actionFunction}`)(...Object.values(env))
  if (scenario === 'double-click') await Promise.all([run(), run()]); else await run()
  if (['busy', 'attempted', 'unconfirmed', 'wrong-text', 'stale', 'already-unmounted'].includes(scenario)) { assert.deepEqual(calls, []); assert.deepEqual(updates, []); continue }
  assert.deepEqual(calls, [[`/osd/${action}`, 'POST', { cluster_id: 17, osd_id: '0', expected_uuid: record.uuid, confirmation: expected }, { ifMatch: record.resource_version }]])
  if (scenario === 'unmounted') assert.equal(updates.length, 3)
  if (scenario === 'success') assert.ok(updates.some(([k, v]) => k === 'status' && v.includes('回读确认')))
  if (scenario === 'failure') assert.ok(updates.some(([k, v]) => k === 'status' && v.includes('不要直接重试')))
}
}
const inspection = readFileSync(new URL('../src/pages/cluster/OSDInspection.tsx', import.meta.url), 'utf8')
assert.ok(inspection.includes('key={`${clusterId}:${osdId}:destroy`}'))
assert.ok(inspection.includes('key={`${clusterId}:${osdId}:lost`}'))
assert.ok(inspection.includes('key={`${clusterId}:${osdId}:purge`}'))

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/OSDIndividualFlags.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('flags.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const compile = node => ts.transpileModule(node.getText(tree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const flagsNode = tree.statements.find(ts.isVariableStatement)
const flags = new Function(`${compile(flagsNode)}; return individualOSDFlags`)()
assert.deepEqual(flags, ['noout', 'noin', 'nodown', 'noup'])
const parserNode = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'individualFlagSnapshot')
const parse = new Function('individualOSDFlags', `${compile(parserNode)}; return individualFlagSnapshot`)(flags)
const record = { stale: false, resource_version: '18446744073709551615', state: ['exists', 'up', 'noout'] }
assert.deepEqual(parse(record), { version: record.resource_version, flags: ['noout'] })
assert.deepEqual(parse({ ...record, state: [] }).flags, [])
for (const patch of [{ state: null }, { state: [null] }, { state: ['up', 'up'] }, { state: [' noout'] }, { stale: true }, { resource_version: '18446744073709551616' }, { resource_version: 9007199254740992 }]) assert.equal(parse({ ...record, ...patch }), null)
const component = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'OSDIndividualFlags')
const save = component.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'save')
for (const scenario of ['set', 'unset', 'failure', 'unmounted', 'busy', 'attempted', 'unconfirmed', 'invalid', 'stale']) {
  const calls = [], updates = [], scope = { current: {} }
  const env = { scope, running: { current: scenario === 'busy' }, attempted: scenario === 'attempted', accepted: scenario !== 'unconfirmed', snapshot: scenario === 'stale' ? null : parse(record), individualOSDFlags: flags, clusterId: 17, osdId: '0', flag: scenario === 'invalid' ? 'noscrub' : 'noout', action: scenario === 'unset' ? 'unset' : 'set',
    setBusy: v => updates.push(['busy', v]), setAttempted: v => updates.push(['attempted', v]), setStatus: v => updates.push(['status', v]),
    mutateResource: async (...args) => { calls.push(args); if (scenario === 'unmounted') scope.current = null; if (scenario === 'failure') throw new Error('offline') } }
  await new Function(...Object.keys(env), `${compile(save)}; return save`)(...Object.values(env))()
  if (['busy', 'attempted', 'unconfirmed', 'invalid', 'stale'].includes(scenario)) { assert.deepEqual(calls, []); assert.deepEqual(updates, []); continue }
  assert.deepEqual(calls, [['/osd/flag/individual', 'PATCH', { cluster_id: 17, osd_id: '0', flag: 'noout', action: env.action }, { ifMatch: record.resource_version }]])
  if (scenario === 'unmounted') assert.equal(updates.length, 3)
  if (scenario === 'set' || scenario === 'unset') assert.ok(updates.some(([k, v]) => k === 'status' && v.includes('回读确认')))
  if (scenario === 'failure') assert.ok(updates.some(([k, v]) => k === 'status' && v.includes('不要直接重试')))
}
assert.ok(readFileSync(new URL('../src/pages/cluster/OSDInspection.tsx', import.meta.url), 'utf8').includes('key={`${clusterId}:${osdId}:individual-flags`}'))

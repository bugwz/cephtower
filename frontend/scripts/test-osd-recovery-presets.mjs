import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/OSDRecoveryPresets.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('presets.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const compile = node => ts.transpileModule(node.getText(tree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const declarations = tree.statements.filter(ts.isVariableStatement)
const presets = new Function(`${declarations.map(compile).join('\n')}; return recoveryPresets`)()
assert.deepEqual(presets, { low: ['1', '1', '1', '0.5'], default: ['1', '3', '1', '0'], high: ['4', '4', '4', '0'] })
const planNode = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'recoveryPresetChange')
const build = new Function('configurationValueError', `${compile(planNode)}; return recoveryPresetChange`)(() => undefined)
const help = { name: 'osd_max_backfills', type: 'uint', flags: [] }
const row = { who: 'osd', name: help.name, value: '0', stale: false, resource_version: '9007199254740993' }
assert.equal(build(help.name, '1', help, [row]).version, '9007199254740993')
assert.equal(build(help.name, '1', help, []).version, undefined)
assert.equal(build(help.name, '1', help, [row]).previous, '0')
for (const rows of [[{ ...row, stale: true }], [{ ...row, resource_version: 9007199254740992 }], [row, row], [{ ...row, value: null }]]) assert.throws(() => build(help.name, '1', help, rows))
for (const metadata of [{}, { ...help, flags: ['no_mon_update'] }, { ...help, name: 'foreign' }]) assert.throws(() => build(help.name, '1', metadata, []))
const component = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'OSDRecoveryPresets')
const applyNode = component.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'apply')
for (const scenario of ['ok', 'failure', 'switched', 'unconfirmed', 'busy']) {
  const writes = [], statuses = [], scope = { current: {} }, running = { current: scenario === 'busy' }
  const env = { scope, running, clusterId: 7, accepted: scenario !== 'unconfirmed', plan: Array.from({ length: 4 }, (_, i) => ({ name: `option${i}`, value: '1', version: '9007199254740993' })),
    setBusy: () => {}, setPlan: () => {}, setAccepted: () => {}, setStatus: value => statuses.push(value),
    mutateResource: async (...args) => { writes.push(args); if (scenario === 'failure' && writes.length === 2) throw new Error('offline'); if (scenario === 'switched') scope.current = null },
    refreshResource: async () => statuses.push('refreshed'), onChanged: () => statuses.push('changed') }
  const run = new Function(...Object.keys(env), `${compile(applyNode)}; return apply`)(...Object.values(env))
  await run()
  assert.equal(writes.length, scenario === 'ok' ? 4 : scenario === 'failure' ? 2 : scenario === 'switched' ? 1 : 0)
  if (writes.length) assert.deepEqual(writes[0], ['/configuration/value', 'PUT', { cluster_id: 7, who: 'osd', name: 'option0', value: '1' }, { ifMatch: '9007199254740993' }])
  if (scenario === 'failure') assert.ok(statuses.at(-1).includes('已确认 1/4'))
  if (scenario === 'ok') assert.deepEqual(statuses.slice(-2), ['refreshed', 'changed'])
  if (scenario === 'switched') assert.equal(statuses.length, 1)
}

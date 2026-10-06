import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/OSDDeviceClass.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('class.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const compile = node => ts.transpileModule(node.getText(tree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const helper = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'osdClassVersion')
const version = new Function(`${compile(helper)}; return osdClassVersion`)()
const record = { stale: false, device_class: 'hdd', resource_version: '18446744073709551615' }
assert.equal(version(record), record.resource_version)
for (const row of [{ ...record, stale: true }, { ...record, device_class: null }, { ...record, resource_version: '18446744073709551616' }, { ...record, resource_version: 9007199254740992 }, { ...record, resource_version: '01' }]) assert.equal(version(row), null)
const component = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'OSDDeviceClass')
const save = component.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'save')
for (const scenario of ['success', 'failure', 'unmounted', 'busy', 'attempted', 'unconfirmed', 'invalid', 'stale']) {
  const calls = [], updates = [], scope = { current: {} }
  const env = { scope, running: { current: scenario === 'busy' }, attempted: scenario === 'attempted', accepted: scenario !== 'unconfirmed', valid: scenario !== 'invalid', version: scenario === 'stale' ? null : record.resource_version, record, clusterId: 17, osdId: '0', value: 'ssd',
    setBusy: v => updates.push(['busy', v]), setAttempted: v => updates.push(['attempted', v]), setStatus: v => updates.push(['status', v]),
    mutateResource: async (...args) => { calls.push(args); if (scenario === 'unmounted') scope.current = null; if (scenario === 'failure') throw new Error('offline') } }
  await new Function(...Object.keys(env), `${compile(save)}; return save`)(...Object.values(env))()
  if (['busy', 'attempted', 'unconfirmed', 'invalid', 'stale'].includes(scenario)) { assert.deepEqual(calls, []); assert.deepEqual(updates, []); continue }
  assert.deepEqual(calls, [['/osd/device/class', 'PUT', { cluster_id: 17, osd_id: '0', device_class: 'ssd', expected_class: 'hdd' }, { ifMatch: record.resource_version }]])
  if (scenario === 'unmounted') assert.equal(updates.length, 3)
  if (scenario === 'success') assert.ok(updates.some(([k, v]) => k === 'status' && v.includes('回读确认')))
  if (scenario === 'failure') assert.ok(updates.some(([k, v]) => k === 'status' && v.includes('不要直接重试')))
}
const inspection = readFileSync(new URL('../src/pages/cluster/OSDInspection.tsx', import.meta.url), 'utf8')
assert.ok(inspection.includes('key={`${clusterId}:${osdId}:device-class`}'))

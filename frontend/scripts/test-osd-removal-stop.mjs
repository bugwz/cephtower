import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/OSDRemovalStop.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('stop.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const compile = node => ts.transpileModule(node.getText(tree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const helper = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'removalStopTarget')
const parse = new Function(`${compile(helper)}; return removalStopTarget`)()
const record = { osd_id: 0, stale: false, resource_version: '18446744073709551615' }
assert.deepEqual(parse(record), { id: '0', version: record.resource_version })
for (const patch of [{ osd_id: null }, { osd_id: '0' }, { osd_id: -1 }, { osd_id: 2147483648 }, { stale: true }, { resource_version: '18446744073709551616' }, { resource_version: 9007199254740992 }]) assert.equal(parse({ ...record, ...patch }), null)
const component = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'OSDRemovalStop')
const node = component.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'stop')
for (const scenario of ['success', 'failure', 'refresh-failure', 'unmounted', 'busy', 'attempted', 'blocked', 'invalid']) {
  const calls = [], updates = [], scope = { current: {} }
  const env = { scope, running: { current: scenario === 'busy' }, attempted: scenario === 'attempted', blocked: scenario === 'blocked', target: scenario === 'invalid' ? null : parse(record), clusterId: 17,
    setBusy: v => updates.push(['busy', v]), setAttempted: v => updates.push(['attempted', v]), setStatus: v => updates.push(['status', v]), onChanged: () => calls.push(['changed']),
    mutateResource: async (...args) => { calls.push(['mutate', ...args]); if (scenario === 'unmounted') scope.current = null; if (scenario === 'failure') throw new Error('offline') },
    refreshResource: async (...args) => { calls.push(['refresh', ...args]); if (scenario === 'refresh-failure') throw new Error('refresh offline') } }
  await new Function(...Object.keys(env), `${compile(node)}; return stop`)(...Object.values(env))()
  if (['busy', 'attempted', 'blocked', 'invalid'].includes(scenario)) { assert.deepEqual(calls, []); assert.deepEqual(updates, []); continue }
  assert.deepEqual(calls[0], ['mutate', '/osd/removal/stop', 'POST', { cluster_id: 17, osd_id: '0' }, { ifMatch: record.resource_version }])
  if (scenario === 'unmounted' || scenario === 'failure') assert.equal(calls.length, 1)
  if (scenario === 'unmounted') assert.equal(updates.length, 3)
  if (scenario === 'success') assert.deepEqual(calls.slice(1), [['refresh', { clusterId: 17, kinds: ['osd', 'osd_removal'] }], ['changed']])
  if (scenario === 'refresh-failure') assert.ok(updates.some(([k, v]) => k === 'status' && v.includes('已确认退出队列，但库存刷新失败')))
}
const pages = readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8')
assert.ok(pages.includes('key={`${selectedClusterId}:${row.osd_id}:${row.resource_version}`}'))
assert.ok(pages.includes('blocked={loading || Boolean(error) || data?.removalMeta?.stale !== false}'))

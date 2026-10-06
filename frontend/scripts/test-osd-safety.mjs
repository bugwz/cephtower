import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/OSDSafetyCheck.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('safety.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const isRecord = value => value !== null && typeof value === 'object' && !Array.isArray(value)
const compile = node => ts.transpileModule(node.getText(tree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const parserNode = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'osdSafetyReport')
const parser = new Function('isRecord', `${compile(parserNode)}; return osdSafetyReport`)(isRecord)
const safe = { is_safe_to_destroy: true, safe_to_destroy: [0], active: [], missing_stats: [], stored_pgs: [] }
assert.equal(parser(safe, '0'), safe)
const unsafe = { ...safe, is_safe_to_destroy: false, safe_to_destroy: [], active: [0] }
assert.equal(parser(unsafe, '0'), unsafe)
for (const report of [null, {}, { ...safe, active: null }, { ...safe, safe_to_destroy: [null] }, { ...safe, active: [0] }, { ...safe, safe_to_destroy: [1] }, { ...safe, is_safe_to_destroy: 'true' }, { ...safe, safe_to_destroy: [] }]) assert.equal(parser(report, '0'), null)
const component = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'OSDSafetyCheck')
const check = component.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'check')
for (const scenario of ['safe', 'unsafe', 'invalid', 'failure', 'unmounted', 'busy']) {
  const calls = [], updates = [], scope = { current: {} }, running = { current: scenario === 'busy' }
  const env = { scope, running, clusterId: 17, osdId: '0', isRecord, osdSafetyReport: parser,
    setLoading: value => updates.push(['loading', value]), setError: value => updates.push(['error', value]), setReport: value => updates.push(['report', value]), setCheckedAt: value => updates.push(['time', value]),
    mutateResource: async (...args) => { calls.push(args); if (scenario === 'unmounted') scope.current = null; if (scenario === 'failure') throw new Error('offline'); return { details: { check: scenario === 'invalid' ? {} : scenario === 'unsafe' ? unsafe : safe } } } }
  const run = new Function(...Object.keys(env), `${compile(check)}; return check`)(...Object.values(env))
  await run()
  if (scenario === 'busy') { assert.deepEqual(calls, []); assert.deepEqual(updates, []); continue }
  assert.deepEqual(calls, [['/osd/removal/check', 'POST', { cluster_id: 17, osd_ids: ['0'] }]])
  if (scenario === 'unmounted') assert.equal(updates.length, 4)
  if (scenario === 'safe' || scenario === 'unsafe') assert.ok(updates.some(([kind, value]) => kind === 'report' && value === (scenario === 'safe' ? safe : unsafe)))
  if (scenario === 'failure' || scenario === 'invalid') assert.ok(updates.some(([kind, value]) => kind === 'error' && value !== ''))
  assert.equal(running.current, false)
}

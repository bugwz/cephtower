import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/monitoring/SilencedAlerts.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('SilencedAlerts.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const functions = tree.statements.filter(node => ts.isFunctionDeclaration(node) && ['silencedAlerts', 'SilencedAlerts'].includes(node.name.text))
const exports = {}
new Function('exports', 'require', 'Alert', 'SilencedAlertsContent', ts.transpileModule(functions.map(node => node.getText(tree)).join('\n'), { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX } }).outputText)(exports, () => ({ jsx: (type, props, key) => ({ type, props, key }) }), 'Alert', 'Content')
const linked = { fingerprint: 'one', status: { silencedBy: ['a', 'b'] } }
const other = { fingerprint: 'two', status: { silencedBy: ['ab'] } }
assert.deepEqual(exports.silencedAlerts([linked, other], 'a'), { items: [linked], incomplete: false })
assert.deepEqual(exports.silencedAlerts([], 'a'), { items: [], incomplete: false })
for (const row of [{}, { status: { silencedBy: null } }, { status: { silencedBy: ['a', 1] } }]) assert.deepEqual(exports.silencedAlerts([linked, row], 'a'), { items: [linked], incomplete: true })
for (const value of [undefined, null, {}, [null], [[]]]) assert.throws(() => exports.silencedAlerts(value, 'a'))
assert.throws(() => exports.silencedAlerts([], ''))
const first = exports.SilencedAlerts({ clusterId: 1, silenceId: 'a' })
assert.notEqual(first.key, exports.SilencedAlerts({ clusterId: 2, silenceId: 'a' }).key)
assert.notEqual(first.key, exports.SilencedAlerts({ clusterId: 1, silenceId: 'b' }).key)
assert.equal(exports.SilencedAlerts({ clusterId: 1, silenceId: undefined }).type, 'Alert')
assert.ok(source.includes('return () => controller.abort()'))
assert.ok(source.includes('if (!controller.signal.aborted) setResult(value)'))
assert.ok(source.includes('silencedAlerts(data.items, silenceId)'))
const list = readFileSync(new URL('../src/pages/ExternalListPage.tsx', import.meta.url), 'utf8')
assert.ok(list.includes('detailGeneration === clusterGeneration.current ? detailRow : null'))
assert.ok(list.includes('setDetailGeneration(clusterGeneration.current); setDetailRow(row)'))
assert.ok(list.includes('visibleDetail && selectedClusterId && definition.detailContent?.(visibleDetail, selectedClusterId)'))
const content = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'SilencedAlertsContent')
const effect = content.body.statements.find(node => ts.isExpressionStatement(node) && ts.isCallExpression(node.expression) && node.expression.expression.getText(tree) === 'useEffect').expression.arguments[0]
const effectCode = ts.transpileModule(`const effect = ${effect.getText(tree)}`, { compilerOptions: { target: ts.ScriptTarget.ES2020 } }).outputText
for (const scenario of ['success', 'failure', 'cancel']) {
  const updates = [], calls = []
  let resolve, reject
  const env = { clusterId: 7, silenceId: 'a', silencedAlerts: exports.silencedAlerts,
    setResult: value => updates.push(['result', value]), setError: value => updates.push(['error', value]),
    jsonInit: (method, body, init) => ({ method, body, ...init }),
    request: (path, init) => { calls.push({ path, init }); return new Promise((yes, no) => { resolve = yes; reject = no }) } }
  const cleanup = new Function(...Object.keys(env), `${effectCode}; return effect()`)(...Object.values(env))
  assert.equal(calls[0].path, '/alert/alerts')
  assert.deepEqual(calls[0].init.body, { cluster_id: 7 })
  if (scenario === 'cancel') cleanup()
  if (scenario === 'failure') reject(new Error('unavailable'))
  else resolve({ items: [linked] })
  await new Promise(done => setTimeout(done, 0))
  if (scenario === 'cancel') { assert.equal(calls[0].init.signal.aborted, true); assert.equal(updates.length, 2) }
  else if (scenario === 'failure') assert.deepEqual(updates[2], ['error', 'unavailable'])
  else assert.deepEqual(updates[2], ['result', { items: [linked], incomplete: false }])
}
console.log('Native silence-alert associations preserve exact IDs and isolate detail scopes')

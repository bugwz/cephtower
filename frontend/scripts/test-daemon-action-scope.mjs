import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const read = path => readFileSync(new URL(path, import.meta.url), 'utf8')
const parse = source => ts.createSourceFile('source.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const compile = source => ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const source = read('../src/pages/cluster/pages.tsx'), tree = parse(source)
const table = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'DaemonTable')
const action = table.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'runAction').getText(tree)
const apiTree = parse(read('../src/api/resource.ts'))
const api = apiTree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'applyDaemonAction').getText(apiTree).replace('export ', '')
for (const scenario of ['success', 'stale', 'unavailable', 'before', 'during', 'failure', 'no-cluster']) {
  const calls = [], messages = [], refreshes = []
  let finish
  const gate = new Promise(resolve => { finish = resolve })
  const apply = new Function('mutateResource', `${compile(api)}; return applyDaemonAction`)(async (...args) => {
    calls.push(args)
    await gate
    if (scenario === 'failure') throw new Error('failed')
  })
  const env = {
    clusterId: scenario === 'no-cluster' ? undefined : 7, active: { current: scenario !== 'before' }, running: { current: false },
    unavailable: scenario === 'unavailable', pendingDaemonAction: '', setPendingDaemonAction() {},
    textValue: value => value || '', applyDaemonAction: apply, operationMutation: { run: fn => fn() },
    message: { success: value => messages.push(value) }, refresh: () => refreshes.push(true)
  }
  const run = new Function(...Object.keys(env), `${compile(action)}; return runAction`)(...Object.values(env))
  const row = { name: 'mds.fs.node1', stale: scenario === 'stale' }
  const pending = run(row, 'restart')
  await run(row, 'restart')
  if (scenario === 'during') env.active.current = false
  finish()
  if (scenario === 'failure') await assert.rejects(pending, /failed/)
  else await pending
  const submitted = ['success', 'during', 'failure'].includes(scenario)
  assert.deepEqual(calls, submitted ? [['/daemon/action', 'POST', { cluster_id: 7, name: 'mds.fs.node1', action: 'restart', force: true }]] : [])
  assert.equal(messages.length, scenario === 'success' ? 1 : 0)
  assert.equal(refreshes.length, scenario === 'success' ? 1 : 0)
  assert.equal(env.running.current, false)
}
assert.equal(source.split('<DaemonTable key={selectedClusterId} clusterId={selectedClusterId}').length - 1, 2)
assert.ok(table.getText(tree).includes('return () => { active.current = false }'))
console.log('Daemon actions use explicit cluster scope and block duplicate or stale submissions')

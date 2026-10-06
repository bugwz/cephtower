import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const read = path => readFileSync(new URL(path, import.meta.url), 'utf8')
const parse = source => ts.createSourceFile('source.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const compile = source => ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const source = read('../src/pages/cluster/pages.tsx'), tree = parse(source)
const table = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'DaemonTable')
const action = table.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'runAction').getText(tree)
const versionSource = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'osdInventoryVersion').getText(tree)
const version = new Function(`${compile(versionSource)}; return osdInventoryVersion`)()
const apiTree = parse(read('../src/api/resource.ts'))
const api = apiTree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'applyDaemonAction').getText(apiTree).replace('export ', '')
for (const scenario of ['success', 'stale', 'unavailable', 'before', 'during', 'failure', 'no-cluster', 'unsafe-version', 'missing-version']) {
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
    textValue: value => value || '', osdInventoryVersion: version, applyDaemonAction: apply, operationMutation: { run: fn => fn() },
    message: { success: value => messages.push(value) }, refresh: () => refreshes.push(true)
  }
  const run = new Function(...Object.keys(env), `${compile(action)}; return runAction`)(...Object.values(env))
  const row = { name: 'mds.fs.node1', stale: scenario === 'stale', resource_version: scenario === 'unsafe-version' ? Number.MAX_SAFE_INTEGER + 1 : scenario === 'missing-version' ? undefined : '18446744073709551615' }
  const pending = run(row, 'restart')
  await run(row, 'restart')
  if (scenario === 'during') env.active.current = false
  finish()
  if (scenario === 'failure') await assert.rejects(pending, /failed/)
  else await pending
  const submitted = ['success', 'during', 'failure'].includes(scenario)
  assert.deepEqual(calls, submitted ? [['/daemon/action', 'POST', { cluster_id: 7, name: 'mds.fs.node1', action: 'restart' }, { ifMatch: '18446744073709551615' }]] : [])
  assert.equal(messages.length, scenario === 'success' ? 1 : 0)
  assert.equal(refreshes.length, scenario === 'success' ? 1 : 0)
  assert.equal(env.running.current, false)
}
assert.equal(source.split('<DaemonTable key={selectedClusterId} clusterId={selectedClusterId}').length - 1, 2)
assert.ok(table.getText(tree).includes('return () => { active.current = false; actionConfirmation.current?.destroy() }'))
const confirmCode = table.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'confirmAction').getText(tree)
for (const scenario of ['success', 'stale', 'inactive', 'during', 'failure', 'no-version']) {
  let modal
  const calls = [], destroys = []
  const env = { clusterId: 7, active: { current: scenario !== 'inactive' }, running: { current: false }, unavailable: false, osdInventoryVersion: version, actionConfirmation: { current: null }, Modal: { confirm: value => { modal = value; return { destroy: () => destroys.push(true) } } }, runAction: async (...args) => { calls.push(args); if (scenario === 'failure') throw new Error('failed') } }
  const confirm = new Function(...Object.keys(env), `${compile(confirmCode)}; return confirmAction`)(...Object.values(env))
  const row = { name: 'mgr.a', stale: scenario === 'stale', resource_version: scenario === 'no-version' ? undefined : '18446744073709551615' }
  confirm(row, 'stop')
  assert.equal(calls.length, 0)
  if (['stale', 'inactive', 'no-version'].includes(scenario)) { assert.equal(modal, undefined); continue }
  assert.match(modal.title, /mgr.a.*7/)
  assert.match(modal.content, /不会使用强制选项/)
  if (scenario === 'during') env.active.current = false
  if (scenario === 'failure') await assert.rejects(modal.onOk(), /failed/)
  else await modal.onOk()
  await modal.onOk()
  assert.equal(calls.length, scenario === 'during' ? 0 : 1)
}
for (const action of ['start', 'stop', 'restart']) assert.ok(table.getText(tree).includes(`confirmAction(row, '${action}')`))
const openPerf = table.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'openPerformance').getText(tree)
for (const name of ['mgr.a', 'mds.fs.node-1', 'osd.1', 'mgr.*', 'mds.', 'mgr.a;stop', '', null]) {
  for (const state of ['active', 'inactive', 'unavailable', 'no-cluster']) {
    const selections = []
    const env = { clusterId: state === 'no-cluster' ? undefined : 7, active: { current: state !== 'inactive' }, unavailable: state === 'unavailable', setPerfSelection: value => selections.push(value) }
    const open = new Function(...Object.keys(env), `${compile(openPerf)}; return openPerformance`)(...Object.values(env))
    open({ name })
    assert.deepEqual(selections, state === 'active' && ['mgr.a', 'mds.fs.node-1'].includes(name) ? [{ clusterId: 7, name }] : [])
  }
}
assert.ok(table.getText(tree).includes('perfSelection?.clusterId === clusterId'))
assert.ok(table.getText(tree).includes('key={JSON.stringify([visiblePerf.clusterId, visiblePerf.name])}'))
assert.ok(table.getText(tree).includes('onCancel={() => setPerfSelection(null)}'))
console.log('Daemon actions use explicit cluster scope and block duplicate or stale submissions')

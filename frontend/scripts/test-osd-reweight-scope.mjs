import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('page.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const component = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'ReweightForm')
const submit = component.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'submit')
const js = ts.transpileModule(submit.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const scenario of ['ok', 'switched-before', 'switched-after', 'busy', 'invalid', 'failure']) {
  let current = scenario !== 'switched-before'
  const calls = [], running = { current: scenario === 'busy' }
  const env = { running, isCurrent: () => current, clusterId: 7, osdID: '12',
    setSubmitting: () => {}, message: { error: () => calls.push('error'), success: () => calls.push('success') },
    operationMutation: { run: fn => fn() },
    reweightOSD: async (...args) => { calls.push(args); if (scenario === 'switched-after') current = false; if (scenario === 'failure') throw new Error('failed') },
    refreshResource: async body => calls.push(body), refresh: async () => calls.push('read') }
  const run = new Function(...Object.keys(env), `${js}; return submit`)(...Object.values(env))
  const result = run({ weight: scenario === 'invalid' ? NaN : 0 })
  if (scenario === 'failure') await assert.rejects(result, /failed/); else await result
  if (['switched-before', 'busy'].includes(scenario)) assert.deepEqual(calls, [])
  if (scenario === 'invalid') assert.deepEqual(calls, ['error'])
  if (scenario === 'ok') assert.deepEqual(calls, [[7, '12', 0], 'success', { clusterId: 7, kind: 'osd' }, 'read'])
  if (scenario === 'switched-after') assert.deepEqual(calls, [[7, '12', 0]])
}
assert.ok(!component.getText(tree).includes('requiredClusterId'))
assert.ok(!component.getText(tree).includes('Modal.destroyAll'))
const initial = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'osdReweightInitial')
const initialJS = ts.transpileModule(initial.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const initialValue = new Function(`${initialJS}; return osdReweightInitial`)()
for (const value of [0, 0.12345, 1]) assert.equal(initialValue(value), value)
for (const value of [undefined, null, -1, 2, NaN, Infinity, '0.5']) assert.equal(initialValue(value), undefined)
assert.ok(component.getText(tree).includes('weight: osdReweightInitial(currentWeight)'))
assert.ok(!component.getText(tree).includes('precision={2}'))
const page = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'OsdManagementPage')
const selection = page.body.statements.flatMap(n => ts.isVariableStatement(n) ? [...n.declarationList.declarations] : []).find(n => n.name.getText(tree) === 'inspectedOSD')
const readInspection = new Function('osdInspection', 'osdScope', `return ${selection.initializer.getText(tree)}`)
const first = { clusterId: 7 }, other = { clusterId: 8 }, returned = { clusterId: 7 }
const row = { id: 0, up: true }, snapshot = { scope: first, row }
assert.equal(readInspection(snapshot, first), row)
assert.equal(readInspection(snapshot, other), null)
assert.equal(readInspection(snapshot, returned), null)
assert.equal(readInspection(null, first), null)
assert.ok(page.getText(tree).includes('setOSDInspection({ scope: osdScope, row })'))
assert.ok(page.getText(tree).includes('inspectedOSD && selectedClusterId && <OSDInspection'))
const actionNode = page.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'runOSDAction')
const actionJS = ts.transpileModule(actionNode.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.React } }).outputText
for (const action of ['in', 'out', 'scrub', 'deep-scrub']) {
 for (const timing of ['current', 'before', 'after']) {
  const scope = {}, scopeRef = { current: scope }, calls = []
  if (timing === 'before') scopeRef.current = {}
  const write = async (...args) => { calls.push(args); if (timing === 'after') scopeRef.current = {} }
  const env = { selectedClusterId: 7, osdScope: scope, osdScopeRef: scopeRef, osdActionRunning: { current: false }, pendingOSDAction: '', setPendingOSDAction: () => {}, operationMutation: { run: fn => fn() }, scrubOSD: write, markOSD: write, refreshResource: async () => calls.push('collect'), refresh: async () => calls.push('read'), message: { success: () => calls.push('success') } }
  await new Function(...Object.keys(env), `${actionJS}; return runOSDAction`)(...Object.values(env))('12', action)
  if (timing === 'before') assert.deepEqual(calls, [])
  else {
   assert.deepEqual(calls[0], [7, '12', action.includes('scrub') ? action === 'deep-scrub' : action])
   assert.deepEqual(calls.slice(1), timing === 'after' ? [] : ['success', 'collect', 'read'])
  }
 }
}
const deleteNode = page.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'deleteOSD')
const deleteJS = ts.transpileModule(deleteNode.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
let confirm
const deleteScope = {}, deleteRef = { current: deleteScope }, writes = []
const deleteEnv = { selectedClusterId: 7, osdScope: deleteScope, osdScopeRef: deleteRef, osdID: () => '0', Modal: { confirm: options => { confirm = options } }, operationMutation: { run: fn => fn() }, mutateResource: async (...args) => writes.push(args), message: { error: () => {}, success: () => {} }, refreshResource: async () => {}, refresh: async () => {} }
await new Function(...Object.keys(deleteEnv), `${deleteJS}; return deleteOSD`)(...Object.values(deleteEnv))({ resource_version: 1 })
deleteRef.current = {}
await assert.rejects(confirm.onOk(), /集群已切换/)
assert.deepEqual(writes, [])
const inspectionSource = readFileSync(new URL('../src/pages/cluster/OSDInspection.tsx', import.meta.url), 'utf8')
const inspectionTree = ts.createSourceFile('inspection.tsx', inspectionSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const epochNode = inspectionTree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'osdHistoryEpoch')
const epochJS = ts.transpileModule(epochNode.getText(inspectionTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const epoch = new Function(`${epochJS}; return osdHistoryEpoch`)()
for (const value of [0, 1, 4294967295]) assert.equal(epoch(value), String(value))
for (const value of [null, undefined, -1, 4294967296, 0.5, '1']) assert.equal(epoch(value), '未采集或格式无效')

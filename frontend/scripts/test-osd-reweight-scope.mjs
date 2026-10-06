import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const flagsSource = readFileSync(new URL('../src/api/resource.ts', import.meta.url), 'utf8')
const flagsTree = ts.createSourceFile('resource.ts', flagsSource, ts.ScriptTarget.Latest, true)
const flagsNode = flagsTree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'listOSDFlags')
const flagsJS = ts.transpileModule(flagsNode.getText(flagsTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const flags of [undefined, null, [], ['noout'], [''], [null], [1], 'noout', [' noout']]) {
  const calls = []
  const readFlags = new Function('getOptionalResource', 'toRecord', `${flagsJS}; return listOSDFlags`)(async (...args) => { calls.push(args); return { item: { data: { flags } } } }, value => value)
  assert.deepEqual(await readFlags(17), Array.isArray(flags) && flags.every(f => typeof f === 'string' && f.length > 0 && f.trim() === f) ? flags : null)
  assert.equal(calls[0][1], 17)
}
const source = readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('page.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const component = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'ReweightForm')
const submit = component.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'submit')
const js = ts.transpileModule(submit.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const scenario of ['ok', 'switched-before', 'switched-after', 'busy', 'invalid', 'failure']) {
  let current = scenario !== 'switched-before'
  const calls = [], running = { current: scenario === 'busy' }
  const env = { running, isCurrent: () => current, clusterId: 7, osdID: '12', version: '18446744073709551615',
    setSubmitting: () => {}, message: { error: () => calls.push('error'), success: () => calls.push('success') },
    operationMutation: { run: fn => fn() },
    reweightOSD: async (...args) => { calls.push(args); if (scenario === 'switched-after') current = false; if (scenario === 'failure') throw new Error('failed') },
    refreshResource: async body => calls.push(body), refresh: async () => calls.push('read') }
  const run = new Function(...Object.keys(env), `${js}; return submit`)(...Object.values(env))
  const result = run({ weight: scenario === 'invalid' ? NaN : 0 })
  if (scenario === 'failure') await assert.rejects(result, /failed/); else await result
  if (['switched-before', 'busy'].includes(scenario)) assert.deepEqual(calls, [])
  if (scenario === 'invalid') assert.deepEqual(calls, ['error'])
  if (scenario === 'ok') assert.deepEqual(calls, [[7, '12', 0, '18446744073709551615'], 'success', { clusterId: 7, kind: 'osd' }, 'read'])
  if (scenario === 'switched-after') assert.deepEqual(calls, [[7, '12', 0, '18446744073709551615']])
}
const reweightNode = flagsTree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'reweightOSD')
const reweightJS = ts.transpileModule(reweightNode.getText(flagsTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const reweightCalls = []
await new Function('mutateResource', `${reweightJS}; return reweightOSD`)((...args) => reweightCalls.push(args))(7, '12', 0, '18446744073709551615')
assert.deepEqual(reweightCalls, [['/osd/action', 'POST', { cluster_id: 7, osd_id: '12', action: 'reweight', weight: 0 }, { ifMatch: '18446744073709551615' }]])
assert.ok(source.includes('version={expectedVersion}'))
assert.ok(source.includes("row.stale === false ? osdInventoryVersion(row.resource_version) ?? undefined : undefined"))
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
for (const action of ['in', 'out', 'down', 'scrub', 'deep-scrub']) {
 for (const timing of ['current', 'before', 'after']) {
  const scope = {}, scopeRef = { current: scope }, calls = []
  if (timing === 'before') scopeRef.current = {}
  const write = async (...args) => { calls.push(args); if (timing === 'after') scopeRef.current = {} }
  const env = { selectedClusterId: 7, osdScope: scope, osdScopeRef: scopeRef, osdActionRunning: { current: false }, pendingOSDAction: '', setPendingOSDAction: () => {}, operationMutation: { run: fn => fn() }, scrubOSD: write, mutateResource: write, refreshResource: async () => calls.push('collect'), refresh: async () => calls.push('read'), message: { success: () => calls.push('success') } }
  await new Function(...Object.keys(env), `${actionJS}; return runOSDAction`)(...Object.values(env))('12', action, undefined, '18446744073709551615')
  if (timing === 'before') assert.deepEqual(calls, [])
  else {
   assert.deepEqual(calls[0], action.includes('scrub') ? [7, '12', action === 'deep-scrub'] : ['/osd/action', 'POST', { cluster_id: 7, osd_id: '12', action }, { ifMatch: '18446744073709551615' }])
   assert.deepEqual(calls.slice(1), timing === 'after' ? [] : ['success', 'collect', 'read'])
  }
 }
}
const deleteNode = page.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'deleteOSD')
const downNode = page.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'confirmOSDState')
const downJS = ts.transpileModule(downNode.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const scenario of ['current', 'switched', 'stale', 'unknown', 'already-down']) {
  const scope = {}, ref = { current: scope }, calls = []
  let modal
  const env = { selectedClusterId: 7, osdScope: scope, osdScopeRef: ref, loading: false, error: '', osdActionRunning: { current: false }, osdID: () => '0', osdInventoryVersion: () => '18446744073709551615', message: { error: () => {} }, Modal: { confirm: v => { modal = v } }, runOSDAction: async (...args) => calls.push(args) }
  new Function(...Object.keys(env), `${downJS}; return confirmOSDState`)(...Object.values(env))({ stale: scenario === 'stale', up: scenario === 'unknown' ? null : scenario !== 'already-down' }, 'down')
  if (['stale', 'unknown', 'already-down'].includes(scenario)) { assert.equal(modal, undefined); continue }
  assert.ok(modal.content.includes('不会停止 OSD 进程'))
  if (scenario === 'switched') { ref.current = {}; await assert.rejects(modal.onOk(), /集群已切换/); assert.deepEqual(calls, []) }
  else { await modal.onOk(); assert.deepEqual(calls, [['0', 'down', undefined, '18446744073709551615']]) }
}
assert.ok(actionNode.getText(tree).includes("mutateResource('/osd/action', 'POST', { cluster_id: selectedClusterId, osd_id: id, action }, { ifMatch: expectedVersion })"))
for (const action of ['in', 'out']) {
  for (const state of [true, false, null, undefined]) {
    const scope = {}, calls = []
    let modal
    const env = { selectedClusterId: 7, osdScope: scope, osdScopeRef: { current: scope }, loading: false, error: '', osdActionRunning: { current: false }, osdID: () => '0', osdInventoryVersion: () => '3', message: { error: () => {} }, Modal: { confirm: v => { modal = v } }, runOSDAction: async (...args) => calls.push(args) }
    new Function(...Object.keys(env), `${downJS}; return confirmOSDState`)(...Object.values(env))({ stale: false, in: state }, action)
    if (state === (action === 'out')) {
      assert.ok(modal.title.includes(action)); await modal.onOk()
      assert.deepEqual(calls, [['0', action, undefined, '3']])
    } else assert.equal(modal, undefined)
  }
}
const deleteJS = ts.transpileModule(deleteNode.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
let confirm
const deleteScope = {}, deleteRef = { current: deleteScope }, writes = []
const versionNode = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'osdInventoryVersion')
const versionJS = ts.transpileModule(versionNode.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const osdInventoryVersion = new Function(`${versionJS}; return osdInventoryVersion`)()
const deleteEnv = { loading: false, error: '', osdInventoryVersion, selectedClusterId: 7, osdScope: deleteScope, osdScopeRef: deleteRef, osdID: () => '0', Modal: { confirm: options => { confirm = options } }, operationMutation: { run: fn => fn() }, mutateResource: async (...args) => writes.push(args), message: { error: () => {}, success: () => {} }, refreshResource: async () => {}, refresh: async () => {} }
const removeOSD = new Function(...Object.keys(deleteEnv), `${deleteJS}; return deleteOSD`)(...Object.values(deleteEnv))
await removeOSD({ resource_version: 1, stale: false })
deleteRef.current = {}
await assert.rejects(confirm.onOk(), /集群已切换/)
assert.deepEqual(writes, [])
deleteRef.current = deleteScope
for (const row of [{ stale: true, resource_version: 1 }, { resource_version: 1 }, { stale: false }, { stale: false, resource_version: 9007199254740992 }]) {
  confirm = null; await removeOSD(row); assert.equal(confirm, null)
}
await removeOSD({ stale: false, resource_version: '9007199254740993' })
await confirm.onOk()
assert.deepEqual(writes[0], ['/osd', 'DELETE', { cluster_id: 7, osd_id: '0', zap: false, preserve_id: false }, { ifMatch: '9007199254740993' }])
assert.ok(confirm.content.includes('不保留 ID'))
await removeOSD({ stale: false, resource_version: '18446744073709551615' }, true)
assert.ok(confirm.title.includes('替换（保留 ID）'))
assert.ok(confirm.content.includes('不代表替换已完成'))
await confirm.onOk()
assert.deepEqual(writes[1], ['/osd', 'DELETE', { cluster_id: 7, osd_id: '0', zap: false, preserve_id: true }, { ifMatch: '18446744073709551615' }])
await removeOSD({ stale: false, resource_version: 1 }, true)
deleteRef.current = {}
await assert.rejects(confirm.onOk(), /集群已切换/)
assert.equal(writes.length, 2)
deleteRef.current = deleteScope
assert.ok(source.includes('onClick={() => deleteOSD(row, true)}'))
for (const value of [null, undefined, 0, -1, 1.5, '0', '1.5', '18446744073709551616']) assert.equal(osdInventoryVersion(value), null)
assert.equal(osdInventoryVersion(1), '1')
const inspectionSource = readFileSync(new URL('../src/pages/cluster/OSDInspection.tsx', import.meta.url), 'utf8')
const inspectionTree = ts.createSourceFile('inspection.tsx', inspectionSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const epochNode = inspectionTree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'osdHistoryEpoch')
const epochJS = ts.transpileModule(epochNode.getText(inspectionTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const epoch = new Function(`${epochJS}; return osdHistoryEpoch`)()
for (const value of [0, 1, 4294967295]) assert.equal(epoch(value), String(value))
for (const value of [null, undefined, -1, 4294967296, 0.5, '1']) assert.equal(epoch(value), '未采集或格式无效')
const networkSource = readFileSync(new URL('../src/pages/cluster/OSDNetwork.tsx', import.meta.url), 'utf8')
const networkTree = ts.createSourceFile('network.tsx', networkSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const helpers = networkTree.statements.filter(n => ts.isFunctionDeclaration(n) && n.name.text.startsWith('osdAddress')).map(n => n.getText(networkTree).replace('export ', '')).join('\n')
const networkJS = ts.transpileModule(helpers, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const { osdAddressRows, osdAddressText, osdAddressNonce } = new Function(`${networkJS}; return { osdAddressRows, osdAddressText, osdAddressNonce }`)()
for (const value of [null, undefined, [], {}, { addrvec: null }, { addrvec: [null] }]) assert.equal(osdAddressRows(value), null)
assert.deepEqual(osdAddressRows({ addrvec: [] }), [])
const address = { type: 'v2', addr: '[::1]:3300', nonce: 0 }
assert.deepEqual(osdAddressRows({ addrvec: [address, address] }), [{ ...address, index: 0 }, { ...address, index: 1 }])
assert.equal(osdAddressText('javascript:alert(1)'), 'javascript:alert(1)')
assert.equal(osdAddressText(''), '空字符串（原生）')
assert.equal(osdAddressNonce(0), '0')
assert.equal(osdAddressNonce(4294967295), '4294967295')
for (const value of [null, -1, 4294967296, 1.5, '0']) assert.equal(osdAddressNonce(value), '未采集或格式无效')
assert.ok(!networkSource.includes('href='))
const stateNode = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'osdStateText')
const stateJS = ts.transpileModule(stateNode.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const stateText = new Function(`${stateJS}; return osdStateText`)()
assert.equal(stateText(['exists', 'up', 'autoout']), 'exists、up、autoout')
assert.equal(stateText([]), '本次未返回状态标记')
for (const value of [null, undefined, 'up', [null], ['']]) assert.equal(stateText(value), '未采集或格式无效')
const usageSource = readFileSync(new URL('../src/pages/cluster/OSDUsage.tsx', import.meta.url), 'utf8')
const usageTree = ts.createSourceFile('usage.tsx', usageSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const usageNode = usageTree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'osdUsageInteger')
const usageJS = ts.transpileModule(usageNode.getText(usageTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const usageInteger = new Function(`${usageJS}; return osdUsageInteger`)()
for (const value of ['0', '9007199254740993', '18446744073709551615']) assert.equal(usageInteger(value), value)
for (const value of [0, null, undefined, '-1', '1.5', '18446744073709551616']) assert.equal(usageInteger(value), '未采集或格式无效')
const latencyNode = usageTree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'osdLatency')
const latencyJS = ts.transpileModule(latencyNode.getText(usageTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const latency = new Function(`${latencyJS}; return osdLatency`)()
for (const value of [0, 0.125, 25]) assert.equal(latency(value), String(value))
for (const value of [null, undefined, -1, NaN, Infinity, '1']) assert.equal(latency(value), '未采集或格式无效')
const snapshotNode = usageTree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'osdSnapshotValue')
const snapshotJS = ts.transpileModule(snapshotNode.getText(usageTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const snapshotValue = new Function('osdUsageInteger', 'osdLatency', `${snapshotJS}; return osdSnapshotValue`)(usageInteger, latency)
assert.equal(snapshotValue({ stats: { kb: '9007199254740993' } }, 'stats', 'kb'), '9007199254740993')
assert.equal(snapshotValue({ stats: { pgs: '0' } }, 'stats', 'pgs'), '0')
assert.equal(snapshotValue({ stats: { utilization: 0 } }, 'stats', 'utilization'), '0%')
assert.equal(snapshotValue({ perf_stats: { apply_latency_ms: 0.125 } }, 'perf_stats', 'apply_latency_ms'), '0.125')
for (const stats of [null, undefined, [], { kb: 12 }, { kb: null }]) assert.equal(snapshotValue({ stats }, 'stats', 'kb'), '未采集或格式无效')
for (const field of ['kb', 'kb_used', 'kb_avail', 'utilization', 'pgs', 'commit_latency_ms', 'apply_latency_ms']) assert.ok(page.getText(tree).includes(`'${field}')`))

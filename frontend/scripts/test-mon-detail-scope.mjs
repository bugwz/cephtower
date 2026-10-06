import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/cluster/MonDetailPage.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('detail.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const wrapper = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'MonDetailPage')
const content = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'MonDetailContent')
assert.ok(wrapper.getText(tree).includes('key={JSON.stringify([selectedClusterId, monName])}'))
assert.ok(wrapper.getText(tree).includes('selectedClusterId={selectedClusterId} monName={monName}'))
assert.ok(content.getText(tree).includes("useState('')"))
assert.ok(content.getText(tree).includes('useResource(loader)'))
const loader = content.body.statements.find(node => ts.isVariableStatement(node) && node.declarationList.declarations[0].name.getText(tree) === 'loader').declarationList.declarations[0].initializer.arguments[0]
const code = ts.transpileModule(`const load = ${loader.getText(tree)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const [selectedClusterId, monName] of [[1, 'a'], [2, 'a'], [2, 'b'], [undefined, 'a'], [1, '']]) {
  const calls = []
  const load = new Function('selectedClusterId', 'monName', 'listResource', 'listMonitorPerfCounters', 'textValue', `${code}; return load`)(selectedClusterId, monName,
    async (path, cluster, options) => { calls.push({ path, cluster, options }); return { items: [{ name: 'other' }, { name: monName }] } },
    async (monitor, cluster) => { calls.push({ monitor, cluster }); return [{ name: 'counter' }] }, value => value)
  const result = await load()
  if (!selectedClusterId || !monName) { assert.equal(result, null); assert.equal(calls.length, 0); continue }
  assert.equal(result.mon.name, monName)
  assert.deepEqual(result.counters, [{ name: 'counter' }])
  assert.deepEqual(calls, [{ path: '/monitors', cluster: selectedClusterId, options: { name: monName } }, { monitor: monName, cluster: selectedClusterId }])
}
console.log('MON detail state is keyed by cluster and monitor, with scoped reads')

const refreshNode = content.body.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'refreshMonDetail')
const refreshJS = ts.transpileModule(refreshNode.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2020 } }).outputText
for (const scenario of ['inactive', 'switched', 'current', 'failed']) {
  const active = { current: scenario !== 'inactive' }
  const calls = []
  const mutation = { run: async (callback, success) => {
    assert.equal(success, false)
    await callback()
    if (scenario === 'switched') active.current = false
    if (scenario === 'failed') throw new Error('collection failed')
  } }
  const refreshMon = new Function('active', 'selectedClusterId', 'monName', 'refreshing', 'setRefreshing', 'operationMutation', 'refreshResource', 'refresh', 'message', `${refreshJS}; return refreshMonDetail`)(active, 7, 'mon-a', false, value => calls.push(['busy', value]), mutation, async body => calls.push(['collect', body.clusterId, body.kinds]), async () => calls.push(['reload']), { success: () => calls.push(['success']) })
  if (scenario === 'failed') await assert.rejects(refreshMon, /collection failed/)
  else await refreshMon()
  const collect = ['collect', 7, ['mon', 'mon_status', 'mon_perf_counter']]
  if (scenario === 'inactive') assert.deepEqual(calls, [])
  if (scenario === 'switched') assert.deepEqual(calls, [['busy', true], collect])
  if (scenario === 'current') assert.deepEqual(calls, [['busy', true], collect, ['success'], ['reload'], ['busy', false]])
  if (scenario === 'failed') assert.deepEqual(calls, [['busy', true], collect, ['busy', false]])
}
assert.ok(source.includes('return () => { active.current = false }'))
console.log('MON refresh stops follow-up work after leaving its mounted scope')
for (const failure of [new Error('counter page failed'), new Error(''), 'unstructured failure']) {
  const load = new Function('selectedClusterId', 'monName', 'listResource', 'listMonitorPerfCounters', 'textValue', `${code}; return load`)(1, 'a',
    async () => ({ items: [{ name: 'a', rank: 0 }] }), async () => { throw failure }, value => value)
  const result = await load()
  assert.equal(result.mon.rank, 0)
  assert.deepEqual(result.counters, [])
  assert.equal(result.counterError, failure instanceof Error && failure.message ? failure.message : '性能计数器请求失败')
}
const primaryFailure = new Function('selectedClusterId', 'monName', 'listResource', 'listMonitorPerfCounters', 'textValue', `${code}; return load`)(1, 'a',
  async () => { throw new Error('monitor request failed') }, async () => [], value => value)
await assert.rejects(primaryFailure(), /monitor request failed/)
assert.ok(source.includes('data?.counterError ? <Alert'))
assert.ok(source.includes('description={data.counterError}'))

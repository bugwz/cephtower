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

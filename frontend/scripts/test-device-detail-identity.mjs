import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('pages.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const page = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'DeviceDetailContent')
const loader = page.body.statements.find(n => ts.isVariableStatement(n) && n.declarationList.declarations[0].name.getText(tree) === 'loader').declarationList.declarations[0].initializer.arguments[0]
const code = ts.transpileModule(`const loader = ${loader.getText(tree)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const scenario of ['unique', 'duplicate', 'missing', 'failure', 'no-cluster']) {
  const calls = []
  const target = { device_id: 'serial', hostname: 'node1', path: '/dev/sda' }
  const items = scenario === 'missing' ? [] : [target]
  if (scenario === 'duplicate') items.push({ ...target, hostname: 'node2' })
  items.push({ device_id: 'other' })
  const env = {
    selectedClusterId: scenario === 'no-cluster' ? undefined : 7,
    decodedDeviceId: 'serial', normalizeDeviceRow: row => row, deviceID: row => row.device_id,
    listAllResources: async (...args) => {
      calls.push(args)
      if (scenario === 'failure') throw new Error('later page failed')
      return { items, stale: true, staleReason: 'inventory stale', observedAt: 'timestamp' }
    }
  }
  const run = new Function(...Object.keys(env), `${code}; return loader`)(...Object.values(env))
  if (['duplicate', 'failure'].includes(scenario)) {
    await assert.rejects(run, scenario === 'duplicate' ? /无法唯一确定/ : /later page failed/)
  } else {
    const result = await run()
    assert.equal(result.device, scenario === 'unique' ? target : null)
    if (scenario !== 'no-cluster') {
      assert.equal(result.stale, true)
      assert.equal(result.staleReason, 'inventory stale')
      assert.equal(result.observedAt, 'timestamp')
    }
  }
  assert.deepEqual(calls, scenario === 'no-cluster' ? [] : [['/devices', 7, { filters: { device_id: ['serial'] } }]])
}
console.log('Device detail uses complete scoped inventory and rejects ambiguous identities')

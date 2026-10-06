import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/OSDGlobalFlags.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('flags.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const compile = node => ts.transpileModule(node.getText(tree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const flags = new Function(`${compile(tree.statements.find(ts.isVariableStatement))}; return globalOSDFlags`)()
assert.deepEqual(flags.map(([name]) => name), ['noin', 'noout', 'noup', 'nodown', 'pause', 'noscrub', 'nodeep-scrub', 'nobackfill', 'norebalance', 'norecover'])
const parserNode = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'globalFlagSnapshot')
const parse = new Function('isRecord', `${compile(parserNode)}; return globalFlagSnapshot`)(v => v !== null && typeof v === 'object' && !Array.isArray(v))
const item = { stale: false, resource_version: '18446744073709551615', data: { flags: ['pauserd', 'pausewr', 'sortbitwise'] } }
assert.deepEqual(parse(item), { version: item.resource_version, flags: item.data.flags })
for (const value of [null, {}, { ...item, stale: true }, { ...item, resource_version: 9007199254740992 }, { ...item, data: { flags: null } }, { ...item, data: { flags: ['noout', 'noout'] } }]) assert.equal(parse(value), null)
const component = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'OSDGlobalFlags')
for (const method of ['read', 'save']) {
  const node = component.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === method)
  for (const scenario of ['success', 'failure', 'unmounted', 'busy']) {
    const calls = [], updates = [], scope = { current: {} }
    const env = { scope, running: { current: scenario === 'busy' }, snapshot: parse(item), accepted: true, globalOSDFlags: flags, flag: 'pause', action: 'unset', clusterId: 17, globalFlagSnapshot: parse,
      setBusy: v => updates.push(['busy', v]), setSnapshot: v => updates.push(['snapshot', v]), setAccepted: v => updates.push(['accepted', v]), setStatus: v => updates.push(['status', v]),
      refreshResource: async (...args) => { calls.push(['refresh', ...args]); if (scenario === 'unmounted') scope.current = null; if (scenario === 'failure') throw new Error('offline') },
      getResource: async (...args) => { calls.push(['get', ...args]); return { item } },
      mutateResource: async (...args) => { calls.push(['mutate', ...args]); if (scenario === 'unmounted') scope.current = null; if (scenario === 'failure') throw new Error('offline') } }
    await new Function(...Object.keys(env), `${compile(node)}; return ${method}`)(...Object.values(env))()
    if (scenario === 'busy') { assert.deepEqual(calls, []); assert.deepEqual(updates, []); continue }
    if (method === 'save') assert.deepEqual(calls, [['mutate', '/osd/flag', 'PATCH', { cluster_id: 17, flag: 'pause', action: 'unset' }, { ifMatch: item.resource_version }]])
    else assert.deepEqual(calls, scenario === 'success' ? [['refresh', { clusterId: 17, kinds: ['osd_flag'] }], ['get', '/osd/flag', 17]] : [['refresh', { clusterId: 17, kinds: ['osd_flag'] }]])
    if (scenario === 'unmounted') assert.equal(updates.length, 4)
  }
}
assert.ok(readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8').includes('<OSDGlobalFlags key={selectedClusterId} clusterId={selectedClusterId} />'))

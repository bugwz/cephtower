import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/api/resource.ts', import.meta.url), 'utf8')
const tree = ts.createSourceFile('resource.ts', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS)
const code = ts.transpileModule(tree.statements.filter(node => ts.isFunctionDeclaration(node) && ['listAllResources', 'listMonitorPerfCounters'].includes(node.name?.text)).map(node => node.getText(tree).replace('export ', '')).join('\n'), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
function loader(listResource) {
  return new Function('listResource', `${code}; return listMonitorPerfCounters`)(listResource)
}
const calls = []
const first = Array.from({ length: 500 }, (_, index) => ({ name: `counter-${index}` }))
const load = loader(async (path, cluster, options) => {
  calls.push({ path, cluster, options })
  return options.cursor ? { items: [{ name: 'last-counter' }], nextCursor: null, stale: false } : { items: [...first], nextCursor: 'next-page', stale: false }
})
const items = await load('mon.with-special-name', 7)
assert.equal(items.length, 501)
assert.equal(items[500].name, 'last-counter')
assert.equal(calls.length, 2)
for (const call of calls) {
  assert.equal(call.path, '/monitor/perf/counters')
  assert.equal(call.cluster, 7)
  assert.equal(call.options.limit, 500)
  assert.deepEqual(call.options.body, { monitor: 'mon.with-special-name' })
}
assert.equal(calls[1].options.cursor, 'next-page')
await assert.rejects(loader(async () => ({ items: [], nextCursor: 'loop', stale: false }))('a', 7), /分页游标重复/)
await assert.rejects(loader(async (_path, _cluster, options) => {
  if (options.cursor) throw new Error('second page failed')
  return { items: [{ name: 'partial' }], nextCursor: 'next', stale: false }
})('a', 7), /second page failed/)
assert.deepEqual(await loader(async () => ({ items: [], nextCursor: null, stale: false }))('a', 7), [])
console.log('MON counter loader follows every cursor and preserves cluster and monitor scope')

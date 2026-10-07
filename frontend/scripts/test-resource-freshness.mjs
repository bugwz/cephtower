import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/api/resource.ts', import.meta.url), 'utf8')
const tree = ts.createSourceFile('resource.ts', source, ts.ScriptTarget.Latest, true)
const functions = tree.statements.filter(node => ts.isFunctionDeclaration(node) && ['listResource', 'listAllResources'].includes(node.name.text))
const code = ts.transpileModule(functions.map(node => node.getText(tree).replace('export ', '')).join('\n'), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const create = pages => new Function('request', 'jsonInit', 'toRecord', `${code}; return { listResource, listAllResources }`)(async () => pages.shift(), (_method, body) => body, value => value ?? {})
const row = stale => ({ data: { name: 'pool' }, stale })
for (const metaStale of [false, true, undefined, null]) {
  for (const rowStale of [false, true, undefined, null]) {
    const result = await create([{ items: [row(rowStale)], meta: { stale: metaStale } }]).listResource('/pools', 7)
    assert.equal(result.stale, metaStale !== false || rowStale !== false)
    assert.equal(result.items[0].stale, rowStale)
  }
  const empty = await create([{ items: [], meta: { stale: metaStale } }]).listResource('/pools', 7)
  assert.equal(empty.stale, metaStale !== false)
}
for (const stale of [false, true, undefined, null]) assert.equal((await create([row(stale)]).listResource('/pool', 7)).stale, stale !== false)
for (const badPage of [0, 1]) {
  const pages = [0, 1].map(index => ({ items: [row(index === badPage ? undefined : false)], meta: { stale: false }, pagination: { next_cursor: index === 0 ? 'next' : null } }))
  const result = await create(pages).listAllResources('/pools', 7)
  assert.equal(result.items.length, 2)
  assert.equal(result.stale, true)
}
const fresh = await create([{ items: [row(false)], meta: { stale: false }, pagination: { next_cursor: 'next' } }, { items: [], meta: { stale: false } }]).listAllResources('/pools', 7)
assert.equal(fresh.stale, false)
console.log('Resource freshness requires explicit page and row evidence and survives pagination')
for (const payload of [null, [], 'invalid', 1]) await assert.rejects(create([payload]).listResource('/pools', 7), /资源响应格式无效/)
for (const items of [null, undefined, {}, 'invalid', [null], [false], [[]], [row(false), 1]]) {
  await assert.rejects(create([{ items, meta: { stale: false } }]).listResource('/pools', 7), /资源列表格式无效/)
  await assert.rejects(create([{ items: [row(false)], meta: { stale: false }, pagination: { next_cursor: 'next' } }, { items, meta: { stale: false } }]).listAllResources('/pools', 7), /资源列表格式无效/)
}
for (const next_cursor of [0, 1, false, {}, []]) await assert.rejects(create([{ items: [], meta: { stale: false }, pagination: { next_cursor } }]).listAllResources('/pools', 7), /资源分页游标格式无效/)
for (const next_cursor of [undefined, null, '']) assert.equal((await create([{ items: [], meta: { stale: false }, pagination: { next_cursor } }]).listAllResources('/pools', 7)).items.length, 0)
console.log('Malformed resource responses are rejected rather than presented as empty or complete inventory')

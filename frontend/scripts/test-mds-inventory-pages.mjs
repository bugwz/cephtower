import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const read = path => readFileSync(new URL(path, import.meta.url), 'utf8')
const parse = source => ts.createSourceFile('source.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const compile = source => ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const source = read('../src/pages/cluster/pages.tsx')
const tree = parse(source)
const page = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'MdsManagementPage')
const loader = page.body.statements.filter(ts.isVariableStatement).flatMap(n => [...n.declarationList.declarations]).find(n => n.name.getText(tree) === 'loader').initializer.arguments[0].getText(tree)
const resourceTree = parse(read('../src/api/resource.ts'))
const all = resourceTree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'listAllResources').getText(resourceTree).replace('export ', '')
const filterTree = parse(read('../src/hooks/useResourceTableFilters.ts'))
const merge = filterTree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'mergeResourceFilters').getText(filterTree).replace('export ', '')
const mergeFilters = new Function(`${compile(merge)}; return mergeResourceFilters`)()
for (const scenario of ['fresh', 'stale', 'unknown', 'failure', 'no-cluster']) {
  const calls = []
  const listResource = async (path, cluster, options) => {
    calls.push({ path, cluster, options })
    if (scenario === 'failure' && path === '/daemons' && options.cursor) throw new Error('second page failed')
    return {
      items: [{ name: `${path}:${options.cursor || 'first'}` }], nextCursor: options.cursor ? null : 'second',
      stale: scenario === 'unknown' ? undefined : scenario === 'stale' && Boolean(options.cursor),
      staleReason: scenario === 'stale' && options.cursor ? 'snapshot expired' : null
    }
  }
  const loadAll = new Function('listResource', `${compile(all)}; return listAllResources`)(listResource)
  const env = { selectedClusterId: scenario === 'no-cluster' ? undefined : 42, listAllResources: loadAll, mergeResourceFilters: mergeFilters, serviceTableFilters: { filters: { service_name: ['mds.fs'] } }, daemonTableFilters: { filters: { hostname: ['node1'] } } }
  const run = new Function(...Object.keys(env), `${compile(`const load = ${loader}`)}; return load`)(...Object.values(env))
  if (scenario === 'failure') await assert.rejects(run(), /second page failed/)
  else {
    const result = await run()
    assert.equal(result.services.length, scenario === 'no-cluster' ? 0 : 2)
    assert.equal(result.daemons.length, scenario === 'no-cluster' ? 0 : 2)
    assert.equal(result.inventoryWarnings.length, ['stale', 'unknown'].includes(scenario) ? 2 : 0)
    if (scenario === 'stale') assert.ok(result.inventoryWarnings.every(value => value.includes('snapshot expired')))
    if (scenario === 'unknown') assert.ok(result.inventoryWarnings.every(value => value.includes('新鲜度未知')))
  }
  assert.equal(calls.length, scenario === 'no-cluster' ? 0 : 4)
  for (const { path, cluster, options } of calls) {
    assert.equal(cluster, 42)
    assert.deepEqual(options.filters, path === '/services' ? { service_type: ['mds'], service_name: ['mds.fs'] } : { daemon_type: ['mds'], hostname: ['node1'] })
  }
}
assert.ok(page.getText(tree).includes('data?.inventoryWarnings.map'))
console.log('MDS inventories retain all pages, scope, errors and independent freshness warnings')

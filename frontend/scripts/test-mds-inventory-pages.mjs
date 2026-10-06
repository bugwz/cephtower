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
  const env = { selectedClusterId: scenario === 'no-cluster' ? undefined : 42, listAllResources: loadAll, mergeResourceFilters: mergeFilters, serviceTableFilters: { filters: { name: ['mds.fs'], type: ['rgw'] } }, daemonTableFilters: { filters: { hostname: ['node1'], type: ['osd'] } } }
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
    assert.deepEqual(options.filters, path === '/services' ? { type: ['mds'], name: ['mds.fs'] } : { type: ['mds'], hostname: ['node1'] })
  }
}
assert.ok(page.getText(tree).includes('data?.inventoryWarnings.map'))
let selectService
function findSelection(node) {
  if (ts.isArrowFunction(node) && node.body.getText(tree).includes('setServiceSelection({ scope: serviceScope, name: row.name })') && !ts.isJsxElement(node.body)) selectService = node
  ts.forEachChild(node, findSelection)
}
findSelection(page)
assert.ok(selectService)
for (const name of ['mds.fs', 'mds.fs-a.node', 'mds.', 'mds.*', 'mds.a;stop', 'rgw.a', null]) {
  for (const state of ['ready', 'loading', 'error', 'no-cluster']) {
    const selections = [], scope = { clusterId: 7 }
    const env = { selectedClusterId: state === 'no-cluster' ? undefined : 7, loading: state === 'loading', error: state === 'error' ? 'failed' : '', row: { name }, serviceScope: scope, setServiceSelection: value => selections.push(value) }
    new Function(...Object.keys(env), `${compile(`const select = ${selectService.getText(tree)}`)}; select()`)(...Object.values(env))
    assert.deepEqual(selections, state === 'ready' && ['mds.fs', 'mds.fs-a.node'].includes(name) ? [{ scope, name }] : [])
  }
}
assert.ok(page.getText(tree).includes('serviceSelection?.scope === serviceScope'))
assert.ok(page.getText(tree).includes('<ServiceDaemons key={JSON.stringify([selectedClusterId, visibleService.name])} clusterId={selectedClusterId} name={visibleService.name} />'))
for (const name of ['MdsManagementPage', 'MgrManagementPage', 'DaemonTable']) {
  const node = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === name)
  assert.doesNotMatch(node.getText(tree), /service_name|service_type|daemon_name|daemon_type|status_desc/)
}
const mgr = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'MgrManagementPage')
const mgrLoader = mgr.body.statements.filter(ts.isVariableStatement).flatMap(n => [...n.declarationList.declarations]).find(n => n.name.getText(tree) === 'loader').initializer.arguments[0].getText(tree)
const mgrCalls = []
const list = async (...args) => { mgrCalls.push(args); return { items: [] } }
const mgrEnv = { selectedClusterId: 9, listResource: list, listAllResources: list, mergeResourceFilters: mergeFilters, moduleTableFilters: { filters: {} }, daemonTableFilters: { filters: { type: ['osd'], name: ['mgr.a'] } } }
await new Function(...Object.keys(mgrEnv), `${compile(`const load = ${mgrLoader}`)}; return load`)(...Object.values(mgrEnv))()
assert.deepEqual(mgrCalls.find(call => call[0] === '/daemons'), ['/daemons', 9, { filters: { type: ['mgr'], name: ['mgr.a'] } }])
for (const scenario of ['fresh', 'stale', 'unknown', 'failure', 'no-cluster']) {
  const calls = []
  const listResource = async (path, cluster, options) => {
    calls.push({ path, cluster, options })
    if (scenario === 'failure' && path === '/daemons' && options.cursor) throw new Error('mgr second page failed')
    return {
      items: [{ name: `${path}:${options.cursor || 'first'}` }], nextCursor: options.cursor ? null : 'second',
      stale: scenario === 'unknown' ? undefined : scenario === 'stale' && Boolean(options.cursor),
      staleReason: scenario === 'stale' && options.cursor ? `${path} snapshot expired` : null
    }
  }
  const loadAll = new Function('listResource', `${compile(all)}; return listAllResources`)(listResource)
  const env = { ...mgrEnv, listAllResources: loadAll, selectedClusterId: scenario === 'no-cluster' ? undefined : 9 }
  const run = new Function(...Object.keys(env), `${compile(`const load = ${mgrLoader}`)}; return load`)(...Object.values(env))
  if (scenario === 'failure') await assert.rejects(run(), /mgr second page failed/)
  else {
    const result = await run()
    assert.equal(result.modules.length, scenario === 'no-cluster' ? 0 : 2)
    assert.equal(result.daemons.length, scenario === 'no-cluster' ? 0 : 2)
    assert.equal(result.inventoryWarnings.length, ['stale', 'unknown'].includes(scenario) ? 2 : 0)
    if (scenario === 'stale') {
      assert.match(result.inventoryWarnings[0], /\/manager\/modules snapshot expired/)
      assert.match(result.inventoryWarnings[1], /\/daemons snapshot expired/)
    }
  }
  assert.equal(calls.length, scenario === 'no-cluster' ? 0 : 4)
  for (const { path, cluster, options } of calls) {
    assert.equal(cluster, 9)
    assert.deepEqual(options.filters, path === '/daemons' ? { type: ['mgr'], name: ['mgr.a'] } : {})
  }
}
assert.ok(mgr.getText(tree).includes('data?.inventoryWarnings.map'))
const daemonTable = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'DaemonTable').getText(tree)
for (const key of ['name', 'type', 'status', 'hostname', 'version']) assert.ok(daemonTable.includes(`key: '${key}'`))
console.log('MDS inventories retain all pages, scope, errors and independent freshness warnings')

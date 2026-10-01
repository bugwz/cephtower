import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/cluster/CrushMapPage.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('CrushMapPage.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const functions = tree.statements.filter((node) => ts.isFunctionDeclaration(node) && ['crushTree', 'crushStatus'].includes(node.name.text))
const exports = {}
new Function('exports', ts.transpileModule(functions.map((fn) => fn.getText(tree)).join('\n'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const nodes = [{ id: 0, name: 'osd.0', type: 'osd', status: 'up' }, { id: -1, name: 'default', type: 'root', children: [0] }]
const result = exports.crushTree({ nodes, roots: [-1] })
assert.equal(result[0].title, 'default (root)')
assert.equal(result[0].children[0].key, '-1/0')
assert.equal(result[0].children[0].title, 'osd.0 (osd)')
assert.equal(result[0].children[0].nodeId, 0)
assert.equal(result[0].children[0].status, 'up')
for (const status of ['up', 'in']) assert.equal(exports.crushStatus(status).color, 'success')
for (const status of ['down', 'out', 'destroyed']) assert.equal(exports.crushStatus(status).color, 'error')
for (const status of [undefined, null, '', 1, true, 'unrecognized']) assert.equal(exports.crushStatus(status).color, 'default')
assert.equal(exports.crushStatus(null).label, '未知')
assert.equal(exports.crushStatus('unrecognized').label, 'unrecognized')
assert.deepEqual(exports.crushTree({ nodes: [], roots: [] }), [])
assert.throws(() => exports.crushTree({ nodes, roots: [99] }))
assert.throws(() => exports.crushTree({ nodes: [{ id: -1, name: 'root', type: 'root', children: [-1] }], roots: [-1] }))
assert.equal(nodes[0].id, 0, 'must not reorder native metadata')
console.log('CRUSH topology tree checks passed')

const usageSource = readFileSync(new URL('../src/pages/cluster/ErasureProfileUsage.tsx', import.meta.url), 'utf8')
const usageTree = ts.createSourceFile('usage.tsx', usageSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const usageFn = usageTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'erasureProfileUsage')
const usageExports = {}
new Function('exports', ts.transpileModule(usageFn.getText(usageTree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(usageExports)
assert.deepEqual(usageExports.erasureProfileUsage('ec', { stale: false, items: [{ name: 'a', erasure_code_profile: 'ec', stale: false }, { name: 'b', erasure_code_profile: 'other', stale: false }] }), { names: ['a'], stale: false })
assert.deepEqual(usageExports.erasureProfileUsage('ec', { stale: true, items: [] }), { names: [], stale: true })
assert.equal(usageExports.erasureProfileUsage('ec', { stale: false, items: [{ name: 'a' }] }).stale, true)
assert.ok(usageSource.includes("listAllResources('/pools', clusterId)"))
const crushUsageFn = usageTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'crushRuleUsage')
new Function('exports', ts.transpileModule(crushUsageFn.getText(usageTree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(usageExports)
assert.deepEqual(usageExports.crushRuleUsage('ssd', 8, { stale: false, items: [{ name: 'by-id', crush_rule: 8, stale: false }, { name: 'by-name', crush_rule: 'ssd', stale: false }, { name: 'other', crush_rule: 0, stale: false }] }), { names: ['by-id', 'by-name'], stale: false })
assert.deepEqual(usageExports.crushRuleUsage('ssd', 8, { stale: true, items: [] }), { names: [], stale: true })

const watcher = tree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'watchCrushMap')
const watcherExports = {}
let pending, scheduled, cancelled = false, received = 0
new Function('exports', 'request', 'jsonInit', 'crushTree', 'setTimeout', 'clearTimeout', ts.transpileModule(watcher.getText(tree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(watcherExports, () => new Promise((resolve) => { pending = resolve }), () => ({}), exports.crushTree, (fn, delay) => { assert.equal(delay, 5000); scheduled = fn; return 1 }, () => { cancelled = true })
const stopWatch = watcherExports.watchCrushMap(1, true, () => received++, () => assert.fail('unexpected error'))
assert.equal(scheduled, undefined, 'do not overlap pending reads')
pending({ nodes, roots: [-1] })
await new Promise((resolve) => setImmediate(resolve))
assert.equal(received, 1)
assert.equal(typeof scheduled, 'function')
scheduled()
stopWatch()
pending({ nodes, roots: [-1] })
await new Promise((resolve) => setImmediate(resolve))
assert.equal(received, 1, 'ignore late response after cleanup')
assert.equal(cancelled, true)

const profileSource = readFileSync(new URL('../src/pages/cluster/ErasureProfilesPanel.tsx', import.meta.url), 'utf8')
assert.ok(profileSource.includes("path: '/erasure/code/profiles'"))
for (const field of ['name', 'plugin', 'k', 'm', 'technique', 'crush-root', 'crush-failure-domain', 'crush-device-class']) {
  assert.ok(profileSource.includes(`key: '${field}'`), `missing native profile field ${field}`)
}
assert.ok(source.includes('<ErasureProfilesPanel />'))
const profileTree = ts.createSourceFile('ErasureProfilesPanel.tsx', profileSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const profileGuard = profileTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'erasureProfileDeleteBlocked')
const profileExports = {}
new Function('exports', ts.transpileModule(profileGuard.getText(profileTree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(profileExports)
assert.equal(profileExports.erasureProfileDeleteBlocked({ name: 'ec', stale: false }), undefined)
for (const row of [{ name: 'ec' }, { name: 'ec', stale: true }, { name: '', stale: false }, { name: null, stale: false }]) assert.ok(profileExports.erasureProfileDeleteBlocked(row))
assert.ok(profileSource.includes("path: '/erasure/code/profile'"))
assert.ok(profileSource.includes("action: 'erasure_code_profile.delete'"))
const detailFunction = profileTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'erasureProfileDetails')
new Function('exports', ts.transpileModule(detailFunction.getText(profileTree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(profileExports)
const profileDetails = profileExports.erasureProfileDetails({ k: '4', m: '2', 'crush-num-failure-domains': '0', l: 0 })
for (const key of ['plugin', 'k', 'm', 'technique', 'l', 'c', 'd', 'scalar_mds', 'packetsize', 'crush-root', 'crush-failure-domain', 'crush-locality', 'crush-num-failure-domains', 'crush-osds-per-failure-domain', 'crush-device-class', 'directory']) assert.ok(profileDetails.some((item) => item.key === key))
assert.equal(profileDetails.find((item) => item.key === 'l').children, '0')
assert.equal(profileDetails.find((item) => item.key === 'crush-num-failure-domains').children, '0')
assert.equal(profileDetails.find((item) => item.key === 'plugin').children, '未提供')

const poolSource = readFileSync(new URL('../src/pages/cluster/PoolManagementPage.tsx', import.meta.url), 'utf8')
assert.ok(poolSource.includes("listAllResources('/pools', selectedClusterId, { filters: poolTableFilters.filters })"), 'pool table must consume all filtered inventory pages')
const resourceSource = readFileSync(new URL('../src/api/resource.ts', import.meta.url), 'utf8')
const resourceTree = ts.createSourceFile('resource.ts', resourceSource, ts.ScriptTarget.Latest, true)
const allPagesFn = resourceTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'listAllResources')
const allPagesCode = ts.transpileModule(allPagesFn.getText(resourceTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const pageCalls = []
const loadAllPages = new Function('listResource', `${allPagesCode}; return listAllResources`)(async (path, cluster, options) => {
  pageCalls.push({ path, cluster, options })
  return options.cursor
    ? { items: [{ name: 'second' }], stale: true, staleReason: 'expired', observedAt: '2026-01-01', nextCursor: null }
    : { items: [{ name: 'first' }], stale: false, observedAt: '2026-01-02', nextCursor: 'next' }
})
const poolFilters = { type: ['replicated'] }
const allPools = await loadAllPages('/pools', 7, { filters: poolFilters })
assert.deepEqual(allPools.items.map((row) => row.name), ['first', 'second'])
assert.equal(allPools.stale, true)
assert.equal(allPools.staleReason, 'expired')
assert.equal(allPools.observedAt, '2026-01-01')
assert.deepEqual(pageCalls.map(({ path, cluster, options }) => [path, cluster, options.filters]), [
  ['/pools', 7, poolFilters], ['/pools', 7, poolFilters]
])
const failedPages = new Function('listResource', `${allPagesCode}; return listAllResources`)(async (_path, _cluster, options) => {
  if (options.cursor) throw new Error('second page unavailable')
  return { items: [{ name: 'first' }], stale: false, nextCursor: 'next' }
})
await assert.rejects(() => failedPages('/pools', 7), /second page unavailable/)
assert.ok(poolSource.includes('压缩后大小与原始大小的比例上限'), 'compression ratio is an upper bound, not a minimum ratio')
assert.ok(poolSource.includes('分配单元对齐和压缩头开销'), 'compression storage is also subject to native allocation constraints')
const poolTree = ts.createSourceFile('PoolManagementPage.tsx', poolSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const objectCountFn = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'poolObjectCount')
const objectCountCode = ts.transpileModule(objectCountFn.getText(poolTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const objectCount = new Function(`${objectCountCode}; return poolObjectCount`)()
assert.equal(objectCount(0), '0')
assert.equal(objectCount(12345), '12,345')
for (const value of [undefined, null, -1, 1.5, Number.MAX_SAFE_INTEGER + 1, '2']) assert.equal(objectCount(value), '未采集')
assert.ok(poolSource.includes('objects_display: poolObjectCount(row.objects)'))
const capacityFns = poolTree.statements.filter((node) => ts.isFunctionDeclaration(node) && ['poolCapacity', 'formatBytes'].includes(node.name.text))
const capacityCode = ts.transpileModule(capacityFns.map((fn) => fn.getText(poolTree).replace('export ', '')).join('\n'), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const capacity = new Function(`${capacityCode}; return poolCapacity`)()
assert.equal(capacity(0), '0 B')
assert.equal(capacity(1024), '1.0 KiB')
assert.equal(capacity(1099511627776), '1.0 TiB')
for (const value of [undefined, null, -1, NaN, Infinity, '1024']) assert.equal(capacity(value), '未采集')
for (const field of ['stored', 'bytes_used', 'max_avail']) {
  assert.ok(poolSource.includes(`${field}_display: poolCapacity(row.${field})`))
  assert.ok(poolSource.includes(`key: '${field}_display'`))
}
const usageDisplayFn = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'poolUsage')
const usageDisplayCode = ts.transpileModule(usageDisplayFn.getText(poolTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const poolUsageDisplay = new Function('numberValue', `${usageDisplayCode}; return poolUsage`)((value) => typeof value === 'number' ? value : undefined)
assert.equal(poolUsageDisplay({}), '未采集')
assert.equal(poolUsageDisplay({ used_percent: 0 }), '0.0%')
assert.equal(poolUsageDisplay({ used_percent: 25 }), '25%')
assert.ok(poolSource.includes("erasure_code_profile: textValue(row.erasure_code_profile, '')"), 'edit form must preserve missing profile data')
assert.ok(!poolSource.includes("const defaultErasureCodeProfile = 'default'"), 'do not invent an existing pool profile')
assert.ok(poolSource.includes("required: formMode === 'create', message: '请选择纠删码配置'"), 'immutable missing profile must not prevent unrelated edits')
const pgStatusFn = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'poolPGStatus')
const pgStatusCode = ts.transpileModule(pgStatusFn.getText(poolTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const pgStatus = new Function('isRecord', `${pgStatusCode}; return poolPGStatus`)((value) => value !== null && typeof value === 'object' && !Array.isArray(value))
assert.equal(pgStatus({ 'active+clean': 8, down: 2 }), '8 active+clean, 2 down')
assert.equal(pgStatus({ 'active+degraded': 3 }), '3 active+degraded')
const detailSource = readFileSync(new URL('../src/pages/cluster/PoolDetailPage.tsx', import.meta.url), 'utf8')
const detailTree = ts.createSourceFile('PoolDetailPage.tsx', detailSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const detailPageFn = detailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'PoolDetailPage')
const detailRefreshFn = detailPageFn.body.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'refreshPoolDetail')
const detailRefreshCode = ts.transpileModule(detailRefreshFn.getText(detailTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const changed of [false, true]) {
  let finish, reads = 0
  const scope = { current: { clusterId: 1, name: 'a' } }
  const refreshDetail = new Function('resourceScope', 'operationMutation', 'refresh', `const selectedClusterId=1, decodedName='a', refreshing=false, setRefreshing=()=>{}; ${detailRefreshCode}; return refreshPoolDetail`)(scope, { run: (_action, success) => { assert.equal(success, false); return new Promise((resolve) => { finish = resolve }) } }, () => { reads++ })
  const pending = refreshDetail()
  if (changed) scope.current = { clusterId: 2, name: 'b' }
  finish()
  await pending
  assert.equal(reads, changed ? 0 : 1, 'old collection completion must not invoke an obsolete detail loader')
}
assert.ok(detailSource.includes('pg_status_display: poolPGStatus(row.pg_status)'))
assert.ok(!detailSource.includes('active+clean'), 'detail view must not fabricate healthy PG states')
for (const field of ['stored', 'bytes_used', 'max_avail', 'compress_bytes_used', 'compress_under_bytes']) assert.ok(detailSource.includes(`poolCapacity(data.${field})`))
assert.ok(detailSource.includes('poolObjectCount(data.objects)'))
assert.ok(detailSource.includes('poolUsage(data)'))
for (const field of ['read_bytes', 'write_bytes']) {
  assert.ok(poolSource.includes(`${field}_display: poolCapacity(row.${field})`))
  assert.ok(detailSource.includes(`poolCapacity(data.${field})`))
}
for (const value of [undefined, null, {}, [], 'active+clean', { down: -1 }, { down: '2' }, { down: 1.5 }, { '': 1 }]) assert.equal(pgStatus(value), '未采集')
assert.ok(poolSource.includes('poolPGStatus(row.pg_status)'), 'pool health must use observed state counts')
assert.ok(!poolSource.includes('`${pgNum} active+clean'), 'PG count must not imply healthy PGs')
assert.ok(poolSource.includes("Form.useWatch('erasure_code_profile', form)"))
assert.ok(poolSource.includes('resourceName(row) === selectedProfileName'))
assert.ok(poolSource.includes('items={erasureProfileDetails(selectedProfile)}'))
assert.ok(poolSource.includes('配置详情尚未采集'))
assert.ok(poolSource.includes("Form.useWatch('crush_rule', form)"))
assert.ok(poolSource.includes('row.rule_name === selectedRuleName'))
assert.ok(poolSource.includes('<CrushRuleDetails row={selectedRule} />'))
assert.ok(poolSource.includes('规则详情尚未采集'))
const poolPage = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'PoolManagementPage')
const deleteGuardFn = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'poolDeleteBlocked')
const deleteGuardCode = ts.transpileModule(deleteGuardFn.getText(poolTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const deleteGuard = new Function('resourceName', `${deleteGuardCode}; return poolDeleteBlocked`)((row) => row.name ?? '')
assert.equal(deleteGuard({ name: 'a', stale: false, resource_version: 3 }), undefined)
assert.equal(deleteGuard({ name: 'a', stale: false, resource_version: 3, flags: ['hashpspool', 'nodelete'] }), '存储池已设置 nodelete 删除保护')
assert.equal(deleteGuard({ name: 'a', stale: false, resource_version: 3, flags: ['hashpspool'] }), undefined)
for (const row of [{ name: 'a' }, { name: 'a', stale: true, resource_version: 3 }, { name: 'a', stale: false }, { name: '', stale: false, resource_version: 3 }]) assert.ok(deleteGuard(row))
const deleteFn = poolPage.body.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'deletePool')
const deleteCode = ts.transpileModule(deleteFn.getText(poolTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const changed of [false, true]) {
  let modal, calls = []
  const scope = { current: { id: 7 } }
  const remove = new Function('Modal', 'clusterScope', 'poolDeleteBlocked', 'resourceName', 'operationMutation', 'mutateResource', 'refreshResource', 'refresh', `const selectedClusterId=7, loading=false, error=null; ${deleteCode}; return deletePool`)(
    { confirm: (options) => { modal = options } }, scope, deleteGuard, (row) => row.name,
    { run: (action) => action() }, (...args) => { calls.push(args) }, async () => {}, async () => {}
  )
  remove({ name: 'a', stale: false, resource_version: 3 })
  assert.ok(modal.content.includes('永久删除池内全部对象'))
  if (changed) {
    scope.current = { id: 8 }
    await assert.rejects(() => modal.onOk(), /集群已切换/)
    assert.equal(calls.length, 0)
  } else {
    await modal.onOk()
    assert.deepEqual(calls, [['/pool', 'DELETE', { cluster_id: 7, pool: 'a' }, { ifMatch: 3 }]])
  }
}
for (const mode of ['create', 'edit']) {
  const fn = poolPage.body.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'submitPool')
  const code = ts.transpileModule(fn.getText(poolTree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
  let finish, touched = false
  const scopeRef = { current: { id: 1 } }
  const run = new Function('clusterScope', 'operationMutation', 'setSubmitting', 'message', 'formMode', 'poolPlacementAvailable', 'poolUpdateBodies', `const selectedClusterId=1, submitting=false, loading=false, error=null, crushRuleOptions=[], erasureCodeProfileOptions=[], editingPool={}, data={}; ${code}; return submitPool`)(scopeRef, { run: () => new Promise((resolve) => { finish = resolve }) }, () => {}, { success: () => { touched = true } }, mode, () => true, () => [{}])
  const pending = run({})
  scopeRef.current = { id: 2 }
  finish()
  await pending
  assert.equal(touched, false, `${mode} pool completion must not update the new cluster UI`)
}
for (const name of ['submitCrushRule', 'submitErasureCodeProfile']) {
  const fn = poolPage.body.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === name)
  const code = ts.transpileModule(fn.getText(poolTree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
  let finish, wrote = false
  const scopeRef = { current: { id: 1 } }
  const run = new Function('clusterScope', 'operationMutation', 'setSubmittingCrushRule', 'setSubmittingErasureCodeProfile', 'setCreatedCrushRules', 'setCreatedErasureCodeProfiles', 'placementValid', 'erasureLocalityValid', `const selectedClusterId=1, submittingCrushRule=false, submittingErasureCodeProfile=false, loading=false, error=null, data={crushNodes:[]}; ${code}; return ${name}`)(scopeRef, { run: () => new Promise((resolve) => { finish = resolve }) }, () => {}, () => {}, () => { wrote = true }, () => { wrote = true }, () => true, () => true)
  const pendingWrite = run({})
  scopeRef.current = { id: 2 }
  finish()
  await pendingWrite
  assert.equal(wrote, false, `${name} must ignore previous cluster completion`)
}
assert.ok(!poolSource.includes('/usr/lib64/ceph/erasure-code'), 'do not invent a deployment-specific plugin directory')
assert.ok(!poolSource.includes('profile.directory'), 'do not copy an unrelated profile directory')
const profileBodyFns = poolTree.statements.filter((node) => ts.isFunctionDeclaration(node) && ['erasureCodeProfileBody', 'positiveInteger'].includes(node.name.text))
const bodyExports = {}
new Function('exports', ts.transpileModule(profileBodyFns.map((fn) => fn.getText(poolTree)).join('\n') + '\nexports.body = erasureCodeProfileBody', { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(bodyExports)
const profileValues = { name: 'ec', plugin: 'isa', k: 4, m: 2 }
for (const directory of [undefined, '', '  ']) assert.ok(!('directory' in JSON.parse(JSON.stringify(bodyExports.body({ ...profileValues, directory }, 1)))))
assert.equal(bodyExports.body({ ...profileValues, directory: ' /custom/plugins ' }, 1).directory, '/custom/plugins')
const poolFunctions = poolTree.statements.filter((node) => ts.isFunctionDeclaration(node) && ['topologyCounts', 'crushRootNames', 'placementDeviceOptions', 'placementValid', 'failureDomainOptions'].includes(node.name.text))
const poolCode = ts.transpileModule(poolFunctions.map((fn) => fn.getText(poolTree)).join('\n') + '\nexports.counts = topologyCounts; exports.roots = crushRootNames; exports.devices = placementDeviceOptions; exports.valid = placementValid; exports.domains = failureDomainOptions', { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
const poolExports = {}
new Function('exports', 'textValue', 'isRecord', poolCode)(poolExports, (value, fallback) => typeof value === 'string' ? value : fallback, (value) => value !== null && typeof value === 'object' && !Array.isArray(value))
const osds = [
  { id: -1, name: 'default', type: 'root', children: [-2, -3] },
  { id: -2, name: 'a', type: 'host', children: [0] },
  { id: -3, name: 'b', type: 'host', children: [1] },
  { id: -4, name: 'archive', type: 'root', children: [2] },
  { id: -5, name: 'empty', type: 'root', children: [] },
  { id: 0, name: 'osd.0', type: 'osd', device_class: 'ssd' },
  { id: 1, name: 'osd.1', type: 'osd', device_class: 'hdd' },
  { id: 2, name: 'osd.2', type: 'osd', device_class: 'ssd' },
  { id: 3, name: 'osd.3', type: 'osd', device_class: 'ssd' }
]
assert.equal(poolExports.counts(osds, 'default').osd, 2)
assert.equal(poolExports.counts(osds, 'default').host, 2)
assert.equal(poolExports.counts(osds, 'default', 'ssd').osd, 1)
assert.equal(poolExports.counts(osds, 'archive').osd, 1)
assert.equal(poolExports.counts(osds, 'a').osd, 1)
assert.equal(poolExports.counts(osds, 'missing').osd, 0)
assert.equal(poolExports.counts([], 'default').osd, 0)
assert.equal(poolExports.counts(osds, 'default', 'ssd').host, 1)
assert.equal(poolExports.counts(osds, 'empty').osd, 0)
assert.ok(poolExports.roots(osds).includes('empty'))
assert.deepEqual(poolExports.roots([]), [])
assert.throws(() => poolExports.counts([{ id: -1, name: 'bad', type: 'root', children: [-1] }], 'bad'))
assert.deepEqual(poolExports.devices(osds, 'archive').map((option) => option.value), ['', 'ssd'])
assert.deepEqual(poolExports.devices(osds, 'empty').map((option) => option.value), [''])
assert.equal(poolExports.valid(osds, 'default', 'host', 'ssd'), true)
assert.equal(poolExports.valid(osds, 'archive', 'osd', 'hdd'), false)
assert.equal(poolExports.valid(osds, 'empty', 'osd'), false)
assert.equal(poolExports.valid(osds, 'default', 'rack'), false)
assert.equal(poolExports.valid(osds, 'default', 'root'), false)
const localityFunction = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'erasureLocalityValid')
const localityExports = {}
new Function('exports', 'placementValid', ts.transpileModule(localityFunction.getText(poolTree) + '\nexports.valid = erasureLocalityValid', { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(localityExports, poolExports.valid)
assert.equal(localityExports.valid({ plugin: 'lrc', crush_root: 'archive', crush_locality: 'host' }, osds), false)
assert.equal(localityExports.valid({ plugin: 'lrc', crush_root: 'archive', crush_locality: 'osd' }, osds), true)
assert.equal(localityExports.valid({ plugin: 'lrc', crush_root: 'archive', crush_locality: undefined }, osds), true)
assert.ok(!poolSource.includes("l: 3, crush_locality: 'host'"), 'plugin changes must preserve topology selections')
assert.deepEqual(poolExports.domains({ osd: 0, host: 0, root: 1 }), [])
console.log('CRUSH failure domain root isolation checks passed')

const placementFunction = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'poolPlacementAvailable')
const placementCode = ts.transpileModule(placementFunction.getText(poolTree) + '\nexports.valid = poolPlacementAvailable', { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
const placement = {}
new Function('exports', placementCode)(placement)
assert.equal(placement.valid({ pool_type: 'replicated', crush_rule: 'replicated_rule' }, [], []), false)
assert.equal(placement.valid({ pool_type: 'erasure', erasure_code_profile: 'default' }, [], []), false)
assert.equal(placement.valid({ pool_type: 'replicated', crush_rule: 'observed' }, [{ value: 'observed' }], []), true)
assert.equal(placement.valid({ pool_type: 'erasure', erasure_code_profile: 'created' }, [], [{ value: 'created' }]), true)
assert.equal(placement.valid({ pool_type: 'replicated', crush_rule: '' }, [{ value: '' }], []), false)
assert.ok(!poolSource.includes(".catch(() => [])"), 'placement load errors must not become empty success')

const ruleFunction = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'readableCrushRule')
const ruleCode = ts.transpileModule(ruleFunction.getText(poolTree) + '\nexports.name = readableCrushRule', { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
const mapping = {}
new Function('exports', 'textValue', ruleCode)(mapping, (value, fallback) => typeof value === 'string' ? value : fallback)
const sparseRules = [{ rule_id: 8, rule_name: 'ssd' }, { rule_id: 0, rule_name: 'hdd' }]
assert.equal(mapping.name(0, sparseRules), 'hdd')
assert.equal(mapping.name('8', sparseRules), 'ssd')
assert.equal(mapping.name(1, sparseRules), '1')
assert.equal(mapping.name(null, sparseRules), '')
assert.equal(mapping.name('8', [...sparseRules, { rule_id: 9, rule_name: '8' }]), '8')
assert.equal(mapping.name('ssd', sparseRules), 'ssd')

const freshnessFunction = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'poolPlacementRows')
const freshnessCode = ts.transpileModule(freshnessFunction.getText(poolTree) + '\nexports.rows = poolPlacementRows', { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
const freshness = {}
new Function('exports', freshnessCode)(freshness)
assert.deepEqual(freshness.rows({ stale: false, items: [] }, 'CRUSH 规则'), [])
assert.throws(() => freshness.rows({ stale: true, items: [] }, 'CRUSH 规则'))
assert.throws(() => freshness.rows({ stale: false, items: [{ stale: true }] }, 'CRUSH 规则'))
assert.throws(() => freshness.rows({ stale: false, items: [{}] }, 'CRUSH 规则'))
assert.throws(() => freshness.rows({ stale: false, items: [{ stale: false, name: 'partial' }] }, '纠删码配置'))
assert.equal(freshness.rows({ stale: false, items: [{ stale: false, plugin: 'isa' }] }, '纠删码配置').length, 1)

const rulesSource = readFileSync(new URL('../src/pages/cluster/CrushRulesPanel.tsx', import.meta.url), 'utf8')
const rulesTree = ts.createSourceFile('CrushRulesPanel.tsx', rulesSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const rulesFunctions = rulesTree.statements.filter((node) => ts.isFunctionDeclaration(node) && ['crushRuleType', 'crushRuleSteps', 'crushRuleDeleteBlocked'].includes(node.name.text))
const rules = {}
new Function('exports', ts.transpileModule(rulesFunctions.map((fn) => fn.getText(rulesTree)).join('\n'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(rules)
assert.equal(rules.crushRuleType(1), '复制 (1)')
assert.equal(rules.crushRuleType(3), '纠删码 (3)')
assert.equal(rules.crushRuleType(5), '类型 5')
assert.equal(rules.crushRuleType(null), '未知')
const steps = [{ op: 'take', item: -1, item_name: 'default' }, { op: 'chooseleaf_firstn', num: 0, type: 'host' }, { op: 'emit' }]
const rendered = rules.crushRuleSteps(steps)
assert.deepEqual(rendered.map((step) => step.sequence), [1, 2, 3])
assert.equal(rendered[0].item, -1)
assert.equal(rendered[1].num, 0)
assert.equal(steps[0].sequence, undefined)
assert.deepEqual(rules.crushRuleSteps([]), [])
for (const invalid of [null, {}, [null], [{ op: 3 }]]) assert.equal(rules.crushRuleSteps(invalid), null)
console.log('CRUSH rule type and native step checks passed')
assert.equal(rules.crushRuleDeleteBlocked({ rule_name: 'a', stale: false }), undefined)
for (const row of [{ rule_name: 'a', stale: true }, { rule_name: 'a' }, { stale: false }, { rule_name: '', stale: false }]) assert.ok(rules.crushRuleDeleteBlocked(row))

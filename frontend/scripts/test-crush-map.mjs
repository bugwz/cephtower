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
const poolTree = ts.createSourceFile('PoolManagementPage.tsx', poolSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
assert.ok(poolSource.includes("Form.useWatch('erasure_code_profile', form)"))
assert.ok(poolSource.includes('resourceName(row) === selectedProfileName'))
assert.ok(poolSource.includes('items={erasureProfileDetails(selectedProfile)}'))
assert.ok(poolSource.includes('配置详情尚未采集'))
const poolPage = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'PoolManagementPage')
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

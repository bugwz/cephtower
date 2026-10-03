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
const techniqueFunctions = poolTree.statements.filter((node) => ts.isFunctionDeclaration(node) && ['erasureCodeTechniqueOptions', 'clayTechniqueForScalar'].includes(node.name.text))
const techniqueCode = ts.transpileModule(techniqueFunctions.map((node) => node.getText(poolTree)).join('\n'), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const techniques = new Function(`${techniqueCode}; return { options: erasureCodeTechniqueOptions, change: clayTechniqueForScalar }`)()
for (const [scalar, expected] of Object.entries({ jerasure: ['reed_sol_van', 'reed_sol_r6_op', 'cauchy_orig', 'cauchy_good', 'liber8tion'], isa: ['reed_sol_van', 'cauchy'], shec: ['single', 'multiple'] })) {
  assert.deepEqual(techniques.options('clay', scalar).map((option) => option.value), expected)
  for (const value of expected) assert.equal(techniques.change(scalar, value), value)
  assert.equal(techniques.change(scalar, 'invalid'), expected[0])
}
assert.equal(techniques.change('shec', 'reed_sol_van'), 'single')
assert.equal(techniques.change('isa', 'multiple'), 'reed_sol_van')
assert.deepEqual(techniques.options('clay', undefined), [])
assert.ok(techniques.options('jerasure').some((option) => option.value === 'liberation'))
let techniqueRules
function findTechniqueRules(node) {
  if (ts.isJsxOpeningElement(node) && node.tagName.getText(poolTree) === 'Form.Item' && node.attributes.properties.some((attribute) => attribute.name?.getText(poolTree) === 'name' && attribute.initializer?.text === 'technique')) {
    techniqueRules = node.attributes.properties.find((attribute) => attribute.name?.getText(poolTree) === 'rules').initializer.expression
  }
  ts.forEachChild(node, findTechniqueRules)
}
findTechniqueRules(poolTree)
const techniqueRulesCode = ts.transpileModule(`const rules = ${techniqueRules.getText(poolTree)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const techniqueRule = new Function('erasureCodeTechniqueOptions', `${techniqueRulesCode}; return rules[1]`)(techniques.options)
for (const scalar of ['jerasure', 'isa', 'shec']) {
  const validator = techniqueRule({ getFieldValue: (name) => name === 'plugin' ? 'clay' : scalar }).validator
  await validator(null, techniques.options('clay', scalar)[0].value)
  for (const value of ['liberation', 'blaum_roth', '', undefined]) await assert.rejects(validator(null, value))
}
assert.ok(poolSource.includes("technique: clayTechniqueForScalar(value, erasureCodeProfileForm.getFieldValue('technique'))"))
console.log('CLAY scalar plugin technique options, transitions and form validation checks passed')
const lrcRuleNode = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'lrcGroupingRule')
const lrcRuleCode = ts.transpileModule(lrcRuleNode.getText(poolTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const lrcRule = new Function(`${lrcRuleCode}; return lrcGroupingRule`)()
const validateLRC = (k, m, l, plugin = 'lrc') => lrcRule({ getFieldValue: (name) => ({ k, m, l, plugin })[name] }).validator()
for (const values of [[4, 2, 3], [4, 2, 6], [8, 4, 3], [6, 3, 3]]) await validateLRC(...values)
for (const values of [[4, 2, 4], [4, 2, 2], [4, 2, 7], [4, 2, 0], [4, 2, undefined], [4, 2, 1.5], [4, 2, '3'], [1, 2, 3], [4, 0, 4], [NaN, 2, 3], [Number.MAX_SAFE_INTEGER, 2, 3]]) await assert.rejects(validateLRC(...values))
await validateLRC(4, 2, undefined, 'isa')
assert.equal((poolSource.match(/ecChunkRule\(erasureCodeTopology\),\s+lrcGroupingRule/g) ?? []).length, 2)
assert.ok(poolSource.includes("dependencies={['k', 'm', 'plugin']} rules={[{ required: true, min: 1, type: 'number' }, lrcGroupingRule]}"))
console.log('LRC native grouping divisibility and plugin isolation checks passed')
const layoutNode = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'lrcLayoutPreview')
const layoutCode = ts.transpileModule(layoutNode.getText(poolTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const lrcPreview = new Function(`${layoutCode}; return lrcLayoutPreview`)()
assert.deepEqual(lrcPreview(4, 2, 3), { groups: 2, dataPerGroup: 2, codingPerGroup: 1, totalChunks: 8 })
assert.deepEqual(lrcPreview(4, 2, 6), { groups: 1, dataPerGroup: 4, codingPerGroup: 2, totalChunks: 7 })
assert.deepEqual(lrcPreview(8, 4, 3), { groups: 4, dataPerGroup: 2, codingPerGroup: 1, totalChunks: 16 })
for (const values of [[4, 2, 4], [4, 2, 2], [4, 2, undefined], [4, 2, 0], [4, 2, 1.5], [4, 2, '3'], [NaN, 2, 3], [Number.MAX_SAFE_INTEGER - 1, 1, Number.MAX_SAFE_INTEGER]]) assert.equal(lrcPreview(...values), null)
for (const field of ['k', 'm', 'l']) assert.ok(poolSource.includes(`Form.useWatch('${field}', erasureCodeProfileForm)`))
assert.ok(poolSource.includes('LRC 分组预览（表单计算）'))
console.log('LRC layout preview preserves native extra parity groups and rejects invalid input')
const shecRuleNode = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'shecParameterRule')
const shecRuleCode = ts.transpileModule(shecRuleNode.getText(poolTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const shecRule = new Function(`${shecRuleCode}; return shecParameterRule`)()
const validateSHEC = (k, m, c, plugin = 'shec') => shecRule({ getFieldValue: (name) => ({ k, m, c, plugin })[name] }).validator()
for (const values of [[4, 3, 2], [12, 8, 8], [10, 10, 1], [2, 1, 1]]) await validateSHEC(...values)
for (const values of [[4, 5, 2], [13, 1, 1], [12, 9, 1], [4, 3, 4], [4, 3, 0], [4, 3, undefined], [4, 3, 1.5], [4, 3, '2'], [NaN, 3, 1], [4, -1, 1]]) await assert.rejects(validateSHEC(...values))
await validateSHEC(16, 4, undefined, 'isa')
assert.equal((poolSource.match(/lrcGroupingRule,\s+shecParameterRule/g) ?? []).length, 2)
assert.ok(poolSource.includes("dependencies={['k', 'm', 'plugin']} rules={[shecParameterRule]}"))
console.log('SHEC native parameter boundaries and plugin isolation checks passed')
const domainRuleNode = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'ecFailureDomainCountRule')
const osdsRuleNode = poolTree.statements.find((node) => ts.isVariableStatement(node) && node.declarationList.declarations[0].name.getText(poolTree) === 'ecOSDsPerFailureDomainRule')
const msrCode = ts.transpileModule(domainRuleNode.getText(poolTree) + '\n' + osdsRuleNode.getText(poolTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const msrRules = new Function('numberValue', 'textValue', `${msrCode}; return { domains: ecFailureDomainCountRule, osds: ecOSDsPerFailureDomainRule }`)((value) => typeof value === 'number' ? value : undefined, (value, fallback) => value ?? fallback)
for (const perDomain of [undefined, 0, 1]) {
  const validator = msrRules.domains({ host: 3 })({ getFieldValue: (name) => name === 'crush_osds_per_failure_domain' ? perDomain : 'host' }).validator
  await validator(null, 0)
  await validator(null, 3)
  await msrRules.osds().validator(null, perDomain)
}
const msrDomainValidator = msrRules.domains({ host: 3 })({ getFieldValue: (name) => name === 'crush_osds_per_failure_domain' ? 2 : 'host' }).validator
for (const value of [undefined, 0, -1, 4]) await assert.rejects(msrDomainValidator(null, value))
await msrDomainValidator(null, 3)
await assert.rejects(msrRules.osds().validator(null, -1))
console.log('Native simple-rule and MSR parameter activation checks passed')
const adjustmentFn = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'poolPGAdjustment')
const adjustmentCode = ts.transpileModule(adjustmentFn.getText(poolTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const adjustment = new Function(`${adjustmentCode}; return poolPGAdjustment`)()
const settled = { pg_num: 32, pgp_num: 32, pg_num_target: 32, pgp_num_target: 32 }
assert.equal(adjustment(settled), '已达到目标数量')
assert.equal(adjustment({ ...settled, pg_num_target: 64 }), '调整中')
assert.equal(adjustment({ ...settled, pgp_num_target: 16 }), '调整中')
assert.equal(adjustment({ pg_num: 16, pgp_num: 32, pg_num_target: 32, pgp_num_target: 16 }), '调整中', 'opposing differences must not cancel each other')
for (const field of Object.keys(settled)) for (const value of [undefined, null, 0, -1, 1.5, '32']) assert.equal(adjustment({ ...settled, [field]: value }), '未采集')
assert.ok(poolSource.includes('pg_adjustment_display: poolPGAdjustment(row)'))
const listFreshnessFn = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'poolListFreshnessWarning')
const listFreshnessCode = ts.transpileModule(listFreshnessFn.getText(poolTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const listFreshness = new Function(`${listFreshnessCode}; return poolListFreshnessWarning`)()
assert.equal(listFreshness({ stale: false, pools: [{ stale: false }] }), undefined)
assert.equal(listFreshness({ stale: false, pools: [] }), undefined)
assert.match(listFreshness({ stale: true, pools: [] }), /历史采集/)
assert.match(listFreshness({ stale: false, pools: [{ stale: false }, { stale: true }] }), /历史采集/)
for (const stale of [undefined, null, 'false', 0]) {
  assert.match(listFreshness({ stale, pools: [] }), /时效未知/)
  assert.match(listFreshness({ stale: false, pools: [{ stale }] }), /时效未知/)
}
assert.ok(poolSource.includes('message={poolListFreshnessWarning(data)}'))
const kindFn = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'poolKind')
const kindCode = ts.transpileModule(kindFn.getText(poolTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const kind = new Function(`${kindCode}; return poolKind`)()
for (const type of ['replicated', 'erasure']) assert.equal(kind({ type }), type)
for (const type of [undefined, null, '', 'unknown', 1, 3, 'other']) assert.equal(kind({ type }), undefined)
assert.ok(poolSource.includes('Boolean(poolEditBlocked(row))'))
const editBlockedFn = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'poolEditBlocked')
const editBlockedCode = ts.transpileModule(editBlockedFn.getText(poolTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const editBlocked = new Function('poolKind', 'resourceName', `${editBlockedCode}; return poolEditBlocked`)(kind, (row) => row.name ?? '')
const editable = { type: 'replicated', name: 'pool', stale: false, resource_version: 2 }
assert.equal(editBlocked(editable), undefined)
assert.equal(editBlocked({ ...editable, flags: ['nodelete'] }), undefined, 'deletion protection must not disable editing')
for (const stale of [true, undefined, null]) assert.match(editBlocked({ ...editable, stale }), /库存/)
for (const resource_version of [undefined, null, 0, -1, 1.5, Infinity]) assert.match(editBlocked({ ...editable, resource_version }), /版本/)
assert.match(editBlocked({ ...editable, type: 'unknown' }), /类型/)
assert.match(editBlocked({ ...editable, name: '' }), /名称/)
assert.ok(poolSource.includes('editingPool ? poolEditBlocked(editingPool)'))
assert.ok(poolSource.includes("if (!kind) throw new Error('池类型未采集或不受支持，无法编辑')"))
const initialFn = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'poolInitialValues')
const initialCode = ts.transpileModule(initialFn.getText(poolTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const initial = new Function('poolKind', `${initialCode}; return poolInitialValues`)(kind)
const pgUpdateCode = ts.transpileModule(poolTree.statements.filter((node) => ts.isFunctionDeclaration(node) && ['poolAutoscaleMode', 'poolPGUpdates'].includes(node.name.text)).map((node) => node.getText(poolTree)).join('\n'), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const pgHelpers = new Function(`${pgUpdateCode}; return { poolPGUpdates, poolAutoscaleMode }`)()
for (const mode of [undefined, null, '', 'invalid', 1]) assert.equal(pgHelpers.poolAutoscaleMode({ pg_autoscale_mode: mode }), undefined)
for (const mode of ['on', 'off', 'warn']) assert.equal(pgHelpers.poolAutoscaleMode({ pg_autoscale_mode: mode }), mode)
assert.deepEqual(pgHelpers.poolPGUpdates({}, {}, 7, 'p'), [])
assert.deepEqual(pgHelpers.poolPGUpdates({}, { pg_autoscale_mode: 'off' }, 7, 'p'), [{ cluster_id: 7, pool: 'p', field: 'pg_autoscale_mode', value: 'off' }])
assert.deepEqual(pgHelpers.poolPGUpdates({ pg_autoscale_mode: 'warn', pg_num: 64 }, { pg_autoscale_mode: 'warn', pg_num: 64 }, 7, 'p'), [])
assert.deepEqual(pgHelpers.poolPGUpdates({ pg_autoscale_mode: 'off' }, { pg_autoscale_mode: 'off', pg_num: 64 }, 7, 'p'), [{ cluster_id: 7, pool: 'p', field: 'pg_num', value: '64' }])
for (const pg_num of [undefined, null]) assert.deepEqual(pgHelpers.poolPGUpdates({ pg_autoscale_mode: 'off', pg_num: 64 }, { pg_autoscale_mode: 'off', pg_num }, 7, 'p'), [])
assert.deepEqual(pgHelpers.poolPGUpdates({ pg_autoscale_mode: 'on' }, { pg_autoscale_mode: 'on', pg_num: 64 }, 7, 'p'), [])
for (const pg_num of [0, -1, 1.5, '64', NaN, Infinity, Number.MAX_SAFE_INTEGER + 1]) assert.throws(() => pgHelpers.poolPGUpdates({}, { pg_autoscale_mode: 'off', pg_num }, 7, 'p'), /PG 数量/)
assert.throws(() => pgHelpers.poolPGUpdates({}, { pg_autoscale_mode: 'invalid' }, 7, 'p'), /自动伸缩模式/)
assert.ok(!initialCode.includes('pg_num: numberValue(row.pg_num) ?? 32'))
assert.ok(poolSource.includes('requests.push(...poolPGUpdates(row, values, clusterId, pool))'))
const quotaUpdateCode = ts.transpileModule(poolTree.statements.filter((node) => ts.isFunctionDeclaration(node) && ['observedPoolQuota', 'poolQuotaUpdates'].includes(node.name.text)).map((node) => node.getText(poolTree)).join('\n'), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const quotaUpdate = new Function('quotaUnits', `${quotaUpdateCode}; return poolQuotaUpdates`)(['B', 'KiB', 'MiB', 'GiB', 'TiB'])
const quotaValues = { quota_unit: 'GiB' }
for (const unknown of [undefined, null, -1, 1.5, '0', NaN, Number.MAX_SAFE_INTEGER + 1]) {
  const row = { quota_max_bytes: unknown, quota_max_objects: unknown, max_bytes: 100, max_objects: 100 }
  assert.deepEqual(quotaUpdate(row, quotaValues, 7, 'p'), [])
  assert.deepEqual(quotaUpdate(row, { ...quotaValues, quota_max_objects: 0 }, 7, 'p'), [{ cluster_id: 7, pool: 'p', operation: 'quota', field: 'max_objects', value: '0' }])
}
assert.deepEqual(quotaUpdate({ quota_max_bytes: 1024 ** 3, quota_max_objects: 12 }, { ...quotaValues, quota_max_bytes: 1, quota_max_objects: 12 }, 7, 'p'), [])
assert.deepEqual(quotaUpdate({ quota_max_bytes: 1024, quota_max_objects: 12 }, { ...quotaValues, quota_max_bytes: null, quota_max_objects: null }, 7, 'p').map(({ field, value }) => ({ field, value })), [{ field: 'max_bytes', value: '0' }, { field: 'max_objects', value: '0' }])
assert.deepEqual(quotaUpdate({}, { ...quotaValues, quota_max_bytes: 2 }, 7, 'p'), [{ cluster_id: 7, pool: 'p', operation: 'quota', field: 'max_bytes', value: String(2 * 1024 ** 3), quota_unit: 'GiB' }])
for (const invalid of [-1, 0.5, NaN, Infinity, Number.MAX_SAFE_INTEGER + 1]) assert.throws(() => quotaUpdate({}, { ...quotaValues, quota_max_objects: invalid }, 7, 'p'), /配额/)
assert.throws(() => quotaUpdate({}, { ...quotaValues, quota_max_bytes: Number.MAX_SAFE_INTEGER }, 7, 'p'), /配额/)
assert.throws(() => quotaUpdate({}, { quota_unit: 'invalid', quota_max_bytes: 0 }, 7, 'p'), /配额/)
assert.ok(initialCode.includes('observedPoolQuota(row.quota_max_bytes)'))
assert.ok(initialCode.includes('quota_max_bytes: quota.value'))
assert.ok(initialCode.includes('quota_max_objects: observedPoolQuota(row.quota_max_objects)'))
assert.ok(poolSource.includes('requests.push(...poolQuotaUpdates(row, values, clusterId, pool))'))
for (const type of [undefined, null, 'unknown', 'other']) assert.throws(() => initial({ type }), /池类型未采集或不受支持/)
const protectionFn = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'poolDataProtection')
const protectionCode = ts.transpileModule(protectionFn.getText(poolTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const protection = new Function(`${protectionCode}; return poolDataProtection`)()
assert.equal(protection({ type: 'replicated', size: 2 }), 'replica: x2')
assert.equal(protection({ type: 'erasure', size: 6 }), '纠删码（分片数未采集）')
const erasurePool = { type: 'erasure', erasure_code_profile: 'archive' }
const erasureProfile = { name: 'archive', k: '4', m: '2', stale: false }
assert.equal(protection(erasurePool, [erasureProfile]), 'EC: 4+2')
assert.equal(protection(erasurePool, [{ ...erasureProfile, k: 6, m: 3 }]), 'EC: 6+3')
for (const profiles of [[], [{ ...erasureProfile, name: 'other' }], [erasureProfile, erasureProfile], [{ ...erasureProfile, stale: true }], [{ ...erasureProfile, stale: undefined }]]) assert.equal(protection(erasurePool, profiles), '纠删码（分片数未采集）')
for (const field of ['k', 'm']) for (const value of [undefined, null, true, 0, -1, 1.5, '1.5', '4oops', Number.MAX_SAFE_INTEGER + 1]) assert.equal(protection(erasurePool, [{ ...erasureProfile, [field]: value }]), '纠删码（分片数未采集）')
assert.ok(poolSource.includes('normalizePoolRow(row, erasureCodeProfileRows)'))
assert.equal(protection({}), '未采集')
for (const size of [undefined, null, 0, -1, 1.5, '3']) assert.equal(protection({ type: 'replicated', size }), '副本数未采集')
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
const ioRateFn = poolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'poolIORate')
const ioRateCode = ts.transpileModule(ioRateFn.getText(poolTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const ioRate = new Function('isRecord', 'poolCapacity', `${ioRateCode}; return poolIORate`)((value) => value !== null && typeof value === 'object' && !Array.isArray(value), capacity)
assert.equal(ioRate({ read_bytes_sec: 1024 }, 'read_bytes_sec'), '1.0 KiB/s')
assert.equal(ioRate({ write_bytes_sec: 0 }, 'write_bytes_sec'), '0 B/s')
assert.equal(ioRate({ read_op_per_sec: 2 }, 'read_op_per_sec'), '2 op/s')
assert.equal(ioRate({ write_op_per_sec: 0 }, 'write_op_per_sec'), '0 op/s')
for (const value of [undefined, null, -1, 1.5, '4', Infinity]) assert.equal(ioRate({ read_bytes_sec: value }, 'read_bytes_sec'), '未采集')
assert.equal(ioRate({}, 'read_bytes_sec'), '未采集')
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
const historySource = readFileSync(new URL('../src/pages/cluster/PoolIOHistory.tsx', import.meta.url), 'utf8')
const historyTree = ts.createSourceFile('PoolIOHistory.tsx', historySource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const historyFns = historyTree.statements.filter((node) => ts.isFunctionDeclaration(node) && ['poolHistoryPoints', 'poolHistoryPath'].includes(node.name.text))
const historyCode = ts.transpileModule(historyFns.map((fn) => fn.getText(historyTree).replace('export ', '')).join('\n'), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const history = new Function('isRecord', `${historyCode}; return { points: poolHistoryPoints, path: poolHistoryPath }`)((value) => value !== null && typeof value === 'object' && !Array.isArray(value))
const response = { result_type: 'matrix', series: [{ metric: { pool_id: '7' }, values: [[1030, '1.5'], [1000, '0'], [1060, 'NaN'], [1090, '2']] }, { metric: { pool_id: '8' }, values: [[1000, '999']] }] }
const points = history.points(response, 7)
assert.deepEqual(points, [{ time: 1000000, value: 0 }, { time: 1030000, value: 1.5 }, { time: 1090000, value: 2 }])
assert.deepEqual(history.points(response, 9), [])
assert.equal((history.path(points).match(/M/g) ?? []).length, 2, 'missing samples must break the chart line')
assert.throws(() => history.points({ result_type: 'vector', series: [] }, 7), /时间序列/)
assert.throws(() => history.points({ ...response, series: [response.series[0], response.series[0]] }, 7), /重复历史序列/)
assert.throws(() => history.points({ result_type: 'matrix', series: [{ metric: { pool_id: '7' }, values: [[1, '2'], [1, '3']] }] }, 7), /重复时间戳/)
assert.ok(historySource.includes('controller.current?.abort()'))
assert.ok(detailSource.includes('data?.history_scope === `${selectedClusterId}/${decodedName}`'))
const detailTree = ts.createSourceFile('PoolDetailPage.tsx', detailSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const detailLoaderCode = ts.transpileModule(detailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'loadPoolDetail').getText(detailTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const scenario of ['fresh', 'stale', 'missing', 'failure', 'invalid', 'wrong-name', 'replicated', 'no-profile', 'no-pool', 'pool-failure']) {
  const calls = []
  const load = new Function('getOptionalResource', 'resourceToRecord', 'poolKind', 'poolDataProtection', 'normalizePoolDetail', `${detailLoaderCode}; return loadPoolDetail`)(async (path, clusterId, body) => {
    calls.push({ path, clusterId, body })
    if (path === '/pool') {
      if (scenario === 'pool-failure') throw new Error('pool read failed')
      if (scenario === 'no-pool') return undefined
      return { item: { name: 'p', type: scenario === 'replicated' ? 'replicated' : 'erasure', size: 3, erasure_code_profile: scenario === 'no-profile' ? undefined : 'ec' } }
    }
    if (scenario === 'failure') throw new Error('profile read failed')
    if (scenario === 'missing') return undefined
    return { item: { name: scenario === 'wrong-name' ? 'other' : 'ec', stale: scenario === 'stale', k: scenario === 'invalid' ? 0 : '4', m: '2' } }
  }, (item) => item, (row) => row.type, protection, (row, profiles) => ({ ...row, data_protection_display: protection(row, profiles) }))
  if (scenario === 'pool-failure') { await assert.rejects(load(7, 'p'), /pool read failed/); continue }
  const result = await load(7, 'p')
  assert.deepEqual(calls[0], { path: '/pool', clusterId: 7, body: { pool: 'p' } })
  if (scenario === 'no-pool') { assert.equal(result, null); assert.equal(calls.length, 1); continue }
  assert.equal(result.history_scope, '7/p')
  if (scenario === 'fresh' || scenario === 'replicated') {
    assert.equal(result.data_protection_display, scenario === 'fresh' ? 'EC: 4+2' : 'replica: x3')
    assert.equal(result.profile_warning, undefined)
  } else {
    assert.equal(result.data_protection_display, '纠删码（分片数未采集）')
    assert.match(result.profile_warning, /无法确认 k\+m/)
  }
  if (scenario === 'replicated' || scenario === 'no-profile') assert.equal(calls.length, 1)
  else assert.deepEqual(calls[1], { path: '/erasure/code/profile', clusterId: 7, body: { name: 'ec' } })
}
assert.ok(detailSource.includes('data_protection_display: poolDataProtection(row, profiles)'))
const minimumSizeFn = detailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'poolMinimumSize')
const minimumSizeCode = ts.transpileModule(minimumSizeFn.getText(detailTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const minimumSize = new Function(`${minimumSizeCode}; return poolMinimumSize`)()
for (const value of [1, 2, 5]) assert.equal(minimumSize(value), String(value))
for (const value of [undefined, null, 0, -1, 1.5, '2', Infinity]) assert.equal(minimumSize(value), '未采集')
assert.ok(detailSource.includes('poolMinimumSize(data.min_size)'))
assert.ok(detailSource.includes('poolPGAdjustment(data)'))
for (const field of ['read_bytes_sec', 'write_bytes_sec', 'read_op_per_sec', 'write_op_per_sec']) {
  assert.ok(poolSource.includes(`poolIORate(row.client_io_rate, '${field}')`))
  assert.ok(detailSource.includes(`poolIORate(data.client_io_rate, '${field}')`))
}
for (const name of ['autoscaleNumber', 'autoscaleBoolean']) {
  const fn = detailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === name)
  const code = ts.transpileModule(fn.getText(detailTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
  const format = new Function(`${code}; return ${name}`)()
  if (name === 'autoscaleNumber') {
    for (const value of [0, 0.25, 1.5]) assert.equal(format(value), String(value))
    for (const value of [null, undefined, '1', -1, NaN, Infinity]) assert.equal(format(value), '未采集')
  } else {
    assert.equal(format(false), '否')
    assert.equal(format(true), '是')
    for (const value of [null, undefined, 'false', 0]) assert.equal(format(value), '未采集')
  }
}
assert.ok(detailSource.includes('renderAutoscaleStatus(data?.autoscale_status)'))
assert.ok(detailSource.includes('poolCapacity(value.logical_used)'))
for (const field of ['raw_used_rate', 'actual_capacity_ratio', 'capacity_ratio']) assert.ok(detailSource.includes(`autoscaleNumber(value.${field})`))
for (const field of ['pg_num_final', 'would_adjust', 'target_bytes', 'subtree_capacity', 'target_ratio', 'effective_target_ratio', 'bias', 'bulk']) assert.ok(detailSource.includes(`value.${field}`))
assert.ok(detailSource.includes("const type = poolKind(row) ?? '未知'"))
const compressionRatioFn = detailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'formatCompressionRatio')
const compressionRatioCode = ts.transpileModule(compressionRatioFn.getText(detailTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const compressionRatio = new Function(`${compressionRatioCode}; return formatCompressionRatio`)()
for (const value of [0, 0.875, 1]) assert.equal(compressionRatio(value), String(value))
for (const value of [undefined, null, '', '0.875', -1, 2, NaN, Infinity]) assert.equal(compressionRatio(value), '未采集')
assert.ok(detailSource.includes('formatCompressionRatio(data.compression_required_ratio)'))
assert.ok(detailSource.includes('所需压缩比（配置阈值）'))
assert.ok(detailSource.includes("textValue(data.compression_algorithm, '未采集')"))
for (const field of ['compression_min_blob_size', 'compression_max_blob_size']) assert.ok(detailSource.includes(`poolCapacity(data.${field})`))
const quotaFn = detailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'formatQuota')
const quotaCode = ts.transpileModule(quotaFn.getText(detailTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const quota = new Function(`${quotaCode}; return formatQuota`)()
assert.equal(quota(0), '无限制（0）')
assert.equal(quota(1024), '1,024')
for (const value of [undefined, null, '', '0', -1, 1.5, Infinity, Number.MAX_SAFE_INTEGER + 1]) assert.equal(quota(value), '未采集')
for (const field of ['quota_max_bytes', 'quota_max_objects']) assert.ok(detailSource.includes(`formatQuota(data.${field})`))
assert.ok(detailSource.includes("textValue(data.compression_mode, '未采集')"))
const detailFreshnessFn = detailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'poolDetailFreshnessWarning')
const detailFreshnessCode = ts.transpileModule(detailFreshnessFn.getText(detailTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const detailFreshness = new Function(`${detailFreshnessCode}; return poolDetailFreshnessWarning`)()
assert.equal(detailFreshness({ stale: false }), undefined)
assert.match(detailFreshness({ stale: true }), /历史采集/)
for (const stale of [undefined, null, 'false', 0]) assert.match(detailFreshness({ stale }), /时效未知/)
assert.ok(detailSource.includes('message={poolDetailFreshnessWarning(data)}'))
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
for (const field of ['pg_num', 'pg_num_target', 'pgp_num', 'pgp_num_target']) assert.ok(detailSource.includes(`poolObjectCount(data.${field})`))
for (const field of ['read_operations', 'write_operations']) assert.ok(detailSource.includes(`poolObjectCount(data.${field})`))
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
for (const changed of [false, true]) for (const collectionFailed of [false, true]) {
  let modal, calls = [], warned = false, reads = 0
  const scope = { current: { id: 7 } }
  const remove = new Function('Modal', 'clusterScope', 'poolDeleteBlocked', 'resourceName', 'operationMutation', 'mutateResource', 'refreshResource', 'refresh', 'message', `const selectedClusterId=7, loading=false, error=null; ${deleteCode}; return deletePool`)(
    { confirm: (options) => { modal = options } }, scope, deleteGuard, (row) => row.name,
    { run: (action) => action() }, (...args) => { calls.push(args) }, async () => { if (collectionFailed) throw new Error('collection unavailable') }, async () => { reads++ }, { warning: (text) => { warned = true; assert.match(text, /删除已执行.*不要重复删除/) } }
  )
  remove({ name: 'a', stale: false, resource_version: 3 })
  assert.ok(modal.content.includes('永久删除池内全部对象'))
  if (changed) {
    scope.current = { id: 8 }
    await assert.rejects(() => modal.onOk(), /集群已切换/)
    assert.equal(calls.length, 0)
    assert.equal(warned, false)
    assert.equal(reads, 0)
  } else {
    await modal.onOk()
    assert.deepEqual(calls, [['/pool', 'DELETE', { cluster_id: 7, pool: 'a' }, { ifMatch: 3 }]])
    assert.equal(warned, collectionFailed)
    assert.equal(reads, 1)
  }
}
for (const mode of ['create', 'edit']) {
  const fn = poolPage.body.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'submitPool')
  const code = ts.transpileModule(fn.getText(poolTree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
  let finish, touched = false, calls = 0
  const scopeRef = { current: { id: 1 } }
  const run = new Function('clusterScope', 'operationMutation', 'setSubmitting', 'message', 'formMode', 'poolPlacementAvailable', 'poolUpdateBodies', 'poolEditBlocked', 'editingPool', `const selectedClusterId=1, submitting=false, loading=false, error=null, crushRuleOptions=[], erasureCodeProfileOptions=[], data={}; ${code}; return submitPool`)(scopeRef, { run: () => { calls++; return calls === 1 ? new Promise((resolve) => { finish = resolve }) : Promise.resolve() } }, () => {}, { success: () => { touched = true } }, mode, () => true, () => [{ field: 'size' }, { operation: 'rename' }], editBlocked, editable)
  const pending = run({})
  scopeRef.current = { id: 2 }
  finish()
  await pending
  assert.equal(calls, 1, `${mode} must not submit remaining mutations after switching clusters`)
  assert.equal(touched, false, `${mode} pool completion must not update the new cluster UI`)
}
{
  const fn = poolPage.body.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'submitPool')
  const code = ts.transpileModule(fn.getText(poolTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
  for (const editingPool of [null, { ...editable, stale: true }, { ...editable, resource_version: 0 }]) {
    let reported = false
    const run = new Function('editingPool', 'poolEditBlocked', 'message', `const selectedClusterId=1, submitting=false, formMode='edit', loading=false, error=null; ${code}; return submitPool`)(editingPool, editBlocked, { error: () => { reported = true } })
    await run({})
    assert.equal(reported, true, 'invalid edits must stop before mutation state or requests are touched')
  }
}
for (const formMode of ['create', 'edit']) for (const failed of [false, true]) {
  const fn = poolPage.body.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'submitPool')
  const code = ts.transpileModule(fn.getText(poolTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
  const events = []
  const env = {
    formMode, editingPool: editable, poolEditBlocked: editBlocked,
    clusterScope: { current: {} }, setSubmitting: () => {},
    operationMutation: { run: async () => { events.push('mutate') } },
    message: { success: () => {}, warning: (text) => { assert.match(text, /不要重复提交/); events.push('warning') } },
    setFormOpen: (open) => { assert.equal(open, false); events.push('close') }, setEditingPool: () => {},
    poolPlacementAvailable: () => true, poolUpdateBodies: () => [{}],
    refreshResource: async (input) => { assert.deepEqual(input, { clusterId: 1, kind: 'pool' }); events.push('collect'); if (failed) throw new Error('collection failed') },
    refresh: async () => { events.push('read') }
  }
  const run = new Function('env', `const { ${Object.keys(env).join(', ')} } = env; const selectedClusterId=1, submitting=false, loading=false, error=null, data={}, crushRuleOptions=[], erasureCodeProfileOptions=[]; ${code}; return submitPool`)(env)
  await run({})
  assert.deepEqual(events, failed ? ['mutate', 'close', 'collect', 'warning', 'read'] : ['mutate', 'close', 'collect', 'read'])
}
for (const failedIndex of [0, 1, 2]) for (const changed of [false, true]) for (const collectionFailed of [false, true]) {
  const fn = poolPage.body.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'submitPool')
  const code = ts.transpileModule(fn.getText(poolTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
  const calls = [], events = [], notices = []
  const clusterScope = { current: {} }
  const requests = [{ field: 'pg_num' }, { operation: 'quota' }, { operation: 'rename' }]
  const env = {
    formMode: 'edit', editingPool: editable, poolEditBlocked: editBlocked,
    clusterScope, setSubmitting: () => {},
    poolUpdateBodies: () => requests,
    operationMutation: { run: async (action) => action() },
    mutateResource: async (path, method, body, options) => {
      assert.equal(path, '/pool'); assert.equal(method, 'PATCH')
      assert.equal(options.ifMatch, Number(editable.resource_version))
      calls.push(body)
      if (calls.length - 1 === failedIndex) {
        if (changed) clusterScope.current = {}
        throw new Error('mutation outcome unavailable')
      }
    },
    Modal: { warning: (notice) => { notices.push(notice); events.push('notice') } },
    message: { success: () => assert.fail('partial failure must not report success'), warning: () => { events.push('collection-warning') } },
    setFormOpen: (open) => { assert.equal(open, false); events.push('close') },
    setEditingPool: (row) => { assert.equal(row, null) },
    refreshResource: async () => { events.push('collect'); if (collectionFailed) throw new Error('collection failed') },
    refresh: async () => { events.push('read') }
  }
  const run = new Function('env', `const { ${Object.keys(env).join(', ')} } = env; const selectedClusterId=1, submitting=false, loading=false, error=null, data={}; ${code}; return submitPool`)(env)
  await run({})
  assert.deepEqual(calls, requests.slice(0, failedIndex + 1), 'must never send later edits or retry the failed edit')
  if (changed) {
    assert.deepEqual(events, [])
  } else {
    assert.equal(notices.length, 1)
    assert.ok(notices[0].content.includes(`已确认成功 ${failedIndex}/3 项`))
    assert.ok(notices[0].content.includes(`第 ${failedIndex + 1} 项请求失败或结果未确认`))
    assert.match(notices[0].content, /不会自动回滚/)
    if (failedIndex > 0) assert.match(notices[0].content, /pg_num/)
    if (failedIndex > 1) assert.match(notices[0].content, /quota/)
    assert.deepEqual(events, collectionFailed ? ['notice', 'close', 'collect', 'collection-warning', 'read'] : ['notice', 'close', 'collect', 'read'])
  }
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
for (const field of ['k', 'm', 'crush_num_failure_domains', 'crush_osds_per_failure_domain']) {
  for (const value of [1.5, -1, NaN, Infinity, Number.MAX_SAFE_INTEGER + 1, '4', false]) assert.throws(() => bodyExports.body({ ...profileValues, [field]: value }, 1))
}
for (const field of ['k', 'm']) for (const value of [undefined, null, 0]) assert.throws(() => bodyExports.body({ ...profileValues, [field]: value }, 1))
for (const [plugin, field] of [['jerasure', 'packetsize'], ['lrc', 'l'], ['shec', 'c'], ['clay', 'd']]) {
  for (const value of [0, -1, 2.5, Infinity, Number.MAX_SAFE_INTEGER + 1, '3']) assert.throws(() => bodyExports.body({ ...profileValues, plugin, [field]: value }, 1))
  assert.equal(bodyExports.body({ ...profileValues, plugin, [field]: 3 }, 1)[field], 3)
  if (field !== 'packetsize') assert.throws(() => bodyExports.body({ ...profileValues, plugin }, 1))
}
for (const value of [undefined, null, 0]) {
  const body = bodyExports.body({ ...profileValues, crush_num_failure_domains: value, crush_osds_per_failure_domain: value }, 1)
  assert.ok(!('crush-num-failure-domains' in body))
  assert.ok(!('crush-osds-per-failure-domain' in body))
}
assert.equal(bodyExports.body({ ...profileValues, plugin: 'jerasure' }, 1).packetsize, undefined)
assert.equal(bodyExports.body({ ...profileValues, crush_num_failure_domains: 2 }, 1)['crush-num-failure-domains'], 2)
console.log('Erasure profile requests reject lossy integer conversion and preserve optional fields')
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

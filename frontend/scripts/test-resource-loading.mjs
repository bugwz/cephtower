import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
import './test-osd-safety.mjs'
import './test-host-create-batch.mjs'
import './test-operation-wait.mjs'
import './test-host-inventory-pages.mjs'
import './test-host-native-summary.mjs'
import './test-host-storage-capacity.mjs'
import './test-device-availability.mjs'
import './test-device-zap-scope.mjs'
import './test-device-identify-scope.mjs'
import './test-device-detail-identity.mjs'

const pgCategorySource = readFileSync(new URL('../src/pages/overview/pgCategory.ts', import.meta.url), 'utf8')
const pgCategoryCode = ts.transpileModule(pgCategorySource, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS } }).outputText
const pgCategoryExports = {}
new Function('exports', pgCategoryCode)(pgCategoryExports)
const classifyPG = pgCategoryExports.pgCategory
assert.equal(classifyPG('active+clean').key, 'clean')
assert.equal(classifyPG('active+clean+scrubbing+deep').key, 'working')
for (const state of ['activating', 'backfill_wait', 'backfilling', 'creating', 'deep', 'degraded', 'forced_backfill', 'forced_recovery', 'peering', 'peered', 'recovering', 'recovery_wait', 'repair', 'scrubbing', 'snaptrim', 'snaptrim_wait']) assert.equal(classifyPG(`active+${state}`).key, 'working')
for (const state of ['backfill_toofull', 'backfill_unfound', 'down', 'incomplete', 'inconsistent', 'recovery_toofull', 'recovery_unfound', 'remapped', 'snaptrim_error', 'stale', 'undersized']) assert.equal(classifyPG(`active+new_state+${state}`).key, 'warning')
for (const value of [null, undefined, '', ' ', 0, {}, 'active+new_state', 'clean+scrubbing+new_state', 'active1', 'constructor', 'ACTIVE']) assert.equal(classifyPG(value).key, 'unknown')
assert.equal(classifyPG(' active + clean + active ').key, 'clean')
console.log('Native PG categories preserve warning precedence and unknown compound states')
const summarizePG = pgCategoryExports.pgSummary
const pgSummaryResult = summarizePG([{ name: 'active+clean', count: 5 }, { name: 'active+scrubbing', count: 2 }, { name: 'active+degraded', count: 1 }, { name: 'undersized+new_state', count: 3 }, { name: 'active+new_state', count: 4 }])
assert.equal(pgSummaryResult.total, 15)
assert.deepEqual(pgSummaryResult.categories.map(({ key, count }) => ({ key, count })), [{ key: 'clean', count: 5 }, { key: 'working', count: 3 }, { key: 'warning', count: 3 }, { key: 'unknown', count: 4 }])
assert.equal(summarizePG([]).total, 0)
assert.equal(summarizePG([{ name: 'active+clean', count: 0 }]).total, 0)
for (const value of [null, undefined, {}, [null], [{ name: 'active' }], [{ name: '', count: 1 }], [{ name: 'active', count: 1 }, { name: 'active', count: 2 }]]) assert.equal(summarizePG(value), null)
for (const count of [null, undefined, '1', -1, 1.5, NaN, Infinity, Number.MAX_SAFE_INTEGER + 1]) assert.equal(summarizePG([{ name: 'active', count }]), null)
assert.equal(summarizePG([{ name: 'active', count: Number.MAX_SAFE_INTEGER }, { name: 'clean', count: 1 }]), null)
console.log('PG category totals distinguish valid zero counts from invalid or missing statistics')
const poolDistribution = pgCategoryExports.poolPGDistribution
const poolPG = poolDistribution({ 'active+clean': 4, 'active+scrubbing': 1, undersized: 2, new_state: 3 })
assert.equal(poolPG.total, 10)
assert.deepEqual(poolPG.rows.map((row) => row.category.key), ['clean', 'working', 'warning', 'unknown'])
assert.equal(poolDistribution({}).total, 0)
assert.equal(poolDistribution({ 'active+clean': 0 }).total, 0)
for (const value of [null, [], false, 'active+clean', { 'active+clean': null }, { 'active+clean': '3' }, { '': 1 }, { active: -1 }, { active: Number.MAX_SAFE_INTEGER, clean: 1 }]) assert.equal(poolDistribution(value), null)
const poolPGDetailSource = readFileSync(new URL('../src/pages/cluster/PoolDetailPage.tsx', import.meta.url), 'utf8')
assert.ok(poolPGDetailSource.includes('PoolPGDistribution key={`${selectedClusterId}/${decodedName}`} value={data?.pg_status}'))
console.log('Pool PG distribution uses scoped raw counts and keeps unavailable data distinct')
const pgTagsSource = readFileSync(new URL('../src/pages/cluster/PoolPGStateTags.tsx', import.meta.url), 'utf8')
const pgTagsTree = ts.createSourceFile('tags.tsx', pgTagsSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const pgTagsNode = pgTagsTree.statements.find((node) => ts.isFunctionDeclaration(node))
const pgTagsCode = ts.transpileModule(pgTagsNode.getText(pgTagsTree).replace('export function', 'function'), { compilerOptions: { target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.React } }).outputText
const renderPGTags = new Function('React', 'Space', 'Tag', 'Typography', 'poolPGDistribution', `${pgTagsCode}; return PoolPGStateTags`)({ createElement: (type, props, ...children) => ({ type, props, children }) }, 'Space', 'Tag', { Text: 'Text' }, poolDistribution)
const renderedPGTags = renderPGTags({ value: { 'active+clean': 3, undersized: 2, new_state: 1 } }).children[0]
assert.deepEqual(renderedPGTags.map((tag) => tag.props.color), ['success', 'warning', 'default'])
assert.ok(renderedPGTags[0].children.join('').includes('active+clean × 3'))
assert.ok(renderedPGTags[2].children.join('').includes('未知'))
assert.equal(renderPGTags({ value: null }).children[0], 'PG 状态不可用')
assert.equal(renderPGTags({ value: {} }).children[0], '未返回 PG（0）')
console.log('Pool PG tags preserve raw states, counts and accessible category labels')

const hostDetailSource = readFileSync(new URL('../src/pages/cluster/HostDetailPage.tsx', import.meta.url), 'utf8')
const hostDetailTree = ts.createSourceFile('host.tsx', hostDetailSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const hostDetailFn = hostDetailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'HostDetailContent')
const hostDetailWrapper = hostDetailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'HostDetailPage').getText(hostDetailTree)
assert.ok(hostDetailWrapper.includes('key={JSON.stringify([selectedClusterId, name])}'))
assert.ok(hostDetailWrapper.includes('name={name} selectedClusterId={selectedClusterId}'))
const hostLoader = hostDetailFn.body.statements.find((node) => ts.isVariableStatement(node) && node.declarationList.declarations[0].name.getText(hostDetailTree) === 'loader').declarationList.declarations[0].initializer.arguments[0]
const hostLoaderCode = ts.transpileModule(`const load = ${hostLoader.getText(hostDetailTree)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const failed of ['none', 'device', 'smart', 'both']) {
  const env = {
    selectedClusterId: 3, decodedName: 'node1', getOptionalResource: async () => ({ item: { hostname: 'node1' } }),
    listAllResources: async (path, cluster) => { assert.equal(cluster, 3); return { items: path === '/devices' ? [{ name: 'disk1', hostname: 'node1' }, { name: 'disk2', hostname: 'node2' }] : [] } },
    getHostDeviceInfo: async () => { if (['device', 'both'].includes(failed)) throw new Error('device unavailable'); return [] },
    getHostSMART: async () => { if (['smart', 'both'].includes(failed)) throw new Error('smart unavailable'); return {} },
    normalizeHostRow: (row) => row, textValue: (value) => value, resourceToRecord: (row) => row, normalizeDaemonRow: (row) => row,
  }
  const load = new Function(...Object.keys(env), `${hostLoaderCode}; return load`)(...Object.values(env))
  const result = await load()
  assert.equal(result.devices.length, 1)
  assert.equal(result.deviceInfoError, ['device', 'both'].includes(failed) ? 'device unavailable' : '')
  assert.equal(result.smartError, ['smart', 'both'].includes(failed) ? 'smart unavailable' : '')
  assert.equal(result.host.hostname, 'node1')
  env.getOptionalResource = async () => null
  const loadMissing = new Function(...Object.keys(env), `${hostLoaderCode}; return load`)(...Object.values(env))
  const missing = await loadMissing()
  assert.equal(missing.host, null)
  assert.equal(missing.devices.length, 1)
}
console.log('Host diagnostic failures remain distinct from empty responses')
for (const stale of [false, true, undefined]) {
  const env = {
    selectedClusterId: 3, decodedName: 'node1', getOptionalResource: async () => null,
    listAllResources: async () => ({ items: [], stale, staleReason: stale ? 'collector offline' : null }),
    getHostDeviceInfo: async () => [], getHostSMART: async () => ({}),
    normalizeHostRow: row => row, textValue: value => value, resourceToRecord: row => row, normalizeDaemonRow: row => row
  }
  const result = await new Function(...Object.keys(env), `${hostLoaderCode}; return load`)(...Object.values(env))()
  assert.equal(result.inventoryWarnings.length, stale === false ? 0 : 2)
  if (stale) assert.ok(result.inventoryWarnings.every(warning => warning.includes('collector offline')))
}
assert.ok(hostDetailSource.includes('data?.inventoryWarnings.map'))
const deleteHostNode = hostDetailFn.body.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'deleteHost')
const deleteHostCode = ts.transpileModule(deleteHostNode.getText(hostDetailTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const hostVersionNode = hostDetailTree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'hostDeleteVersion')
const hostVersionCode = ts.transpileModule(hostVersionNode.getText(hostDetailTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const hostDeleteVersion = new Function(`${hostVersionCode}; return hostDeleteVersion`)()
assert.equal(hostDeleteVersion({ stale: false, resource_version: '18446744073709551615' }), '18446744073709551615')
for (const host of [null, {}, { stale: true, resource_version: 1 }, { stale: false, resource_version: 9007199254740992 }, { stale: false, resource_version: '01' }, { stale: false, resource_version: '18446744073709551616' }]) assert.equal(hostDeleteVersion(host), null)
for (const timing of ['current', 'before-confirm', 'during-request']) {
  const active = { current: true }
  let confirmation
  let writes = 0
  let navigations = 0
  const env = {
    selectedClusterId: 3, host: { hostname: 'node1', stale: false, resource_version: '18446744073709551615' }, hostName: (row) => row.hostname,
    hostDeleteVersion, loading: false, error: '', pendingAction: '',
    message: { success() {}, error() {} }, active, deleteConfirmation: { current: null },
    Modal: { confirm(options) { confirmation = options; return { destroy() {} } } },
    operationMutation: { run: (fn) => fn() },
    mutateResource: async (...args) => { assert.deepEqual(args, ['/host', 'DELETE', { cluster_id: 3, host: 'node1' }, { ifMatch: '18446744073709551615' }]); writes++; if (timing === 'during-request') active.current = false },
    navigate: () => { navigations++ },
  }
  const remove = new Function(...Object.keys(env), `${deleteHostCode}; return deleteHost`)(...Object.values(env))
  await remove()
  if (timing === 'before-confirm') active.current = false
  await confirmation.onOk()
  await confirmation.onOk()
  assert.equal(writes, timing === 'before-confirm' ? 0 : 1)
  assert.equal(navigations, timing === 'current' ? 1 : 0)
}
assert.ok(hostDetailFn.getText(hostDetailTree).includes('deleteConfirmation.current?.destroy()'))
console.log('Host deletion cannot submit or navigate from an inactive detail page')
const hostActionNode = hostDetailFn.body.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'runHostAction')
const hostLabelNode = hostDetailFn.body.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'submitHostLabel')
const hostLabelCode = ts.transpileModule(hostLabelNode.getText(hostDetailTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const scenario of ['add', 'rm', 'failure', 'unmounted', 'stale', 'running', 'no-version', 'invalid-action']) {
  const calls = [], updates = [], active = { current: true }
  const env = {
    selectedClusterId: 3, host: { hostname: 'node1', stale: scenario === 'stale', resource_version: '18446744073709551615' },
    active, hostName: row => row.hostname, hostDeleteVersion, submitting: false, loading: false, error: '',
    labelRunning: { current: scenario === 'running' }, actionRunning: { current: false }, labelVersion: scenario === 'no-version' ? null : '18446744073709551614',
    setSubmitting: value => updates.push(['busy', value]), setLabelVersion: value => updates.push(['version', value]),
    setLabelModalOpen: value => updates.push(['open', value]), message: { error() {}, success() {} }, refresh: () => updates.push(['refresh']),
    operationMutation: { run: fn => fn() }, mutateResource: async (...args) => { calls.push(args); if (scenario === 'unmounted') active.current = false; if (scenario === 'failure') throw new Error('failed') }
  }
  const save = new Function(...Object.keys(env), `${hostLabelCode}; return submitHostLabel`)(...Object.values(env))
  const values = { label: ' custom ', action: scenario === 'rm' ? 'rm' : scenario === 'invalid-action' ? 'other' : 'add' }
  if (scenario === 'failure') await assert.rejects(save(values), /failed/); else await save(values)
  if (['stale', 'running', 'no-version', 'invalid-action'].includes(scenario)) { assert.deepEqual(calls, []); continue }
  assert.deepEqual(calls, [['/host', 'PATCH', { cluster_id: 3, host: 'node1', labels_add: scenario === 'rm' ? [] : ['custom'], labels_remove: scenario === 'rm' ? ['custom'] : [] }, { ifMatch: '18446744073709551614' }]])
  if (scenario === 'unmounted') assert.deepEqual(updates, [['busy', true]])
  else assert.ok(updates.some(([key, value]) => key === 'version' && value === null))
}
const hostActionCode = ts.transpileModule(hostActionNode.getText(hostDetailTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const action of ['maintenance_enter', 'maintenance_exit', 'drain', 'stop_drain', 'rescan']) {
  for (const scenario of ['success', 'stale', 'loading', 'error', 'running', 'before-confirm', 'during-request', 'failure']) {
    const active = { current: true }, calls = [], updates = []
    let confirmation
    const env = {
      selectedClusterId: 3, host: { hostname: 'node1', stale: scenario === 'stale', resource_version: '18446744073709551615' },
      hostName: row => row.hostname, hostDeleteVersion, active, actionRunning: { current: scenario === 'running' },
      pendingAction: '', loading: scenario === 'loading', error: scenario === 'error' ? 'offline' : '',
      message: { error() {}, success: value => updates.push(value) }, deleteConfirmation: { current: null },
      Modal: { confirm: options => { confirmation = options; return { destroy() {} } } },
      setPendingAction: value => updates.push(value), operationMutation: { run: fn => fn() }, refresh: async () => updates.push('refresh'),
      mutateResource: async (...args) => { calls.push(args); if (scenario === 'during-request') active.current = false; if (scenario === 'failure') throw new Error('failed') }
    }
    const run = new Function(...Object.keys(env), `${hostActionCode}; return runHostAction`)(...Object.values(env))
    await run(action)
    if (['stale', 'loading', 'error', 'running'].includes(scenario)) { assert.equal(confirmation, undefined); assert.deepEqual(calls, []); continue }
    assert.ok(confirmation.title.includes('集群 3 / 主机 node1'))
    assert.deepEqual(calls, [])
    if (scenario === 'before-confirm') active.current = false
    if (scenario === 'failure') await assert.rejects(confirmation.onOk(), /failed/); else await confirmation.onOk()
    await confirmation.onOk()
    assert.equal(calls.length, scenario === 'before-confirm' ? 0 : 1)
    if (calls.length) assert.deepEqual(calls[0], ['/host/action', 'POST', { cluster_id: 3, host: 'node1', action }, { ifMatch: '18446744073709551615' }])
    assert.equal(updates.includes('refresh'), scenario === 'success')
  }
}
const hostPerfNode = hostDetailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'hostDaemonSupportsPerf')
const hostPerfCode = ts.transpileModule(hostPerfNode.getText(hostDetailTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const hostPerfSupported = new Function(`${hostPerfCode}; return hostDaemonSupportsPerf`)()
for (const name of ['osd.0', 'mon.a', 'mgr.a.x', 'mds.fs.a']) assert.equal(hostPerfSupported(name), true)
for (const name of [null, '', 'osd.', 'prometheus.a', 'osd.*', 'osd.1;ls', 'rgw.store.a', 'rbd-mirror.a']) assert.equal(hostPerfSupported(name), false)
const hostDaemonTable = hostDetailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'HostDaemonTable').getText(hostDetailTree)
assert.ok(hostDetailTree.text.includes('HostDaemonTable key={`${clusterId}:${hostname}`}'))
assert.ok(hostDaemonTable.includes('row.daemon_display === perfName'))
assert.ok(hostDaemonTable.includes('DaemonPerf key={`${clusterId}:${visibleName}`}'))
console.log('Host daemon performance entry point is restricted and scope keyed')
const daemonStateHelpers = {}
for (const name of ['nativeDaemonStatus', 'daemonStatusColor', 'daemonStatusText']) {
  const node = hostDetailTree.statements.find((item) => ts.isFunctionDeclaration(item) && item.name.text === name)
  const code = ts.transpileModule(node.getText(hostDetailTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
  daemonStateHelpers[name] = new Function(`${code}; return ${name}`)()
}
for (const [code, state] of [[-2, 'unknown'], [-1, 'error'], [0, 'stopped'], [1, 'running'], [2, 'starting']]) {
  assert.equal(daemonStateHelpers.nativeDaemonStatus({ status: code }), state)
  assert.equal(daemonStateHelpers.nativeDaemonStatus({ status: String(code) }), state)
}
assert.equal(daemonStateHelpers.nativeDaemonStatus({ status: 0, status_desc: 'running' }), 'stopped')
assert.equal(daemonStateHelpers.nativeDaemonStatus({ status: null, status_desc: 'starting' }), 'starting')
assert.equal(daemonStateHelpers.nativeDaemonStatus({ status: 99 }), 'unknown')
assert.equal(daemonStateHelpers.nativeDaemonStatus({}), 'unknown')
for (const state of ['not running', 'not ok', 'broken']) {
  assert.equal(daemonStateHelpers.daemonStatusColor(state), 'default')
  assert.equal(daemonStateHelpers.daemonStatusText(state), state)
}
assert.equal(daemonStateHelpers.daemonStatusColor('running'), 'success')
assert.equal(daemonStateHelpers.daemonStatusText('stopped'), '已停止')
console.log('Native daemon status codes and unrecognized descriptions remain distinct')
for (const state of ['constructor', '__proto__', 'toString', 'hasOwnProperty', 'valueOf']) {
  const normalized = daemonStateHelpers.nativeDaemonStatus({ status: state })
  assert.equal(normalized, state)
  assert.equal(daemonStateHelpers.daemonStatusColor(normalized), 'default')
  assert.equal(daemonStateHelpers.daemonStatusText(normalized), state)
}
const deviceHealthNode = hostDetailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'renderDeviceHealth')
const deviceHealthCode = ts.transpileModule(deviceHealthNode.getText(hostDetailTree), { compilerOptions: { target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.React } }).outputText
const renderDeviceHealth = new Function('React', 'Tag', 'textValue', `${deviceHealthCode}; return renderDeviceHealth`)({ createElement: (_, props, label) => ({ color: props.color, label }) }, 'Tag', (value, fallback = '-') => typeof value === 'string' && value.trim() ? value : fallback)
for (const state of ['constructor', '__proto__', 'toString', 'hasOwnProperty', 'valueOf', 'new-health-state']) {
  assert.deepEqual(renderDeviceHealth(state), { label: state, color: 'default' })
}
assert.deepEqual(renderDeviceHealth('good'), { label: '良好', color: 'success' })
assert.deepEqual(renderDeviceHealth('failed'), { label: '异常', color: 'error' })
assert.deepEqual(renderDeviceHealth(null), { label: '未知', color: 'default' })
console.log('Unrecognized daemon and device health states cannot resolve inherited properties')
const deviceInfoTable = hostDetailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'HostDeviceInfoTable')
assert.ok(deviceInfoTable.getText(hostDetailTree).includes("key: 'life_expectancy_stamp'"))
const timeSource = readFileSync(new URL('../src/utils/time.ts', import.meta.url), 'utf8')
const timeCode = ts.transpileModule(timeSource, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS } }).outputText
const timeExports = {}
new Function('exports', timeCode)(timeExports)
for (const value of [undefined, null, '']) assert.equal(timeExports.formatDateTime(value), '-')
assert.equal(timeExports.formatDateTime('2026-10-03 01:02:03.000000'), '2026-10-03 01:02:03.000000')
assert.match(timeExports.formatDateTime('2026-10-03T01:02:03Z'), /^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/)
console.log('Device prediction creation timestamp display checks passed')
const inventoryNode = hostDetailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'normalizeInventoryDeviceRow')
const inventoryCode = ts.transpileModule(inventoryNode.getText(hostDetailTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const osdNamesNode = hostDetailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'inventoryOSDNames')
const osdNamesCode = ts.transpileModule(osdNamesNode.getText(hostDetailTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const capacityModule = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/cluster/hostStorageCapacity.ts', import.meta.url), 'utf8'), { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS } }).outputText)(capacityModule)
const normalizeInventory = new Function('isRecord', 'stringArray', 'textValue', 'numberValue', 'formatBytes', 'hostStorageCapacity', `${osdNamesCode}; ${inventoryCode}; return normalizeInventoryDeviceRow`)(
  (v) => v !== null && typeof v === 'object' && !Array.isArray(v),
  (v) => Array.isArray(v) ? v.filter((item) => typeof item === 'string') : [],
  (v, fallback) => v ?? fallback, (v) => typeof v === 'number' ? v : undefined, String, capacityModule.hostStorageCapacity)
assert.match(normalizeInventory({ size_bytes: '18446744073709551615' }, []).size_display, /18446744073709551615 B/)
assert.equal(normalizeInventory({ size_bytes: '0' }, []).size_display, '0 B')
assert.match(normalizeInventory({}, []).size_display, /未知/)
const lsmRow = normalizeInventory({ lsm_data: { health: 'Good', transport: 'SAS', rpm: '7200', linkSpeed: '12000', ledSupport: { IDENTsupport: 'Supported', IDENTstatus: 'Off', FAILsupport: 'Unsupported', FAILstatus: 'Unknown' } } }, [])
assert.equal(lsmRow.rpm_display, '7200')
assert.equal(lsmRow.link_speed_display, '12000')
assert.equal(lsmRow.health_display, 'Good')
assert.equal(lsmRow.identify_led_display, 'Supported / Off')
assert.equal(lsmRow.fault_led_display, 'Unsupported / Unknown')
assert.equal(normalizeInventory({}, []).identify_led_display, '未知 / 未知')
assert.equal(normalizeInventory({ lsm_data: { errors: ['query failed', 'unsupported transport'] } }, []).lsm_errors_display, 'query failed；unsupported transport')
assert.equal(normalizeInventory({ lsm_data: { errors: [] } }, []).lsm_errors_display, '未报告查询错误（不代表健康）')
for (const errors of [null, undefined, ['failed', 1], 'failed']) assert.equal(normalizeInventory({ lsm_data: { errors } }, []).lsm_errors_display, '未报告')
for (const rotational of [undefined, null, '', 'unknown', 2]) {
  const row = normalizeInventory({ rotational }, [])
  assert.equal(row.type_display, '未知')
  assert.equal(row.availability_display, 'unknown')
}
for (const rotational of [true, '1', 1]) assert.equal(normalizeInventory({ sys_api: { rotational } }, []).type_display, 'HDD')
for (const rotational of [false, '0', 0]) assert.equal(normalizeInventory({ rotational }, []).type_display, 'SSD')
assert.equal(normalizeInventory({ rotational: false, sys_api: { rotational: '1' } }, []).type_display, 'SSD')
assert.equal(normalizeInventory({ device_type: 'nvme' }, []).type_display, 'NVME')
assert.equal(normalizeInventory({ available: false }, []).availability_display, 'unavailable')
assert.equal(normalizeInventory({ available: true }, []).availability_display, 'available')
assert.equal(normalizeInventory({ available: 'false' }, []).availability_display, 'unknown')
console.log('Disk inventory distinguishes missing type and availability from explicit values')
const osdDisk = normalizeInventory({ path: '/dev/sda', osd_ids: [0, 1, '1'], lvs: [{ osd_id: '01' }, { osd_id: '2' }] }, [{ device_names: ['sda'], daemons_display: ['osd.1', 'osd.3', 'mon.a'] }])
assert.deepEqual(osdDisk.osd_display, ['osd.0', 'osd.1', 'osd.2', 'osd.3'])
const cephDeviceNode = hostDetailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'normalizeCephDeviceRow')
const cephDeviceCode = ts.transpileModule(cephDeviceNode.getText(hostDetailTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const normalizeCephDevice = new Function('isRecord', 'stringArray', 'textValue', 'formatLifeExpectancy', `${cephDeviceCode}; return normalizeCephDeviceRow`)(
  (v) => v !== null && typeof v === 'object' && !Array.isArray(v),
  (v) => Array.isArray(v) ? v.filter((item) => typeof item === 'string') : [],
  (v, fallback) => typeof v === 'string' && v.trim() ? v : fallback, () => '-')
const sharedDevice = { devid: 'shared-disk', location: [{ host: 'node1', dev: 'sda' }, { host: 'node2', dev: 'sdb' }, { dev: 'sdc' }, null], daemons: ['osd.3'] }
const node1Device = normalizeCephDevice(sharedDevice, 'node1')
const node2Device = normalizeCephDevice(sharedDevice, 'node2')
assert.deepEqual(node1Device.device_names, ['sda'])
assert.equal(node1Device.name_display, 'sda')
assert.deepEqual(node2Device.device_names, ['sdb'])
assert.equal(normalizeCephDevice(sharedDevice, 'node3').name_display, '-')
assert.deepEqual(normalizeInventory({ path: '/dev/sdb' }, [node1Device]).osd_display, [])
assert.deepEqual(normalizeInventory({ path: '/dev/sda' }, [node1Device]).osd_display, ['osd.3'])
assert.deepEqual(normalizeInventory({ path: '/dev/sdb' }, [node2Device]).osd_display, ['osd.3'])
for (const location of [null, {}, [], [null], [{ host: 'node1' }]]) assert.deepEqual(normalizeCephDevice({ location }, 'node1').device_names, [])
assert.deepEqual(normalizeInventory({ path: '/dev/sda' }, [{ name_display: 'sda', device_names: [], daemons_display: ['osd.9'] }]).osd_display, [])
console.log('Host device locations and derived OSD associations stay within the selected host')
assert.deepEqual(normalizeInventory({ osd_ids: [null, false, -1, 1.5, '', 'osd.', 'osd.-1', 'mon.1', Number.MAX_SAFE_INTEGER + 1] }, []).osd_display, [])
assert.deepEqual(normalizeInventory({ osd_ids: ['9007199254740993'] }, []).osd_display, ['osd.9007199254740993'])
const diagnosticDisk = normalizeInventory({ lsm_data: { health: 'Fail', serialNum: 'lsm-serial' }, rejected_reasons: ['Has a FileSystem', 'LVM detected'] }, [])
assert.equal(diagnosticDisk.serial_display, 'lsm-serial')
assert.equal(diagnosticDisk.health_display, 'Fail')
assert.deepEqual(diagnosticDisk.rejected_reasons_display, ['Has a FileSystem', 'LVM detected'])
assert.equal(normalizeInventory({ lsm_data: { serialNum: 'lsm-serial' } }, []).serial_display, 'lsm-serial')
const physicalTable = hostDetailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'HostPhysicalDiskTable').getText(hostDetailTree)
for (const field of ['serial_display', 'health_display', 'rejected_reasons_display']) assert.ok(physicalTable.includes(`key: '${field}'`))
const smartNode = hostDetailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'normalizeSMARTData')
const smartCode = ts.transpileModule(smartNode.getText(hostDetailTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const smartNumberNode = hostDetailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'smartMetricNumber')
const smartNumberCode = ts.transpileModule(smartNumberNode.getText(hostDetailTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const smartNumber = new Function(`${smartNumberCode}; return smartMetricNumber`)()
const normalizeSmart = new Function('isRecord', 'textValue', 'smartMetricNumber', 'formatTemperature', 'formatHours', 'formatWear', `${smartCode}; return normalizeSMARTData`)(
  (v) => v !== null && typeof v === 'object' && !Array.isArray(v), (v, fallback) => v ?? fallback,
  smartNumber, (v) => v, (v) => v, (v) => v)
for (const value of [undefined, null, '', ' ', '\t', false, [], {}, '0x10', 'Infinity', NaN, Infinity, '9007199254740993']) {
  const row = normalizeSmart({ disk: { temperature: { current: value }, percentage_used: value } })[0]
  assert.equal(row.temperature_display, undefined)
  assert.equal(row.wear_level_display, undefined)
}
for (const [value, expected] of [['0', 0], [0, 0], ['42.5', 42.5], ['-5', -5], ['120', 120]]) {
  assert.equal(smartNumber(value), expected)
}
const numericSmart = normalizeSmart({ disk: { nvme_smart_health_information_log: { temperature: '42', percentage_used: '0' } } })[0]
assert.equal(numericSmart.temperature_display, 42)
assert.equal(numericSmart.wear_level_display, 0)
const smartRows = normalizeSmart({ disk1: { smart_status: { passed: true } }, disk2: { error: 'unsupported device', smartctl_error_code: -22 }, disk3: { error: '', smart_status: { passed: true } }, disk4: null, disk5: { smart_status: { passed: false } } })
assert.equal(smartRows.length, 5)
assert.equal(smartRows[0].health_display, 'good')
assert.equal(smartRows[1].health_display, 'unavailable')
assert.equal(smartRows[1].smartctl_error_code, -22)
assert.equal(smartRows[1].error_display, 'unsupported device')
assert.equal(smartRows[2].health_display, 'unavailable')
assert.equal(smartRows[3].health_display, 'unavailable')
assert.equal(smartRows[4].health_display, 'bad')
assert.deepEqual(normalizeSmart({}), [])
console.log('Per-device SMART failures remain visible with unknown health')
const smartWearRows = normalizeSmart({
  nvme: { nvme_smart_health_information_log: { percentage_used: 0, power_on_hours: 13 }, ssd_life_left: 100 },
  remaining: { ssd_life_left: 80 },
  worn: { percentage_used: 120, power_on_time: { hours: 0 }, nvme_smart_health_information_log: { power_on_hours: 13 } },
})
assert.equal(smartWearRows[0].wear_level_display, 0)
assert.equal(smartWearRows[0].power_on_hours_display, 13)
assert.equal(smartWearRows[0].ssd_life_left_display, '100')
assert.equal(smartWearRows[1].wear_level_display, undefined)
assert.equal(smartWearRows[1].ssd_life_left_display, '80')
assert.equal(smartWearRows[2].wear_level_display, 120)
assert.equal(smartWearRows[2].power_on_hours_display, 0)
console.log('SMART used-life and remaining-life fields are kept distinct')
const detailedSmart = normalizeSmart({ disk: { nvme_smart_health_information_log: { data_units_written: '18446744073709551615' }, ata_smart_attributes: { table: [{ raw: { value: '9007199254740993' } }] }, scsi_error_counter_log: { read: { total_uncorrected_errors: '0' } } } })[0]
assert.ok(JSON.stringify(detailedSmart.smart_report).includes('18446744073709551615'))
assert.ok(JSON.stringify(detailedSmart.smart_report).includes('9007199254740993'))
assert.ok(JSON.stringify(detailedSmart.smart_report).includes('scsi_error_counter_log'))
assert.equal(smartRows[1].smart_report.error, 'unsupported device')
const hoursNode = hostDetailTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'formatHours')
const hoursCode = ts.transpileModule(hoursNode.getText(hostDetailTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const smartHours = new Function(`${hoursCode}; return formatHours`)()
assert.equal(smartHours('18446744073709551615'), '18446744073709551615 小时')
assert.equal(smartHours('0'), '0 小时')
for (const value of [undefined, null, '', 'not-hours', Number.MAX_SAFE_INTEGER + 1]) assert.equal(smartHours(value), '-')
console.log('SMART native detail and exact power-on hour display checks passed')
const protocolSource = readFileSync(new URL('../src/pages/cluster/SMARTDetails.tsx', import.meta.url), 'utf8')
const protocolTree = ts.createSourceFile('protocol.tsx', protocolSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const protocolHelpers = {}
for (const name of ['ataRows', 'scsiRows']) {
  const fn = protocolTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === name)
  const code = ts.transpileModule(fn.getText(protocolTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
  protocolHelpers[name] = new Function('isRecord', `${code}; return ${name}`)((v) => v !== null && typeof v === 'object' && !Array.isArray(v))
}
const ataReport = { table: [{ id: '9', raw: { value: '18446744073709551615' }, value: '0' }, { id: '9' }] }
const ataTable = protocolHelpers.ataRows(ataReport)
assert.equal(ataTable[0].raw_value, '18446744073709551615')
assert.equal(ataTable[0].value, '0')
assert.equal(ataTable[1].raw_value, null)
assert.notEqual(ataTable[0].row_key, ataTable[1].row_key)
assert.deepEqual(protocolHelpers.scsiRows({ read: { total_uncorrected_errors: '9007199254740993' } }), [{ operation: 'read', total_uncorrected_errors: '9007199254740993' }])
for (const value of [null, [], { table: [null] }]) assert.equal(protocolHelpers.ataRows(value), null)
for (const value of [null, [], { read: null }]) assert.equal(protocolHelpers.scsiRows(value), null)
assert.deepEqual(protocolHelpers.ataRows({ table: [] }), [])
assert.deepEqual(protocolHelpers.scsiRows({}), [])
console.log('SMART protocol tables preserve counters and reject invalid section shapes')

const hardwareSource = readFileSync(new URL('../src/pages/cluster/HostHardware.tsx', import.meta.url), 'utf8')
const hardwareTree = ts.createSourceFile('hardware.tsx', hardwareSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const hardwareFn = hardwareTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'HardwareCategory')
const hardwareLoader = hardwareFn.body.statements.find((node) => ts.isVariableStatement(node) && node.declarationList.declarations[0].name.getText(hardwareTree) === 'loader').declarationList.declarations[0].initializer.arguments[0]
const hardwareCode = ts.transpileModule(`const load = ${hardwareLoader.getText(hardwareTree)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const response of [{ host: 'node1', category: 'memory', items: [] }, { host: 'other', category: 'memory', items: [] }, { host: 'node1', category: 'fans', items: [] }, { host: 'node1', category: 'memory', items: null }]) {
  const calls = []
  const load = new Function('request', 'jsonInit', 'clusterId', 'host', 'category', `${hardwareCode}; return load`)(async (...args) => { calls.push(args); return response }, (method, body) => ({ method, body }), 7, 'node1', 'memory')
  if (response.host === 'node1' && response.category === 'memory' && Array.isArray(response.items)) assert.equal(await load(), response)
  else await assert.rejects(load())
  assert.deepEqual(calls, [['/host/hardware', { method: 'GET', body: { cluster_id: 7, host: 'node1', category: 'memory' } }]])
}
assert.ok(hardwareSource.includes('key={`${clusterId}:${host}:${category}`}'))
assert.ok(hardwareSource.includes('未知（未返回）'))
console.log('Hardware request identity and category scope checks passed')

const firmwareResponse = { host: 'node1', category: 'firmwares', items: [{ id: 'bios', version: '01.02' }] }
const firmwareCalls = []
const firmwareLoad = new Function('request', 'jsonInit', 'clusterId', 'host', 'category', `${hardwareCode}; return load`)(async (...args) => { firmwareCalls.push(args); return firmwareResponse }, (method, body) => ({ method, body }), 7, 'node1', 'firmwares')
assert.equal(await firmwareLoad(), firmwareResponse)
assert.deepEqual(firmwareCalls, [['/host/hardware', { method: 'GET', body: { cluster_id: 7, host: 'node1', category: 'firmwares' } }]])
const hostHardwareFn = hardwareTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'HostHardware')
let hardwareOptions
function findHardwareOptions(node) {
  if (ts.isJsxAttribute(node) && node.name.text === 'options') hardwareOptions = node.initializer.expression
  ts.forEachChild(node, findHardwareOptions)
}
findHardwareOptions(hostHardwareFn)
const optionsCode = ts.transpileModule(`const options = ${hardwareOptions.getText(hardwareTree)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const hardwareOptionsForHost = new Function('host', `${optionsCode}; return options`)
assert.equal(hardwareOptionsForHost('node1').some((option) => option.value === 'firmwares'), true)
assert.equal(hardwareOptionsForHost('').some((option) => option.value === 'firmwares'), false)
assert.equal(hardwareOptionsForHost('node1').some((option) => option.value === 'fullreport'), true)
assert.equal(hardwareOptionsForHost('').some((option) => option.value === 'fullreport'), false)
assert.equal(hardwareOptionsForHost('').some((option) => option.value === 'criticals'), true)
assert.equal(hardwareOptionsForHost('node1').some((option) => option.value === 'criticals'), true)
console.log('Firmware request and host-only selection checks passed')
const fullReportSource = readFileSync(new URL('../src/pages/cluster/HostHardwareFullReport.tsx', import.meta.url), 'utf8')
const fullReportTree = ts.createSourceFile('fullreport.tsx', fullReportSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const fullReportFn = fullReportTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'HostHardwareFullReport')
const fullReportLoader = fullReportFn.body.statements.find((node) => ts.isVariableStatement(node) && node.declarationList.declarations[0].name.getText(fullReportTree) === 'loader').declarationList.declarations[0].initializer.arguments[0]
const fullReportCode = ts.transpileModule(`const load = ${fullReportLoader.getText(fullReportTree)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const fullReportResponse = { host: 'node1', category: 'fullreport', serial_number: '00123', report: '{"capacity_bytes":18446744073709551615}' }
for (const response of [fullReportResponse, { ...fullReportResponse, host: 'other' }, { ...fullReportResponse, category: 'memory' }, { ...fullReportResponse, report: {} }, { ...fullReportResponse, report: ' ' }, { ...fullReportResponse, serial_number: 123 }]) {
  const calls = []
  const load = new Function('request', 'jsonInit', 'clusterId', 'host', `${fullReportCode}; return load`)(async (...args) => { calls.push(args); return response }, (method, body) => ({ method, body }), 7, 'node1')
  if (response === fullReportResponse) assert.equal(await load(), fullReportResponse)
  else await assert.rejects(load())
  assert.deepEqual(calls, [['/host/hardware', { method: 'GET', body: { cluster_id: 7, host: 'node1', category: 'fullreport' } }]])
}
assert.ok(!fullReportSource.includes('JSON.parse'))
console.log('Full hardware report identity, precision and host-only selection checks passed')
const healthSource = readFileSync(new URL('../src/pages/cluster/hardwareHealth.ts', import.meta.url), 'utf8')
const healthCode = ts.transpileModule(healthSource, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS } }).outputText
const healthModule = {}
new Function('exports', healthCode)(healthModule)
for (const value of [null, undefined, '', '  ', 0, false]) assert.equal(healthModule.hardwareHealthGroup(value), 'unknown')
assert.equal(healthModule.hardwareHealthGroup('OK'), 'ok')
for (const value of ['Warning', 'Critical', 'Unknown', 'NEW_STATUS', 'ok']) assert.equal(healthModule.hardwareHealthGroup(value), 'other')
assert.deepEqual(healthModule.hardwareHealthCounts([{ health: 'OK' }, { health: 'Warning' }, { health: 'Critical' }, {}, { health: '' }]), { total: 5, ok: 1, other: 2, unknown: 2 })
assert.deepEqual(healthModule.hardwareHealthCounts([]), { total: 0, ok: 0, other: 0, unknown: 0 })
for (const [category, keys] of Object.entries({ memory: ['description'], storage: ['description', 'model', 'capacity_bytes', 'protocol', 'serial_number'], processors: ['model', 'total_cores', 'total_threads'], network: ['name', 'speed_mbps'], power: ['name', 'model', 'manufacturer'], fans: ['name'], firmwares: ['name', 'version', 'release_date'] })) {
  const columns = healthModule.hardwareAttributeColumns(category)
  assert.deepEqual(columns.map((column) => column.key), keys)
  for (const column of columns) {
    assert.equal(column.render('18446744073709551615'), '18446744073709551615')
    assert.equal(column.render('0'), '0')
    assert.equal(column.render('01.02'), '01.02')
    for (const value of [null, undefined, '', '  ', false, {}, []]) assert.equal(column.render(value), '未知（未返回）')
  }
}
assert.deepEqual(healthModule.hardwareAttributeColumns('constructor'), [])
assert.deepEqual(healthModule.hardwareAttributeColumns('criticals').map((column) => column.key), ['category', 'name'])
console.log('Native hardware attribute column checks passed')
console.log('Hardware health counts and unknown-state classification checks passed')
const summarySource = readFileSync(new URL('../src/pages/cluster/HostHardwareSummary.tsx', import.meta.url), 'utf8')
const summaryTree = ts.createSourceFile('summary.tsx', summarySource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const summaryFn = summaryTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'loadHardwareSummary')
const summaryCode = ts.transpileModule(summaryFn.getText(summaryTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const summaryCalls = []
const summaryLoad = new Function('request', 'jsonInit', 'hardwareHealthCounts', `${summaryCode}; return loadHardwareSummary`)(async (_, init) => {
  summaryCalls.push(init.body)
  const { host, category } = init.body
  if (category === 'power') throw new Error('node-proxy unavailable')
  return { host: category === 'fans' ? 'wrong-host' : host, category, items: category === 'storage' ? [] : [{ host: 'node1', health: 'OK' }, { host: host || 'node2' }], observed_at: '2026-10-03T00:00:00Z' }
}, (method, body) => ({ method, body }), healthModule.hardwareHealthCounts)
const summaryRows = await summaryLoad(8, 'node1')
assert.equal(summaryCalls.length, 6)
assert.ok(summaryCalls.every((call) => call.cluster_id === 8 && call.host === 'node1'))
assert.equal(summaryRows[0].ok, 1)
assert.equal(summaryRows[0].unknown, 1)
assert.equal(summaryRows[1].total, 0)
for (const row of summaryRows.slice(4)) { assert.equal(row.total, null); assert.ok(row.error) }
assert.equal(summaryRows[0].reported_hosts, 1)
const clusterSummaryRows = await summaryLoad(9, '')
assert.equal(clusterSummaryRows[0].reported_hosts, 2)
assert.equal(clusterSummaryRows[1].reported_hosts, 0)
assert.equal(clusterSummaryRows[4].reported_hosts, null)
assert.ok(summaryCalls.slice(6).every((call) => call.cluster_id === 9 && call.host === ''))
console.log('Hardware summary partial failures and exact request scope checks passed')

const perfSource = readFileSync(new URL('../src/pages/cluster/DaemonPerf.tsx', import.meta.url), 'utf8')
const perfTree = ts.createSourceFile('perf.tsx', perfSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const perfFn = perfTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'DaemonPerf')
const perfLoader = perfFn.body.statements.find((node) => ts.isVariableStatement(node) && node.declarationList.declarations[0].name.getText(perfTree) === 'loader').declarationList.declarations[0].initializer.arguments[0]
const perfCode = ts.transpileModule(`const load = ${perfLoader.getText(perfTree)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const response of [{ daemon_name: 'osd.1', items: [] }, { daemon_name: 'osd.2', items: [] }, { daemon_name: 'osd.1', items: null }]) {
  const calls = []
  const load = new Function('request', 'jsonInit', 'clusterId', 'name', `${perfCode}; return load`)(async (...args) => { calls.push(args); return response }, (method, body) => ({ method, body }), 7, 'osd.1')
  if (response.daemon_name === 'osd.1' && Array.isArray(response.items)) assert.equal(await load(), response)
  else await assert.rejects(load())
  assert.deepEqual(calls, [['/daemon/perf', { method: 'GET', body: { cluster_id: 7, name: 'osd.1' } }]])
}
assert.ok(!perfSource.includes('JSON.parse'))

const rateSource = readFileSync(new URL('../src/pages/cluster/daemonPerfRate.ts', import.meta.url), 'utf8')
const rateCode = ts.transpileModule(rateSource.replace('export function', 'function'), { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS } }).outputText
const rate = new Function('exports', `${rateCode}; return daemonPerfRate`)({})
const counter = (raw_value, type = 10) => ({ name: 'osd.bytes', units: 'bytes', type, raw_value })
assert.equal(rate(counter('9007199254740994'), counter('9007199254740993'), 2000), '0.500000 /s')
assert.equal(rate(counter('18446744073709551615'), counter('0'), 1000), '18446744073709551615.000000 /s')
assert.equal(rate(counter('1'), counter('0'), 3000), '0.333333 /s')
assert.equal(rate(counter('0'), counter('0'), 1000), '0.000000 /s')
assert.match(rate(counter('0'), counter('1'), 1000), /重置/)
assert.match(rate(counter('1'), undefined, 1000), /无基线/)
for (const value of [null, undefined, 1, '-1', '1.5', '{}']) assert.match(rate(counter(value), counter('0'), 1000), /格式无效/)
for (const ms of [0, -1, NaN, Infinity, 0.5]) assert.match(rate(counter('1'), counter('0'), ms), /较新的快照/)
for (const type of [1, 2, 5, 6, 9, 14, 18, null]) assert.match(rate(counter('1', type), counter('0'), 1000), /不适用/)
assert.match(rate(counter('18446744073709551616'), counter('0'), 1000), /超出/)
assert.match(rate(counter('1'), { ...counter('0'), units: 'none' }, 1000), /定义已变化/)
console.log('Native uint64 snapshot rate precision and reset checks passed')

const osdSource = readFileSync(new URL('../src/pages/cluster/OSDInspection.tsx', import.meta.url), 'utf8')
const osdTree = ts.createSourceFile('osd.tsx', osdSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const osdSmartStatusNode = osdTree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'osdSMARTStatus')
const osdSmartStatusCode = ts.transpileModule(osdSmartStatusNode.getText(osdTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const smartStatus = new Function(`${osdSmartStatusCode}; return osdSMARTStatus`)()
assert.equal(smartStatus({ smart_status: { passed: true } }), '设备报告通过')
assert.equal(smartStatus({ smart_status: { passed: false } }), '设备报告未通过')
assert.equal(smartStatus({ error: '', smart_status: { passed: true } }), '读取失败，健康未知')
for (const passed of [null, undefined, 0, 1, 'true']) assert.equal(smartStatus({ smart_status: { passed } }), '未返回明确健康状态')
for (const report of [null, [], 'healthy']) assert.equal(smartStatus(report), '报告无效，健康未知')
const deviceNode = osdTree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'osdDeviceRecords')
const deviceCode = ts.transpileModule(deviceNode.getText(osdTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const deviceRecords = new Function(`${deviceCode}; return osdDeviceRecords`)()
const nativeDevices = [{ devid: 'disk-1', location: [{ host: 'node-a', dev: 'sda', path: '/dev/sda' }], daemons: ['osd.0'], wear_level: 0 }]
assert.equal(deviceRecords({ devices: nativeDevices }), nativeDevices)
assert.deepEqual(deviceRecords({ devices: [] }), [])
for (const devices of [undefined, null, {}, [null], [{}], [{ devid: 0 }], [{ devid: ' ' }]]) assert.equal(deviceRecords({ devices }), null)
const osdNode = osdTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'OSDInspection')
const osdCode = ts.transpileModule(osdNode.getText(osdTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.React } }).outputText
const osdInspection = new Function('React', 'Tabs', 'RecordDetail', 'Diagnostic', 'DaemonPerf', 'Alert', 'Descriptions', 'osdHistoryEpoch', 'OSDNetwork', 'OSDUsage', 'OSDSafetyCheck', 'OSDDeviceClass', 'OSDIndividualFlags', 'OSDDestroy', 'OSDLost', 'OSDPurge', `${osdCode}; return OSDInspection`)(
  { createElement: (component, props) => ({ component, props }) }, 'Tabs', 'RecordDetail', 'Diagnostic', 'DaemonPerf', 'Alert', { Item: 'Item' }, String, 'OSDNetwork', 'OSDUsage', 'OSDSafetyCheck', 'OSDDeviceClass', 'OSDIndividualFlags', 'OSDDestroy', 'OSDLost', 'OSDPurge')
const perfKeys = new Set()
const diagnosticKeys = new Set()
for (const [clusterId, osdId] of [[1, '0'], [1, '12'], [2, '12']]) {
  const panel = osdInspection({ clusterId, osdId, record: {} })
  const safety = panel.props.items.find(item => item.key === 'safety').children
  assert.equal(safety.props.key, `${clusterId}:${osdId}:safety`)
  assert.equal(safety.props.clusterId, clusterId)
  assert.equal(safety.props.osdId, osdId)
  for (const section of ['metadata', 'devices', 'smart', 'histogram']) {
    const diagnostic = panel.props.items.find(item => item.key === section).children
    assert.equal(diagnostic.props.key, `${clusterId}:${osdId}:${section}`)
    diagnosticKeys.add(diagnostic.props.key)
  }
  const devices = panel.props.items.find((item) => item.key === 'devices')
  assert.equal(devices.children.props.clusterId, clusterId)
  assert.equal(devices.children.props.osdId, osdId)
  assert.equal(devices.children.props.section, 'devices')
  const smart = panel.props.items.find((item) => item.key === 'smart')
  assert.equal(smart.children.props.clusterId, clusterId)
  assert.equal(smart.children.props.osdId, osdId)
  assert.equal(smart.children.props.section, 'smart')
  const tab = panel.props.items.find((item) => item.key === 'perf')
  assert.equal(tab.children.component, 'DaemonPerf')
  assert.equal(tab.children.props.clusterId, clusterId)
  assert.equal(tab.children.props.name, `osd.${osdId}`)
  perfKeys.add(tab.children.props.key)
}
assert.equal(perfKeys.size, 3, 'cluster or OSD changes must reset the performance baseline')
assert.equal(diagnosticKeys.size, 12, 'each target and section must own a fresh diagnostic state')
console.log('OSD performance navigation and snapshot scope checks passed')

const gatewaySource = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const gatewayTree = ts.createSourceFile('gateway.tsx', gatewaySource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const gatewayNode = gatewayTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'GatewayManagementPage')
const gatewayCode = ts.transpileModule(gatewayNode.getText(gatewayTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.React } }).outputText
for (const clusterId of [undefined, 1, 2]) {
  const render = new Function('React', 'useClusterContext', 'ResourceListPage', 'definitions', `${gatewayCode}; return GatewayManagementPage`)(
    { createElement: (component, props) => ({ component, props }) }, () => ({ selectedClusterId: clusterId }), 'ResourceListPage', { gatewayManagement: {} })
  assert.equal(render().props.key, clusterId ?? 'none')
}
assert.ok(gatewaySource.includes('<ServiceDaemons key={`${clusterId}:${row.name}`} clusterId={clusterId} name={row.name} />'))
console.log('RGW gateway detail cluster isolation checks passed')
let gatewayDefinition
function findGatewayDefinition(node) {
  if (ts.isPropertyAssignment(node) && node.name.getText(gatewayTree) === 'gatewayManagement') gatewayDefinition = node.initializer
  ts.forEachChild(node, findGatewayDefinition)
}
findGatewayDefinition(gatewayTree)
const gatewayColumnsNode = gatewayDefinition.properties.find((node) => node.name.getText(gatewayTree) === 'columns').initializer
const gatewayColumnsCode = ts.transpileModule(`const columns = ${gatewayColumnsNode.getText(gatewayTree)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const gatewayColumns = new Function(`${gatewayColumnsCode}; return columns`)()
const gatewayKeys = gatewayColumns.map((column) => column.key)
for (const key of ['name', 'running', 'size', 'unmanaged', 'last_refresh', 'networks', 'ports', 'service_url', 'virtual_ip', 'container_image_name']) assert.ok(gatewayKeys.includes(key))
for (const key of ['service_name', 'status']) assert.ok(!gatewayKeys.includes(key), `service inventory does not provide ${key}`)
const managementMode = gatewayColumns.find((column) => column.key === 'unmanaged').render
assert.equal(managementMode(false), '编排器管理')
assert.equal(managementMode(true), '非托管')
assert.equal(managementMode(null), '未采集')

const logsSource = readFileSync(new URL('../src/pages/monitoring/RuntimeLogsPage.tsx', import.meta.url), 'utf8')
const logsTree = ts.createSourceFile('logs.tsx', logsSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const highlightNode = logsTree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'runtimeLogHighlights')
const highlightExports = {}
new Function('exports', ts.transpileModule(highlightNode.getText(logsTree), { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS } }).outputText)(highlightExports)
const highlightLogs = highlightExports.runtimeLogHighlights
assert.deepEqual(highlightLogs('Error: ERROR', 'error'), [
  { text: 'Error', match: true }, { text: ': ', match: false }, { text: 'ERROR', match: true }
])
for (const search of ['[ERR]', '.*', '\\', 'a+b?', '${x}', '(x)', '^$']) {
  assert.deepEqual(highlightLogs(`before ${search} after`, search), [
    { text: 'before ', match: false }, { text: search, match: true }, { text: ' after', match: false }
  ])
}
for (const [text, search] of [['中文\n中文', '中文'], ['😀😀', '😀'], ['İ error', 'error'], ['<script>alert(1)</script>', 'script'], ['abc', 'missing'], ['', 'x'], ['abc', '']]) {
  assert.equal(highlightLogs(text, search).map(part => part.text).join(''), text)
}
assert.deepEqual(highlightLogs('İ error', 'error'), [{ text: 'İ ', match: false }, { text: 'error', match: true }])
assert.deepEqual(highlightLogs(null, 'x'), [{ text: '', match: false }])
assert.ok(logsSource.includes('<mark key={index}>{part.text}</mark>'))
assert.ok(!logsSource.includes('dangerouslySetInnerHTML'))
const visibleLogsNode = logsTree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'visibleRuntimeLogs')
const visibleLogsExports = {}
new Function('exports', ts.transpileModule(visibleLogsNode.getText(logsTree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(visibleLogsExports)
const oldLogScope = JSON.stringify([1, 'cluster', 'debug', 100])
const oldLogSnapshot = { scope: oldLogScope, rows: [{ message: 'old' }], observed: '2026-10-06', error: 'old error' }
assert.equal(visibleLogsExports.visibleRuntimeLogs(oldLogSnapshot, oldLogScope), oldLogSnapshot)
for (const scope of [[2, 'cluster', 'debug', 100], [1, 'audit', 'debug', 100], [1, 'cluster', 'info', 100], [1, 'cluster', 'debug', 30]].map(JSON.stringify)) {
  assert.deepEqual(visibleLogsExports.visibleRuntimeLogs(oldLogSnapshot, scope), { scope, rows: [], error: '' })
}
assert.deepEqual(visibleLogsExports.visibleRuntimeLogs(null, oldLogScope).rows, [])
assert.ok(logsSource.includes('const { rows, observed, error } = visibleRuntimeLogs(snapshot, scope)'))
const logKeyNode = logsTree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'runtimeLogKey')
const keyExports = {}
new Function('exports', ts.transpileModule(logKeyNode.getText(logsTree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(keyExports)
const keyRow = { channel: 'audit', name: 'client/a', rank: 'b', stamp: '2026-10-06', seq: '9007199254740993' }
assert.notEqual(keyExports.runtimeLogKey(keyRow), keyExports.runtimeLogKey({ ...keyRow, name: 'client', rank: 'a/b' }))
assert.notEqual(keyExports.runtimeLogKey(keyRow), keyExports.runtimeLogKey({ ...keyRow, seq: '9007199254740992' }))
assert.equal(keyExports.runtimeLogKey(keyRow), keyExports.runtimeLogKey({ ...keyRow, message: 'updated text' }))
assert.ok(logsSource.includes('rowKey={runtimeLogKey}'))
const logDetailsNode = logsTree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'runtimeLogDetails')
const detailExports = {}
new Function('exports', ts.transpileModule(logDetailsNode.getText(logsTree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(detailExports)
const logDetails = detailExports.runtimeLogDetails({ rank: '0', seq: '18446744073709551615', message: '', addrs: { addrvec: [{ addr: '[::1]:3300', type: 'v2' }] } })
assert.equal(logDetails.find(item => item.key === 'seq').children, '18446744073709551615')
assert.equal(logDetails.find(item => item.key === 'rank').children, '0')
assert.equal(logDetails.find(item => item.key === 'message').children, '')
assert.equal(logDetails.find(item => item.key === 'channel').children, '未提供')
assert.ok(logDetails.find(item => item.key === 'addrs').children.includes('[::1]:3300'))
assert.ok(logsSource.includes('items={runtimeLogDetails(row)}'))
const logCopyFunctions = logsTree.statements.filter(node => ts.isFunctionDeclaration(node) && ['runtimeLogsText', 'copyRuntimeLogs'].includes(node.name.text))
const logCopyCode = ts.transpileModule(logCopyFunctions.map(node => node.getText(logsTree).replace('export ', '')).join('\n'), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const copied = []
const logCopy = new Function('navigator', `${logCopyCode}; return { runtimeLogsText, copyRuntimeLogs }`)({ clipboard: { writeText: async text => copied.push(text) } })
const copyRows = [{ stamp: '2026-10-06', channel: 'audit', priority: '[INF]', name: 'mon.a', message: '中文\nline two' }]
assert.equal(logCopy.runtimeLogsText(copyRows), '2026-10-06 [audit] [INF] mon.a: 中文\nline two')
await logCopy.copyRuntimeLogs(copyRows)
assert.deepEqual(copied, [logCopy.runtimeLogsText(copyRows)])
await assert.rejects(new Function('navigator', `${logCopyCode}; return copyRuntimeLogs`)({})(copyRows), /不支持剪贴板/)
await assert.rejects(new Function('navigator', `${logCopyCode}; return copyRuntimeLogs`)({ clipboard: { writeText: async () => { throw new Error('permission denied') } } })(copyRows), /permission denied/)
assert.ok(logsSource.includes('copyRuntimeLogs(filtered)'))
assert.ok(logsSource.includes('const content = runtimeLogsText(filtered)'))
const logsPanelNode = logsTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'RuntimeLogsPanel')
const logsPanelCode = ts.transpileModule(logsPanelNode.getText(logsTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.React } }).outputText
for (const id of [undefined, 1, 2]) {
  const panel = new Function('useClusterContext', 'RuntimeLogsContent', 'React', `${logsPanelCode}; return RuntimeLogsPanel`)(() => ({ selectedClusterId: id }), 'logs-content', { createElement: (component, props) => ({ component, props }) })
  const result = panel({ compact: true })
  assert.equal(result.props.key, id ?? 'none')
  assert.equal(result.props.selectedClusterId, id)
  assert.equal(result.props.compact, true)
}
assert.ok(logsSource.includes('return () => { abort.abort(); clearTimeout(timer) }'))
const logFilterNode = logsTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'runtimeLogMatches')
const logFilterCode = ts.transpileModule(logFilterNode.getText(logsTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const matchesLog = new Function(`${logFilterCode}; return runtimeLogMatches`)()
const log = { stamp: '2026-10-03T08:00:00Z', message: 'healthy', channel: 'audit', priority: '[INF]' }
assert.equal(matchesLog(log, 'AUDIT', '', ''), true)
assert.equal(matchesLog(log, 'missing', '', ''), false)
assert.equal(matchesLog(log, '', '2026-10-03T16:00:00+08:00', '2026-10-03T08:00:00Z'), true)
assert.equal(matchesLog(log, '', '2026-10-03T08:00:01Z', ''), false)
assert.equal(matchesLog(log, '', '', '2026-10-03T07:59:59Z'), false)
assert.equal(matchesLog(log, '', '2026-10-04T00:00:00Z', '2026-10-02T00:00:00Z'), false)
assert.equal(matchesLog({ stamp: 'unknown' }, '', '2026-10-03T00:00:00Z', ''), false)
assert.equal(matchesLog({ stamp: 'unknown' }, '', '', ''), true)
assert.equal(matchesLog(log, '', '', '', '[INF]'), true)
assert.equal(matchesLog(log, '', '', '', '[ERR]'), false)
assert.equal(matchesLog({ ...log, message: '[ERR] in text' }, '[ERR]', '', '', '[ERR]'), false)
assert.equal(matchesLog({ ...log, priority: '[NEW]' }, '', '', '', ''), true)
assert.equal(matchesLog(log, '', '2026-10-04T00:00:00Z', '', '[INF]'), false)
assert.ok(logsSource.includes('const content = runtimeLogsText(filtered)'))
console.log('Runtime log time bounds and filtered download checks passed')

const serviceSource = readFileSync(new URL('../src/pages/cluster/ServiceDaemons.tsx', import.meta.url), 'utf8')
const serviceTree = ts.createSourceFile('service.tsx', serviceSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const detailFormatter = serviceTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'daemonDetailText')
const detailCode = ts.transpileModule(detailFormatter.getText(serviceTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const formatDetail = new Function(`${detailCode}; return daemonDetailText`)()
for (const [value, expected] of [[null, '未返回'], [undefined, '未返回'], [false, '否'], [true, '是'], [0, '0'], ['18446744073709551615', '18446744073709551615']]) assert.equal(formatDetail(value), expected)
assert.equal(formatDetail({ pending: false }), '{\n  "pending": false\n}')
for (const key of ['memory_request', 'container_id', 'container_image_id', 'container_image_digests', 'ip', 'ports', 'systemd_unit', 'created', 'started', 'last_deployed', 'last_configured', 'pending_daemon_config']) assert.ok(serviceSource.includes(`['${key}',`))
assert.ok(serviceSource.includes('<summary>展开运行详情</summary>'))
assert.ok(serviceSource.includes('<DaemonRuntimeDetails row={row} />'))
assert.ok(hostDetailTree.text.includes('<DaemonRuntimeDetails row={{ ...row, daemon_type: row.type_display }} />'))
for (const key of ['rank', 'rank_generation']) assert.ok(serviceSource.includes(`['${key}',`))
const serviceFn = serviceTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'ServiceDaemons')
assert.ok(serviceSource.includes('perf?.clusterId === clusterId && perf.service === name'))
assert.ok(serviceSource.includes('daemon: String(row.daemon_name)'))
assert.ok(serviceSource.includes('key={`${visiblePerf.clusterId}:${visiblePerf.service}:${visiblePerf.daemon}`}'))
const loaderDeclaration = serviceFn.body.statements.find((node) => ts.isVariableStatement(node) && node.declarationList.declarations[0].name.getText(serviceTree) === 'loader')
const loaderArrow = loaderDeclaration.declarationList.declarations[0].initializer.arguments[0]
const serviceLoaderCode = ts.transpileModule(`const load = ${loaderArrow.getText(serviceTree)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const response of [{ service_name: 'rgw.a', items: [] }, { service_name: 'rgw.b', items: [] }, { service_name: 'rgw.a', items: null }]) {
  const requests = []
  const load = new Function('request', 'jsonInit', 'clusterId', 'name', `${serviceLoaderCode}; return load`)(async (...args) => { requests.push(args); return response }, (method, body) => ({ method, body }), 9, 'rgw.a')
  if (response.service_name === 'rgw.a' && Array.isArray(response.items)) assert.equal(await load(), response)
  else await assert.rejects(load())
  assert.deepEqual(requests, [['/service/daemons', { method: 'GET', body: { cluster_id: 9, name: 'rgw.a' } }]])
}
const servicePage = readFileSync(new URL('../src/pages/cluster/ServicePage.tsx', import.meta.url), 'utf8')
assert.ok(servicePage.includes('detail?.clusterId === selectedClusterId'))
assert.ok(servicePage.includes('listAllResources(\'/services\''))
assert.ok(servicePage.includes('key={`${visibleDetail.clusterId}:${visibleDetail.name}`}'))
assert.ok(readFileSync(new URL('../src/pages/index.ts', import.meta.url), 'utf8').includes('serviceManagement: ServicePage'))
console.log('Service daemon request and navigation checks passed')

const servicePageTree = ts.createSourceFile('ServicePage.tsx', servicePage, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const identityCode = ts.transpileModule(servicePageTree.statements.filter((node) => ts.isFunctionDeclaration(node) && ['serviceName', 'serviceType', 'serviceId'].includes(node.name.text)).map((node) => node.getText(servicePageTree)).join('\n'), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const serviceId = new Function('textValue', `${identityCode}; return serviceId`)((value, fallback) => typeof value === 'string' ? value : fallback)
for (const [name, type, id] of [['mon', 'mon', ''], ['rgw.foo', 'rgw', 'foo'], ['rgw.realm.zone', 'rgw', 'realm.zone'], ['mds.fs', 'mds', 'fs']]) assert.equal(serviceId({ name, type }), id)
assert.ok(servicePage.includes('<Select disabled={Boolean(editingService)} options={serviceTypeOptions}'))
console.log('Service edit identity checks passed')
assert.ok(servicePage.includes("['mds', 'rgw', 'nfs', 'smb'].includes(selectedServiceType)"))
assert.ok(servicePage.includes("onChange={() => form.setFieldValue('service_id', undefined)}"))
assert.ok(servicePage.includes('name="unmanaged" label="非托管" valuePropName="checked"'))
assert.ok(servicePage.includes("typeof values.unmanaged === 'boolean' ? { unmanaged: values.unmanaged } : {}"))
assert.ok(servicePage.includes("unmanaged: typeof row.unmanaged === 'boolean' ? row.unmanaged : undefined"))
assert.ok(servicePage.includes('name="networks" label="绑定网段"'))
assert.ok(servicePage.includes('Array.isArray(values.networks) ? { networks: values.networks } : {}'))

for (const key of ['unmanaged', 'last_refresh', 'ports', 'events', 'service_url', 'virtual_ip', 'container_image_name', 'container_image_id', 'ceph_created_at']) assert.ok(servicePage.includes(`key: '${key}'`))
assert.ok(servicePage.includes('serviceMeta: services, daemonMeta: daemons'))
assert.ok(servicePage.includes('data?.serviceMeta?.stale && <Alert'))
assert.ok(servicePage.includes('data?.daemonMeta?.stale && <Alert'))
console.log('Service runtime metadata and stale inventory bindings passed')

const serviceContent = servicePageTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'ServicePageContent')
const submitServiceNode = serviceContent.body.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'submitService')
const submitServiceCode = ts.transpileModule(submitServiceNode.getText(servicePageTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const scenario of ['ok', 'inactive', 'unmount', 'stale', 'error']) {
  const calls = []
  let resolve, reject
  const env = {
    active: { current: scenario !== 'inactive' }, running: { current: false }, selectedClusterId: 7, loading: false, error: '',
    editingService: { name: 'rgw.a', resource_version: '9007199254740993', stale: scenario === 'stale' }, serviceWritable: (row) => !row.stale,
    setSubmitting: () => {}, setFormOpen: () => calls.push('close'), parsePlacement: JSON.parse, serviceName: (row) => row.name,
    message: { error: () => {}, warning: () => calls.push('warning'), success: () => calls.push('success') }, refreshAfterMutation: async () => calls.push('collect'),
    mutateResource: (...args) => { calls.push(args); return new Promise((yes, no) => { resolve = yes; reject = no }) },
  }
  const submit = new Function(...Object.keys(env), `${submitServiceCode}; return submitService`)(...Object.values(env))
  const pending = submit({ service_type: 'rgw', service_id: 'a', placement_json: '{}' })
  await submit({ service_type: 'rgw', service_id: 'a', placement_json: '{}' })
  if (['inactive', 'stale'].includes(scenario)) { await pending; assert.deepEqual(calls, []); continue }
  assert.deepEqual(calls[0], ['/service', 'PATCH', { cluster_id: 7, name: 'rgw.a', service_type: 'rgw', service_id: 'a', placement: {} }, { ifMatch: '9007199254740993' }])
  assert.equal(calls.length, 1)
  if (scenario === 'unmount') env.active.current = false
  if (scenario === 'error') { reject(new Error('failure')); await assert.rejects(pending) } else { resolve(); await pending }
  assert.equal(calls.includes('success'), scenario === 'ok')
  assert.equal(calls.includes('collect'), ['ok', 'error'].includes(scenario))
  assert.equal(calls.includes('close'), ['ok', 'error'].includes(scenario))
  assert.equal(env.running.current, false)
}
assert.ok(servicePage.includes("<ServicePageContent key={selectedClusterId ?? 'none'}"))
assert.ok(servicePage.includes("if (!active.current) throw new Error('集群已切换，请重新确认删除')"))
assert.ok(!servicePage.includes('window.setTimeout'))
console.log('Service mutation cluster isolation and duplicate submission checks passed')

const deleteServiceNode = serviceContent.body.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'deleteService')
const deleteServiceCode = ts.transpileModule(deleteServiceNode.getText(servicePageTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const scenario of ['ok', 'error', 'unmount', 'refresh-error']) {
  const calls = []
  let dialog
  const active = { current: true }, running = { current: false }
  const env = {
    active, running, selectedClusterId: 7, loading: false, error: '', serviceWritable: () => true,
    serviceName: (row) => row.name, setSubmitting: () => {},
    Modal: { confirm: (options) => { dialog = options; return { destroy: () => calls.push('destroy') } } },
    message: { error: () => {}, warning: () => calls.push('warning'), success: () => calls.push('success') },
    mutateResource: async () => { calls.push('mutate'); if (scenario === 'unmount') active.current = false; if (scenario === 'error') throw new Error('unconfirmed') },
    refreshAfterMutation: async () => { calls.push('refresh'); if (scenario === 'refresh-error') throw new Error('refresh failed') },
  }
  const remove = new Function(...Object.keys(env), `${deleteServiceCode}; return deleteService`)(...Object.values(env))
  await remove({ name: 'rgw.a', resource_version: '9007199254740993' })
  if (['error', 'refresh-error'].includes(scenario)) await assert.rejects(dialog.onOk())
  else await dialog.onOk()
  assert.equal(calls.filter((call) => call === 'mutate').length, 1)
  assert.equal(calls.includes('destroy'), true)
  assert.equal(calls.includes('refresh'), scenario !== 'unmount')
  assert.equal(running.current, false)
}
console.log('Service uncertain deletion recovery checks passed')

const daemonActionNode = serviceContent.body.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'runDaemonAction')
const daemonActionCode = ts.transpileModule(daemonActionNode.getText(servicePageTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const scenario of ['ok', 'error', 'stale', 'switch-before', 'switch-after']) {
  const calls = []
  let dialog
  const active = { current: true }, running = { current: false }
  const env = {
    active, running, selectedClusterId: 7, loading: false, error: '', serviceWritable: () => scenario !== 'stale',
    daemonActions: ['start', 'stop', 'restart', 'redeploy'].map((value) => ({ value, label: value })),
    textValue: (value) => value, setSubmitting: () => {},
    Modal: { confirm: (options) => { dialog = options; return { destroy: () => calls.push('destroy') } } },
    message: { warning: () => calls.push('warning'), success: () => calls.push('success') },
    mutateResource: async (...args) => { calls.push(args); if (scenario === 'switch-after') active.current = false; if (scenario === 'error') throw new Error('failure') },
    refreshAfterMutation: async () => calls.push('refresh'),
  }
  const run = new Function(...Object.keys(env), `${daemonActionCode}; return runDaemonAction`)(...Object.values(env))
  await run({ name: 'osd.1', resource_version: '9007199254740993' }, 'restart')
  if (scenario === 'stale') { assert.equal(dialog, undefined); continue }
  if (scenario === 'switch-before') active.current = false
  if (['error', 'switch-before'].includes(scenario)) await assert.rejects(dialog.onOk())
  else await dialog.onOk()
  if (scenario === 'switch-before') { assert.deepEqual(calls, []); continue }
  assert.deepEqual(calls[0], ['/daemon/action', 'POST', { cluster_id: 7, name: 'osd.1', action: 'restart' }, { ifMatch: '9007199254740993' }])
  assert.equal(calls.includes('refresh'), scenario !== 'switch-after')
  assert.equal(calls.includes('success'), scenario === 'ok')
  assert.equal(calls.includes('destroy'), true)
  assert.equal(running.current, false)
}
console.log('Daemon action confirmation, version, and scope checks passed')

const writableNode = servicePageTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'serviceWritable')
const writableCode = ts.transpileModule(writableNode.getText(servicePageTree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const writable = new Function('serviceName', `${writableCode}; return serviceWritable`)((row) => row.name)
for (const version of [1, '1', '9007199254740993']) assert.equal(writable({ name: 'mgr', stale: false, resource_version: version }), true)
for (const version of [0, -1, 1.5, Number.MAX_SAFE_INTEGER + 1, undefined, '', '0', '1e3']) assert.equal(writable({ name: 'mgr', stale: false, resource_version: version }), false)
assert.equal(writable({ name: 'mgr', stale: true, resource_version: 1 }), false)

const source = readFileSync(new URL('../src/hooks.ts', import.meta.url), 'utf8')
const tree = ts.createSourceFile('hooks.ts', source, ts.ScriptTarget.Latest, true)
const fn = tree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'useResource')
const code = ts.transpileModule(fn.getText(tree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText

// Deterministic hook lifecycle harness: explicitly render, flush effects, and settle requests.
function harness() {
  const slots = [], effects = []
  let cursor = 0, clusterId = 1, writes = 0
  const same = (a, b) => a?.length === b.length && a.every((value, i) => Object.is(value, b[i]))
  const useState = (initial) => {
    const i = cursor++
    slots[i] ??= { value: initial }
    return [slots[i].value, (value) => { writes++; slots[i].value = typeof value === 'function' ? value(slots[i].value) : value }]
  }
  const useRef = (value) => { const i = cursor++; return slots[i] ??= { current: value } }
  const useCallback = (value, deps) => {
    const i = cursor++
    if (!same(slots[i]?.deps, deps)) slots[i] = { value, deps }
    return slots[i].value
  }
  const useEffect = (effect, deps) => {
    const i = cursor++
    if (!same(slots[i]?.deps, deps)) {
      effects.push(() => { slots[i]?.cleanup?.(); slots[i] = { deps, effect, cleanup: effect() } })
    }
  }
  const hook = new Function('useState', 'useRef', 'useCallback', 'useEffect', 'useClusterContext', `${code}; return useResource`)(useState, useRef, useCallback, useEffect, () => ({ selectedClusterId: clusterId }))
  return {
    render(loader, cluster = clusterId) { clusterId = cluster; cursor = 0; return hook(loader) },
    flush() { while (effects.length) effects.shift()() },
    unmount() { slots.forEach((slot) => slot.cleanup?.()) },
    strictReplay() { slots.forEach((slot) => { if (slot.effect) { slot.cleanup?.(); slot.cleanup = slot.effect() } }) },
    get writes() { return writes }
  }
}
function loaderQueue() {
  const pending = []
  return { pending, loader: () => new Promise((resolve, reject) => pending.push({ resolve, reject })) }
}
const settle = async () => { await Promise.resolve(); await Promise.resolve() }

for (const oldFails of [false, true]) {
  const h = harness(), q = loaderQueue()
  h.render(q.loader); h.flush()
  const refresh = h.render(q.loader).refresh({ showLoading: false })
  q.pending[1].resolve('newest'); await refresh
  oldFails ? q.pending[0].reject(new Error('obsolete failure')) : q.pending[0].resolve('obsolete data')
  await settle()
  const result = h.render(q.loader)
  assert.equal(result.data, 'newest'); assert.equal(result.error, ''); assert.equal(result.loading, false)
}
for (const change of ['loader', 'cluster']) for (const oldFails of [false, true]) {
  const h = harness(), old = loaderQueue(), next = change === 'loader' ? loaderQueue() : old
  h.render(old.loader); h.flush(); old.pending[0].resolve('old data'); await settle()
  const obsoleteRefresh = h.render(old.loader).refresh
  const pending = obsoleteRefresh()
  const cluster = change === 'cluster' ? 2 : 1
  let result = h.render(next.loader, cluster)
  assert.equal(result.data, null, 'old resource must disappear before effect execution')
  assert.equal(result.loading, true)
  await obsoleteRefresh()
  assert.equal(old.pending.length, 2, 'obsolete refresh callbacks must not start requests')
  h.flush()
  oldFails ? old.pending[1].reject(new Error('old failure')) : old.pending[1].resolve('old refresh')
  await pending
  next.pending.at(-1).reject(new Error('new failure')); await settle()
  result = h.render(next.loader, cluster)
  assert.equal(result.data, null); assert.equal(result.error, 'new failure'); assert.equal(result.loading, false)
}
{
  const h = harness(), q = loaderQueue()
  h.render(q.loader); h.flush(); q.pending[0].resolve('cached'); await settle()
  const pending = h.render(q.loader).refresh({ showLoading: false })
  assert.equal(h.render(q.loader).data, 'cached')
  assert.equal(h.render(q.loader).loading, false)
  q.pending[1].reject(new Error('refresh failed')); await pending
  assert.equal(h.render(q.loader).data, 'cached')
  assert.equal(h.render(q.loader).error, 'refresh failed')
  const old = h.render(q.loader).refresh()
  const latest = h.render(q.loader).refresh()
  q.pending[3].resolve('latest'); await latest
  q.pending[2].resolve('older'); await old
  assert.equal(h.render(q.loader).data, 'latest')
}
for (const fails of [false, true]) {
  const h = harness(), q = loaderQueue()
  h.render(q.loader); h.flush()
  const refresh = h.render(q.loader).refresh
  const pending = refresh()
  h.unmount()
  const writes = h.writes
  for (const request of q.pending) fails ? request.reject(new Error('unmounted')) : request.resolve('unmounted')
  await pending; await settle(); await refresh()
  assert.equal(h.writes, writes); assert.equal(q.pending.length, 2)
}
{
  const h = harness(), q = loaderQueue()
  h.render(q.loader); h.flush(); h.strictReplay()
  q.pending[0].resolve('discarded strict effect'); await settle()
  assert.equal(h.render(q.loader).data, null)
  q.pending[1].resolve('active strict effect'); await settle()
  assert.equal(h.render(q.loader).data, 'active strict effect')
}
console.log('Resource loading scope and request ordering checks passed')
import './test-host-address-editor.mjs'

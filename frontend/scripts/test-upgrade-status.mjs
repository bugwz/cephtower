import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/cluster/UpgradePage.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('UpgradePage.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
assert.ok(source.includes("UpgradeContent key={selectedClusterId ?? 'none'}"))
const content = tree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'UpgradeContent')
const controlNode = content.body.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'control')
const controlCode = ts.transpileModule(controlNode.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const phase of ['active', 'before', 'during']) {
  const active = { current: phase !== 'before' }
  const writes = []; const ui = []
  const env = { active, needsCollection: false, pending: { action: 'pause', clusterId: 3 }, selectedClusterId: 3,
    record: { data: {}, stale: false, resource_version: 7 }, operation: { loading: false, run: (fn) => fn() },
    upgradeControlAllowed: () => true, controlLabels: { pause: '暂停升级' },
    mutateResource: async (...args) => { writes.push(args); if (phase === 'during') active.current = false },
    refreshResource: async () => {}, message: { success: () => ui.push('success') },
    setPending: () => ui.push('pending'), setRevision: () => ui.push('revision'), setError: () => ui.push('error') }
  const control = new Function(...Object.keys(env), `${controlCode}; return control`)(...Object.values(env))
  await control()
  assert.equal(writes.length, phase === 'before' ? 0 : 1)
  assert.equal(ui.length, phase === 'active' ? 3 : 0)
  if (writes.length) assert.deepEqual(writes[0][2], { cluster_id: 3, action: 'pause' })
}
const functions = tree.statements.filter((node) => ts.isFunctionDeclaration(node) && ['upgradeStatusFields', 'upgradeControlAllowed'].includes(node.name.text))
for (const stage of ['mutation', 'collection']) {
  let locked = false; let pending = true; let error = ''; let writes = 0
  const env = { active: { current: true }, needsCollection: false, pending: { action: 'pause', clusterId: 3 }, selectedClusterId: 3,
    record: { data: {}, stale: false, resource_version: 7 }, operation: { loading: false, run: (fn) => fn() },
    upgradeControlAllowed: () => true, mutateResource: async () => { writes++; if (stage === 'mutation') throw new Error('uncertain') },
    refreshResource: async () => { throw new Error('uncertain') }, setNeedsCollection: (value) => { locked = value },
    setPending: (value) => { pending = value }, setError: (value) => { error = value } }
  const invoke = () => new Function(...Object.keys(env), `${controlCode}; return control`)(...Object.values(env))()
  await invoke()
  assert.equal(locked, true); assert.equal(pending, null); assert.equal(error, 'uncertain')
  env.needsCollection = true
  await invoke()
  assert.equal(writes, 1)
}
assert.ok(source.includes('disabled={needsCollection || loading || operation.loading}'))
assert.ok(source.includes('setNeedsCollection(false)'))
const code = ts.transpileModule(functions.map((fn) => fn.getText(tree)).join('\n'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
const exports = {}
new Function('exports', code)(exports)
const fields = (value) => Object.fromEntries(exports.upgradeStatusFields(value))
for (const action of ['pause', 'resume', 'stop']) {
  assert.equal(exports.upgradeControlAllowed({}, false, action), false)
  assert.equal(exports.upgradeControlAllowed({ in_progress: false, is_paused: false }, false, action), false)
  assert.equal(exports.upgradeControlAllowed({ in_progress: true, is_paused: action === 'resume' }, true, action), false)
  assert.equal(exports.upgradeControlAllowed({ in_progress: true, is_paused: action === 'resume' }, false, action), true)
}
assert.equal(exports.upgradeControlAllowed({ in_progress: true, is_paused: true }, false, 'pause'), false)
assert.equal(exports.upgradeControlAllowed({ in_progress: true, is_paused: false }, false, 'resume'), false)
assert.equal(exports.upgradeControlAllowed({ in_progress: true }, false, 'pause'), false)
assert.equal(exports.upgradeControlAllowed({ in_progress: true }, false, 'resume'), false)
assert.equal(fields({})['升级状态'], '未知')
assert.equal(fields({ in_progress: false })['升级状态'], '未在升级')
assert.equal(fields({ in_progress: true, is_paused: true })['升级状态'], '已暂停')
assert.equal(fields({ in_progress: true, is_paused: false })['升级状态'], '进行中')
assert.equal(fields({ in_progress: true })['升级状态'], '进行中（暂停状态未知）')
const actual = fields({ target_image: 'quay.io/ceph/ceph:v20.2.2', progress: '2/5 daemons upgraded', which: 'all hosts', services_complete: ['mgr', 'mon'], message: 'waiting' })
assert.equal(actual['目标镜像'], 'quay.io/ceph/ceph:v20.2.2')
assert.equal(actual['进度'], '2/5 daemons upgraded')
assert.equal(actual['已完成服务'], 'mgr、mon')
assert.equal(actual['升级范围'], 'all hosts')
assert.equal(actual['状态消息'], 'waiting')
assert.equal(fields({ services_complete: [] })['已完成服务'], '无')
assert.equal(fields({ services_complete: [{}] })['已完成服务'], '未知')
assert.equal(fields({})['目标镜像'], '未提供')
console.log('Native upgrade status display checks passed')

const watchFunction = tree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'watchUpgradeStatus')
const watchCode = ts.transpileModule(watchFunction.getText(tree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
const flush = async () => { await Promise.resolve(); await Promise.resolve(); await Promise.resolve() }
for (const auto of [false, true]) {
  const watcher = {}
  const timers = new Map()
  const calls = []
  const results = []
  const errors = []
  let settle
  new Function('exports', 'getResource', 'setTimeout', 'clearTimeout', watchCode)(watcher,
    (...args) => { calls.push(args); return new Promise((resolve, reject) => { settle = { resolve, reject } }) },
    (callback, delay) => { assert.equal(delay, 10000); timers.set(1, callback); return 1 },
    (id) => timers.delete(id))
  const stop = watcher.watchUpgradeStatus(7, auto, (item) => results.push(item), (error) => errors.push(error))
  assert.equal(calls.length, 1)
  assert.equal(calls[0][0], '/upgrade')
  assert.equal(calls[0][1], 7)
  assert.equal(timers.size, 0, 'must wait for request completion')
  settle.resolve({ item: { stale: false } })
  await flush()
  assert.equal(results.length, 1)
  assert.equal(timers.size, auto ? 1 : 0)
  if (auto) {
    const next = timers.get(1); timers.clear(); void next()
    settle.reject(new Error('offline')); await flush()
    assert.equal(errors.length, 1)
    assert.equal(timers.size, 1, 'must continue polling after failure')
    const pending = timers.get(1); timers.clear(); void pending()
    stop()
    assert.equal(calls.at(-1)[3].signal.aborted, true)
    settle.resolve({ item: { stale: false } }); await flush()
    assert.equal(results.length, 1, 'must ignore results after cleanup')
  } else stop()
  assert.equal(timers.size, 0)
}

const daemonSource = readFileSync(new URL('../src/pages/cluster/UpgradeDaemons.tsx', import.meta.url), 'utf8')
const daemonTree = ts.createSourceFile('UpgradeDaemons.tsx', daemonSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const daemonFunction = daemonTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'upgradeDaemonRows')
const daemonExports = {}
new Function('exports', ts.transpileModule(daemonFunction.getText(daemonTree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(daemonExports)
const daemonItems = [{ natural_key: 'mgr.a', name: 'mgr.a', type: 'mgr', hostname: 'node-a', version: '20.2.2', container_image: 'ceph:v20.2.2', stale: false }, { name: 'mon.b', type: 'mon', version: null, stale: true }]
const daemonRows = daemonExports.upgradeDaemonRows(daemonItems, '')
assert.equal(daemonRows[0].image, 'ceph:v20.2.2')
assert.equal(daemonRows[0].freshness, '有效')
assert.equal(daemonRows[1].version, '未知')
assert.equal(daemonRows[1].freshness, '已过期')
assert.equal(daemonExports.upgradeDaemonRows(daemonItems, ' NODE-A ').length, 1)
assert.equal(daemonExports.upgradeDaemonRows(daemonItems, '20.2').length, 1)
assert.equal(daemonExports.upgradeDaemonRows(daemonItems, 'missing').length, 0)
assert.deepEqual(daemonExports.upgradeDaemonRows([], ''), [])
for (const type of ['mgr', 'mon', 'crash', 'osd', 'mds', 'rgw', 'rbd-mirror', 'cephfs-mirror', 'iscsi', 'nfs']) {
  assert.equal(daemonExports.upgradeDaemonRows([{ name: 'arbitrary', type }], '').length, 1)
}
for (const type of ['prometheus', 'grafana', 'alertmanager', 'node-exporter', 'unknown', undefined, null]) {
  assert.deepEqual(daemonExports.upgradeDaemonRows([{ name: 'osd.1', type }], ''), [])
}

const checkSource = readFileSync(new URL('../src/pages/cluster/UpgradeCheck.tsx', import.meta.url), 'utf8')
const checkTree = ts.createSourceFile('UpgradeCheck.tsx', checkSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const checkFunctions = checkTree.statements.filter((node) => ts.isFunctionDeclaration(node) && ['upgradeCheckVersion', 'upgradeTargetBody', 'upgradeCheckData', 'upgradeStartAllowed'].includes(node.name.text))
const checkExports = {}
new Function('exports', ts.transpileModule(checkFunctions.map((fn) => fn.getText(checkTree)).join('\n'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(checkExports)
assert.equal(checkExports.upgradeCheckVersion(' 20.2.2 '), '20.2.2')
assert.deepEqual(checkExports.upgradeTargetBody('version', '20.2.2'), { version: '20.2.2' })
assert.deepEqual(checkExports.upgradeTargetBody('image', ' registry:5000/ceph@sha256:abc '), { image: 'registry:5000/ceph@sha256:abc' })
for (const image of ['', '--image', 'image name', 'image;cmd']) assert.throws(() => checkExports.upgradeTargetBody('image', image))
const idle = { stale: false, data: { in_progress: false } }
assert.equal(checkExports.upgradeStartAllowed(idle, '20.2.2', '20.2.2'), true)
for (const record of [null, { ...idle, stale: true }, { data: {} }, { ...idle, data: { in_progress: true } }]) assert.equal(checkExports.upgradeStartAllowed(record, '20.2.2', '20.2.2'), false)
assert.equal(checkExports.upgradeStartAllowed(idle, '20.2.2', ''), false)
assert.equal(checkExports.upgradeStartAllowed(idle, '20.2.3', '20.2.2'), false)
for (const version of ['', 'v20.2.2', '--image', '20.2']) assert.throws(() => checkExports.upgradeCheckVersion(version))
const report = { target_name: 'ceph:v20.2.2', target_id: 'abc', target_version: '20.2.2', needs_update: { 'mon.a': { current_name: 'ceph:old', current_id: null, current_version: '19.2.1', ignored: 'not-a-column' } }, up_to_date: [], non_ceph_image_daemons: ['prometheus.a'] }
assert.deepEqual(checkExports.upgradeCheckData({ check: report }).rows, [{ name: 'mon.a', current_name: 'ceph:old', current_id: null, current_version: '19.2.1' }])
assert.deepEqual(checkExports.upgradeCheckData({ check: { ...report, needs_update: {} } }).rows, [])
for (const check of [null, {}, { ...report, target_name: null }, { ...report, up_to_date: null }, { ...report, needs_update: [] }, { ...report, needs_update: { 'mon.a': {} } }]) assert.throws(() => checkExports.upgradeCheckData({ check }))
console.log('Upgrade compatibility report checks passed')

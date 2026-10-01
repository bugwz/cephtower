import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/cluster/UpgradePage.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('UpgradePage.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const functions = tree.statements.filter((node) => ts.isFunctionDeclaration(node) && ['upgradeStatusFields', 'upgradeControlAllowed'].includes(node.name.text))
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

const checkSource = readFileSync(new URL('../src/pages/cluster/UpgradeCheck.tsx', import.meta.url), 'utf8')
const checkTree = ts.createSourceFile('UpgradeCheck.tsx', checkSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const checkFunctions = checkTree.statements.filter((node) => ts.isFunctionDeclaration(node) && ['upgradeCheckVersion', 'upgradeCheckData', 'upgradeStartAllowed'].includes(node.name.text))
const checkExports = {}
new Function('exports', ts.transpileModule(checkFunctions.map((fn) => fn.getText(checkTree)).join('\n'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(checkExports)
assert.equal(checkExports.upgradeCheckVersion(' 20.2.2 '), '20.2.2')
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

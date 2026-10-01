import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/cluster/UpgradePage.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('UpgradePage.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const fn = tree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'upgradeStatusFields')
const code = ts.transpileModule(fn.getText(tree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
const exports = {}
new Function('exports', code)(exports)
const fields = (value) => Object.fromEntries(exports.upgradeStatusFields(value))
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

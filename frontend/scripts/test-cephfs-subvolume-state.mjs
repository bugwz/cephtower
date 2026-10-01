import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/file/cephfsSubvolumeState.ts', import.meta.url), 'utf8')
const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
const exports = {}
new Function('exports', compiled)(exports)
for (const [state, text] of Object.entries({ complete: '可用', init: '初始化中', pending: '等待克隆', 'in-progress': '克隆进行中', failed: '克隆失败', canceled: '克隆已取消', 'snapshot-retained': '已删除，仅保留快照' })) {
  assert.equal(exports.subvolumeState(state).text, text)
  assert.equal(exports.subvolumeReadyReason({ state }) === undefined, state === 'complete')
}
for (const value of [null, undefined, '', true, 1, 'available', 'toString', '__proto__', 'new-state']) {
  assert.equal(exports.subvolumeState(value).color, 'default')
  assert.ok(exports.subvolumeReadyReason({ state: value }))
}
assert.equal(exports.subvolumeType('clone'), '克隆子卷')
assert.equal(exports.subvolumeType('subvolume'), '普通子卷')
assert.equal(exports.subvolumeType(undefined), '未知')
assert.equal(exports.subvolumeType('new-type'), '未知（new-type）')
const pages = readFileSync(new URL('../src/pages/file/pages.tsx', import.meta.url), 'utf8')
const start = pages.indexOf('  subvolumes: {')
const config = pages.slice(start, pages.indexOf('\n  },', start))
for (const field of ['state', 'type', 'pool_namespace']) assert.ok(config.includes(`key: '${field}'`))
assert.match(config, /disabledWhen: subvolumeReadyReason/)
assert.match(config, /detailContent:.*subvolumeReadyReason\(row\).*<SubvolumeSnapshotVisibility/)
const resourcePage = readFileSync(new URL('../src/pages/ResourceListPage.tsx', import.meta.url), 'utf8')
assert.match(resourcePage, /<TableAction disabled=\{mutationBlocked \|\| Boolean\(definition\.updateAction\.disabledWhen\?\.\(row\)\)\}/)
console.log('CephFS subvolume readiness checks passed')

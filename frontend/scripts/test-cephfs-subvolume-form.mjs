import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

function load(url) {
  const compiled = ts.transpileModule(readFileSync(url, 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
  const exports = {}
  new Function('exports', 'require', compiled)(exports, (path) => load(new URL(`${path}.ts`, url)))
  return exports
}
const helpers = load(new URL('../src/pages/file/cephfsSubvolumeForm.ts', import.meta.url))
const group = load(new URL('../src/pages/file/cephfsGroupForm.ts', import.meta.url))
assert.deepEqual(helpers.subvolumeUpdateInitialValues({ bytes_quota: 'infinite' }), { size: undefined, unlimited: true, no_shrink: false })
assert.deepEqual(helpers.subvolumeUpdateInitialValues({ bytes_quota: 2048 }), { size: '2048', unlimited: false, no_shrink: false })
for (const size of ['9007199254740991', '9007199254740993', '9223372036854775807']) {
  assert.equal(helpers.subvolumeUpdateInitialValues({ bytes_quota: size }).size, size)
  assert.equal(helpers.subvolumeUpdateBody({ size }, 1, 'cephfs', 'volume', 'team').size, size)
  assert.equal(group.groupUpdateBody({ size }, 1, 'cephfs', 'team').size, size)
}
for (const value of ['9223372036854775808', Number.MAX_SAFE_INTEGER + 1, null, undefined, true, 0, -1, 1.5, 'Infinity', '2e3']) assert.equal(helpers.subvolumeUpdateInitialValues({ bytes_quota: value }).size, undefined)
assert.deepEqual(helpers.subvolumeUpdateBody({ unlimited: true, size: 123, no_shrink: true }, 1, 'cephfs', 'volume', 'team'), { cluster_id: 1, fs: 'cephfs', subvolume: 'volume', group: 'team', unlimited: true, no_shrink: false })
assert.deepEqual(helpers.subvolumeUpdateBody({ size: '2048', no_shrink: true }, 1, 'cephfs', 'volume', '_nogroup'), { cluster_id: 1, fs: 'cephfs', subvolume: 'volume', group: '_nogroup', size: '2048', no_shrink: true })
for (const size of ['', undefined, null, true, 1024, 0, -1, 1.5, '2e3', '9223372036854775808', Number.MAX_SAFE_INTEGER + 1]) {
  assert.throws(() => helpers.subvolumeUpdateBody({ size }, 1, 'cephfs', 'volume', 'team'), /整数字节数/)
  assert.throws(() => group.groupUpdateBody({ size }, 1, 'cephfs', 'team'), /整数字节数/)
}
const pages = readFileSync(new URL('../src/pages/file/pages.tsx', import.meta.url), 'utf8')
const start = pages.indexOf('  subvolumes: {')
const config = pages.slice(start, pages.indexOf('\n  },', start))
assert.match(config, /initialValues: subvolumeUpdateInitialValues/)
assert.match(config, /subvolumeUpdateBody\(values, clusterId, fsName\(row\), subvolumeName\(row\), groupName\(row\)\)/)
assert.match(config, /name: 'size'.*pattern:.*请输入正整数字节数/)
assert.match(config, /name: 'no_shrink'.*不允许配额低于已用空间.*visibleWhen/)
console.log('CephFS subvolume quota form checks passed')

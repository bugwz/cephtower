import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const exports = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwStorageDetails.ts', import.meta.url), 'utf8'), {
  compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 }
}).outputText)(exports)
const rows = exports.rgwStorageRows
const stats = { size: 0, size_actual: 4096, size_utilized: 7, num_objects: 2 }
assert.deepEqual(rows({ stats }), [{ category: '汇总', size: '0', actual: '4096', utilized: '7', objects: '2' }])
assert.deepEqual(rows({ 'rgw.main': stats, future: stats }, true).map(row => row.category), ['rgw.main', 'future'])
assert.deepEqual(rows({}, true), [])
for (const value of [undefined, null, [], 'bad']) assert.equal(rows(value), undefined)
for (const value of [{}, { stats: [] }, { stats: null }]) assert.equal(rows(value), undefined)
assert.equal(rows({ bad: null }, true), undefined)
for (const size of [undefined, null, -1, 0.5, NaN, Infinity, '1', Number.MAX_SAFE_INTEGER + 1]) {
  assert.equal(rows({ stats: { ...stats, size } })[0].size, '未返回或超出精确显示范围')
}
assert.equal(rows({ stats: { num_objects: 0 } })[0].objects, '0')
assert.equal(rows({ stats: { size_kb: 1 } })[0].size, '未返回或超出精确显示范围')
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.equal(pages.match(/<RgwStorage value=\{value\} \/>/g).length, 2)
assert.equal(pages.match(/<RgwStorage value=\{value\} categorized \/>/g).length, 1)
console.log('RGW storage statistics preserve category, units, zero and unavailable values')

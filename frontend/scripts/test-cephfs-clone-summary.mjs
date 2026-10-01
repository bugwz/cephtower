import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import React from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import ts from 'typescript'

const nativeRequire = createRequire(import.meta.url)
function load(url) {
  const compiled = ts.transpileModule(readFileSync(url, 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText
  const exports = {}
  new Function('exports', 'require', compiled)(exports, (path) => path.startsWith('.') ? load(new URL(`${path}.ts`, url)) : nativeRequire(path))
  return exports
}
const helpers = load(new URL('../src/pages/file/cephfsCloneSummary.ts', import.meta.url))
assert.equal(helpers.cloneSource({ volume: 'cephfs', subvolume: 'source', snapshot: 'snap' }), 'cephfs/_nogroup/source@snap')
assert.equal(helpers.cloneSource({ volume: 'cephfs', group: 'team', subvolume: 'source', snapshot: 'snap' }), 'cephfs/team/source@snap')
assert.equal(helpers.cloneSource('N/A'), '来源不可用')
for (const value of [null, [], {}, { volume: 'cephfs' }, true]) assert.equal(helpers.cloneSource(value), '—')
const report = { 'percentage cloned': '42.5%', 'amount cloned': '4.2K/10.0K', 'files cloned': '21/50' }
assert.deepEqual(helpers.cloneProgress(report), { percent: 42.5, amount: '4.2K/10.0K', files: '21/50' })
assert.equal(helpers.cloneProgress({ 'percentage cloned': '0%' }).percent, 0)
assert.equal(helpers.cloneProgress({ 'percentage cloned': '100%' }).percent, 100)
for (const value of ['101%', '-1%', 'NaN%', 'Infinity%', 42, '42', null, '']) assert.equal(helpers.cloneProgress({ 'percentage cloned': value }).percent, undefined)
assert.equal(helpers.cloneFailure({ errno: '5', error_msg: 'Input/output error' }), 'Input/output error (errno 5)')
assert.equal(helpers.cloneFailure({ errno: 0, error_msg: 'canceled' }), 'canceled (errno 0)')
assert.equal(helpers.cloneFailure({ error_msg: 'failed' }), 'failed')
assert.equal(helpers.cloneFailure({ errno: '-5' }), 'errno -5')
for (const value of [null, [], true, {}, { errno: 'NaN' }]) assert.equal(helpers.cloneFailure(value), '—')
const components = load(new URL('../src/pages/file/CephFSCloneProgress.tsx', import.meta.url))
const render = (report) => renderToStaticMarkup(React.createElement(components.CephFSCloneProgress, { report }))
const html = render(report)
assert.match(html, /42\.50%/)
assert.match(html, /已克隆容量：4\.2K\/10\.0K/)
assert.match(html, /已克隆文件：21\/50/)
assert.equal(render(null), '—')
assert.doesNotMatch(render({ 'files cloned': '0/50' }), /ant-progress/)
assert.match(render({ 'files cloned': '0/50' }), /0\/50/)
const pages = readFileSync(new URL('../src/pages/file/pages.tsx', import.meta.url), 'utf8')
assert.match(pages, /key: 'clone_progress'.*CephFSCloneProgress/)
assert.match(pages, /key: 'source'.*cloneSource\(row.clone_source \?\? value\)/)
assert.match(pages, /key: 'clone_failure'.*cloneFailure/)
console.log('CephFS native clone report checks passed')

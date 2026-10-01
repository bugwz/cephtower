import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import React from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import ts from 'typescript'

const nativeRequire = createRequire(import.meta.url)
function load(url) {
  const source = readFileSync(url, 'utf8')
  const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText
  const exports = {}
  new Function('exports', 'require', compiled)(exports, (path) => path.startsWith('.') ? load(new URL(`${path}.ts`, url)) : nativeRequire(path))
  return exports
}
const summary = load(new URL('../src/pages/file/cephfsSubvolumeSummary.ts', import.meta.url))
assert.equal(summary.cephFSBytes(1024), '1.00 KiB')
assert.equal(summary.cephFSBytes(0), '0 B')
assert.equal(summary.cephFSBytes('9007199254740993'), '8.00 PiB')
for (const value of [null, undefined, true, -1, 1.5, Number.MAX_SAFE_INTEGER + 1, 'undefined', '1e3']) assert.equal(summary.cephFSBytes(value), '—')
assert.equal(summary.cephFSQuota('infinite'), '无限制')
assert.deepEqual(summary.cephFSUsage('infinite', 2048, 'undefined'), { text: '2.00 KiB（无限制）', percent: undefined, barPercent: undefined })
assert.deepEqual(summary.cephFSUsage(4096, 0, '0.00'), { text: '0 B / 4.00 KiB', percent: 0, barPercent: 0 })
assert.equal(summary.cephFSUsage(4096, 2048, '50.00').percent, 50)
const over = summary.cephFSUsage(4096, 8192, '200.00')
assert.equal(over.percent, 200)
assert.equal(over.barPercent, 100)
for (const value of [null, undefined, '', 'undefined', 'NaN', 'Infinity', '50%', -1, true, Infinity]) assert.equal(summary.cephFSUsage(4096, 2048, value).percent, undefined)
assert.deepEqual(summary.cephFSPermissions(0o40755), { octal: '0755', owner: 'rwx', group: 'r-x', others: 'r-x' })
assert.deepEqual(summary.cephFSPermissions(0), { octal: '0000', owner: '---', group: '---', others: '---' })
assert.deepEqual(summary.cephFSPermissions(0o4777), { octal: '4777', owner: 'rws', group: 'rwx', others: 'rwx' })
assert.deepEqual(summary.cephFSPermissions(0o7644), { octal: '7644', owner: 'rwS', group: 'r-S', others: 'r-T' })
for (const value of [null, undefined, 'rwx', -1, 0o200000]) assert.equal(summary.cephFSPermissions(value), undefined)

const components = load(new URL('../src/pages/file/CephFSResourceUsage.tsx', import.meta.url))
const render = (props) => renderToStaticMarkup(React.createElement(components.CephFSUsage, props))
assert.match(render({ quota: 4096, used: 2048, percent: '50.00' }), /50\.00%/)
assert.match(render({ quota: 4096, used: 0, percent: '0.00' }), /0\.00%/)
assert.match(render({ quota: 4096, used: 8192, percent: '200.00' }), /200\.00%/)
const infinite = render({ quota: 'infinite', used: 2048, percent: 'undefined' })
assert.match(infinite, /2\.00 KiB（无限制）/)
assert.doesNotMatch(infinite, /undefined|ant-progress/)
assert.match(renderToStaticMarkup(React.createElement(components.CephFSPermissions, { mode: 0o40755 })), /rwxr-xr-x \(0755\)/)
const pages = readFileSync(new URL('../src/pages/file/pages.tsx', import.meta.url), 'utf8')
for (const name of ['subvolumeGroups', 'subvolumes']) {
  const start = pages.indexOf(`  ${name}: {`)
  const config = pages.slice(start, pages.indexOf('\n  },', start))
  for (const renderer of ['cephFSQuota', 'cephFSBytes', 'CephFSUsage', 'CephFSPermissions']) assert.ok(config.includes(renderer), `${name} missing ${renderer}`)
}
console.log('CephFS quota usage and permissions checks passed')

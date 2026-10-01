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
const helpers = load(new URL('../src/pages/file/cephfsSnapshotDependencies.ts', import.meta.url))
const row = { has_pending_clones: 'yes', pending_clones: [{ name: 'default-clone' }, { name: 'team-clone', target_group: 'team' }], orphan_clones_count: 2 }
assert.deepEqual(helpers.snapshotDependencies(row), { pending: 'yes', clones: [{ name: 'default-clone', group: '_nogroup' }, { name: 'team-clone', group: 'team' }], orphans: 2 })
assert.ok(helpers.snapshotDeleteReason(row))
assert.deepEqual(helpers.snapshotDependencies({ has_pending_clones: 'no' }), { pending: 'no', clones: [], orphans: 0 })
assert.equal(helpers.snapshotDeleteReason({ has_pending_clones: 'no' }), undefined)
assert.ok(helpers.snapshotDeleteReason({ has_pending_clones: 'no', orphan_clones_count: 1 }))
assert.ok(helpers.snapshotDeleteReason({ pending_clones: [{ name: 'clone' }] }))
assert.deepEqual(helpers.snapshotDependencies({}), { pending: 'unknown', clones: undefined, orphans: undefined })
assert.equal(helpers.snapshotPendingText(undefined), '未知')
assert.equal(helpers.snapshotDependencies({ has_pending_clones: true }).pending, 'unknown')
for (const value of [null, 'yes', {}, [{ name: '' }], [{ name: 'clone', target_group: 123 }], [null], [true]]) assert.equal(helpers.snapshotDependencies({ has_pending_clones: 'yes', pending_clones: value }).clones, undefined)
for (const value of [null, -1, 1.5, '2', true, Number.MAX_SAFE_INTEGER + 1]) assert.equal(helpers.snapshotDependencies({ orphan_clones_count: value }).orphans, undefined)
const components = load(new URL('../src/pages/file/SnapshotCloneDependenciesPanel.tsx', import.meta.url))
const render = (row) => renderToStaticMarkup(React.createElement(components.SnapshotCloneDependenciesPanel, { row }))
const html = render(row)
for (const text of ['当前不能删除此快照', '_nogroup', 'default-clone', 'team-clone', '检测到孤儿克隆记录']) assert.ok(html.includes(text))
assert.match(render({}), /未知不表示没有克隆依赖/)
assert.match(render({ has_pending_clones: 'no' }), /无待处理克隆目标/)
assert.doesNotMatch(render({ has_pending_clones: 'no' }), /当前不能删除此快照/)
const pages = readFileSync(new URL('../src/pages/file/pages.tsx', import.meta.url), 'utf8')
const start = pages.indexOf('  cephfsSnapshots: {')
const config = pages.slice(start, pages.indexOf('\n  },', start))
assert.match(config, /detailContent:.*SnapshotCloneDependenciesPanel/)
assert.match(config, /disabledWhen: snapshotDeleteReason/)
assert.match(config, /key: 'has_pending_clones'.*snapshotPendingText/)
assert.match(config, /key: 'orphan_clones_count'/)
console.log('CephFS snapshot clone dependency checks passed')

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/ServiceInventoryDetails.tsx', import.meta.url), 'utf8')
const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.React } }).outputText
const exports = {}
new Function('exports', 'require', 'React', code)(exports, () => ({ Alert: 'Alert', Descriptions: 'Descriptions' }), { createElement: (type, props, ...children) => ({ type, props, children }) })
for (const stale of [false, true, undefined]) {
  const result = exports.ServiceInventoryDetails({ row: { stale, type: 'mds', unmanaged: false, networks: ['10.0.0.0/24'], ports: [], container_image_id: 'sha256:123', events: ['deployment failed', 'retry scheduled'] } })
  const items = result.children[2].props.items
  const text = key => items.find(item => item.key === key).children.children[0]
  assert.equal(text('type'), 'mds')
  assert.equal(text('unmanaged'), '否')
  assert.match(text('networks'), /10.0.0.0\/24/)
  assert.equal(text('ports'), '本次未报告条目')
  assert.equal(text('container_image_id'), 'sha256:123')
  assert.equal(text('last_refresh'), '未返回')
  assert.match(text('events'), /deployment failed/)
  assert.match(text('events'), /retry scheduled/)
  assert.equal(JSON.stringify(result).includes('库存已过期'), stale !== false)
}
const page = readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('page.tsx', page, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const mds = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'MdsManagementPage').getText(tree)
assert.ok(mds.includes('<ServiceInventoryDetails row={row} />'))
console.log('MDS service details preserve deployment fields, empty events and stale snapshots')

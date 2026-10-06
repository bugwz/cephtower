import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/monitoring/AlertRoutingDetails.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('AlertRoutingDetails.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const node = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'alertRoutingValues')
const exports = {}
new Function('exports', ts.transpileModule(node.getText(tree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const values = exports.alertRoutingValues
assert.deepEqual(values(['silence-1', 'silence-2']), ['silence-1', 'silence-2'])
assert.deepEqual(values(['same', 'same']), ['same', 'same'])
assert.deepEqual(values([{ name: 'email' }, { name: '<script>' }], true), ['email', '<script>'])
for (const mode of [false, true]) {
  assert.deepEqual(values([], mode), [])
  for (const invalid of [null, undefined, {}, '', [null], [[]], [0], [''], [' ']]) assert.equal(values(invalid, mode), null)
}
assert.equal(values(['legacy-string'], true), null)
assert.equal(values([{ name: 'ok' }, { name: 0 }], true), null)
assert.equal(values(['ok', {}]), null)
const page = readFileSync(new URL('../src/pages/monitoring/pages.tsx', import.meta.url), 'utf8')
assert.ok(page.includes('detailContent: row => <AlertRoutingDetails row={row} />'))
for (const field of ['status.silencedBy', 'status.inhibitedBy', 'row.receivers']) assert.ok(source.includes(field))
assert.ok(source.includes('copyable'))
assert.ok(!source.includes('dangerouslySetInnerHTML'))
console.log('Alert routing details distinguish missing associations from empty native lists')

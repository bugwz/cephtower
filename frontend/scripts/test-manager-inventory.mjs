import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/ManagerInventory.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('inventory.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const fn = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'managerServices')
const exports = {}
new Function('exports', ts.transpileModule(fn.getText(tree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
assert.deepEqual(exports.managerServices({}), [])
assert.deepEqual(exports.managerServices({ dashboard: 'https://host:8443/', empty: '', unsafe: 'javascript:alert(1)' }), [{ name: 'dashboard', uri: 'https://host:8443/' }, { name: 'empty', uri: '' }, { name: 'unsafe', uri: 'javascript:alert(1)' }])
for (const value of [undefined, null, [], '', { a: null }, { a: 7 }]) assert.equal(exports.managerServices(value), null)
assert.ok(!source.includes('href='))
assert.ok(source.includes("listAllResources('/managers', clusterId)"))
assert.ok(source.includes('row.active !== true'))
assert.ok(readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8').includes('<ManagerInventory key={selectedClusterId} clusterId={selectedClusterId} />'))
console.log('Manager inventory preserves service URIs as inert text and scopes active services')

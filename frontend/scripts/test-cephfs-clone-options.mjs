import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/file/pages.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('pages.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const names = new Set(['cloneTargetGroupOptions', 'resourceName', 'fsName'])
const functions = tree.statements.filter((node) => ts.isFunctionDeclaration(node) && names.has(node.name?.text))
assert.equal(functions.length, 3)
const compiled = ts.transpileModule(functions.map((node) => node.getText(tree)).join('\n'), {
  compilerOptions: { module: ts.ModuleKind.CommonJS }
}).outputText
const calls = []
const loader = new Function('listAllResources', `${compiled}; return cloneTargetGroupOptions`)(async (...args) => {
  calls.push(args)
  return { items: [{ name: 'team' }, { name: '_nogroup' }, { name: 'team' }] }
})
assert.deepEqual(await loader(42, { fs: 'cephfs-a' }), [
  { label: '默认组（_nogroup）', value: '_nogroup' }, { label: 'team', value: 'team' }
])
assert.deepEqual(calls, [['/filesystem/subvolume/groups', 42, { body: { fs: 'cephfs-a' } }]])
await assert.rejects(loader(42), /缺少源文件系统/)
assert.equal(calls.length, 1)
const failing = new Function('listAllResources', `${compiled}; return cloneTargetGroupOptions`)(async () => { throw new Error('offline') })
await assert.rejects(failing(42, { fs: 'cephfs-a' }), /offline/)
console.log('CephFS clone target group scope checks passed')

const dependentSource = readFileSync(new URL('../src/pages/dependentFormFields.ts', import.meta.url), 'utf8')
const dependentExports = {}
new Function('exports', ts.transpileModule(dependentSource, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(dependentExports)
const fields = [{ name: 'subvolume', optionsDependencies: ['group'] }, { name: 'group', optionsDependencies: ['fs'] }, { name: 'fs' }, { name: 'name' }]
assert.deepEqual(dependentExports.dependentFormFields(fields, ['fs']), ['group', 'subvolume'])
assert.deepEqual(dependentExports.dependentFormFields(fields, ['group']), ['subvolume'])
assert.deepEqual(dependentExports.dependentFormFields(fields, ['name']), [])
const snapshotFunctions = tree.statements.filter((node) => ts.isFunctionDeclaration(node) && ['snapshotSubvolumeOptions', 'resourceName'].includes(node.name?.text))
const snapshotCode = ts.transpileModule(snapshotFunctions.map((node) => node.getText(tree)).join('\n'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
const scopedCalls = []
const snapshotLoader = new Function('listAllResources', 'subvolumeReadyReason', `${snapshotCode}; return snapshotSubvolumeOptions`)(async (...args) => {
  scopedCalls.push(args)
  return { items: [{ name: 'same-name', state: 'complete' }, { name: 'pending-clone', state: 'pending' }] }
}, (row) => row.state === 'complete' ? undefined : 'unavailable')
assert.deepEqual(await snapshotLoader(7, undefined, { fs: 'a' }), [])
assert.equal(scopedCalls.length, 0)
for (const group of ['_nogroup', 'team']) {
  assert.deepEqual(await snapshotLoader(7, undefined, { fs: 'a', group }), [{ label: 'same-name', value: 'same-name' }])
  assert.deepEqual(scopedCalls.at(-1), ['/filesystem/subvolumes', 7, { body: { fs: 'a', group } }])
}
console.log('CephFS snapshot dependent selection checks passed')

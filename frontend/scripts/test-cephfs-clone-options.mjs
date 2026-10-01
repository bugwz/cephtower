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

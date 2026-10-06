import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const exports = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/cluster/monQuorumGroups.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const rows = [{ name: 'a', in_quorum: true }, { name: 'b', in_quorum: false }, ...[undefined, null, 0, 'false', 'true'].map((in_quorum, i) => ({ name: `unknown-${i}`, in_quorum }))]
const groups = exports.monQuorumGroups(rows)
assert.deepEqual(groups.map(group => group.rows.length), [1, 1, 5])
assert.equal(groups[0].rows[0], rows[0])
assert.equal(groups[1].rows[0], rows[1])
assert.deepEqual(exports.monQuorumGroups([]).map(group => group.rows), [[], [], []])
assert.equal(rows.length, 7)
const source = readFileSync(new URL('../src/api/resource.ts', import.meta.url), 'utf8')
const tree = ts.createSourceFile('resource.ts', source, ts.ScriptTarget.Latest, true)
const code = ts.transpileModule(tree.statements.filter(node => ts.isFunctionDeclaration(node) && ['listAllResources', 'listMonitors'].includes(node.name?.text)).map(node => node.getText(tree).replace('export ', '')).join('\n'), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const calls = []
const load = new Function('listResource', `${code}; return listMonitors`)(async (path, cluster, options) => {
  calls.push({ path, cluster, options })
  return { items: [options.cursor ? rows[1] : rows[0]], nextCursor: options.cursor ? null : 'next' }
})
assert.deepEqual(await load(42, { name: ['a', 'b'] }), rows.slice(0, 2))
assert.equal(calls.length, 2)
for (const call of calls) {
  assert.equal(call.path, '/monitors')
  assert.equal(call.cluster, 42)
  assert.deepEqual(call.options.filters, { name: ['a', 'b'] })
}
const page = readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8')
assert.ok(page.includes('monQuorumGroups(data?.mons ?? [])'))
assert.ok(page.includes('data={group.rows}'))
assert.ok(page.includes('数量为当前筛选结果'))
console.log('MON quorum groups preserve unknown states and load every filtered page')

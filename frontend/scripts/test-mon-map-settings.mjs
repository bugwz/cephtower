import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/MonMapSettings.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('map.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const node = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'monMapSetting')
const exports = {}
new Function('exports', ts.transpileModule(node.getText(tree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
for (const [value, expected] of [[false, '否'], [true, '是'], [0, '0'], ['', '空（原生值）'], ['[a,b]', '[a,b]'], ['squid', 'squid']]) assert.equal(exports.monMapSetting(value), expected)
for (const value of [undefined, null, [], {}, NaN, Infinity, 1.5]) assert.equal(exports.monMapSetting(value), '未返回或格式无效')
for (const field of ['min_mon_release', 'min_mon_release_name', 'election_strategy', 'stretch_mode', 'tiebreaker_mon', 'disallowed_leaders', 'removed_ranks']) assert.ok(source.includes(`value?.${field}`))
assert.ok(readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8').includes('<MonMapSettings value={data?.status} />'))
console.log('MON map settings preserve false, zero and native empty strings')

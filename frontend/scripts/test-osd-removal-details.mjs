import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/OSDRemovalDetails.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('details.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const compile = node => ts.transpileModule(node.getText(tree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
function helper(name) {
  const node = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === name)
  return new Function(`${compile(node)}; return ${name}`)()
}
const bool = helper('removalBoolean'), weight = helper('removalWeight'), time = helper('removalTimestamp')
assert.equal(bool(true), '是'); assert.equal(bool(false), '否')
for (const value of [null, undefined, 0, 1, 'false', 'true', {}]) assert.equal(bool(value), '未知')
assert.equal(weight(0), '0'); assert.equal(weight(1.25), '1.25')
for (const value of [null, undefined, -1, NaN, Infinity, '0']) assert.equal(weight(value), '未采集或格式无效')
assert.equal(time(null), '尚无时间记录')
assert.equal(time('2020-09-14T11:41:53.960463Z'), '2020-09-14T11:41:53.960463Z')
for (const value of [undefined, 0, '', ' ']) assert.equal(time(value), '未采集或格式无效')
const arrays = tree.statements.filter(ts.isVariableStatement).map(compile).join('\n')
const fields = new Function(`${arrays}; return [removalBooleanFields, removalTimeFields]`)()
assert.deepEqual(fields[0].map(([key]) => key), ['started', 'draining', 'stopped', 'replace', 'replace_block', 'replace_db', 'replace_wal', 'force', 'zap'])
assert.deepEqual(fields[1].map(([key]) => key), ['process_started_at', 'drain_started_at', 'drain_stopped_at', 'drain_done_at'])
const pages = readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8')
assert.ok(pages.includes("{ key: 'zap', title: '清盘', render: removalBoolean }"))
assert.ok(pages.includes('key={`${selectedClusterId}:${row.osd_id}:details`}'))

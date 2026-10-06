import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/OSDCapacityThresholds.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('thresholds.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const helper = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'osdCapacityThreshold')
const code = ts.transpileModule(helper.getText(tree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const format = new Function(`${code}; return osdCapacityThreshold`)()
assert.equal(format(0.85), '0.85（约 85%）')
assert.equal(format(0), '0（约 0%）')
assert.equal(format(1), '1（约 100%）')
for (const value of [null, undefined, '0.85', NaN, Infinity, -0.1, 1.1]) assert.equal(format(value), '未采集或格式无效')
for (const field of ['nearfull_ratio', 'backfillfull_ratio', 'full_ratio']) assert.ok(source.includes(`osdCapacityThreshold(values.${field})`))
assert.ok(source.includes("await refreshResource({ clusterId, kinds: ['osd_flag'] })"))
assert.ok(source.includes("await getResource('/osd/flag', clusterId)"))

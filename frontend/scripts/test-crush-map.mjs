import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/cluster/CrushMapPage.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('CrushMapPage.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const functions = tree.statements.filter((node) => ts.isFunctionDeclaration(node) && ['crushTree', 'crushStatus'].includes(node.name.text))
const exports = {}
new Function('exports', ts.transpileModule(functions.map((fn) => fn.getText(tree)).join('\n'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const nodes = [{ id: 0, name: 'osd.0', type: 'osd', status: 'up' }, { id: -1, name: 'default', type: 'root', children: [0] }]
const result = exports.crushTree({ nodes, roots: [-1] })
assert.equal(result[0].title, 'default (root)')
assert.equal(result[0].children[0].key, '-1/0')
assert.equal(result[0].children[0].title, 'osd.0 (osd)')
assert.equal(result[0].children[0].nodeId, 0)
assert.equal(result[0].children[0].status, 'up')
for (const status of ['up', 'in']) assert.equal(exports.crushStatus(status).color, 'success')
for (const status of ['down', 'out', 'destroyed']) assert.equal(exports.crushStatus(status).color, 'error')
for (const status of [undefined, null, '', 1, true, 'unrecognized']) assert.equal(exports.crushStatus(status).color, 'default')
assert.equal(exports.crushStatus(null).label, '未知')
assert.equal(exports.crushStatus('unrecognized').label, 'unrecognized')
assert.deepEqual(exports.crushTree({ nodes: [], roots: [] }), [])
assert.throws(() => exports.crushTree({ nodes, roots: [99] }))
assert.throws(() => exports.crushTree({ nodes: [{ id: -1, name: 'root', type: 'root', children: [-1] }], roots: [-1] }))
assert.equal(nodes[0].id, 0, 'must not reorder native metadata')
console.log('CRUSH topology tree checks passed')

const poolSource = readFileSync(new URL('../src/pages/cluster/PoolManagementPage.tsx', import.meta.url), 'utf8')
const poolTree = ts.createSourceFile('PoolManagementPage.tsx', poolSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const poolFunctions = poolTree.statements.filter((node) => ts.isFunctionDeclaration(node) && ['topologyCounts', 'crushPath'].includes(node.name.text))
const poolCode = ts.transpileModule(poolFunctions.map((fn) => fn.getText(poolTree)).join('\n') + '\nexports.counts = topologyCounts', { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
const poolExports = {}
new Function('exports', 'textValue', 'isRecord', poolCode)(poolExports, (value, fallback) => typeof value === 'string' ? value : fallback, (value) => value !== null && typeof value === 'object' && !Array.isArray(value))
const osds = [
  { host: 'a', device_class: 'ssd', crush_path: { root: 'default', host: 'a' } },
  { host: 'b', device_class: 'hdd', crush_path: { root: 'default', host: 'b' } },
  { host: 'c', device_class: 'ssd', crush_path: { root: 'archive', host: 'c' } },
  { host: 'unknown', device_class: 'ssd' }
]
assert.equal(poolExports.counts(osds, 'default').osd, 2)
assert.equal(poolExports.counts(osds, 'default').host, 2)
assert.equal(poolExports.counts(osds, 'default', 'ssd').osd, 1)
assert.equal(poolExports.counts(osds, 'archive').osd, 1)
assert.equal(poolExports.counts(osds, 'a').osd, 1)
assert.equal(poolExports.counts(osds, 'missing').osd, 0)
assert.equal(poolExports.counts([], 'default').osd, 0)
console.log('CRUSH failure domain root isolation checks passed')

const rulesSource = readFileSync(new URL('../src/pages/cluster/CrushRulesPanel.tsx', import.meta.url), 'utf8')
const rulesTree = ts.createSourceFile('CrushRulesPanel.tsx', rulesSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const rulesFunctions = rulesTree.statements.filter((node) => ts.isFunctionDeclaration(node) && ['crushRuleType', 'crushRuleSteps', 'crushRuleDeleteBlocked'].includes(node.name.text))
const rules = {}
new Function('exports', ts.transpileModule(rulesFunctions.map((fn) => fn.getText(rulesTree)).join('\n'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(rules)
assert.equal(rules.crushRuleType(1), '复制 (1)')
assert.equal(rules.crushRuleType(3), '纠删码 (3)')
assert.equal(rules.crushRuleType(5), '类型 5')
assert.equal(rules.crushRuleType(null), '未知')
const steps = [{ op: 'take', item: -1, item_name: 'default' }, { op: 'chooseleaf_firstn', num: 0, type: 'host' }, { op: 'emit' }]
const rendered = rules.crushRuleSteps(steps)
assert.deepEqual(rendered.map((step) => step.sequence), [1, 2, 3])
assert.equal(rendered[0].item, -1)
assert.equal(rendered[1].num, 0)
assert.equal(steps[0].sequence, undefined)
assert.deepEqual(rules.crushRuleSteps([]), [])
for (const invalid of [null, {}, [null], [{ op: 3 }]]) assert.equal(rules.crushRuleSteps(invalid), null)
console.log('CRUSH rule type and native step checks passed')
assert.equal(rules.crushRuleDeleteBlocked({ rule_name: 'a', stale: false }), undefined)
for (const row of [{ rule_name: 'a', stale: true }, { rule_name: 'a' }, { stale: false }, { rule_name: '', stale: false }]) assert.ok(rules.crushRuleDeleteBlocked(row))

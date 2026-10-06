import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/monitoring/MetricTrend.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('MetricTrend.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const fn = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'metricTrend')
const exports = {}
new Function('exports', ts.transpileModule(fn.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const meta = { start: '1970-01-01T00:00:00Z', end: '1970-01-01T00:02:00Z', step_seconds: 30 }
const point = (time, value, status = '有效值') => ({ index: time, timestamp: String(time), value: String(value), utc: 'time', status })
const chart = exports.metricTrend([point(0, -10), point(30, 0), point(60, 10)], meta)
assert.equal(chart.min, -10); assert.equal(chart.max, 10)
assert.deepEqual(chart.points.map(point => [point.x, point.y]), [[60, 140], [230, 85], [400, 30]])
assert.equal(chart.path.match(/L/g).length, 2)
for (const points of [
  [point(0, 0), point(60, 10)],
  [point(0, 0), point(30, 'NaN', '非有限值（不代表零）'), point(60, 10)],
  [point(0, 0), point(30, 'bad', '格式异常'), point(60, 10)],
  [point(30, 0), point(30, 10)],
  [point(60, 0), point(30, 10)]
]) assert.equal(exports.metricTrend(points, meta).path.includes('L'), false)
assert.equal(exports.metricTrend([point(0, 0)], meta).points[0].y, 85)
assert.equal(exports.metricTrend([point(0, 0), point(30, 0)], meta).points[1].y, 85)
const extreme = exports.metricTrend([point(0, -1e308), point(30, 1e308)], meta)
assert.ok(extreme.points.every(point => Number.isFinite(point.x) && Number.isFinite(point.y)))
assert.equal(exports.metricTrend([point(-1, 1), point(121, 1)], meta), null)
assert.equal(exports.metricTrend([point(0, 'NaN')], meta), null)
assert.equal(exports.metricTrend([], meta), null)
for (const bad of [undefined, {}, { ...meta, step_seconds: 0 }, { ...meta, end: meta.start }, { ...meta, start: 'bad' }]) assert.equal(exports.metricTrend([point(0, 1)], bad), null)
assert.ok(source.includes('<title>{point.sample.utc}: {point.sample.value}</title>'))
console.log('Metric trend bounds, gap handling and finite coordinates passed')

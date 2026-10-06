import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/MonSessionHistory.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('history.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const fn = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'monSessionSeries')
const exports = {}
new Function('exports', ts.transpileModule(fn.getText(tree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const row = { metric: { cluster: 'fsid', ceph_daemon: 'mon.a', instance: 'host1' }, values: [[1, '9007199254740993'], [2, 'NaN']] }
const data = { result_type: 'matrix', meta: { cluster_fsid: 'fsid', mon_name: 'a' }, series: [row] }
assert.equal(exports.monSessionSeries(data, 'a'), data.series)
assert.deepEqual(exports.monSessionSeries({ ...data, series: [] }, 'a'), [])
for (const bad of [{ ...data, meta: {} }, { ...data, result_type: 'vector' }, { ...data, series: [null] }, { ...data, series: [{ ...row, metric: { cluster: 'other', ceph_daemon: 'mon.a' } }] }, { ...data, series: [{ ...row, values: null }] }]) assert.throws(() => exports.monSessionSeries(bad, 'a'))
assert.throws(() => exports.monSessionSeries(data, 'b'))
assert.ok(source.includes('controller.current?.abort()'))
assert.ok(source.includes('if (!active.signal.aborted) setData(result)'))
assert.ok(source.includes('metricId: \'mon_sessions\', monName'))
assert.ok(readFileSync(new URL('../src/pages/cluster/MonDetailPage.tsx', import.meta.url), 'utf8').includes('<MonSessionHistory key={JSON.stringify([selectedClusterId, monName])}'))
console.log('MON session history keeps raw samples and rejects cross-scope sequences')
const component = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'MonSessionHistory')
const queryNode = component.body.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'query')
const js = ts.transpileModule(queryNode.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2020 } }).outputText
for (const leave of [false, true]) {
  const controller = { current: null }, updates = []
  const query = new Function('controller', 'setLoading', 'setError', 'setData', 'queryMetricRange', 'clusterId', 'monName', 'monSessionSeries', `${js}; return query`)(controller, value => updates.push(['busy', value]), value => updates.push(['error', value]), value => updates.push(['data', value]), async (id, input, init) => {
    assert.equal(id, 7); assert.equal(input.monName, 'a'); assert.equal(input.metricId, 'mon_sessions')
    assert.equal(Date.parse(input.end) - Date.parse(input.start), 3600000)
    assert.equal(init.signal, controller.current.signal)
    if (leave) controller.current.abort()
    return data
  }, 7, 'a', exports.monSessionSeries)
  await query()
  assert.deepEqual(updates.slice(0, 3), [['busy', true], ['error', ''], ['data', null]])
  assert.deepEqual(updates.slice(3), leave ? [] : [['data', data], ['busy', false]])
}

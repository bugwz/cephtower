import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/HostDetailPage.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('host.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const component = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'HostPerformancePanel')
const effect = component.body.statements.find(node => ts.isExpressionStatement(node) && ts.isCallExpression(node.expression) && node.expression.expression.getText(tree) === 'useEffect').expression.arguments[0]
const code = ts.transpileModule(`const effect = ${effect.getText(tree)};`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const cancelled of [false, true]) {
  const notices = [], pending = []
  const env = { clusterId: 7, hostname: 'node', address: '192.0.2.1', featureStatus: {}, hostPerformanceMetrics: [{ key: 'cpu', metricId: 'host_cpu_usage' }, { key: 'mem', metricId: 'host_memory_usage' }], setLoading: () => {}, setQueryError: () => {}, setValues: () => {}, setNotices: value => notices.push(value), metricValueForHost: () => 0, queryMetric: (id, input) => new Promise((resolve, reject) => pending.push({ id, input, resolve, reject })) }
  const run = new Function(...Object.keys(env), `${code}; return effect`)(...Object.values(env))
  const cleanup = run()
  assert.deepEqual(notices, [{}])
  assert.ok(pending.every(call => call.id === 7))
  if (cancelled) cleanup()
  const meta = { warnings: ['partial CPU data'], infos: ['hint'] }
  pending[0].resolve({ meta }); pending[1].reject(new Error('unavailable'))
  await new Promise(resolve => setTimeout(resolve, 0))
  assert.deepEqual(notices, cancelled ? [{}] : [{}, { cpu: meta }])
}
assert.ok(source.includes('<MetricNotices meta={notices[metric.key]} source={metric.title} />'))

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/monitoring/MetricNotices.tsx', import.meta.url), 'utf8')
const exports = {}
new Function('exports', 'require', 'React', ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.React, target: ts.ScriptTarget.ES2022 } }).outputText)(exports, () => ({ Alert: 'Alert', Space: 'Space' }), { createElement: (type, props, ...children) => ({ type, props, children }) })
const render = meta => exports.MetricNotices({ meta }).children.flat()
for (const meta of [undefined, {}, { warnings: null, infos: 'bad' }]) assert.deepEqual(render(meta), [])
const notices = render({ warnings: ['partial data', '<script>not executed</script>', null, 2, ''], infos: ['sample omitted'] })
assert.deepEqual(notices.map(item => item.props.type), ['warning', 'warning', 'info'])
assert.deepEqual(notices.map(item => item.props.description), ['partial data', '<script>not executed</script>', 'sample omitted'])
assert.ok(notices[0].props.message.includes('不完整'))
const named = exports.MetricNotices({ source: 'GET 带宽', meta: { warnings: ['partial'], infos: ['hint'] } }).children.flat()
assert.ok(named.every(item => item.props.message.startsWith('GET 带宽 · Prometheus')))
assert.deepEqual(named.map(item => item.props.description), ['partial', 'hint'])

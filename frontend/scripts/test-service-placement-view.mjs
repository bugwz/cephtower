import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/cluster/ServicePlacement.tsx', import.meta.url), 'utf8')
const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.React } }).outputText
const exports = {}
new Function('exports', 'require', 'React', code)(exports, name => name === 'antd' ? { Alert: 'Alert', Descriptions: 'Descriptions', Space: 'Space' } : { isRecord: value => value !== null && typeof value === 'object' && !Array.isArray(value) }, { createElement: (type, props, ...children) => ({ type, props, children }) })
const placement = { hosts: ['node1:10.0.0.1=daemon-a', 'node2'], count: 2, count_per_host: 1, label: 'storage', host_pattern: { pattern: 'node[12]', pattern_type: 'regex' }, future: false }
const before = JSON.stringify(placement)
const fields = exports.servicePlacementFields(placement)
assert.deepEqual(fields.map(field => field.label), ['指定主机', '实例数量', '每主机实例数量', '主机标签', '主机匹配规则', 'future'])
assert.equal(fields.at(-1).value, 'false')
const render = (placement, unmanaged) => JSON.stringify(exports.ServicePlacement({ placement, unmanaged }))
const full = render(placement, true)
for (const text of ['node1:10.0.0.1=daemon-a', 'storage', 'regex', '原始放置配置', '不自动执行']) assert.ok(full.includes(text), text)
assert.equal(JSON.stringify(placement), before)
for (const unmanaged of [false, undefined, null, 'true']) assert.ok(!render(placement, unmanaged).includes('不自动执行'))
for (const invalid of [undefined, null, false, [], 'node1']) assert.ok(render(invalid, false).includes('放置配置未返回或格式无效'))
assert.ok(render({}, false).includes('具体默认值由 Ceph 决定'))
assert.deepEqual(exports.servicePlacementFields({ count: 0, label: '', hosts: [], future: null }).map(field => field.value), ['0', '空字符串', '[]', '未报告'])
const page = readFileSync(new URL('../src/pages/cluster/ServicePage.tsx', import.meta.url), 'utf8')
assert.ok(page.includes('<ServicePlacement placement={value} unmanaged={row.unmanaged} />'))
console.log('Service placement details retain native host specs, regex patterns and unknown fields')

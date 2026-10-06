import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/cluster/DeviceHardwareSummary.tsx', import.meta.url), 'utf8')
const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.React } }).outputText
const exports = {}
new Function('exports', 'require', 'React', code)(exports, name => name === 'antd' ? { Alert: 'Alert', Card: 'Card', Descriptions: { Item: 'Item' }, Space: 'Space', Typography: { Text: 'Text' } } : { isRecord: value => value !== null && typeof value === 'object' && !Array.isArray(value) }, { createElement: (type, props, ...children) => ({ type, props, children }) })
const render = (lsm_data, stale = false, rowStale = false) => JSON.stringify(exports.DeviceHardwareSummary({ device: { lsm_data, stale: rowStale }, stale }))
const full = render({ serialNum: 'disk-serial', health: 'Unknown', mediaType: 'HDD', transport: 'SAS', rpm: '7200', linkSpeed: '12000', ledSupport: { IDENTsupport: 'Supported', IDENTstatus: 'Off', FAILsupport: 'Unsupported', FAILstatus: 'Unknown' }, errors: ['query failed', 'controller unavailable'] })
for (const value of ['disk-serial', 'Unknown', 'HDD', 'SAS', '7200', '12000', 'Supported', 'Off', 'Unsupported', 'query failed', 'controller unavailable', '不是操作后的实时回读']) assert.ok(full.includes(value))
assert.ok(!full.includes('库存已过期'))
for (const value of [undefined, null, {}, [], { health: true, ledSupport: [] }, { errors: ['valid', 1] }]) assert.ok(render(value).includes('未报告'))
assert.ok(render({ errors: [] }).includes('未报告查询错误（不代表健康）'))
for (const [stale, rowStale] of [[true, false], [false, true], [null, false], [false, null]]) assert.ok(render({}, stale, rowStale).includes('库存已过期'))
const page = readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8')
assert.ok(page.includes('device && <DeviceHardwareSummary device={device} stale={data?.stale} />'))
console.log('Device hardware detail preserves native LSM reports and unknown or stale states')

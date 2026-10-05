import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const jsx = (type, props) => ({ type, props })
function load(file, require = () => { throw new Error('unexpected import') }) {
  const exports = {}
  new Function('exports', 'require', ts.transpileModule(readFileSync(new URL(`../src/pages/object/${file}`, import.meta.url), 'utf8'), {
    compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX }
  }).outputText)(exports, require)
  return exports
}
const identity = load('rgwUserIdentity.ts')
const state = load('rgwBucketState.ts')
const view = load('RgwBucketSummary.tsx', name => {
  if (name === 'react/jsx-runtime') return { jsx, jsxs: jsx }
  if (name === 'antd') return { Descriptions: 'Descriptions' }
  if (name === './rgwUserIdentity') return identity
  if (name === './rgwBucketState') return state
  throw new Error(name)
})
const row = { name: 'photos', tenant: '', id: 'opaque:id', owner: 'RGW123', versioning: 'suspended', mfa_enabled: false,
  object_lock_enabled: true, zonegroup: 'zone-id', placement_rule: 'default/STANDARD', creation_time: '2026-10-03 00:00:00.123456Z', mtime: 'raw-time' }
const items = view.RgwBucketSummary({ row }).props.items
const values = Object.fromEntries(items.map(item => [item.key, item.children]))
assert.equal(items.length, 11)
for (const key of ['name', 'id', 'owner', 'zonegroup']) assert.equal(values[key], row[key])
assert.equal(values.tenant, '默认租户')
assert.equal(values.placement, row.placement_rule)
assert.equal(values.created, row.creation_time)
assert.equal(values.modified, row.mtime)
assert.ok(values.versioning.includes('不等同于从未启用'))
assert.equal(values.mfa, '未启用')
assert.equal(values.lock, '已启用')
for (const invalid of [undefined, null, 0, false, {}, []]) {
  const bad = Object.fromEntries(Object.keys(row).map(key => [key, invalid]))
  const fields = view.RgwBucketSummary({ row: bad }).props.items
  for (const item of fields.filter(item => !['mfa', 'lock'].includes(item.key))) assert.ok(item.children.includes('未返回'), item.key)
}
assert.ok(items.find(item => item.key === 'lock').label.includes('非保留策略'))
console.log('Bucket summary preserves native identities, states and timestamps')

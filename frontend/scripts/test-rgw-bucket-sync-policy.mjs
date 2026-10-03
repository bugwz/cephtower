import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketSyncPolicy.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(api)
const group = { id: 'team-policy', status: 'enabled', pipes: [], data_flow: {} }
const summary = value => api.rgwBucketSyncPolicy(value)
assert.match(summary({ groups: [group] }), /已启用.*不代表 Zonegroup/)
assert.match(summary({ groups: [{ ...group, status: 'allowed' }] }), /允许（未启用）/)
assert.match(summary({ groups: [{ ...group, status: 'forbidden' }] }), /禁止/)
for (const status of ['future', 'constructor', 'Enabled']) assert.match(summary({ groups: [{ ...group, status }] }), /未知/)
assert.match(summary({ groups: [] }), /无桶本地同步组.*不代表/)
for (const value of [null, {}, [], { groups: null }, { groups: [null] }, { groups: [group, group] }, { groups: [{ ...group, status: null }] }, { groups: [{ ...group, pipes: {} }] }, { groups: [{ ...group, data_flow: [] }] }]) assert.match(summary(value), /不可用/)
assert.match(summary({ groups: [{ ...group, id: '<script>' }] }), /"<script>"/)
console.log('bucket local sync policy summary checks passed')
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketSyncGroupForm.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(api)
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const source = ts.createSourceFile('pages.tsx', pages, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let action
function visit(node) {
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(source) === 'title' && p.initializer.text === '修改桶同步组状态')) {
    const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    action = new Function(...Object.keys(api), `${code}; return action`)(...Object.values(api))
  }
  ts.forEachChild(node, visit)
}
visit(source)
assert.ok(action)
assert.equal(action.path, '/rgw/bucket/sync/group')
assert.equal(action.method, 'PATCH')
const row = { natural_key: 'AGJ1Y2tldA', bucket_sync_policy: { groups: [group] } }
const initial = action.initialValues(row)
assert.equal(initial.status, undefined)
assert.equal(initial.group_id, undefined)
assert.equal(initial.confirm_change, undefined)
assert.ok(action.disabledWhen({ ...row, stale: true }))
assert.ok(action.disabledWhen({ ...row, bucket_sync_policy: { groups: [] } }))
for (const status of ['allowed', 'forbidden']) {
  const values = { ...initial, group_id: group.id, status, confirm_change: 'acknowledged' }
  assert.deepEqual(action.buildBody(values, 7, row), { cluster_id: 7, bucket_id: row.natural_key, group_id: group.id, status, expected_status: 'enabled' })
  assert.match(action.confirmation(values, row), /AGJ1Y2tldA.*team-policy.*不会新建数据流或管道/)
  for (const change of [{ bucket_id: 'other' }, { group_id: 'other' }, { status: 'enabled' }, { status: 'unknown' }, { confirm_change: true }]) assert.throws(() => action.buildBody({ ...values, ...change }, 7, row))
}
console.log('bucket sync group form and action binding checks passed')

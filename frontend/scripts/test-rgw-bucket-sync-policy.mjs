import assert from 'node:assert/strict'
import './test-rgw-bucket-sync-flows.mjs'
import './test-rgw-bucket-sync-pipes.mjs'
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
let action, createAction, deleteAction, flowAction, deleteFlowAction, deletePipeAction
function visit(node) {
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(source) === 'title' && p.initializer.text === '删除桶同步管道')) {
    const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    deletePipeAction = new Function(...Object.keys(api), `${code}; return action`)(...Object.values(api))
  }
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(source) === 'title' && p.initializer.text === '删除桶数据流')) {
    const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    deleteFlowAction = new Function(...Object.keys(api), `${code}; return action`)(...Object.values(api))
  }
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(source) === 'title' && p.initializer.text === '创建桶数据流')) {
    const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    flowAction = new Function(...Object.keys(api), `${code}; return action`)(...Object.values(api))
  }
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(source) === 'title' && p.initializer.text === '删除桶同步组')) {
    const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    deleteAction = new Function(...Object.keys(api), `${code}; return action`)(...Object.values(api))
  }
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(source) === 'title' && p.initializer.text === '修改桶同步组状态')) {
    const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    action = new Function(...Object.keys(api), `${code}; return action`)(...Object.values(api))
  }
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(source) === 'title' && p.initializer.text === '创建桶同步组')) {
    const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    createAction = new Function(...Object.keys(api), `${code}; return action`)(...Object.values(api))
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
assert.ok(createAction)
assert.equal(createAction.path, '/rgw/bucket/sync/group')
assert.equal(createAction.method, 'POST')
const emptyRow = { ...row, bucket_sync_policy: { groups: [] } }
assert.equal(createAction.disabledWhen(emptyRow), undefined)
assert.ok(createAction.disabledWhen({ ...row, bucket_sync_policy: null }))
assert.ok(createAction.disabledWhen({ ...row, stale: true }))
const creation = createAction.initialValues(emptyRow)
assert.equal(creation.status, undefined)
assert.equal(creation.confirm_create, undefined)
for (const status of ['enabled', 'allowed', 'forbidden']) {
  const values = { ...creation, group_id: ' 新组 ', status, confirm_create: 'acknowledged' }
  assert.deepEqual(createAction.buildBody(values, 7, row), { cluster_id:7, bucket_id:row.natural_key, group_id:' 新组 ', status })
  assert.match(createAction.confirmation(values, row), /空数据流和空管道.*不建立可工作的复制链路/)
  for (const change of [{ group_id: group.id }, { group_id:'' }, { group_id:'-bad' }, { group_id:'a\nb' }, { group_id:'\ud800' }, { group_id:'中'.repeat(171) }, { bucket_id:'other' }, { status:'unknown' }, { confirm_create:true }]) assert.throws(() => createAction.buildBody({ ...values, ...change }, 7, row))
}
console.log('bucket sync group creation form checks passed')
assert.equal(deleteAction.method, 'DELETE')
assert.equal(deleteAction.path, '/rgw/bucket/sync/group')
assert.ok(deleteAction.disabledWhen(emptyRow))
assert.ok(deleteAction.disabledWhen({ ...row, stale: true }))
const deletion = { ...deleteAction.initialValues(row), group_id: group.id, confirm_delete: 'acknowledged' }
assert.deepEqual(deleteAction.buildBody(deletion, 7, row), { cluster_id: 7, bucket_id: row.natural_key, group_id: group.id, expected_group: JSON.stringify(group) })
assert.match(deleteAction.confirmation(deletion, row), /全部数据流、管道.*forbidden.*不自动回滚/)
for (const change of [{ bucket_id: 'other' }, { group_id: 'other' }, { confirm_delete: true }]) assert.throws(() => deleteAction.buildBody({ ...deletion, ...change }, 7, row))
console.log('bucket sync group deletion form checks passed')
assert.equal(flowAction.path, '/rgw/bucket/sync/flow')
assert.equal(flowAction.method, 'POST')
assert.ok(flowAction.disabledWhen(emptyRow))
assert.ok(flowAction.disabledWhen({ ...row, stale: true }))
const flowBase = { ...flowAction.initialValues(row), group_id: group.id, confirm_flow: 'acknowledged' }
const sym = { ...flowBase, flow_type: 'symmetrical', flow_id: ' 流 ', zones_json: '["b","a"]' }
const dir = { ...flowBase, flow_type: 'directional', source_zone: 'a', dest_zone: 'b' }
assert.deepEqual(flowAction.buildBody(sym, 7, row), { cluster_id: 7, bucket_id: row.natural_key, group_id: group.id, expected_group: JSON.stringify(group), flow_type: 'symmetrical', flow_id: ' 流 ', zones: ['b', 'a'] })
assert.deepEqual(flowAction.buildBody(dir, 7, row), { cluster_id: 7, bucket_id: row.natural_key, group_id: group.id, expected_group: JSON.stringify(group), flow_type: 'directional', source_zone: 'a', dest_zone: 'b' })
for (const values of [sym, dir]) {
  assert.match(flowAction.confirmation(values, row), /不创建管道.*不自动回滚/)
  for (const change of [{ bucket_id: 'other' }, { group_id: 'other' }, { confirm_flow: true }, { flow_type: 'unknown' }]) assert.throws(() => flowAction.buildBody({ ...values, ...change }, 7, row))
}
for (const change of [{ flow_id: '-bad' }, { zones_json: '[]' }, { zones_json: '["a","a"]' }, { zones_json: '["*"]' }, { zones_json: '["a,b"]' }, { zones_json: '[3]' }, { zones_json: 'broken' }, { source_zone: 'a' }]) assert.throws(() => flowAction.buildBody({ ...sym, ...change }, 7, row))
for (const change of [{ source_zone: 'b' }, { dest_zone: '' }, { flow_id: 'f' }, { zones_json: '[]' }]) assert.throws(() => flowAction.buildBody({ ...dir, ...change }, 7, row))
const existingRow = { ...row, bucket_sync_policy: { groups: [{ ...group, data_flow: { symmetrical: [{ id: ' 流 ', zones: ['a'] }], directional: [{ source_zone: 'a', dest_zone: 'b' }] } }] } }
for (const values of [sym, dir]) assert.throws(() => flowAction.buildBody(values, 7, existingRow))
console.log('bucket sync flow creation form and binding checks passed')
assert.equal(deleteFlowAction.path, '/rgw/bucket/sync/flow')
assert.equal(deleteFlowAction.method, 'DELETE')
assert.ok(deleteFlowAction.disabledWhen({ ...existingRow, stale: true }))
const deleteBase = { ...deleteFlowAction.initialValues(existingRow), group_id: group.id, confirm_flow_delete: 'acknowledged' }
const deleteSym = { ...deleteBase, flow_type: 'symmetrical', flow_id: ' 流 ' }
const deleteDir = { ...deleteBase, flow_type: 'directional', source_zone: 'source-id', dest_zone: 'dest-id' }
for (const values of [deleteSym, deleteDir]) {
  const body = deleteFlowAction.buildBody(values, 7, existingRow)
  assert.equal(body.cluster_id, 7)
  assert.equal(body.expected_group, JSON.stringify(existingRow.bucket_sync_policy.groups[0]))
  assert.equal(body.zones, undefined)
  assert.match(deleteFlowAction.confirmation(values, existingRow), /保留组状态、其他流和管道.*不自动回滚/)
  for (const change of [{ bucket_id: 'other' }, { group_id: 'other' }, { confirm_flow_delete: true }, { flow_type: 'unknown' }]) assert.throws(() => deleteFlowAction.buildBody({ ...values, ...change }, 7, existingRow))
  assert.throws(() => deleteFlowAction.buildBody(values, 7, row))
}
assert.match(deleteFlowAction.confirmation(deleteSym, existingRow), /包含全部 Zone/)
assert.equal(deleteFlowAction.buildBody(deleteDir, 7, existingRow).source_zone, 'source-id')
for (const change of [{ flow_id: 'missing' }, { source_zone: 'a' }]) assert.throws(() => deleteFlowAction.buildBody({ ...deleteSym, ...change }, 7, existingRow))
for (const change of [{ source_zone: 'dest-id' }, { dest_zone: '*' }, { source_zone: 'a;b' }, { flow_id: 'f' }]) assert.throws(() => deleteFlowAction.buildBody({ ...deleteDir, ...change }, 7, existingRow))
console.log('bucket sync flow deletion form and binding checks passed')
assert.equal(deletePipeAction.path, '/rgw/bucket/sync/pipe')
assert.equal(deletePipeAction.method, 'DELETE')
const selectedPipe = { id: ' 管道 ', source: { bucket: '*', zones: ['*'] }, dest: { bucket: 'team/photos:marker', zones: ['Zone B'] }, params: { mode: 'user' } }
const pipeGroup = { ...group, pipes: [selectedPipe] }
const pipeRow = { ...row, bucket_sync_policy: { groups: [pipeGroup] } }
const pipeValues = { ...deletePipeAction.initialValues(pipeRow), group_id: group.id, pipe_id: selectedPipe.id, confirm_pipe_delete: 'acknowledged' }
assert.deepEqual(deletePipeAction.buildBody(pipeValues, 7, pipeRow), { cluster_id: 7, bucket_id: row.natural_key, group_id: group.id, pipe_id: selectedPipe.id, expected_group: JSON.stringify(pipeGroup) })
assert.match(deletePipeAction.confirmation(pipeValues, pipeRow), /team\/photos:marker.*保留组状态、数据流及其他管道.*不自动回滚/)
assert.ok(deletePipeAction.disabledWhen({ ...pipeRow, stale: true }))
for (const change of [{ bucket_id: 'other' }, { group_id: 'other' }, { pipe_id: 'missing' }, { confirm_pipe_delete: true }]) assert.throws(() => deletePipeAction.buildBody({ ...pipeValues, ...change }, 7, pipeRow))
for (const pipes of [[], null, [null], [selectedPipe, selectedPipe]]) assert.throws(() => deletePipeAction.buildBody(pipeValues, 7, { ...pipeRow, bucket_sync_policy: { groups: [{ ...group, pipes }] } }))
console.log('bucket sync pipe deletion form and binding checks passed')

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
import './test-rgw-bucket-notification-write.mjs'
const api = {}
const jsx = (type, props) => ({ type, props })
const compile = file => ts.transpileModule(readFileSync(new URL(file, import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText
new Function('exports', 'require', compile('../src/pages/object/RgwBucketNotifications.tsx'))(api, name => name === 'antd' ? { Table: 'Table' } : { jsx, jsxs: jsx })
new Function('exports', compile('../src/pages/object/rgwBucketConfiguration.ts'))(api)
const rule = { id: '<script>', topic: 'arn:aws:sns:east::topic', events: ['s3:ObjectCreated:*', 'future'], filters: [{ kind: 'S3Key', name: 'regex', value: ' x.* ' }, { kind: 'S3Metadata', name: 'x-amz-meta-a', value: '' }, { kind: 'S3Tags', name: '标签', value: '<img>' }] }
assert.deepEqual(api.bucketNotificationData([rule, rule]), [rule, rule])
assert.deepEqual(api.bucketNotificationData([]), [])
for (const invalid of [null, {}, [null], [{ ...rule, id: 1 }], [{ ...rule, topic: null }], [{ ...rule, events: [1] }], [{ ...rule, filters: null }], [{ ...rule, filters: [null] }], [{ ...rule, filters: [{ kind: 'unknown', name: 'a', value: 'b' }] }], [{ ...rule, filters: [{ kind: 'S3Key', name: 'prefix', value: null }] }]]) assert.equal(api.bucketNotificationData(invalid), undefined)
for (const configured of [false, undefined, null]) assert.match(api.RgwBucketNotifications({ value: [], configured }).props.children, /不可用/)
const table = api.RgwBucketNotifications({ value: [rule, rule], configured: true }).props.children.find(child => child.type === 'Table')
assert.equal(table.props.rowKey, 'index')
assert.deepEqual(table.props.dataSource.map(row => row.index), [0, 1])
assert.equal(table.props.columns[0].render(rule.id), '"<script>"')
assert.equal(table.props.columns[1].render(rule.topic), JSON.stringify(rule.topic))
assert.equal(table.props.columns[2].render(rule.events)[1].props.children, '"future"')
assert.match(table.props.columns[2].render([]), /不推断/)
assert.equal(table.props.columns[3].render(rule.filters).length, 3)
assert.equal(table.props.columns[3].render([]), '未返回过滤条件')
assert.match(table.props.locale.emptyText, /无事件通知/)
assert.ok(api.rgwBucketConfigurationReadOptions.some(option => option.value === 'notification'))
assert.ok(!api.rgwBucketConfigurationOptions.some(option => option.value === 'notification'))
const row = { bucket_id: 'AGJ1Y2tldA', kind: 'notification', configured: true, document: '<NotificationConfiguration/>', notifications: [] }
assert.ok(api.rgwBucketConfigurationEditBlocked(row))
assert.ok(api.rgwBucketConfigurationDeleteBlocked(row))
// Verify the actual page column binding, not just a disconnected component.
const source = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const page = ts.createSourceFile('pages.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let column
function visit(node) {
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(page) === 'key' && ts.isStringLiteral(p.initializer) && p.initializer.text === 'notifications')) {
    const code = ts.transpileModule(`exports.column = ${node.getText(page)}`, { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX } }).outputText
    const output = {}
    new Function('exports', 'require', 'RgwBucketNotifications', code)(output, () => ({ jsx, jsxs: jsx }), api.RgwBucketNotifications)
    column = output.column
  }
  ts.forEachChild(node, visit)
}
visit(page)
assert.equal(column.render([], row).type, api.RgwBucketNotifications)
assert.equal(column.render([], { kind: 'policy' }), '—')
console.log('bucket notification presentation and read-only bindings passed')

const form = {}
new Function('exports', compile('../src/pages/object/rgwBucketNotificationDelete.ts'))(form)
const populated = { ...row, notifications: [rule], document: '<NotificationConfiguration><TopicConfiguration/></NotificationConfiguration>' }
const values = { ...form.bucketNotificationDeleteInitial(populated), mode: 'single', notification_id: rule.id, confirm_notification: 'acknowledged' }
assert.equal(form.bucketNotificationDeleteInitial(populated).mode, undefined)
assert.equal(form.bucketNotificationDeleteInitial(populated).confirm_notification, undefined)
assert.deepEqual(form.bucketNotificationDeleteInput(values, populated), { bucket_id: row.bucket_id, mode: 'single', notification_id: rule.id, expected_document: populated.document })
assert.equal(form.bucketNotificationDeleteInput({ ...values, mode: 'all', notification_id: '' }, populated).mode, 'all')
assert.ok(form.bucketNotificationDeleteBlocked(row))
for (const patch of [{ configured: false }, { kind: 'policy' }, { bucket_id: '' }, { document: null }, { notifications: null }, { notifications: [null] }]) assert.ok(form.bucketNotificationDeleteBlocked({ ...populated, ...patch }))
for (const patch of [{ bucket_id: 'other' }, { mode: undefined }, { notification_id: '' }, { notification_id: 'absent' }, { mode: 'all' }, { confirm_notification: undefined }]) assert.throws(() => form.bucketNotificationDeleteInput({ ...values, ...patch }, populated))
assert.throws(() => form.bucketNotificationDeleteInput(values, { ...populated, notifications: [rule, rule] }), /重复/)
assert.throws(() => form.bucketNotificationDeleteInput(values, { ...populated, document: 'x'.repeat(1024 * 1024) }), /上限/)
const exact = ' a&+%/中 '
assert.equal(form.bucketNotificationDeleteInput({ ...values, notification_id: exact }, { ...populated, notifications: [{ ...rule, id: exact }] }).notification_id, exact)
for (const text of ['指定通知', '不删除 Bucket', '独立 Topic', '排队消息', '映射', '原子锁', '回滚', '重试']) assert.ok(form.bucketNotificationDeleteConfirmation(values, populated).includes(text))
assert.match(form.bucketNotificationDeleteConfirmation({ ...values, mode: 'all', notification_id: '' }, populated), /全部通知规则/)
let action
function findAction(node) {
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(page) === 'title' && ts.isStringLiteral(p.initializer) && p.initializer.text === '删除 Bucket 通知规则')) {
    const code = ts.transpileModule(`const action = ${node.getText(page)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    action = new Function(...Object.keys(form), `${code}; return action`)(...Object.values(form))
  }
  ts.forEachChild(node, findAction)
}
findAction(page)
assert.equal(action.path, '/rgw/bucket/notification')
assert.equal(action.method, 'DELETE')
assert.equal(action.visibleWhen(populated), true)
assert.equal(action.visibleWhen({ kind: 'policy' }), false)
assert.equal(action.disabledWhen, form.bucketNotificationDeleteBlocked)
assert.equal(action.initialValues, form.bucketNotificationDeleteInitial)
assert.equal(action.confirmation, form.bucketNotificationDeleteConfirmation)
assert.equal(action.fields.find(field => field.name === 'bucket_id').readOnly, true)
assert.deepEqual(action.buildBody(values, 42, populated), { cluster_id: 42, ...form.bucketNotificationDeleteInput(values, populated) })
console.log('bucket notification deletion forms, snapshot and action bindings passed')

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api = {}
const compile = file => ts.transpileModule(readFileSync(new URL(file, import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText
new Function('exports', compile('../src/pages/object/rgwBucketNotificationForm.ts'))(api)
const row = { bucket_id: 'AGJ1Y2tldA', kind: 'notification', configured: true, document: '<NotificationConfiguration/>', notifications: [] }
const initial = api.notificationFormInitial(row)
const rule = { id: ' exact <&中 ', topic: 'arn:aws:sns:east:RGW12345678901234567:topic', events: [], filters: [{ kind: 'S3Key', name: 'regex', value: ' a.* ' }, { kind: 'S3Tags', name: 'tag', value: '' }] }
const values = { ...initial, notification_draft: { ...initial.notification_draft, mode: 'create', rule }, confirm_notification: 'acknowledged' }
assert.deepEqual(api.notificationFormInput(values, row), { bucket_id: row.bucket_id, mode: 'create', rule, expected_document: row.document })
assert.equal(initial.notification_draft.mode, '')
assert.equal(initial.confirm_notification, undefined)
for (const invalid of [{ configured: false }, { notifications: null }, { notifications: [null] }, { kind: 'policy' }, { document: '' }]) assert.ok(api.notificationFormBlocked({ ...row, ...invalid }))
for (const invalid of [{ confirm_notification: undefined }, { bucket_id: 'other' }, { notification_draft: null }]) assert.throws(() => api.notificationFormInput({ ...values, ...invalid }, row))
for (const patch of [{ id: '' }, { topic: 'bad' }, { events: ['s3:ObjectRestore:*'] }, { filters: [{ kind: 'S3Key', name: 'invalid', value: 'x' }] }, { filters: [rule.filters[0], rule.filters[0]] }, { id: '\u0000' }, { id: '\ud800' }]) assert.throws(() => api.notificationFormInput({ ...values, notification_draft: { ...values.notification_draft, rule: { ...rule, ...patch } } }, row))
const existing = { ...row, notifications: [rule] }
assert.throws(() => api.notificationFormInput(values, existing))
const edit = { ...values, notification_draft: { mode: 'edit', selected: rule.id, rule, existing: [] } }
assert.equal(api.notificationFormInput(edit, existing).mode, 'edit')
assert.throws(() => api.notificationFormInput(edit, { ...existing, notifications: [rule, rule] }))
assert.throws(() => api.notificationFormInput({ ...edit, notification_draft: { ...edit.notification_draft, selected: 'wrong' } }, existing))
assert.throws(() => api.notificationFormInput(values, { ...row, document: 'x'.repeat(1024 * 1024) }), /上限/)
const normalized = api.notificationEditableRule({ ...rule, events: ['s3:ObjectLifecycle:Expiration:AbortMPU'] })
assert.equal(normalized.events[0], 's3:ObjectLifecycle:Expiration:AbortMultipartUpload')
normalized.filters[0].value = 'changed'
assert.equal(rule.filters[0].value, ' a.* ')
for (const warning of ['ObjectCreated:*', 'ObjectRemoved:*', '先删除', '空窗', 'GetTopicAttributes', '原子锁', '回滚', 'C++', '不证明投递']) assert.ok(api.notificationFormConfirmation(values, row).includes(warning))
const jsx = (type, props) => ({ type, props })
const editor = {}
new Function('exports', 'require', compile('../src/pages/object/RgwBucketNotificationEditor.tsx'))(editor, name => name.includes('rgwBucketNotificationForm') ? api : name === 'antd' ? { Alert: 'Alert', Button: 'Button', Input: Object.assign('Input', { TextArea: 'TextArea' }), Select: 'Select', Space: 'Space' } : { jsx, jsxs: jsx })
let changed
const draft = { ...values.notification_draft, existing: [rule] }
function flatten(node) { return !node || typeof node !== 'object' ? [] : Array.isArray(node) ? node.flatMap(flatten) : [node, ...flatten(node.props?.children)] }
const enabled = flatten(editor.RgwBucketNotificationEditor({ value: draft, onChange: value => { changed = value } }))
enabled.find(node => node.props?.['aria-label'] === '通知 ID').props.onChange({ target: { value: 'new-id' } })
assert.equal(changed.rule.id, 'new-id')
assert.equal(rule.id, ' exact <&中 ')
changed = undefined
const disabled = flatten(editor.RgwBucketNotificationEditor({ value: draft, disabled: true, onChange: value => { changed = value } }))
for (const node of disabled) {
  if (node.props?.onClick) { assert.equal(node.props.disabled, true); node.props.onClick() }
  if (node.props?.['aria-label'] === '通知 ID') { assert.equal(node.props.disabled, true); node.props.onChange({ target: { value: 'bad' } }) }
}
assert.equal(changed, undefined)
const pageText = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const page = ts.createSourceFile('pages.tsx', pageText, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let action
function visit(node) {
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(page) === 'title' && ts.isStringLiteral(p.initializer) && p.initializer.text === '创建或编辑 Bucket 通知')) {
    const code = ts.transpileModule(`const action = ${node.getText(page)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.React } }).outputText
    action = new Function(...Object.keys(api), 'React', 'RgwBucketNotificationEditor', `${code}; return action`)(...Object.values(api), { createElement: (type, props) => ({ type, props }) }, editor.RgwBucketNotificationEditor)
  }
  ts.forEachChild(node, visit)
}
visit(page)
assert.equal(action.path, '/rgw/bucket/notification')
assert.equal(action.method, 'POST')
assert.equal(action.visibleWhen(row), true)
assert.equal(action.visibleWhen({ kind: 'policy' }), false)
assert.equal(action.disabledWhen, api.notificationFormBlocked)
assert.equal(action.initialValues, api.notificationFormInitial)
assert.equal(action.confirmation, api.notificationFormConfirmation)
assert.deepEqual(action.buildBody(values, 42, row), { cluster_id: 42, ...api.notificationFormInput(values, row) })
assert.equal(action.fields.find(field => field.name === 'notification_draft').renderControl(true).props.disabled, true)
console.log('notification write form, editor scope and native-event checks passed')

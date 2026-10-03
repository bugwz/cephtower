import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
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

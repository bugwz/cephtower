import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
import './test-rgw-bucket-sync-policy.mjs'
const api = {}
const jsx = (type, props) => ({ type, props })
new Function('exports', 'require', ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwBucketReplication.tsx', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText)(api, name => name === 'antd' ? { Table: 'Table' } : { jsx, jsxs: jsx })
const rule = { id: '<rule>', status: 'Enabled', priority: '9007199254740993', destination_bucket: 'arn:aws:s3:::target' }
const data = { role: '', rules: [rule, { ...rule, status: 'Disabled', priority: null }] }
assert.deepEqual(api.bucketReplicationData(data), data)
for (const invalid of [null, {}, [], { ...data, role: null }, { role: '', rules: [null] }, { role: '', rules: [{ ...rule, priority: 1 }] }, { role: '', rules: [{ ...rule, destination_bucket: null }] }]) assert.equal(api.bucketReplicationData(invalid), undefined)
assert.match(api.bucketReplicationStatus('Enabled'), /不代表复制完成/)
assert.equal(api.bucketReplicationStatus('Disabled'), '规则停用')
for (const status of ['future', 'constructor', 'enabled']) assert.match(api.bucketReplicationStatus(status), /未知/)
assert.match(api.RgwBucketReplication({ value: null, configured: false }).props.children, /不存在.*不代表/)
assert.equal(api.RgwBucketReplication({ value: data, configured: false }).props.children, '复制配置不可用')
const table = api.RgwBucketReplication({ value: data, configured: true }).props.children.find(child => child.type === 'Table')
assert.equal(table.props.dataSource.length, 2)
assert.match(table.props.columns[2].render(rule.priority), /9007199254740993/)
assert.equal(table.props.columns[2].render(null), '未返回')
assert.match(table.props.columns[3].render(''), /原生空值/)
assert.equal(typeof table.props.columns[0].render('<script>'), 'string')
const empty = api.RgwBucketReplication({ value: { role: '', rules: [] }, configured: true })
assert.match(empty.props.children[2].props.locale.emptyText, /无 S3 复制规则.*不推断/)
console.log('bucket replication presentation tests passed')

const form = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketReplicationForm.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(form)
const row = { bucket_id: 'AGJ1Y2tldA', kind: 'replication', configured: true, document: '<ReplicationConfiguration><Role/></ReplicationConfiguration>' }
const values = { ...form.bucketReplicationFormInitial(row), confirm_replication: 'acknowledged' }
assert.deepEqual(form.bucketReplicationFormInput(values, row), { bucket_id: row.bucket_id, expected_document: row.document })
assert.equal(form.bucketReplicationFormInput(values, { ...row, configured: false, document: null }).expected_document, '')
assert.throws(() => form.bucketReplicationFormInput({ ...values, bucket_id: 'other' }, row), /不可更改/)
assert.throws(() => form.bucketReplicationFormInput({ ...values, confirm_replication: undefined }, row), /确认/)
for (const invalid of [{ ...row, kind: 'acl' }, { ...row, bucket_id: '' }, { ...row, configured: undefined }, { ...row, document: null }, { ...row, configured: false }]) assert.equal(typeof form.bucketReplicationFormBlocked(invalid), 'string')
assert.throws(() => form.bucketReplicationFormInput(values, { ...row, document: 'x'.repeat(1024 * 1024) }), /上限/)
const confirmation = form.bucketReplicationFormConfirmation(values, row)
for (const text of ['全部 S3', '上层', '不证明', '租户', 'ListBuckets', '原子锁', '回滚']) assert.ok(confirmation.includes(text))
const page = ts.createSourceFile('pages.tsx', readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8'), ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let action
function visit(node) {
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(page) === 'title' && ts.isStringLiteral(p.initializer) && p.initializer.text === '设置 Dashboard 桶复制规则')) {
    const source = ts.transpileModule(`const action = ${node.getText(page)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    action = new Function(...Object.keys(form), `${source}; return action`)(...Object.values(form))
  }
  ts.forEachChild(node, visit)
}
visit(page)
assert.equal(action.path, '/rgw/bucket/replication')
assert.equal(action.method, 'POST')
assert.equal(action.visibleWhen(row), true)
assert.equal(action.visibleWhen({ kind: 'acl' }), false)
assert.equal(action.disabledWhen, form.bucketReplicationFormBlocked)
assert.equal(action.initialValues, form.bucketReplicationFormInitial)
assert.equal(action.confirmation, form.bucketReplicationFormConfirmation)
assert.deepEqual(action.buildBody(values, 42, row), { cluster_id: 42, bucket_id: row.bucket_id, expected_document: row.document })
assert.equal(action.fields.find(f => f.name === 'bucket_id').readOnly, true)
console.log('bucket replication write form, snapshot and action checks passed')

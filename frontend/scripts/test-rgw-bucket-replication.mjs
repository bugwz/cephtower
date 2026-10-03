import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
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

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const exports = {}
new Function('exports', 'require', ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwBucketLifecycleRules.tsx', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText)(exports, name => name === 'antd' ? { Table: 'Table' } : { jsx: (type, props) => ({ type, props }), jsxs: (type, props) => ({ type, props }) })
const selector = { kind: 'Filter', and: true, prefix: ' <prefix> ', tags: [{ key: 'a', value: '1' }, { key: 'a', value: '' }], object_size_greater_than: '9007199254740993', object_size_less_than: null, archive_zone: true }
const actions = [
  { type: 'Expiration', fields: { Days: '90' } },
  { type: 'NoncurrentVersionExpiration', fields: { NoncurrentDays: '30', NewerNoncurrentVersions: '0' } },
  { type: 'AbortIncompleteMultipartUpload', fields: { DaysAfterInitiation: '7' } },
  { type: 'Transition', fields: { Days: '0', StorageClass: '<cold>' } },
  { type: 'Transition', fields: { Days: '60', StorageClass: 'archive' } },
  { type: 'NoncurrentVersionTransition', fields: { NoncurrentDays: '1', StorageClass: 'cold' } }
]
const rule = { id: '<rule>', status: 'Enabled', selector, actions }
assert.deepEqual(exports.bucketLifecycleRows([rule])[0].actions, actions)
for (const value of [undefined, {}, [null], [{ ...rule, status: 'unknown' }], [{ ...rule, selector: null }], [{ ...rule, selector: { ...selector, object_size_greater_than: 9007199254740992 } }], [{ ...rule, actions: [] }], [{ ...rule, actions: [{ type: 'unknown', fields: { Days: '1' } }] }], [{ ...rule, actions: [{ type: 'Transition', fields: { Days: 1 } }] }]]) assert.equal(exports.bucketLifecycleRows(value), undefined)
assert.equal(exports.RgwBucketLifecycleRules({ value: [], configured: false }).props.children, '未配置生命周期')
for (const input of [{ value: [], configured: true }, { value: [rule], configured: false }, { value: undefined, configured: true }]) assert.equal(exports.RgwBucketLifecycleRules(input).props.children, '生命周期数据不可用')
const text = exports.lifecycleSelectorText(selector)
assert.match(text, /Filter \/ And/)
assert.match(text, /9007199254740993/)
assert.match(text, /"a" = "1"\n标签："a" = ""/)
assert.match(text, /归档区域条件：存在/)
assert.match(exports.lifecycleSelectorText({ ...selector, kind: 'Prefix', and: false, tags: [], archive_zone: false, prefix: '', object_size_greater_than: null }), /旧式 Prefix\n前缀：""/)
assert.match(exports.lifecycleSelectorText({ ...selector, tags: [], archive_zone: false, prefix: null, object_size_greater_than: null }), /无显式过滤条件/)
const actionText = exports.lifecycleActionsText(actions)
for (const phrase of ['当前版本过期', '非当前版本过期', '终止未完成分段上传', '当前版本转换', '非当前版本转换', '目标存储类别："<cold>"', '天数："0"', '保留较新非当前版本数量："0"']) assert.ok(actionText.includes(phrase))
assert.match(exports.lifecycleActionsText([{ type: 'Expiration', fields: { ExpiredObjectDeleteMarker: 'false' } }]), /清除过期删除标记："false"/)
const table = exports.RgwBucketLifecycleRules({ value: [rule], configured: true })
assert.equal(table.props.columns.find(column => column.dataIndex === 'id').render('<rule>'), '"<rule>"')
assert.equal(table.props.columns.find(column => column.dataIndex === 'status').render('Disabled'), '禁用')
assert.equal(table.props.columns.find(column => column.dataIndex === 'selector').render(selector).props.children, text)
assert.equal(table.props.columns.find(column => column.dataIndex === 'actions').render(actions).props.children, actionText)
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.match(pages, /key: 'lifecycle_rules'.*row.kind === 'lifecycle'.*<RgwBucketLifecycleRules value=\{value\} configured=\{row.configured\}/)
console.log('bucket lifecycle rendering tests passed')

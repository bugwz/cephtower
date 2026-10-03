import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const helpers = {}
const compile = name => ts.transpileModule(readFileSync(new URL(`../src/pages/object/${name}`, import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText
new Function('exports', compile('rgwBucketLifecycleForm.ts'))(helpers)
const selector = { kind: 'Filter', and: false, prefix: null, tags: [], object_size_greater_than: null, object_size_less_than: null, archive_zone: false }
const actions = [
  { type: 'Expiration', fields: { Days: '90' } },
  { type: 'NoncurrentVersionExpiration', fields: { NoncurrentDays: '30', NewerNoncurrentVersions: '0' } },
  { type: 'AbortIncompleteMultipartUpload', fields: { DaysAfterInitiation: '7' } },
  { type: 'Transition', fields: { Days: '0', StorageClass: 'COLD' } },
  { type: 'NoncurrentVersionTransition', fields: { NoncurrentDays: '1', StorageClass: 'COLD' } }
]
const rule = { id: 'rule<&\r', status: 'Disabled', selector, actions }
const row = { bucket_id: 'AGJ1Y2tldA', kind: 'lifecycle', configured: true, lifecycle_rules: [rule] }
const values = helpers.lifecycleFormInitial(row)
values.lifecycle_draft.rules[0].actions[0].fields.Days = '91'
assert.equal(rule.actions[0].fields.Days, '90')
const input = helpers.lifecycleFormInput(helpers.lifecycleFormInitial(row), row)
assert.equal(input.document, readFileSync(new URL('../../backend/internal/integration/ceph/s3/testdata/lifecycle-editor.xml', import.meta.url), 'utf8').trimEnd())
assert.deepEqual(Object.keys(input), ['bucket_id', 'kind', 'document'])
assert.match(input.document, /<ID>rule&lt;&amp;&#13;<\/ID>/)
for (const action of actions) assert.ok(input.document.includes(`<${action.type}>`))
assert.match(helpers.lifecycleFormConfirmation(values, row), /整体替换.*永久删除.*不代表对象处理已完成/)
for (const changed of [{ bucket_id: 'other' }, { kind: 'cors' }]) assert.throws(() => helpers.lifecycleFormInput({ ...values, ...changed }, row))
assert.ok(helpers.lifecycleFormBlocked({ ...row, configured: false }))
assert.deepEqual(helpers.lifecycleFormInitial({ ...row, configured: false, lifecycle_rules: [] }).lifecycle_draft.rules, [])
const document = patch => helpers.lifecycleDocument({ rules: [{ ...rule, ...patch }] })
assert.match(document({ selector: { ...selector, and: true, prefix: '', tags: [{ key: '<a>', value: '' }, { key: '<a>', value: '2' }], object_size_greater_than: '9007199254740993', archive_zone: true }, actions: [actions[3]] }), /<And><Prefix><\/Prefix><Tag><Key>&lt;a&gt;<\/Key>/)
assert.match(document({ selector: { ...selector, kind: 'Prefix', prefix: '' } }), /<Prefix><\/Prefix>/)
assert.match(document({ actions: [{ type: 'Expiration', fields: { Date: '2030-01-01T00:00:00Z' } }] }), /<Date>2030-01-01T00:00:00Z<\/Date>/)
for (const patch of [
  { id: '\u0000' }, { id: '界'.repeat(86) }, { actions: [] },
  { selector: { ...selector, object_size_greater_than: '18446744073709551616' } },
  { selector: { ...selector, kind: 'Prefix', prefix: '', tags: [{ key: 'a', value: 'b' }] } },
  { selector: { ...selector, tags: [{ key: 'a', value: 'b' }] } },
  { actions: [actions[3], actions[3]] },
  { actions: [{ type: 'Expiration', fields: { Days: '1', Date: '2030' } }] },
  { actions: [{ type: 'Expiration', fields: { Days: '0' } }] },
  { actions: [{ type: 'Expiration', fields: { Days: '2147483648' } }] },
  { actions: [{ type: 'Expiration', fields: { ExpiredObjectDeleteMarker: 'false' } }] },
  { actions: [{ type: 'Expiration', fields: { ExpiredObjectDeleteMarker: 'yes' } }] },
  { actions: [{ type: 'Transition', fields: { Days: '1x', StorageClass: 'COLD' } }] },
  { actions: [{ type: 'Transition', fields: { Days: '0' } }] },
  { actions: [{ type: 'NoncurrentVersionExpiration', fields: {} }] },
  { actions: [{ type: 'Unknown', fields: {} }] }
]) assert.throws(() => document(patch))
assert.throws(() => helpers.lifecycleDocument({ rules: [rule, rule] }))

const editor = {}
new Function('exports', 'require', compile('RgwBucketLifecycleEditor.tsx'))(editor, name => name === './rgwBucketLifecycleForm' ? helpers : name === 'antd' ? { Alert: 'Alert', Button: 'Button', Checkbox: 'Checkbox', Input: { TextArea: 'TextArea' }, Select: 'Select', Space: 'Space' } : { jsx: (type, props) => ({ type, props }), jsxs: (type, props) => ({ type, props }) })
function nodes(node) { return Array.isArray(node) ? node.flatMap(nodes) : node && typeof node === 'object' ? [node, ...nodes(node.props?.children)] : [] }
let changed
const render = (draft, disabled = false) => nodes(editor.RgwBucketLifecycleEditor({ value: draft, disabled, onChange: value => { changed = value } }))
const find = (list, label) => list.find(node => node.props?.['aria-label'] === label)
let list = render({ rules: [rule] })
find(list, '规则 1 ID').props.onChange({ target: { value: 'updated' } })
assert.equal(changed.rules[0].id, 'updated')
find(list, '规则 1 状态').props.onChange('Enabled')
assert.equal(changed.rules[0].status, 'Enabled')
find(list, '规则 1 设置 object_size_greater_than').props.onChange({ target: { checked: true } })
assert.equal(changed.rules[0].selector.object_size_greater_than, '')
find(list, '规则 1 动作 4 Days').props.onChange({ target: { value: '2' } })
assert.equal(changed.rules[0].actions[3].fields.Days, '2')
find(list, '规则 1 动作 1 设置 Days').props.onChange({ target: { checked: false } })
assert.ok(!('Days' in changed.rules[0].actions[0].fields))
list.find(node => node.type === 'Button' && node.props.children === '添加标签').props.onClick()
assert.deepEqual(changed.rules[0].selector.tags, [{ key: '', value: '' }])
list.find(node => node.type === 'Button' && node.props.children === '当前版本转换').props.onClick()
assert.equal(changed.rules[0].actions.length, 6)
list.find(node => node.type === 'Button' && node.props.children === '移除动作').props.onClick()
assert.equal(changed.rules[0].actions.length, 4)
list.find(node => node.type === 'Button' && node.props.children === '移除规则').props.onClick()
assert.deepEqual(changed.rules, [])
render({ rules: [] }).find(node => node.type === 'Button' && node.props.children === '添加规则').props.onClick()
assert.equal(changed.rules[0].status, 'Disabled')
assert.deepEqual(changed.rules[0].actions, [])
changed = undefined
list = render({ rules: [rule] }, true)
for (const node of list.filter(node => ['Button', 'TextArea', 'Select', 'Checkbox'].includes(node.type))) assert.equal(node.props.disabled, true)
find(list, '规则 1 ID').props.onChange({ target: { value: 'blocked' } })
list.find(node => node.type === 'Button' && node.props.children === '添加规则').props.onClick()
assert.equal(changed, undefined)
console.log('bucket lifecycle form and interaction tests passed')

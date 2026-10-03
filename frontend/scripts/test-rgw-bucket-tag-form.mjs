import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const helpers = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketTagForm.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(helpers)
const entries = [{ key: ' same ', value: '' }, { key: ' same ', value: '&< >"\'\r\n\t😀' }]
const value = { entries }
const document = helpers.bucketTagFormDocument(value)
assert.equal(document, '<Tagging xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><TagSet><Tag><Key> same </Key><Value></Value></Tag><Tag><Key> same </Key><Value>&amp;&lt; &gt;&quot;&apos;&#13;\n\t😀</Value></Tag></TagSet></Tagging>')
assert.ok(helpers.bucketTagFormDocument({ entries: [] }).includes('<TagSet></TagSet>'))
const boundary = { key: '中'.repeat(42) + 'aa', value: '😀'.repeat(64) }
assert.ok(helpers.bucketTagFormDocument({ entries: Array.from({ length: 50 }, () => boundary) }))
for (const invalid of [null, {}, [], { entries: null }, { entries: [null] }, { entries: [{ key: '', value: '' }] }, { entries: [{ key: 'a', value: 1 }] }, { entries: [{ key: '中'.repeat(43), value: '' }] }, { entries: [{ key: 'a', value: '😀'.repeat(65) }] }, { entries: Array.from({ length: 51 }, () => boundary) }, ...['\0', '\b', '\uD800', '\uFFFE'].map(character => ({ entries: [{ key: 'a', value: character }] }))]) {
  assert.throws(() => helpers.bucketTagFormDocument(invalid))
}
const row = { bucket_id: 'AGJ1Y2tldA', kind: 'tagging', configured: true, tags: entries }
const initial = helpers.bucketTagFormInitial(row)
assert.deepEqual(initial.tag_set, value)
initial.tag_set.entries[0].key = 'changed'
assert.equal(entries[0].key, ' same ', 'form edits must not mutate the loaded row')
const values = helpers.bucketTagFormInitial(row)
assert.deepEqual(helpers.bucketTagFormInput(values, row), { bucket_id: row.bucket_id, kind: 'tagging', document })
assert.throws(() => helpers.bucketTagFormInput({ ...values, bucket_id: 'other' }, row))
assert.throws(() => helpers.bucketTagFormInput({ ...values, kind: 'policy' }, row))
const confirmation = helpers.bucketTagFormConfirmation(values, row)
assert.ok(confirmation.includes(row.bucket_id) && confirmation.includes('2 条') && confirmation.includes('整体替换'))
assert.ok(!confirmation.includes(entries[1].value))
const emptyRow = { ...row, configured: false, tags: [] }
assert.ok(helpers.bucketTagFormConfirmation(helpers.bucketTagFormInitial(emptyRow), emptyRow).includes('空标签集合，而不是删除标签属性'))
for (const invalid of [{}, { ...row, kind: 'policy' }, { ...row, configured: undefined }, { ...row, tags: undefined }, { ...row, configured: false }, { ...row, bucket_id: ' id' }]) assert.ok(helpers.bucketTagFormBlocked(invalid))

const component = {}
const componentSource = readFileSync(new URL('../src/pages/object/RgwBucketTagEditor.tsx', import.meta.url), 'utf8')
new Function('exports', 'require', ts.transpileModule(componentSource, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText)(component, name => {
  if (name === 'react/jsx-runtime') return { jsx: (type, props) => ({ type, props }), jsxs: (type, props) => ({ type, props }) }
  if (name === 'antd') return { Alert: 'alert', Button: 'button', Input: { TextArea: 'input' }, Space: 'space', Typography: { Text: 'text' } }
  if (name === './rgwBucketTagForm') return helpers
  throw new Error(name)
})
function nodes(node, type) {
  if (!node || typeof node !== 'object') return []
  if (Array.isArray(node)) return node.flatMap(child => nodes(child, type))
  return [...(node.type === type ? [node] : []), ...nodes(node.props?.children, type)]
}
const changes = []
const render = (disabled = false, formValue = value) => component.RgwBucketTagEditor({ value: formValue, disabled, id: 'tags', onChange: next => changes.push(next) })
const control = render()
assert.equal(changes.length, 0)
const inputs = nodes(control, 'input'), buttons = nodes(control, 'button')
assert.equal(inputs.length, 4)
assert.equal(inputs[0].props.value, ' same ')
assert.equal(inputs[1].props.value, '')
inputs[2].props.onChange({ target: { value: 'second' } })
assert.deepEqual(changes.pop().entries, [entries[0], { key: 'second', value: entries[1].value }])
inputs[1].props.onChange({ target: { value: 'new' } })
assert.deepEqual(changes.pop().entries, [{ key: entries[0].key, value: 'new' }, entries[1]])
buttons[0].props.onClick()
assert.deepEqual(changes.pop(), { entries: [entries[1]] })
buttons.find(button => button.props.children === '添加标签').props.onClick()
assert.deepEqual(changes.pop(), { entries: [...entries, { key: '', value: '' }] })
buttons.find(button => button.props.children === '清空表单条目').props.onClick()
assert.deepEqual(changes.pop(), { entries: [] })
for (const input of nodes(render(true), 'input')) { assert.equal(input.props.disabled, true); input.props.onChange({ target: { value: 'ignored' } }) }
for (const button of nodes(render(true), 'button')) { assert.equal(button.props.disabled, true); button.props.onClick() }
assert.equal(changes.length, 0)
const full = render(false, { entries: Array.from({ length: 50 }, () => boundary) })
const add = nodes(full, 'button').find(button => button.props.children === '添加标签')
assert.equal(add.props.disabled, true)
add.props.onClick()
assert.equal(changes.length, 0)
assert.equal(component.RgwBucketTagEditor({ value: undefined }).type, 'alert')
const external = readFileSync(new URL('../src/pages/ExternalListPage.tsx', import.meta.url), 'utf8')
assert.ok(external.includes('renderFormControl(field, submitting)'))
assert.ok(external.includes('field.renderControl(disabled || field.readOnly)'))
assert.ok(external.includes('action.visibleWhen(row)'))
assert.ok(external.includes('openForm(action, row)'))
console.log('Bucket tag forms preserve entries, encode safe native XML and disable draft changes during submission')

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/ResourceListPage.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('ResourceListPage.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const declaration = tree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name?.text === 'renderFormControl')
assert.ok(declaration)
const code = ts.transpileModule(declaration.getText(tree), { compilerOptions: { jsx: ts.JsxEmit.React, target: ts.ScriptTarget.ES2022 } }).outputText
const React = { createElement: (type, props) => ({ type, props }) }
const render = new Function('React', 'Select', `${code}; return renderFormControl`)(React, 'Select')
for (const multiple of [false, true]) {
  const field = { type: 'select', multiple, options: [{ label: 'SMB', value: 'smb' }], placeholder: '选择标签' }
  const control = render(field)
  assert.equal(control.type, 'Select')
  assert.equal(control.props.allowClear, true)
  assert.equal(control.props.mode, multiple ? 'multiple' : undefined)
  assert.deepEqual(control.props.options, field.options)
  assert.equal(control.props.placeholder, field.placeholder)
  assert.equal(render({ ...field, required: true }).props.allowClear, false)
  const readOnly = render({ ...field, readOnly: true })
  assert.equal(readOnly.props.allowClear, false)
  assert.equal(readOnly.props.disabled, true)
}
console.log('Resource form select clearing checks passed')

const blockSource = readFileSync(new URL('../src/pages/block/pages.tsx', import.meta.url), 'utf8')
const blockTree = ts.createSourceFile('pages.tsx', blockSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let scheduleNode
function findSchedule(node) {
  if (ts.isObjectLiteralExpression(node) && node.properties.some((property) => ts.isPropertyAssignment(property) && property.name.getText(blockTree) === 'path' && property.initializer.getText(blockTree) === "'/rbd/mirroring/schedule'")) scheduleNode = node
  ts.forEachChild(node, findSchedule)
}
findSchedule(blockTree)
assert.ok(scheduleNode)
const scheduleCode = ts.transpileModule(`const schedule = ${scheduleNode.getText(blockTree)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const schedule = new Function(`${scheduleCode}; return schedule`)()
const poolRow = { pool: 'images' }
const addValues = { action: 'mirror-schedule-add', interval: '12h', remove_interval: '1d', start_time: '01:00' }
assert.deepEqual(schedule.buildBody(addValues, 7, poolRow), { cluster_id: 7, pool: 'images', action: 'mirror-schedule-add', interval: '12h', start_time: '01:00' })
const removeValues = { ...addValues, action: 'mirror-schedule-remove', remove_interval: '' }
assert.deepEqual(schedule.buildBody(removeValues, 7, poolRow), { cluster_id: 7, pool: 'images', action: 'mirror-schedule-remove' })
assert.match(schedule.confirmation(removeValues, poolRow), /全部池级调度/)
assert.deepEqual(schedule.buildBody({ ...removeValues, remove_interval: '1d' }, 7, poolRow), { cluster_id: 7, pool: 'images', action: 'mirror-schedule-remove', interval: '1d', start_time: '01:00' })
const intervalField = schedule.fields.find((field) => field.name === 'interval')
assert.equal(intervalField.required, true)
assert.equal(intervalField.visibleWhen(addValues), true)
assert.equal(intervalField.visibleWhen(removeValues), false)
assert.equal(schedule.fields.find((field) => field.name === 'start_time').visibleWhen(removeValues), false)
console.log('Pool mirror schedule form scope checks passed')

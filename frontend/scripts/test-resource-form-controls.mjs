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

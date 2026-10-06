import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8')
const file = ts.createSourceFile('pages.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const node = file.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'deviceUsage')
const code = ts.transpileModule(node.getText(file), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const usage = new Function('deviceReasonValues', 'isUsedDeviceReason', 'readableDeviceReason', `${code}; return deviceUsage`)(v => v ?? [], v => v === 'LVM', v => v)
assert.equal(usage({ available: true }).state, 'available')
assert.equal(usage({ available: false }).state, 'unavailable')
assert.equal(usage({ available: false, rejected_reasons: ['LVM'] }).state, 'used')
for (const available of [undefined, null, 'false', 0, {}]) {
  const result = usage({ available, rejected_reasons: ['LVM'] })
  assert.equal(result.state, 'unknown')
  assert.deepEqual(result.notes, ['LVM'])
}
console.log('Device availability retains unknown native states and rejection diagnostics')

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('page.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const component = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'ReweightForm')
const submit = component.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'submit')
const js = ts.transpileModule(submit.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const scenario of ['ok', 'switched-before', 'switched-after', 'busy', 'invalid', 'failure']) {
  let current = scenario !== 'switched-before'
  const calls = [], running = { current: scenario === 'busy' }
  const env = { running, isCurrent: () => current, clusterId: 7, osdID: '12',
    setSubmitting: () => {}, message: { error: () => calls.push('error'), success: () => calls.push('success') },
    operationMutation: { run: fn => fn() },
    reweightOSD: async (...args) => { calls.push(args); if (scenario === 'switched-after') current = false; if (scenario === 'failure') throw new Error('failed') },
    refreshResource: async body => calls.push(body), refresh: async () => calls.push('read') }
  const run = new Function(...Object.keys(env), `${js}; return submit`)(...Object.values(env))
  const result = run({ weight: scenario === 'invalid' ? NaN : 0 })
  if (scenario === 'failure') await assert.rejects(result, /failed/); else await result
  if (['switched-before', 'busy'].includes(scenario)) assert.deepEqual(calls, [])
  if (scenario === 'invalid') assert.deepEqual(calls, ['error'])
  if (scenario === 'ok') assert.deepEqual(calls, [[7, '12', 0], 'success', { clusterId: 7, kind: 'osd' }, 'read'])
  if (scenario === 'switched-after') assert.deepEqual(calls, [[7, '12', 0]])
}
assert.ok(!component.getText(tree).includes('requiredClusterId'))
assert.ok(!component.getText(tree).includes('Modal.destroyAll'))
const initial = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'osdReweightInitial')
const initialJS = ts.transpileModule(initial.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const initialValue = new Function(`${initialJS}; return osdReweightInitial`)()
for (const value of [0, 0.12345, 1]) assert.equal(initialValue(value), value)
for (const value of [undefined, null, -1, 2, NaN, Infinity, '0.5']) assert.equal(initialValue(value), undefined)
assert.ok(component.getText(tree).includes('weight: osdReweightInitial(currentWeight)'))
assert.ok(!component.getText(tree).includes('precision={2}'))

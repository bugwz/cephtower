import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/cluster/HostPage.tsx', import.meta.url), 'utf8')
const file = ts.createSourceFile('HostPage.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const page = file.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'HostPage')
const node = page.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'submitHost')
const code = ts.transpileModule(node.getText(file), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const stale of [false, true]) {
  let finish
  const pending = new Promise(resolve => { finish = resolve })
  const calls = [], results = [], messages = [], refreshes = []
  const scope = { active: true }
  const env = {
    selectedClusterId: 7, submitting: false, createRunning: { current: false }, createAttempted: false, createScope: scope,
    setSubmitting() {}, setCreateAttempted() {}, setCreateResults: value => results.push(value),
    expandHostnames: () => ['node1', 'node2'], form: { setFields() {} },
    operationMutation: { run: fn => fn() }, setFormOpen() {},
    mutateResource: async (path, method, body) => { calls.push(body); if (body.hostname === 'node1') throw new Error('unconfirmed'); await pending },
    message: { success: value => messages.push(value), warning: value => messages.push(value) },
    refresh: async () => refreshes.push(true)
  }
  const submit = new Function(...Object.keys(env), `${code}; return submitHost`)(...Object.values(env))
  const work = submit({ hostname: 'node[1-2]', labels: ['osd'], maintenance: true })
  await Promise.resolve()
  await submit({ hostname: 'node[1-2]' })
  assert.equal(calls.length, 2)
  assert.equal(results.length, 0, 'must wait for the remaining request after failure')
  if (stale) scope.active = false
  finish()
  await work
  assert.equal(env.createRunning.current, false)
  assert.equal(results.length, stale ? 0 : 1)
  assert.equal(refreshes.length, stale ? 0 : 1)
  if (!stale) {
    assert.match(results[0][0], /node1.*未确认/)
    assert.match(results[0][1], /node2.*已确认/)
    assert.match(messages[0], /不要直接重复提交/)
  }
}
assert.match(source, /okButtonProps=\{\{ disabled: createAttempted \}\}/)
console.log('Host batches wait for all outcomes and suppress stale-scope results')

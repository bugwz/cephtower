import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/hooks/useMutationOperation.ts', import.meta.url), 'utf8')
const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS } }).outputText
for (const scenario of ['success', 'failure', 'synchronous-throw']) {
  const exports = {}, loading = [], messages = [], ref = { current: false }, lifetime = { current: { active: true, generation: 0 } }
  new Function('exports', 'require', code)(exports, name => name === 'react' ? { useRef: initial => typeof initial === 'boolean' ? ref : lifetime, useEffect() {}, useState: () => [false, value => loading.push(value)] } : { message: { success: value => messages.push(value) } })
  const oldRender = exports.useMutationOperation()
  let release, calls = 0
  const gate = new Promise(resolve => { release = resolve })
  const first = oldRender.run(async () => { calls++; await gate; if (scenario === 'failure') throw new Error('native failure'); return 'result' }, false)
  assert.equal(ref.current, true)
  await assert.rejects(oldRender.run(async () => { calls++ }), /已有操作正在执行/)
  const nextRender = exports.useMutationOperation()
  await assert.rejects(nextRender.run(async () => { calls++ }), /已有操作正在执行/)
  assert.equal(calls, 1)
  assert.deepEqual(loading, [true])
  release()
  if (scenario === 'failure') await assert.rejects(first, /native failure/)
  else assert.equal(await first, 'result')
  assert.equal(ref.current, false)
  assert.deepEqual(loading, [true, false])
  if (scenario === 'synchronous-throw') {
    await assert.rejects(nextRender.run(() => { throw new Error('sync failure') }), /sync failure/)
    assert.equal(ref.current, false)
  }
  assert.equal(await nextRender.run(async () => 'next', '完成'), 'next')
  assert.deepEqual(messages, ['完成'])
  assert.equal(ref.current, false)
}
console.log('Mutation hook synchronously excludes duplicate calls across render closures and releases failures')

for (const scenario of ['before', 'during', 'failure', 'strict-replay']) {
  const exports = {}, loading = [], messages = [], ref = { current: false }, lifetime = { current: { active: true, generation: 0 } }
  let effect
  new Function('exports', 'require', code)(exports, name => name === 'react' ? { useRef: initial => typeof initial === 'boolean' ? ref : lifetime, useEffect: setup => { effect = setup }, useState: () => [false, value => loading.push(value)] } : { message: { success: value => messages.push(value) } })
  const operation = exports.useMutationOperation()
  const cleanup = effect()
  let release, calls = 0
  const gate = new Promise(resolve => { release = resolve })
  if (scenario === 'before') {
    cleanup()
    await assert.rejects(operation.run(async () => { calls++ }), /操作页面已关闭/)
    assert.equal(calls, 0)
    assert.deepEqual(loading, [])
    continue
  }
  const pending = operation.run(async () => { calls++; await gate; if (scenario === 'failure') throw new Error('failed'); return 'completed' })
  cleanup()
  if (scenario === 'strict-replay') effect()
  release()
  if (scenario === 'failure') await assert.rejects(pending, /failed/)
  else assert.equal(await pending, 'completed')
  assert.deepEqual(messages, [])
  assert.deepEqual(loading, scenario === 'strict-replay' ? [true, false] : [true])
  assert.equal(ref.current, false)
  if (scenario === 'strict-replay') {
    await operation.run(async () => 'new operation')
    assert.equal(messages.length, 1)
  }
}
console.log('Mutation lifetime rejects closed-page calls and suppresses stale success notifications')

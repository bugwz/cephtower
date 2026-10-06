import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/hooks/useMutationOperation.ts', import.meta.url), 'utf8')
const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS } }).outputText
for (const scenario of ['success', 'failure', 'synchronous-throw']) {
  const exports = {}, loading = [], messages = [], ref = { current: false }
  new Function('exports', 'require', code)(exports, name => name === 'react' ? { useRef: () => ref, useState: () => [false, value => loading.push(value)] } : { message: { success: value => messages.push(value) } })
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

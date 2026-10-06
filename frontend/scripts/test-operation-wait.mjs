import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/api/resource.ts', import.meta.url), 'utf8')
const file = ts.createSourceFile('resource.ts', source, ts.ScriptTarget.Latest, true)
const node = file.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'waitForOperation')
const code = ts.transpileModule(node.getText(file), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const base = { cluster_id: 7, operation_id: 8, status: 'queued' }
class ApiRequestError extends Error { constructor(message, status, code) { super(message); this.code = code } }
for (const response of [null, {}, { ...base, status: 'cancelled' }, { ...base, status: 'succeeded', cluster_id: 9 }, { ...base, status: 'succeeded', operation_id: 9 }]) {
  const calls = [], notifications = []
  const env = { ApiRequestError, notifyApiError: e => notifications.push(e), toApiErrorDetail: e => e, wait: async () => {}, request: async (...args) => { calls.push(args); return response }, jsonInit: (_method, body) => body }
  const run = new Function(...Object.keys(env), `${code}; return waitForOperation`)(...Object.values(env))
  await assert.rejects(run(base), e => e.code === 'operation_response_invalid')
  assert.equal(calls.length, 1)
  assert.deepEqual(calls[0][1], { cluster_id: 7, operation_id: 8 })
  assert.equal(notifications.length, 1)
}
for (const status of ['succeeded', 'failed']) {
  const env = { ApiRequestError, notifyApiError() {}, toApiErrorDetail: e => e, wait: async () => {}, request: async () => ({ ...base, status, result: { details: 'confirmed' }, error_code: 'native_failure' }), jsonInit: () => ({}) }
  const run = new Function(...Object.keys(env), `${code}; return waitForOperation`)(...Object.values(env))
  if (status === 'succeeded') assert.deepEqual(await run(base), { details: 'confirmed' })
  else await assert.rejects(run(base), e => e.code === 'native_failure')
  await assert.rejects(run({ ...base, operation_id: Number.MAX_SAFE_INTEGER + 1 }), e => e.code === 'operation_response_invalid')
}
console.log('Operation polling validates terminal state and stable task identity')

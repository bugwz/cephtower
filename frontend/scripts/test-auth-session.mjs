import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/api/auth.ts', import.meta.url), 'utf8')
const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS } }).outputText
const exports = {}
let identity = { id: 7, username: 'operator', display_name: 'Operator', role: 'operator', status: 'active' }
let failure, requests = 0
const client = {
  async request(path, options) {
    requests++
    assert.equal(path, '/auth/me')
    assert.equal(options.method, 'GET')
    if (failure) throw failure
    return identity
  },
}
new Function('exports', 'require', 'localStorage', code)(exports, () => client, {
  getItem() { throw new Error('identity must not come from local storage') },
  setItem() { throw new Error('identity must not be cached') },
})
assert.equal((await exports.currentUser()).id, 7)
identity = { ...identity, display_name: 'Updated name' }
assert.equal((await exports.currentUser()).display_name, 'Updated name')
failure = new Error('session expired')
await assert.rejects(exports.currentUser(), /session expired/)
assert.equal(requests, 3)
console.log('Session restoration uses the token identity, refreshes changes and rejects expired sessions')

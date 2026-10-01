import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

function load(path, require = () => { throw new Error('unexpected import') }) {
  const source = readFileSync(new URL(path, import.meta.url), 'utf8')
  const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText
  const exports = {}
  new Function('exports', 'require', compiled)(exports, require)
  return exports
}

const nativeTime = '2020-01-01 01:02:03'
const cacheTime = '2026-10-01T00:00:00Z'
let data = { ceph_created_at: nativeTime }
const resource = load('../src/api/resource.ts', (path) => {
  assert.equal(path, './client')
  return {
    jsonInit: (method, body) => ({ method, body }),
    request: async () => ({ items: [{ name: 'volume', created_at: cacheTime, data }] })
  }
})
const time = load('../src/utils/time.ts')
const result = await resource.listResource('/filesystem/subvolumes', 1)
assert.equal(result.items[0].ceph_created_at, nativeTime)
assert.equal(result.items[0].created_at, cacheTime)
assert.equal(time.formatDateTime(result.items[0].ceph_created_at), nativeTime)
assert.equal(time.isDateTimeField('ceph_created_at'), true)
data = {}
const missing = await resource.listResource('/filesystem/subvolumes', 1)
assert.equal(time.formatDateTime(missing.items[0].ceph_created_at), '-')
assert.equal(missing.items[0].created_at, cacheTime)

// The three CephFS lists must select native time, never cache metadata.
const pages = readFileSync(new URL('../src/pages/file/pages.tsx', import.meta.url), 'utf8')
for (const name of ['subvolumeGroups', 'subvolumes', 'cephfsSnapshots']) {
  const start = pages.indexOf(`  ${name}: {`)
  assert.ok(start >= 0)
  const end = pages.indexOf('\n  },', start)
  const config = pages.slice(start, end)
  assert.match(config, /key: 'ceph_created_at'.*formatDateTime\(value\)/)
  assert.doesNotMatch(config, /key: 'created_at'/)
  if (name === 'subvolumes') assert.match(config, /key: 'bytes_pcent'/)
}
console.log('CephFS native creation time checks passed')

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/cluster/hostStorageCapacity.ts', import.meta.url), 'utf8')
const exports = {}
new Function('exports', ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(exports)
const capacity = exports.hostStorageCapacity
assert.equal(capacity([]), '0 B')
assert.equal(capacity([{ size_bytes: 0 }]), '0 B')
assert.equal(capacity([{ size_bytes: 512 }, { size_bytes: '512' }]), '约 1.00 KiB（1024 B）')
assert.match(capacity([{ size_bytes: '18446744073709551615' }, { size_bytes: '18446744073709551615' }]), /36893488147419103230 B/)
for (const value of [undefined, null, -1, 0.1, {}, '', '01', '1e3', Number.MAX_SAFE_INTEGER + 1, '18446744073709551616']) {
  assert.match(capacity([{ size_bytes: 1024 }, { size_bytes: value }]), /未知/)
}
const hostSource = readFileSync(new URL('../src/pages/cluster/HostPage.tsx', import.meta.url), 'utf8')
assert.ok(hostSource.includes('storage_display: hostStorageCapacity(hostDevices)'))
const clusterSource = readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8')
assert.ok(clusterSource.includes('size_display: hostStorageCapacity([{ size_bytes: row.size_bytes ?? row.size }])'))
console.log('Host storage totals retain exact bytes and reject incomplete capacity inputs')

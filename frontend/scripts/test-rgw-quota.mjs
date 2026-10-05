import assert from 'node:assert/strict'
import './test-rgw-realm-token.mjs'
import './test-rgw-realm-delete.mjs'
import './test-rgw-zonegroup-delete.mjs'
import './test-rgw-sync-status.mjs'
import './test-rgw-topology.mjs'
import './test-rgw-zone-pools.mjs'
import './test-rgw-realm-transfer.mjs'
import './test-rgw-realm-import.mjs'
import './test-rgw-realm-setup.mjs'
import './test-rgw-realm-migration.mjs'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const exports = {}
const code = ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwQuotaDetails.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText
new Function('exports', code)(exports)
const details = exports.rgwQuotaDetails
assert.deepEqual(details({ enabled: true, max_size: 0, max_objects: 0 }), { state: '已启用', size: '0', objects: '0' })
assert.deepEqual(details({ enabled: true, max_size: -1, max_objects: -2 }), { state: '已启用', size: '无限制', objects: '无限制' })
assert.equal(details({ enabled: true, max_size: 4096, max_objects: 20 }).size, '4096')
assert.deepEqual(details({ enabled: false, max_size: 4096 }), { state: '未启用' })
for (const value of [null, undefined, [], '']) assert.equal(details(value).state, '配额信息不可用')
for (const enabled of [undefined, null, 'false', 0, 1]) assert.equal(details({ enabled }).state, '配额启用状态未知')
for (const value of [undefined, null, '1', NaN, Infinity, 0.5, Number.MAX_SAFE_INTEGER + 1]) assert.equal(details({ enabled: true, max_size: value }).size, '未返回或超出精确显示范围')
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.equal(pages.match(/<RgwQuota value=\{value\} \/>/g).length, 1)
const accountDetails = readFileSync(new URL('../src/pages/object/RgwAccountDetails.tsx', import.meta.url), 'utf8')
assert.ok(accountDetails.includes('<RgwQuota value={row.quota} />'))
assert.ok(accountDetails.includes('<RgwQuota value={row.bucket_quota} />'))
console.log('RGW quota display preserves enabled, unlimited, zero and unavailable states')
const form = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwQuotaForm.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(form)
assert.deepEqual(form.rgwQuotaInitial(null), { enabled: undefined, max_size: undefined, max_objects: undefined })
assert.deepEqual(form.rgwQuotaInitial({ enabled: false, max_size: -1, max_objects: 0 }), { enabled: 'disable', max_size: -1, max_objects: 0 })
for (const enabled of ['enable', 'disable']) {
  assert.deepEqual(form.rgwQuotaInput({ enabled, max_size: -1, max_objects: 0 }), { enabled: enabled === 'enable', max_size: -1, max_objects: 0 })
}
for (const value of [null, undefined, '', '0', false, -2, 0.5, NaN, Infinity, Number.MAX_SAFE_INTEGER + 1]) {
  for (const key of ['max_size', 'max_objects']) {
    assert.throws(() => form.rgwQuotaInput({ enabled: 'enable', max_size: 0, max_objects: 0, [key]: value }))
    assert.equal(form.rgwQuotaInitial({ [key]: value })[key], undefined)
  }
}
for (const enabled of [null, undefined, true, false, 'false']) assert.throws(() => form.rgwQuotaInput({ enabled, max_size: 0, max_objects: 0 }))
assert.equal(pages.match(/\.\.\.rgwQuotaInput\(values\)/g).length, 3)
assert.equal(pages.match(/initialValues: \(row\) => rgwQuotaInitial/g).length, 3)

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const exports = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwUserFlags.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
assert.equal(exports.rgwUserSuspension(0), '未暂停')
for (const value of [1, 2, 255]) assert.equal(exports.rgwUserSuspension(value), '已暂停')
for (const value of [null, undefined, true, false, '1', -1, 256, 0.5, NaN]) assert.equal(exports.rgwUserSuspension(value), '暂停状态未知')
assert.equal(exports.rgwUserBooleanFlag(true), '是')
assert.equal(exports.rgwUserBooleanFlag(false), '否')
for (const value of [undefined, null, 0, 1, 'true']) assert.equal(exports.rgwUserBooleanFlag(value), '未知')
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.equal(pages.match(/render: rgwUserSuspension/g).length, 1)
assert.equal(pages.match(/render: rgwUserBooleanFlag/g).length, 2)
console.log('RGW user flags preserve native integer and boolean representations')
const identity = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwUserIdentity.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(identity)
assert.equal(identity.rgwIdentityText('', '默认租户'), '默认租户')
assert.equal(identity.rgwIdentityText('future-type', '空'), 'future-type')
for (const value of [null, undefined, 0, false, []]) assert.equal(identity.rgwIdentityText(value, '空'), '未返回或格式无效')
assert.deepEqual(identity.rgwIdentityList([]), [])
assert.deepEqual(identity.rgwIdentityList(['mfa:one', 'group,id']), ['mfa:one', 'group,id'])
for (const value of [null, undefined, {}, [null], [''], [1]]) assert.equal(identity.rgwIdentityList(value), undefined)
assert.ok(pages.includes('<RgwUserIdentityDetails row={row} />'))
assert.ok(pages.includes('<RgwUserPlacementDetails row={row} />'))
const components = readFileSync(new URL('../src/pages/object/RgwUserIdentityDetails.tsx', import.meta.url), 'utf8')
assert.ok(components.includes("rgwIdentityText(row.default_placement, '未显式设置')"))
assert.ok(components.includes("rgwIdentityText(row.default_storage_class, '未显式设置')"))
assert.ok(components.includes('<Identifiers value={row.placement_tags} />'))
assert.equal(identity.rgwIdentityText('', '未显式设置'), '未显式设置')
assert.equal(identity.rgwIdentityText('custom-archive', '未显式设置'), 'custom-archive')
assert.deepEqual(identity.rgwIdentityList(['placement,tag', 'cold']), ['placement,tag', 'cold'])
const tags = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwUserTags.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(tags)
assert.deepEqual(tags.rgwUserTags([]), [])
assert.deepEqual(tags.rgwUserTags([{ key: 'team', val: 'one' }, { key: 'team', val: '' }, { key: '', val: '<tag>' }]), [
  { id: 0, key: 'team', value: 'one' }, { id: 1, key: 'team', value: '' }, { id: 2, key: '', value: '<tag>' }
])
for (const value of [undefined, null, {}, [null], [{ Key: 'role', Value: 'wrong schema' }], [{ key: 'a', val: 0 }]]) assert.equal(tags.rgwUserTags(value), undefined)
assert.ok(components.includes('<RgwUserTagsTable value={row.tags} />'))
assert.ok(components.includes("rgwIdentityText(row.create_date, '未提供时间')"))
assert.equal(identity.rgwIdentityText('2026-10-03T12:34:56.123456789Z', '未提供时间'), '2026-10-03T12:34:56.123456789Z')

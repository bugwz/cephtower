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
const userDetailsSource = readFileSync(new URL('../src/pages/object/RgwUserDetails.tsx', import.meta.url), 'utf8')
const placementForm = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwUserPlacementForm.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(placementForm)
for (const storage of [undefined, null, '', 'STANDARD', 'ARCHIVE']) assert.deepEqual(placementForm.rgwUserPlacementInput({ default_placement: 'custom', default_storage_class: storage }), { default_placement: 'custom', default_storage_class: storage ?? '' })
for (const value of [undefined, null, '', false, 1, ' ', 'bad\nname']) assert.throws(() => placementForm.rgwUserPlacementInput({ default_placement: value }))
for (const value of [false, 1, ' ', 'bad\nclass']) assert.throws(() => placementForm.rgwUserPlacementInput({ default_placement: 'custom', default_storage_class: value }))
assert.ok(pages.includes('...rgwUserPlacementInput(values)'))
assert.ok(pages.includes('detailContent: (row) => <RgwUserDetails row={row} />'))
const userDetailsExports = {}
new Function('exports', 'require', ts.transpileModule(userDetailsSource, { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX } }).outputText)(userDetailsExports, (name) => {
  if (name === 'react/jsx-runtime') return { jsx: (type, props) => ({ type, props }) }
  if (name === 'antd') return { Tabs: 'Tabs' }
  if (name === './RgwUserIdentityDetails') return { RgwUserIdentityDetails: 'Identity', RgwUserPlacementDetails: 'Placement' }
  throw new Error(`unexpected import ${name}`)
})
for (const row of [{}, { uid: 'tenant$user', account_id: 'RGW123', tags: [], placement_tags: ['archive'] }]) {
  const view = userDetailsExports.RgwUserDetails({ row })
  assert.equal(view.type, 'Tabs')
  assert.deepEqual(view.props.items.map(item => item.key), ['identity', 'placement'])
  assert.equal(view.props.items[0].children.type, 'Identity')
  assert.equal(view.props.items[1].children.type, 'Placement')
  for (const item of view.props.items) assert.equal(item.children.props.row, row)
}
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
assert.ok(components.includes('<Identifiers value={row.managed_user_policies} />'))
assert.ok(components.includes("row.account_id === '' || row.type === 'root' ? '不适用'"))
assert.deepEqual(identity.rgwIdentityList(['arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess']), ['arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess'])
const policy = {}
new Function('exports', 'require', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwUserPolicy.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(policy, () => identity)
const arn = 'arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess'
const accountUser = { account_id: 'RGW123', type: 'rgw', managed_user_policies: [arn] }
assert.deepEqual(policy.rgwUserPolicyOptions(accountUser), [{ label: arn, value: arn }])
assert.deepEqual(policy.rgwUserPolicyInput({ action: 'detach', existing_policy: arn, policy_arn: 'stale' }, accountUser), { action: 'detach', policy_arn: arn })
assert.deepEqual(policy.rgwUserPolicyInput({ action: 'attach', policy_arn: arn }, { ...accountUser, managed_user_policies: [] }), { action: 'attach', policy_arn: arn })
assert.throws(() => policy.rgwUserPolicyInput({ action: 'attach', policy_arn: arn }, accountUser))
assert.throws(() => policy.rgwUserPolicyInput({ action: 'detach', existing_policy: arn }, { ...accountUser, managed_user_policies: undefined }))
assert.throws(() => policy.rgwUserPolicyInput({ action: 'detach', existing_policy: arn }, { ...accountUser, managed_user_policies: [] }))
for (const row of [undefined, {}, { account_id: '', type: 'rgw' }, { account_id: 'RGW123', type: 'root' }]) {
  assert.ok(policy.rgwUserPolicyBlocked(row))
  assert.throws(() => policy.rgwUserPolicyInput({ action: 'attach', policy_arn: arn }, row))
}
for (const value of ['', '--policy', arn + '\n', arn + ';other', null]) assert.throws(() => policy.rgwUserPolicyInput({ action: 'attach', policy_arn: value }, accountUser))
assert.ok(pages.includes("path: '/rgw/user/policy'"))
assert.ok(pages.includes('disabledWhen: rgwUserPolicyBlocked'))
assert.ok(pages.includes('...rgwUserPolicyInput(values, row)'))
const roleViews = {}
const jsx = (type, props) => ({ type, props })
new Function('exports', 'require', ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwRolePolicyDetails.tsx', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX } }).outputText)(roleViews, (name) => name === 'react/jsx-runtime' ? { jsx, jsxs: jsx } : name === './rgwUserIdentity' ? identity : {})
const managed = roleViews.RgwRoleManagedPolicies
assert.ok(managed({ value: undefined }).props.children.includes('省略'))
assert.equal(managed({ value: [] }).props.children, '托管策略列表为空')
for (const value of [null, {}, [null], [''], [{ PolicyArn: arn }]]) assert.equal(managed({ value }).props.children, '托管策略格式无效')
const renderedPolicies = managed({ value: [arn, 'arn:aws:iam::aws:policy/path/CustomPolicy'] })
assert.equal(renderedPolicies.type, 'ul')
assert.equal(renderedPolicies.props.style.overflowWrap, 'anywhere')
assert.deepEqual(renderedPolicies.props.children.map(item => item.props.children), [arn, 'arn:aws:iam::aws:policy/path/CustomPolicy'])
assert.ok(pages.includes("title: '直接关联的托管策略 ARN', ellipsis: false, render: (value) => <RgwRoleManagedPolicies value={value} />"))

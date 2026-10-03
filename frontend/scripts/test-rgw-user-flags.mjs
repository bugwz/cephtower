import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
import './test-rgw-user-account.mjs'
import './test-rgw-user-keys.mjs'
import './test-rgw-user-account-root.mjs'
import './test-rgw-user-account-migration.mjs'
import './test-rgw-migration-account-options.mjs'
import './test-rgw-user-create-account.mjs'
import './test-rgw-user-subuser.mjs'
const exports = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwUserFlags.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
assert.equal(exports.rgwUserSuspension(0), '未暂停')
for (const value of [1, 2, 255]) assert.equal(exports.rgwUserSuspension(value), '已暂停')
for (const value of [null, undefined, true, false, '1', -1, 256, 0.5, NaN]) assert.equal(exports.rgwUserSuspension(value), '暂停状态未知')
assert.equal(exports.rgwUserBooleanFlag(true), '是')
assert.equal(exports.rgwUserBooleanFlag(false), '否')
for (const value of [undefined, null, 0, 1, 'true']) assert.equal(exports.rgwUserBooleanFlag(value), '未知')
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const operationMask = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwUserOperationMask.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(operationMask)
assert.equal(operationMask.rgwUserOperationMaskOptions.length, 7)
for (const mask of ['read', 'write', 'delete', 'read,write', 'read,delete', 'write,delete', 'read,write,delete']) assert.deepEqual(operationMask.rgwUserOperationMaskInput({ op_mask: mask }), { op_mask: mask })
for (const mask of [undefined, null, false, '', '*', 'none', 'read,read', 'read\n', 'read,unknown', ['read']]) assert.throws(() => operationMask.rgwUserOperationMaskInput({ op_mask: mask }))
assert.ok(pages.includes('options: rgwUserOperationMaskOptions'))
assert.ok(pages.includes('...rgwUserOperationMaskInput(values)'))
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
for (const tags of ['fast', 'fast,archive', 'space tag, raw ', '--option=value']) assert.deepEqual(placementForm.rgwUserPlacementTagsInput({ placement_tags_csv: tags }), { placement_tags_csv: tags })
for (const tags of [undefined, null, false, [], '', ' ', ',', 'a,', ',a', 'a,,b', 'a, ,b', 'a\nb', 'a\0b']) assert.throws(() => placementForm.rgwUserPlacementTagsInput({ placement_tags_csv: tags }))
assert.ok(pages.includes('...rgwUserPlacementTagsInput(values)'))
assert.ok(pages.includes('JSON.stringify(rgwUserPlacementTagsInput(values).placement_tags_csv)'))
const placementSource = ts.createSourceFile('pages.tsx', pages, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const userIDFunction = placementSource.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === 'userId')
assert.ok(userIDFunction)
const userIDExports = {}
new Function('exports', ts.transpileModule('export ' + userIDFunction.getText(placementSource), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(userIDExports)
for (const uid of ['user', 'tenant$user', 'tenant$namespace$user', '$namespace$user']) assert.equal(userIDExports.userId({ uid, user_id: 'wrong', natural_key: 'wrong' }), uid)
for (const uid of [' raw ', 'tenant/user', '-user']) assert.throws(() => userIDExports.userId({ uid }), /完整 UID/)
for (const row of [undefined, {}, { user_id: 'local', tenant: 'tenant' }, { natural_key: 'tenant$user' }, { full_user_id: 'tenant$user' }, { name: 'user' }, ...[null, false, 1, '', ' ', 'bad\nuid', 'bad\0uid'].map(uid => ({ uid, user_id: 'local' }))]) assert.throws(() => userIDExports.userId(row), /完整 UID/)
const placementActions = new Map()
function findPlacementActions(node) {
  if (ts.isObjectLiteralExpression(node)) {
    const title = node.properties.find(property => ts.isPropertyAssignment(property) && property.name.getText(placementSource) === 'title')
    if (title && ["'设置用户默认放置'", "'替换用户放置标签'"].includes(title.initializer.getText(placementSource))) {
      const expression = node.getText(placementSource)
      placementActions.set(title.initializer.text, new Function('rgwUserPlacementInput', 'rgwUserPlacementTagsInput', 'userId', `return (${expression})`)(placementForm.rgwUserPlacementInput, placementForm.rgwUserPlacementTagsInput, userIDExports.userId))
    }
  }
  ts.forEachChild(node, findPlacementActions)
}
findPlacementActions(placementSource)
assert.equal(placementActions.size, 2)
const placementRow = { uid: 'tenant$user' }
for (const [title, values] of [
  ['设置用户默认放置', { default_placement: ' custom ', default_storage_class: 'ARCHIVE' }],
  ['设置用户默认放置', { default_placement: 'custom', default_storage_class: '' }],
  ['替换用户放置标签', { placement_tags_csv: 'fast, raw ' }]
]) {
  const action = placementActions.get(title)
  assert.equal(action.path, '/rgw/user')
  assert.equal(action.method, 'PATCH')
  assert.deepEqual(action.buildBody(values, 'cluster', placementRow), { cluster_id: 'cluster', uid: placementRow.uid, ...values })
  const confirmation = action.confirmation(values, placementRow)
  assert.ok(confirmation.includes(JSON.stringify(placementRow.uid)))
  for (const value of Object.values(values)) {
    assert.ok(confirmation.includes(value === '' ? '原生默认类（空值）' : JSON.stringify(value)))
  }
  assert.throws(() => action.buildBody({}, 'cluster', placementRow))
  assert.throws(() => action.confirmation({}, placementRow))
  assert.throws(() => action.buildBody(values, 'cluster', { user_id: 'local', tenant: 'tenant' }), /完整 UID/)
  assert.throws(() => action.confirmation(values, { user_id: 'local', tenant: 'tenant' }), /完整 UID/)
}
assert.ok(pages.includes('detailContent: (row, clusterId) => <RgwUserDetails row={row} clusterId={clusterId} />'))
const userDetailsExports = {}
new Function('exports', 'require', ts.transpileModule(userDetailsSource, { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX } }).outputText)(userDetailsExports, (name) => {
  if (name === 'react/jsx-runtime') return { jsx: (type, props) => ({ type, props }), jsxs: (type, props) => ({ type, props }) }
  if (name === 'antd') return { Tabs: 'Tabs' }
  if (name === './RgwUserIdentityDetails') return { RgwUserIdentityDetails: 'Identity', RgwUserPlacementDetails: 'Placement' }
  if (name === './RgwUserAccountDetails') return { RgwUserAccountDetails: 'Account' }
  if (name === './RgwUserKeys') return { RgwUserKeys: 'Keys' }
  if (name === './RgwPermissions') return { RgwPermissions: 'Permissions' }
  if (name === './RgwQuota') return { RgwQuota: 'Quota' }
  if (name === './RgwRateLimit') return { RgwRateLimit: 'RateLimit' }
  if (name === './RgwStorage') return { RgwStorage: 'Storage' }
  if (name === './rgwStorageDetails') return { rgwStorageScope: (scope, account) => ({ scope, account }) }
  throw new Error(`unexpected import ${name}`)
})
for (const row of [{}, { uid: 'tenant$user', account_id: 'RGW123', stats_scope: 'account', tags: [], placement_tags: ['archive'], caps: [], subusers: [], user_quota: { enabled: false }, bucket_quota: { enabled: true, max_size: 0 }, rate_limit: { enabled: false }, storage_stats: { stats: { num_objects: 0 } } }]) {
  const view = userDetailsExports.RgwUserDetails({ row, clusterId: 42 })
  assert.equal(view.type, 'Tabs')
  assert.deepEqual(view.props.items.map(item => item.key), ['identity', 'placement', 'caps', 'subusers', 'quota', 'bucket-quota', 'rate-limit', 'usage', 'account', 'keys'])
  assert.equal(view.props.items[0].children.type, 'Identity')
  assert.equal(view.props.items[1].children.type, 'Placement')
  for (const item of view.props.items.slice(0, 2)) assert.equal(item.children.props.row, row)
  for (const [index, type, field] of [[2, 'Permissions', 'caps'], [3, 'Permissions', 'subusers'], [4, 'Quota', 'user_quota'], [5, 'Quota', 'bucket_quota'], [6, 'RateLimit', 'rate_limit']]) {
    assert.equal(view.props.items[index].children.type, type)
    assert.equal(view.props.items[index].children.props.value, row[field])
  }
  assert.equal(view.props.items[3].children.props.subusers, true)
  const usage = view.props.items[7].children.props.children
  assert.deepEqual(usage[0].props.children, { scope: row.stats_scope, account: row.account_id })
  assert.equal(usage[1].type, 'Storage')
  assert.equal(usage[1].props.value, row.storage_stats)
  assert.equal(view.props.items[8].children.type, 'Account')
  assert.deepEqual(view.props.items[8].children.props, { clusterId: 42, accountId: row.account_id })
  assert.equal(view.props.items[9].children.type, 'Keys')
  assert.equal(view.props.items[9].children.props.row, row)
}
const components = readFileSync(new URL('../src/pages/object/RgwUserIdentityDetails.tsx', import.meta.url), 'utf8')
const identityView = {}
new Function('exports', 'require', ts.transpileModule(components, { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX } }).outputText)(identityView, (name) => {
  if (name === 'react/jsx-runtime') return { jsx: (type, props) => ({ type, props }), jsxs: (type, props) => ({ type, props }) }
  if (name === 'antd') return { Descriptions: 'Descriptions' }
  if (name === './rgwUserIdentity') return identity
  if (name === './RgwUserTagsTable') return { RgwUserTagsTable: 'Tags' }
  throw new Error(`unexpected import ${name}`)
})
for (const row of [
  { full_user_id: 'user', user_id: 'user', tenant: '' },
  { full_user_id: 'tenant$namespace$user', user_id: 'user', tenant: 'tenant', namespace: 'namespace' },
  { full_user_id: '$namespace$user', user_id: 'user', tenant: '', namespace: 'namespace' }
]) {
  const fields = Object.fromEntries(identityView.RgwUserIdentityDetails({ row }).props.items.map(item => [item.key, item.children]))
  assert.equal(fields['full-uid'], row.full_user_id)
  assert.equal(fields['local-id'], row.user_id)
  assert.equal(fields.tenant, row.tenant || '默认租户')
  assert.equal(fields.namespace, row.namespace ?? '未返回（原生命令在命名空间为空时省略）')
}
for (const value of [null, false, 1, []]) {
  const fields = Object.fromEntries(identityView.RgwUserIdentityDetails({ row: { full_user_id: value, user_id: value, namespace: value } }).props.items.map(item => [item.key, item.children]))
  for (const key of ['full-uid', 'local-id', 'namespace']) assert.equal(fields[key], '未返回或格式无效')
}
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

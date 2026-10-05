import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const exports = {}
const jsx = (type, props) => ({ type, props })
new Function('exports', 'require', ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwRoleDetails.tsx', import.meta.url), 'utf8'), {
  compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX }
}).outputText)(exports, name => {
  if (name === 'react/jsx-runtime') return { jsx, jsxs: jsx }
  if (name === 'antd') return { Tabs: 'Tabs' }
  if (name === './RgwRolePolicyDetails') return { RgwPolicyDocument: 'Document', RgwRolePolicyDetails: 'Inline', RgwRoleManagedPolicies: 'Managed' }
  if (name === './RgwRoleTagsTable') return { RgwRoleTagsTable: 'Tags' }
  if (name === './RgwRoleSummary') return { RgwRoleSummary: 'Summary' }
  throw new Error(name)
})
const fields = ['AssumeRolePolicyDocument', 'PermissionPolicies', 'ManagedPermissionPolicies', 'Tags']
for (const row of [{}, Object.fromEntries(fields.map(field => [field, null])), {
  AssumeRolePolicyDocument: '{"Statement":[]}', PermissionPolicies: [], ManagedPermissionPolicies: [], Tags: []
}, {
  AssumeRolePolicyDocument: '<literal>', PermissionPolicies: [{ PolicyName: 'example', PolicyValue: '{"Statement":[]}' }],
  ManagedPermissionPolicies: ['arn:aws:iam::aws:policy/ReadOnlyAccess'], Tags: [{ Key: 'team', Value: 'storage' }]
}]) {
  const result = exports.RgwRoleDetails({ row })
  assert.ok(result.props.children[0].props.children.includes('不能单独据此判断完整有效权限'))
  const tabs = result.props.children[1]
  assert.equal(tabs.type, 'Tabs')
  assert.deepEqual(tabs.props.items.map(item => item.key), ['trust', 'inline', 'managed', 'tags', 'summary'])
  assert.deepEqual(tabs.props.items.map(item => item.children.type), ['Document', 'Inline', 'Managed', 'Tags', 'Summary'])
  assert.equal(tabs.props.items[4].children.props.row, row)
  fields.forEach((field, index) => assert.equal(tabs.props.items[index].children.props.value, row[field]))
}
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.ok(pages.includes('detailContent: (row) => <RgwRoleDetails row={row} />'))
console.log('RGW role details bind native policy and tag fields without inventing defaults')
function load(file, require = () => { throw new Error('unexpected dependency') }) {
  const exports = {}
  new Function('exports', 'require', ts.transpileModule(readFileSync(new URL(`../src/pages/object/${file}`, import.meta.url), 'utf8'), {
    compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX }
  }).outputText)(exports, require)
  return exports
}
const policies = load('rgwRolePolicies.ts')
const identities = load('rgwUserIdentity.ts')
const summary = load('RgwRoleSummary.tsx', name => {
  if (name === 'react/jsx-runtime') return { jsx, jsxs: jsx }
  if (name === 'antd') return { Descriptions: 'Descriptions' }
  if (name === './rgwUserIdentity') return identities
  throw new Error(name)
})
const summaryRows = row => Object.fromEntries(summary.RgwRoleSummary({ row }).props.items.map(item => [item.key, item.children]))
assert.deepEqual(summaryRows({ RoleName: 'team$reader', RoleId: 'role-id', AccountId: '', Path: '/', Arn: 'arn:role', Description: '<literal>', MaxSessionDuration: 3600, CreateDate: 'native-time' }), {
  name: 'team$reader', id: 'role-id', account: '未关联账户（租户作用域角色）', path: '/', arn: 'arn:role', description: '<literal>', duration: '3600', created: 'native-time'
})
assert.equal(summaryRows({ AccountId: 'RGW123', MaxSessionDuration: 43200 }).account, 'RGW123')
assert.equal(summaryRows({ MaxSessionDuration: 43200 }).duration, '43200')
for (const value of [undefined, null, false, {}, [], 0, 3599, 43201, 3600.5, '3600']) {
  assert.equal(summaryRows({ MaxSessionDuration: value }).duration, '未返回或格式无效')
}
for (const value of [undefined, null, false, 123, {}, []]) {
  const result = summaryRows(Object.fromEntries(['RoleName', 'RoleId', 'AccountId', 'Path', 'Arn', 'Description', 'CreateDate'].map(key => [key, value])))
  assert.ok(Object.values(result).every(item => item === '未返回或格式无效'))
}
const tags = load('rgwRoleTags.ts')
const policyView = load('RgwRolePolicyDetails.tsx', name => {
  if (name === 'react/jsx-runtime') return { jsx, jsxs: jsx }
  if (name === './rgwRolePolicies') return policies
  if (name === './rgwUserIdentity') return identities
  throw new Error(name)
})
const tagView = load('RgwRoleTagsTable.tsx', name => {
  if (name === 'react/jsx-runtime') return { jsx, jsxs: jsx }
  if (name === './rgwRoleTags') return tags
  if (name === 'antd') return { Table: 'Table' }
  throw new Error(name)
})
for (const render of [policyView.RgwRolePolicyDetails, tagView.RgwRoleTagsTable]) {
  assert.ok(render({ value: undefined }).props.children.includes('原生命令'))
  for (const value of [null, {}, 'bad', 0, false]) assert.ok(render({ value }).props.children.includes('格式无效'))
}
assert.ok(policyView.RgwRolePolicyDetails({ value: undefined }).props.children.includes('不能据此判断完整有效权限'))
assert.equal(policyView.RgwRolePolicyDetails({ value: [] }).props.children, '内联策略列表为空')
assert.deepEqual(tagView.RgwRoleTagsTable({ value: [] }).props.dataSource, [])
const inline = policyView.RgwRolePolicyDetails({ value: [{ PolicyName: '<name>', PolicyValue: '<raw-policy>' }] })
assert.equal(inline.props.children[0].props.children[0].props.children, '<name>')
assert.equal(inline.props.children[0].props.children[1].props.value, '<raw-policy>')

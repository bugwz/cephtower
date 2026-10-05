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
  assert.deepEqual(tabs.props.items.map(item => item.key), ['trust', 'inline', 'managed', 'tags'])
  assert.deepEqual(tabs.props.items.map(item => item.children.type), ['Document', 'Inline', 'Managed', 'Tags'])
  fields.forEach((field, index) => assert.equal(tabs.props.items[index].children.props.value, row[field]))
}
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.ok(pages.includes('detailContent: (row) => <RgwRoleDetails row={row} />'))
console.log('RGW role details bind native policy and tag fields without inventing defaults')

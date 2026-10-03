import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api = {}
const jsx = (type, props) => ({ type, props })
new Function('exports', 'require', ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwBucketAcl.tsx', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText)(api, name => name === 'antd' ? { Table: 'Table' } : { jsx, jsxs: jsx })
const owner = { id: 'team$owner', display_name: '<owner>' }
const grantee = { type: 'CanonicalUser', id: owner.id, display_name: '', uri: '', email_address: '' }
const acl = { owner, grants: [{ grantee, permissions: ['READ', 'WRITE_ACP', 'FUTURE_PERMISSION'] }] }
assert.deepEqual(api.bucketAclData(acl), acl)
assert.equal(api.bucketAclAudience(grantee, owner.id), '所有者')
assert.equal(api.bucketAclAudience({ ...grantee, id: 'other' }, owner.id), '指定用户')
assert.match(api.bucketAclAudience({ ...grantee, type: 'Group', uri: 'http://acs.amazonaws.com/groups/global/AllUsers' }, owner.id), /匿名/)
assert.match(api.bucketAclAudience({ ...grantee, type: 'Group', uri: 'http://acs.amazonaws.com/groups/global/AuthenticatedUsers' }, owner.id), /不限本账户/)
assert.equal(api.bucketAclAudience({ ...grantee, type: 'Group', uri: 'https://evil/AllUsers' }, owner.id), '其他授权组')
assert.match(api.bucketAclAudience({ ...grantee, type: 'future' }, owner.id), /未知/)
for (const value of [null, {}, [], { ...acl, owner: null }, { ...acl, grants: [{}] }, { ...acl, grants: [{ grantee, permissions: [3] }] }, { ...acl, grants: [{ grantee: { ...grantee, id: '' }, permissions: [] }] }]) assert.equal(api.bucketAclData(value), undefined)
assert.equal(api.RgwBucketAcl({ value: acl, configured: false }).props.children, 'ACL 数据不可用')
const view = api.RgwBucketAcl({ value: acl, configured: true })
const table = view.props.children.find(child => child.type === 'Table')
assert.equal(table.props.dataSource.length, 1)
assert.match(table.props.columns[2].render(acl.grants[0].permissions), /READ.*WRITE_ACP.*FUTURE_PERMISSION/)
const identity = table.props.columns[1].render({ ...grantee, display_name: '<script>alert(1)</script>' })
assert.equal(typeof identity.props.children, 'string')
assert.equal(identity.props.dangerouslySetInnerHTML, undefined)
assert.match(api.RgwBucketAcl({ value: { owner, grants: [] }, configured: true }).props.children[2].props.locale.emptyText, /不推断/)
console.log('bucket ACL presentation tests passed')

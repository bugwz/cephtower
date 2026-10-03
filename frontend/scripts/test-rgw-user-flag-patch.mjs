import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const exports = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwUserFlagPatch.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const patch = exports.rgwUserFlagPatch
assert.deepEqual(patch({}), {})
assert.deepEqual(patch({ suspended: 'keep', system: 'keep' }), {})
assert.deepEqual(patch({ suspended: 'disable', system: 'enable' }), { suspended: false, system: true })
assert.deepEqual(patch({ suspended: 'enable', system: null }), { suspended: true })
for (const value of [false, true, 0, 1, '', 'false', 'unexpected']) assert.throws(() => patch({ system: value }))
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.ok(pages.includes('...rgwUserFlagPatch(values)'))
assert.ok(!pages.includes('suspended: Boolean(values.suspended)'))
assert.ok(pages.includes("suspended: 'keep'"))
assert.ok(pages.includes("system: 'keep'"))
console.log('RGW user flag updates omit unchanged values and preserve explicit false')
const role = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwRoleEdit.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(role)
assert.deepEqual(role.rgwRoleInitial({}), { assume_role_policy: undefined, max_session_duration: undefined })
const row = { AssumeRolePolicyDocument: '{}', MaxSessionDuration: 7200 }
assert.throws(() => role.rgwRolePatch(role.rgwRoleInitial(row), row))
assert.deepEqual(role.rgwRolePatch({ assume_role_policy: '{"Statement":[]}', max_session_duration: 7200 }, row), { assume_role_policy: '{"Statement":[]}' })
assert.deepEqual(role.rgwRolePatch({ assume_role_policy: '', max_session_duration: 3600 }, row), { max_session_duration: 3600 })
for (const duration of [0, 3599, 43201, 3600.5, '3600']) assert.throws(() => role.rgwRolePatch({ max_session_duration: duration }, row))
for (const policy of ['[]', 'null', 'invalid', '1']) assert.throws(() => role.rgwRolePatch({ assume_role_policy: policy }, row))
assert.ok(pages.includes('...rgwRolePatch(values, row)'))
const tags = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwRoleTags.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(tags)
assert.deepEqual(tags.rgwRoleTags([{ Key: 'env', Value: 'test' }, { Key: 'empty', Value: '' }]), [{ id: 0, key: 'env', value: 'test' }, { id: 1, key: 'empty', value: '' }])
assert.deepEqual(tags.rgwRoleTags([]), [])
for (const value of [null, undefined, {}, [null], [[]], [{ Key: 'env' }], [{ Key: 1, Value: 'test' }]]) assert.equal(tags.rgwRoleTags(value), undefined)
assert.ok(pages.includes('<RgwRoleTagsTable value={value} />'))
assert.ok(pages.includes("key: 'RoleId'"))
const policies = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwRolePolicies.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(policies)
const document = '{"Condition":{"count":9007199254740993},"Statement":[]}'
assert.deepEqual(policies.rgwRolePolicies([{ PolicyName: 'native', PolicyValue: document }]), [{ id: 0, name: 'native', document }])
assert.deepEqual(policies.rgwRolePolicies([]), [])
for (const value of [null, undefined, {}, [null], [{ PolicyName: '', PolicyValue: '{}' }], [{ PolicyName: 'x', PolicyValue: {} }]]) assert.equal(policies.rgwRolePolicies(value), undefined)
assert.ok(pages.includes('<RgwRolePolicyDetails value={value} />'))
assert.ok(pages.includes('<RgwPolicyDocument value={value} />'))
const policyRow = { PermissionPolicies: [{ PolicyName: 'native', PolicyValue: document }] }
assert.deepEqual(policies.rgwPolicyDeleteOptions(policyRow), [{ label: 'native', value: 'native' }])
assert.deepEqual(policies.rgwPolicyMutation({ action: 'delete', existing_policy: 'native', policy_name: 'stale', policy_document: 'stale' }, policyRow), { action: 'delete', policy_name: 'native' })
assert.throws(() => policies.rgwPolicyMutation({ action: 'delete', existing_policy: 'missing' }, policyRow))
assert.throws(() => policies.rgwPolicyDeleteOptions({}))
assert.throws(() => policies.rgwPolicyDeleteOptions({ PermissionPolicies: [...policyRow.PermissionPolicies, ...policyRow.PermissionPolicies] }))
assert.deepEqual(policies.rgwPolicyMutation({ action: 'put', policy_name: 'new', policy_document: document }), { action: 'put', policy_name: 'new', policy_document: document })
for (const policy_document of ['null', '[]', 'invalid']) assert.throws(() => policies.rgwPolicyMutation({ action: 'put', policy_name: 'new', policy_document }))
assert.throws(() => policies.rgwPolicyMutation({ action: 'unknown' }))
assert.ok(pages.includes('...rgwPolicyMutation(values, row)'))

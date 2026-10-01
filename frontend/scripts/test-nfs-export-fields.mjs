import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
import { createRequire } from 'node:module'

const source = readFileSync(new URL('../src/pages/file/nfsExportFields.ts', import.meta.url), 'utf8')
const exports = {}
new Function('exports', ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(exports)
const row = { cluster_id: 'nfs-a', pseudo: '/share', path: '/data', access_type: 'RO', fsal: { name: 'CEPH', fs_name: 'cephfs-a', user_id: 'nfs.user' } }
assert.deepEqual(exports.nfsExportInitialValues(row), { cluster: 'nfs-a', pseudo: '/share', path: '/data', filesystem: 'cephfs-a', access_type: 'RO', squash: undefined, security_label: undefined, transports: undefined, protocols: undefined, sectype: undefined, clients: undefined, fsal_type: 'CEPH', rgw_export_type: 'user', rgw_bucket: undefined, rgw_user_id: 'nfs.user' })
assert.deepEqual(exports.nfsFSALBody({ fsal_type: 'RGW', rgw_user_id: 'owner', filesystem: 'stale' }), { fsal_type: 'RGW', path: '/', rgw_user_id: 'owner' })
assert.deepEqual(exports.nfsFSALBody({ fsal_type: 'CEPH', filesystem: 'fs', rgw_user_id: 'stale' }), { fsal_type: 'CEPH', filesystem: 'fs' })
assert.deepEqual(exports.nfsFSALBody({ fsal_type: 'RGW', rgw_export_type: 'bucket', path: 'stale', rgw_bucket: 'bucket', rgw_user_id: 'stale' }), { fsal_type: 'RGW', path: 'bucket' })
assert.deepEqual(exports.nfsRGWBucketChoices([{ tenant: '', bucket: 'a' }, { tenant: 'tenant', bucket: 'a' }, { bucket: 'unknown' }, { tenant: '', bucket: 'a' }, { tenant: '', bucket: 'x/y' }]), [{ label: 'a', value: 'a' }])
assert.equal(exports.nfsExportInitialValues({ fsal: { name: 'RGW' }, path: 'a' }).rgw_bucket, 'a')
assert.equal(exports.nfsExportInitialValues({ fsal: { name: 'RGW' }, path: 'bucket' }).rgw_export_type, 'bucket')
assert.equal(exports.nfsExportInitialValues({ fsal: { name: 'RGW' }, path: '/' }).rgw_export_type, 'user')
assert.deepEqual(exports.nfsRGWUserChoices([{ uid: 'tenant$user', user_id: 'user', display_name: 'Owner' }, { uid: 'plain' }, { uid: 'plain' }, { user_id: 'unqualified' }, { uid: '' }]), [{ label: 'Owner (tenant$user)', value: 'tenant$user' }, { label: 'plain', value: 'plain' }])
assert.deepEqual(exports.nfsClientsBody(''), {})
assert.deepEqual(exports.nfsClientsBody('[]'), { client_rules: [] })
assert.deepEqual(exports.nfsClientsBody('[{"addresses":["10.0.0.0/8"],"access_type":null,"squash":null}]'), { client_rules: [{ addresses: ['10.0.0.0/8'], access_type: '', squash: '' }] })
for (const value of ['null', '{}', '[{}]', '[{"addresses":["x;}"]}]', '[']) assert.throws(() => exports.nfsClientsBody(value))
assert.deepEqual(exports.nfsSecurityTypeBody('none, sys, krb5,krb5i,krb5p'), { sectype: ['none', 'sys', 'krb5', 'krb5i', 'krb5p'] })
assert.deepEqual(exports.nfsSecurityTypeBody('default'), { sectype: [] })
assert.deepEqual(exports.nfsSecurityTypeBody(''), {})
assert.equal(exports.nfsExportInitialValues({ ...row, sectype: ['sys', 'krb5p'] }).sectype, 'sys,krb5p')
for (const value of ['sys,sys', 'invalid', 'sys,', true]) assert.throws(() => exports.nfsSecurityTypeBody(value))
for (const protocols of [[3], [4], [4, 3]]) {
  const selection = exports.nfsExportInitialValues({ ...row, protocols }).protocols
  assert.deepEqual(exports.nfsProtocolBody(selection), { protocols: [...protocols].sort() })
}
for (const protocols of [[], [3, 3], [5], ['4'], null]) assert.equal(exports.nfsExportInitialValues({ ...row, protocols }).protocols, undefined)
assert.deepEqual(exports.nfsProtocolBody(undefined), {})
assert.throws(() => exports.nfsProtocolBody('5'))
for (const transports of [['TCP'], ['UDP'], ['UDP', 'TCP']]) {
  const selection = exports.nfsExportInitialValues({ ...row, transports }).transports
  assert.deepEqual(exports.nfsTransportBody(selection), { transports: [...transports].sort() })
}
for (const transports of [[], ['TCP', 'TCP'], ['SCTP'], null]) assert.equal(exports.nfsExportInitialValues({ ...row, transports }).transports, undefined)
assert.deepEqual(exports.nfsTransportBody(undefined), {})
assert.throws(() => exports.nfsTransportBody('SCTP'))
assert.equal(exports.nfsExportInitialValues({ ...row, security_label: false }).security_label, 'disabled')
assert.equal(exports.nfsExportInitialValues({ ...row, security_label: true }).security_label, 'enabled')
assert.equal(exports.nfsExportInitialValues({ ...row, squash: 'all_squash' }).squash, 'all_squash')
assert.equal(exports.nfsExportInitialValues({ ...row, squash: 'native-alias' }).squash, undefined)
assert.equal(exports.nfsExportEditReason(row), undefined)
assert.equal(exports.nfsExportInitialValues({ ...row, access_type: 'RW' }).access_type, 'RW')
assert.equal(exports.nfsExportEditReason({ ...row, fsal: { name: 'RGW' } }), undefined)
assert.ok(exports.nfsExportEditReason({ ...row, fsal: { name: 'UNKNOWN' } }))
assert.equal(exports.nfsExportEditReason({ ...row, access_type: 'NONE' }), undefined)
for (const fsal of [null, undefined, [], 'CEPH']) assert.deepEqual(exports.nfsFSAL({ fsal }), {})
console.log('NFS native export field checks passed')
const pagesSource = readFileSync(new URL('../src/pages/file/pages.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('pages.tsx', pagesSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const userLoader = tree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name?.text === 'nfsRGWUserOptions')
const userLoaderCode = ts.transpileModule(userLoader.getText(tree), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText
const userCalls = []
const loadUsers = new Function('listAllResources', 'nfsRGWUserChoices', `${userLoaderCode}; return nfsRGWUserOptions`)(async (...args) => { userCalls.push(args); return { items: [{ uid: 'tenant$user' }] } }, exports.nfsRGWUserChoices)
assert.deepEqual(await loadUsers(7, undefined, { fsal_type: 'CEPH' }), [])
assert.equal(userCalls.length, 0)
assert.deepEqual(await loadUsers(7, undefined, { fsal_type: 'RGW', rgw_export_type: 'bucket' }), [])
assert.equal(userCalls.length, 0)
assert.deepEqual(await loadUsers(7, undefined, { fsal_type: 'RGW', rgw_export_type: 'user' }), [{ label: 'tenant$user', value: 'tenant$user' }])
assert.deepEqual(userCalls, [['/rgw/users', 7]])
const bucketLoader = tree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name?.text === 'nfsRGWBucketOptions')
const bucketLoaderCode = ts.transpileModule(bucketLoader.getText(tree), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText
const bucketCalls = []
const loadBuckets = new Function('listAllResources', 'nfsRGWBucketChoices', `${bucketLoaderCode}; return nfsRGWBucketOptions`)(async (...args) => { bucketCalls.push(args); return { items: [{ tenant: '', bucket: 'a' }] } }, exports.nfsRGWBucketChoices)
assert.deepEqual(await loadBuckets(9, undefined, { fsal_type: 'RGW', rgw_export_type: 'user' }), [])
assert.equal(bucketCalls.length, 0)
assert.deepEqual(await loadBuckets(9, undefined, { fsal_type: 'RGW', rgw_export_type: 'bucket' }), [{ label: 'a', value: 'a' }])
assert.deepEqual(bucketCalls, [['/rgw/buckets', 9]])
const identity = tree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name?.text === 'exportId')
assert.ok(identity)
const code = ts.transpileModule(identity.getText(tree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
const exportId = new Function(`${code}; return exportId`)()
assert.equal(exportId({ export_id: 2, natural_key: 'bmZzLWEAMg' }), 'bmZzLWEAMg')
assert.equal(exportId({ export_id: 2 }), '')

// Exercise the controlled editor callbacks without a running Ceph cluster or browser.
const require = createRequire(import.meta.url)
const editorSource = readFileSync(new URL('../src/pages/file/NFSClientsEditor.tsx', import.meta.url), 'utf8')
const editorExports = {}
const editorCode = ts.transpileModule(editorSource, { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX } }).outputText
new Function('require', 'exports', editorCode)((name) => {
  if (name === './nfsExportFields') return exports
  if (name === 'antd') return { Alert: 'Alert', Button: 'Button', Card: 'Card', Input: 'Input', Select: 'Select', Space: 'Space', Typography: { Text: 'Text' } }
  return require(name)
}, editorExports)
function nodes(tree) {
  if (!tree || typeof tree !== 'object') return []
  if (Array.isArray(tree)) return tree.flatMap(nodes)
  return [tree, ...nodes(tree.props?.children), ...nodes(tree.props?.extra)]
}
let value
const renderEditor = () => editorExports.NFSClientsEditor({ value, onChange: (next) => { value = next } })
let elements = nodes(renderEditor())
elements.find((node) => node.type === 'Button' && node.props.children === '新增客户端规则').props.onClick()
assert.deepEqual(JSON.parse(value), [{ addresses: [''], access_type: '', squash: '' }])
elements = nodes(renderEditor())
elements.find((node) => node.type === 'Input').props.onChange({ target: { value: '10.0.0.0/8, host.example.com' } })
assert.deepEqual(exports.nfsClientsBody(value).client_rules[0].addresses, ['10.0.0.0/8', 'host.example.com'])
elements = nodes(renderEditor())
elements.find((node) => node.type === 'Select' && node.props['aria-label'].endsWith('访问类型')).props.onChange('RO')
assert.equal(exports.nfsClientsBody(value).client_rules[0].access_type, 'RO')
elements = nodes(renderEditor())
elements.find((node) => node.type === 'Button' && node.props.danger).props.onClick()
assert.deepEqual(exports.nfsClientsBody(value), { client_rules: [] })
value = JSON.stringify([{ addresses: ['*'], access_type: null, squash: 'root' }])
elements = nodes(renderEditor())
assert.ok(elements.find((node) => node.type === 'Select' && node.props.options.some((option) => option.value === 'root')))
value = '[{}]'
assert.equal(renderEditor().type, 'Alert')
assert.equal(value, '[{}]')
console.log('NFS client editor interaction checks passed')

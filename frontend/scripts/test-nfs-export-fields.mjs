import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
import { createRequire } from 'node:module'

const source = readFileSync(new URL('../src/pages/file/nfsExportFields.ts', import.meta.url), 'utf8')
const exports = {}
new Function('exports', ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const row = { cluster_id: 'nfs-a', pseudo: '/share', path: '/data', access_type: 'RO', fsal: { name: 'CEPH', fs_name: 'cephfs-a', user_id: 'nfs.user' } }
assert.deepEqual(exports.nfsExportInitialValues(row), { cluster: 'nfs-a', pseudo: '/share', path: '/data', filesystem: 'cephfs-a', access_type: 'RO', squash: undefined, security_label: undefined, transports: undefined, protocols: undefined, sectype: undefined, clients: undefined })
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
assert.ok(exports.nfsExportEditReason({ ...row, fsal: { name: 'RGW' } }))
assert.equal(exports.nfsExportEditReason({ ...row, access_type: 'NONE' }), undefined)
for (const fsal of [null, undefined, [], 'CEPH']) assert.deepEqual(exports.nfsFSAL({ fsal }), {})
console.log('NFS native export field checks passed')
const pagesSource = readFileSync(new URL('../src/pages/file/pages.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('pages.tsx', pagesSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
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

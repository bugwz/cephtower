import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketMFA.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(api)
const row = { bucket_id: 'AGJ1Y2tldA', kind: 'versioning', configured: true, document: '<VersioningConfiguration/>', versioning: { status: '', mfa_delete: null } }
assert.match(api.bucketVersioningSummary(row.versioning, row), /尚未启用.*未返回/)
assert.match(api.bucketVersioningSummary({ status: 'Suspended', mfa_delete: null }, row), /暂停.*未知/)
assert.equal(api.bucketVersioningSummary(null, row), '版本控制配置不可用')
assert.equal(api.bucketVersioningSummary(null, { kind: 'policy' }), '—')
const initial = api.bucketMFAInitial(row)
assert.equal(initial.status, undefined)
assert.equal(initial.mfa_delete, undefined)
assert.equal(initial.mfa_token, '')
const values = { ...initial, status: 'Enabled', mfa_delete: 'Enabled', mfa_serial_secret: 'private-device', mfa_token: '001234', confirm_mfa: 'acknowledged' }
const input = api.bucketMFAInput(values, row)
assert.equal(input.mfa_token, '001234')
assert.equal(input.expected_document, row.document)
for (const invalid of [{ configured: false }, { document: null }, { kind: 'policy' }, { versioning: { status: 'Enabled', mfa_delete: null } }, { versioning: null }]) assert.ok(api.bucketMFABlocked({ ...row, ...invalid }))
for (const patch of [{ bucket_id: 'other' }, { status: '' }, { mfa_delete: 'future' }, { mfa_serial_secret: 'x y' }, { mfa_serial_secret: 'x\r\ny' }, { mfa_token: 1234 }, { mfa_token: '12 34' }, { mfa_token: '' }, { confirm_mfa: undefined }]) assert.throws(() => api.bucketMFAInput({ ...values, ...patch }, row))
assert.throws(() => api.bucketMFAInput(values, { ...row, versioning: { status: 'Enabled', mfa_delete: 'Enabled' } }), /没有变化/)
const confirmation = api.bucketMFAConfirmation(values, row)
for (const text of ['HTTPS', '过期', '不注册', '回滚', '对象锁', '删除保护', '原子锁']) assert.ok(confirmation.includes(text))
for (const secret of [values.mfa_token, values.mfa_serial_secret]) assert.ok(!confirmation.includes(secret))
const source = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const page = ts.createSourceFile('pages.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let action
function visit(node) {
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(page) === 'title' && ts.isStringLiteral(p.initializer) && p.initializer.text === '设置 MFA Delete 与版本控制')) {
    const code = ts.transpileModule(`const action = ${node.getText(page)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    action = new Function(...Object.keys(api), `${code}; return action`)(...Object.values(api))
  }
  ts.forEachChild(node, visit)
}
visit(page)
assert.equal(action.path, '/rgw/bucket/mfa')
assert.equal(action.method, 'PATCH')
assert.equal(action.visibleWhen(row), true)
assert.equal(action.visibleWhen({ kind: 'policy' }), false)
assert.equal(action.disabledWhen, api.bucketMFABlocked)
assert.equal(action.initialValues, api.bucketMFAInitial)
assert.equal(action.confirmation, api.bucketMFAConfirmation)
for (const name of ['mfa_serial_secret', 'mfa_token']) assert.equal(action.fields.find(field => field.name === name).type, 'password')
assert.deepEqual(action.buildBody(values, 42, row), { cluster_id: 42, ...input })
console.log('MFA state, credential handling and form binding tests passed')

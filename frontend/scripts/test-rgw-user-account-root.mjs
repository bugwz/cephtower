import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const helpers = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwUserAccountRoot.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(helpers)
const account = 'RGW12345678901234567'
for (const type of ['root', 'rgw']) {
  const row = { account_id: account, type }
  assert.equal(helpers.rgwUserAccountRootBlocked(row), undefined)
  for (const action of ['enable', 'disable']) assert.deepEqual(helpers.rgwUserAccountRootInput({ account_root: action }, row), { account_root: action === 'enable', expected_account_id: account })
  for (const value of [undefined, null, true, false, '', 'true']) assert.throws(() => helpers.rgwUserAccountRootInput({ account_root: value }, row))
}
for (const row of [undefined, {}, { account_id: '' }, { account_id: account, type: 'ldap' }, { account_id: account + ' ', type: 'rgw' }]) {
  assert.ok(helpers.rgwUserAccountRootBlocked(row))
  assert.throws(() => helpers.rgwUserAccountRootInput({ account_root: 'enable' }, row))
}
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const source = ts.createSourceFile('pages.tsx', pages, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let actionNode
function visit(node) {
  if (ts.isObjectLiteralExpression(node) && node.properties.some(prop => ts.isPropertyAssignment(prop) && prop.name.getText(source) === 'title' && ts.isStringLiteral(prop.initializer) && prop.initializer.text === '设置账户根用户')) actionNode = node
  ts.forEachChild(node, visit)
}
visit(source)
assert.ok(actionNode)
const code = ts.transpileModule(`const action = ${actionNode.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const action = new Function('rgwUserAccountRootBlocked', 'rgwUserAccountRootInput', 'userId', `${code}; return action`)(helpers.rgwUserAccountRootBlocked, helpers.rgwUserAccountRootInput, row => row.uid)
for (const value of ['enable', 'disable']) {
  const row = { uid: 'tenant$user', account_id: account, type: 'rgw' }
  assert.deepEqual(action.buildBody({ account_root: value }, 7, row), { cluster_id: 7, uid: row.uid, account_root: value === 'enable', expected_account_id: account })
  const text = action.confirmation({ account_root: value }, row)
  assert.ok(text.includes(row.uid) && text.includes(account) && text.includes('权限'))
}
assert.equal(action.path, '/rgw/user')
assert.equal(action.method, 'PATCH')
assert.ok(action.disabledWhen({}))
console.log('RGW account root mutation guards and confirmation checks passed')

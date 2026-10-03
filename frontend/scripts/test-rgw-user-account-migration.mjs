import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const helpers = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwUserAccountMigration.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(helpers)
const row = { uid: 'tenant$user', account_id: '', display_name: 'user-name', type: 'rgw' }
const values = { target_account_id: 'RGW12345678901234567', migration_confirm_uid: row.uid }
assert.equal(helpers.rgwUserAccountMigrationBlocked(row), undefined)
assert.deepEqual(helpers.rgwUserAccountMigrationInput(values, row), values)
for (const bad of [{}, { ...row, account_id: null }, { ...row, account_id: values.target_account_id }, { ...row, type: 'root' }, { ...row, display_name: 'invalid name' }]) {
  assert.ok(helpers.rgwUserAccountMigrationBlocked(bad))
  assert.throws(() => helpers.rgwUserAccountMigrationInput(values, bad))
}
for (const bad of [{}, { ...values, migration_confirm_uid: 'user' }, { ...values, target_account_id: values.target_account_id + ' ' }]) assert.throws(() => helpers.rgwUserAccountMigrationInput(bad, row))
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const source = ts.createSourceFile('pages.tsx', pages, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let node
function visit(value) {
  if (ts.isObjectLiteralExpression(value) && value.properties.some(prop => ts.isPropertyAssignment(prop) && prop.name.getText(source) === 'title' && ts.isStringLiteral(prop.initializer) && prop.initializer.text === '迁入账户（不可逆）')) node = value
  ts.forEachChild(value, visit)
}
visit(source)
assert.ok(node)
const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const optionsLoader = async () => []
const action = new Function('rgwUserAccountMigrationBlocked', 'rgwUserAccountMigrationInput', 'userId', 'loadRgwMigrationAccountOptions', `${code}; return action`)(helpers.rgwUserAccountMigrationBlocked, helpers.rgwUserAccountMigrationInput, row => row.uid, optionsLoader)
const targetField = action.fields.find(field => field.name === 'target_account_id')
assert.equal(targetField.type, 'select')
assert.equal(targetField.optionsLoader, optionsLoader)
assert.deepEqual(action.buildBody(values, 7, row), { cluster_id: 7, uid: row.uid, ...values })
const confirmation = action.confirmation(values, row)
for (const text of [row.uid, values.target_account_id, '不可逆', 'Bucket', '部分', '不能直接重试']) assert.ok(confirmation.includes(text))
assert.equal(action.path, '/rgw/user')
assert.equal(action.method, 'PATCH')
assert.ok(action.disabledWhen({}))
console.log('RGW account migration validates explicit identity and irreversible confirmation')

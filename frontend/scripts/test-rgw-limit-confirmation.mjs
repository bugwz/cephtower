import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const env = {}
for (const name of ['rgwQuotaForm', 'rgwRateLimitForm']) {
  new Function('exports', ts.transpileModule(readFileSync(new URL(`../src/pages/object/${name}.ts`, import.meta.url), 'utf8'), {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 }
  }).outputText)(env)
}
const quota = { enabled: 'disable', max_size: 1025, max_objects: 0 }
const rate = { enabled: 'enable', max_read_ops: 0, max_write_ops: 1, max_read_bytes: 1024, max_write_bytes: Number.MAX_SAFE_INTEGER }
const confirmation = env.rgwQuotaConfirmation(quota, 'target', 'scope')
for (const expected of ['target', 'scope', '关闭', '1025 字节', '2048 字节', '对象上限：0', '仍保存']) assert.ok(confirmation.includes(expected))
assert.ok(env.rgwQuotaConfirmation({ ...quota, max_size: Number.MAX_SAFE_INTEGER }, '', '').includes('9007199254740992 字节'))
assert.ok(env.rgwQuotaConfirmation({ ...quota, max_size: 0 }, '', '').includes('取整后 0 字节'))
assert.ok(env.rgwQuotaConfirmation({ ...quota, max_size: -1, max_objects: -1 }, '', '').includes('容量上限：无限制（-1）；对象上限：无限制（-1）'))
const rateText = env.rgwRateLimitConfirmation(rate, 'target')
for (const expected of ['target', '启用', '读请求数：无限制（0）', '写请求数：1', '读取字节数：1024', '写入字节数：9007199254740991', '每 RGW 每分钟']) assert.ok(rateText.includes(expected))
assert.ok(env.rgwRateLimitConfirmation({ ...rate, enabled: 'disable' }, '').includes('关闭'))
assert.throws(() => env.rgwQuotaConfirmation({ ...quota, max_size: '0' }, '', ''))
assert.throws(() => env.rgwRateLimitConfirmation({ ...rate, max_read_ops: -1 }, ''))

// Evaluate the real action definitions, including both mapped quota scopes.
const source = ts.createSourceFile('pages.tsx', readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8'), ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
for (const name of ['userId', 'bucketId']) {
  const declaration = source.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === name)
  assert.ok(declaration)
  new Function('exports', ts.transpileModule('export ' + declaration.getText(source), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(env)
}
const paths = new Set(['/rgw/user/quota', '/rgw/account/quota', '/rgw/bucket/quota', '/rgw/user/ratelimit', '/rgw/bucket/ratelimit'])
let checked = 0
function visit(node) {
  if (ts.isObjectLiteralExpression(node)) {
    const path = node.properties.find(property => ts.isPropertyAssignment(property) && property.name.getText(source) === 'path')?.initializer
    if (path && ts.isStringLiteral(path) && paths.has(path.text)) {
      const scopes = path.text === '/rgw/user/quota' ? ['user', 'bucket'] : path.text === '/rgw/account/quota' ? ['account', 'bucket'] : [undefined]
      for (const scope of scopes) {
        const code = ts.transpileModule('export const action = ' + node.getText(source), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
        const output = {}
        new Function('exports', 'scope', ...Object.keys(env), code)(output, scope, ...Object.values(env))
        const row = { uid: 'tenant$user', account_id: 'RGW123', natural_key: 'dGVhbQBwaG90b3M' }
        const values = path.text.endsWith('/quota') ? quota : rate
        const text = output.action.confirmation(values, row)
        const body = output.action.buildBody(values, 7, row)
        assert.ok(text.includes(JSON.stringify(body.uid ?? body.account_id ?? body.bucket_id)))
        assert.equal(body.cluster_id, 7)
        assert.equal(body.enabled, values.enabled === 'enable')
        if (scope) assert.equal(body.scope, scope)
        if (scope === 'bucket') assert.ok(text.includes('默认 Bucket 配额'))
        if (scope === 'account') assert.ok(text.includes('账户总配额'))
        if (scope === 'user') assert.ok(text.includes('用户总配额'))
        assert.throws(() => output.action.confirmation({ ...values, enabled: undefined }, row))
        checked++
      }
    }
  }
  ts.forEachChild(node, visit)
}
visit(source)
assert.equal(checked, 7)
console.log('RGW quota and rate-limit confirmations preserve target, scope and exact limits')

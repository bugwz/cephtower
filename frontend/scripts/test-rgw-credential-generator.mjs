import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { webcrypto } from 'node:crypto'
import ts from 'typescript'
const helpers = {}
const generatorSource = readFileSync(new URL('../src/pages/object/rgwCredentialGenerator.ts', import.meta.url), 'utf8')
new Function('exports', ts.transpileModule(generatorSource, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(helpers)
for (const [kind, length, alphabet] of [['access', 20, 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567'], ['secret', 40, 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_']]) {
  let buffer
  const generated = helpers.generateRGWCredential(kind, { getRandomValues: bytes => { buffer = bytes; for (let i = 0; i < bytes.length; i++) bytes[i] = i; return bytes } })
  assert.equal(generated, Array.from({ length }, (_, i) => alphabet[i % alphabet.length]).join(''))
  assert.ok(buffer.every(byte => byte === 0))
  for (let i = 0; i < 8; i++) {
    const random = helpers.generateRGWCredential(kind, webcrypto)
    assert.equal(random.length, length)
    assert.ok([...random].every(character => alphabet.includes(character)))
  }
  const counts = new Map()
  for (let byte = 0; byte < 256; byte++) {
    const value = helpers.generateRGWCredential(kind, { getRandomValues: bytes => bytes.fill(byte) })
    counts.set(value[0], (counts.get(value[0]) ?? 0) + 1)
  }
  assert.equal(counts.size, alphabet.length)
  assert.ok([...counts.values()].every(count => count === 256 / alphabet.length))
}
assert.throws(() => helpers.generateRGWCredential('access', {}))
assert.throws(() => helpers.generateRGWCredential('access', null))
assert.throws(() => helpers.generateRGWCredential('unsupported', webcrypto))
let failedBuffer
assert.throws(() => helpers.generateRGWCredential('secret', { getRandomValues: bytes => { failedBuffer = bytes; bytes.fill(17); throw new Error('failed') } }))
assert.ok(failedBuffer.every(byte => byte === 0))
assert.ok(!generatorSource.includes('Math.random'))

const component = {}
const errors = [], changes = []
let calls = 0, fail = false
const jsx = (type, props) => ({ type, props })
const code = ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwGeneratedCredentialInput.tsx', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText
new Function('exports', 'require', code)(component, name => {
  if (name === 'react/jsx-runtime') return { jsx, jsxs: jsx }
  if (name === 'antd') return { Button: 'button', Input: { Password: 'password' }, Space: 'space', Typography: { Text: 'text' } }
  if (name.endsWith('appMessage')) return { message: { error: value => errors.push(value) } }
  if (name.endsWith('rgwCredentialGenerator')) return { generateRGWCredential: kind => { calls++; if (fail) throw new Error('sensitive-exception'); return kind + '-generated' } }
  throw new Error(name)
})
const render = value => component.RgwGeneratedCredentialInput({ kind: 'secret', value, id: 'credential-field', onChange: value => changes.push(value) })
let control = render('existing')
assert.equal(calls, 0, 'render must not generate or mutate credentials')
const [input, button] = control.props.children
assert.equal(input.type, 'password')
assert.equal(input.props.value, 'existing')
assert.equal(input.props.id, 'credential-field')
assert.equal(input.props.autoComplete, 'new-password')
assert.equal(button.props.htmlType, 'button')
assert.equal(button.props.children, '重新生成并替换')
input.props.onChange({ target: { value: 'manual' } })
button.props.onClick()
assert.deepEqual(changes, ['manual', 'secret-generated'])
assert.equal(calls, 1)
fail = true
button.props.onClick()
assert.deepEqual(changes, ['manual', 'secret-generated'])
assert.equal(errors.length, 1)
assert.ok(!errors[0].includes('sensitive-exception'))
control = render(undefined)
assert.equal(control.props.children[0].props.value, '')
assert.equal(control.props.children[1].props.children, '安全随机生成')

const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const source = ts.createSourceFile('pages.tsx', pages, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const expected = new Map([
  ['新建 RGW 用户', { access_key: 'access', secret_key: 'secret' }],
  ['创建 S3 访问密钥', { access_key: 'access', secret_key: 'secret' }],
  ['创建子用户', { access_key: 'access', secret_key: 'secret' }],
  ['轮换 S3 访问密钥', { secret_key: 'secret' }],
  ['轮换 Swift 子用户密钥', { secret_key: 'secret' }]
])
const prop = (node, name) => node.properties.find(item => ts.isPropertyAssignment(item) && item.name.getText(source) === name)?.initializer
function visit(node) {
  if (ts.isObjectLiteralExpression(node)) {
    const title = prop(node, 'title')
    if (title && ts.isStringLiteral(title) && expected.has(title.text)) {
      const wanted = expected.get(title.text)
      const fields = prop(node, 'fields')
      assert.ok(ts.isArrayLiteralExpression(fields))
      for (const field of fields.elements) {
        const name = prop(field, 'name')?.text
        const renderer = prop(field, 'renderControl')
        if (wanted[name]) {
          assert.ok(renderer?.getText(source).includes(`kind="${wanted[name]}"`))
          delete wanted[name]
        } else assert.equal(renderer, undefined, 'existing Access Keys must not get a generator')
      }
      assert.deepEqual(wanted, {})
      expected.delete(title.text)
    }
  }
  ts.forEachChild(node, visit)
}
visit(source)
assert.equal(expected.size, 0)
console.log('RGW credential generation uses secure random bytes and explicit controlled input changes')

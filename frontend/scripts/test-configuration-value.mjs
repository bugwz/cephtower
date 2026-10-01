import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/cluster/configurationValue.ts', import.meta.url), 'utf8')
const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS } }).outputText
const exports = {}
new Function('exports', code)(exports)
const validate = (type, value, metadata = {}) => exports.configurationValueError({ name: 'test', type, ...metadata }, 'test', value)
for (const [type, value] of [['int', '-9223372036854775808'], ['int', '9223372036854775807'], ['uint', '18E'], ['size', '15EiB'], ['size', '4GB'], ['size', '+4Gi'], ['int', '2M'], ['uint', '0'], ['float', '1.25e-3']]) assert.equal(validate(type, value), undefined, `${type}: ${value}`)
for (const [type, value] of [['int', '9223372036854775808'], ['uint', '-1'], ['size', '-1'], ['size', '16EiB'], ['int', '10E'], ['uint', '20E'], ['int', '1.2'], ['int', '2MiB'], ['size', '1.5G'], ['size', '4gb'], ['size', '4Bi'], ['float', '1e999']]) assert.ok(validate(type, value), `${type}: ${value}`)
for (const type of ['int', 'uint', 'size', 'float']) assert.ok(validate(type, ''))
assert.equal(validate('int', '9007199254740993', { min: '9007199254740993' }), undefined)
assert.match(validate('int', '9007199254740992', { min: '9007199254740993' }), /不小于/)
assert.match(validate('int', '9007199254740993', { max: '9007199254740992' }), /不大于/)
assert.match(validate('int', '1', { max: '0' }), /不大于/)
assert.equal(validate('size', '1KiB', { min: '1024', max: '1024' }), undefined)
assert.equal(validate('uint', '1K', { min: '1000', max: '1000' }), undefined)
assert.match(validate('float', '-0.1', { min: '0' }), /不小于/)
assert.match(validate('float', '0.1', { max: '0' }), /不大于/)
assert.equal(validate('float', '0x1p2'), undefined, 'leave alternate native float syntax to Ceph')
for (const value of ['bad', '1.2.3', '1e', '2K', ' ']) assert.ok(validate('float', value))
for (const type of ['str', 'secs', 'addr', 'uuid', 'unknown']) assert.equal(validate(type, ''), undefined)
assert.equal(exports.configurationValueError(null, 'test', ''), undefined)
assert.equal(exports.configurationValueError({ name: 'other', type: 'int' }, 'test', ''), undefined)
const page = readFileSync(new URL('../src/pages/cluster/ConfigurationPage.tsx', import.meta.url), 'utf8')
assert.ok(page.includes("configurationValueError(help, form.getFieldValue('name'), value ?? '')"))
console.log('Configuration numeric syntax and exact range checks passed')

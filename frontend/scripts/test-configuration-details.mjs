import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/cluster/ConfigurationPage.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('ConfigurationPage.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const names = ['configurationOverrides', 'configurationList', 'configurationRuntime']
const code = ts.transpileModule(tree.statements.filter((node) => ts.isFunctionDeclaration(node) && names.includes(node.name.text)).map((node) => node.getText(tree)).join('\n'), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const helpers = new Function(`${code}; return { ${names.join(', ')} }`)()
const rows = [
  { name: 'option', who: 'global', value: 'false', natural_key: 'a' },
  { name: 'option', who: 'osd/class:ssd', value: '0', natural_key: 'b' },
  { name: 'option', who: 'osd.1', value: '', natural_key: 'c' },
  { name: 'other', who: 'global', value: 'secret', natural_key: 'd' },
  { name: 'option', who: 'mgr', value: '[REDACTED]', natural_key: 'e' }
]
assert.deepEqual(helpers.configurationOverrides(rows, 'option'), [rows[0], rows[1], rows[2], rows[4]])
for (const name of [undefined, null, '', 'missing', 'opt']) assert.deepEqual(helpers.configurationOverrides(rows, name), [])
assert.deepEqual(helpers.configurationOverrides([], 'option'), [])
assert.equal(helpers.configurationList(['osd', 'mgr']), 'osd、mgr')
assert.equal(helpers.configurationList([]), '无')
for (const value of [undefined, null, false, 'osd', [1], [{}]]) assert.equal(helpers.configurationList(value), '未采集')
assert.equal(helpers.configurationRuntime(true), '是')
assert.equal(helpers.configurationRuntime(false), '否')
for (const value of [undefined, null, 0, 1, 'false', 'true']) assert.equal(helpers.configurationRuntime(value), '未采集')
for (const name of ['flags', 'services', 'tags', 'enum_values', 'see_also']) assert.ok(source.includes(`configurationList(detail.${name})`))
assert.ok(source.includes('configurationOverrides(data?.values ?? [], detail.name)'))
assert.ok(source.includes('覆盖值不等同于某个守护进程最终生效的值'))
assert.ok(source.includes("value === '' ? '空字符串'"))
console.log('Configuration metadata and scoped override checks passed')

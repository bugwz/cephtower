import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/cluster/ManagerModuleDetails.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('details.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const names = ['managerOptionRows', 'managerOptionValue', 'managerModuleOverrides']
const code = ts.transpileModule(tree.statements.filter((node) => ts.isFunctionDeclaration(node) && names.includes(node.name.text)).map((node) => node.getText(tree).replace('export ', '')).join('\n'), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const { managerOptionRows, managerOptionValue, managerModuleOverrides } = new Function(`${code}; return { ${names.join(', ')} }`)()
const overrides = [
  { name: 'mgr/test/port', who: 'mgr', value: '0' },
  { name: 'mgr/test/node-a/port', who: 'mgr', value: '1' },
  { name: 'mgr/test/secret', who: 'global', value: '[REDACTED]' },
  { name: 'mgr/test/port', who: 'mgr/host:node-a', value: '' },
  { name: 'mgr/test-other/port', value: '2' }, { name: 'mgr/test/', value: '3' }, { name: 'other', value: '4' }
]
assert.deepEqual(managerModuleOverrides(overrides, 'test'), overrides.slice(0, 4))
for (const name of [undefined, null, '', 'test/port']) assert.deepEqual(managerModuleOverrides(overrides, name), [])
assert.ok(source.includes("listAllResources('/configuration/values', clusterId)"))
assert.ok(source.includes('overrides.items.every((row) => row.stale === false)'))
const metadata = { option: { name: 'option', default_value: '9007199254740993', min: '0', max: '', enum_allowed: [], flags: 0, tags: ['test'] } }
assert.deepEqual(managerOptionRows(metadata), [{ ...metadata.option, option_name: 'option' }])
assert.equal(metadata.option.option_name, undefined, 'do not mutate API inventory')
assert.deepEqual(managerOptionRows({}), [])
for (const value of [undefined, null, [], '', { bad: null }, { bad: [] }, { '': {} }]) assert.equal(managerOptionRows(value), null)
for (const [value, expected] of [[false, 'false'], [0, '0'], ['', '空字符串'], ['None', 'None'], ['9007199254740993', '9007199254740993'], [[], '无'], [['a', 'b'], 'a、b']]) assert.equal(managerOptionValue(value), expected)
for (const value of [undefined, null, {}, [1], NaN, Infinity]) assert.equal(managerOptionValue(value), '未采集')
for (const field of ['type', 'level', 'default_value', 'min', 'max', 'desc', 'long_desc', 'enum_allowed', 'flags', 'tags', 'see_also']) assert.ok(source.includes(`'${field}'`))
assert.ok(source.includes('不代表当前守护进程的生效值'))
const page = readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8')
assert.ok(page.includes('moduleSelection?.scope === scope ? moduleSelection.row : null'))
assert.ok(page.includes('<ManagerModuleDetails'))
console.log('Manager module metadata and unknown value checks passed')

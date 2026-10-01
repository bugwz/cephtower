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

const page = tree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'ConfigurationPage')
const mutationCode = ts.transpileModule(page.body.statements.filter((node) => ts.isFunctionDeclaration(node) && ['run', 'collect', 'save', 'remove'].includes(node.name.text)).map((node) => node.getText(tree)).join('\n'), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const action of ['save', 'remove']) for (const timing of ['before', 'mutation', 'collection', 'unchanged']) {
  const scope = { clusterId: 7, moduleName: 'test' }, scopeRef = { current: scope }, events = []
  let modal, finishMutation, finishCollection
  const env = {
    scope, scopeRef, selectedClusterId: 7, running: { current: false }, editing: { resource_version: 3 },
    setBusy: () => {}, setOpen: () => { events.push('close') }, message: { success: () => { events.push('success') } },
    mutateResource: async (path, method, body, options) => {
      assert.equal(path, '/configuration/value'); assert.equal(body.cluster_id, 7); assert.equal(options.ifMatch, '3')
      assert.equal(method, action === 'save' ? 'PUT' : 'DELETE')
      events.push('mutate'); await new Promise((resolve) => { finishMutation = resolve })
    },
    refreshResource: async (body) => {
      assert.deepEqual(body, { clusterId: 7, kinds: ['config_value', 'config_option'] })
      events.push('collect'); await new Promise((resolve) => { finishCollection = resolve })
    },
    refresh: async () => { events.push('read') }, Modal: { confirm: (options) => { modal = options } }
  }
  const functions = new Function('env', `const { ${Object.keys(env).join(', ')} } = env; ${mutationCode}; return { save, remove, run }`)(env)
  if (action === 'remove') functions.remove({ who: 'global', name: 'test', resource_version: 3 })
  const invoke = () => action === 'save' ? functions.save({ who: 'global', name: 'test', value: '0' }) : modal.onOk()
  if (timing === 'before') {
    scopeRef.current = { ...scope }
    if (action === 'remove') assert.throws(invoke, /已切换/)
    else await invoke()
    assert.deepEqual(events, []); continue
  }
  const pending = invoke()
  await functions.run(async () => assert.fail('synchronous lock must reject overlapping work'))
  if (timing === 'mutation') scopeRef.current = { ...scope }
  finishMutation()
  for (let i = 0; i < 5; i++) await Promise.resolve()
  if (timing === 'mutation') {
    await pending; assert.deepEqual(events, ['mutate']); continue
  }
  if (timing === 'collection') scopeRef.current = { ...scope }
  finishCollection(); await pending
  const expected = action === 'save' ? ['mutate', 'close', 'success', 'collect'] : ['mutate', 'success', 'collect']
  if (timing === 'unchanged') expected.push('read')
  assert.deepEqual(events, expected)
  assert.equal(env.running.current, false)
}
console.log('Configuration mutation scope checks passed')

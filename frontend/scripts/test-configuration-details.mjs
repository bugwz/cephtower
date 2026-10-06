import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
import './test-mon-public-addresses.mjs'

const source = readFileSync(new URL('../src/pages/cluster/ConfigurationPage.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('ConfigurationPage.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const names = ['configurationOverrides', 'configurationList', 'configurationRuntime', 'filterConfigurationOptions', 'configurationMetadataBatch', 'configurationWriteBlocked', 'configurationMonWriteBlocked', 'configurationHelpDescription', 'currentConfigurationHelp', 'localizedConfigurationTarget']
const code = ts.transpileModule(tree.statements.filter((node) => ts.isFunctionDeclaration(node) && names.includes(node.name.text)).map((node) => node.getText(tree)).join('\n'), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const helpers = new Function(`${code}; return { ${names.join(', ')} }`)()
assert.equal(helpers.localizedConfigurationTarget('mgr/dashboard/server_port', 'node-a.x1'), 'mgr/dashboard/node-a.x1/server_port')
assert.equal(helpers.localizedConfigurationTarget('mgr/dashboard/server_port', ''), 'mgr/dashboard/server_port')
assert.equal(helpers.localizedConfigurationTarget('mgr/dashboard/node-a.x1/server_port', undefined), 'mgr/dashboard/node-a.x1/server_port')
for (const instance of ['../', '..', '-flag', 'a/b', 'a b', null, 1, 'a'.repeat(257)]) assert.equal(helpers.localizedConfigurationTarget('mgr/dashboard/server_port', instance), undefined)
assert.equal(helpers.localizedConfigurationTarget('osd_memory_target', 'a'), undefined)
assert.equal(helpers.localizedConfigurationTarget('mgr/dashboard/a/server_port', 'b'), undefined)
{
  const scope = { clusterId: 1, moduleName: undefined }
  for (const help of [{ name: 'test', type: 'bool' }, { name: 'test', enum_values: ['a', 'b'] }, null]) {
    const snapshot = { scope, name: 'test', help, error: help ? '' : 'failed' }
    assert.equal(helpers.currentConfigurationHelp(snapshot, scope, 'test', true), snapshot)
    assert.equal(helpers.currentConfigurationHelp(snapshot, scope, 'other', true), null)
    assert.equal(helpers.currentConfigurationHelp(snapshot, { ...scope }, 'test', true), null)
    assert.equal(helpers.currentConfigurationHelp(snapshot, scope, 'test', false), null)
  }
  assert.equal(helpers.currentConfigurationHelp(null, scope, 'test', true), null)
  assert.equal(helpers.currentConfigurationHelp({ scope, name: 'test', help: { name: 'other' }, error: '' }, scope, 'test', true), null)
  assert.ok(source.includes('currentConfigurationHelp(helpSnapshot, scope, watchedName, open)'))
  assert.ok(source.includes('const help = currentHelp?.help ?? null'))
  assert.ok(source.includes("const helpError = currentHelp?.error ?? ''"))
  assert.ok(source.includes("if (value.name !== watchedName) throw new Error"))
}
const editable = { who: 'global', name: 'test', stale: false, resource_version: 3 }
assert.equal(helpers.configurationWriteBlocked(editable), undefined)
assert.equal(helpers.configurationWriteBlocked({ ...editable, resource_version: '3' }), undefined)
for (const resource_version of ['9007199254740993', '18446744073709551615']) assert.equal(helpers.configurationWriteBlocked({ ...editable, resource_version }), undefined)
for (const resource_version of ['18446744073709551616', '01', '+1', '1e3', ' 1', '1.0', '0']) assert.ok(helpers.configurationWriteBlocked({ ...editable, resource_version }))
const invalidRows = [
  { ...editable, stale: true }, { ...editable, stale: undefined },
  { ...editable, who: '' }, { ...editable, name: undefined },
  ...[undefined, null, 0, -1, 1.5, true, {}, '', 'bad', Number.MAX_SAFE_INTEGER + 1].map((resource_version) => ({ ...editable, resource_version }))
]
for (const row of invalidRows) assert.ok(helpers.configurationWriteBlocked(row))
const protectedOption = { name: 'test', flags: ['no_mon_update'], can_update_at_runtime: true }
assert.equal(helpers.configurationMonWriteBlocked(protectedOption, 'test'), true)
assert.equal(helpers.configurationMonWriteBlocked(protectedOption, 'other'), false)
assert.equal(helpers.configurationMonWriteBlocked(null, 'test'), false)
assert.match(helpers.configurationHelpDescription(protectedOption), /不能通过 Monitor/)
assert.match(helpers.configurationHelpDescription({ can_update_at_runtime: false }), /可能需要重启/)
for (const value of [undefined, null, 'false']) assert.match(helpers.configurationHelpDescription({ can_update_at_runtime: value }), /未采集/)
assert.equal(helpers.configurationHelpDescription({ can_update_at_runtime: true, default: false }), '默认值：false')
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
const options = [{ name: 'option' }, { name: 'unset' }, { name: 'other' }]
assert.deepEqual(helpers.filterConfigurationOptions(options, rows, 'configured', true), [options[0], options[2]])
assert.deepEqual(helpers.filterConfigurationOptions(options, rows, 'unconfigured', true), [options[1]])
assert.deepEqual(helpers.filterConfigurationOptions(options, [], 'unconfigured', true), options)
for (const filter of ['configured', 'unconfigured']) assert.deepEqual(helpers.filterConfigurationOptions(options, rows, filter, false), [])
assert.deepEqual(helpers.filterConfigurationOptions(options, rows, 'all', false), options)
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

{
  let inflight = 0, maximum = 0
  const calls = []
  const result = await helpers.configurationMetadataBatch(['a', 'b', 'c', 'd', 'e', 'f', 'a'], async (name) => {
    calls.push(name); maximum = Math.max(maximum, ++inflight)
    await Promise.resolve(); inflight--
    if (name === 'b') throw new Error('unavailable')
    return { name: name === 'c' ? 'mismatch' : name, default: '0' }
  }, () => true)
  assert.equal(maximum, 4)
  assert.equal(calls.length, 6)
  assert.deepEqual(Object.keys(result.items).sort(), ['a', 'd', 'e', 'f'])
  assert.deepEqual(result.failed.sort(), ['b', 'c'])
  assert.equal(result.items.a.default, '0')
}
{
  let active = true
  const pending = [], calls = []
  const result = helpers.configurationMetadataBatch(['a', 'b', 'c', 'd', 'e'], (name) => {
    calls.push(name); return new Promise((resolve) => pending.push(() => resolve({ name })))
  }, () => active)
  assert.equal(calls.length, 4)
  active = false
  pending.forEach((resolve) => resolve())
  await result
  assert.equal(calls.length, 4, 'scope cancellation must stop queued reads')
}
assert.ok(source.includes('scopeRef.current !== scope || controller.signal.aborted'))
assert.ok(source.includes('names.forEach((name) => { delete items[name] })'))
assert.ok(source.includes('filteredOptions.slice((currentPage - 1) * optionPageSize, currentPage * optionPageSize)'))

const page = tree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'ConfigurationPage')
const mutationCode = ts.transpileModule(page.body.statements.filter((node) => ts.isFunctionDeclaration(node) && ['run', 'collect', 'refreshAfterMutation', 'save', 'remove'].includes(node.name.text)).map((node) => node.getText(tree)).join('\n'), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const version of [3, '9007199254740993', '18446744073709551615']) for (const action of ['save', 'remove']) for (const timing of ['before', 'mutation', 'collection', 'unchanged']) {
  const target = { ...editable, resource_version: version }
  const scope = { clusterId: 7, moduleName: 'test' }, scopeRef = { current: scope }, events = []
  let modal, finishMutation, finishCollection
  const env = {
    scope, scopeRef, selectedClusterId: 7, running: { current: false }, editing: target,
    loading: false, error: null, configurationWriteBlocked: helpers.configurationWriteBlocked, help: null, configurationMonWriteBlocked: helpers.configurationMonWriteBlocked, localizedConfigurationTarget: helpers.localizedConfigurationTarget,
    setBusy: () => {}, setOpen: () => { events.push('close') }, message: { success: () => { events.push('success') } },
    mutateResource: async (path, method, body, options) => {
      assert.equal(path, '/configuration/value'); assert.equal(body.cluster_id, 7); assert.equal(options.ifMatch, String(version))
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
  if (action === 'remove') functions.remove(target)
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

for (const instance of ['', 'node-a.x1', '../bad']) {
  const scope = {}, bodies = []
  const env = {
    ...helpers, scope, scopeRef: { current: scope }, selectedClusterId: 7, running: { current: false }, editing: null,
    loading: false, error: null, help: null, setBusy: () => {}, setOpen: () => {},
    message: { success: () => {}, warning: () => {}, error: () => {} },
    mutateResource: async (_path, _method, body) => bodies.push(body), refreshResource: async () => {}, refresh: async () => {}
  }
  const save = new Function('env', `const { ${Object.keys(env).join(', ')} } = env; ${mutationCode}; return save`)(env)
  await save({ who: 'mgr', name: 'mgr/dashboard/server_port', instance, value: '8443' })
  assert.deepEqual(bodies, instance === '../bad' ? [] : [{ cluster_id: 7, who: 'mgr', name: instance ? `mgr/dashboard/${instance}/server_port` : 'mgr/dashboard/server_port', value: '8443' }])
}

for (const action of ['save', 'remove']) for (const collectionFailed of [false, true]) for (const mutationFailed of [false, true]) {
  const scope = {}, events = []
  let modal
  const env = {
    scope, scopeRef: { current: scope }, selectedClusterId: 7, running: { current: false }, editing: editable,
    loading: false, error: null, configurationWriteBlocked: helpers.configurationWriteBlocked, help: null, configurationMonWriteBlocked: helpers.configurationMonWriteBlocked, localizedConfigurationTarget: helpers.localizedConfigurationTarget,
    setBusy: () => {}, setOpen: () => events.push('close'),
    message: { success: () => events.push('success'), warning: (text) => { assert.match(text, /修改已执行.*不要重复提交/); events.push('warning') } },
    mutateResource: async () => { events.push('mutate'); if (mutationFailed) throw new Error('write failed') },
    refreshResource: async () => { events.push('collect'); if (collectionFailed) throw new Error('collection failed') },
    refresh: async () => { events.push('read') }, Modal: { confirm: (options) => { modal = options } }
  }
  const functions = new Function('env', `const { ${Object.keys(env).join(', ')} } = env; ${mutationCode}; return { save, remove }`)(env)
  functions.remove(editable)
  const invoke = () => action === 'save' ? functions.save({ who: 'global', name: 'test', value: '0' }) : modal.onOk()
  if (mutationFailed) {
    await assert.rejects(invoke, /write failed/)
    assert.deepEqual(events, ['mutate'])
  } else {
    await invoke()
    const expected = action === 'save' ? ['mutate', 'close', 'success', 'collect'] : ['mutate', 'success', 'collect']
    if (collectionFailed) expected.push('warning')
    expected.push('read')
    assert.deepEqual(events, expected)
  }
  assert.equal(env.running.current, false)
}
console.log('Configuration collection failures remain distinct from mutation failures')
{
  let reported = false
  const save = new Function('configurationMonWriteBlocked', 'help', 'message', `const loading=false, error=null, editing=null; ${mutationCode}; return save`)(helpers.configurationMonWriteBlocked, protectedOption, { error: () => { reported = true } })
  await save({ name: 'test', who: 'global', value: 'x' })
  assert.equal(reported, true, 'protected options must stop before a request is sent')
}
for (const editing of [...invalidRows, editable]) {
  let errors = 0
  const env = { editing, loading: false, error: null, configurationWriteBlocked: helpers.configurationWriteBlocked, message: { error: () => { errors++ } } }
  const functions = new Function('env', `const { ${Object.keys(env).join(', ')} } = env; ${mutationCode}; return { save, remove }`)(env)
  await functions.save({ who: 'different', name: 'test', value: 'x' })
  assert.equal(errors, 1, 'invalid records or changed identity must stop before acquiring locks or sending requests')
  if (editing !== editable) { functions.remove(editing); assert.equal(errors, 2) }
}

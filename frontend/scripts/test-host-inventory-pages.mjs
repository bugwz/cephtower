import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/cluster/HostPage.tsx', import.meta.url), 'utf8')
const file = ts.createSourceFile('HostPage.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const page = file.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'HostPage')
const nicNode = file.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'hostNICCount')
const nicCode = ts.transpileModule(nicNode.getText(file).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const nicCount = new Function('isRecord', `${nicCode}; return hostNICCount`)(value => value !== null && typeof value === 'object' && !Array.isArray(value))
for (const value of [0, 4, '0', '4', '18446744073709551615']) assert.equal(nicCount({ native_summary: { nic_count: value } }), String(value))
for (const value of [null, undefined, 'N/A', '', -1, 1.5, Number.MAX_SAFE_INTEGER + 1, {}, [], true, '04']) assert.equal(nicCount({ native_summary: { nic_count: value } }), '未知')
assert.equal(nicCount({}), '未知')
assert.ok(source.includes('render: (_value, row) => hostNICCount(row)'))
let nativeColumns
function findColumns(node) {
  if (ts.isSpreadElement(node) && node.expression.getText(file).startsWith("['server', 'cpu_summary'")) nativeColumns = node.expression
  ts.forEachChild(node, findColumns)
}
findColumns(file)
assert.ok(nativeColumns)
const columnCode = ts.transpileModule(`const columns = ${nativeColumns.getText(file)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const columns = new Function('isRecord', 'textValue', `${columnCode}; return columns`)(value => value !== null && typeof value === 'object' && !Array.isArray(value), value => value == null ? '-' : String(value))
assert.equal(columns.length, 6)
for (const column of columns) {
  const key = column.key.replace('native_', '')
  assert.equal(column.render(null, { native_summary: { [key]: '12/120 TB' } }), '12/120 TB')
  assert.equal(column.render(null, { native_summary: { [key]: 'N/A' } }), 'N/A')
  assert.equal(column.render(null, {}), '-')
  assert.equal(column.filterKey, false)
}
const declaration = page.body.statements.filter(ts.isVariableStatement).flatMap(n => [...n.declarationList.declarations]).find(n => n.name.getText(file) === 'loader')
const callback = declaration.initializer.arguments[0].getText(file)
const resource = readFileSync(new URL('../src/api/resource.ts', import.meta.url), 'utf8')
const resourceFile = ts.createSourceFile('resource.ts', resource, ts.ScriptTarget.Latest, true)
const all = resourceFile.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'listAllResources').getText(resourceFile).replace('export ', '')
const compile = code => ts.transpileModule(code, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const fail of [false, true]) {
  const calls = []
  const listResource = async (path, cluster, options) => {
    calls.push({ path, cluster, options })
    if (fail && path === '/devices' && options.cursor) throw new Error('second page unavailable')
    return { items: [{ name: `${path}:${options.cursor || 'first'}` }], nextCursor: options.cursor ? null : 'second', stale: path === '/daemons' && Boolean(options.cursor), staleReason: path === '/daemons' && options.cursor ? 'daemon snapshot stale' : null }
  }
  const loadAll = new Function('listResource', `${compile(all)}; return listAllResources`)(listResource)
  const load = new Function('listAllResources', 'selectedClusterId', 'hostTableFilters', 'normalizeHostRow', `${compile(`const load = ${callback}`)}; return load`)(loadAll, 42, { filters: { status: ['maintenance'] } }, (host, daemons, devices) => ({ host, daemons, devices }))
  if (fail) await assert.rejects(load(), /second page unavailable/)
  else {
    const result = await load()
    assert.equal(result.hosts.length, 2)
    assert.equal(result.hosts[0].daemons.length, 2)
    assert.equal(result.hosts[0].devices.length, 2)
    assert.equal(result.stale, true)
    assert.match(result.staleReason, /daemon snapshot stale/)
  }
  assert.equal(calls.length, 6)
  assert.ok(calls.every(call => call.cluster === 42))
  assert.ok(calls.filter(call => call.path === '/hosts').every(call => call.options.filters.status[0] === 'maintenance'))
}
console.log('Host inventory consumes every page with fixed cluster scope and stale metadata')

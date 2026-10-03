import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const compile = name => ts.transpileModule(readFileSync(new URL(`../src/pages/object/${name}`, import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText
const model = {}
new Function('exports', compile('rgwUserKeyRows.ts'))(model)
const rows = model.rgwUserKeyRows
assert.deepEqual(rows([]), [])
for (const value of [undefined, null, {}, '', [null], [[]], [false]]) assert.equal(rows(value), undefined)
const input = [{ user: 'tenant$user:sub', active: false, access_key: 'must-not-render-access', secret_key: 'must-not-render-secret' }]
assert.deepEqual(rows(input), [{ key: 0, user: 'tenant$user:sub', state: '未启用' }])
assert.ok(!JSON.stringify(rows(input)).includes('must-not-render'))
assert.equal(rows([{ user: 'test', active: true }])[0].state, '已启用')
for (const active of [undefined, null, 0, 1, 'false', 'true']) assert.equal(rows([{ user: 'test', active }])[0].state, '启用状态未返回或无效')
for (const user of [undefined, null, false, '', 1]) assert.equal(rows([{ user }])[0].user, '所属用户未返回或无效')
const view = {}
const jsx = (type, props) => ({ type, props })
new Function('exports', 'require', compile('RgwUserKeys.tsx'))(view, name => {
  if (name === 'react/jsx-runtime') return { jsx, jsxs: jsx }
  if (name === 'antd') return { Table: 'Table' }
  if (name === './rgwUserKeyRows') return model
  throw new Error(name)
})
for (const protocol of ['S3', 'Swift']) {
  assert.match(view.RgwUserKeyTable({ protocol }).props.children.join(''), /未返回/)
  const empty = view.RgwUserKeyTable({ protocol, value: [] })
  assert.equal(empty.props.locale.emptyText, `未配置 ${protocol} 密钥`)
  assert.equal(empty.props.pagination, false)
  const table = view.RgwUserKeyTable({ protocol, value: input })
  assert.deepEqual(table.props.dataSource, rows(input))
  assert.deepEqual(table.props.columns.map(column => column.dataIndex), ['user', 'state'])
  assert.deepEqual(view.RgwUserKeyTable({ protocol, value: Array(6).fill(input[0]) }).props.pagination, { pageSize: 5 })
}
const row = { keys: input, swift_keys: [] }
const children = view.RgwUserKeys({ row }).props.children
assert.equal(children[2].props.value, row.keys)
assert.equal(children[2].props.protocol, 'S3')
assert.equal(children[4].props.value, row.swift_keys)
assert.equal(children[4].props.protocol, 'Swift')
console.log('RGW key metadata preserves protocol, identity and unknown states without secrets')

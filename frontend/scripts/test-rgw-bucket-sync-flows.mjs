import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api = {}
const jsx = (type, props) => ({ type, props })
new Function('exports', 'require', ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwBucketSyncFlows.tsx', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText)(api, name => name === 'antd' ? { Table: 'Table' } : { jsx, jsxs: jsx })
const policy = data_flow => ({ groups: [{ id: '<group>', data_flow }] })
const flow = {
  symmetrical: [{ id: '双向', zones: [' zone-a ', '<zone-b>'] }],
  directional: [{ source_zone: 'zone-b', dest_zone: 'zone-c' }]
}
const groups = api.bucketSyncFlows(policy(flow))
assert.equal(groups[0].rows.length, 2)
assert.deepEqual(groups[0].rows[0], { index: 0, kind: '对称', id: '"双向"', source: '不适用', destination: '不适用', zones: '" zone-a " / "<zone-b>"' })
assert.deepEqual(groups[0].rows[1], { index: 1, kind: '定向', id: '原生数据无 ID', source: '"zone-b"', destination: '"zone-c"', zones: '不适用' })
assert.equal(api.bucketSyncFlows(policy({}))[0].unavailable, false)
assert.deepEqual(api.bucketSyncFlows({ groups: [] }), [])
for (const value of [null, {}, [], { groups: null }, { groups: [null] }, { groups: [{ id: 5 }] }, { groups: [{ id: 'g' }, { id: 'g' }] }]) assert.equal(api.bucketSyncFlows(value), undefined)
for (const malformed of [null, [], { symmetrical: null }, { directional: {} }, { symmetrical: [null] }, { symmetrical: [{ id: 'g', zones: [1] }] }, { directional: [{ source_zone: 'x' }] }]) assert.equal(api.bucketSyncFlows(policy(malformed))[0].unavailable, true)
assert.deepEqual(api.bucketSyncFlows(policy({ future: [] }))[0].extensions, ['future'])
assert.equal(api.bucketSyncFlows(policy({ ...flow, directional: [...flow.directional, null] }))[0].rows.length, 2)
function nodes(node) {
  if (Array.isArray(node)) return node.flatMap(nodes)
  if (!node || typeof node !== 'object') return []
  return [node, ...nodes(node.props?.children)]
}
const view = api.RgwBucketSyncFlows({ value: policy(flow) })
const all = nodes(view)
const table = all.find(node => node.type === 'Table')
assert.equal(table.props.dataSource.length, 2)
assert.equal(table.props.columns.length, 5)
assert.ok(all.every(node => node.props?.dangerouslySetInnerHTML === undefined))
for (const [value, pattern] of [[policy({}), /此组无桶本地/], [policy({ future: [] }), /不推断/], [policy(null), /不推断/]]) {
  const t = nodes(api.RgwBucketSyncFlows({ value })).find(node => node.type === 'Table')
  assert.match(t.props.locale.emptyText, pattern)
}
const many = { symmetrical: Array.from({ length: 6 }, (_, i) => ({ id: String(i), zones: [] })) }
assert.equal(nodes(api.RgwBucketSyncFlows({ value: policy(many) })).find(node => node.type === 'Table').props.pagination.pageSize, 5)
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const zoneView = api.RgwBucketSyncFlows({ value: policy(flow), scope: 'zonegroup' })
assert.equal(nodes(zoneView).find(node => node.type === 'Table').props.dataSource[0].zones, '" zone-a " / "<zone-b>"')
assert.match(JSON.stringify(zoneView), /Zonegroup get 返回 Zone ID/)
assert.doesNotMatch(JSON.stringify(zoneView), /桶本地/)
assert.match(JSON.stringify(api.RgwBucketSyncFlows({ value: null, scope: 'zonegroup' })), /Zonegroup/)
assert.match(nodes(api.RgwBucketSyncFlows({ value: policy({}), scope: 'zonegroup' })).find(node => node.type === 'Table').props.locale.emptyText, /无Zonegroup数据流/)
assert.match(pages, /<RgwBucketSyncFlows value=\{value\} scope="zonegroup" \/>/)
assert.match(pages, /<RgwBucketSyncFlows value=\{value\} \/>/)
assert.match(pages, /JSON.stringify\(value, null, 2\)/)
console.log('bucket native symmetrical and directional flow presentation checks passed')

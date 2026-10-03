import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api = {}
const jsx = (type, props) => ({ type, props })
new Function('exports', 'require', ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwBucketSyncPipes.tsx', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText)(api, name => name === 'antd' ? { Table: 'Table' } : { jsx, jsxs: jsx })
const pipe = {
  id: '<pipe>', source: { bucket: 'tenant/photos:marker', zones: ['*'] }, dest: { bucket: '*', zones: [' zone ', '<zone>'] },
  params: { source: { filter: { prefix: '', tags: [{ key: '<key>', value: '' }] } }, dest: { acl_translation: { owner: 'team$user' }, storage_class: 'COLD' }, priority: 0, mode: 'user', user: 'team$user', extension: { enabled: true } }
}
const policy = pipes => ({ groups: [{ id: 'g', pipes }] })
const row = api.bucketSyncPipes(policy([pipe]))[0].rows[0]
assert.equal(row.id, '"<pipe>"')
assert.equal(row.sourceBucket, '"tenant/photos:marker"')
assert.equal(row.destBucket, '"*"')
assert.match(row.sourceZones, /所有 Zone.*数据流/)
assert.equal(row.destZones, '" zone " / "<zone>"')
assert.deepEqual(row.params, pipe.params)
for (const [zones, pattern] of [[undefined, /未返回.*不推断/], [[], /列表为空/], [null, /格式不可用/], [[4], /格式不可用/]]) {
  assert.match(api.bucketSyncPipes(policy([{ ...pipe, source: { ...pipe.source, zones } }]))[0].rows[0].sourceZones, pattern)
}
for (const value of [null, {}, [], { groups: null }, { groups: [null] }, { groups: [{ id: 'g' }, { id: 'g' }] }]) assert.equal(api.bucketSyncPipes(value), undefined)
assert.deepEqual(api.bucketSyncPipes({ groups: [] }), [])
assert.equal(api.bucketSyncPipes(policy([]))[0].unavailable, false)
for (const pipes of [null, {}, [null], [{ ...pipe, dest: null }], [pipe, pipe]]) assert.equal(api.bucketSyncPipes(policy(pipes))[0].unavailable, true)
const summary = api.bucketPipeParameters(pipe.params)
assert.match(summary, /用户模式.*\n用户："team\$user".*\n优先级：0/)
assert.match(summary, /源前缀：""/)
assert.match(summary, /<key>.*value/)
assert.match(summary, /目标 ACL 转换.*team\$user/)
assert.match(summary, /目标存储类："COLD"/)
assert.match(api.bucketPipeParameters({ ...pipe.params, mode: 'constructor' }), /未知/)
assert.match(api.bucketPipeParameters({ ...pipe.params, mode: 'system', priority: Number.MAX_SAFE_INTEGER + 1 }), /系统模式.*\n.*\n优先级：不可用/)
assert.match(api.bucketPipeParameters(null), /不可用/)
assert.match(api.bucketPipeParameters({}), /未返回/)
function nodes(node) {
  if (Array.isArray(node)) return node.flatMap(nodes)
  if (!node || typeof node !== 'object') return []
  return [node, ...nodes(node.props?.children)]
}
const table = nodes(api.RgwBucketSyncPipes({ value: policy([pipe]) })).find(node => node.type === 'Table')
assert.equal(table.props.columns.length, 6)
assert.equal(table.props.dataSource[0].params.extension.enabled, true)
const details = table.props.columns[5].render(pipe.params)
assert.ok(nodes(details).every(node => !node.props?.dangerouslySetInnerHTML))
assert.match(JSON.stringify(details), /extension/)
for (const [pipes, pattern] of [[[], /此组无/], [null, /不推断/]]) assert.match(nodes(api.RgwBucketSyncPipes({ value: policy(pipes) })).find(node => node.type === 'Table').props.locale.emptyText, pattern)
assert.equal(nodes(api.RgwBucketSyncPipes({ value: policy(Array.from({ length: 6 }, (_, i) => ({ ...pipe, id: String(i) }))) })).find(node => node.type === 'Table').props.pagination.pageSize, 5)
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.match(pages, /<RgwBucketSyncPipes value=\{value\} \/>/)
console.log('bucket native sync pipe presentation checks passed')

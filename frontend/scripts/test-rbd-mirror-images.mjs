import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const ui = {}, jsx = (type, props) => ({ type, props })
const source = readFileSync(new URL('../src/pages/block/RbdMirrorImages.tsx', import.meta.url), 'utf8')
new Function('exports', 'require', ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText)(ui, name => name === 'antd' ? Object.fromEntries(['Alert', 'Space', 'Table', 'Descriptions'].map(name => [name, name])) : { jsx, jsxs: jsx })
const daemon = { service_id: 'svc', instance_id: '9007199254740993', daemon_id: 'client.mirror', hostname: 'host' }
const peer = { site_name: '', mirror_uuid: 'remote', state: 'down+unknown', description: '<script>native</script>', last_update: '2026-01-01 12:00:00' }
const sample = { name: 'image', global_id: 'global', state: 'up+replaying', description: 'replaying', last_update: '2026-01-01 12:00:01', daemon_service: daemon, peer_sites: [peer] }
const rows = ui.rbdMirrorImageRows([sample, {}])
assert.equal(rows[0].state, 'up+replaying')
assert.equal(rows[0].daemon.instance_id, '9007199254740993')
assert.equal(rows[0].peers[0].name, '')
assert.equal(rows[0].peers[0].updated, peer.last_update)
assert.equal(rows[1].state, '未返回或无效')
assert.equal(rows[1].peers, undefined)
for (const value of [null, undefined, {}, [null], [[]]]) assert.equal(ui.RbdMirrorImages({ value }).props.type, 'warning')
assert.equal(ui.RbdMirrorImages({ value: [] }).props.type, 'info')
for (const peers of [null, {}, [null], [[]]]) assert.equal(ui.rbdMirrorImageRows([{ peer_sites: peers }])[0].peers, undefined)
const table = ui.RbdMirrorImages({ value: [sample] }).props.children.find(node => node.type === 'Table')
assert.equal(table.props.columns.length, 6)
for (const state of ['syncing', 'starting_replay', 'replaying']) assert.equal(ui.rbdMirrorStateCategory(`up+${state}`), '同步或重放中')
for (const state of ['stopping_replay', 'stopped']) assert.equal(ui.rbdMirrorStateCategory(`up+${state}`), '停止中或已停止')
for (const state of ['unknown', 'unknown (42)', 'error', 'syncing', 'starting_replay', 'replaying', 'stopping_replay', 'stopped']) assert.equal(ui.rbdMirrorStateCategory(`down+${state}`), '需关注')
assert.equal(ui.rbdMirrorStateCategory('up+error'), '需关注')
for (const state of [undefined, null, 0, '', 'up+unknown', 'up+unknown (42)', 'up+future', 'down+future', 'up+replaying ', 'UP+replaying', 'up+replaying\n']) assert.equal(ui.rbdMirrorStateCategory(state), '未知或无效')
assert.equal(rows[0].category, '同步或重放中')
assert.equal(rows[0].peers[0].category, '需关注')
assert.equal(rows[1].category, '未知或无效')
const classification = table.props.columns.at(-1)
assert.equal(classification.filters.length, 4)
assert.equal(classification.onFilter('需关注', rows[0]), false)
assert.equal(classification.onFilter('同步或重放中', rows[0]), true)
const summary = ui.RbdMirrorImages({value:[sample,{}, {state:'up+stopped'}, {state:'down+error'}]}).props.children[1].props.children.join('')
assert.ok(summary.includes('需关注 1；同步或重放中 1；停止中或已停止 1；未知或无效 1'))
const expanded = table.props.expandable.expandedRowRender(rows[0]).props.children
assert.equal(expanded[1].props.columns.at(-1).onFilter('需关注', rows[0].peers[0]), true)
assert.deepEqual(expanded[0].props.items.map(item => item.children.props.children), Object.values(daemon))
assert.equal(expanded[1].props.columns[0].render(''), '名称未解析')
assert.equal(expanded[1].props.columns[3].render(peer.description).props.children, peer.description)
const missing = table.props.expandable.expandedRowRender(rows[1]).props.children
assert.ok(missing.every(node => node.type === 'Alert'))
const empty = table.props.expandable.expandedRowRender(ui.rbdMirrorImageRows([{peer_sites:[]}])[0]).props.children
assert.equal(empty[1].props.message, '远端站点状态列表为空')
assert.ok(readFileSync(new URL('../src/pages/block/pages.tsx', import.meta.url), 'utf8').includes('<RbdMirrorImages value={value} />'))
console.log('RBD mirror image local, remote, and daemon fields preserve native states')

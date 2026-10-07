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
assert.equal(table.props.columns.length, 10)
for (const mode of ['journal', 'snapshot']) {
  const row = ui.rbdMirrorImageRows([{mirror_mode:mode, mirror_primary:false, mirror_image_state:'enabled'}])[0]
  assert.equal(row.mode,mode)
  assert.equal(row.role,'非主端')
  assert.equal(row.mirrorState,'enabled')
}
assert.equal(ui.rbdMirrorImageRows([{mirror_primary:true}])[0].role,'主端')
for(const state of ['creating','disabling']) assert.equal(ui.rbdMirrorImageRows([{mirror_image_state:state}])[0].mirrorState,state)
for (const value of [undefined,null,0,'false',{},[]]) {
  const row = ui.rbdMirrorImageRows([{mirror_mode:value,mirror_primary:value,mirror_image_state:value}])[0]
  assert.equal(row.mode,'未返回有效模式')
  assert.equal(row.role,'未返回有效角色')
  assert.equal(row.mirrorState,'未返回有效配置状态')
}
for (const state of ['syncing', 'starting_replay', 'replaying']) assert.equal(ui.rbdMirrorStateCategory(`up+${state}`), '同步或重放中')
for (const state of ['stopping_replay', 'stopped']) assert.equal(ui.rbdMirrorStateCategory(`up+${state}`), '停止中或已停止')
for (const state of ['unknown', 'unknown (42)', 'error', 'syncing', 'starting_replay', 'replaying', 'stopping_replay', 'stopped']) assert.equal(ui.rbdMirrorStateCategory(`down+${state}`), '需关注')
assert.equal(ui.rbdMirrorStateCategory('up+error'), '需关注')
for (const state of [undefined, null, 0, '', 'up+unknown', 'up+unknown (42)', 'up+future', 'down+future', 'up+replaying ', 'UP+replaying', 'up+replaying\n']) assert.equal(ui.rbdMirrorStateCategory(state), '未知或无效')
assert.equal(rows[0].category, '同步或重放中')
assert.equal(rows[0].peers[0].category, '需关注')
assert.equal(rows[1].category, '未知或无效')
const classification = table.props.columns.find(column => column.dataIndex === 'category')
assert.equal(classification.filters.length, 4)
assert.equal(classification.onFilter('需关注', rows[0]), false)
assert.equal(classification.onFilter('同步或重放中', rows[0]), true)
const summary = ui.RbdMirrorImages({value:[sample,{}, {state:'up+stopped'}, {state:'down+error'}]}).props.children[1].props.children.join('')
assert.ok(summary.includes('需关注 1；同步或重放中 1；停止中或已停止 1；未知或无效 1'))
const expanded = table.props.expandable.expandedRowRender(rows[0]).props.children
assert.equal(expanded[1].props.columns.find(column => column.dataIndex === 'category').onFilter('需关注', rows[0].peers[0]), true)
assert.deepEqual(expanded[0].props.items.map(item => item.children.props.children), Object.values(daemon))
assert.equal(expanded[1].props.columns[0].render(''), '名称未解析')
assert.equal(expanded[1].props.columns[3].render(peer.description).props.children, peer.description)
const missing = table.props.expandable.expandedRowRender(rows[1]).props.children
assert.ok(missing.slice(0,2).every(node => node.type === 'Alert'))
const empty = table.props.expandable.expandedRowRender(ui.rbdMirrorImageRows([{peer_sites:[]}])[0]).props.children
assert.equal(empty[1].props.message, '远端站点状态列表为空')
const metrics = {bytes_per_second:'0', seconds_until_synced:'18446744073709551615', syncing_percent:'0.5', entries_behind_primary:'9007199254740993'}
assert.deepEqual(ui.RbdReplayMetrics({value:metrics}).props.items.map(item=>item.children), [...Object.values(metrics),'未返回有效值'])
assert.equal(ui.RbdReplayMetrics({value:{replay_state:'idle'}}).props.items.at(-1).children,'idle（快照重放空闲，不代表所有站点同步完成）')
assert.equal(ui.rbdReplayStateText('syncing'),'syncing（快照复制中）')
for(const value of [undefined,null,true,0,'future','IDLE','idle ']) assert.equal(ui.rbdReplayStateText(value),'未返回有效值')
assert.ok(ui.RbdReplayMetrics({value:undefined}).props.items.every(item=>item.children==='未返回有效值'))
const enriched = ui.rbdMirrorImageRows([{...sample,replay_metrics:metrics,peer_sites:[{...peer,replay_metrics:{bytes_per_second:'1'}}]}])[0]
const details = table.props.expandable.expandedRowRender(enriched).props.children
assert.equal(details[2].props.value,metrics)
const lag = (value, mode) => ui.RbdReplayMetrics({value,mode}).props.items.find(item=>item.key==='entries_behind_primary').children
assert.equal(lag(metrics,'snapshot'),'不适用（快照同步模式）')
assert.equal(lag(undefined,'snapshot'),'不适用（快照同步模式）')
assert.equal(lag(metrics,'journal'),'9007199254740993')
assert.equal(lag({entries_behind_primary:'0'},'journal'),'0')
assert.equal(lag(undefined,'journal'),'未返回有效值')
for(const mode of [undefined,null,'future','Snapshot']) assert.equal(lag(undefined,mode),'未返回有效值')
const snapshotRow = ui.rbdMirrorImageRows([{...sample,mirror_mode:'snapshot',replay_metrics:metrics}])[0]
const snapshotDetails = table.props.expandable.expandedRowRender(snapshotRow).props.children
assert.equal(snapshotDetails[2].props.mode,'snapshot')
assert.equal(snapshotDetails[1].props.columns.find(column=>column.dataIndex==='metrics').render(metrics).props.mode,undefined)
assert.equal(details[1].props.columns.find(column=>column.dataIndex==='metrics').render(enriched.peers[0].metrics).props.value.bytes_per_second,'1')
for(const value of ['0','50','100']) assert.equal(ui.RbdBootstrapProgress({value}).props.children,`${value}%（引导复制阶段，不代表整体同步完成）`)
for(const value of [undefined,null,0,'','101','-1','50.5',' 50','50\n']) assert.equal(ui.RbdBootstrapProgress({value}).props.children,'未返回有效复制阶段进度')
const bootstrap = ui.rbdMirrorImageRows([{...sample,bootstrap_percent:'0',peer_sites:[{...peer,bootstrap_percent:'100'}]}])[0]
assert.equal(table.props.columns.find(column=>column.dataIndex==='bootstrap').render(bootstrap.bootstrap).props.value,'0')
assert.equal(details[1].props.columns.find(column=>column.dataIndex==='bootstrap').render(bootstrap.peers[0].bootstrap).props.value,'100')
assert.ok(readFileSync(new URL('../src/pages/block/pages.tsx', import.meta.url), 'utf8').includes('<RbdMirrorImages value={value} mode={row.mode} />'))
for(const value of [undefined,[],[sample]]) assert.equal(ui.RbdMirrorImages({value,mode:'disabled'}).props.message,'池同步已禁用，未查询镜像同步运行状态；不代表池内没有镜像')
for(const mode of [undefined,'image','pool','init-only','future']) assert.equal(ui.RbdMirrorImages({value:undefined,mode}).props.type,'warning')
console.log('RBD mirror image local, remote, and daemon fields preserve native states')

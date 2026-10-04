import assert from 'node:assert/strict'
import './test-rgw-period-commit.mjs'
import './test-rgw-zonegroup-sync-group.mjs'
import './test-rgw-bucket-sync-flows.mjs'
import './test-rgw-bucket-sync-pipes.mjs'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketSyncPolicy.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(api)
const group = { id: 'team-policy', status: 'enabled', pipes: [], data_flow: {} }
const summary = value => api.rgwBucketSyncPolicy(value)
assert.match(summary({ groups: [group] }), /已启用.*不代表 Zonegroup/)
assert.match(summary({ groups: [{ ...group, status: 'allowed' }] }), /允许（未启用）/)
assert.match(summary({ groups: [{ ...group, status: 'forbidden' }] }), /禁止/)
for (const status of ['future', 'constructor', 'Enabled']) assert.match(summary({ groups: [{ ...group, status }] }), /未知/)
assert.match(summary({ groups: [] }), /无桶本地同步组.*不代表/)
for (const value of [null, {}, [], { groups: null }, { groups: [null] }, { groups: [group, group] }, { groups: [{ ...group, status: null }] }, { groups: [{ ...group, pipes: {} }] }, { groups: [{ ...group, data_flow: [] }] }]) assert.match(summary(value), /不可用/)
assert.match(summary({ groups: [{ ...group, id: '<script>' }] }), /"<script>"/)
console.log('bucket local sync policy summary checks passed')
assert.match(api.rgwBucketSyncPolicy({ groups: [group] }, 'zonegroup'), /已启用.*Zonegroup.*period.*最终有效策略/)
assert.match(api.rgwBucketSyncPolicy({ groups: [] }, 'zonegroup'), /无Zonegroup同步组.*空策略不代表停止全部复制/)
assert.match(api.rgwBucketSyncPolicy(null, 'zonegroup'), /Zonegroup同步策略不可用/)
assert.doesNotMatch(api.rgwBucketSyncPolicy({ groups: [group] }, 'zonegroup'), /桶本地/)
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketSyncGroupForm.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(api)
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const source = ts.createSourceFile('pages.tsx', pages, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let action, createAction, deleteAction, flowAction, deleteFlowAction, deletePipeAction, createPipeAction, updateFlowAction, updatePipeAction, pipeZonesAction
function visit(node) {
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(source) === 'title' && p.initializer.text === '编辑桶同步管道 Zone')) {
    const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    pipeZonesAction = new Function(...Object.keys(api), `${code}; return action`)(...Object.values(api))
  }
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(source) === 'title' && p.initializer.text === '编辑桶同步管道配置')) {
    const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    updatePipeAction = new Function(...Object.keys(api), `${code}; return action`)(...Object.values(api))
  }
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(source) === 'title' && p.initializer.text === '编辑桶对称数据流')) {
    const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    updateFlowAction = new Function(...Object.keys(api), `${code}; return action`)(...Object.values(api))
  }
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(source) === 'title' && p.initializer.text === '创建桶同步管道')) {
    const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    createPipeAction = new Function(...Object.keys(api), `${code}; return action`)(...Object.values(api))
  }
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(source) === 'title' && p.initializer.text === '删除桶同步管道')) {
    const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    deletePipeAction = new Function(...Object.keys(api), `${code}; return action`)(...Object.values(api))
  }
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(source) === 'title' && p.initializer.text === '删除桶数据流')) {
    const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    deleteFlowAction = new Function(...Object.keys(api), `${code}; return action`)(...Object.values(api))
  }
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(source) === 'title' && p.initializer.text === '创建桶数据流')) {
    const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    flowAction = new Function(...Object.keys(api), `${code}; return action`)(...Object.values(api))
  }
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(source) === 'title' && p.initializer.text === '删除桶同步组')) {
    const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    deleteAction = new Function(...Object.keys(api), `${code}; return action`)(...Object.values(api))
  }
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(source) === 'title' && p.initializer.text === '修改桶同步组状态')) {
    const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    action = new Function(...Object.keys(api), `${code}; return action`)(...Object.values(api))
  }
  if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(source) === 'title' && p.initializer.text === '创建桶同步组')) {
    const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
    createAction = new Function(...Object.keys(api), `${code}; return action`)(...Object.values(api))
  }
  ts.forEachChild(node, visit)
}
visit(source)
assert.ok(action)
assert.equal(action.path, '/rgw/bucket/sync/group')
assert.equal(action.method, 'PATCH')
const row = { natural_key: 'AGJ1Y2tldA', bucket_sync_policy: { groups: [group] } }
const pipeEditGroup = { ...group, pipes: [{ id: ' 管道 ', source: { bucket: '*', zones: ['A'] }, dest: { bucket: '*', zones: ['*'] }, params: { mode: 'user', user: 'old', priority: 123 } }] }
const pipeEditRow = { ...row, bucket_sync_policy: { groups: [pipeEditGroup] } }
const pipeZonesValues = { bucket_id: row.natural_key, group_id: group.id, pipe_id: ' 管道 ', source_zones_json: '["a","b"]', dest_zones_json: '["*"]', confirm_pipe_zones: 'acknowledged' }
assert.equal(pipeZonesAction.method, 'PATCH')
assert.equal(pipeZonesAction.path, '/rgw/bucket/sync/pipe/zones')
assert.deepEqual(pipeZonesAction.buildBody(pipeZonesValues, 7, pipeEditRow), { cluster_id: 7, bucket_id: row.natural_key, group_id: group.id, pipe_id: ' 管道 ', expected_group: JSON.stringify(pipeEditGroup), source_zones: ['a','b'], dest_zones: ['*'] })
assert.match(pipeZonesAction.confirmation(pipeZonesValues, pipeEditRow), /先增后删.*直接切换.*部分生效.*不自动回滚或重试/)
for (const field of ['source_zones_json', 'dest_zones_json']) {
  for (const value of ['[]','["*","a"]','["a","a"]','["a;b"]','["a=b"]','["a b"]','["-a"]','[null]','{}','broken','["\\ud800"]']) assert.throws(() => pipeZonesAction.buildBody({ ...pipeZonesValues, [field]: value }, 7, pipeEditRow))
}
for (const change of [{ bucket_id: 'wrong' }, { group_id:'wrong' }, { pipe_id:'wrong' }, { confirm_pipe_zones:true }]) assert.throws(() => pipeZonesAction.buildBody({ ...pipeZonesValues, ...change }, 7, pipeEditRow))
assert.ok(pipeZonesAction.disabledWhen({ ...pipeEditRow, stale: true }))
console.log('bucket pipe zone membership form checks passed')
const pipeEditValues = { bucket_id: row.natural_key, group_id: group.id, pipe_id: ' 管道 ', source_bucket: 'photos', dest_bucket: '*', mode: 'system', confirm_pipe_update: 'acknowledged' }
const loadedZones=pipeZonesAction.changedValues({pipe_id:pipeEditValues.pipe_id},pipeEditValues,pipeEditRow)
assert.deepEqual(JSON.parse(loadedZones.source_zones_json),['A'])
assert.deepEqual(JSON.parse(loadedZones.dest_zones_json),['*'])
assert.equal(loadedZones.confirm_pipe_zones,undefined)
assert.equal(loadedZones.pipe_load_error,undefined)
assert.deepEqual(pipeZonesAction.fields.find(f=>f.name==='pipe_id').optionsDependencies,['group_id'])
assert.deepEqual(await pipeZonesAction.fields.find(f=>f.name==='pipe_id').optionsLoader(7,pipeEditRow,pipeEditValues),[{value:' 管道 ',label:' 管道 '}])
assert.equal(pipeZonesAction.changedValues({group_id:'other'},pipeEditValues,pipeEditRow).source_zones_json,undefined)
assert.equal(pipeZonesAction.changedValues({group_id:'other'},pipeEditValues,pipeEditRow).pipe_id,undefined)
assert.deepEqual(pipeZonesAction.changedValues({source_zones_json:'[]'},pipeEditValues,pipeEditRow),{})
for(const zones of [[],['orphan'],['*'],['a','b']]) {
 const current=structuredClone(pipeEditRow)
 current.bucket_sync_policy.groups[0].pipes[0].source.zones=zones
 const filled=pipeZonesAction.changedValues({pipe_id:pipeEditValues.pipe_id},pipeEditValues,current)
 assert.equal(filled.pipe_load_error,undefined)
 assert.deepEqual(JSON.parse(filled.source_zones_json),zones)
}
for(const zones of [undefined,null,{},['a','a'],['*','a'],[null],['a b'],['-a'],['a;b'],['\ud800']]) {
 const current=structuredClone(pipeEditRow)
 current.bucket_sync_policy.groups[0].pipes[0].dest.zones=zones
 const failed=pipeZonesAction.changedValues({pipe_id:pipeEditValues.pipe_id},pipeEditValues,current)
 assert.ok(failed.pipe_load_error)
 assert.equal(failed.source_zones_json,undefined)
 assert.equal(failed.confirm_pipe_zones,undefined)
 assert.throws(()=>pipeZonesAction.buildBody({...pipeZonesValues,pipe_load_error:failed.pipe_load_error},7,current))
}
assert.ok(pipeZonesAction.changedValues({pipe_id:pipeEditValues.pipe_id},pipeEditValues,{...pipeEditRow,stale:true}).pipe_load_error)
assert.equal(updatePipeAction.method, 'PATCH')
for(const user of ['team$ns$user$extra','$ns$$','team$ns$user$']) {
 assert.equal(updatePipeAction.buildBody({...pipeEditValues,mode:'user',user},7,pipeEditRow).user,user)
 const current=structuredClone(pipeEditRow)
 current.bucket_sync_policy.groups[0].pipes[0].params.user=user
 const filled=updatePipeAction.changedValues({pipe_id:pipeEditValues.pipe_id},pipeEditValues,current)
 assert.equal(filled.pipe_load_error,undefined)
 assert.equal(filled.user,user)
}
assert.equal(updatePipeAction.path, '/rgw/bucket/sync/pipe')
assert.deepEqual(await updatePipeAction.fields.find(f=>f.name==='group_id').optionsLoader(7,pipeEditRow),[{value:group.id,label:group.id}])
const pipePicker=updatePipeAction.fields.find(f=>f.name==='pipe_id')
assert.deepEqual(pipePicker.optionsDependencies,['group_id'])
assert.deepEqual(await pipePicker.optionsLoader(7,pipeEditRow,pipeEditValues),[{value:' 管道 ',label:' 管道 '}])
assert.deepEqual(await pipePicker.optionsLoader(7,pipeEditRow,{}),[])
const loadedPipe=updatePipeAction.changedValues({pipe_id:pipeEditValues.pipe_id},pipeEditValues,pipeEditRow)
assert.equal(loadedPipe.source_bucket,'*')
assert.equal(loadedPipe.dest_bucket,'*')
assert.equal(loadedPipe.mode,'user')
assert.equal(loadedPipe.user,'old')
assert.equal(loadedPipe.priority,undefined)
assert.equal(loadedPipe.confirm_pipe_update,undefined)
assert.equal(loadedPipe.acl_mode,'preserve')
assert.equal(loadedPipe.pipe_load_error,undefined)
assert.deepEqual(updatePipeAction.changedValues({priority:3},pipeEditValues,pipeEditRow),{})
assert.equal(updatePipeAction.changedValues({group_id:'other'},pipeEditValues,pipeEditRow).pipe_id,undefined)
assert.equal(updatePipeAction.changedValues({group_id:'other'},pipeEditValues,pipeEditRow).source_bucket,undefined)
for(const key of ['team/photos:instance','photos','*','team/*:instance']) {
 const current=structuredClone(pipeEditRow)
 current.bucket_sync_policy.groups[0].pipes[0].source.bucket=key
 current.bucket_sync_policy.groups[0].pipes[0].params.mode='system'
 const filled=updatePipeAction.changedValues({pipe_id:pipeEditValues.pipe_id},pipeEditValues,current)
 assert.equal(filled.pipe_load_error,undefined)
 assert.equal(filled.user,undefined)
 const rebuilt=(filled.source_tenant?filled.source_tenant+'/':'')+filled.source_bucket+(filled.source_bucket_id==='*'?'':':'+filled.source_bucket_id)
 assert.equal(rebuilt,key)
}
for(const key of [undefined,null,'','team//photos','photos:','*/photos','photos:*','photos:one:two','photos bad','a\\b']) {
 const current=structuredClone(pipeEditRow)
 current.bucket_sync_policy.groups[0].pipes[0].source.bucket=key
 const failed=updatePipeAction.changedValues({pipe_id:pipeEditValues.pipe_id},pipeEditValues,current)
 assert.ok(failed.pipe_load_error)
 assert.equal(failed.source_bucket,undefined)
 assert.equal(failed.confirm_pipe_update,undefined)
 assert.throws(()=>updatePipeAction.buildBody({...pipeEditValues,pipe_load_error:failed.pipe_load_error},7,current))
}
assert.ok(updatePipeAction.changedValues({pipe_id:pipeEditValues.pipe_id},pipeEditValues,{...pipeEditRow,stale:true}).pipe_load_error)
const duplicatePipeRow=structuredClone(pipeEditRow)
duplicatePipeRow.bucket_sync_policy.groups[0].pipes.push(duplicatePipeRow.bucket_sync_policy.groups[0].pipes[0])
await assert.rejects(()=>pipePicker.optionsLoader(7,duplicatePipeRow,pipeEditValues))
assert.ok(updatePipeAction.changedValues({pipe_id:pipeEditValues.pipe_id},pipeEditValues,duplicatePipeRow).pipe_load_error)
assert.deepEqual(updatePipeAction.buildBody(pipeEditValues, 7, pipeEditRow), { cluster_id: 7, bucket_id: row.natural_key, group_id: group.id, pipe_id: ' 管道 ', expected_group: JSON.stringify(pipeEditGroup), source_bucket: 'photos', dest_bucket: '*', source_tenant: '', dest_tenant: '', source_bucket_id: '*', dest_bucket_id: '*', mode: 'system' })
assert.match(updatePipeAction.confirmation(pipeEditValues, pipeEditRow), /system 模式保留已存储 UID.*保留 Zone 成员.*不自动回滚/)
assert.equal(updatePipeAction.buildBody({ ...pipeEditValues, mode: 'user', user: 'tenant$uid' }, 7, pipeEditRow).user, 'tenant$uid')
for (const priority of [-2147483648,0,2147483647]) {
 assert.equal(updatePipeAction.buildBody({...pipeEditValues,priority},7,pipeEditRow).priority,priority)
 assert.ok(updatePipeAction.confirmation({...pipeEditValues,priority},pipeEditRow).includes(`优先级：${priority}`))
}
for (const priority of [-2147483649,2147483648,0.5,'0',NaN]) assert.throws(()=>updatePipeAction.buildBody({...pipeEditValues,priority},7,pipeEditRow))
for (const priority of [undefined,null,'']) assert.equal(updatePipeAction.buildBody({...pipeEditValues,priority},7,pipeEditRow).priority,undefined)
assert.equal(updatePipeAction.fields.find(field=>field.name==='priority').max,2147483647)
for (const storage_class of ['COLD','归档',' spaced ']) {
 const values={...pipeEditValues,storage_class_mode:'set',storage_class}
 assert.equal(updatePipeAction.buildBody(values,7,pipeEditRow).storage_class,storage_class)
 assert.ok(updatePipeAction.confirmation(values,pipeEditRow).includes(JSON.stringify(storage_class)))
}
assert.equal(updatePipeAction.buildBody({...pipeEditValues,storage_class_mode:'empty',storage_class:'stale'},7,pipeEditRow).storage_class,'')
assert.equal(updatePipeAction.buildBody({...pipeEditValues,storage_class_mode:'preserve',storage_class:'stale'},7,pipeEditRow).storage_class,undefined)
for (const storage_class of ['',null,'-x','x\n','x'.repeat(513),'\ud800']) assert.throws(()=>updatePipeAction.buildBody({...pipeEditValues,storage_class_mode:'set',storage_class},7,pipeEditRow))
assert.throws(()=>updatePipeAction.buildBody({...pipeEditValues,storage_class_mode:'unknown'},7,pipeEditRow))
assert.equal(updatePipeAction.fields.find(field=>field.name==='storage_class').visibleWhen({storage_class_mode:'empty'}),false)
for (const source_prefix of [' photos/ ','目录/','--leading','a=b','x'.repeat(1024)]) {
 const values={...pipeEditValues,prefix_mode:'set',source_prefix}
 assert.equal(updatePipeAction.buildBody(values,7,pipeEditRow).source_prefix,source_prefix)
 assert.ok(updatePipeAction.confirmation(values,pipeEditRow).includes(JSON.stringify(source_prefix)))
}
assert.equal(updatePipeAction.buildBody({...pipeEditValues,prefix_mode:'empty',source_prefix:'stale'},7,pipeEditRow).source_prefix,'')
assert.equal(updatePipeAction.buildBody({...pipeEditValues,prefix_mode:'remove',source_prefix:'stale'},7,pipeEditRow).source_prefix,undefined)
assert.equal(updatePipeAction.buildBody({...pipeEditValues,prefix_mode:'preserve',source_prefix:'stale'},7,pipeEditRow).prefix_mode,undefined)
for (const source_prefix of ['',null,'x\n','x'.repeat(1025),'\ud800']) assert.throws(()=>updatePipeAction.buildBody({...pipeEditValues,prefix_mode:'set',source_prefix},7,pipeEditRow))
assert.throws(()=>updatePipeAction.buildBody({...pipeEditValues,prefix_mode:'unknown'},7,pipeEditRow))
assert.equal(updatePipeAction.fields.find(field=>field.name==='source_prefix').visibleWhen({prefix_mode:'empty'}),false)
const tagValues={...pipeEditValues,tags_mode:'change',tags_remove_json:'[{"key":"k","value":"old"}]',tags_add_json:'[{"key":"k","value":""},{"key":"k","value":"v=more"},{"key":"","value":""}]'}
assert.deepEqual(updatePipeAction.buildBody(tagValues,7,pipeEditRow).tags_add,[{key:'k',value:''},{key:'k',value:'v=more'},{key:'',value:''}])
assert.deepEqual(updatePipeAction.buildBody(tagValues,7,pipeEditRow).tags_remove,[{key:'k',value:'old'}])
assert.ok(updatePipeAction.confirmation(tagValues,pipeEditRow).includes('不修改对象本身的标签'))
assert.equal(updatePipeAction.buildBody({...tagValues,tags_mode:'preserve',tags_add_json:'invalid'},7,pipeEditRow).tags_add,undefined)
for (const tags_add_json of ['null','{}','invalid','[{"key":"k," ,"value":"v"}]','[{"key":"k=","value":"v"}]','[{"key":"k","value":"old"}]','[{"key":"k","value":"v"},{"key":"k","value":"v"}]','[{"key":"k"}]',JSON.stringify([{key:'k',value:'\n'}]),JSON.stringify([{key:'k',value:'x'.repeat(1025)}]),JSON.stringify([{key:'k',value:'\ud800'}])]) assert.throws(()=>updatePipeAction.buildBody({...tagValues,tags_add_json},7,pipeEditRow))
assert.throws(()=>updatePipeAction.buildBody({...pipeEditValues,tags_mode:'change'},7,pipeEditRow))
assert.equal(updatePipeAction.fields.find(field=>field.name==='tags_remove_json').visibleWhen({tags_mode:'preserve'}),false)
for (const dest_owner of ['u','team$u','$ns$u','team$ns$u','team$ns$u$extra']) {
 const values={...pipeEditValues,acl_mode:'set',dest_owner}
 assert.equal(updatePipeAction.buildBody(values,7,pipeEditRow).dest_owner,dest_owner)
 assert.ok(updatePipeAction.confirmation(values,pipeEditRow).includes(JSON.stringify(dest_owner)))
}
assert.equal(updatePipeAction.buildBody({...pipeEditValues,acl_mode:'remove',dest_owner:'stale'},7,pipeEditRow).dest_owner,'')
assert.equal(updatePipeAction.buildBody({...pipeEditValues,acl_mode:'preserve',dest_owner:'stale'},7,pipeEditRow).dest_owner,undefined)
for (const dest_owner of ['',null,'-u','u x','u\n','$u','team$','team$$u','$ns$','x'.repeat(513),'\ud800']) assert.throws(()=>updatePipeAction.buildBody({...pipeEditValues,acl_mode:'set',dest_owner},7,pipeEditRow))
assert.throws(()=>updatePipeAction.buildBody({...pipeEditValues,acl_mode:'unknown'},7,pipeEditRow))
assert.equal(updatePipeAction.fields.find(field=>field.name==='dest_owner').visibleWhen({acl_mode:'remove'}),false)
for (const change of [{ bucket_id:'other' }, { group_id:'other' }, { pipe_id:'missing' }, { confirm_pipe_update:true }, { source_bucket:'' }, { source_bucket:'a/b' }, { mode:'unknown' }, { user:'unexpected' }, { mode:'user' }]) assert.throws(() => updatePipeAction.buildBody({ ...pipeEditValues, ...change }, 7, pipeEditRow))
assert.ok(updatePipeAction.disabledWhen({ ...pipeEditRow, stale: true }))
assert.equal(pipeEditGroup.pipes[0].params.user, 'old')
console.log('bucket pipe selector and identity editing checks passed')
const updateGroup = { ...group, data_flow: { symmetrical: [{ id: ' 流 ', zones: ['Zeta', 'Alpha'] }] } }
const updateRow = { ...row, bucket_sync_policy: { groups: [updateGroup] } }
const updateValues = { bucket_id: row.natural_key, group_id: group.id, flow_id: ' 流 ', zones_json: '["b","c"]', confirm_flow_update: 'acknowledged' }
assert.equal(updateFlowAction.method, 'PATCH')
assert.equal(updateFlowAction.path, '/rgw/bucket/sync/flow')
assert.deepEqual(updateFlowAction.buildBody(updateValues, 7, updateRow), { cluster_id: 7, bucket_id: row.natural_key, group_id: group.id, flow_id: ' 流 ', expected_group: JSON.stringify(updateGroup), zones: ['b', 'c'] })
assert.match(updateFlowAction.confirmation(updateValues, updateRow), /先添加再移除.*非事务.*部分生效.*不自动回滚或重试/)
assert.ok(updateFlowAction.disabledWhen({ ...updateRow, stale: true }))
for (const change of [{ bucket_id:'wrong' }, { group_id:'wrong' }, { flow_id:'wrong' }, { confirm_flow_update:true }, ...['[]','["a","a"]','["*"]','["a;b"]','["a=b"]','["-a"]','["a b"]','[null]','{}','broken'].map(zones_json => ({ zones_json }))]) assert.throws(() => updateFlowAction.buildBody({ ...updateValues, ...change }, 7, updateRow))
assert.throws(() => updateFlowAction.buildBody(updateValues, 7, { ...updateRow, bucket_sync_policy: { groups: [{ ...updateGroup, data_flow: { symmetrical: [updateGroup.data_flow.symmetrical[0], updateGroup.data_flow.symmetrical[0]] } }] } }))
assert.deepEqual(updateGroup.data_flow.symmetrical[0].zones, ['Zeta', 'Alpha'])
console.log('bucket symmetrical flow membership edit checks passed')
const initial = action.initialValues(row)
assert.equal(initial.status, undefined)
assert.equal(initial.group_id, undefined)
assert.equal(initial.confirm_change, undefined)
assert.ok(action.disabledWhen({ ...row, stale: true }))
assert.ok(action.disabledWhen({ ...row, bucket_sync_policy: { groups: [] } }))
for (const status of ['allowed', 'forbidden']) {
  const values = { ...initial, group_id: group.id, status, confirm_change: 'acknowledged' }
  assert.deepEqual(action.buildBody(values, 7, row), { cluster_id: 7, bucket_id: row.natural_key, group_id: group.id, status, expected_status: 'enabled' })
  assert.match(action.confirmation(values, row), /AGJ1Y2tldA.*team-policy.*不会新建数据流或管道/)
  for (const change of [{ bucket_id: 'other' }, { group_id: 'other' }, { status: 'enabled' }, { status: 'unknown' }, { confirm_change: true }]) assert.throws(() => action.buildBody({ ...values, ...change }, 7, row))
}
console.log('bucket sync group form and action binding checks passed')
assert.ok(createAction)
assert.equal(createAction.path, '/rgw/bucket/sync/group')
assert.equal(createAction.method, 'POST')
const emptyRow = { ...row, bucket_sync_policy: { groups: [] } }
assert.equal(createAction.disabledWhen(emptyRow), undefined)
assert.ok(createAction.disabledWhen({ ...row, bucket_sync_policy: null }))
assert.ok(createAction.disabledWhen({ ...row, stale: true }))
const creation = createAction.initialValues(emptyRow)
assert.equal(creation.status, undefined)
assert.equal(creation.confirm_create, undefined)
for (const status of ['enabled', 'allowed', 'forbidden']) {
  const values = { ...creation, group_id: ' 新组 ', status, confirm_create: 'acknowledged' }
  assert.deepEqual(createAction.buildBody(values, 7, row), { cluster_id:7, bucket_id:row.natural_key, group_id:' 新组 ', status })
  assert.match(createAction.confirmation(values, row), /空数据流和空管道.*不建立可工作的复制链路/)
  for (const change of [{ group_id: group.id }, { group_id:'' }, { group_id:'-bad' }, { group_id:'a\nb' }, { group_id:'\ud800' }, { group_id:'中'.repeat(171) }, { bucket_id:'other' }, { status:'unknown' }, { confirm_create:true }]) assert.throws(() => createAction.buildBody({ ...values, ...change }, 7, row))
}
console.log('bucket sync group creation form checks passed')
assert.equal(deleteAction.method, 'DELETE')
assert.equal(deleteAction.path, '/rgw/bucket/sync/group')
assert.ok(deleteAction.disabledWhen(emptyRow))
assert.ok(deleteAction.disabledWhen({ ...row, stale: true }))
const deletion = { ...deleteAction.initialValues(row), group_id: group.id, confirm_delete: 'acknowledged' }
assert.deepEqual(deleteAction.buildBody(deletion, 7, row), { cluster_id: 7, bucket_id: row.natural_key, group_id: group.id, expected_group: JSON.stringify(group) })
assert.match(deleteAction.confirmation(deletion, row), /全部数据流、管道.*forbidden.*不自动回滚/)
for (const change of [{ bucket_id: 'other' }, { group_id: 'other' }, { confirm_delete: true }]) assert.throws(() => deleteAction.buildBody({ ...deletion, ...change }, 7, row))
console.log('bucket sync group deletion form checks passed')
assert.equal(flowAction.path, '/rgw/bucket/sync/flow')
assert.equal(flowAction.method, 'POST')
assert.ok(flowAction.disabledWhen(emptyRow))
assert.ok(flowAction.disabledWhen({ ...row, stale: true }))
const flowBase = { ...flowAction.initialValues(row), group_id: group.id, confirm_flow: 'acknowledged' }
const sym = { ...flowBase, flow_type: 'symmetrical', flow_id: ' 流 ', zones_json: '["b","a"]' }
const dir = { ...flowBase, flow_type: 'directional', source_zone: 'a', dest_zone: 'b' }
assert.deepEqual(flowAction.buildBody(sym, 7, row), { cluster_id: 7, bucket_id: row.natural_key, group_id: group.id, expected_group: JSON.stringify(group), flow_type: 'symmetrical', flow_id: ' 流 ', zones: ['b', 'a'] })
assert.deepEqual(flowAction.buildBody(dir, 7, row), { cluster_id: 7, bucket_id: row.natural_key, group_id: group.id, expected_group: JSON.stringify(group), flow_type: 'directional', source_zone: 'a', dest_zone: 'b' })
for (const values of [sym, dir]) {
  assert.match(flowAction.confirmation(values, row), /不创建管道.*不自动回滚/)
  for (const change of [{ bucket_id: 'other' }, { group_id: 'other' }, { confirm_flow: true }, { flow_type: 'unknown' }]) assert.throws(() => flowAction.buildBody({ ...values, ...change }, 7, row))
}
for (const change of [{ flow_id: '-bad' }, { zones_json: '[]' }, { zones_json: '["a","a"]' }, { zones_json: '["*"]' }, { zones_json: '["a,b"]' }, { zones_json: '[3]' }, { zones_json: 'broken' }, { source_zone: 'a' }]) assert.throws(() => flowAction.buildBody({ ...sym, ...change }, 7, row))
for (const change of [{ source_zone: 'b' }, { dest_zone: '' }, { flow_id: 'f' }, { zones_json: '[]' }]) assert.throws(() => flowAction.buildBody({ ...dir, ...change }, 7, row))
const existingRow = { ...row, bucket_sync_policy: { groups: [{ ...group, data_flow: { symmetrical: [{ id: ' 流 ', zones: ['a'] }], directional: [{ source_zone: 'a', dest_zone: 'b' }] } }] } }
for (const values of [sym, dir]) assert.throws(() => flowAction.buildBody(values, 7, existingRow))
console.log('bucket sync flow creation form and binding checks passed')
assert.equal(deleteFlowAction.path, '/rgw/bucket/sync/flow')
assert.equal(deleteFlowAction.method, 'DELETE')
assert.ok(deleteFlowAction.disabledWhen({ ...existingRow, stale: true }))
const deleteBase = { ...deleteFlowAction.initialValues(existingRow), group_id: group.id, confirm_flow_delete: 'acknowledged' }
const deleteSym = { ...deleteBase, flow_type: 'symmetrical', flow_id: ' 流 ' }
const deleteDir = { ...deleteBase, flow_type: 'directional', source_zone: 'source-id', dest_zone: 'dest-id' }
for (const values of [deleteSym, deleteDir]) {
  const body = deleteFlowAction.buildBody(values, 7, existingRow)
  assert.equal(body.cluster_id, 7)
  assert.equal(body.expected_group, JSON.stringify(existingRow.bucket_sync_policy.groups[0]))
  assert.equal(body.zones, undefined)
  assert.match(deleteFlowAction.confirmation(values, existingRow), /保留组状态、其他流和管道.*不自动回滚/)
  for (const change of [{ bucket_id: 'other' }, { group_id: 'other' }, { confirm_flow_delete: true }, { flow_type: 'unknown' }]) assert.throws(() => deleteFlowAction.buildBody({ ...values, ...change }, 7, existingRow))
  assert.throws(() => deleteFlowAction.buildBody(values, 7, row))
}
assert.match(deleteFlowAction.confirmation(deleteSym, existingRow), /包含全部 Zone/)
assert.equal(deleteFlowAction.buildBody(deleteDir, 7, existingRow).source_zone, 'source-id')
for (const change of [{ flow_id: 'missing' }, { source_zone: 'a' }]) assert.throws(() => deleteFlowAction.buildBody({ ...deleteSym, ...change }, 7, existingRow))
for (const change of [{ source_zone: 'dest-id' }, { dest_zone: '*' }, { source_zone: 'a;b' }, { flow_id: 'f' }]) assert.throws(() => deleteFlowAction.buildBody({ ...deleteDir, ...change }, 7, existingRow))
console.log('bucket sync flow deletion form and binding checks passed')
assert.equal(deletePipeAction.path, '/rgw/bucket/sync/pipe')
assert.equal(deletePipeAction.method, 'DELETE')
const selectedPipe = { id: ' 管道 ', source: { bucket: '*', zones: ['*'] }, dest: { bucket: 'team/photos:marker', zones: ['Zone B'] }, params: { mode: 'user' } }
const pipeGroup = { ...group, pipes: [selectedPipe] }
const pipeRow = { ...row, bucket_sync_policy: { groups: [pipeGroup] } }
const pipeValues = { ...deletePipeAction.initialValues(pipeRow), group_id: group.id, pipe_id: selectedPipe.id, confirm_pipe_delete: 'acknowledged' }
assert.deepEqual(deletePipeAction.buildBody(pipeValues, 7, pipeRow), { cluster_id: 7, bucket_id: row.natural_key, group_id: group.id, pipe_id: selectedPipe.id, expected_group: JSON.stringify(pipeGroup) })
assert.match(deletePipeAction.confirmation(pipeValues, pipeRow), /team\/photos:marker.*保留组状态、数据流及其他管道.*不自动回滚/)
assert.ok(deletePipeAction.disabledWhen({ ...pipeRow, stale: true }))
for (const change of [{ bucket_id: 'other' }, { group_id: 'other' }, { pipe_id: 'missing' }, { confirm_pipe_delete: true }]) assert.throws(() => deletePipeAction.buildBody({ ...pipeValues, ...change }, 7, pipeRow))
for (const pipes of [[], null, [null], [selectedPipe, selectedPipe]]) assert.throws(() => deletePipeAction.buildBody(pipeValues, 7, { ...pipeRow, bucket_sync_policy: { groups: [{ ...group, pipes }] } }))
console.log('bucket sync pipe deletion form and binding checks passed')
assert.equal(createPipeAction.method, 'POST')
assert.equal(createPipeAction.path, '/rgw/bucket/sync/pipe')
assert.ok(createPipeAction.disabledWhen({ ...row, stale: true }))
const pipeCreate = { ...createPipeAction.initialValues(row), group_id: group.id, pipe_id: ' new ', source_zones_json: '["b","a"]', dest_zones_json: '["*"]', source_tenant: 'team', source_bucket: 'photos', source_bucket_id: 'marker', dest_bucket: '*', mode: 'system', confirm_pipe_create: 'acknowledged' }
assert.deepEqual(createPipeAction.buildBody(pipeCreate, 7, row), { cluster_id: 7, bucket_id: row.natural_key, group_id: group.id, pipe_id: ' new ', expected_group: JSON.stringify(group), source_zones: ['b','a'], dest_zones: ['*'], source_tenant: 'team', source_bucket: 'photos', source_bucket_id: 'marker', dest_tenant: '', dest_bucket: '*', dest_bucket_id: '*', mode: 'system' })
assert.match(createPipeAction.confirmation(pipeCreate, row), /空租户不限定租户.*优先级 0.*不自动回滚/)
assert.equal(createPipeAction.buildBody({ ...pipeCreate, mode: 'user', user: 'team$u' }, 7, row).user, 'team$u')
assert.equal(createPipeAction.initialValues(row).mode, undefined)
for (const change of [{ bucket_id: 'other' }, { group_id: 'other' }, { pipe_id: '-bad' }, { source_zones_json: '["*","a"]' }, { source_zones_json: '["a","a"]' }, { dest_zones_json: '["a;b"]' }, { dest_zones_json: '["a=b"]' }, { dest_zones_json: '[]' }, { source_bucket: '' }, { source_bucket: 'team/photos' }, { source_bucket: 'pho*' }, { source_tenant: 'a/b' }, { source_bucket_id: 'a:b' }, { mode: 'unknown' }, { mode: 'user' }, { user: 'u' }, { mode: 'user', user: '$u' }, { confirm_pipe_create: true }]) assert.throws(() => createPipeAction.buildBody({ ...pipeCreate, ...change }, 7, row))
assert.throws(() => createPipeAction.buildBody({ ...pipeCreate, pipe_id: selectedPipe.id }, 7, pipeRow))
console.log('bucket sync pipe creation form and binding checks passed')

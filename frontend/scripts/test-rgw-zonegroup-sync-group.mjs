import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api={}
const pipeFields={}
new Function('exports',ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketSyncGroupForm.ts',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022}}).outputText)(pipeFields)
new Function('exports','require',ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwZonegroupSyncGroup.ts',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022}}).outputText)(api,()=>pipeFields)
const source=ts.createSourceFile('pages.tsx',readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8'),ts.ScriptTarget.Latest,true,ts.ScriptKind.TSX)
let action,createAction,deleteAction,flowAction,flowDeleteAction,flowUpdateAction,pipeAction,pipeDeleteAction,pipeUpdateAction
function visit(node){if(ts.isObjectLiteralExpression(node)){const title=node.properties.find(p=>ts.isPropertyAssignment(p)&&p.name.getText(source)==='title')?.initializer.text;if(['修改 Zonegroup 同步组状态','创建 Zonegroup 同步组','删除 Zonegroup 同步组','创建 Zonegroup 同步流','删除 Zonegroup 同步流','修改 Zonegroup 对称流成员','创建 Zonegroup 同步管道','删除 Zonegroup 同步管道','编辑 Zonegroup 管道选择器与身份'].includes(title)){const code=ts.transpileModule(`const action=${node.getText(source)}`,{compilerOptions:{target:ts.ScriptTarget.ES2022}}).outputText;const found=new Function(...Object.keys(api),`${code};return action`)(...Object.values(api));if(title==='创建 Zonegroup 同步组')createAction=found;else if(title==='删除 Zonegroup 同步组')deleteAction=found;else if(title==='创建 Zonegroup 同步流')flowAction=found;else if(title==='删除 Zonegroup 同步流')flowDeleteAction=found;else if(title==='修改 Zonegroup 对称流成员')flowUpdateAction=found;else if(title==='创建 Zonegroup 同步管道')pipeAction=found;else if(title==='删除 Zonegroup 同步管道')pipeDeleteAction=found;else if(title==='编辑 Zonegroup 管道选择器与身份')pipeUpdateAction=found;else action=found}}ts.forEachChild(node,visit)}visit(source)
const group={id:' g ',status:'allowed',data_flow:{},pipes:[]}
for(const realm of ['','realm']){
 const row={id:'zg',name:'east',realm_id:realm,sync_policy:{groups:[group]}}
 const values={...action.initialValues(row),group_id:group.id,status:'enabled',confirm_change:'acknowledged'}
 assert.deepEqual(action.buildBody(values,7,row),{cluster_id:7,name:'east',zonegroup_id:'zg',realm_id:realm,group_id:group.id,expected_group:JSON.stringify(group),status:'enabled'})
 assert.match(action.confirmation(values,row),realm?/其他待提交变更/:/不提交 Period/)
 for(const change of [{zonegroup_id:'other'},{name:'west'},{realm_id:'wrong'},{group_id:'missing'},{status:'allowed'},{status:'unknown'},{confirm_change:true}])assert.throws(()=>action.buildBody({...values,...change},7,row))
 assert.ok(action.disabledWhen({...row,stale:true}))
 assert.ok(action.disabledWhen({...row,sync_policy:{groups:[group,group]}}))
}
assert.equal(action.path,'/rgw/zonegroup/sync/group')
assert.equal(action.method,'PATCH')
assert.equal(deleteAction.method,'DELETE')
for(const realm of ['','realm']){
 const row={id:'zg',name:'east',realm_id:realm,sync_policy:{groups:[group]}}
 const values={...deleteAction.initialValues(row),group_id:group.id,confirm_delete:'acknowledged'}
 assert.deepEqual(deleteAction.buildBody(values,7,row),{cluster_id:7,name:'east',zonegroup_id:'zg',realm_id:realm,group_id:group.id,expected_group:JSON.stringify(group)})
 assert.match(deleteAction.confirmation(values,row),/全部数据流和管道.*forbidden.*不自动回滚或重试/)
 for(const change of [{zonegroup_id:'other'},{group_id:'missing'},{realm_id:'wrong'},{confirm_delete:true}])assert.throws(()=>deleteAction.buildBody({...values,...change},7,row))
 assert.ok(deleteAction.disabledWhen({...row,sync_policy:{groups:[]}}))
 assert.ok(deleteAction.disabledWhen({...row,stale:true}))
}
assert.equal(createAction.method,'POST')
for(const realm of ['','realm']){
 const row={id:'zg',name:'east',realm_id:realm,sync_policy:{groups:[]}}
 assert.equal(createAction.disabledWhen(row),undefined)
 assert.ok(action.disabledWhen(row))
 const values={...createAction.initialValues(row),group_id:' 新组 ',status:'allowed',confirm_create:'acknowledged'}
 assert.deepEqual(createAction.buildBody(values,7,row),{cluster_id:7,name:'east',zonegroup_id:'zg',realm_id:realm,group_id:' 新组 ',status:'allowed',expected_policy:'{"groups":[]}'})
 assert.match(createAction.confirmation(values,row),/空同步组.*不代表建立复制链路.*非事务/)
 for(const change of [{group_id:''},{group_id:'-bad'},{group_id:'a\nb'},{group_id:'\ud800'},{status:undefined},{confirm_create:true},{realm_id:'wrong'}])assert.throws(()=>createAction.buildBody({...values,...change},7,row))
 assert.throws(()=>createAction.buildBody({...values,group_id:group.id},7,{...row,sync_policy:{groups:[group]}}))
 assert.ok(createAction.disabledWhen({...row,sync_policy:null}))
}
console.log('zonegroup sync status and publication form checks passed')
assert.equal(flowAction.path,'/rgw/zonegroup/sync/flow')
assert.equal(flowAction.method,'POST')
for(const realm of ['','realm']){
 const row={id:'zg',name:'east',realm_id:realm,zones:[{id:'a',name:'west'},{id:'b',name:'east'}],sync_policy:{groups:[group]}}
 const base={...flowAction.initialValues(row),group_id:group.id,confirm_flow:'acknowledged'}
 for(const fields of [{flow_type:'symmetrical',flow_id:'f',zones:'b,a'},{flow_type:'directional',source_zone:'a',dest_zone:'b'}]){
  const values={...base,...fields}
  const body=flowAction.buildBody(values,7,row)
  assert.equal(body.cluster_id,7);assert.equal(body.expected_group,JSON.stringify(group));assert.equal(body.realm_id,realm)
  if(fields.zones)assert.deepEqual(body.zones,['b','a']);else assert.equal(body.dest_zone,'b')
  assert.match(flowAction.confirmation(values,row),/非事务.*不自动回滚或重试/)
  for(const change of [{confirm_flow:true},{realm_id:'wrong'},{group_id:'missing'},{flow_type:'unknown'}])assert.throws(()=>flowAction.buildBody({...values,...change},7,row))
  assert.throws(()=>flowAction.buildBody(values,7,{...row,zones:[{id:'a',name:'x'},{id:'a',name:'y'}]}))
 }
 for(const fields of [{flow_type:'symmetrical',flow_id:'f',zones:'a,a'},{flow_type:'symmetrical',flow_id:'f',zones:'missing'},{flow_type:'directional',source_zone:'a',dest_zone:'a'},{flow_type:'directional',flow_id:'f',source_zone:'a',dest_zone:'b'}])assert.throws(()=>flowAction.buildBody({...base,...fields},7,row))
}
console.log('zonegroup sync flow creation form and route checks passed')
assert.equal(flowDeleteAction.path,'/rgw/zonegroup/sync/flow')
assert.equal(flowDeleteAction.method,'DELETE')
for(const realm of ['','realm']){
 const g={...group,data_flow:{symmetrical:[{id:'f',zones:['orphan']}],directional:[{source_zone:'orphan',dest_zone:'z'}]}}
 const row={id:'zg',name:'east',realm_id:realm,zones:[],sync_policy:{groups:[g]}}
 const base={...flowDeleteAction.initialValues(row),group_id:g.id,confirm_flow_delete:'acknowledged'}
 for(const fields of [{flow_type:'symmetrical',flow_id:'f'},{flow_type:'directional',source_zone:'orphan',dest_zone:'z'}]){
  const values={...base,...fields}
  const body=flowDeleteAction.buildBody(values,7,row)
  assert.deepEqual(body,{cluster_id:7,name:'east',zonegroup_id:'zg',realm_id:realm,group_id:g.id,expected_group:JSON.stringify(g),...fields})
  assert.match(flowDeleteAction.confirmation(values,row),/保留组状态和管道.*不删除对象副本.*非事务.*不自动回滚或重试/)
  for(const change of [{confirm_flow_delete:true},{realm_id:'wrong'},{group_id:'missing'},{zones:'z'},{flow_type:'unknown'}])assert.throws(()=>flowDeleteAction.buildBody({...values,...change},7,row))
  const duplicate={...g,data_flow:{...g.data_flow,[fields.flow_type]:[...g.data_flow[fields.flow_type],...g.data_flow[fields.flow_type]]}}
  assert.throws(()=>flowDeleteAction.buildBody(values,7,{...row,sync_policy:{groups:[duplicate]}}))
 }
 for(const fields of [{flow_type:'symmetrical',flow_id:'missing'},{flow_type:'symmetrical',flow_id:'f',source_zone:'z'},{flow_type:'directional',source_zone:'z',dest_zone:'orphan'},{flow_type:'directional',source_zone:'orphan',dest_zone:'z',flow_id:'f'}])assert.throws(()=>flowDeleteAction.buildBody({...base,...fields},7,row))
}
console.log('zonegroup sync flow deletion form and route checks passed')
assert.equal(flowUpdateAction.method,'PATCH')
assert.equal(flowUpdateAction.path,'/rgw/zonegroup/sync/flow')
for(const realm of ['','realm']){
 const g={...group,data_flow:{symmetrical:[{id:'f',zones:['orphan']}]}}
 const row={id:'zg',name:'east',realm_id:realm,zones:[{id:'a',name:'west'},{id:'b',name:'east'}],sync_policy:{groups:[g]}}
 const values={...flowUpdateAction.initialValues(row),group_id:g.id,flow_id:'f',zones:'b,a',confirm_flow_update:'acknowledged'}
 assert.deepEqual(flowUpdateAction.buildBody(values,7,row),{cluster_id:7,name:'east',zonegroup_id:'zg',realm_id:realm,group_id:g.id,flow_id:'f',zones:['b','a'],expected_group:JSON.stringify(g)})
 assert.match(flowUpdateAction.confirmation(values,row),/先添加后移除.*中间成员集合可能扩大.*非事务.*不自动回滚或重试/)
 for(const change of [{confirm_flow_update:true},{realm_id:'wrong'},{group_id:'missing'},{flow_id:'missing'},{zones:''},{zones:'a,a'},{zones:'unknown'},{zones:'*'}])assert.throws(()=>flowUpdateAction.buildBody({...values,...change},7,row))
 for(const zones of [[],['a','a'],['bad id']])assert.throws(()=>flowUpdateAction.buildBody(values,7,{...row,sync_policy:{groups:[{...g,data_flow:{symmetrical:[{id:'f',zones}]}}]}}))
 const current={...g,data_flow:{symmetrical:[{id:'f',zones:['a','b']}]}}
 assert.throws(()=>flowUpdateAction.buildBody(values,7,{...row,sync_policy:{groups:[current]}}))
 assert.throws(()=>flowUpdateAction.buildBody(values,7,{...row,stale:true}))
}
console.log('zonegroup symmetrical flow membership form checks passed')
assert.equal(pipeAction.path,'/rgw/zonegroup/sync/pipe')
assert.equal(pipeAction.method,'POST')
for(const realm of ['','realm']){
 const row={id:'zg',name:'east',realm_id:realm,zones:[{id:'a',name:'west'},{id:'b',name:'east'}],sync_policy:{groups:[group]}}
 const values={...pipeAction.initialValues(row),group_id:group.id,pipe_id:'p',source_zones_json:'["b","a"]',dest_zones_json:'["*"]',source_bucket:'photos',source_tenant:'team',dest_bucket:'*',mode:'system',confirm_pipe_create:'acknowledged'}
 const body=pipeAction.buildBody(values,7,row)
 assert.deepEqual(body,{cluster_id:7,name:'east',zonegroup_id:'zg',realm_id:realm,group_id:group.id,pipe_id:'p',source_zones:['b','a'],dest_zones:['*'],source_bucket:'photos',source_tenant:'team',source_bucket_id:'*',dest_tenant:'',dest_bucket:'*',dest_bucket_id:'*',mode:'system',expected_group:JSON.stringify(group)})
 assert.match(pipeAction.confirmation(values,row),/空租户不限租户.*优先级 0.*非事务.*不自动回滚或重试/)
 assert.equal(pipeAction.buildBody({...values,mode:'user',user:'team$u'},7,row).user,'team$u')
 for(const change of [{realm_id:'wrong'},{pipe_id:''},{source_zones_json:'["missing"]'},{dest_zones_json:'["*","a"]'},{mode:'user'},{user:'u'},{source_bucket:'a/b'},{confirm_pipe_create:true}])assert.throws(()=>pipeAction.buildBody({...values,...change},7,row))
 assert.throws(()=>pipeAction.buildBody(values,7,{...row,sync_policy:{groups:[{...group,pipes:[{id:'p'}]}]}}))
 assert.throws(()=>pipeAction.buildBody(values,7,{...row,zones:[]}))
 assert.deepEqual(pipeAction.buildBody({...values,source_zones_json:'["*"]'},7,{...row,zones:[]}).source_zones,['*'])
}
console.log('zonegroup pipe creation form and binding checks passed')
assert.equal(pipeDeleteAction.path,'/rgw/zonegroup/sync/pipe')
assert.equal(pipeDeleteAction.method,'DELETE')
for(const realm of ['','realm']){
 const pipe={id:'p',source:{zones:['orphan'],bucket:'team/photos'},params:{mode:'user',user:'team$u'}}
 const g={...group,pipes:[pipe]}
 const row={id:'zg',name:'east',realm_id:realm,zones:[],sync_policy:{groups:[g]}}
 const values={...pipeDeleteAction.initialValues(row),group_id:g.id,pipe_id:'p',confirm_pipe_delete:'acknowledged'}
 assert.deepEqual(pipeDeleteAction.buildBody(values,7,row),{cluster_id:7,name:'east',zonegroup_id:'zg',realm_id:realm,group_id:g.id,pipe_id:'p',expected_group:JSON.stringify(g)})
 assert.match(pipeDeleteAction.confirmation(values,row),/整个管道.*高级参数将移除.*不删除已有对象副本.*非事务.*不自动回滚或重试/)
 for(const change of [{pipe_id:'missing'},{pipe_id:'-bad'},{group_id:'missing'},{realm_id:'wrong'},{zonegroup_id:'wrong'},{confirm_pipe_delete:true}])assert.throws(()=>pipeDeleteAction.buildBody({...values,...change},7,row))
 for(const pipes of [[],[pipe,pipe],[null],[{id:1}]]){
  assert.throws(()=>pipeDeleteAction.buildBody(values,7,{...row,sync_policy:{groups:[{...g,pipes}]}}))
 }
 assert.throws(()=>pipeDeleteAction.buildBody(values,7,{...row,stale:true}))
 assert.equal(pipeDeleteAction.buildBody({...values,source_zones:['*']},7,row).source_zones,undefined)
}
console.log('zonegroup pipe deletion form and binding checks passed')
assert.equal(pipeUpdateAction.path,'/rgw/zonegroup/sync/pipe')
assert.equal(pipeUpdateAction.method,'PATCH')
for(const realm of ['','realm']){
 const pipe={id:'p',source:{zones:['orphan'],bucket:'team/photos'},params:{mode:'user',user:'team$u'}}
 const g={...group,pipes:[pipe]}
 const row={id:'zg',name:'east',realm_id:realm,zones:[],sync_policy:{groups:[g]}}
 const values={...pipeUpdateAction.initialValues(row),group_id:g.id,pipe_id:'p',source_bucket:'*',dest_bucket:'archive',mode:'system',confirm_pipe_update:'acknowledged'}
 assert.deepEqual(pipeUpdateAction.buildBody(values,7,row),{cluster_id:7,name:'east',zonegroup_id:'zg',realm_id:realm,group_id:g.id,pipe_id:'p',source_tenant:'',source_bucket:'*',source_bucket_id:'*',dest_tenant:'',dest_bucket:'archive',dest_bucket_id:'*',mode:'system',expected_group:JSON.stringify(g)})
 assert.equal(pipeUpdateAction.buildBody({...values,mode:'user',user:'new$u'},7,row).user,'new$u')
 assert.match(pipeUpdateAction.confirmation(values,row),/保留已存储 UID.*保留 Zone 成员、过滤器.*非事务.*不自动回滚或重试/)
 for(const change of [{pipe_id:'missing'},{group_id:'missing'},{realm_id:'wrong'},{confirm_pipe_update:true},{mode:'user'},{user:'u'},{source_bucket:'a/b'}])assert.throws(()=>pipeUpdateAction.buildBody({...values,...change},7,row))
 assert.equal(pipeUpdateAction.buildBody({...values,source_zones_json:'["*"]'},7,row).source_zones,undefined)
 assert.equal(row.sync_policy.groups[0].pipes.length,1)
}
console.log('zonegroup pipe selector and identity form checks passed')

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api={}
new Function('exports',ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwZonegroupSyncGroup.ts',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022}}).outputText)(api)
const source=ts.createSourceFile('pages.tsx',readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8'),ts.ScriptTarget.Latest,true,ts.ScriptKind.TSX)
let action,createAction,deleteAction,flowAction
function visit(node){if(ts.isObjectLiteralExpression(node)){const title=node.properties.find(p=>ts.isPropertyAssignment(p)&&p.name.getText(source)==='title')?.initializer.text;if(['修改 Zonegroup 同步组状态','创建 Zonegroup 同步组','删除 Zonegroup 同步组','创建 Zonegroup 同步流'].includes(title)){const code=ts.transpileModule(`const action=${node.getText(source)}`,{compilerOptions:{target:ts.ScriptTarget.ES2022}}).outputText;const found=new Function(...Object.keys(api),`${code};return action`)(...Object.values(api));if(title==='创建 Zonegroup 同步组')createAction=found;else if(title==='删除 Zonegroup 同步组')deleteAction=found;else if(title==='创建 Zonegroup 同步流')flowAction=found;else action=found}}ts.forEachChild(node,visit)}visit(source)
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

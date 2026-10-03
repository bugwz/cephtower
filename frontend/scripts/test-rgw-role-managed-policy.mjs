import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const api={}
new Function('exports',ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwRoleManagedPolicy.ts',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022}}).outputText)(api)
const text=readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8')
const source=ts.createSourceFile('pages.tsx',text,ts.ScriptTarget.Latest,true,ts.ScriptKind.TSX)
let action
function visit(node) {
  if (ts.isObjectLiteralExpression(node)&&node.properties.some(p=>ts.isPropertyAssignment(p)&&p.name.getText(source)==='path'&&ts.isStringLiteral(p.initializer)&&p.initializer.text==='/rgw/role/managed/policy')) {
    const code=ts.transpileModule(`const action=${node.getText(source)}`,{compilerOptions:{target:ts.ScriptTarget.ES2022}}).outputText
    action=new Function(...Object.keys(api),`${code};return action`)(...Object.values(api))
  }
  ts.forEachChild(node,visit)
}
visit(source)
assert.equal(action.method,'PATCH')
const row={AccountId:'RGW12345678901234567',RoleName:'role',RoleId:'id',Arn:'arn:aws:iam::RGW12345678901234567:role/path/role'}
const policy='arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess'
for (const mode of ['attach','detach']) {
  const current={...row,...(mode==='detach'?{ManagedPermissionPolicies:[policy]}:{})}
  const initial=action.initialValues(current)
  const values={...initial,mode,policy_arn:policy,owner_uid:'tenant$user',confirm_managed_policy:'acknowledged'}
  assert.equal(action.disabledWhen(current),undefined)
  const body=action.buildBody(values,42,current)
  assert.deepEqual(body,{cluster_id:42,account_id:row.AccountId,name:'role',owner_uid:'tenant$user',mode,policy_arn:policy,expected_role_id:'id',expected_role_arn:row.Arn,expected_policies:mode==='attach'?[]:[policy]})
  const warning=action.confirmation(values,current)
  for (const text of ['Account','HTTPS','原子锁','回滚','临时会话','内联策略','永久']) assert.ok(warning.includes(text))
  for (const change of [{account_id:'other'},{name:'other'},{owner_uid:''},{confirm_managed_policy:undefined},{mode:'unknown'},{policy_arn:'bad'},{policy_arn:policy+' '},{policy_arn:policy+'\n'}]) assert.throws(()=>action.buildBody({...values,...change},42,current))
  for (const change of [{AccountId:''},{RoleId:undefined},{Arn:row.Arn.replace(row.AccountId,'other')},{stale:true},{ManagedPermissionPolicies:null},{ManagedPermissionPolicies:[policy,policy]}]) {
    assert.ok(action.disabledWhen({...current,...change}))
    assert.throws(()=>action.buildBody(values,42,{...current,...change}))
  }
  assert.throws(()=>action.buildBody({...values,mode:mode==='attach'?'detach':'attach'},42,current))
}
for (const key of ['name','account_id']) assert.equal(action.fields.find(field=>field.name===key).readOnly,true)
const picker=action.fields.find(field=>field.name==='policy_arn')
assert.equal(picker.type,'select')
assert.deepEqual(picker.optionsDependencies,['mode','policy_source'])
const choices=await picker.optionsLoader(42,row,{mode:'attach'})
assert.equal(choices.length,6)
assert.equal(new Set(choices.map(item=>item.value)).size,6)
assert.ok(choices.every(item=>item.value===`arn:aws:iam::aws:policy/${item.label}`))
assert.equal((await picker.optionsLoader(42,{...row,ManagedPermissionPolicies:[policy]},{mode:'attach'})).length,5)
const custom='arn:aws:iam::aws:policy/FuturePolicy'
assert.deepEqual(await picker.optionsLoader(42,{...row,ManagedPermissionPolicies:[custom]},{mode:'detach'}),[{label:custom,value:custom}])
assert.deepEqual(await picker.optionsLoader(42,row,{}),[])
const values={...action.initialValues(row),mode:'attach',owner_uid:'tenant$user',confirm_managed_policy:'acknowledged',policy_arn:policy,custom_policy_arn:custom}
assert.equal(action.buildBody(values,42,row).policy_arn,policy)
assert.equal(action.buildBody({...values,policy_source:'custom'},42,row).policy_arn,custom)
assert.throws(()=>action.buildBody({...values,policy_arn:custom},42,row))
assert.throws(()=>action.buildBody({...values,policy_source:'unknown'},42,row))
assert.throws(()=>action.buildBody({...values,policy_source:'custom',custom_policy_arn:''},42,row))
const detached=action.buildBody({...values,mode:'detach',policy_source:'custom'},42,{...row,ManagedPermissionPolicies:[policy]})
assert.equal(detached.policy_arn,policy)
for (const mode of ['attach','detach',undefined]) for (const policy_source of ['reference','custom',undefined]) {
  const state={mode,policy_source}
  assert.equal(picker.visibleWhen(state),mode==='detach'||(mode==='attach'&&policy_source==='reference'))
  assert.equal(action.fields.find(field=>field.name==='custom_policy_arn').visibleWhen(state),mode==='attach'&&policy_source==='custom')
  assert.equal(action.fields.find(field=>field.name==='policy_source').visibleWhen(state),mode==='attach')
}
console.log('Account role managed policy identity, snapshot and action tests passed')

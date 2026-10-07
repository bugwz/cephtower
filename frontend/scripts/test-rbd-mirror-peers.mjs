import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const ui = {}, jsx = (type,props)=>({type,props})
const source=readFileSync(new URL('../src/pages/block/RbdMirrorPeers.tsx',import.meta.url),'utf8')
new Function('exports','require',ts.transpileModule(source,{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.ReactJSX}}).outputText)(ui,name=>name==='antd'?{Alert:'Alert',Space:'Space',Table:'Table'}:{jsx,jsxs:jsx})
const uuid='aaaaaaaa-0000-0000-0000-000000000001', mirror='bbbbbbbb-0000-0000-0000-000000000002'
const peer={uuid,mirror_uuid:mirror,site_name:'<remote>',client_name:'client.mirror',direction:'tx-only',key:'DO-NOT-DISPLAY'}
const row={peers:[peer]}
assert.deepEqual(ui.rbdMirrorPeerOptions(row),[{value:uuid,label:`<remote> · ${uuid}`}])
assert.equal(ui.rbdMirrorPeerIdentity({uuid},row),uuid)
for(const value of [mirror,'',undefined,uuid+'\n']) assert.throws(()=>ui.rbdMirrorPeerIdentity({uuid:value},row))
for(const peers of [undefined,null,{},[],[null],[{}],[{uuid:'invalid'}],[{uuid:uuid+'\n'}],[peer,peer],[peer,{...peer,uuid:uuid.toUpperCase()}]]) assert.deepEqual(ui.rbdMirrorPeerOptions({peers}),[])
const table=ui.RbdMirrorPeers({value:[peer]}).props.children.find(node=>node.type==='Table')
assert.equal(table.props.columns.length,5)
assert.deepEqual(table.props.dataSource,[{index:0,uuid,site_name:'<remote>',mirror_uuid:mirror,client_name:'client.mirror',direction:'tx-only'}])
assert.ok(!JSON.stringify(table).includes('DO-NOT-DISPLAY'))
assert.equal(ui.RbdMirrorPeers({value:[]}).props.type,'info')
assert.equal(ui.RbdMirrorPeers({value:null}).props.type,'warning')
const pages=readFileSync(new URL('../src/pages/block/pages.tsx',import.meta.url),'utf8')
assert.ok(pages.includes('<RbdMirrorPeers value={value} />'))
assert.equal((pages.match(/uuid:rbdMirrorPeerIdentity\(values,row\)/g)??[]).length,2)
assert.equal((pages.match(/optionsLoader:async\(_clusterId,row\)=>rbdMirrorPeerOptions\(row\)/g)??[]).length,2)
const ast=ts.createSourceFile('pages.tsx',pages,ts.ScriptTarget.Latest,true,ts.ScriptKind.TSX)
const actions=[]
function visit(node){
  if(ts.isObjectLiteralExpression(node)&&node.properties.some(prop=>ts.isPropertyAssignment(prop)&&prop.name.getText(ast)==='title'&&ts.isStringLiteral(prop.initializer)&&['编辑远端站点','删除远端站点'].includes(prop.initializer.text))) actions.push(node)
  ts.forEachChild(node,visit)
}
visit(ast)
assert.equal(actions.length,2)
for(const node of actions){
  const exports={}
  new Function('exports','rbdMirrorPeerOptions','rbdMirrorPeerIdentity',ts.transpileModule(`export const action=${node.getText(ast)}`,{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022}}).outputText)(exports,ui.rbdMirrorPeerOptions,ui.rbdMirrorPeerIdentity)
  const action=exports.action, scoped={...row,pool:'pool-a'}
  assert.equal(action.disabledWhen(scoped),undefined)
  assert.ok(action.disabledWhen({peers:[]}))
  assert.deepEqual(await action.fields[0].optionsLoader(7,scoped),ui.rbdMirrorPeerOptions(scoped))
  const body=action.buildBody({uuid,field:'direction',value:'rx-only'},7,scoped)
  assert.equal(body.uuid,uuid);assert.equal(body.pool,'pool-a');assert.equal(body.cluster_id,7)
  assert.throws(()=>action.buildBody({uuid:mirror},7,scoped))
  assert.throws(()=>action.buildBody({uuid},7,{pool:'pool-b',peers:[]}))
  if(action.confirmation) assert.ok(action.confirmation({uuid},scoped).includes(uuid))
}
console.log('RBD peer inventory keeps peer UUID distinct and rejects ambiguous selections')

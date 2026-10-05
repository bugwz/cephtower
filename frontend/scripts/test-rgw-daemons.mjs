import assert from 'node:assert/strict'
import './test-rgw-daemon-status.mjs'
import './test-rgw-daemon-perf.mjs'
import './test-rgw-daemon-history.mjs'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const code=ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwDaemonsPage.tsx',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.ReactJSX}}).outputText
for(const file of ['HostDetailPage','ServiceDaemons','ServicePage']){
 const source=readFileSync(new URL(`../src/pages/cluster/${file}.tsx`,import.meta.url),'utf8')
 const match=source.match(/\/\^\(mon\|mgr\|mds\|osd\)[^\n]*?\$\//)
 assert.ok(match,`${file} must constrain native tell performance targets`)
 const pattern=new RegExp(match[0].slice(1,-1))
 for(const name of ['rgw.a','rgw.123','rbd-mirror.a','osd.','osd.*'])assert.equal(pattern.test(name),false)
 for(const name of ['osd.1','mon.a','mgr.a','mds.fs.a'])assert.equal(pattern.test(name),true)
}
const ui={},states=[],refs=[],calls=[];let si=0,ri=0,deps,cleanup,pending
const jsx=(type,props)=>({type,props}),react={useState:v=>{const i=si++;if(!(i in states))states[i]=v;return[states[i],v=>states[i]=v]},useRef:v=>refs[ri++]??(refs[ri-1]={current:v}),useEffect:(fn,next)=>{if(JSON.stringify(next)!==JSON.stringify(deps)){deps=next;pending=()=>{cleanup?.();cleanup=fn()}}}}
new Function('exports','require',code)(ui,name=>name==='react'?react:name==='antd'?Object.fromEntries(['Alert','Button','Card','Input','Space','Table'].map(n=>[n,n])):name.includes('api/client')?{jsonInit:(method,body,opts)=>({method,body,...opts}),request:(path,init)=>new Promise((resolve,reject)=>calls.push({path,...init,resolve,reject}))}:name.includes('ClusterContext')?{useClusterContext:()=>({selectedClusterId:7})}:{jsx,jsxs:jsx})
const row={service_map_id:'1',id:'rgw.a',hostname:'host',version:'ceph version',realm_name:'r',realm_id:'realm-id',zonegroup_name:'g',zonegroup_id:'gid',zone_name:'z',zone_id:'zone-id',metadata:{os:'Linux',num_handles:'9007199254740993'},listeners:[{frontend:'frontend_config#0',tls:true,port:443}],listeners_complete:true}
const data={items:[row,{...row,service_map_id:'2'}],source:'service_map',observed_at:'now'}
assert.equal(ui.daemonRegistrationData(data).items.length,2)
assert.deepEqual(ui.daemonRegistrationData({...data,items:[{...row,metadata:{...row.metadata,password:'private'}}]}).items[0].metadata,row.metadata)
for(const metadata of [null,[],{os:1}])assert.throws(()=>ui.daemonRegistrationData({...data,items:[{...row,metadata}]}))
assert.deepEqual(ui.daemonRegistrationData({...data,items:[{...row,password:'private'}]}).items,[row])
for(const listeners of [null,[{frontend:'other',tls:true,port:443}],[{frontend:'frontend_config#0',tls:'true',port:443}],[{frontend:'frontend_config#0',tls:true,port:0}]])assert.throws(()=>ui.daemonRegistrationData({...data,items:[{...row,listeners}]}))
for(const bad of [null,{}, {...data,source:'other'},{...data,items:[row,row]},{...data,items:[{...row,id:''}]},{...data,items:[{...row,hostname:null}]}])assert.throws(()=>ui.daemonRegistrationData(bad))
function nodes(n){return Array.isArray(n)?n.flatMap(nodes):n&&typeof n==='object'?[n,...nodes(n.props?.children)]:[]}
function render(id=7){si=ri=0;const tree=nodes(ui.RgwDaemonsView({clusterId:id}));const effect=pending;pending=undefined;effect?.();return tree}
const tick=()=>new Promise(r=>setTimeout(r,0))
render();let tree=render();const old=tree.find(n=>n.type==='Button');old.props.onClick()
assert.equal(calls[0].path,'/rgw/daemons');assert.deepEqual(calls[0].body,{cluster_id:7});assert.equal(calls[0].cache,'no-store')
calls[0].resolve(data);await tick();tree=render();assert.equal(tree.find(n=>n.type==='Table').props.dataSource.length,2)
for(const field of ['realm_id','zone_id']){
 const column=tree.find(n=>n.type==='Table').props.columns.find(c=>c.dataIndex===field)
 assert.ok(column);assert.equal(column.render(row[field]),row[field]);assert.equal(column.render(''),'—')
 tree.find(n=>n.type==='Input').props.onChange({target:{value:row[field]}})
 assert.equal(render().find(n=>n.type==='Table').props.dataSource.length,2)
 assert.throws(()=>ui.daemonRegistrationData({...data,items:[{...row,[field]:null}]}))
}
const portColumn=tree.find(n=>n.type==='Table').props.columns.at(-1)
const detail=tree.find(n=>n.type==='Table').props.expandable.expandedRowRender(row)
assert.equal(detail.props.clusterId,7);assert.equal(detail.props.serviceMapId,'1');assert.equal(detail.props.isCurrent(),true)
assert.deepEqual(detail.props.metadata,row.metadata)
assert.ok(JSON.stringify(portColumn.render(null,row)).includes('HTTPS 443'))
assert.ok(JSON.stringify(portColumn.render(null,{...row,listeners:[],listeners_complete:false})).includes('无法解析'))
tree.find(n=>n.type==='Input').props.onChange({target:{value:'missing'}});assert.equal(render().find(n=>n.type==='Table').props.dataSource.length,0)
old.props.onClick();render(8);assert.equal(calls[1].signal.aborted,true);old.props.onClick();assert.equal(calls.length,2);calls[1].resolve(data);await tick();assert.equal(render(8).some(n=>n.type==='Table'),false)
assert.equal(detail.props.isCurrent(),false)
render(8).find(n=>n.type==='Button').props.onClick();calls[2].reject(new Error('private-error'));await tick();assert.ok(render(8).some(n=>n.type==='Alert'&&n.props.type==='error'))
const final=render(8).find(n=>n.type==='Button');final.props.onClick();cleanup();final.props.onClick();assert.equal(calls.length,4);assert.equal(calls[3].signal.aborted,true)
calls[3].resolve(data);await tick();assert.equal(states[1].data,undefined)
assert.match(readFileSync(new URL('../src/navigation.ts',import.meta.url),'utf8'),/key: 'rgwDaemons'.*permission: 'storage'/)
console.log('RGW daemon registrations validate identity, projection, filtering and request isolation')

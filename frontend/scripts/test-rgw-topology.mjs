import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api={},calls=[]
let response
const client={jsonInit:(method,body,options)=>({method,body,...options}),request:async(path,init)=>{calls.push({path,...init});return response(path)}}
const compile=file=>ts.transpileModule(readFileSync(new URL(file,import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.React}}).outputText
new Function('exports','require',compile('../src/pages/object/rgwTopologyData.ts'))(api,()=>client)
const realm={id:'r',name:'realm',is_default:true}
const zone={id:'z',name:'zone',realm_id:'r',system_key:{secret_key:'must-not-display'}}
const member={id:'z',name:'zone',tier_type:'archive',endpoints:['https://example']}
const group={id:'g',name:'group',realm_id:'r',master_zone:'z',zones:[member,{id:'remote',name:'remote'}]}
const tree=api.buildRgwTopology([realm],[group],[zone,{id:'orphan',name:'orphan',realm_id:''}])
assert.equal(tree.length,2)
assert.equal(tree[0].children[0].children[0].details.role,'主 Zone')
assert.match(tree[0].children[0].children[1].details.local_detail,/不能据此认定远端/)
assert.match(tree[1].details.relationship,/未列出所属/)
assert.ok(!JSON.stringify(tree).includes('must-not-display'))
assert.match(api.buildRgwTopology([],[{...group,realm_id:''}],[])[0].details.relationship,/未关联/)
assert.match(api.buildRgwTopology([],[group],[])[0].details.relationship,/未找到/)
assert.match(api.buildRgwTopology([realm],[{...group,master_zone:'missing'}],[])[0].children[0].details.warning,/主 Zone/)
assert.match(api.buildRgwTopology([realm],[group],[{...zone,name:'changed'}])[0].children[0].children[0].details.warning,/不一致/)
for(const rows of [[realm,realm],[{id:'',name:'x'}],[null],[{id:'x'}]])assert.throws(()=>api.buildRgwTopology(rows,[],[]))
assert.throws(()=>api.buildRgwTopology([],[{...group,zones:undefined}],[]))
assert.throws(()=>api.buildRgwTopology([],[{...group,zones:[member,member]}],[]))
assert.deepEqual(api.buildRgwTopology([],[],[]),[])
const envelope=(items,next=null,stale=false)=>({items:items.map(data=>({data,stale:false,observed_at:'2026-10-05T00:00:00Z'})),pagination:{next_cursor:next},meta:{stale}})
response=path=>path.startsWith('/rgw/realms')?envelope([realm]):path.startsWith('/rgw/zonegroups')?envelope([group]):path.includes('cursor=next')?envelope([{id:'orphan',name:'orphan'}],null,true):envelope([zone],'next')
const result=await api.readRgwTopology(7)
assert.equal(result.nodes.length,2);assert.equal(result.stale,true)
assert.equal(calls.length,4)
assert.ok(calls.every(call=>call.method==='GET'&&call.body.cluster_id===7&&call.cache==='no-store'))
assert.ok(calls.some(call=>call.path.includes('cursor=next')))
for(const bad of [null,{},envelope([null]),{...envelope([]),meta:undefined},{...envelope([]),pagination:{next_cursor:5}}]){
 response=()=>bad;await assert.rejects(api.readRgwTopology(7))
}
response=()=>envelope([],'repeat');await assert.rejects(api.readRgwTopology(7),/pagination/)
response=()=>{throw new Error('capability unavailable')};await assert.rejects(api.readRgwTopology(7))

const refs=[],states=[],effects=[],pending=[]
let cursor=0,resolve,reject,count=0
const hooks={useRef(v){const i=cursor++;return refs[i]??={current:v}},useState(v){const i=cursor++;if(!(i in states))states[i]=v;return[states[i],v=>{states[i]=v}]},useEffect(fn,deps){const i=cursor++;if(!effects[i]||deps.some((v,n)=>v!==effects[i].deps[n]))pending.push(()=>{effects[i]?.cleanup?.();effects[i]={deps,cleanup:fn()}})}}
const view={}
new Function('exports','require','React',compile('../src/pages/object/RgwTopology.tsx'))(view,name=>name==='react'?hooks:name==='antd'?{Alert:'Alert',Button:'Button',Card:'Card',Space:'Space',Tree:'Tree'}:{readRgwTopology:()=>{count++;return new Promise((yes,no)=>{resolve=yes;reject=no})}},{createElement:(type,props,...children)=>({type,props:{...props,children}})})
let props={clusterId:7}
function render(){cursor=0;const tree=view.RgwTopologyView(props);pending.splice(0).forEach(fn=>fn());return tree}
function nodes(node){return !node||typeof node!=='object'?[]:[node,...(node.props?.children??[]).flat(Infinity).flatMap(nodes)]}
const find=type=>nodes(render()).find(node=>node.type===type)
const settle=async()=>{await Promise.resolve();await Promise.resolve()}
render();assert.equal(count,0)
const click=find('Button').props.onClick;click();click();assert.equal(count,1)
resolve(result);await settle();assert.deepEqual(find('Tree').props.treeData,result.nodes)
find('Tree').props.onSelect(['r'],{selected:true,node:result.nodes[0]});assert.match(find('pre').props.children[0],/realm/)
find('Button').props.onClick();assert.equal(find('Tree'),undefined);assert.equal(find('pre'),undefined)
const previous=resolve;props={clusterId:8};render();previous(result);await settle();assert.equal(find('Tree'),undefined)
find('Button').props.onClick();reject(new Error('private error'));await settle();assert.ok(!JSON.stringify(render()).includes('private error'))
find('Button').props.onClick();effects.forEach(effect=>effect?.cleanup?.());resolve(result);await settle();assert.equal(find('Tree'),undefined)
props={clusterId:undefined};render();assert.equal(find('Button').props.disabled,true)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8'),/function MultisitePage\(\)[\s\S]*?<RgwTopologyView key=\{selectedClusterId/)
console.log('Local RGW topology preserves identities, orphan nodes, paging and cluster isolation')

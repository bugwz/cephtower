import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api={}
const jsx=(type,props)=>({type,props})
new Function('exports','require',ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwCurrentPeriod.tsx',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.ReactJSX}}).outputText)(api,name=>name==='antd'?{Table:'Table'}:{jsx,jsxs:jsx})
const period={id:'p',realm_id:'r',epoch:2,realm_epoch:1,master_zone:'<zone>',period_map:{zonegroups:[{id:'g',name:'<name>',master_zone:'z',zones:[],sync_policy:{groups:[]}}]}}
const props={value:period,realm:'r',current:'p'}
function nodes(node){if(Array.isArray(node))return node.flatMap(nodes);if(!node||typeof node!=='object')return [];return [node,...nodes(node.props?.children)]}
assert.match(api.currentPeriodSummary(period,'r','p'),/epoch 2.*Realm epoch 1/)
for(const value of [null,{}, {...period,realm_id:'other'},{...period,id:'other'}])assert.match(api.currentPeriodSummary(value,'r','p'),/不可用/)
assert.match(api.currentPeriodSummary({...period,epoch:2**53},'r','p'),/非精确/)
const view=api.RgwCurrentPeriod(props)
const table=nodes(view).find(n=>n.type==='Table')
assert.equal(table.props.dataSource[0].name,'<name>')
assert.match(JSON.stringify(view),/不是待提交配置差异/)
assert.ok(nodes(view).every(n=>!n.props?.dangerouslySetInnerHTML))
for(const groups of [null,{},[null],[{id:'g'}],[{id:'g',name:'a'},{id:'g',name:'b'}]])assert.match(JSON.stringify(api.RgwCurrentPeriod({...props,value:{...period,period_map:{zonegroups:groups}}})),/列表不可用/)
assert.equal(nodes(api.RgwCurrentPeriod({...props,value:{...period,period_map:{zonegroups:[]}}})).find(n=>n.type==='Table').props.locale.emptyText,'此 Period 的 Zonegroup 列表为空')
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8'),/<RgwCurrentPeriod value=\{value\} realm=\{row.id\} current=\{row.current_period\} \/>/)
console.log('realm current period snapshot presentation checks passed')

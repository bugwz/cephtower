import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const compile=name=>ts.transpileModule(readFileSync(new URL(`../src/pages/object/${name}`,import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.ReactJSX}}).outputText
const rules={},view={},jsx=(type,props)=>({type,props})
const require=name=>name==='antd'?{Table:'Table'}:name==='./RgwBucketLifecycleRules'?rules:{jsx,jsxs:jsx}
new Function('exports','require',compile('RgwBucketLifecycleRules.tsx'))(rules,require)
new Function('exports','require',compile('RgwBucketLifecycleTransitions.tsx'))(view,require)
const selector={kind:'Filter',and:true,prefix:'images/',tags:[{key:'a',value:''}],object_size_greater_than:'9007199254740993',object_size_less_than:null,archive_zone:false}
const rule={id:'<rule>',status:'Disabled',selector,actions:[{type:'Expiration',fields:{Days:'30'}},{type:'Transition',fields:{Days:'0',StorageClass:'COLD'}},{type:'Transition',fields:{Date:'2030-01-01T00:00:00Z',StorageClass:'DATE'}},{type:'NoncurrentVersionTransition',fields:{NoncurrentDays:'3',StorageClass:'HISTORY'}}]}
const original=JSON.stringify(rule),rows=view.bucketLifecycleTransitions([rule,rule])
assert.equal(rows.length,6);assert.equal(new Set(rows.map(row=>row.key)).size,6)
assert.equal(rows[0].ruleIndex,1);assert.equal(rows[3].ruleIndex,2)
assert.equal(rows[0].status,'Disabled');assert.equal(rows[0].storageClass,'"COLD"');assert.match(rows[0].timing,/"0"/)
assert.match(rows[1].timing,/UTC 日期.*2030/);assert.equal(rows[2].version,'非当前版本');assert.match(rows[2].timing,/非当前版本.*"3"/)
assert.match(rows[0].selector,/9007199254740993/);assert.match(rows[0].selector,/images/)
assert.equal(JSON.stringify(rule),original)
const show=(value,configured)=>view.RgwBucketLifecycleTransitions({value,configured})
assert.equal(show([],false).props.children,'未配置生命周期')
for(const [value,configured] of [[[],true],[[rule],false],[null,true],[[rule],undefined],[[{...rule,status:'unknown'}],true]])assert.match(show(value,configured).props.children,/不可用/)
assert.match(show([{...rule,actions:[rule.actions[0]]}],true).props.children,/没有当前或非当前版本/)
for(const fields of [{Days:'1',Date:'2030-01-01',StorageClass:'COLD'},{StorageClass:'COLD'},{Days:'1',NoncurrentDays:'3',StorageClass:'COLD'}]) {
 assert.match(view.bucketLifecycleTransitions([{...rule,actions:[{type:'Transition',fields}]}])[0].timing,/异常/)
}
assert.equal(view.bucketLifecycleTransitions([{...rule,actions:[{type:'Transition',fields:{Days:'1'}}]}])[0].storageClass,'未返回')
const empty=view.bucketLifecycleTransitions([{...rule,actions:[{type:'Transition',fields:{Days:'0',StorageClass:''}}]}])[0]
assert.equal(empty.storageClass,'""')
const table=show([rule],true)
assert.equal(table.type,'Table');assert.equal(table.props.dataSource.length,3)
for(const name of ['status','version','storageClass']) {
 const column=table.props.columns.find(column=>column.dataIndex===name)
 assert.equal(column.onFilter(rows[0][name],rows[0]),true)
 assert.equal(column.onFilter('not-a-value',rows[0]),false)
 assert.ok(column.filters.length)
}
assert.equal(table.props.columns.find(c=>c.dataIndex==='id').render('<rule>'),'"<rule>"')
const pages=readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8')
assert.match(pages,/key: 'lifecycle_transitions'.*filterKey: false.*row.kind === 'lifecycle'.*<RgwBucketLifecycleTransitions value=\{row.lifecycle_rules\}/)
console.log('Lifecycle tiering displays every scoped transition, exact text and filterable native targets')

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const jsx=(type,props)=>({type,props}),runtime={jsx,jsxs:jsx,Fragment:'Fragment'}
const compile=name=>ts.transpileModule(readFileSync(new URL(`../src/pages/object/${name}.tsx`,import.meta.url),'utf8'),{compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.CommonJS,jsx:ts.JsxEmit.ReactJSX}}).outputText
const parser={},api={}
new Function('exports','require',compile('RgwBucketLifecycleRules'))(parser,name=>name==='react/jsx-runtime'?runtime:{Table:'Table'})
new Function('exports','require',compile('RgwBucketLifecycleDocument'))(api,name=>name==='react/jsx-runtime'?runtime:name==='./RgwBucketLifecycleRules'?parser:{Tabs:'Tabs',Alert:'Alert'})
const rule={id:'<script>not HTML</script>',status:'Disabled',selector:{kind:'Filter',and:true,prefix:'',tags:[{key:'a',value:''},{key:'a',value:'b'}],object_size_greater_than:'9007199254740993',object_size_less_than:null,archive_zone:false},actions:[{type:'Expiration',fields:{Days:'0',ExpiredObjectDeleteMarker:'false'}}]}
const document='<LifecycleConfiguration>\r\n  <!-- keep raw -->\r\n</LifecycleConfiguration>'
const render=(rules=[rule],configured=true,raw=document)=>api.RgwBucketLifecycleDocument({document:raw,rules,configured})
const view=render()
assert.equal(view.props.defaultActiveKey,'xml')
assert.equal(view.props.items[0].children.props.children,document)
assert.deepEqual(JSON.parse(view.props.items[1].children.props.children[1].props.children),[rule])
assert.equal(render([{...rule,private:'not projected'}]).props.items[1].children.props.children[1].props.children.includes('not projected'),false)
const second={...rule,id:'second'}
assert.deepEqual(JSON.parse(render([rule,second]).props.items[1].children.props.children[1].props.children).map(rule=>rule.id),[rule.id,'second'])
for(const rules of [undefined,null,[],[{}]]){
  const result=api.RgwBucketLifecycleDocument({document,rules,configured:true})
  assert.equal(result.props.items[0].children.props.children,document)
  assert.equal(result.props.items[1].children.props.type,'warning')
}
assert.equal(render([],false,null).props.children,'未配置生命周期，无配置文档')
for(const configured of [null,'true',0])assert.equal(render([rule],configured).props.children,'生命周期配置文档不可用')
for(const raw of [null,'',0])assert.equal(render([rule],true,raw).props.children,'生命周期配置文档不可用')
const pages=readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8')
assert.ok(pages.includes("row.kind === 'lifecycle' ? <RgwBucketLifecycleDocument document={value} rules={row.lifecycle_rules} configured={row.configured} />"))
console.log('Lifecycle XML and structured JSON views preserve raw text, field precision and unavailable states')

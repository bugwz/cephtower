import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api = {}
const jsx = (type, props) => ({ type, props })
new Function('exports','require',ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwTopicDetails.tsx',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.ReactJSX}}).outputText)(api,()=>({jsx,jsxs:jsx}))
assert.equal(api.topicText('18446744073709551615'),'18446744073709551615')
assert.equal(api.topicText('0'),'0')
assert.equal(api.topicText('None'),'None')
assert.equal(api.topicText(''),'原生空值')
for (const value of [undefined,null,42,{}]) assert.equal(api.topicText(value),'未返回或不可用')
assert.equal(api.topicBoolean(false),'否')
assert.equal(api.topicBoolean(true),'是')
for (const value of [undefined,'false',0]) assert.equal(api.topicBoolean(value),'未返回或不可用')
assert.match(api.topicEndpoint({push_endpoint:'https://host/path',endpoint_redacted:true}),/host\/path.*隐藏/)
assert.match(api.topicEndpoint({push_endpoint:null,endpoint_redacted:true}),/不可用/)
const details = api.RgwTopicDetails({row:{policy:'<script>bad</script>',opaqueData:'<tag>',max_retries:'18446744073709551615'}})
assert.match(details.props.children[0].props.children,/不代表.*消息已投递/)
const fields = details.props.children.find(child => child.type === 'dl').props.children
const policy = fields.find(field => field.props.children[0].props.children === 'Topic Policy')
assert.equal(policy.props.children[1].props.children.props.children,'<script>bad</script>')
const text = readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8')
const source = ts.createSourceFile('pages.tsx',text,ts.ScriptTarget.Latest,true,ts.ScriptKind.TSX)
let definition
function visit(node) {
  if (ts.isPropertyAssignment(node) && node.name.getText(source) === 'rgwTopics' && ts.isObjectLiteralExpression(node.initializer)) {
    const code = ts.transpileModule(`const definition = ${node.initializer.getText(source)}`,{compilerOptions:{target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.React}}).outputText
    definition = new Function(...Object.keys(api),`${code}; return definition`)(...Object.values(api))
  }
  ts.forEachChild(node,visit)
}
visit(source)
assert.equal(definition.path,'/rgw/topics')
assert.deepEqual(definition.requiredCapabilities,['rgw_admin'])
assert.deepEqual(definition.rowKeyCandidates,['natural_key'])
for (const key of ['name','owner','scope','arn','push_endpoint','persistent','time_to_live','max_retries','retry_sleep_duration']) assert.ok(definition.columns.some(column=>column.key===key))
assert.equal(definition.columns.find(column=>column.key==='scope').render(''),'全局租户')
assert.equal(definition.columns.find(column=>column.key==='scope').render(undefined),'未返回或不可用')
assert.equal(definition.createAction,undefined)
const navigation = readFileSync(new URL('../src/navigation.ts',import.meta.url),'utf8')
assert.match(navigation,/key: 'rgwTopics'.*path: '\/object\/topics'/)
const refresh = readFileSync(new URL('../src/pages/ResourceListPage.tsx',import.meta.url),'utf8')
assert.match(refresh,/'\/rgw\/topics': \['rgw_topic'\]/)
console.log('RGW topic display, precision, redaction notices and page bindings passed')

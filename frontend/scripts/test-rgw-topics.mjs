import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api = {}
new Function('exports',ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwTopicPolicy.ts',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022}}).outputText)(api)
new Function('exports','require',ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwTopicAttribute.ts',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022}}).outputText)(api,()=>api)
new Function('exports','require',ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwTopicEndpoint.ts',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022}}).outputText)(api,()=>api)
new Function('exports',ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwTopicDelete.ts',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022}}).outputText)(api)
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
for (const scope of ['', 'team', 'RGW12345678901234567']) {
  const key = `${scope}:events`
  const row = { natural_key: Buffer.from(key).toString('base64url'), metadata_key: key, name:'events', scope, arn:'arn:topic', metadata_version:'{"tag":"t","ver":9007199254740993}' }
  const action = definition.deleteAction
  assert.equal(action.path,'/rgw/topic')
  assert.equal(action.action,'rgw_topic.delete')
  assert.equal(action.risk,'high')
  assert.equal(action.disabledWhen(row),undefined)
  assert.deepEqual(action.buildBody(row,42),{cluster_id:42,topic_id:row.natural_key,expected_version:row.metadata_version})
  assert.equal(action.resourceKey(row),key)
  for (const warning of ['持久化队列','永久丢失','不会自动清理','主 Zone','原子锁','回滚']) assert.ok(action.confirmation(row).includes(warning))
  for (const change of [{stale:true},{natural_key:'other'},{metadata_version:null},{metadata_key:'other:events'},{scope:undefined}]) assert.throws(()=>action.buildBody({...row,...change},42))
}
const navigation = readFileSync(new URL('../src/navigation.ts',import.meta.url),'utf8')
for (const scope of ['', 'team', 'RGW12345678901234567']) {
  const row = {scope,name:'events',metadata_key:`${scope}:events`,natural_key:Buffer.from(`${scope}:events`).toString('base64url'),arn:`arn:aws:sns:default:${scope}:events`,push_endpoint:'https://old/path',endpoint_redacted:true,stored_secret:true}
  const action = definition.extraActions[2]
  const initial = action.initialValues(row)
  assert.equal(action.path,'/rgw/topic/endpoint')
  assert.equal(action.method,'PATCH')
  assert.equal(action.disabledWhen(row),undefined)
  assert.equal(initial.endpoint_secret,undefined)
  assert.equal(action.fields.find(field=>field.name==='endpoint_secret').type,'password')
  for (const endpoint_secret of ['https://new/path?q=private&x=1','amqp://user:private-password@broker/vhost','amqps://broker/vhost','kafka://broker:9092','http://host/path']) {
    const values = {...initial,endpoint_mode:'replace',endpoint_secret,confirm_endpoint:'acknowledged'}
    assert.deepEqual(action.buildBody(values,42,row),{cluster_id:42,...initial,expected_endpoint:row.push_endpoint,expected_redacted:true,expected_stored_secret:true,endpoint_secret})
    const confirmation = action.confirmation(values,row)
    assert.ok(!confirmation.includes(endpoint_secret))
    assert.ok(!confirmation.includes('private-password'))
    for (const warning of ['永久丢失','EndpointArgs','隐藏凭据','HTTPS','回滚','脱敏快照']) assert.ok(confirmation.includes(warning))
    for (const change of [{stale:true},{natural_key:'other'},{push_endpoint:null},{endpoint_redacted:undefined},{stored_secret:undefined}]) assert.throws(()=>action.buildBody(values,42,{...row,...change}))
    assert.throws(()=>action.buildBody({...values,topic_id:'other'},42,row))
    assert.throws(()=>action.buildBody({...values,confirm_endpoint:undefined},42,row))
  }
  const clear = {...initial,endpoint_mode:'clear',confirm_endpoint:'acknowledged',endpoint_secret:'must-not-be-sent'}
  assert.equal(action.buildBody(clear,42,row).endpoint_secret,'')
  for (const endpoint_secret of ['','ftp://host/a','https://user@host/a','https://user:@host/a','https://:pass@host/a','https://host?query=1','https://host/\n','https://[::1]/','https://host/'+'x'.repeat(1024*1024)]) assert.throws(()=>action.buildBody({...initial,endpoint_mode:'replace',confirm_endpoint:'acknowledged',endpoint_secret},42,row))
  assert.throws(()=>action.buildBody({...initial,endpoint_mode:'replace',confirm_endpoint:'acknowledged',endpoint_secret:row.push_endpoint},42,{...row,endpoint_redacted:false}))
}
for (const scope of ['', 'team', 'RGW12345678901234567']) {
  const row = {scope,name:'events',metadata_key:`${scope}:events`,natural_key:Buffer.from(`${scope}:events`).toString('base64url'),arn:`arn:aws:sns:default:${scope}:events`,opaqueData:'old',persistent:true,time_to_live:'None',max_retries:'10',retry_sleep_duration:'0'}
  const action = definition.extraActions[1]
  assert.equal(action.path,'/rgw/topic/attribute')
  assert.equal(action.method,'PATCH')
  assert.equal(action.disabledWhen(row),undefined)
  const initial = action.initialValues(row)
  for (const [attribute,extra,expected_value,value] of [
    ['OpaqueData',{mode:'set',value:' 中文 &+<>\n'},'old',' 中文 &+<>\n'],
    ['OpaqueData',{mode:'clear'},'old',''],
    ['persistent',{persistent:'false'},'true','false'],
    ['time_to_live',{mode:'set',value:'0'},'None','0'],
    ['max_retries',{mode:'set',value:'2147483647'},'10','2147483647'],
    ['max_retries',{mode:'default'},'10','None'],
    ['retry_sleep_duration',{mode:'set',value:'1'},'0','1']
  ]) {
    const values = {...initial,attribute,...extra}
    assert.deepEqual(action.buildBody(values,42,row),{cluster_id:42,...initial,attribute,expected_value,value})
    for (const change of [{stale:true},{natural_key:'other'},{arn:'wrong'},{scope:undefined}]) assert.throws(()=>action.buildBody(values,42,{...row,...change}))
    assert.throws(()=>action.buildBody({...values,topic_id:'other'},42,row))
    for (const warning of ['HTTPS','原子锁','回滚','Policy']) assert.ok(action.confirmation(values,row).includes(warning))
  }
  for (const value of ['-1','2147483648','01','1.5','','None',' 1','1e3']) assert.throws(()=>action.buildBody({...initial,attribute:'max_retries',mode:'set',value},42,row))
  for (const values of [
    {attribute:'password',value:'secret'}, {attribute:'persistent',persistent:'1'},
    {attribute:'persistent',persistent:'true'}, {attribute:'time_to_live',mode:'default'},
    {attribute:'OpaqueData',mode:'set',value:''}, {attribute:'OpaqueData',mode:'default'},
    {attribute:'max_retries',mode:'clear'}, {attribute:'max_retries',value:'1'},
    {attribute:'OpaqueData',mode:'set',value:'x'.repeat(1024*1024)}
  ]) assert.throws(()=>action.buildBody({...initial,...values},42,row))
  assert.throws(()=>action.buildBody({...initial,attribute:'persistent',persistent:'false'},42,{...row,persistent:undefined}))
  assert.match(action.confirmation({...initial,attribute:'persistent',persistent:'false'},row),/永久丢失/)
  assert.match(action.confirmation({...initial,attribute:'max_retries',mode:'default'},row),/全局默认不是 0/)
}
for (const scope of ['', 'team', 'RGW12345678901234567']) {
  const row = {scope,name:'events',metadata_key:`${scope}:events`,natural_key:Buffer.from(`${scope}:events`).toString('base64url'),arn:`arn:aws:sns:default:${scope}:events`,policy:'{"Statement":[]}'}
  const action = definition.extraActions[0]
  const values = {...action.initialValues(row),policy_mode:'set',policy:'{"Statement":[],"Id":"中文&+"}'}
  assert.equal(action.path,'/rgw/topic/policy')
  assert.equal(action.method,'PATCH')
  assert.equal(action.disabledWhen(row),undefined)
  assert.deepEqual(action.buildBody(values,42,row),{cluster_id:42,topic_id:row.natural_key,topic_arn:row.arn,expected_policy:row.policy,policy:values.policy})
  assert.equal(action.buildBody({...values,policy_mode:'clear'},42,row).policy,'')
  for (const change of [{policy_mode:undefined},{policy:'[]'},{policy:row.policy},{topic_id:'other'},{policy:'{"Id":"'+'x'.repeat(1024*1024)+'"}'}]) assert.throws(()=>action.buildBody({...values,...change},42,row))
  for (const change of [{stale:true},{policy:null},{arn:'arn:aws:sns:default:other:events'},{natural_key:'other'}]) assert.throws(()=>action.buildBody(values,42,{...row,...change}))
  for (const warning of ['HTTPS','回读权限','原子锁','回滚','清除不保证']) assert.ok(action.confirmation(values,row).includes(warning))
}
assert.match(navigation,/key: 'rgwTopics'.*path: '\/object\/topics'/)
const refresh = readFileSync(new URL('../src/pages/ResourceListPage.tsx',import.meta.url),'utf8')
assert.match(refresh,/'\/rgw\/topics': \['rgw_topic'\]/)
assert.match(refresh,/action.path === '\/rgw\/topic'.*kinds: \['rgw_topic'\]/)
console.log('RGW topic display, precision, redaction notices and page bindings passed')

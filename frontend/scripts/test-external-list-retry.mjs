import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const jsx=(type,props)=>({type,props}),runtime={jsx,jsxs:jsx,Fragment:'Fragment'}
const compile=source=>ts.transpileModule(source,{compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.CommonJS,jsx:ts.JsxEmit.ReactJSX}}).outputText
const api={}
new Function('exports','require',compile(readFileSync(new URL('../src/components/Page.tsx',import.meta.url),'utf8')))(api,name=>name==='react/jsx-runtime'?runtime:{Alert:'Alert',Card:'Card',Spin:'Spin'})
const source=readFileSync(new URL('../src/pages/ExternalListPage.tsx',import.meta.url),'utf8')
const ast=ts.createSourceFile('external.tsx',source,ts.ScriptTarget.Latest,true,ts.ScriptKind.TSX)
let action
function visit(node){
  if(ts.isJsxAttribute(node)&&node.name.getText(ast)==='errorAction')action=node.initializer.expression.getText(ast)
  ts.forEachChild(node,visit)
}
visit(ast);assert.ok(action)
const code=compile(`export function recovery(Space, Button, ReloadOutlined, filterFormContent, refreshing, selectedClusterId, reload) { return ${action} }`)
const view={}
new Function('exports','require',code)(view,()=>runtime)
let reads=0
const filters={type:'FilterForm'},reload=()=>{reads++}
const recovery=view.recovery('Space','Button','ReloadOutlined',filters,false,7,reload)
function nodes(node){return !node||typeof node!=='object'?[]:[node,...[node.props?.children].flat(Infinity).flatMap(nodes)]}
const children={type:'MutationSurface'}
const failed=nodes(api.Page({error:'endpoint unavailable',errorAction:recovery,children}))
assert.equal(failed.find(node=>node.type==='Alert').props.description,'endpoint unavailable')
assert.ok(failed.includes(filters));assert.ok(!failed.includes(children))
failed.find(node=>node.type==='Button').props.onClick();assert.equal(reads,1)
for(const [refreshing,cluster] of [[true,7],[false,undefined]]){
  const tree=nodes(view.recovery('Space','Button','ReloadOutlined',null,refreshing,cluster,reload))
  assert.equal(tree.find(node=>node.type==='Button').props.disabled,true)
}
assert.ok(!nodes(api.Page({loading:true,error:'old error',errorAction:recovery,children})).includes(recovery))
assert.ok(!nodes(api.Page({error:'',errorAction:recovery,children})).includes(recovery))
assert.ok(nodes(api.Page({error:'',errorAction:recovery,children})).includes(children))
assert.ok(!nodes(api.Page({error:'failed',children})).includes(children))
console.log('External read failures retain retry and filter controls without exposing mutations')

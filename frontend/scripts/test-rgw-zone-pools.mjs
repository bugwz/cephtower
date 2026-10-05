import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api={},jsx=(type,props)=>({type,props})
new Function('exports','require',ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwZonePoolReferences.tsx',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,jsx:ts.JsxEmit.ReactJSX}}).outputText)(api,name=>name==='antd'?{Table:'Table',Alert:'Alert'}:{jsx,jsxs:jsx})
function nodes(node){return Array.isArray(node)?node.flatMap(nodes):node&&typeof node==='object'?[node,...nodes(node.props?.children)]:[]}
const refs=[{field:'log_pool',pool:'<pool>',namespace:'',raw:'<pool>'},{field:'user_uid_pool',pool:'meta',namespace:'users.uid',raw:'meta:users.uid'}]
const view=api.RgwZonePoolReferences({row:{pool_references:refs,pool_references_complete:true}})
const table=nodes(view).find(node=>node.type==='Table')
assert.deepEqual(table.props.dataSource,refs)
assert.match(table.props.columns.find(column=>column.dataIndex==='namespace').render(''),/默认命名空间/)
assert.equal(table.props.columns.find(column=>column.dataIndex==='namespace').render('<namespace>'),'<namespace>')
assert.ok(nodes(view).every(node=>!node.props?.dangerouslySetInnerHTML))
assert.match(JSON.stringify(view),/不证明池存在/)
for(const value of [null,{},[null],[{pool:'p'}],[refs[0],refs[0]]])assert.match(JSON.stringify(api.RgwZonePoolReferences({row:{pool_references:value}})),/未采集或格式异常/)
const partial=api.RgwZonePoolReferences({row:{pool_references:[],pool_references_complete:false,pool_reference_issues:['missing field']}})
assert.ok(nodes(partial).some(node=>node.type==='Alert'))
assert.match(nodes(partial).find(node=>node.type==='Table').props.locale.emptyText,/不代表没有引用/)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8'),/key: 'pool_references'.*<RgwZonePoolReferences row=\{row\}/)
console.log('Zone pool references preserve exact names, namespaces and incomplete evidence')

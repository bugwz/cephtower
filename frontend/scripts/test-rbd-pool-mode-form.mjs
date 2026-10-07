import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const exports = {}
const source = readFileSync(new URL('../src/pages/block/rbdMirrorPoolMode.ts',import.meta.url),'utf8')
new Function('exports',ts.transpileModule(source,{compilerOptions:{module:ts.ModuleKind.CommonJS}}).outputText)(exports)
const row = {pool:'pool-a',mode:'image',peers:[]}
for(const mode of ['disabled','image','pool']) assert.deepEqual(exports.mirrorPoolModeBody({pool:'foreign',mode},7,row),{cluster_id:7,pool:'pool-a',mode})
for(const pool of [undefined,null,7,'','pool\n']) assert.throws(()=>exports.mirrorPoolModeBody({mode:'pool'},7,{pool}))
for(const mode of [undefined,'future','',[],['disabled']]) assert.throws(()=>exports.mirrorPoolModeBody({mode},7,row))
for(const peers of [undefined,null,{},[{}]]) assert.throws(()=>exports.mirrorPoolModeBody({mode:'disabled'},7,{...row,peers}))
assert.equal(exports.mirrorPoolModeBody({mode:'disabled'},7,{pool:'pool-a',mode:'disabled'}).mode,'disabled')
const pages = readFileSync(new URL('../src/pages/block/pages.tsx',import.meta.url),'utf8')
const ast = ts.createSourceFile('pages.tsx',pages,ts.ScriptTarget.Latest,true,ts.ScriptKind.TSX)
let found
function visit(node) {
  if(ts.isObjectLiteralExpression(node)&&node.properties.some(prop=>ts.isPropertyAssignment(prop)&&prop.name.getText(ast)==='title'&&ts.isStringLiteral(prop.initializer)&&prop.initializer.text==='设置池同步模式')) found=node
  ts.forEachChild(node,visit)
}
visit(ast)
assert.ok(found)
assert.equal(found.parent.name.getText(ast),'updateAction')
const actionExports={}
new Function('exports','mirrorPoolIdentity','mirrorPoolModeBody',ts.transpileModule(`export const action=${found.getText(ast)}`,{compilerOptions:{module:ts.ModuleKind.CommonJS}}).outputText)(actionExports,exports.mirrorPoolIdentity,exports.mirrorPoolModeBody)
const action=actionExports.action
assert.equal(action.fields[0].readOnly,true)
assert.deepEqual(action.initialValues(row),{pool:'pool-a',mode:'image'})
assert.equal(action.initialValues({...row,mode:'init-only'}).mode,undefined)
assert.ok(action.confirmation({pool:'foreign',mode:'pool'},row).includes('pool-a'))
assert.ok(!action.confirmation({pool:'foreign',mode:'pool'},row).includes('foreign'))
assert.equal(action.buildBody({pool:'foreign',mode:'pool'},7,row).pool,'pool-a')
console.log('RBD pool mode form is scoped to selected inventory and guards peer disablement')

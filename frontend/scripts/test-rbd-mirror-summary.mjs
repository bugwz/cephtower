import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const ui = {}, jsx = (type, props) => ({ type, props })
const source = readFileSync(new URL('../src/pages/block/RbdMirrorSummary.tsx', import.meta.url), 'utf8')
new Function('exports', 'require', ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText)(ui, name => name === 'antd' ? Object.fromEntries(['Alert','Descriptions','Space','Table','Tag'].map(name=>[name,name])) : {jsx,jsxs:jsx})
for(const [value, expected] of [[0,'0'],[12,'12'],['9007199254740993','9007199254740993']]) assert.equal(ui.mirrorStateCount(value),expected)
for(const value of [null,undefined,true,-1,1.5,NaN,Infinity,9007199254740992,'','01','1e3','12\n',' 1']) assert.equal(ui.mirrorStateCount(value),'未返回有效数量')
for(const [value,color] of [['OK','green'],['WARNING','orange'],['ERROR','red'],['UNKNOWN','default'],['future','default'],[undefined,'default']]) assert.equal(ui.mirrorHealth(value).props.color,color)
assert.equal(ui.mirrorHealth('<script>').props.children,'<script>')
assert.equal(ui.RbdMirrorSummary({mode:'disabled',value:{health:'OK'}}).props.message,'池同步已停用，未查询运行健康状态')
for(const value of [null,undefined,[],'bad']) assert.equal(ui.RbdMirrorSummary({mode:'image',value}).props.type,'warning')
const tree=ui.RbdMirrorSummary({mode:'image',value:{health:'WARNING',daemon_health:'OK',image_health:'ERROR',states:{replaying:0,error:2,'unknown (42)':1}}})
assert.deepEqual(tree.props.children[0].props.items.map(item=>item.children.props.children),['WARNING','OK','ERROR'])
assert.deepEqual(tree.props.children[2].props.dataSource,[{state:'replaying',count:'0'},{state:'error',count:'2'},{state:'unknown (42)',count:'1'}])
assert.equal(ui.RbdMirrorSummary({mode:'image',value:{states:{}}}).props.children[2].props.type,'info')
assert.equal(ui.RbdMirrorSummary({mode:'image',value:{}}).props.children[2].props.type,'warning')
assert.ok(readFileSync(new URL('../src/pages/block/pages.tsx',import.meta.url),'utf8').includes('<RbdMirrorSummary value={value} mode={row.mode} />'))
console.log('RBD pool mirror summary distinguishes native health, zero counts, and unavailable data')

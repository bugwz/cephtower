import assert from 'node:assert/strict'
import './test-rgw-current-period.mjs'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwPeriodCommit.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(api)
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const source = ts.createSourceFile('pages.tsx', pages, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let action
function visit(node) {
 if (ts.isObjectLiteralExpression(node) && node.properties.some(p => ts.isPropertyAssignment(p) && p.name.getText(source) === 'title' && p.initializer.text === '提交 Realm Period')) {
 const code = ts.transpileModule(`const action = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
 action = new Function(...Object.keys(api), `${code}; return action`)(...Object.values(api))
 }
 ts.forEachChild(node,visit)
}
visit(source)
const row={id:'realm-id',current_period:'period-id'}
const values={...action.initialValues(row),confirm_commit:'acknowledged'}
assert.equal(action.path,'/rgw/period/commit')
assert.deepEqual(action.buildBody(values,7,row),{cluster_id:7,realm_id:'realm-id',expected_current_period:'period-id'})
assert.match(action.confirmation(values,row),/realm-id.*全部待提交.*不自动回滚或重试/)
for(const change of [{realm_id:'wrong'},{expected_current_period:'wrong'},{confirm_commit:true}]) assert.throws(()=>action.buildBody({...values,...change},7,row))
for(const bad of [{},{...row,stale:true},{...row,id:4},{...row,current_period:''}]) assert.ok(action.disabledWhen(bad))
assert.doesNotMatch(pages,/function PeriodCommitPanel/)
console.log('realm-scoped period commit form checks passed')

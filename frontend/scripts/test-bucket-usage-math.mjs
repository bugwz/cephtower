import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
export const math = {}
const code = ts.transpileModule(readFileSync(new URL('../src/pages/object/bucketUsageMath.ts', import.meta.url), 'utf8'), {compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022}}).outputText
new Function('exports',code)(math)
assert.equal(math.averageBucketObjectBytes({object_count:'0',size_actual_bytes:'0'}),'不适用（对象数为 0）')
assert.equal(math.averageBucketObjectBytes({object_count:'3',size_actual_bytes:'10'}),'3.33 B/对象')
assert.equal(math.averageBucketObjectBytes({object_count:'1',size_actual_bytes:'36893488147419103230'}),'36893488147419103230.00 B/对象')
assert.equal(math.averageBucketObjectBytes({object_count:'1000',size_actual_bytes:'1'}),'0.00 B/对象')
console.log('Bucket average sizes preserve fixed-point precision and zero-object semantics')

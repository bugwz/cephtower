import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const helpers = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwLifecycleProgress.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(helpers)
const render = helpers.rgwLifecycleProgress
for (const value of [undefined, null, {}, [], { found: false }, { found: true, entry: null }, { found: false, entry: {} }, { found: true, entry: { bucket: 'key', status: 'COMPLETE', started: 0 } }]) assert.equal(render(value), '生命周期进度不可用')
assert.match(render({ found: false, entry: null }), /无生命周期处理记录.*不代表未配置规则/)
for (const [status, label] of [['UNINITIAL', '尚未初始化'], ['PROCESSING', '处理中'], ['FAILED', '处理失败'], ['COMPLETE', '本轮处理完成']]) {
  const text = render({ found: true, entry: { bucket: 'team:photos:m1', status, started: 'Fri, 02 Oct 2026 00:00:00 GMT' } })
  assert.ok(text.includes(label))
  assert.ok(text.includes(`原生状态：${status}`))
  if (status !== 'UNINITIAL') assert.ok(text.includes('GMT'))
}
assert.match(render({ found: true, entry: { bucket: 'key', status: 'COMPLETE', started: null } }), /不保证所有对象.*开始时间：不可用/)
for (const status of ['FUTURE', 'toString', '__proto__']) assert.match(render({ found: true, entry: { bucket: 'key', status, started: null } }), /未知状态/)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8'), /key: 'lifecycle_progress'.*render: rgwLifecycleProgress/)
console.log('lifecycle progress display checks passed')

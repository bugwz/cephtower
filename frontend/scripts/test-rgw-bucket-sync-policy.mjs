import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketSyncPolicy.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(api)
const group = { id: 'team-policy', status: 'enabled', pipes: [], data_flow: {} }
const summary = value => api.rgwBucketSyncPolicy(value)
assert.match(summary({ groups: [group] }), /已启用.*不代表 Zonegroup/)
assert.match(summary({ groups: [{ ...group, status: 'allowed' }] }), /允许（未启用）/)
assert.match(summary({ groups: [{ ...group, status: 'forbidden' }] }), /禁止/)
for (const status of ['future', 'constructor', 'Enabled']) assert.match(summary({ groups: [{ ...group, status }] }), /未知/)
assert.match(summary({ groups: [] }), /无桶本地同步组.*不代表/)
for (const value of [null, {}, [], { groups: null }, { groups: [null] }, { groups: [group, group] }, { groups: [{ ...group, status: null }] }, { groups: [{ ...group, pipes: {} }] }, { groups: [{ ...group, data_flow: [] }] }]) assert.match(summary(value), /不可用/)
assert.match(summary({ groups: [{ ...group, id: '<script>' }] }), /"<script>"/)
console.log('bucket local sync policy summary checks passed')

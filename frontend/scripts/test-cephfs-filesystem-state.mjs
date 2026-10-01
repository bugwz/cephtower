import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/file/cephfsFilesystemState.ts', import.meta.url), 'utf8')
const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
const exports = {}
new Function('exports', compiled)(exports)
assert.equal(exports.filesystemEnabledText(true), '已启用')
assert.equal(exports.filesystemEnabledText(false), '未启用')
for (const value of [null, undefined, '', 0, 1, 'true', 'false']) assert.equal(exports.filesystemEnabledText(value), '未知')
console.log('CephFS filesystem state checks passed')

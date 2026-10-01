import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/TelemetryPage.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('telemetry.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const node = tree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'telemetryStatusValue')
const code = ts.transpileModule(node.getText(tree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const display = new Function(`${code}; return telemetryStatusValue`)()
for (const [value, expected] of [[false, '关闭'], [true, '开启'], [undefined, '未提供'], [null, '未提供'], ['', '空字符串'], [0, '0'], ['24', '24'], [{}, '未知']]) assert.equal(display(value), expected)
assert.ok(source.includes("request<{ status: ApiRecord, observed_at: string }>('/manager/telemetry/status'"))
assert.ok(!source.includes('mutateResource'))
console.log('Telemetry status display checks passed')

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/osdOperationalStatus.ts', import.meta.url), 'utf8')
const tree = ts.createSourceFile('status.ts', source, ts.ScriptTarget.Latest, true)
const helper = tree.statements.find(n => ts.isFunctionDeclaration(n))
const code = ts.transpileModule(helper.getText(tree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const status = new Function('isRecord', `${code}; return osdOperationalStatus`)(value => value !== null && typeof value === 'object' && !Array.isArray(value))
const osd = { stale: false }, removal = { stale: false, osd_id: 0 }
assert.equal(status(osd, '0', [], false, false), '未在移除队列中')
assert.equal(status(osd, '0', [removal], false, false), '移除队列中')
assert.equal(status(osd, '0', [{ ...removal, replace: true }], false, false), '移除队列中（保留 ID 替换）')
assert.equal(status(osd, '1', [removal], false, false), '未在移除队列中')
for (const rows of [null, undefined, {}, [null], [{}], [{ ...removal, stale: true }], [{ ...removal, osd_id: '0' }], [{ ...removal, osd_id: -1 }], [{ ...removal, osd_id: 2147483648 }], [removal, removal]]) assert.ok(status(osd, '0', rows, false, false).startsWith('未知'))
for (const stale of [true, undefined, null]) assert.ok(status(osd, '0', [], stale, false).startsWith('未知'))
assert.ok(status({ stale: true }, '0', [], false, false).startsWith('未知'))
assert.ok(status(osd, '0', [], false, true).startsWith('未知'))
for (const id of ['all', '01', '-1', '2147483648']) assert.ok(status(osd, id, [], false, false).startsWith('未知'))
const pages = readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8')
assert.ok(pages.includes('osdOperationalStatus(row, osdID(row), data?.removals, data?.removalMeta?.stale, loading || Boolean(error))'))

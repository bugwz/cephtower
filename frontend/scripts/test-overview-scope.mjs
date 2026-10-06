import assert from 'node:assert/strict'
import './test-overview-capacity.mjs'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/overview/OverviewPage.tsx', import.meta.url), 'utf8')
assert.ok(source.includes("key={selectedClusterId ?? 'none'} selectedClusterId={selectedClusterId}"))
assert.ok(source.includes('return () => { active.current = false }'))
const tree = ts.createSourceFile('OverviewPage.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const content = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'OverviewContent')
const toggle = content.body.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'toggleHealth')
const js = ts.transpileModule(toggle.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2020 } }).outputText
for (const scenario of ['inactive', 'switch-during-write', 'current']) {
  const active = { current: scenario !== 'inactive' }
  const calls = []
  const mutation = { run: async callback => { await callback(); if (scenario === 'switch-during-write') active.current = false } }
  const fn = new Function('active', 'selectedClusterId', 'mutatingHealth', 'textValue', 'setMutatingHealth', 'operationMutation', 'mutateResource', 'setMuteTarget', 'refreshResource', 'refresh', 'message', `${js}; return toggleHealth`)(active, 7, false, value => String(value), value => calls.push(['busy', value]), mutation, async (path, method, body) => calls.push(['write', body.cluster_id, body.code]), () => calls.push(['close']), async body => calls.push(['collect', body.clusterId]), async () => calls.push(['refresh']), { success: () => calls.push(['success']) })
  await fn({ code: 'OSD_DOWN' }, false, { ttl: '1h' })
  if (scenario === 'inactive') assert.deepEqual(calls, [])
  if (scenario === 'switch-during-write') assert.deepEqual(calls, [['busy', true], ['write', 7, 'OSD_DOWN']])
  if (scenario === 'current') assert.deepEqual(calls, [['busy', true], ['write', 7, 'OSD_DOWN'], ['success'], ['close'], ['collect', 7], ['refresh'], ['busy', false]])
}
console.log('Overview health actions remain bound to their mounted cluster scope')

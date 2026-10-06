import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
import './test-osd-recovery-presets.mjs'
import './test-osd-device-class.mjs'
import './test-osd-individual-flags.mjs'
import './test-osd-global-flags.mjs'
import './test-osd-removal-stop.mjs'
import './test-osd-removal-details.mjs'
const scrubSource = readFileSync(new URL('../src/pages/cluster/OSDScrubConfiguration.tsx', import.meta.url), 'utf8')
const scrubTree = ts.createSourceFile('scrub.tsx', scrubSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const scrubOptionsNode = scrubTree.statements.find(n => ts.isVariableStatement(n))
const scrubOptionsJS = ts.transpileModule(scrubOptionsNode.getText(scrubTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const scrubOptions = new Function(`${scrubOptionsJS}; return osdScrubOptions`)()
assert.equal(scrubOptions.length, 31)
assert.equal(new Set(scrubOptions).size, 31)
for (const name of ['osd_scrub_during_recovery', 'osd_scrub_begin_week_day', 'osd_deep_scrub_interval', 'osd_scrub_sleep', 'osd_deep_scrub_large_omap_object_value_sum_threshold', 'osd_scrub_max_preemptions', 'osd_shallow_scrub_chunk_min']) assert.ok(scrubOptions.includes(name))
const scrubPanelNode = scrubTree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'OSDScrubConfiguration')
const scrubPanelJS = ts.transpileModule(scrubPanelNode.getText(scrubTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.React } }).outputText
const scrubPanel = new Function('React', 'Alert', 'ConfigurationPage', 'osdScrubOptions', `${scrubPanelJS}; return OSDScrubConfiguration`)(
  { createElement: (type, props, ...children) => ({ type, props, children }) }, 'Alert', 'ConfigurationPage', scrubOptions)()
assert.equal(scrubPanel.children[1].props.optionNames, scrubOptions)
const pagesSource = readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8')
assert.ok(pagesSource.includes('const scrubConfigOpen = scrubConfigScope === osdScope'))
assert.ok(pagesSource.includes('setScrubConfigScope(osdScope)'))
assert.ok(pagesSource.includes('scrubConfigOpen && selectedClusterId && <OSDScrubConfiguration'))
const recoverySource = readFileSync(new URL('../src/pages/cluster/OSDRecoveryConfiguration.tsx', import.meta.url), 'utf8')
const recoveryTree = ts.createSourceFile('recovery.tsx', recoverySource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const recoveryOptionsNode = recoveryTree.statements.find(n => ts.isVariableStatement(n))
const recoveryOptionsJS = ts.transpileModule(recoveryOptionsNode.getText(recoveryTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const recoveryOptions = new Function(`${recoveryOptionsJS}; return osdRecoveryOptions`)()
assert.deepEqual(recoveryOptions, ['osd_max_backfills', 'osd_recovery_max_active', 'osd_recovery_max_single_start', 'osd_recovery_sleep', 'osd_op_queue', 'osd_mclock_profile', 'osd_mclock_override_recovery_settings'])
assert.ok(recoverySource.includes('optionNames={osdRecoveryOptions}'))
assert.ok(pagesSource.includes('const recoveryConfigOpen = recoveryConfigScope === osdScope'))
assert.ok(pagesSource.includes('setRecoveryConfigScope(osdScope)'))
assert.ok(pagesSource.includes('recoveryConfigOpen && selectedClusterId && <OSDRecoveryConfiguration'))
const source = readFileSync(new URL('../src/pages/cluster/OSDSafetyCheck.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('safety.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const isRecord = value => value !== null && typeof value === 'object' && !Array.isArray(value)
const compile = node => ts.transpileModule(node.getText(tree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const parserNode = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'osdSafetyReport')
const parser = new Function('isRecord', `${compile(parserNode)}; return osdSafetyReport`)(isRecord)
const safe = { is_safe_to_destroy: true, safe_to_destroy: [0], active: [], missing_stats: [], stored_pgs: [] }
assert.equal(parser(safe, '0'), safe)
const unsafe = { ...safe, is_safe_to_destroy: false, safe_to_destroy: [], active: [0] }
assert.equal(parser(unsafe, '0'), unsafe)
for (const report of [null, {}, { ...safe, active: null }, { ...safe, safe_to_destroy: [null] }, { ...safe, active: [0] }, { ...safe, safe_to_destroy: [1] }, { ...safe, is_safe_to_destroy: 'true' }, { ...safe, safe_to_destroy: [] }]) assert.equal(parser(report, '0'), null)
const component = tree.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'OSDSafetyCheck')
const check = component.body.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'check')
for (const scenario of ['safe', 'unsafe', 'invalid', 'failure', 'unmounted', 'busy']) {
  const calls = [], updates = [], scope = { current: {} }, running = { current: scenario === 'busy' }
  const env = { scope, running, clusterId: 17, osdId: '0', isRecord, osdSafetyReport: parser,
    setLoading: value => updates.push(['loading', value]), setError: value => updates.push(['error', value]), setReport: value => updates.push(['report', value]), setCheckedAt: value => updates.push(['time', value]),
    mutateResource: async (...args) => { calls.push(args); if (scenario === 'unmounted') scope.current = null; if (scenario === 'failure') throw new Error('offline'); return { details: { check: scenario === 'invalid' ? {} : scenario === 'unsafe' ? unsafe : safe } } } }
  const run = new Function(...Object.keys(env), `${compile(check)}; return check`)(...Object.values(env))
  await run()
  if (scenario === 'busy') { assert.deepEqual(calls, []); assert.deepEqual(updates, []); continue }
  assert.deepEqual(calls, [['/osd/removal/check', 'POST', { cluster_id: 17, osd_ids: ['0'] }]])
  if (scenario === 'unmounted') assert.equal(updates.length, 4)
  if (scenario === 'safe' || scenario === 'unsafe') assert.ok(updates.some(([kind, value]) => kind === 'report' && value === (scenario === 'safe' ? safe : unsafe)))
  if (scenario === 'failure' || scenario === 'invalid') assert.ok(updates.some(([kind, value]) => kind === 'error' && value !== ''))
  assert.equal(running.current, false)
}

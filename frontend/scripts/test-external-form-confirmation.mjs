import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/ExternalListPage.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('external.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let submitNode, filterNode, cleanNode
function visit(node) {
  if (ts.isFunctionDeclaration(node) && node.name?.text === 'submitForm') submitNode = node
  if (ts.isFunctionDeclaration(node) && node.name?.text === 'applyFilters') filterNode = node
  if (ts.isFunctionDeclaration(node) && node.name?.text === 'cleanRecord') cleanNode = node
  ts.forEachChild(node, visit)
}
visit(tree)
assert.ok(submitNode)
const code = ts.transpileModule(submitNode.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
assert.ok(filterNode && cleanNode)
const filterCode=ts.transpileModule(`${filterNode.getText(tree)}\n${cleanNode.getText(tree)}`,{compilerOptions:{target:ts.ScriptTarget.ES2022}}).outputText

async function run(scenario) {
  const calls = [], states = []
  let confirmations = 0, resets = 0, successes = 0, refreshes = 0, errors = 0
  const env = {
    selectedClusterId: 7, formClusterId: 7, submitting: false, mutationBlocked: false,
    currentClusterId: { current: 7 }, clusterGeneration: { current: 0 }, formGeneration: 0, activeRow: { bucket_id: 'row' },
    activeAction: {
      title: 'Edit', path: '/rgw/bucket/policy', method: 'PATCH', successMessage: 'done',
      disabledWhen: () => scenario === 'blocked' ? 'blocked' : undefined,
      buildBody: (values, clusterId, row) => {
        if (scenario === 'invalid') throw new Error('invalid')
        assert.equal(row.bucket_id, 'row')
        return { cluster_id: clusterId, document: values.document }
      },
      confirmation: () => scenario === 'no-confirmation' ? undefined : 'Replace entire document?'
    },
    Modal: { confirm: (options) => {
      confirmations++
      if (scenario === 'switch-before-confirm') env.currentClusterId.current = 8
      if (scenario === 'round-trip-switch' || scenario === 'unmounted') env.clusterGeneration.current++
      if (scenario === 'filter-before-confirm') env.clusterGeneration.current++
      if (scenario === 'cancel') options.onCancel()
      else if (scenario !== 'close') options.onOk()
      options.afterClose()
    } },
    operationMutation: { run: async (executor) => executor() },
    mutateResource: async (...args) => { calls.push(args); if (scenario === 'switch-after-send') env.currentClusterId.current = 8; if(scenario==='filter-after-send')env.clusterGeneration.current++ },
    form: { resetFields: () => { resets++ } },
    setSubmitting: (value) => states.push(value),
    setFormOpen: (value) => assert.equal(value, false),
    message: { warning: () => {}, success: () => { successes++ }, error: () => { errors++ } },
    refresh: () => { refreshes++ }
  }
  if (scenario === 'stale-form') env.formClusterId = 8
  if (scenario === 'stale-query-form') env.clusterGeneration.current++
  const submit = new Function('env', `with (env) { ${code}; return submitForm }`)(env)
  await submit({ document: 'original' })
  assert.equal(errors, scenario === 'invalid' ? 1 : 0)
  const sent = ['approve', 'no-confirmation', 'switch-after-send','filter-after-send'].includes(scenario)
  assert.equal(calls.length, sent ? 1 : 0, scenario)
  if (sent) assert.deepEqual(calls[0], ['/rgw/bucket/policy', 'PATCH', { cluster_id: 7, document: 'original' }])
  assert.equal(confirmations, ['invalid', 'blocked', 'stale-form', 'stale-query-form', 'no-confirmation'].includes(scenario) ? 0 : 1, scenario)
  const completed = ['approve', 'no-confirmation'].includes(scenario)
  for (const count of [resets, successes, refreshes]) assert.equal(count, completed ? 1 : 0, scenario)
  assert.deepEqual(states, ['blocked', 'stale-form', 'stale-query-form'].includes(scenario) ? [] : [true, false], scenario)
}
for (const scenario of ['approve', 'cancel', 'close', 'invalid', 'blocked', 'stale-form', 'stale-query-form', 'filter-before-confirm', 'filter-after-send', 'switch-before-confirm', 'round-trip-switch', 'unmounted', 'switch-after-send', 'no-confirmation']) await run(scenario)
const changed=[],env={clusterGeneration:{current:5},form:{resetFields:()=>changed.push('reset')}}
for(const key of ['setFormOpen','setActiveAction','setActiveRow','setDetailRow','setFormClusterId','setQueryBody'])env[key]=value=>{assert.equal(env.clusterGeneration.current,6);changed.push([key,value])}
new Function('env',`with(env){${filterCode};return applyFilters}`)(env)({bucket_id:'new-bucket',kind:'cors',empty:''})
assert.deepEqual(changed,['reset',['setFormOpen',false],['setActiveAction',null],['setActiveRow',undefined],['setDetailRow',null],['setFormClusterId',undefined],['setQueryBody',{bucket_id:'new-bucket',kind:'cors'}]])
assert.ok(source.includes('onFinish={applyFilters}'))
assert.ok(source.includes('setFormGeneration(clusterGeneration.current)'))
assert.ok(source.includes('setFormClusterId(selectedClusterId)'))
assert.ok(source.includes('Boolean(definition.updateAction.disabledWhen?.(row))'))
assert.ok(source.includes('readOnly={field.readOnly}'))
assert.ok(source.includes('disabled={submitting}'))
assert.ok(source.includes('useEffect(() => () => { clusterGeneration.current += 1 }, [])'))
assert.equal((source.match(/if \(!isCurrentScope\(\)\) return/g) ?? []).length, 4)
console.log('External form confirmation preserves scope and cancels stale or rejected submissions')

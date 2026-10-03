import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/ResourceListPage.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('ResourceListPage.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const declaration = tree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name?.text === 'renderFormControl')
assert.ok(declaration)
const code = ts.transpileModule(declaration.getText(tree), { compilerOptions: { jsx: ts.JsxEmit.React, target: ts.ScriptTarget.ES2022 } }).outputText
const React = { createElement: (type, props) => ({ type, props }) }
const render = new Function('React', 'Select', `${code}; return renderFormControl`)(React, 'Select')
for (const multiple of [false, true]) {
  const field = { type: 'select', multiple, options: [{ label: 'SMB', value: 'smb' }], placeholder: '选择标签' }
  const control = render(field)
  assert.equal(control.type, 'Select')
  assert.equal(control.props.allowClear, true)
  assert.equal(control.props.mode, multiple ? 'multiple' : undefined)
  assert.deepEqual(control.props.options, field.options)
  assert.equal(control.props.placeholder, field.placeholder)
  assert.equal(render({ ...field, required: true }).props.allowClear, false)
  const readOnly = render({ ...field, readOnly: true })
  assert.equal(readOnly.props.allowClear, false)
  assert.equal(readOnly.props.disabled, true)
}
console.log('Resource form select clearing checks passed')

const blockSource = readFileSync(new URL('../src/pages/block/pages.tsx', import.meta.url), 'utf8')
const blockTree = ts.createSourceFile('pages.tsx', blockSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let scheduleNode
function findSchedule(node) {
  if (ts.isObjectLiteralExpression(node) && node.properties.some((property) => ts.isPropertyAssignment(property) && property.name.getText(blockTree) === 'path' && property.initializer.getText(blockTree) === "'/rbd/mirroring/schedule'")) scheduleNode = node
  ts.forEachChild(node, findSchedule)
}
findSchedule(blockTree)
assert.ok(scheduleNode)
const scheduleCode = ts.transpileModule(`const schedule = ${scheduleNode.getText(blockTree)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const schedule = new Function(`${scheduleCode}; return schedule`)()
const poolRow = { pool: 'images' }
const addValues = { action: 'mirror-schedule-add', interval: '12h', remove_interval: '1d', start_time: '01:00' }
assert.deepEqual(schedule.buildBody(addValues, 7, poolRow), { cluster_id: 7, pool: 'images', action: 'mirror-schedule-add', interval: '12h', start_time: '01:00' })
const removeValues = { ...addValues, action: 'mirror-schedule-remove', remove_interval: '' }
assert.deepEqual(schedule.buildBody(removeValues, 7, poolRow), { cluster_id: 7, pool: 'images', action: 'mirror-schedule-remove' })
assert.match(schedule.confirmation(removeValues, poolRow), /全部池级调度/)
assert.deepEqual(schedule.buildBody({ ...removeValues, remove_interval: '1d' }, 7, poolRow), { cluster_id: 7, pool: 'images', action: 'mirror-schedule-remove', interval: '1d', start_time: '01:00' })
const intervalField = schedule.fields.find((field) => field.name === 'interval')
assert.equal(intervalField.required, true)
assert.equal(intervalField.visibleWhen(addValues), true)
assert.equal(intervalField.visibleWhen(removeValues), false)
assert.equal(schedule.fields.find((field) => field.name === 'start_time').visibleWhen(removeValues), false)
console.log('Pool mirror schedule form scope checks passed')

const mirrorExports = {}
const mirrorCode = ts.transpileModule(readFileSync(new URL('../src/pages/block/rbdMirrorScheduleRows.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText
new Function('exports', mirrorCode)(mirrorExports)
const mirrorRows = mirrorExports.mirrorScheduleRows
const items = [{ interval: '1h', start_time: '01:00:00+08:00' }, { interval: '1d', start_time: '' }]
const rows = mirrorRows([
  { pool: '-', namespace: '-', image: '-', items },
  { pool: 'images', namespace: '-', image: '-', items },
  { pool: 'images', namespace: 'team', image: '-', items },
  { pool: 'images', namespace: '', image: 'vm', items }
])
assert.deepEqual(rows.filter((_, index) => index % 2 === 0).map((row) => row.scope), ['集群', '池', '命名空间', '镜像'])
assert.equal(rows[6].target, 'images / 默认命名空间 / vm')
assert.equal(rows[0].startTime, '01:00:00+08:00')
assert.equal(rows[1].startTime, '未指定')
assert.equal(new Set(rows.map((row) => row.key)).size, 8)
assert.deepEqual(mirrorRows([]), [])
for (const value of [null, {}, [null], [{}], [{ pool: 'images', namespace: '-', image: '-', items: null }], [{ pool: '-', namespace: 'team', image: '-', items }], [{ pool: 'images', namespace: '-', image: '-', items: [{}] }]]) assert.equal(mirrorRows(value), undefined)
console.log('Mirror schedule scope display checks passed')

const imageScheduleDetails = mirrorExports.imageMirrorScheduleDetails
for (const [origin, label] of [['cluster', '继承集群'], ['pool', '继承池'], ['namespace', '继承命名空间'], ['', '镜像专属']]) {
  const details = imageScheduleDetails({ name: origin === 'cluster' ? '' : 'images/team/vm', inherited_from: origin, schedule_time: '2026-10-03 12:30:00', schedule_interval: items })
  assert.equal(details.origin, label)
  assert.equal(details.target, origin === 'cluster' ? '集群' : 'images/team/vm')
  assert.equal(details.nextRun, '2026-10-03 12:30:00')
  assert.equal(details.intervals[0].startTime, '01:00:00+08:00')
  assert.equal(details.intervals[1].startTime, '未指定')
}
assert.equal(imageScheduleDetails({ name: 'images/vm', schedule_interval: items }).origin, '镜像专属')
assert.equal(imageScheduleDetails({ name: 'images/vm', schedule_interval: items }).nextRun, '未返回下次运行时间')
for (const value of [null, [], {}, { name: 'images/vm', schedule_interval: [] }, { name: 'images/vm', inherited_from: 'constructor', schedule_interval: items }, { name: 'images/vm', schedule_interval: [{}] }, { name: 'images/vm', schedule_time: 123, schedule_interval: items }]) assert.equal(imageScheduleDetails(value), undefined)
console.log('Image mirror schedule inheritance display checks passed')

let namespaceScheduleNode
function findNamespaceSchedule(node) {
  if (ts.isObjectLiteralExpression(node) && node.properties.some((property) => ts.isPropertyAssignment(property) && property.name.getText(blockTree) === 'path' && property.initializer.getText(blockTree) === "'/rbd/namespace/schedule'")) namespaceScheduleNode = node
  ts.forEachChild(node, findNamespaceSchedule)
}
findNamespaceSchedule(blockTree)
assert.ok(namespaceScheduleNode)
const namespaceHelpers = blockTree.statements.filter((node) => ts.isFunctionDeclaration(node) && ['namespacePool', 'namespaceName'].includes(node.name?.text)).map((node) => node.getText(blockTree)).join('\n')
const namespaceCode = ts.transpileModule(`${namespaceHelpers}\nconst schedule = ${namespaceScheduleNode.getText(blockTree)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const namespaceSchedule = new Function(`${namespaceCode}; return schedule`)()
assert.deepEqual(namespaceSchedule.buildBody(addValues, 9, { pool: 'images', namespace: 'team' }), { cluster_id: 9, pool: 'images', namespace: 'team', action: 'mirror-schedule-add', interval: '12h', start_time: '01:00' })
assert.deepEqual(namespaceSchedule.buildBody(removeValues, 9, { pool: 'images', namespace: 'team' }), { cluster_id: 9, pool: 'images', namespace: 'team', action: 'mirror-schedule-remove' })
assert.match(namespaceSchedule.confirmation(removeValues, { pool: 'images', namespace: 'team' }), /images\/team.*全部命名空间级调度/)
assert.equal(namespaceSchedule.fields.find((field) => field.name === 'interval').required, true)
console.log('Namespace mirror schedule form checks passed')

const runtimeExports = {}
const runtimeCode = ts.transpileModule(readFileSync(new URL('../src/pages/block/rbdRuntimeFields.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText
new Function('exports', runtimeCode)(runtimeExports)
assert.equal(runtimeExports.runtimeValue('18446744073709551615'), '18446744073709551615')
assert.equal(runtimeExports.runtimeValue(0), '0')
assert.equal(runtimeExports.runtimeValue(false), '否')
assert.equal(runtimeExports.runtimeValue(true), '是')
for (const value of [undefined, null, NaN, Infinity, 9007199254740992, {}, []]) assert.equal(runtimeExports.runtimeValue(value), '未返回或值无效')
assert.deepEqual(runtimeExports.runtimeWatchers([]), [])
assert.equal(runtimeExports.runtimeWatchers(null), undefined)
assert.equal(runtimeExports.runtimeWatchers([null]), undefined)
assert.deepEqual(runtimeExports.runtimeWatchers([{ address: '10.0.0.1:0/1', client: '9007199254740993', cookie: '18446744073709551615' }]), [{ key: '0', address: '10.0.0.1:0/1', client: '9007199254740993', cookie: '18446744073709551615' }])
assert.equal(runtimeExports.runtimeDetails(null, runtimeExports.cacheFields), undefined)
const cacheRows = runtimeExports.runtimeDetails({ clean: false, dirty_bytes: '0', hits_full: '18446744073709551615' }, runtimeExports.cacheFields)
assert.equal(cacheRows.find((row) => row.key === 'clean').children, '否')
assert.equal(cacheRows.find((row) => row.key === 'dirty_bytes').children, '0')
assert.equal(cacheRows.find((row) => row.key === 'hits_full').children, '18446744073709551615')
assert.equal(cacheRows.length, 20)
assert.equal(runtimeExports.runtimeDetails({ source_pool_namespace: '' }, runtimeExports.migrationFields).find((row) => row.key === 'source_pool_namespace').children, '默认命名空间')
console.log('RBD runtime status field checks passed')

const configExports = {}
const configCode = ts.transpileModule(readFileSync(new URL('../src/pages/block/rbdConfigurationRows.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText
new Function('exports', configCode)(configExports)
const configRows = configExports.rbdConfigurationRows([
  { name: 'rbd_qos_iops_limit', value: '18446744073709551615', source: 'image' },
  { name: 'rbd_cache', value: false, source: 'pool' },
  { name: 'rbd_zero', value: 0, source: 'config' },
  { name: 'rbd_empty', value: '', source: 'unknown (9)' },
  { name: 'rbd_missing' }
])
assert.equal(configRows[0].value, '18446744073709551615')
assert.equal(configRows[1].value, 'false')
assert.equal(configRows[2].value, '0')
assert.equal(configRows[3].value, '（空字符串）')
assert.deepEqual(configRows.map((row) => row.sourceLabel), ['镜像级覆盖', '池级覆盖', '客户端配置', '未知来源：unknown (9)', '来源未返回'])
assert.equal(configExports.filterRbdConfiguration(configRows, ' IOPS ', 'image').length, 1)
assert.equal(configExports.filterRbdConfiguration(configRows, 'IOPS', 'pool').length, 0)
assert.equal(configExports.filterRbdConfiguration(configRows, '', '').length, 5)
assert.deepEqual(configExports.rbdConfigurationRows([]), [])
for (const value of [null, {}, [null], [{}], [{ name: '' }]]) assert.equal(configExports.rbdConfigurationRows(value), undefined)
assert.equal(configExports.rbdConfigurationRows([{ name: 'x', value: 9007199254740992, source: 'constructor' }])[0].value, '未返回或值无效')
console.log('RBD configuration values and source filters passed')

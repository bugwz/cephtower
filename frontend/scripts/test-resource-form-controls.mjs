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

const protectionSource = readFileSync(new URL('../src/pages/block/RbdSnapshotProtection.tsx', import.meta.url), 'utf8')
const protectionTree = ts.createSourceFile('protection.tsx', protectionSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const protectionNode = protectionTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'snapshotProtection')
const protectionCode = ts.transpileModule(protectionNode.getText(protectionTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const protection = new Function(`${protectionCode}; return snapshotProtection`)()
assert.deepEqual(protection(true), { label: '已保护', color: 'green' })
assert.deepEqual(protection(false), { label: '未保护', color: 'blue' })
for (const value of [undefined, null, 0, 1, 'false', 'true', {}, []]) assert.deepEqual(protection(value), { label: '保护状态未知', color: 'default' })
console.log('Snapshot protection tags distinguish explicit booleans from unknown values')

const trashSource = readFileSync(new URL('../src/pages/block/RbdTrashStatus.tsx', import.meta.url), 'utf8')
const trashTree = ts.createSourceFile('trash.tsx', trashSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const trashNode = trashTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'rbdTrashStatus')
const trashCode = ts.transpileModule(trashNode.getText(trashTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const trashStatus = new Function(`${trashCode}; return rbdTrashStatus`)()
const deadline = 'Sat Oct  3 12:00:00 2026'
assert.deepEqual(trashStatus(`protected until ${deadline}`), { label: '延期保护中（采集时）', color: 'blue', deadline })
assert.deepEqual(trashStatus(`expired at ${deadline}`), { label: '已到期（采集时）', color: 'orange', deadline })
for (const value of [null, undefined, 0, '']) assert.equal(trashStatus(value).label, '延期状态未知')
for (const value of ['expired at ', 'protected until ', 'unexpected']) assert.deepEqual(trashStatus(value), { label: '无法识别的延期状态', color: 'default', deadline: value })
console.log('Trash deferment display preserves native times and collection-time state')
const restoreNode = trashTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'rbdTrashRestoreReason')
const restoreCode = ts.transpileModule(restoreNode.getText(trashTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const restoreReason = new Function(`${restoreCode}; return rbdTrashRestoreReason`)()
for (const trash_source of ['USER', 'USER_PARENT', 'MIRRORING']) assert.equal(restoreReason({ trash_source }), undefined)
assert.match(restoreReason({ trash_source: 'MIGRATION' }), /迁移来源/)
assert.match(restoreReason({ trash_source: 'REMOVING' }), /正在删除/)
for (const trash_source of [null, undefined, '', 'user', 'unknown', ['USER'], {}]) assert.match(restoreReason({ trash_source }), /来源未知/)

const parentSource = readFileSync(new URL('../src/pages/block/RbdParent.tsx', import.meta.url), 'utf8')
const parentTree = ts.createSourceFile('parent.tsx', parentSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const parentNode = parentTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'rbdParentDetails')
const parentCode = ts.transpileModule(parentNode.getText(parentTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const parentDetails = new Function(`${parentCode}; return rbdParentDetails`)()
const parent = { pool: 'images', pool_namespace: 'team', image: 'base', snapshot: 'v1', id: 'abc', trash: false }
assert.deepEqual(parentDetails(parent), { path: 'images/team/base@v1', id: 'abc', trash: '不在回收站' })
assert.equal(parentDetails({ ...parent, pool_namespace: '', trash: true }).path, 'images/base@v1')
assert.equal(parentDetails({ ...parent, trash: true }).trash, '位于回收站')
assert.equal(parentDetails({ pool: 'p', pool_namespace: '', image: 'i', snapshot: 's' }).trash, '未返回')
assert.equal(parentDetails({ ...parent, trash: 'false' }).trash, '未返回')
assert.equal(parentDetails({ ...parent, image: ' base ' }).path, 'images/team/ base @v1')
for (const value of [null, [], {}, { ...parent, snapshot: null }, { ...parent, pool_namespace: undefined }, { ...parent, pool_namespace: 1 }]) assert.equal(parentDetails(value), undefined)
console.log('Native RBD parent paths preserve namespace and explicit trash state')

const childSource = readFileSync(new URL('../src/pages/block/RbdChildren.tsx', import.meta.url), 'utf8')
const childTree = ts.createSourceFile('children.tsx', childSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const childNode = childTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'rbdChildRows')
const childCode = ts.transpileModule(childNode.getText(childTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const childRows = new Function(`${childCode}; return rbdChildRows`)()
assert.deepEqual(childRows([]), [])
assert.deepEqual(childRows([{ pool: 'p', pool_namespace: '', image: 'i', id: 'abc', trash: true }]), [{ key: 0, pool: 'p', namespace: '默认命名空间', image: 'i', id: 'abc', trash: '位于回收站' }])
assert.equal(childRows([{ pool: '', pool_namespace: 'team', image: '', id: 'orphan', trash: false }])[0].image, '名称未解析')
assert.equal(childRows([{ pool: 'p', pool_namespace: 'team', image: 'i' }])[0].trash, '未返回')
for (const value of [null, {}, [null], [{}], [{ pool: 'p', image: 'i' }]]) assert.equal(childRows(value), undefined)
console.log('RBD child dependencies retain namespace, trash membership and unresolved names')
const deleteReasonNode = childTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'rbdSnapshotDeleteReason')
const deleteReasonCode = ts.transpileModule(deleteReasonNode.getText(childTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const snapshotDeleteReason = new Function(`${deleteReasonCode}; return rbdSnapshotDeleteReason`)()
assert.equal(snapshotDeleteReason({ is_protected: false, children: [] }), undefined)
assert.match(snapshotDeleteReason({ is_protected: true, children: [] }), /取消快照保护/)
for (const is_protected of [undefined, null, 'false', 0]) assert.match(snapshotDeleteReason({ is_protected, children: [] }), /保护状态未知/)
for (const children of [undefined, null, {}, '']) assert.match(snapshotDeleteReason({ is_protected: false, children }), /依赖信息不可用/)
for (const children of [[{ trash: true }], [{}], [null]]) assert.match(snapshotDeleteReason({ is_protected: false, children }), /仍有子镜像/)

const usageExports = {}
const usageCode = ts.transpileModule(readFileSync(new URL('../src/pages/block/rbdUsage.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText
new Function('exports', usageCode)(usageExports)
const usageText = usageExports.rbdUsageText
assert.equal(usageText(0, ['fast-diff']), '0')
assert.equal(usageText(4096, ['layering', 'fast-diff']), '4096')
assert.equal(usageText(4096, ['layering']), '不可用：未启用 fast-diff')
for (const features of [null, undefined, 'fast-diff', [null]]) assert.equal(usageText(0, features), '特性信息不可用，无法确认用量采集条件')
for (const value of [undefined, null, -1, 0.5, NaN, Infinity, Number.MAX_SAFE_INTEGER + 1, '123']) assert.equal(usageText(value, ['fast-diff']), '用量未返回或无效，请重新采集')
console.log('RBD usage distinguishes missing fast-diff, unavailable statistics and valid zero')

const blockSource = readFileSync(new URL('../src/pages/block/pages.tsx', import.meta.url), 'utf8')
assert.ok(blockSource.includes("key: 'image_created_at', title: '镜像创建时间（命令原值）'"))
const resourceApiSource = readFileSync(new URL('../src/api/resource.ts', import.meta.url), 'utf8')
const resourceApiTree = ts.createSourceFile('resource.ts', resourceApiSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS)
const listResourceNode = resourceApiTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'listResource')
const listResourceCode = ts.transpileModule(listResourceNode.getText(resourceApiTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const readImages = new Function('request', 'jsonInit', 'toRecord', `${listResourceCode}; return listResource`)(
  async () => ({ items: [{ created_at: 'inventory-time', data: { image_created_at: 'native-image-time' } }, { created_at: 'inventory-only', data: {} }] }),
  (method, body) => ({ method, body }), (value) => value
)
const imageTimes = (await readImages('/rbd/images', 7)).items
assert.equal(imageTimes[0].created_at, 'inventory-time')
assert.equal(imageTimes[0].image_created_at, 'native-image-time')
assert.equal(imageTimes[1].image_created_at, undefined)
console.log('Native image creation time remains separate from inventory creation time')
assert.ok(blockSource.includes('disabledWhen: rbdSnapshotDeleteReason'))
assert.ok(blockSource.includes('disabledWhen: rbdTrashRestoreReason'))
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
let globalScheduleNode
function findGlobalSchedule(node) {
  if (ts.isObjectLiteralExpression(node) && node.properties.some((property) => ts.isPropertyAssignment(property) && property.name.getText(blockTree) === 'path' && property.initializer.getText(blockTree) === "'/rbd/mirroring/global/schedule'")) globalScheduleNode = node
  ts.forEachChild(node, findGlobalSchedule)
}
findGlobalSchedule(blockTree)
const globalScheduleCode = ts.transpileModule(`const action = ${globalScheduleNode.getText(blockTree)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const globalSchedule = new Function(`${globalScheduleCode}; return action`)()
assert.deepEqual(globalSchedule.buildBody(addValues, 7, poolRow), { cluster_id: 7, action: 'mirror-schedule-add', interval: '12h', start_time: '01:00' })
assert.deepEqual(globalSchedule.buildBody(removeValues, 7, poolRow), { cluster_id: 7, action: 'mirror-schedule-remove' })
assert.deepEqual(globalSchedule.buildBody({ ...removeValues, remove_interval: '1d' }, 7), { cluster_id: 7, action: 'mirror-schedule-remove', interval: '1d', start_time: '01:00' })
assert.match(globalSchedule.confirmation(addValues), /集群全局/)
assert.match(globalSchedule.confirmation(removeValues), /全部全局/)
assert.equal(globalSchedule.fields.find((field) => field.name === 'start_time').visibleWhen(removeValues), false)
assert.ok(source.includes('definition.toolbarActions?.map'))
assert.ok(source.includes('onClick={() => openForm(action)}'))
assert.ok(source.includes('currentClusterId.current !== formClusterId || clusterGeneration.current !== generation'))
console.log('Global schedule toolbar action preserves cluster scope and explicit confirmation')

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
const emptyScopes = [
  { pool: '-', namespace: '-', image: '-', items: [] },
  { pool: 'images', namespace: '-', image: '-', items: [] },
  { pool: 'images', namespace: 'team', image: '-', items: [] },
  { pool: 'images', namespace: '', image: 'vm', items: [] }
]
assert.deepEqual(mirrorRows(emptyScopes).map((row) => row.scope), ['集群', '池', '命名空间', '镜像'])
assert.ok(mirrorRows(emptyScopes).every((row) => row.interval === '此范围未返回调度项'))
assert.equal(mirrorRows([emptyScopes[0], { ...emptyScopes[1], items }]).length, 3)
assert.equal(new Set(mirrorRows(emptyScopes).map((row) => row.key)).size, 4)
const duplicateScope = { pool: '-', namespace: '-', image: '-', items }
assert.equal(mirrorRows([duplicateScope, { ...duplicateScope, items: [] }]), undefined)
assert.equal(mirrorRows([{ pool: 'images', image: 'vm', items }]), undefined)
assert.equal(mirrorRows([{ pool: 'images', namespace: '-', image: 'vm', items }]), undefined)
for (const value of [null, {}, [null], [{}], [{ pool: 'images', namespace: '-', image: '-', items: null }], [{ pool: '-', namespace: 'team', image: '-', items }], [{ pool: 'images', namespace: '-', image: '-', items: [{}] }]]) assert.equal(mirrorRows(value), undefined)
console.log('Mirror schedule scope display checks passed')
const liveSource = readFileSync(new URL('../src/pages/block/LiveMirrorSchedules.tsx', import.meta.url), 'utf8')
const liveTree = ts.createSourceFile('LiveMirrorSchedules.tsx', liveSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const liveFn = liveTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'LiveMirrorSchedules')
const liveLoader = liveFn.body.statements.find((node) => ts.isVariableStatement(node) && node.declarationList.declarations[0].name.getText(liveTree) === 'loader').declarationList.declarations[0].initializer.arguments[0]
const liveCode = ts.transpileModule(`const loader = ${liveLoader.getText(liveTree)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const loadLive = (payload) => new Function('request', 'jsonInit', 'mirrorScheduleRows', 'clusterId', `${liveCode}; return loader`)(async (path, init) => {
  assert.equal(path, '/rbd/mirroring/schedules'); assert.deepEqual(init, { method: 'GET', body: { cluster_id: 9 } }); return payload
}, (method, body) => ({ method, body }), mirrorRows, 9)()
const liveData = { schedules: [{ pool: '-', namespace: '-', image: '-', items }], observed_at: '2026-10-03T00:00:00Z' }
assert.deepEqual(await loadLive(liveData), liveData)
assert.deepEqual((await loadLive({ ...liveData, schedules: [] })).schedules, [])
for (const payload of [null, {}, { ...liveData, schedules: null }, { ...liveData, schedules: [duplicateScope, duplicateScope] }, { ...liveData, observed_at: null }]) await assert.rejects(loadLive(payload))
assert.ok(blockSource.includes('<LiveMirrorSchedules key={`${selectedClusterId}/${scheduleRevision}`} clusterId={selectedClusterId} />'))
assert.ok(blockSource.includes('onFormMutationSuccess={() => setScheduleRevision((value) => value + 1)}'))
assert.ok(liveSource.includes('下方为上次成功读取结果，不代表当前配置。'))
console.log('Live mirror schedules preserve cluster scope, empty results and read failures')
const statusSource = readFileSync(new URL('../src/pages/block/LiveMirrorScheduleStatus.tsx', import.meta.url), 'utf8')
const statusTree = ts.createSourceFile('status.tsx', statusSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const statusRowsNode = statusTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'scheduledImageRows')
const statusRowsCode = ts.transpileModule(statusRowsNode.getText(statusTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const statusRows = new Function(`${statusRowsCode}; return scheduledImageRows`)()
assert.deepEqual(statusRows([]), [])
const pending = { image: 'images/team/vm', schedule_time: '2026-10-03 12:00:00' }
assert.deepEqual(statusRows([pending, pending]), [{ key: 0, ...pending }, { key: 1, ...pending }])
for (const value of [null, {}, [null], [{}], [{ ...pending, image: ' ' }], [{ ...pending, schedule_time: null }]]) assert.equal(statusRows(value), undefined)
const statusFn = statusTree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'LiveMirrorScheduleStatus')
const statusLoaderNode = statusFn.body.statements.find((node) => ts.isVariableStatement(node) && node.declarationList.declarations[0].name.getText(statusTree) === 'loader').declarationList.declarations[0].initializer.arguments[0]
const statusLoaderCode = ts.transpileModule(`const loader = ${statusLoaderNode.getText(statusTree)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const loadStatus = (payload) => new Function('request', 'jsonInit', 'clusterId', 'scheduledImageRows', `${statusLoaderCode}; return loader()`)(async (path, init) => {
  assert.equal(path, '/rbd/mirroring/schedule/status'); assert.deepEqual(init, { method: 'GET', body: { cluster_id: 9 } }); return payload
}, (method, body) => ({ method, body }), 9, statusRows)
assert.deepEqual(await loadStatus({ scheduled_images: [pending], observed_at: 'now' }), { rows: [{ key: 0, ...pending }], observedAt: 'now' })
for (const payload of [{}, { scheduled_images: null, observed_at: 'now' }, { scheduled_images: [], observed_at: null }]) await assert.rejects(loadStatus(payload))
assert.ok(blockSource.includes('key={`status/${selectedClusterId}/${scheduleRevision}`}'))
console.log('Live pending mirror tasks preserve native times, repeated images and isolated scope')
const resourceFn = tree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'ResourceListPage')
const deleteNode = resourceFn.body.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'deleteRow')
const deleteCode = ts.transpileModule(deleteNode.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const risk of ['high', 'medium']) for (const phase of ['current', 'confirm-switch', 'confirm-return', 'unmount', 'result-switch', 'refresh-switch']) {
  let confirmation, writes = 0, notices = 0, refreshes = 0
  const currentClusterId = { current: 7 }, clusterGeneration = { current: 0 }
  const switchScope = () => { currentClusterId.current = 9; clusterGeneration.current++ }
  const env = {
    selectedClusterId: 7, currentClusterId, clusterGeneration, mutationBlocked: false,
    definition: { deleteAction: { risk, path: '/rbd/image/snapshot', resourceKey: () => 'snapshot', buildBody: () => ({ cluster_id: 7 }) } },
    message: { success() { notices++ }, warning() {}, error() {} },
    Modal: { confirm(options) { confirmation = options } },
    operationMutation: { run: async (fn) => { const result = await fn(); if (phase === 'result-switch') switchScope(); return result } },
    mutateResource: async (path, method, body, options) => {
      writes++; assert.equal(method, 'DELETE'); assert.equal(body.cluster_id, 7)
      assert.deepEqual(options, risk === 'high' ? { ifMatch: 3 } : undefined)
    },
    refreshResource: async () => { if (phase === 'refresh-switch') switchScope() },
    refresh: async () => { refreshes++ }
  }
  const remove = new Function(...Object.keys(env), `${deleteCode}; return deleteRow`)(...Object.values(env))
  await remove({ resource_version: 3 })
  if (phase === 'confirm-switch' || phase === 'confirm-return') switchScope()
  if (phase === 'confirm-return') { currentClusterId.current = 7; clusterGeneration.current++ }
  if (phase === 'unmount') clusterGeneration.current++
  await confirmation.onOk()
  const blocked = ['confirm-switch', 'confirm-return', 'unmount'].includes(phase)
  assert.equal(writes, blocked ? 0 : 1)
  assert.equal(notices, blocked || phase === 'result-switch' ? 0 : 1)
  assert.equal(refreshes, phase === 'current' ? 1 : 0)
}
assert.ok(source.includes('useEffect(() => () => { clusterGeneration.current += 1 }, [])'))
console.log('Resource deletion confirmations and completions remain bound to the original scope')
const submitNode = resourceFn.body.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'submitForm')
const submitCode = ts.transpileModule(submitNode.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const mode of ['success', 'cancel', 'switch-confirm', 'switch-result', 'failed', 'refresh-failed']) {
  const currentClusterId = { current: 7 }
  const clusterGeneration = { current: 0 }
  let writes = 0, notified = 0
  const env = {
    selectedClusterId: 7, formClusterId: 7, submitting: false, mutationBlocked: false, currentClusterId, clusterGeneration,
    activeRow: undefined, activeAction: { path: '/rbd/mirroring/global/schedule', method: 'POST', confirmation: () => 'confirm', buildBody: () => ({ cluster_id: 7 }) },
    setSubmitting() {}, closeForm() {}, message: { success() {} }, form: {},
    Modal: { confirm(options) { if (mode === 'switch-confirm') currentClusterId.current = 9; if (mode === 'cancel') options.onCancel(); else options.onOk() } },
    operationMutation: { run: async (fn) => { const result = await fn(); if (mode === 'switch-result') { currentClusterId.current = 9; clusterGeneration.current++ } return result } },
    mutateResource: async () => { writes++; if (mode === 'failed') throw new Error('operation failed'); return {} },
    onFormMutationSuccess: () => { notified++ },
    refreshResource: async () => { if (mode === 'refresh-failed') throw new Error('inventory unavailable') }, refresh: async () => {},
  }
  const submit = new Function(...Object.keys(env), `${submitCode}; return submitForm`)(...Object.values(env))
  if (mode === 'failed' || mode === 'refresh-failed') await assert.rejects(submit({})); else await submit({})
  assert.equal(writes, mode === 'cancel' || mode === 'switch-confirm' ? 0 : 1)
  assert.equal(notified, mode === 'success' || mode === 'refresh-failed' ? 1 : 0)
}
console.log('Successful current-cluster mutations invalidate live schedules before inventory refresh')

const imageScheduleDetails = mirrorExports.imageMirrorScheduleDetails
for (const [origin, label] of [['cluster', '继承集群'], ['pool', '继承池'], ['namespace', '继承命名空间'], ['', '镜像专属']]) {
  const details = imageScheduleDetails({ name: origin === 'cluster' ? '' : 'images/team/vm', inherited_from: origin, schedule_status: 'available', schedule_time: '2026-10-03 12:30:00', schedule_interval: items })
  assert.equal(details.origin, label)
  assert.equal(details.target, origin === 'cluster' ? '集群' : 'images/team/vm')
  assert.equal(details.nextRun, '2026-10-03 12:30:00')
  assert.equal(details.intervals[0].startTime, '01:00:00+08:00')
  assert.equal(details.intervals[1].startTime, '未指定')
}
assert.equal(imageScheduleDetails({ name: 'images/vm', schedule_interval: items }).origin, '镜像专属')
assert.equal(imageScheduleDetails({ name: 'images/vm', schedule_interval: items }).nextRun, '运行状态不可用或尚未采集')
assert.equal(imageScheduleDetails({ name: 'images/vm', schedule_interval: items, schedule_status: 'available' }).nextRun, '本次采集未返回此镜像的待执行任务')
assert.equal(imageScheduleDetails({ name: 'images/vm', schedule_interval: items, schedule_status: 'unavailable', schedule_time: 'old' }).nextRun, '运行状态不可用或尚未采集')
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

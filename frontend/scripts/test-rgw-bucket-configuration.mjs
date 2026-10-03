import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
import './test-external-form-confirmation.mjs'
import './test-rgw-bucket-tag-form.mjs'
import './test-rgw-bucket-lifecycle.mjs'
import './test-rgw-bucket-lifecycle-form.mjs'
import './test-rgw-bucket-acl.mjs'
import './test-rgw-bucket-replication.mjs'
import './test-rgw-bucket-notifications.mjs'
import './test-rgw-topics.mjs'
const helpers = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketNotificationForm.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(helpers)
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketNotificationDelete.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(helpers)
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketReplicationForm.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(helpers)
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketAclForm.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(helpers)
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketObjectLockForm.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(helpers)
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketObjectLockSummary.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(helpers)
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketLifecycleForm.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(helpers)
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketCorsForm.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(helpers)
const cors = {}
new Function('exports', 'require', ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwBucketCorsRules.tsx', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText)(cors, (name) => name === 'antd' ? { Table: 'Table' } : { jsx: (type, props) => ({ type, props }), jsxs: (type, props) => ({ type, props }) })
const corsRule = { id: '<rule>', allowed_origins: ['*'], allowed_methods: ['GET'], allowed_headers: [], expose_headers: ['b', 'a', 'b'], max_age_seconds: 0 }
assert.deepEqual(cors.bucketCorsRows([corsRule])[0].expose_headers, ['b', 'a', 'b'])
for (const value of [undefined, {}, [null], [{ ...corsRule, max_age_seconds: -1 }], [{ ...corsRule, allowed_origins: [1] }]]) assert.equal(cors.bucketCorsRows(value), undefined)
assert.equal(cors.RgwBucketCorsRules({ value: [], configured: false }).props.children, '未配置 CORS')
assert.equal(cors.RgwBucketCorsRules({ value: [], configured: true }).props.children, 'CORS 数据不可用')
const corsTable = cors.RgwBucketCorsRules({ value: [corsRule], configured: true })
const ageColumn = corsTable.props.columns.find(column => column.dataIndex === 'max_age_seconds')
assert.equal(ageColumn.render(0), '0 秒')
assert.equal(ageColumn.render(null), '未设置')
assert.equal(corsTable.props.columns.find(column => column.dataIndex === 'id').render('<rule>'), '"<rule>"')
const deletion = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketDelete.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(deletion)
const deleteRow = { natural_key: 'dGVhbQBidWNrZXQ', name: 'bucket', tenant: 'team' }
assert.deepEqual(deletion.bucketDeleteInput(deleteRow), { bucket_id: deleteRow.natural_key })
assert.equal(deletion.bucketDeleteBlocked(deleteRow), undefined)
assert.match(deletion.bucketDeleteConfirmation(deleteRow), /team.*dGVhbQBidWNrZXQ.*不会清空对象、历史版本或删除标记/)
assert.match(deletion.bucketDeleteConfirmation({ ...deleteRow, tenant: '' }), /全局租户/)
for (const id of ['', ' id ', 12, null, 'a/b', 'a:b']) {
  assert.ok(deletion.bucketDeleteBlocked({ natural_key: id }))
  assert.throws(() => deletion.bucketDeleteInput({ natural_key: id }))
}
assert.deepEqual(deletion.bucketDeleteInput({ bucket_id: 'AGJ1Y2tldA' }), { bucket_id: 'AGJ1Y2tldA' })
const versioning = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketVersioningForm.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(versioning)
for (const value of [undefined, 'off', 'future', '', null]) assert.equal(versioning.bucketVersioningInitial({ versioning: value }).versioning, undefined)
for (const value of ['enabled', 'suspended']) {
  const values = versioning.bucketVersioningInitial({ versioning: value })
  assert.deepEqual(versioning.bucketVersioningInput(values, 'AGJ1Y2tldA'), { bucket_id: 'AGJ1Y2tldA', versioning: value })
  assert.match(versioning.bucketVersioningConfirmation(values, 'AGJ1Y2tldA'), /AGJ1Y2tldA.*MFA.*回读核验/)
}
for (const value of [undefined, null, 'off', true, 'Enabled']) assert.throws(() => versioning.bucketVersioningInput({ versioning: value }, 'AGJ1Y2tldA'))
assert.throws(() => versioning.bucketVersioningInput({ versioning: 'enabled' }, ''))
assert.match(versioning.bucketVersioningConfirmation({ versioning: 'suspended' }, 'AGJ1Y2tldA'), /可能覆盖 null 版本/)
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketEncryptionForm.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(helpers)
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketEncryptionSummary.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(helpers)
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketConfiguration.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(helpers)
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwBucketTagForm.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(helpers)
for (const [kind, document] of [['policy', '{"Statement":[],"large":9007199254740993}'], ['cors', '<CORSConfiguration/>'], ['lifecycle', '<LifecycleConfiguration/>'], ['encryption', '<ServerSideEncryptionConfiguration/>'], ['tagging', '<Tagging><TagSet/></Tagging>']]) {
  const input = { bucket_id: 'AGJ1Y2tldA', kind, document }
  assert.deepEqual(helpers.rgwBucketConfigurationInput(input), input)
}
for (const values of [{ kind: 'cors', document: '{}' }, { kind: 'policy', document: 'null' }, { kind: 'policy', document: '[]' }, { kind: 'policy', document: '{' }, { kind: 'unknown', document: '{}' }, { kind: 'policy', document: '' }]) assert.throws(() => helpers.rgwBucketConfigurationInput({ bucket_id: 'id', ...values }))
assert.throws(() => helpers.rgwBucketConfigurationInput({ bucket_id: 'id', kind: 'cors', document: '<CORSConfiguration>' + 'x'.repeat(1024 * 1024) + '</CORSConfiguration>' }))
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
const source = ts.createSourceFile('pages.tsx', pages, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
let definition
function visit(node) {
  if (ts.isObjectLiteralExpression(node) && node.properties.some(item => ts.isPropertyAssignment(item) && item.name.getText(source) === 'title' && item.initializer.text === 'Bucket 配置文档')) {
    const code = ts.transpileModule(`const definition = ${node.getText(source)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.React } }).outputText
    definition = new Function(...Object.keys(helpers), `${code}; return definition`)(...Object.values(helpers))
  }
  ts.forEachChild(node, visit)
}
visit(source)
assert.ok(definition)
const aclAction = definition.extraActions.find(action => action.title === '替换 Bucket ACL')
const aclRow = { bucket_id: 'AGJ1Y2tldA', kind: 'acl', configured: true, acl: { owner: { id: 'owner' }, grants: [] } }
const aclInitial = aclAction.initialValues(aclRow)
assert.equal(aclInitial.acl, undefined)
assert.equal(aclInitial.confirm_replace, undefined)
assert.equal(aclAction.path, '/rgw/bucket/acl')
assert.equal(aclAction.method, 'PATCH')
assert.ok(aclAction.visibleWhen(aclRow))
assert.ok(!aclAction.visibleWhen({ kind: 'policy' }))
assert.ok(aclAction.disabledWhen({ ...aclRow, configured: false }))
assert.ok(aclAction.disabledWhen({ ...aclRow, acl: {} }))
assert.throws(() => aclAction.buildBody(aclInitial, 7, aclRow))
for (const acl of ['private', 'public-read', 'public-read-write', 'authenticated-read']) {
  const values = { ...aclInitial, acl, confirm_replace: 'acknowledged' }
  assert.deepEqual(aclAction.buildBody(values, 7, aclRow), { cluster_id: 7, bucket_id: aclRow.bucket_id, acl })
  assert.match(aclAction.confirmation(values, aclRow), /AGJ1Y2tldA.*现有自定义授权将被移除/)
  assert.match(aclAction.confirmation(values, aclRow), /private 也不保证最终私有/)
  for (const change of [{ bucket_id: 'other' }, { acl: 'future' }, { acl: '' }, { confirm_replace: true }]) assert.throws(() => aclAction.buildBody({ ...values, ...change }, 7, aclRow))
}
const objectLockAction = definition.extraActions.find(action => action.title === '修改对象锁默认保留')
const objectLockRow = { bucket_id: 'AGJ1Y2tldA', kind: 'object-lock', configured: false, object_lock: null }
assert.ok(objectLockAction.visibleWhen(objectLockRow))
assert.equal(objectLockAction.disabledWhen(objectLockRow), undefined)
const objectLockValues = { ...objectLockAction.initialValues(objectLockRow), retention_action: 'set', mode: 'GOVERNANCE', period: '030', unit: 'Days', confirm_lock: 'acknowledged' }
assert.equal(objectLockAction.initialValues(objectLockRow).confirm_lock, undefined)
assert.equal(objectLockAction.initialValues(objectLockRow).retention_action, undefined)
assert.match(objectLockAction.buildBody(objectLockValues, 7, objectLockRow).document, /<Days>30<\/Days>/)
assert.match(objectLockAction.confirmation(objectLockValues, objectLockRow), /无法关闭.*不会解除已有对象.*Legal Hold/)
assert.ok(!objectLockAction.buildBody({ ...objectLockValues, retention_action: 'clear' }, 7, objectLockRow).document.includes('<Rule>'))
for (const patch of [{ period: '0' }, { period: '2147483648' }, { period: '1x' }, { unit: 'days' }, { mode: 'future' }, { confirm_lock: undefined }, { bucket_id: 'other' }, { kind: 'policy' }]) assert.throws(() => objectLockAction.buildBody({ ...objectLockValues, ...patch }, 7, objectLockRow))
assert.equal(helpers.objectLockFormInitial({ ...objectLockRow, configured: true, object_lock: { enabled: true, default_retention: { mode: 'COMPLIANCE', days: null, years: '2' } } }).unit, 'Years')
const lifecycleAction = definition.extraActions.find(action => action.title === '编辑生命周期规则')
const lifecycleRow = { bucket_id: 'AGJ1Y2tldA', kind: 'lifecycle', configured: false, lifecycle_rules: [] }
assert.ok(lifecycleAction.visibleWhen(lifecycleRow))
assert.ok(!lifecycleAction.visibleWhen({ kind: 'cors' }))
assert.equal(lifecycleAction.disabledWhen(lifecycleRow), undefined)
assert.deepEqual(lifecycleAction.initialValues(lifecycleRow).lifecycle_draft, { rules: [] })
assert.equal(typeof lifecycleAction.fields.find(field => field.name === 'lifecycle_draft').renderControl, 'function')
assert.throws(() => lifecycleAction.buildBody(lifecycleAction.initialValues(lifecycleRow), 7, lifecycleRow))
const corsAction = definition.extraActions.find(action => action.title === '编辑 CORS 规则')
const corsRow = { bucket_id: 'AGJ1Y2tldA', kind: 'cors', configured: true, cors_rules: [corsRule] }
const corsValues = corsAction.initialValues(corsRow)
corsValues.cors_draft.rules[0].expose_headers.push('extra')
assert.deepEqual(corsRow.cors_rules[0].expose_headers, ['b', 'a', 'b'])
assert.ok(corsAction.visibleWhen(corsRow))
assert.ok(!corsAction.visibleWhen({ kind: 'policy' }))
assert.equal(corsAction.disabledWhen(corsRow), undefined)
assert.equal(corsAction.method, 'PATCH')
assert.equal(corsAction.path, '/rgw/bucket/policy')
assert.match(corsAction.buildBody(corsValues, 7, corsRow).document, /<ID>&lt;rule&gt;<\/ID>/)
assert.match(corsAction.confirmation(corsValues, corsRow), /整体替换.*不替代访问权限策略/)
assert.throws(() => corsAction.buildBody({ ...corsValues, bucket_id: 'different' }, 7, corsRow))
for (const patch of [{ allowed_origins: [] }, { allowed_origins: ['**'] }, { allowed_headers: [''] }, { allowed_methods: ['PATCH'] }, { id: 'é'.repeat(128) }, { id: '\u0000' }, { max_age_seconds: -1 }]) assert.throws(() => helpers.corsDocument({ rules: [{ ...corsRule, ...patch }] }))
assert.throws(() => helpers.corsDocument({ rules: [] }))
assert.match(helpers.corsDocument({ rules: [{ ...corsRule, id: '\r<&', max_age_seconds: null }] }), /<ID>&#13;&lt;&amp;<\/ID>/)
assert.ok(!helpers.corsDocument({ rules: [{ ...corsRule, max_age_seconds: null }] }).includes('MaxAgeSeconds'))
const editor = {}
new Function('exports', 'require', ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwBucketCorsEditor.tsx', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText)(editor, name => name === './rgwBucketCorsForm' ? helpers : name === 'antd' ? { Alert: 'Alert', Button: 'Button', Input: { TextArea: 'TextArea' }, InputNumber: 'InputNumber', Select: 'Select', Space: 'Space' } : { jsx: (type, props) => ({ type, props }), jsxs: (type, props) => ({ type, props }) })
function flattenCors(node) { return Array.isArray(node) ? node.flatMap(flattenCors) : node && typeof node === 'object' ? [node, ...flattenCors(node.props?.children)] : [] }
let edited
const controls = flattenCors(editor.RgwBucketCorsEditor({ value: { rules: [corsRule] }, onChange: value => { edited = value } }))
controls.find(node => node.type === 'Button' && node.props.children === '添加规则').props.onClick()
assert.equal(edited.rules.length, 2)
controls.find(node => node.type === 'TextArea' && node.props['aria-label'] === '规则 1 ID').props.onChange({ target: { value: ' changed ' } })
assert.equal(edited.rules[0].id, ' changed ')
assert.equal(corsRule.id, '<rule>')
const twoRules = flattenCors(editor.RgwBucketCorsEditor({ value: { rules: [corsRule, { ...corsRule, id: 'second' }] }, onChange: value => { edited = value } }))
twoRules.find(node => node.type === 'Button' && node.props.children === '下移').props.onClick()
assert.deepEqual(edited.rules.map(rule => rule.id), ['second', '<rule>'])
twoRules.find(node => node.type === 'Button' && node.props.children === '移除规则').props.onClick()
assert.deepEqual(edited.rules.map(rule => rule.id), ['second'])
const newCorsRow = { ...corsRow, configured: false, cors_rules: [] }
assert.equal(corsAction.initialValues(newCorsRow).cors_draft.rules.length, 0)
assert.equal(corsAction.disabledWhen(newCorsRow), undefined)
edited = undefined
const locked = flattenCors(editor.RgwBucketCorsEditor({ value: { rules: [corsRule] }, disabled: true, onChange: value => { edited = value } }))
for (const node of locked.filter(node => node.type === 'Button')) node.props.onClick()
assert.equal(edited, undefined)
const encryptionAction = definition.extraActions.find(action => action.title === '编辑 Bucket 默认加密')
const encryptionRow = { bucket_id: 'AGJ1Y2tldA', kind: 'encryption', configured: true, encryption: { rule_exists: true, algorithm: 'aws:kms', kms_master_key_id: ' key<&\r\n😀 ', bucket_key_enabled: true } }
const encryptionValues = encryptionAction.initialValues(encryptionRow)
assert.equal(encryptionValues.kms_master_key_id, encryptionRow.encryption.kms_master_key_id)
assert.equal(encryptionAction.disabledWhen(encryptionRow), undefined)
assert.ok(encryptionAction.visibleWhen(encryptionRow))
assert.ok(!encryptionAction.visibleWhen({ kind: 'policy' }))
assert.equal(encryptionAction.method, 'PATCH')
assert.equal(encryptionAction.path, '/rgw/bucket/policy')
const encryptionBody = encryptionAction.buildBody(encryptionValues, 7, encryptionRow)
assert.equal(encryptionBody.cluster_id, 7)
assert.match(encryptionBody.document, /<KMSMasterKeyID> key&lt;&amp;&#13;\n😀 <\/KMSMasterKeyID>/)
assert.match(encryptionBody.document, /<BucketKeyEnabled>true<\/BucketKeyEnabled>/)
assert.match(encryptionAction.confirmation(encryptionValues, encryptionRow), /不会重新加密已有对象/)
assert.ok(!encryptionAction.confirmation(encryptionValues, encryptionRow).includes('key<&'))
for (const change of [{ bucket_id: 'other' }, { kind: 'policy' }, { algorithm: 'other' }, { kms_master_key_id: '' }, { kms_master_key_id: '\u0000' }, { kms_master_key_id: '\ud800' }, { bucket_key_enabled: true }, { algorithm: 'AES256' }]) assert.throws(() => encryptionAction.buildBody({ ...encryptionValues, ...change }, 7, encryptionRow))
const aesBody = encryptionAction.buildBody({ ...encryptionValues, algorithm: 'AES256', kms_master_key_id: '', bucket_key_enabled: 'false' }, 7, encryptionRow)
assert.ok(!aesBody.document.includes('KMSMasterKeyID'))
assert.match(aesBody.document, /<SSEAlgorithm>AES256<\/SSEAlgorithm>/)
assert.ok(encryptionAction.disabledWhen({ ...encryptionRow, encryption: { ...encryptionRow.encryption, algorithm: 'future' } }))
assert.ok(encryptionAction.disabledWhen({ ...encryptionRow, configured: undefined }))
const fresh = encryptionAction.initialValues({ ...encryptionRow, configured: false, encryption: null })
assert.equal(fresh.algorithm, undefined)
assert.throws(() => encryptionAction.buildBody(fresh, 7, { ...encryptionRow, configured: false, encryption: null }))
const tagAction = definition.extraActions.find(action => action.title === '逐条编辑 Bucket 标签')
assert.ok(tagAction.visibleWhen({ kind: 'tagging' }))
assert.ok(!tagAction.visibleWhen({ kind: 'policy' }))
const tagRow = { bucket_id: 'AGJ1Y2tldA', kind: 'tagging', configured: true, tags: [{ key: 'a', value: '' }] }
const tagValues = tagAction.initialValues(tagRow)
assert.equal(tagAction.disabledWhen(tagRow), undefined)
assert.deepEqual(tagAction.buildBody(tagValues, 7, tagRow), { cluster_id: 7, ...helpers.bucketTagFormInput(tagValues, tagRow) })
assert.equal(tagAction.path, '/rgw/bucket/policy')
assert.equal(tagAction.method, 'PATCH')
assert.ok(tagAction.confirmation(tagValues, tagRow).includes('整体替换'))
assert.equal(typeof tagAction.fields.find(field => field.name === 'tag_set').renderControl, 'function')
assert.equal(definition.buildQuery({ kind: 'cors' }).toString(), 'kind=cors')
assert.deepEqual(definition.filterFields.find(field => field.name === 'kind').options, helpers.rgwBucketConfigurationReadOptions)
assert.deepEqual(definition.columns.map(column => column.key), ['bucket_id', 'kind', 'configured', 'tags', 'encryption', 'object_lock', 'acl', 'replication', 'notifications', 'cors_rules', 'lifecycle_rules', 'content_type', 'document'])
const readonlyReplication = { bucket_id: 'AGJ1Y2tldA', kind: 'replication', configured: true, document: '<ReplicationConfiguration/>' }
assert.ok(helpers.rgwBucketConfigurationEditBlocked(readonlyReplication))
assert.ok(helpers.rgwBucketConfigurationDeleteBlocked(readonlyReplication))
const deletableReplication = { ...readonlyReplication, replication: { role: '', rules: [{ id: 'r', status: 'Enabled', priority: '1', destination_bucket: 'target' }] } }
assert.equal(helpers.rgwBucketConfigurationDeleteBlocked(deletableReplication), undefined)
assert.deepEqual(helpers.rgwBucketConfigurationDeleteInput(deletableReplication), { bucket_id: 'AGJ1Y2tldA', kind: 'replication' })
assert.match(helpers.rgwBucketConfigurationDeleteConfirmation(deletableReplication), /移除全部 S3 复制规则.*不会关闭其他桶本地或 Zonegroup 同步策略/)
assert.ok(helpers.rgwBucketConfigurationDeleteBlocked({ ...deletableReplication, replication: { rules: [] } }))
assert.ok(helpers.rgwBucketConfigurationEditBlocked(deletableReplication))
assert.ok(helpers.rgwBucketConfigurationReadOptions.some(option => option.value === 'replication'))
const readonlyAcl = { bucket_id: 'AGJ1Y2tldA', kind: 'acl', configured: true, document: '<AccessControlPolicy/>' }
assert.ok(helpers.rgwBucketConfigurationEditBlocked(readonlyAcl))
assert.ok(helpers.rgwBucketConfigurationDeleteBlocked(readonlyAcl))
assert.ok(helpers.rgwBucketConfigurationReadOptions.some(option => option.value === 'acl'))
assert.ok(!helpers.rgwBucketConfigurationOptions.some(option => option.value === 'acl'))
const lockSummary = definition.columns.find(column => column.key === 'object_lock').render
assert.match(lockSummary(null, { kind: 'object-lock', configured: false }), /未启用/)
assert.match(lockSummary({ enabled: true, default_retention: null }, { kind: 'object-lock', configured: true }), /未设置默认保留期/)
for (const mode of ['GOVERNANCE', 'COMPLIANCE', 'FUTURE']) assert.ok(lockSummary({ enabled: true, default_retention: { mode, days: '30', years: null } }, { kind: 'object-lock', configured: true }).includes(mode))
assert.match(lockSummary({ enabled: true, default_retention: { mode: 'COMPLIANCE', days: null, years: '2' } }, { kind: 'object-lock', configured: true }), /2 年/)
assert.match(lockSummary({ enabled: true, default_retention: { mode: 'COMPLIANCE', days: '0', years: null } }, { kind: 'object-lock', configured: true }), /无效期限/)
assert.match(lockSummary(undefined, { kind: 'object-lock', configured: true }), /不可用/)
assert.equal(lockSummary(null, { kind: 'policy' }), '—')
const readonlyLock = { bucket_id: 'AGJ1Y2tldA', kind: 'object-lock', configured: true, document: '<ObjectLockConfiguration/>' }
assert.ok(helpers.rgwBucketConfigurationEditBlocked(readonlyLock))
assert.ok(helpers.rgwBucketConfigurationDeleteBlocked(readonlyLock))
assert.throws(() => helpers.rgwBucketConfigurationInput(readonlyLock))
const encryption = definition.columns.find(column => column.key === 'encryption').render
assert.equal(encryption(null, { kind: 'policy' }), '—')
assert.match(encryption(null, { kind: 'encryption', configured: false }), /未设置/)
for (const value of [undefined, null, {}, { rule_exists: true }]) assert.match(encryption(value, { kind: 'encryption', configured: true }), /不可用/)
const config = { rule_exists: false, algorithm: '', kms_master_key_id: '', bucket_key_enabled: false }
assert.match(encryption(config, { kind: 'encryption', configured: true }), /没有默认加密规则/)
assert.match(encryption({ ...config, rule_exists: true }, { kind: 'encryption', configured: true }), /算法：.*空/)
assert.match(encryption({ ...config, rule_exists: true, algorithm: 'future:algorithm', kms_master_key_id: ' key ', bucket_key_enabled: true }, { kind: 'encryption', configured: true }), /future:algorithm.*" key ".*启用.*不代表已有对象/)
const status = definition.columns.find(column => column.key === 'configured').render
assert.equal(status(true), '已配置')
assert.equal(status(false), '未配置')
for (const value of [undefined, null, 0, 1, 'false']) assert.equal(status(value), '状态不可用')
const input = { bucket_id: 'id', kind: 'cors', document: '<CORSConfiguration/>' }
assert.deepEqual(definition.createAction.buildBody(input, 7), { cluster_id: 7, ...input })
assert.equal(definition.createAction.method, 'PATCH')
assert.equal(definition.createAction.path, '/rgw/bucket/policy')
assert.equal(definition.updateAction.path, '/rgw/bucket/policy')
assert.equal(definition.updateAction.method, 'PATCH')
for (const kind of ['policy', 'cors', 'lifecycle', 'encryption', 'tagging']) {
  const document = kind === 'policy' ? '{ "Statement":[], "large":9007199254740993 }\n' : '<NativeXML>原文\n  保留格式</NativeXML>'
  const row = { bucket_id: 'AGJ1Y2tldA', kind, configured: true, document }
  assert.equal(definition.updateAction.disabledWhen(row), undefined)
  const values = definition.updateAction.initialValues(row)
  assert.deepEqual(values, { bucket_id: row.bucket_id, kind, document })
  assert.deepEqual(definition.updateAction.buildBody(values, 7, row), { cluster_id: 7, ...values })
  assert.throws(() => definition.updateAction.buildBody({ ...values, bucket_id: 'other' }, 7, row))
  assert.throws(() => definition.updateAction.buildBody({ ...values, kind: kind === 'policy' ? 'cors' : 'policy' }, 7, row))
  for (const action of [definition.createAction, definition.updateAction]) {
    const confirmation = action.confirmation(values, action === definition.updateAction ? row : undefined)
    assert.ok(confirmation.includes(row.bucket_id) && confirmation.includes(kind) && confirmation.includes('整体替换'))
    assert.ok(confirmation.includes('外部并发') && !confirmation.includes(document))
  }
  const absent = { ...row, configured: false, document: null }
  assert.equal(definition.updateAction.disabledWhen(absent), undefined)
  assert.equal(definition.updateAction.initialValues(absent).document, '')
  assert.throws(() => definition.updateAction.buildBody(definition.updateAction.initialValues(absent), 7, absent))
}
for (const row of [{}, { bucket_id: 'id', kind: 'policy' }, { bucket_id: 'id', kind: 'policy', configured: true, document: null }, { bucket_id: 'id', kind: 'policy', configured: false, document: '{}' }, { bucket_id: 'id', kind: 'versioning', configured: false, document: null }]) {
  assert.ok(definition.updateAction.disabledWhen(row))
  assert.throws(() => definition.updateAction.initialValues(row))
}
for (const name of ['bucket_id', 'kind']) assert.equal(definition.updateAction.fields.find(field => field.name === name).readOnly, true)
assert.equal(definition.deleteAction.path, '/rgw/bucket/policy')
assert.equal(definition.deleteAction.action, 'rgw_bucket_policy.delete')
assert.equal(definition.deleteAction.risk, 'high')
for (const kind of ['policy', 'cors', 'lifecycle', 'encryption', 'tagging']) {
  const row = { bucket_id: 'dGVhbQBiaWc', kind, configured: true, document: 'private-document' }
  assert.equal(definition.deleteAction.disabledWhen(row), undefined)
  assert.deepEqual(definition.deleteAction.buildBody(row, 7), { cluster_id: 7, bucket_id: row.bucket_id, kind })
  assert.equal(definition.deleteAction.resourceKey(row), `${row.bucket_id} / ${kind}`)
  const confirmation = definition.deleteAction.confirmation(row)
  assert.ok(confirmation.includes(row.bucket_id) && confirmation.includes(kind))
  assert.ok(confirmation.includes('不会删除 Bucket 或对象') && confirmation.includes('外部并发'))
  assert.ok(!confirmation.includes(row.document))
}
for (const row of [{}, { configured: false }, { configured: 'true' }, { configured: true, bucket_id: 'id', kind: 'versioning' }, { configured: true, bucket_id: ' id', kind: 'policy' }]) {
  assert.ok(definition.deleteAction.disabledWhen(row))
  assert.throws(() => definition.deleteAction.buildBody(row, 7))
}
const externalPage = readFileSync(new URL('../src/pages/ExternalListPage.tsx', import.meta.url), 'utf8')
assert.ok(externalPage.includes('const blocked = action.disabledWhen?.(row)'))
assert.ok(externalPage.includes('Boolean(definition.deleteAction.disabledWhen?.(row))'))
assert.equal((externalPage.match(/content: action\.confirmation\?\.\(row\)/g) ?? []).length, 2)
assert.ok(readFileSync(new URL('../src/pages/ExternalListPage.tsx', import.meta.url), 'utf8').includes('definition.buildQuery?.(queryBody)'))
console.log('Bucket configuration preserves raw JSON/XML and query-scoped reads')

const tagSource = readFileSync(new URL('../src/pages/object/RgwBucketTagEntries.tsx', import.meta.url), 'utf8')
const tagTree = ts.createSourceFile('tags.tsx', tagSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const tagNode = tagTree.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === 'rgwBucketTagEntries')
const tagCode = ts.transpileModule(tagNode.getText(tagTree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const entries = new Function(`${tagCode}; return rgwBucketTagEntries`)()
const tags = [{ key: '重复', value: '' }, { key: '重复', value: '<script>文本</script>' }, { key: ' 空白 ', value: ' 值 ' }]
assert.deepEqual(entries(tags), tags.map((tag, index) => ({ ...tag, index })))
assert.deepEqual(entries([]), [])
for (const value of [undefined, null, {}, [null], [{ key: 'a' }], [{ key: 'a', value: 1 }]]) assert.equal(entries(value), undefined)
assert.ok(tagSource.includes('rowKey="index"'))
assert.ok(!tagSource.includes('dangerouslySetInnerHTML'))
assert.ok(definition.filterFields.find(field => field.name === 'kind').options.some(option => option.value === 'tagging'))
console.log('Bucket tag entries preserve duplicates, empty values and Unicode without HTML execution')

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const compile = name => ts.transpileModule(readFileSync(new URL(`../src/pages/object/${name}`, import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX } }).outputText
const lookup = {}
const calls = []
let inventory = { items: [], stale: false, observedAt: 'snapshot' }
let failure
new Function('exports', 'require', compile('rgwUserAccountLookup.ts'))(lookup, () => ({ listAllResources: async (...args) => {
  calls.push(args)
  if (failure) throw failure
  return inventory
} }))
assert.deepEqual(await lookup.loadRgwUserAccount(7, 'RGW123'), { account: undefined, stale: false, observedAt: 'snapshot' })
assert.deepEqual(calls.pop(), ['/rgw/accounts', 7])
const account = { account_id: 'RGW123', account_name: 'team', tenant: '', email: '', stale: false, observed_at: 'account-snapshot' }
inventory.items = [{ account_id: 'other' }, account]
assert.deepEqual(await lookup.loadRgwUserAccount(8, 'RGW123'), { account, stale: false, observedAt: 'account-snapshot' })
assert.deepEqual(calls.pop(), ['/rgw/accounts', 8])
inventory.stale = true
assert.equal((await lookup.loadRgwUserAccount(8, 'RGW123')).stale, true)
inventory.stale = false
delete account.stale
assert.equal((await lookup.loadRgwUserAccount(8, 'RGW123')).stale, true)
inventory.items.push(account)
await assert.rejects(lookup.loadRgwUserAccount(8, 'RGW123'), /重复 ID/)
failure = new Error('offline')
await assert.rejects(lookup.loadRgwUserAccount(8, 'RGW123'), /offline/)
const count = calls.length
for (const cluster of [0, -1, undefined, NaN, 1.5]) await assert.rejects(lookup.loadRgwUserAccount(cluster, 'RGW123'))
for (const id of ['', ' RGW123', 'RGW123 ']) await assert.rejects(lookup.loadRgwUserAccount(7, id))
assert.equal(calls.length, count)

const components = {}
const jsx = (type, props, key) => ({ type, props, key })
let state = { data: null, loading: false, error: '', refresh: () => {} }
const loaders = []
new Function('exports', 'require', compile('RgwUserAccountDetails.tsx'))(components, name => {
  if (name === 'react/jsx-runtime') return { jsx, jsxs: jsx }
  if (name === 'react') return { useCallback: fn => fn }
  if (name === 'antd') return Object.fromEntries(['Alert', 'Button', 'Descriptions', 'Space', 'Spin'].map(name => [name, name]))
  if (name === '../../hooks') return { useResource: loader => { loaders.push(loader); return state } }
  if (name === './rgwUserIdentity') return { rgwIdentityText: (value, empty) => typeof value === 'string' ? value || empty : '未返回或格式无效' }
  if (name === './rgwUserAccountLookup') return lookup
  throw new Error(name)
})
assert.equal(components.RgwUserAccountDetails({ clusterId: 7, accountId: '' }).props.message, '此用户未关联账户')
for (const accountId of [undefined, null, 1, false, ' RGW123']) assert.match(components.RgwUserAccountDetails({ clusterId: 7, accountId }).props.message, /无效/)
assert.match(components.RgwUserAccountDetails({ accountId: 'RGW123' }).props.message, /选择集群/)
const first = components.RgwUserAccountDetails({ clusterId: 7, accountId: 'RGW123' })
assert.notEqual(first.key, components.RgwUserAccountDetails({ clusterId: 8, accountId: 'RGW123' }).key)
assert.notEqual(first.key, components.RgwUserAccountDetails({ clusterId: 7, accountId: 'RGW456' }).key)
assert.deepEqual(first.props, { clusterId: 7, accountId: 'RGW123' })
failure = undefined
inventory.items = [account]
components.AccountInventory(first.props)
await loaders.pop()()
assert.deepEqual(calls.pop(), ['/rgw/accounts', 7])
state = { ...state, data: { account, stale: true, observedAt: 'time' }, error: 'offline' }
let children = components.AccountInventory(first.props).props.children.filter(Boolean)
assert.ok(children.some(child => child.type === 'Alert' && child.props.type === 'warning'))
assert.ok(children.some(child => child.type === 'Alert' && child.props.type === 'error' && child.props.description.includes('上次读取')))
const fields = children.find(child => child.type === 'Descriptions').props.items
assert.deepEqual(fields.map(field => field.children), ['RGW123', 'team', '默认租户', '未设置邮箱', 'time'])
state = { ...state, data: { stale: false }, error: '' }
children = components.AccountInventory(first.props).props.children.filter(Boolean)
assert.ok(children.some(child => child.type === 'Alert' && child.props.message.includes('不能据此断定账户不存在')))
console.log('RGW user account lookup preserves cluster scope, exact identity and unavailable states')

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const jsx = (type, props) => ({ type, props })
function load(name, dependencies = {}) {
  const output = {}
  const code = ts.transpileModule(readFileSync(new URL(`../src/pages/object/${name}`, import.meta.url), 'utf8'), {
    compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX, target: ts.ScriptTarget.ES2022 }
  }).outputText
  new Function('exports', 'require', code)(output, id => {
    if (id === 'react/jsx-runtime') return { jsx, jsxs: jsx }
    if (id === 'antd') return { Alert: 'Alert', Tabs: 'Tabs', Descriptions: 'Descriptions' }
    assert.ok(id in dependencies, id)
    return dependencies[id]
  })
  return output
}
const details = load('rgwRateLimitDetails.ts')
const rate = load('RgwRateLimit.tsx', { './rgwRateLimitDetails': details })
const { RgwGlobalRateLimit } = load('RgwGlobalRateLimit.tsx', { './RgwRateLimit': rate })
for (const value of [undefined, null, [], '', 0, false]) {
  const view = RgwGlobalRateLimit({ value })
  assert.equal(view.type, 'Alert')
  assert.equal(view.props.type, 'warning')
}
const value = {
  user_ratelimit: { enabled: true, max_read_ops: 0, max_write_ops: 1, max_read_bytes: 1024, max_write_bytes: Number.MAX_SAFE_INTEGER },
  bucket_ratelimit: { enabled: false },
  anonymous_ratelimit: { enabled: 'unknown' }
}
for (const input of [value, {}, { ...value, anonymous_ratelimit: undefined }]) {
  const view = RgwGlobalRateLimit({ value: input })
  const [notice, tabs] = view.props.children
  assert.ok(notice.props.message.includes('每 RGW 每分钟'))
  assert.ok(notice.props.description.includes('不代表'))
  assert.deepEqual(tabs.props.items.map(item => item.key), ['user_ratelimit', 'bucket_ratelimit', 'anonymous_ratelimit'])
  for (const item of tabs.props.items) {
    assert.equal(item.children.type, rate.RgwRateLimit)
    assert.equal(item.children.props.value, input[item.key])
    const rendered = item.children.type(item.children.props)
    const state = rendered.props.items[0].children
    if (input[item.key] === undefined) assert.equal(state, '限流信息不可用')
    else if (item.key === 'bucket_ratelimit') assert.equal(state, '未启用')
    else if (item.key === 'anonymous_ratelimit') assert.equal(state, '限流启用状态未知')
    else assert.deepEqual(rendered.props.items.map(row => row.children), ['已启用', '无限制', '1', '1024', '9007199254740991'])
  }
}
const pages = readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8')
assert.match(pages, /key: 'global_rate_limit'[^\n]+render: \(value\) => <RgwGlobalRateLimit value=\{value\} \/>/)
console.log('Global RGW rate limit scopes preserve unavailable, disabled and exact values')

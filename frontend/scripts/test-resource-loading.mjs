import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/hooks.ts', import.meta.url), 'utf8')
const tree = ts.createSourceFile('hooks.ts', source, ts.ScriptTarget.Latest, true)
const fn = tree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'useResource')
const code = ts.transpileModule(fn.getText(tree).replace('export ', ''), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText

// Deterministic hook lifecycle harness: explicitly render, flush effects, and settle requests.
function harness() {
  const slots = [], effects = []
  let cursor = 0, clusterId = 1, writes = 0
  const same = (a, b) => a?.length === b.length && a.every((value, i) => Object.is(value, b[i]))
  const useState = (initial) => {
    const i = cursor++
    slots[i] ??= { value: initial }
    return [slots[i].value, (value) => { writes++; slots[i].value = typeof value === 'function' ? value(slots[i].value) : value }]
  }
  const useRef = (value) => { const i = cursor++; return slots[i] ??= { current: value } }
  const useCallback = (value, deps) => {
    const i = cursor++
    if (!same(slots[i]?.deps, deps)) slots[i] = { value, deps }
    return slots[i].value
  }
  const useEffect = (effect, deps) => {
    const i = cursor++
    if (!same(slots[i]?.deps, deps)) {
      effects.push(() => { slots[i]?.cleanup?.(); slots[i] = { deps, effect, cleanup: effect() } })
    }
  }
  const hook = new Function('useState', 'useRef', 'useCallback', 'useEffect', 'useClusterContext', `${code}; return useResource`)(useState, useRef, useCallback, useEffect, () => ({ selectedClusterId: clusterId }))
  return {
    render(loader, cluster = clusterId) { clusterId = cluster; cursor = 0; return hook(loader) },
    flush() { while (effects.length) effects.shift()() },
    unmount() { slots.forEach((slot) => slot.cleanup?.()) },
    strictReplay() { slots.forEach((slot) => { if (slot.effect) { slot.cleanup?.(); slot.cleanup = slot.effect() } }) },
    get writes() { return writes }
  }
}
function loaderQueue() {
  const pending = []
  return { pending, loader: () => new Promise((resolve, reject) => pending.push({ resolve, reject })) }
}
const settle = async () => { await Promise.resolve(); await Promise.resolve() }

for (const oldFails of [false, true]) {
  const h = harness(), q = loaderQueue()
  h.render(q.loader); h.flush()
  const refresh = h.render(q.loader).refresh({ showLoading: false })
  q.pending[1].resolve('newest'); await refresh
  oldFails ? q.pending[0].reject(new Error('obsolete failure')) : q.pending[0].resolve('obsolete data')
  await settle()
  const result = h.render(q.loader)
  assert.equal(result.data, 'newest'); assert.equal(result.error, ''); assert.equal(result.loading, false)
}
for (const change of ['loader', 'cluster']) for (const oldFails of [false, true]) {
  const h = harness(), old = loaderQueue(), next = change === 'loader' ? loaderQueue() : old
  h.render(old.loader); h.flush(); old.pending[0].resolve('old data'); await settle()
  const obsoleteRefresh = h.render(old.loader).refresh
  const pending = obsoleteRefresh()
  const cluster = change === 'cluster' ? 2 : 1
  let result = h.render(next.loader, cluster)
  assert.equal(result.data, null, 'old resource must disappear before effect execution')
  assert.equal(result.loading, true)
  await obsoleteRefresh()
  assert.equal(old.pending.length, 2, 'obsolete refresh callbacks must not start requests')
  h.flush()
  oldFails ? old.pending[1].reject(new Error('old failure')) : old.pending[1].resolve('old refresh')
  await pending
  next.pending.at(-1).reject(new Error('new failure')); await settle()
  result = h.render(next.loader, cluster)
  assert.equal(result.data, null); assert.equal(result.error, 'new failure'); assert.equal(result.loading, false)
}
{
  const h = harness(), q = loaderQueue()
  h.render(q.loader); h.flush(); q.pending[0].resolve('cached'); await settle()
  const pending = h.render(q.loader).refresh({ showLoading: false })
  assert.equal(h.render(q.loader).data, 'cached')
  assert.equal(h.render(q.loader).loading, false)
  q.pending[1].reject(new Error('refresh failed')); await pending
  assert.equal(h.render(q.loader).data, 'cached')
  assert.equal(h.render(q.loader).error, 'refresh failed')
  const old = h.render(q.loader).refresh()
  const latest = h.render(q.loader).refresh()
  q.pending[3].resolve('latest'); await latest
  q.pending[2].resolve('older'); await old
  assert.equal(h.render(q.loader).data, 'latest')
}
for (const fails of [false, true]) {
  const h = harness(), q = loaderQueue()
  h.render(q.loader); h.flush()
  const refresh = h.render(q.loader).refresh
  const pending = refresh()
  h.unmount()
  const writes = h.writes
  for (const request of q.pending) fails ? request.reject(new Error('unmounted')) : request.resolve('unmounted')
  await pending; await settle(); await refresh()
  assert.equal(h.writes, writes); assert.equal(q.pending.length, 2)
}
{
  const h = harness(), q = loaderQueue()
  h.render(q.loader); h.flush(); h.strictReplay()
  q.pending[0].resolve('discarded strict effect'); await settle()
  assert.equal(h.render(q.loader).data, null)
  q.pending[1].resolve('active strict effect'); await settle()
  assert.equal(h.render(q.loader).data, 'active strict effect')
}
console.log('Resource loading scope and request ordering checks passed')

import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

function load(file, imports = {}) {
  const source = readFileSync(new URL(`../src/pages/file/${file}.ts`, import.meta.url), 'utf8')
  const compiled = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS } }).outputText
  const exports = {}
  new Function('exports', 'require', compiled)(exports, (name) => imports[name])
  return exports
}
const series = load('cephfsPerformanceSeries')
const { rankSample, rankActivity, rankClientCount } = load('cephfsRankMetrics', { './cephfsPerformanceSeries': series })
const row = { name: 'a', gid: '1', rank: '0', state: 'active' }
const sample = (requests, replay, clients, seconds = 0, gid = '1') => ({ name: 'a', gid, observed_at: new Date(seconds * 1000).toISOString(), counters: [
  { name: 'mds_server.handle_client_request', value: requests }, { name: 'mds_log.replay', value: replay }, { name: 'mds_sessions.session_count', value: clients }
] })
const first = sample('100', '200', '9007199254740993')
const next = sample('120', '250', '9007199254740993', 10)
assert.equal(rankActivity(row, [next], { a: [first, next] }), 2)
assert.equal(rankActivity({ ...row, state: 'standby-replay' }, [next], { a: [first, next] }), 5)
assert.equal(rankActivity({ ...row, state: 'reconnect' }, [next], { a: [first, next] }), undefined)
assert.equal(rankSample({ ...row, gid: '2' }, [next]), undefined)
assert.equal(rankSample(row, [{ ...next, error: 'failed' }]), undefined)
assert.equal(rankActivity(row, [next], { a: [first] }), undefined)
assert.equal(rankClientCount([row], [next]), '9007199254740993')
assert.equal(rankClientCount([{ ...row, rank: '0-s' }], [next]), undefined)
assert.equal(rankClientCount([row, { ...row, name: 'b', rank: '1', gid: '2' }], [sample('1','1','0'), { ...next, name: 'b', gid: '2' }]), '9007199254740993')
console.log('CephFS rank metric checks passed')

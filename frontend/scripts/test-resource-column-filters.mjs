import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

function load(path, require = () => { throw new Error('unexpected import') }) {
  const source = readFileSync(new URL(path, import.meta.url), 'utf8')
  const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText
  const exports = {}
  new Function('exports', 'require', compiled)(exports, require)
  return exports
}
const helpers = load('../src/pages/resourceColumnFilters.ts')
const columns = [
  { key: 'filesystem', title: '文件系统' },
  { key: 'state_label', title: '状态', filterKey: 'state' },
  { key: 'state', title: '原生状态' },
  { key: 'orphan_clones_count', title: '孤儿克隆', filterKey: false }
]
assert.deepEqual(helpers.resourceFilterFields(columns), ['filesystem', 'state'])
assert.deepEqual(helpers.resourceFilterFields([]), [])
assert.deepEqual(helpers.resourceColumnFilters(columns[1], { state: ['complete', 'snapshot-retained'], state_label: ['wrong'] }, { state: ['complete'] }), {
  key: 'state', filterMultiple: true, filterSearch: true,
  filters: [{ text: 'complete', value: 'complete' }, { text: 'snapshot-retained', value: 'snapshot-retained' }], filteredValue: ['complete']
})
assert.deepEqual(helpers.resourceColumnFilters(columns[3], { orphan_clones_count: ['0'] }, {}), { key: 'orphan_clones_count' })
assert.equal(helpers.resourceColumnFilters(columns[0], {}, {}).filteredValue, null)

const requests = []
const resource = load('../src/api/resource.ts', (path) => {
  assert.equal(path, './client')
  return { jsonInit: (method, body) => ({ method, body }), request: async (url, init) => { requests.push({ url, init }); return { items: [], filter_options: {} } } }
})
await resource.listResourceFilterOptions('/filesystem/subvolume/groups', helpers.resourceFilterFields(columns), 1)
assert.equal(new URL(requests[0].url, 'https://example.invalid').searchParams.get('fields'), 'filesystem,state')
const config = helpers.resourceColumnFilters(columns[1], { state: ['complete'] }, {})
await resource.listResource('/filesystem/subvolumes', 1, { filters: { [config.key]: ['complete'] } })
const query = new URL(requests[1].url, 'https://example.invalid').searchParams
assert.equal(query.get('filter.state'), 'complete')
assert.equal(query.get('filter.state_label'), null)
const page = readFileSync(new URL('../src/pages/ResourceListPage.tsx', import.meta.url), 'utf8')
assert.match(page, /resourceFilterFields\(definition.columns\)/)
assert.match(page, /resourceColumnFilters\(column, filterOptions, columnFilters\)/)
const pages = readFileSync(new URL('../src/pages/file/pages.tsx', import.meta.url), 'utf8')
const start = pages.indexOf('  subvolumeGroups: {')
const group = pages.slice(start, pages.indexOf('\n  },', start))
assert.match(group, /key: 'filesystem'/)
assert.doesNotMatch(group, /key: 'fs'/)
assert.match(pages, /key: 'orphan_clones_count'.*filterKey: false/)
console.log('Resource column filter checks passed')

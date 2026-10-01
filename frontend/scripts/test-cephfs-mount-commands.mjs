import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { spawnSync } from 'node:child_process'
import React from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import ts from 'typescript'

const nativeRequire = createRequire(import.meta.url)
const fsid = '00000000-0000-0000-0000-000000000001'
function load(url) {
  const compiled = ts.transpileModule(readFileSync(url, 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } }).outputText
  const exports = {}
  new Function('exports', 'require', compiled)(exports, (path) => {
    if (path === '../../state/ClusterContext') return { useClusterContext: () => ({ selectedCluster: { id: 1, fsid } }) }
    return path.startsWith('.') ? load(new URL(`${path}.ts`, url)) : nativeRequire(path)
  })
  return exports
}
const helpers = load(new URL('../src/pages/file/cephfsMountCommands.ts', import.meta.url))
const input = { fsid, filesystem: 'cephfs', path: '/volumes/team/volume', client: 'client.app', mountDirectory: '/mnt/cephfs' }
const commands = helpers.cephFSMountCommands(input)
assert.equal(commands.kernel, `sudo mount -t ceph 'app@${fsid}.cephfs=/volumes/team/volume' '/mnt/cephfs'`)
assert.equal(commands.fuse, "sudo ceph-fuse --id='app' --client_fs='cephfs' -r '/volumes/team/volume' '/mnt/cephfs'")
assert.match(commands.nfs, /NFS_HOST:EXPORT_PATH/)
assert.doesNotMatch(commands.nfs, /volumes\/team/)
for (const patch of [{ fsid: '' }, { fsid: 'invalid' }, { filesystem: '' }, { filesystem: 'fs=x' }, { path: '' }, { path: 'relative' }, { path: '/bad\npath' }, { client: '' }, { client: 'app; printf injected' }, { mountDirectory: '/' }, { mountDirectory: 'relative' }]) {
  const result = helpers.cephFSMountCommands({ ...input, ...patch })
  assert.ok(result.error)
  assert.equal(result.kernel, undefined)
  assert.equal(result.fuse, undefined)
}
const unusual = { ...input, path: "/volumes/it's a path; $(printf injected)", mountDirectory: "/mnt/it's a mount" }
const quoted = helpers.cephFSMountCommands(unusual)
// Parse the generated shell arguments using `set`, never invoke mount or sudo.
function parse(command) {
  const result = spawnSync('/bin/sh', ['-c', `set -- ${command}; printf '%s\\n' "$@"`], { encoding: 'utf8' })
  assert.equal(result.status, 0, result.stderr)
  return result.stdout.trimEnd().split('\n')
}
assert.deepEqual(parse(quoted.kernel), ['sudo', 'mount', '-t', 'ceph', `app@${fsid}.cephfs=${unusual.path}`, unusual.mountDirectory])
assert.deepEqual(parse(quoted.fuse), ['sudo', 'ceph-fuse', '--id=app', '--client_fs=cephfs', '-r', unusual.path, unusual.mountDirectory])
const components = load(new URL('../src/pages/file/CephFSAttachCommands.tsx', import.meta.url))
const html = renderToStaticMarkup(React.createElement(components.CephFSAttachCommands, input))
for (const text of ['Linux 内核客户端', 'FUSE 客户端', 'NFS 模板', '--client_fs=', fsid, '不会执行挂载或获取密钥']) assert.ok(html.includes(text))
const missing = renderToStaticMarkup(React.createElement(components.CephFSAttachCommands, { ...input, fsid: '' }))
assert.match(missing, /未获取到有效的集群 FSID/)
assert.doesNotMatch(missing, /sudo mount/)
const mismatch = renderToStaticMarkup(React.createElement(components.CephFSSubvolumeMount, { clusterId: 2, filesystem: 'cephfs', path: input.path }))
assert.doesNotMatch(mismatch, /sudo mount/)
const matched = renderToStaticMarkup(React.createElement(components.CephFSSubvolumeMount, { clusterId: 1, filesystem: 'cephfs', path: input.path }))
assert.match(matched, /sudo mount/)
const pages = readFileSync(new URL('../src/pages/file/pages.tsx', import.meta.url), 'utf8')
assert.match(pages, /detailContent:.*subvolumeReadyReason\(row\).*CephFSSubvolumeMount.*path=\{row.path\}/)
console.log('CephFS mount command checks passed')

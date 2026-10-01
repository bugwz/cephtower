function quote(value: string): string {
  return `'${value.replace(/'/g, `'"'"'`)}'`
}

export function cephFSMountCommands({ fsid, filesystem, path, client, mountDirectory }: { fsid: unknown; filesystem: unknown; path: unknown; client: string; mountDirectory: string }) {
  if (typeof fsid !== 'string' || !/^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i.test(fsid)) return { error: '未获取到有效的集群 FSID，请刷新集群信息。' }
  if (typeof filesystem !== 'string' || !/^[A-Za-z0-9_.-]+$/.test(filesystem)) return { error: '文件系统名称不可用。' }
  if (typeof path !== 'string' || !path.startsWith('/') || /[\x00-\x1f\x7f]/.test(path)) return { error: '原生子卷路径不可用，不能生成挂载命令。' }
  const username = client.trim().replace(/^client\./, '')
  if (!/^[A-Za-z0-9_][A-Za-z0-9_.-]*$/.test(username)) return { error: '请输入 CephX 客户端名称，例如 app 或 client.app。' }
  if (!mountDirectory.startsWith('/') || mountDirectory === '/' || /[\x00-\x1f\x7f]/.test(mountDirectory)) return { error: '请输入非根目录的本地绝对挂载路径。' }
  return {
    kernel: `sudo mount -t ceph ${quote(`${username}@${fsid}.${filesystem}=${path}`)} ${quote(mountDirectory)}`,
    fuse: `sudo ceph-fuse --id=${quote(username)} --client_fs=${quote(filesystem)} -r ${quote(path)} ${quote(mountDirectory)}`,
    nfs: `sudo mount -t nfs -o 'port=NFS_PORT' 'NFS_HOST:EXPORT_PATH' ${quote(mountDirectory)}`
  }
}

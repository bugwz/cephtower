import { Alert, Card, Input, Space, Typography } from 'antd'
import { useState } from 'react'
import { useClusterContext } from '../../state/ClusterContext'
import { cephFSMountCommands } from './cephfsMountCommands'

export function CephFSSubvolumeMount({ clusterId, filesystem, path }: { clusterId?: number; filesystem: string; path: unknown }) {
  const { selectedCluster } = useClusterContext()
  return <CephFSAttachCommands fsid={selectedCluster?.id === clusterId ? selectedCluster?.fsid : undefined} filesystem={filesystem} path={path} />
}

export function CephFSAttachCommands({ fsid, filesystem, path }: { fsid: unknown; filesystem: unknown; path: unknown }) {
  const [client, setClient] = useState('CLIENT_USER')
  const [mountDirectory, setMountDirectory] = useState('/mnt/cephfs')
  const commands = cephFSMountCommands({ fsid, filesystem, path, client, mountDirectory })
  return <Card title="子卷挂载命令">
    <Space direction="vertical" style={{ width: '100%' }}>
      <Alert type="info" showIcon message="仅生成客户端命令，不会执行挂载或获取密钥" description="将 CLIENT_USER 替换为有权访问此子卷的 CephX 用户；在客户端配置对应集群的 ceph.conf 和用户 keyring，确认本地挂载目录已存在。CephTower 的服务器凭据不会复制到客户端。" />
      <label>CephX 客户端<Input value={client} onChange={(event) => setClient(event.target.value)} placeholder="app 或 client.app" /></label>
      <label>客户端挂载目录<Input value={mountDirectory} onChange={(event) => setMountDirectory(event.target.value)} /></label>
      {commands.error ? <Alert type="warning" showIcon message={commands.error} /> : <>
        <Typography.Text strong>Linux 内核客户端</Typography.Text>
        <Typography.Paragraph copyable={{ text: commands.kernel }}><code style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{commands.kernel}</code></Typography.Paragraph>
        <Typography.Text strong>FUSE 客户端</Typography.Text>
        <Typography.Paragraph copyable={{ text: commands.fuse }}><code style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{commands.fuse}</code></Typography.Paragraph>
        <Typography.Text strong>NFS 模板（需要已有 NFS 导出）</Typography.Text>
        <Typography.Text type="secondary">请手动替换 NFS_HOST、NFS_PORT、EXPORT_PATH。本模板不代表此子卷已有导出，EXPORT_PATH 也不一定等于 CephFS 内部路径。</Typography.Text>
        <Typography.Paragraph copyable={{ text: commands.nfs }}><code style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{commands.nfs}</code></Typography.Paragraph>
      </>}
    </Space>
  </Card>
}

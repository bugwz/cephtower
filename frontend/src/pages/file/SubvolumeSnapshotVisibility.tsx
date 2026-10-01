import { Alert, Button, Card, Modal, Space, Switch, Typography } from 'antd'
import { useEffect, useState } from 'react'
import { jsonInit, request } from '../../api/client'
import { mutateResource } from '../../api/resource'
import { useFeatureRequirements } from '../../hooks/useFeatureRequirements'
import { useMutationOperation } from '../../hooks/useMutationOperation'

interface Visibility { visible: boolean; observed_at: string }
const path = '/filesystem/subvolume/snapshot/visibility'

export function SubvolumeSnapshotVisibility({ clusterId, filesystem, subvolume, group, resourceVersion }: { clusterId?: number; filesystem: string; subvolume: string; group: string; resourceVersion?: string }) {
  const [data, setData] = useState<Visibility | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [revision, setRevision] = useState(0)
  const operation = useMutationOperation()
  const feature = useFeatureRequirements(clusterId, { requiredCapabilities: ['cephfs_volume'] })
  useEffect(() => {
    const controller = new AbortController()
    setData(null); setError(''); setLoading(false)
    if (!clusterId || !filesystem || !subvolume) return
    setLoading(true)
    void request<Visibility>(path, jsonInit('GET', { cluster_id: clusterId, fs: filesystem, subvolume, group }, { signal: controller.signal, suppressErrorNotification: true })).then((result) => {
      if (typeof result.visible !== 'boolean') throw new Error('快照可见性响应无效')
      if (!controller.signal.aborted) setData(result)
    }).catch((err) => {
      if (!controller.signal.aborted) setError(err instanceof Error ? err.message : '读取快照可见性失败')
    }).finally(() => { if (!controller.signal.aborted) setLoading(false) })
    return () => controller.abort()
  }, [clusterId, filesystem, subvolume, group, revision])

  function change(visible: boolean) {
    Modal.confirm({ title: visible ? '允许快照浏览' : '隐藏子卷快照', content: '此设置仅影响尊重可见性配置的客户端，不会删除快照。确认修改？', onOk: async () => {
      try {
        await operation.run(() => mutateResource(path, 'PUT', { cluster_id: clusterId, fs: filesystem, subvolume, group, visible }, resourceVersion ? { ifMatch: resourceVersion } : undefined), '快照可见性已更新并验证')
        setRevision((value) => value + 1)
      } catch (err) {
        setData(null)
        setError(err instanceof Error ? err.message : '设置失败，请重新读取实际状态')
      }
    } })
  }

  return <Card title="子卷快照浏览" extra={<Button disabled={!clusterId || operation.loading} loading={loading} onClick={() => setRevision((value) => value + 1)}>实时读取</Button>}>
    <Space direction="vertical" style={{ width: '100%' }}>
      <Alert type="warning" showIcon message="需要客户端启用 client_respect_subvolume_snapshot_visibility=true" description="本操作不会自动修改客户端配置。当前仅 FUSE/libcephfs 客户端支持；隐藏后这些客户端将无法浏览 .snap 或执行依赖其访问的快照操作。旧版 Ceph 不支持此命令时会显示读取错误，不会假定为已启用。" />
      {error && <Alert type="error" showIcon message={error} />}
      <Space><Typography.Text>允许快照浏览：{data == null ? '未知' : data.visible ? '是' : '否'}</Typography.Text>
        <Switch checked={data?.visible ?? false} loading={operation.loading} disabled={!clusterId || loading || data == null || feature.loading || feature.blocked || Boolean(feature.error)} onChange={change} />
      </Space>
      {data && <Typography.Text type="secondary">观测时间：{new Date(data.observed_at).toLocaleString()}</Typography.Text>}
    </Space>
  </Card>
}

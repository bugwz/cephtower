import { Alert, Button, Card, Progress, Tag, Tooltip, Typography } from 'antd'
import { useEffect, useState } from 'react'
import { jsonInit, request } from '../../api/client'
import { AppTable } from '../../components/AppTable'
import { formatPoolBytes, poolUsagePercent } from './cephfsPoolCapacity'

interface PoolUsage {
  id: string; name: string; type: 'metadata' | 'data'
  stored: string | null; available: string | null; size: string | null
  bytes_used: string | null; error?: string
}
interface PoolList { filesystem: string; items: PoolUsage[]; observed_at: string }

export function CephFSPoolUsage({ clusterId, filesystem, refreshToken }: { clusterId?: number; filesystem: string; refreshToken: number }) {
  const [data, setData] = useState<PoolList | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [revision, setRevision] = useState(0)
  useEffect(() => {
    const controller = new AbortController()
    setData(null)
    setError('')
    setLoading(false)
    if (!clusterId || !filesystem) return
    setLoading(true)
    void request<PoolList>('/filesystem/pools', jsonInit('GET', {
      cluster_id: clusterId, fs: filesystem
    }, { signal: controller.signal, suppressErrorNotification: true })).then((result) => {
      if (!controller.signal.aborted) setData(result)
    }).catch((err) => {
      if (!controller.signal.aborted) setError(err instanceof Error ? err.message : '读取文件系统存储池容量失败')
    }).finally(() => { if (!controller.signal.aborted) setLoading(false) })
    return () => controller.abort()
  }, [clusterId, filesystem, refreshToken, revision])

  const bytes = (value: string | null) => <Tooltip title={value == null ? '统计缺失' : `${value} 字节（精确值）`}>{formatPoolBytes(value)}</Tooltip>
  return <Card title="文件系统存储池容量" extra={<Button loading={loading} disabled={!clusterId} onClick={() => setRevision((value) => value + 1)}>实时读取</Button>}>
    <Alert type="info" showIcon message="逻辑容量 = stored + max_avail；物理占用包含副本或纠删码开销" description="这是当前放置约束与集群空间下的池容量估计，不是独占或预留容量；共享池可能还保存其他文件系统的数据。数值通过整数计算保留精度。" />
    {error && <Alert type="error" showIcon message={error} />}
    {data && <Typography.Text type="secondary">观测时间：{new Date(data.observed_at).toLocaleString()}</Typography.Text>}
    <AppTable<PoolUsage> loading={loading} rowKey="id" pagination={false} dataSource={data?.items ?? []} columns={[
      { title: '池', dataIndex: 'name', render: (value, row) => value || `Pool ID ${row.id}` },
      { title: 'ID', dataIndex: 'id' },
      { title: '用途', dataIndex: 'type', render: (value) => <Tag>{value === 'metadata' ? '元数据' : '数据'}</Tag> },
      { title: '逻辑存储量', dataIndex: 'stored', render: bytes },
      { title: '最大可用量', dataIndex: 'available', render: bytes },
      { title: '估算逻辑容量', dataIndex: 'size', render: bytes },
      { title: '物理占用', dataIndex: 'bytes_used', render: bytes },
      { title: '逻辑使用率', render: (_, row) => {
        const percent = poolUsagePercent(row.stored, row.size)
        return row.error ? <Typography.Text type="warning">{row.error}</Typography.Text> : percent === undefined ? '—' : <Progress percent={percent} size="small" style={{ minWidth: 120 }} />
      } }
    ]} />
  </Card>
}

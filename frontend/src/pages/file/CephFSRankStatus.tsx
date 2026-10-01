import { Alert, Card, Space, Tag, Typography } from 'antd'
import { useEffect, useState } from 'react'
import { jsonInit, request } from '../../api/client'
import { AppTable } from '../../components/AppTable'
import { type PerformanceSample } from './cephfsPerformanceSeries'
import { rankActivity, rankClientCount, rankSample } from './cephfsRankMetrics'

interface Rank { rank: string; name: string; gid: string; state: string; laggy: boolean; version: string }
interface Topology { ranks: Rank[]; standbys: Rank[]; metadata_error?: string; observed_at: string }
type Sample = PerformanceSample & { name: string; error?: string }

export function CephFSRankStatus({ clusterId, filesystem, automatic, revision, samples, history }: {
  clusterId?: number; filesystem: string; automatic: boolean; revision: number
  samples: Sample[]; history: Record<string, PerformanceSample[]>
}) {
  const [topology, setTopology] = useState<Topology | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  useEffect(() => {
    const controller = new AbortController()
    let timer: ReturnType<typeof setTimeout> | undefined
    setTopology(null)
    setError('')
    async function load() {
      if (!clusterId || !filesystem) { setLoading(false); return }
      setLoading(true)
      try {
        const result = await request<Topology>('/filesystem/mds', jsonInit('GET', { cluster_id: clusterId, fs: filesystem }, { signal: controller.signal, suppressErrorNotification: true }))
        if (!controller.signal.aborted) { setTopology(result); setError('') }
      } catch (err) {
        if (!controller.signal.aborted) { setError(err instanceof Error ? err.message : '读取 MDS 拓扑失败'); setTopology(null) }
      } finally {
        if (!controller.signal.aborted) { setLoading(false); if (automatic) timer = setTimeout(() => void load(), 10000) }
      }
    }
    void load()
    return () => { controller.abort(); if (timer) clearTimeout(timer) }
  }, [automatic, clusterId, filesystem, revision])

  const sampleFor = (row: Rank) => rankSample(row, samples)
  const metric = (row: Rank, name: string) => sampleFor(row)?.counters.find((counter) => counter.name === name)?.value ?? '—'
  const activity = (row: Rank) => {
    const value = rankActivity(row, samples, history)
    return value === undefined ? '—' : `${value.toLocaleString(undefined, { maximumFractionDigits: 2 })} ${row.state === 'standby-replay' ? 'Evts' : 'Reqs'}/s`
  }
  const clients = rankClientCount(topology?.ranks ?? [], samples)
  return <Card type="inner" title="Rank 与备用 MDS" loading={loading && !topology}>
    {error && <Alert type="error" showIcon message={error} />}
    {topology?.metadata_error && <Alert type="warning" showIcon message="MDS 版本元数据读取失败" description={topology.metadata_error} />}
    <Typography.Paragraph>客户端会话数：{clients ?? '—'}（优先 Rank 0；未求和，避免重复统计）</Typography.Paragraph>
    <AppTable<Rank> rowKey={(row) => `${row.rank}:${row.gid}`} pagination={false} dataSource={topology?.ranks ?? []} columns={[
      { title: 'Rank', dataIndex: 'rank' },
      { title: '状态', render: (_, row) => <Space><Tag color={row.state === 'failed' ? 'error' : 'blue'}>{row.state}</Tag>{row.laggy && <Tag color="warning">laggy</Tag>}</Space> },
      { title: 'MDS', dataIndex: 'name', render: (value) => value || '—' },
      { title: '活动', render: (_, row) => activity(row) },
      { title: 'Dentries', render: (_, row) => metric(row, 'mds_mem.dn') },
      { title: 'Inodes', render: (_, row) => metric(row, 'mds_mem.ino') },
      { title: 'Dirs', render: (_, row) => metric(row, 'mds_mem.dir') },
      { title: 'Caps', render: (_, row) => metric(row, 'mds_mem.cap') },
      { title: 'Ceph 版本', dataIndex: 'version', render: (value) => value || '—' },
      { title: '指标采样时间', render: (_, row) => sampleFor(row) ? new Date(sampleFor(row)!.observed_at).toLocaleString() : '—' }
    ]} />
    <Typography.Paragraph type="secondary">全局备用 MDS（尚未分配给某个文件系统，不能视为此 FS 独占资源）</Typography.Paragraph>
    <AppTable<Rank> rowKey="gid" pagination={false} dataSource={topology?.standbys ?? []} columns={[
      { title: '名称', dataIndex: 'name' }, { title: 'GID', dataIndex: 'gid' },
      { title: '状态', dataIndex: 'state' }, { title: 'Ceph 版本', dataIndex: 'version', render: (value) => value || '—' }
    ]} />
    {topology && <Typography.Text type="secondary">拓扑观测时间：{new Date(topology.observed_at).toLocaleString()}；指标来自独立采样，仅按匹配的 GID 关联。</Typography.Text>}
  </Card>
}

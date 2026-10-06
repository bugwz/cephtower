import { CopyOutlined, DownloadOutlined, ReloadOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Descriptions, Input, Select, Space, Switch, Tag, Typography } from 'antd'
import { useEffect, useState } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'
import { AppTable } from '../../components/AppTable'
import { Page } from '../../components/Page'
import { ResourceMetaBar } from '../../components/ResourceMetaBar'
import { useClusterContext } from '../../state/ClusterContext'
import { message } from '../../utils/appMessage'

export function runtimeLogsText(rows: ApiRecord[]): string {
  return rows.map((row) => `${row.stamp} [${row.channel}] ${row.priority} ${row.name}: ${row.message}`).join('\n')
}

export function runtimeLogDetails(row: ApiRecord) {
  return [['name', '来源'], ['rank', '来源 Rank'], ['seq', '原生序列号'], ['stamp', '原生时间'], ['channel', '频道'], ['priority', '级别'], ['addrs', '来源地址'], ['message', '消息']].map(([key, label]) => ({
    key, label, children: row[key] == null ? '未提供' : typeof row[key] === 'object' ? JSON.stringify(row[key], null, 2) : String(row[key])
  }))
}

export function runtimeLogKey(row: ApiRecord): string {
  return JSON.stringify([row.channel, row.name, row.rank, row.stamp, row.seq])
}

interface RuntimeLogSnapshot { scope: string; rows: ApiRecord[]; observed?: string; error: string }
export function visibleRuntimeLogs(snapshot: RuntimeLogSnapshot | null, scope: string): RuntimeLogSnapshot {
  return snapshot?.scope === scope ? snapshot : { scope, rows: [], error: '' }
}

export async function copyRuntimeLogs(rows: ApiRecord[]): Promise<void> {
  if (!navigator.clipboard?.writeText) throw new Error('当前浏览器不支持剪贴板写入，请下载日志。')
  await navigator.clipboard.writeText(runtimeLogsText(rows))
}

function runtimeLogMatches(row: ApiRecord, search: string, start: string, end: string, priority = ''): boolean {
  if (priority && row.priority !== priority) return false
  const text = [row.message, row.name, row.stamp, row.channel, row.priority].map((value) => String(value ?? '')).join(' ').toLowerCase()
  if (!text.includes(search.toLowerCase())) return false
  if (!start && !end) return true
  const stamp = Date.parse(String(row.stamp ?? ''))
  const lower = start ? Date.parse(start) : -Infinity
  const upper = end ? Date.parse(end) : Infinity
  return Number.isFinite(stamp) && !Number.isNaN(lower) && !Number.isNaN(upper) && stamp >= lower && stamp <= upper
}

export function RuntimeLogsPage() {
  return <RuntimeLogsPanel />
}

export function RuntimeLogsPanel({ compact = false }: { compact?: boolean }) {
  const { selectedClusterId } = useClusterContext()
  return <RuntimeLogsContent key={selectedClusterId ?? 'none'} selectedClusterId={selectedClusterId} compact={compact} />
}

function RuntimeLogsContent({ compact, selectedClusterId }: { compact: boolean; selectedClusterId?: number }) {
  const [channel, setChannel] = useState('cluster')
  const [level, setLevel] = useState('debug')
  const [limit, setLimit] = useState(100)
  const [auto, setAuto] = useState(true)
  const [revision, setRevision] = useState(0)
  const [loading, setLoading] = useState(false)
  const scope = JSON.stringify([selectedClusterId, channel, level, limit])
  const [snapshot, setSnapshot] = useState<RuntimeLogSnapshot | null>(null)
  const { rows, observed, error } = visibleRuntimeLogs(snapshot, scope)
  const [search, setSearch] = useState('')
  const [start, setStart] = useState('')
  const [end, setEnd] = useState('')
  const [priority, setPriority] = useState('')
  useEffect(() => {
    const abort = new AbortController()
    let timer: ReturnType<typeof setTimeout> | undefined
    setSnapshot(null); setLoading(Boolean(selectedClusterId))
    async function read() {
      if (!selectedClusterId) return
      setLoading(true)
      try {
        const data = await request<{ items: ApiRecord[]; observed_at: string }>('/logs', jsonInit('GET', {
          cluster_id: selectedClusterId, channel, level, limit
        }, { signal: abort.signal, suppressErrorNotification: true }))
        if (!abort.signal.aborted) setSnapshot({ scope, rows: data.items, observed: data.observed_at, error: '' })
      } catch (err) {
        if (!abort.signal.aborted) setSnapshot(current => ({ ...visibleRuntimeLogs(current, scope), error: err instanceof Error && err.message ? err.message : '读取日志失败' }))
      } finally {
        if (!abort.signal.aborted) {
          setLoading(false)
          if (auto) timer = setTimeout(read, 10000)
        }
      }
    }
    void read()
    return () => { abort.abort(); clearTimeout(timer) }
  }, [selectedClusterId, channel, level, limit, auto, revision, scope])
  const invalidRange = Boolean(start && end && Date.parse(start) > Date.parse(end))
  const filtered = rows.filter((row) => runtimeLogMatches(row, search, start, end, priority))
  function download() {
    const content = runtimeLogsText(filtered)
    const url = URL.createObjectURL(new Blob([content], { type: 'text/plain;charset=utf-8' }))
    const link = document.createElement('a'); link.href = url; link.download = `ceph-${channel === '*' ? 'all' : channel}.log`; link.click()
    setTimeout(() => URL.revokeObjectURL(url), 1000)
  }
  const content = <>
    {compact && error && <Alert type="error" message={error} />}
    <Space direction="vertical" size={16} className="page-stack">
      <Alert type="info" showIcon message="显示 MON 日志缓冲区的最近记录。自动刷新每 10 秒读取一次；此视图不提供长期日志归档。" />
      <Card>
        <Space wrap>
          {!compact && <><Select aria-label="日志频道" value={channel} onChange={setChannel} style={{ width: 160 }} options={[
            { value: 'cluster', label: '集群日志' }, { value: 'audit', label: 'Ceph 审计日志' }, { value: 'cephadm', label: 'Cephadm' }, { value: '*', label: '全部频道' }
          ]} />
          <Select aria-label="最低级别" value={level} onChange={setLevel} style={{ width: 140 }} options={['debug', 'info', 'sec', 'warn', 'error'].map((value) => ({ value, label: `最低 ${value}` }))} />
          <Select aria-label="读取条数" value={limit} onChange={setLimit} options={[30, 100, 300, 500].map((value) => ({ value, label: `${value} 条` }))} /></>}
          <Switch checked={auto} onChange={setAuto} checkedChildren="自动刷新" unCheckedChildren="已暂停" />
          <Button icon={<ReloadOutlined />} loading={loading} disabled={!selectedClusterId} onClick={() => setRevision((n) => n + 1)}>刷新</Button>
          {!compact && <><Button icon={<DownloadOutlined />} disabled={!filtered.length} onClick={download}>下载当前结果</Button>
          <Button icon={<CopyOutlined />} disabled={!filtered.length} onClick={() => { void copyRuntimeLogs(filtered).then(() => message.success('已复制当前筛选日志')).catch((err: unknown) => message.error(err instanceof Error ? err.message : '复制日志失败')) }}>复制当前结果</Button>
          <Input.Search allowClear placeholder="搜索消息、来源、时间、频道或级别" value={search} onChange={(event) => setSearch(event.target.value)} />
          <Select aria-label="精确日志级别" value={priority} onChange={setPriority} style={{ width: 170 }} options={[
            { value: '', label: '显示全部返回级别' }, ...['[DBG]', '[INF]', '[SEC]', '[WRN]', '[ERR]'].map((value) => ({ value, label: `仅显示 ${value}` }))
          ]} />
          <label>起始时间<Input type="datetime-local" step={1} aria-label="日志起始时间" value={start} onChange={(event) => setStart(event.target.value)} /></label>
          <label>结束时间<Input type="datetime-local" step={1} aria-label="日志结束时间" value={end} onChange={(event) => setEnd(event.target.value)} /></label>
          <Button disabled={!start && !end && !search && !priority} onClick={() => { setStart(''); setEnd(''); setSearch(''); setPriority('') }}>清除筛选</Button>
          <Typography.Text type="secondary">时间按浏览器本地时区输入。精确级别、时间和搜索仅筛选已读取的最近日志；下载使用相同筛选。若目标级别低于请求的最低级别，请先降低最低级别。</Typography.Text></>}
        </Space>
        {invalidRange && <Alert type="warning" message="起始时间不能晚于结束时间" />}
        <ResourceMetaBar observedAt={observed} />
        {error && observed && <Alert type="warning" message="刷新失败，以下为上次成功获取的日志。" />}
        <AppTable<ApiRecord> size="small" loading={loading && !rows.length} dataSource={filtered}
          expandable={{ expandedRowRender: (row) => <Descriptions size="small" bordered column={1} items={runtimeLogDetails(row)} style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }} /> }}
          rowKey={runtimeLogKey}
          pagination={{ defaultPageSize: 30, showSizeChanger: true }} columns={[
            { title: '时间', dataIndex: 'stamp', width: 230 },
            { title: '频道', dataIndex: 'channel', width: 100 },
            { title: '级别', dataIndex: 'priority', width: 85, render: (value) => <Tag color={String(value).includes('ERR') ? 'error' : String(value).includes('WRN') ? 'warning' : 'default'}>{String(value)}</Tag> },
            { title: '来源', dataIndex: 'name', width: 130 },
            { title: '消息', dataIndex: 'message', render: (value) => <Typography.Text style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{String(value)}</Typography.Text> }
          ]} />
      </Card>
    </Space>
  </>
  return compact ? <Card title="集群日志" style={{ marginTop: 16 }}>{content}</Card> : <Page title="运行日志" error={error}>{content}</Page>
}

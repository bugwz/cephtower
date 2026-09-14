import { DownloadOutlined, ReloadOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Input, Select, Space, Switch, Tag, Typography } from 'antd'
import { useEffect, useState } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'
import { AppTable } from '../../components/AppTable'
import { Page } from '../../components/Page'
import { ResourceMetaBar } from '../../components/ResourceMetaBar'
import { useClusterContext } from '../../state/ClusterContext'

export function RuntimeLogsPage() {
  const { selectedClusterId } = useClusterContext()
  const [channel, setChannel] = useState('cluster')
  const [level, setLevel] = useState('debug')
  const [limit, setLimit] = useState(100)
  const [auto, setAuto] = useState(true)
  const [revision, setRevision] = useState(0)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [rows, setRows] = useState<ApiRecord[]>([])
  const [observed, setObserved] = useState<string>()
  const [search, setSearch] = useState('')
  useEffect(() => {
    const abort = new AbortController()
    let timer: ReturnType<typeof setTimeout> | undefined
    setRows([]); setObserved(undefined); setError(''); setLoading(Boolean(selectedClusterId))
    async function read() {
      if (!selectedClusterId) return
      setLoading(true)
      try {
        const data = await request<{ items: ApiRecord[]; observed_at: string }>('/logs', jsonInit('GET', {
          cluster_id: selectedClusterId, channel, level, limit
        }, { signal: abort.signal, suppressErrorNotification: true }))
        if (!abort.signal.aborted) { setRows(data.items); setObserved(data.observed_at); setError('') }
      } catch (err) {
        if (!abort.signal.aborted) setError(err instanceof Error ? err.message : '读取日志失败')
      } finally {
        if (!abort.signal.aborted) {
          setLoading(false)
          if (auto) timer = setTimeout(read, 10000)
        }
      }
    }
    void read()
    return () => { abort.abort(); clearTimeout(timer) }
  }, [selectedClusterId, channel, level, limit, auto, revision])
  const filtered = rows.filter((row) => `${row.message} ${row.name} ${row.stamp}`.toLowerCase().includes(search.toLowerCase()))
  function download() {
    const content = filtered.map((row) => `${row.stamp} [${row.channel}] ${row.priority} ${row.name}: ${row.message}`).join('\n')
    const url = URL.createObjectURL(new Blob([content], { type: 'text/plain;charset=utf-8' }))
    const link = document.createElement('a'); link.href = url; link.download = `ceph-${channel === '*' ? 'all' : channel}.log`; link.click()
    setTimeout(() => URL.revokeObjectURL(url), 1000)
  }
  return <Page title="运行日志" error={error}>
    <Space direction="vertical" size={16} className="page-stack">
      <Alert type="info" showIcon message="显示 MON 日志缓冲区的最近记录。自动刷新每 10 秒读取一次；此视图不提供长期日志归档。" />
      <Card>
        <Space wrap>
          <Select aria-label="日志频道" value={channel} onChange={setChannel} style={{ width: 160 }} options={[
            { value: 'cluster', label: '集群日志' }, { value: 'audit', label: 'Ceph 审计日志' }, { value: 'cephadm', label: 'Cephadm' }, { value: '*', label: '全部频道' }
          ]} />
          <Select aria-label="最低级别" value={level} onChange={setLevel} style={{ width: 140 }} options={['debug', 'info', 'sec', 'warn', 'error'].map((value) => ({ value, label: `最低 ${value}` }))} />
          <Select aria-label="读取条数" value={limit} onChange={setLimit} options={[30, 100, 300, 500].map((value) => ({ value, label: `${value} 条` }))} />
          <Switch checked={auto} onChange={setAuto} checkedChildren="自动刷新" unCheckedChildren="已暂停" />
          <Button icon={<ReloadOutlined />} loading={loading} disabled={!selectedClusterId} onClick={() => setRevision((n) => n + 1)}>刷新</Button>
          <Button icon={<DownloadOutlined />} disabled={!filtered.length} onClick={download}>下载当前结果</Button>
          <Input.Search allowClear placeholder="搜索消息、来源或时间" value={search} onChange={(event) => setSearch(event.target.value)} />
        </Space>
        <ResourceMetaBar observedAt={observed} />
        {error && observed && <Alert type="warning" message="刷新失败，以下为上次成功获取的日志。" />}
        <AppTable<ApiRecord> size="small" loading={loading && !rows.length} dataSource={filtered}
          rowKey={(row) => `${row.channel}/${row.name}/${row.rank}/${row.stamp}/${row.seq}`}
          pagination={{ defaultPageSize: 30, showSizeChanger: true }} columns={[
            { title: '时间', dataIndex: 'stamp', width: 230 },
            { title: '频道', dataIndex: 'channel', width: 100 },
            { title: '级别', dataIndex: 'priority', width: 85, render: (value) => <Tag color={String(value).includes('ERR') ? 'error' : String(value).includes('WRN') ? 'warning' : 'default'}>{String(value)}</Tag> },
            { title: '来源', dataIndex: 'name', width: 130 },
            { title: '消息', dataIndex: 'message', render: (value) => <Typography.Text style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{String(value)}</Typography.Text> }
          ]} />
      </Card>
    </Space>
  </Page>
}

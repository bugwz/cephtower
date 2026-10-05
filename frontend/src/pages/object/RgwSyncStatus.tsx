import { Alert, Button, Card, Descriptions, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'
import { rgwSyncCounters, rgwSyncReportSections } from './rgwSyncReport'

export function RgwSyncCounterDetails({ section }: { section: string }) {
  const counters = rgwSyncCounters(section)
  if (!counters) return null
  const items = [
    { key: 'full', label: '全量同步阶段分片数（非已完成）', value: counters.full },
    { key: 'incremental', label: '增量同步阶段分片数', value: counters.incremental },
    { key: 'total', label: '报告分片总数', value: counters.total },
    { key: 'remaining', label: counters.remainingUnit === 'buckets' ? '待同步桶数' : '待同步条目数', value: counters.remaining },
    { key: 'behind', label: '落后分片数', value: counters.behind },
    { key: 'recovering', label: '恢复中分片数', value: counters.recovering },
    { key: 'oldestChange', label: '最早未应用增量变更（原始时间与时区）', value: counters.oldestChange },
    { key: 'oldestShard', label: '最早未应用增量变更所在分片', value: counters.oldestShard },
  ].filter(item => item.value !== undefined).map(({ key, label, value }) => ({ key, label, children: value }))
  return <Descriptions column={1} size="small" items={items} />
}

export function RgwSyncStatus({ row, clusterId }: { row: ApiRecord; clusterId?: number }) {
  const scope = JSON.stringify([clusterId, row.id, row.name, row.stale])
  const current = useRef(scope)
  current.current = scope
  const sequence = useRef(0)
  const locked = useRef(false)
  const mounted = useRef(true)
  const abort = useRef<AbortController>()
  const [state, setState] = useState({ scope, report: '', error: '', busy: false })
  const scoped = state.scope === scope
  const sections = scoped && state.report ? rgwSyncReportSections(state.report) : undefined
  const valid = !!clusterId && typeof row.id === 'string' && !!row.id && typeof row.name === 'string' && !!row.name && row.stale !== true
  useEffect(() => {
    mounted.current = true
    sequence.current++
    locked.current = false
    setState({ scope, report: '', error: '', busy: false })
    return () => { mounted.current = false; abort.current?.abort(); sequence.current++ }
  }, [scope])
  async function read() {
    if (!valid || !scoped || locked.current || !mounted.current || current.current !== scope) return
    locked.current = true
    abort.current?.abort()
    const controller = new AbortController()
    abort.current = controller
    const ticket = ++sequence.current
    setState({ scope, report: '', error: '', busy: true })
    try {
      const result = await request<{ report: string }>('/rgw/zone/sync/status', jsonInit('POST', { cluster_id: clusterId, zone_id: row.id, name: row.name }, { signal: controller.signal, cache: 'no-store', suppressErrorNotification: true }))
      if (!mounted.current || controller.signal.aborted || current.current !== scope || sequence.current !== ticket) return
      if (typeof result?.report !== 'string' || !result.report.trim() || result.report.length > 1048576) throw new Error('invalid report')
      setState({ scope, busy: false, report: result.report, error: '' })
    } catch {
      if (mounted.current && !controller.signal.aborted && current.current === scope && sequence.current === ticket) setState({ scope, busy: false, report: '', error: '同步报告读取失败或 Zone 身份发生变化。请检查本地 Zone 配置、集群访问权限并重新采集。' })
    } finally {
      if (mounted.current && current.current === scope && sequence.current === ticket) locked.current = false
    }
  }
  return <Card size="small" title="Multisite 同步状态">
    <Space direction="vertical" style={{ width: '100%' }}>
      <Alert type="info" message="按需读取此 Zone 的原生同步报告" description="展示 Realm、Zonegroup、元数据同步及各来源的数据分片进度、落后和恢复信息。报告读取成功不代表所有数据已同步；主 Zone 的 no sync 只表示不接收元数据同步。无来源的数据段不代表双向同步已完成。不会触发同步、修改配置或自动轮询。" />
      {!valid && <Alert type="warning" message="Zone 身份或库存状态不可用，请重新采集后再试" />}
      <Button disabled={!valid || !scoped || state.busy} loading={scoped && state.busy} onClick={() => void read()}>读取同步报告</Button>
      {scoped && state.error && <Alert type="error" message={state.error} />}
      {sections && <>
        <Card size="small" title="站点身份与原生采样信息"><pre style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{sections.identity}</pre></Card>
        <Card size="small" title="元数据同步"><RgwSyncCounterDetails section={sections.metadata}/><pre style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{sections.metadata}</pre></Card>
        {sections.sources.length === 0 && <Alert type="info" message="原生报告未列出数据同步来源；这不是同步完成的证明" />}
        {sections.sources.map((source, index) => <Card key={index} size="small" title={`数据同步来源 ${index + 1}`}><RgwSyncCounterDetails section={source}/><pre style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{source}</pre></Card>)}
      </>}
      {scoped && state.report && <details open={!sections}><summary>完整原生同步报告</summary><pre aria-label="Zone 原生同步报告" style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere', maxHeight: 520, overflow: 'auto' }}>{state.report}</pre></details>}
    </Space>
  </Card>
}

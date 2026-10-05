import { Alert, Button, Card, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'

export function RgwSyncStatus({ row, clusterId }: { row: ApiRecord; clusterId?: number }) {
  const scope = JSON.stringify([clusterId, row.id, row.name, row.stale])
  const current = useRef(scope)
  current.current = scope
  const sequence = useRef(0)
  const locked = useRef(false)
  const [state, setState] = useState({ scope, report: '', error: '', busy: false })
  const scoped = state.scope === scope
  const valid = !!clusterId && typeof row.id === 'string' && !!row.id && typeof row.name === 'string' && !!row.name && row.stale !== true
  useEffect(() => {
    sequence.current++
    locked.current = false
    setState({ scope, report: '', error: '', busy: false })
    return () => { sequence.current++ }
  }, [scope])
  async function read() {
    if (!valid || !scoped || locked.current) return
    locked.current = true
    const ticket = ++sequence.current
    setState({ scope, report: '', error: '', busy: true })
    try {
      const result = await request<{ report: string }>('/rgw/zone/sync/status', jsonInit('POST', { cluster_id: clusterId, zone_id: row.id, name: row.name }, { cache: 'no-store', suppressErrorNotification: true }))
      if (current.current !== scope || sequence.current !== ticket) return
      if (typeof result?.report !== 'string' || !result.report.trim() || result.report.length > 1048576) throw new Error('invalid report')
      setState({ scope, busy: false, report: result.report, error: '' })
    } catch {
      if (current.current === scope && sequence.current === ticket) setState({ scope, busy: false, report: '', error: '同步报告读取失败或 Zone 身份发生变化。请检查本地 Zone 配置、集群访问权限并重新采集。' })
    } finally {
      if (current.current === scope && sequence.current === ticket) locked.current = false
    }
  }
  return <Card size="small" title="Multisite 同步状态">
    <Space direction="vertical" style={{ width: '100%' }}>
      <Alert type="info" message="按需读取此 Zone 的原生同步报告" description="展示 Realm、Zonegroup、元数据同步及各来源的数据分片进度、落后和恢复信息。报告读取成功不代表所有数据已同步；主 Zone 的 no sync 只表示不接收元数据同步。无来源的数据段不代表双向同步已完成。不会触发同步、修改配置或自动轮询。" />
      {!valid && <Alert type="warning" message="Zone 身份或库存状态不可用，请重新采集后再试" />}
      <Button disabled={!valid || !scoped || state.busy} loading={scoped && state.busy} onClick={() => void read()}>读取同步报告</Button>
      {scoped && state.error && <Alert type="error" message={state.error} />}
      {scoped && state.report && <pre aria-label="Zone 原生同步报告" style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere', maxHeight: 520, overflow: 'auto' }}>{state.report}</pre>}
    </Space>
  </Card>
}

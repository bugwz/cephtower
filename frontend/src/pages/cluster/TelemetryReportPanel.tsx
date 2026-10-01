import { Alert, Button, Card, Select, Space, Typography } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { jsonInit, request } from '../../api/client'

interface Report { mode: string; report_json?: string; message?: string; observed_at: string }

export function TelemetryReportPanel({ clusterId }: { clusterId: number }) {
  const [mode, setMode] = useState<'current' | 'preview'>('preview')
  const [report, setReport] = useState<Report | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const pending = useRef<AbortController | null>(null)
  useEffect(() => () => { pending.current?.abort() }, [])
  async function load() {
    if (pending.current && !pending.current.signal.aborted) return
    const controller = new AbortController(); pending.current = controller
    setLoading(true); setReport(null); setError('')
    try {
      const value = await request<Report>('/manager/telemetry/report', jsonInit('GET', { cluster_id: clusterId, mode }, { signal: controller.signal, suppressErrorNotification: true }))
      if (controller.signal.aborted) return
      if (value.mode !== mode) throw new Error('报告模式与请求不一致')
      setReport(value)
    } catch (reason) { if (!controller.signal.aborted) setError(reason instanceof Error ? reason.message : '读取失败') }
    finally { if (!controller.signal.aborted) { pending.current = null; setLoading(false) } }
  }
  return <Card title="遥测报告预览">
    <Space direction="vertical" style={{ width: '100%' }}>
      <Alert type="info" showIcon message="仅生成本地样本，不启用遥测、不发送报告。报告可能包含集群与设备信息；常见凭据已脱敏。" description="当前集合使用 show-all；最新集合使用 preview-all，可能包含尚未同意的新数据集合。Ceph 返回无报告提示时不会自动切换模式。" />
      <Space wrap><Select value={mode} style={{ minWidth: 250 }} options={[{ value: 'preview', label: '最新可用集合（含设备信息）' }, { value: 'current', label: '当前已同意的集合' }]} onChange={(value) => { pending.current?.abort(); pending.current = null; setMode(value); setReport(null); setError(''); setLoading(false) }} /><Button loading={loading} onClick={load}>生成并读取样本</Button></Space>
      {error && <Alert type="error" message={error} />}
      {report && <>
        <Typography.Text type="secondary">生成时间：{report.observed_at}</Typography.Text>
        {report.message && <Alert type="info" message={report.message} />}
        {report.report_json !== undefined && <Typography.Paragraph copyable={{ text: report.report_json }}><pre style={{ maxHeight: 600, overflow: 'auto', whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{report.report_json}</pre></Typography.Paragraph>}
      </>}
    </Space>
  </Card>
}

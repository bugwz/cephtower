import { Alert, Button, Card, Descriptions, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { isRecord, type ApiRecord } from '../../api/client'
import { mutateResource } from '../../api/resource'

export function osdSafetyReport(value: unknown, osdId: string): ApiRecord | null {
  if (!isRecord(value) || typeof value.is_safe_to_destroy !== 'boolean') return null
  const seen = new Set<number>()
  for (const field of ['safe_to_destroy', 'active', 'missing_stats', 'stored_pgs']) {
    const ids = value[field]
    if (!Array.isArray(ids)) return null
    for (const id of ids) {
      if (!Number.isInteger(id) || id < 0 || id > 2147483647 || String(id) !== osdId || seen.has(id)) return null
      seen.add(id)
    }
  }
  if (value.is_safe_to_destroy !== ((value.safe_to_destroy as number[]).length === 1)) return null
  return value
}

export function OSDSafetyCheck({ clusterId, osdId }: { clusterId: number; osdId: string }) {
  const [report, setReport] = useState<ApiRecord | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [checkedAt, setCheckedAt] = useState('')
  const running = useRef(false)
  const scope = useRef<object | null>({})
  useEffect(() => { scope.current = {}; return () => { scope.current = null } }, [])
  async function check() {
    if (running.current || !scope.current) return
    const current = scope.current
    running.current = true
    setLoading(true); setError(''); setReport(null); setCheckedAt('')
    try {
      const result = await mutateResource('/osd/removal/check', 'POST', { cluster_id: clusterId, osd_ids: [osdId] })
      if (scope.current !== current) return
      const parsed = osdSafetyReport(isRecord(result.details) ? result.details.check : null, osdId)
      if (!parsed) throw new Error('安全检查响应无效，不能判断是否安全')
      setReport(parsed); setCheckedAt(new Date().toLocaleString())
    } catch (err) {
      if (scope.current === current) setError(err instanceof Error ? err.message : '检查失败，安全状态未知')
    } finally {
      running.current = false
      if (scope.current === current) setLoading(false)
    }
  }
  return <Card title={`集群 ${clusterId} / OSD ${osdId} 安全检查`}>
    <Space direction="vertical" style={{ width: '100%' }}>
      <Alert type="info" message="只读执行 ceph osd safe-to-destroy。结果仅代表检查当时，不会删除 OSD，也不替代编排器删除检查或后续操作确认。" />
      <Button loading={loading} disabled={loading} onClick={() => void check()}>执行安全检查</Button>
      {error && <Alert type="error" message={error} />}
      {report && <>
        <Alert type={report.is_safe_to_destroy === true ? 'success' : 'warning'} message={report.is_safe_to_destroy === true ? '本次原生检查通过' : '本次未确认安全，禁止据此推断可销毁'} description={`检查完成时间：${checkedAt}`} />
        <Descriptions bordered column={1}>{[['safe_to_destroy', '安全集合'], ['active', '仍承载 PG'], ['missing_stats', '缺少统计'], ['stored_pgs', '仍存储 PG']].map(([key, label]) => <Descriptions.Item key={key} label={label}>{(report[key] as number[]).join(', ') || '本次无此类目标'}</Descriptions.Item>)}</Descriptions>
      </>}
    </Space>
  </Card>
}

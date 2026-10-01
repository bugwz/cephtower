import { Alert, Button, Card, Descriptions, Input, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { mutateResource } from '../../api/resource'
import type { ApiRecord } from '../../api/client'
import { DataTable } from '../../components/DataTable'

export function upgradeCheckVersion(value: string) {
  const version = value.trim()
  if (!/^\d+\.\d+\.\d+$/.test(version)) throw new Error('请输入目标版本号，例如 20.2.2（不带 v 前缀）')
  return version
}

export function upgradeCheckData(details: unknown) {
  const object = (value: unknown): ApiRecord => {
    if (!value || typeof value !== 'object' || Array.isArray(value)) throw new Error('升级检查报告格式异常')
    return value as ApiRecord
  }
  const report = object(object(details).check)
  for (const field of ['target_name', 'target_id', 'target_version']) if (typeof report[field] !== 'string' || !report[field]) throw new Error('升级检查缺少目标信息')
  for (const field of ['up_to_date', 'non_ceph_image_daemons']) if (!Array.isArray(report[field]) || !(report[field] as unknown[]).every((value) => typeof value === 'string')) throw new Error('升级检查缺少守护进程列表')
  const rows = Object.entries(object(report.needs_update)).map(([name, value]) => {
    const current = object(value)
    for (const field of ['current_name', 'current_id', 'current_version']) if (!(field in current) || (current[field] !== null && typeof current[field] !== 'string')) throw new Error('升级检查守护进程信息异常')
    return { name, current_name: current.current_name, current_id: current.current_id, current_version: current.current_version }
  })
  return { report, rows }
}

export function UpgradeCheck({ clusterId }: { clusterId: number }) {
  const [version, setVersion] = useState('')
  const [result, setResult] = useState<ReturnType<typeof upgradeCheckData> | null>(null)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const active = useRef(true)
  const running = useRef(false)
  useEffect(() => { active.current = true; return () => { active.current = false } }, [])
  async function check() {
    if (running.current) return
    running.current = true; setBusy(true); setError(''); setResult(null)
    try {
      const target = upgradeCheckVersion(version)
      const response = await mutateResource('/upgrade/check', 'POST', { cluster_id: clusterId, version: target })
      if (active.current) setResult(upgradeCheckData(response.details))
    } catch (err) { if (active.current) setError(err instanceof Error ? err.message : '检查失败') }
    finally { running.current = false; if (active.current) setBusy(false) }
  }
  return <Card title="升级前检查" style={{ marginTop: 16 }}>
    <Space direction="vertical" style={{ width: '100%' }}>
      <Alert type="info" message="检查会查询目标容器镜像，可能需要较长时间；不会启动升级。结果仅代表检查时刻的镜像兼容性，不保证升级一定成功。" />
      <Space><Input aria-label="升级检查目标版本" placeholder="例如 20.2.2" value={version} disabled={busy} onChange={(event) => { setVersion(event.target.value); setResult(null); setError('') }} /><Button loading={busy} disabled={!version.trim()} onClick={() => void check()}>检查目标版本</Button></Space>
      {error && <Alert type="error" message={error} />}
      {result && <>
        <Descriptions column={1} items={[
          { key: 'target', label: '目标镜像', children: String(result.report.target_name) },
          { key: 'version', label: '目标版本', children: String(result.report.target_version) },
          { key: 'id', label: '镜像 ID', children: String(result.report.target_id) },
          { key: 'digest', label: '镜像摘要', children: typeof result.report.target_digest === 'string' ? result.report.target_digest : '未提供' },
          { key: 'current', label: '已是目标版本', children: (result.report.up_to_date as string[]).join('、') || '无' },
          { key: 'other', label: '非 Ceph 镜像守护进程', children: (result.report.non_ceph_image_daemons as string[]).join('、') || '无' }
        ]} />
        <Card size="small" title={`待升级守护进程（${result.rows.length}）`}><DataTable data={result.rows} columns={[{ key: 'name', title: '守护进程' }, { key: 'current_name', title: '当前镜像' }, { key: 'current_id', title: '当前镜像 ID' }, { key: 'current_version', title: '当前版本' }]} /></Card>
      </>}
    </Space>
  </Card>
}

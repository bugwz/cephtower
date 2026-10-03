import { Alert, Button, Card, Space, Table, Typography } from 'antd'
import { useCallback } from 'react'
import { jsonInit, request } from '../../api/client'
import { useResource } from '../../hooks'

export function scheduledImageRows(value: unknown) {
  if (!Array.isArray(value)) return undefined
  const rows: { key: number; image: string; schedule_time: string }[] = []
  for (const [key, item] of value.entries()) {
    if (!item || typeof item.image !== 'string' || !item.image.trim() || typeof item.schedule_time !== 'string' || !item.schedule_time.trim()) return undefined
    rows.push({ key, image: item.image, schedule_time: item.schedule_time })
  }
  return rows
}

export function LiveMirrorScheduleStatus({ clusterId }: { clusterId: number }) {
  const loader = useCallback(async () => {
    const result = await request<{ scheduled_images: unknown; observed_at: string }>('/rbd/mirroring/schedule/status', jsonInit('GET', { cluster_id: clusterId }))
    const rows = scheduledImageRows(result.scheduled_images)
    if (!rows || typeof result.observed_at !== 'string') throw new Error('调度运行状态响应不完整')
    return { rows, observedAt: result.observed_at }
  }, [clusterId])
  const { data, loading, error, refresh } = useResource(loader)
  return <Card title="镜像快照待执行任务（实时读取）" extra={<Button loading={loading} onClick={() => refresh()}>重读运行状态</Button>}>
    <Space direction="vertical" style={{ width: '100%' }}>
      <Alert type="info" message="读取 rbd mirror snapshot schedule status；以下为调度器返回的待执行任务，不是完成记录，也不证明镜像正在同步。本页操作成功后自动重读，外部变化请手动重读。" />
      {error && <Alert type="error" message={error} description={data ? '下方为上次成功读取结果，不代表当前运行状态。' : '无法读取运行状态，不能据此判断没有待执行任务。'} />}
      {data && <><Typography.Text type="secondary">读取时间：{data.observedAt}</Typography.Text><Table size="small" rowKey="key" dataSource={data.rows} pagination={{ pageSize: 10 }} locale={{ emptyText: '本次读取未返回待执行任务' }} columns={[
        { title: '镜像路径', dataIndex: 'image' },
        { title: '计划执行时间（命令原值）', dataIndex: 'schedule_time' }
      ]} /></>}
      {!data && !error && <Typography.Text type="secondary">正在读取运行状态…</Typography.Text>}
    </Space>
  </Card>
}

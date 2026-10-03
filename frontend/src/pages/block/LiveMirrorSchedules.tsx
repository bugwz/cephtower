import { Alert, Button, Card, Space, Typography } from 'antd'
import { useCallback } from 'react'
import { jsonInit, request } from '../../api/client'
import { useResource } from '../../hooks'
import { MirrorSchedules } from './MirrorSchedules'
import { mirrorScheduleRows } from './rbdMirrorScheduleRows'

export function LiveMirrorSchedules({ clusterId }: { clusterId: number }) {
  const loader = useCallback(async () => {
    const result = await request<{ schedules: unknown; observed_at: string }>('/rbd/mirroring/schedules', jsonInit('GET', { cluster_id: clusterId }))
    if (!mirrorScheduleRows(result.schedules) || typeof result.observed_at !== 'string') throw new Error('调度响应不完整')
    return result
  }, [clusterId])
  const { data, loading, error, refresh } = useResource(loader)
  return <Card title="全范围镜像快照调度（实时读取）" extra={<Button loading={loading} onClick={() => refresh()}>重读调度</Button>}>
    <Space direction="vertical" style={{ width: '100%' }}>
      <Alert type="info" message="直接读取 rbd mirror snapshot schedule list --recursive，不依赖池库存。显示集群、池、命名空间和镜像范围；本页操作成功后自动重读，外部变更请手动重读。配置不代表同步正在运行。" />
      {error && <Alert type="error" message={error} description={data ? '下方为上次成功读取结果，不代表当前配置。' : '无法读取调度，不能据此判断未配置。'} />}
      {data && <><Typography.Text type="secondary">读取时间：{data.observed_at}</Typography.Text><MirrorSchedules value={data.schedules} status="available" /></>}
      {!data && !error && <Typography.Text type="secondary">正在读取调度…</Typography.Text>}
    </Space>
  </Card>
}

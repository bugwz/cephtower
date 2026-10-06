import { Alert, Button, Descriptions, Space } from 'antd'
import { useCallback } from 'react'
import { isRecord } from '../../api/client'
import { getResource, refreshResource } from '../../api/resource'
import { useResource } from '../../hooks'

export function osdCapacityThreshold(value: unknown): string {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0 && value <= 1 ? `${value}（约 ${Number((value * 100).toFixed(4))}%）` : '未采集或格式无效'
}

export function OSDCapacityThresholds({ clusterId }: { clusterId: number }) {
  const loader = useCallback(async () => {
    await refreshResource({ clusterId, kinds: ['osd_flag'] })
    const result = await getResource('/osd/flag', clusterId)
    if (!isRecord(result.item) || !isRecord(result.item.data)) throw new Error('OSD 容量阈值响应无效')
    return result.item
  }, [clusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const values = data && isRecord(data.data) ? data.data : {}
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Alert type="info" message="OSD 容量阈值（只读）" description="来自 ceph osd dump 的集群配置比例，不是当前 OSD 利用率。Nearfull 预警容量压力，Backfillfull 限制回填，Full 限制写入；缺失阈值不使用默认值替代。" />
    <Button loading={loading} disabled={loading} onClick={() => refresh()}>重新采集容量阈值</Button>
    {error && <Alert type="error" message={error} description="下方若有数据，为上次成功读取的快照。" />}
    {data && data.stale !== false && <Alert type="warning" message="阈值库存已过期或新鲜度未知。" />}
    <Descriptions bordered column={3}>
      <Descriptions.Item label="Nearfull">{osdCapacityThreshold(values.nearfull_ratio)}</Descriptions.Item>
      <Descriptions.Item label="Backfillfull">{osdCapacityThreshold(values.backfillfull_ratio)}</Descriptions.Item>
      <Descriptions.Item label="Full">{osdCapacityThreshold(values.full_ratio)}</Descriptions.Item>
    </Descriptions>
  </Space>
}

import { Alert, Button, Space, Typography } from 'antd'
import { useCallback } from 'react'
import { jsonInit, request } from '../../api/client'
import { useResource } from '../../hooks'

export function HostHardwareFullReport({ clusterId, host }: { clusterId: number; host: string }) {
  const loader = useCallback(async () => {
    const result = await request<{ host: string; category: string; serial_number?: string | null; report: string; observed_at: string }>('/host/hardware', jsonInit('GET', { cluster_id: clusterId, host, category: 'fullreport' }))
    if (result.host !== host || result.category !== 'fullreport' || typeof result.report !== 'string' || result.report.trim() === '') throw new Error('完整硬件报告与所选主机不匹配或数据无效')
    if (result.serial_number != null && typeof result.serial_number !== 'string') throw new Error('完整硬件报告序列号类型无效')
    return result
  }, [clusterId, host])
  const { data, loading, error, refresh } = useResource(loader)
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Alert type="info" message="node-proxy 完整报告（脱敏）" description="按原始 JSON 文本保留精确数值，包含采集到的硬件属性和固件清单。数据来自 node-proxy 缓存，不证明覆盖所有硬件或反映实时状态；本入口不执行配置或升级操作。" />
    <Button loading={loading} onClick={() => refresh()}>重新读取完整报告</Button>
    {error && <Alert type="error" message={error} description={data ? '下方为上次成功读取的报告，不代表当前状态。' : '请检查 node-proxy、编排器支持和访问权限。'} />}
    {data && <>
      <Typography.Text>主机：{data.host}；序列号：{data.serial_number || '未返回'}；读取时间：{data.observed_at}</Typography.Text>
      <Typography.Paragraph copyable={{ text: data.report }}>复制脱敏报告</Typography.Paragraph>
      <pre style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere', maxHeight: 640, overflowY: 'auto' }}>{data.report}</pre>
    </>}
  </Space>
}

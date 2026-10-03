import { Alert, Button, Descriptions, Space, Tag } from 'antd'
import { useCallback } from 'react'
import { jsonInit, request } from '../../api/client'
import { useResource } from '../../hooks'

interface Info { manager: string; plugins: string[]; directory: string; observed_at: string }

export function ErasureCodeInfo({ clusterId }: { clusterId: number }) {
  const loader = useCallback(async () => {
    const result = await request<Info>('/erasure/code/info', jsonInit('GET', { cluster_id: clusterId }))
    if (typeof result.manager !== 'string' || !result.manager.startsWith('mgr.') || !Array.isArray(result.plugins) || !result.plugins.length || result.plugins.some((value) => typeof value !== 'string' || !value.trim()) || typeof result.directory !== 'string' || !result.directory.trim() || typeof result.observed_at !== 'string') throw new Error('EC 运行配置响应不完整')
    return result
  }, [clusterId])
  const { data, error, loading, refresh } = useResource(loader)
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Alert type="info" message="活动 MGR 的 EC 运行配置（只读）" description="来自 config show-with-defaults，不等于各 OSD 已安装插件清单。下方插件为 osd_erasure_code_plugins 原值；参考界面还提供实验性 SHEC / CLAY。此信息不会自动覆盖表单插件或目录。" />
    <Button size="small" loading={loading} onClick={() => refresh()}>重新读取 EC 运行配置</Button>
    {error && <Alert type="error" message={error} description={data ? '下方保留上次读取结果，不代表当前状态。' : '无法确认集群插件配置和目录；没有使用本地默认值替代。'} />}
    {data && <Descriptions size="small" column={1} items={[
      { key: 'manager', label: '读取目标', children: data.manager },
      { key: 'plugins', label: '配置插件', children: <Space wrap>{data.plugins.map((plugin) => <Tag key={plugin}>{plugin}</Tag>)}</Space> },
      { key: 'directory', label: '插件目录', children: data.directory },
      { key: 'observed_at', label: '读取时间', children: data.observed_at },
    ]} />}
  </Space>
}

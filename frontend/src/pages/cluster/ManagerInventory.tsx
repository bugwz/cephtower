import { Alert, Button, Space, Typography } from 'antd'
import { useCallback, useEffect, useRef, useState } from 'react'
import type { ApiRecord } from '../../api/client'
import { listAllResources, refreshResource } from '../../api/resource'
import { AppTable } from '../../components/AppTable'
import { useResource } from '../../hooks'
import { message } from '../../utils/appMessage'

export function managerServices(value: unknown): { name: string; uri: string }[] | null {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return null
  const entries = Object.entries(value)
  if (entries.some(([name, uri]) => !name || typeof uri !== 'string')) return null
  return entries.map(([name, uri]) => ({ name, uri: uri as string }))
}

export function ManagerInventory({ clusterId }: { clusterId: number }) {
  const active = useRef(true)
  const running = useRef(false)
  const [collecting, setCollecting] = useState(false)
  const [collectionError, setCollectionError] = useState('')
  useEffect(() => {
    active.current = true
    return () => { active.current = false }
  }, [])
  const loader = useCallback(() => listAllResources('/managers', clusterId), [clusterId])
  const { data, loading, error, refresh } = useResource(loader)
  async function collect() {
    if (!active.current || running.current) return
    running.current = true
    setCollecting(true); setCollectionError('')
    try {
      await refreshResource({ clusterId, kind: 'mgr' })
      if (!active.current) return
      message.success('MGR 采集完成，正在重新读取库存')
      await refresh()
    } catch (err) {
      if (active.current) setCollectionError(err instanceof Error ? err.message : 'MGR 采集失败')
    } finally {
      running.current = false
      if (active.current) setCollecting(false)
    }
  }
  return <Space direction="vertical" className="page-stack">
    <Alert type="info" message="来自 ceph mgr dump 的采集快照，不代表实时状态。服务 URI 由活动 MGR 发布，仅展示原文，不探测可达性；备用 MGR 不继承活动实例的服务。" />
    <Space>
      <Button loading={loading} disabled={collecting} onClick={() => void refresh()}>重新读取库存</Button>
      <Button loading={collecting} disabled={loading} onClick={() => void collect()}>重新采集 MGR</Button>
    </Space>
    {collectionError && <Alert type="error" message={`MGR 采集失败：${collectionError}`} />}
    {error && <Alert type="error" message={error} />}
    {data && (data.stale !== false || data.items.some(row => row.stale !== false)) && <Alert type="warning" message="MGR 库存已过期或新鲜度未知，请重新采集集群。" />}
    <AppTable<ApiRecord> dataSource={data?.items ?? []} loading={loading} rowKey="natural_key" pagination={{ defaultPageSize: 10 }} columns={[
      { title: '名称', dataIndex: 'name' },
      { title: '角色（采集时）', dataIndex: 'active', render: value => value === true ? '活动' : value === false ? '备用' : '未知' },
      { title: '活动 MGR 可用（原生）', dataIndex: 'available', render: value => value === true ? '是' : value === false ? '否' : '未知' },
      { title: '地址', dataIndex: 'address' },
      { title: '采集时间', dataIndex: 'observed_at' }
    ]} expandable={{ expandedRowRender: row => {
      if (row.active !== true) return <span>此记录不是活动 MGR，未关联活动实例的服务 URI。</span>
      const services = managerServices(row.services)
      if (!services) return <Alert type="warning" message="未取得有效的模块服务映射，不能认定没有服务。" />
      return <AppTable dataSource={services} rowKey="name" size="small" pagination={{ defaultPageSize: 5 }} locale={{ emptyText: '本次采集未发布模块服务' }} columns={[
        { title: '模块', dataIndex: 'name' },
        { title: '原生 URI（未探测）', dataIndex: 'uri', render: value => <Typography.Text copyable style={{ overflowWrap: 'anywhere' }}>{value || '空字符串'}</Typography.Text> }
      ]} />
    } }} />
  </Space>
}

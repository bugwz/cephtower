import { Alert, Card, Input, Space } from 'antd'
import { useEffect, useState } from 'react'
import { listAllResources, type ResourceListResult } from '../../api/resource'
import type { ApiRecord } from '../../api/client'
import { DataTable } from '../../components/DataTable'
import { ResourceMetaBar } from '../../components/ResourceMetaBar'

export function upgradeDaemonRows(items: ApiRecord[], search: string) {
  const types = new Set(['mgr', 'mon', 'crash', 'osd', 'mds', 'rgw', 'rbd-mirror', 'cephfs-mirror', 'iscsi', 'nfs'])
  const text = (value: unknown) => typeof value === 'string' && value.trim() ? value : '未知'
  const digests = (value: unknown) => Array.isArray(value) && value.every(item => typeof item === 'string' && item.trim() !== '') ? (value.length ? value.join('、') : '未报告摘要') : '未知'
  const query = search.trim().toLowerCase()
  return items.filter((item) => typeof item.type === 'string' && types.has(item.type)).map((item) => ({
    natural_key: item.natural_key,
    name: text(item.name), hostname: text(item.hostname), version: text(item.version),
    image: text(item.container_image), image_id: text(item.container_image_id), image_digests: digests(item.container_image_digests), freshness: item.stale === true ? '已过期' : item.stale === false ? '有效' : '未知'
  })).filter((item) => !query || [item.name, item.hostname, item.version, item.image, item.image_id, item.image_digests].some((value) => value.toLowerCase().includes(query)))
}

export function UpgradeDaemons({ clusterId, revision }: { clusterId: number; revision: number }) {
  const [data, setData] = useState<ResourceListResult | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [search, setSearch] = useState('')
  useEffect(() => {
    let active = true
    setData(null); setError(''); setLoading(true)
    void listAllResources('/daemons', clusterId)
      .then((result) => { if (active) setData(result) })
      .catch((err) => { if (active) setError(err instanceof Error ? err.message : '守护进程版本读取失败') })
      .finally(() => { if (active) setLoading(false) })
    return () => { active = false }
  }, [clusterId, revision])
  return <Card title="守护进程版本" loading={loading} style={{ marginTop: 16 }}>
    <Space direction="vertical" style={{ width: '100%' }}>
      {error && <Alert type="error" message={error} />}
      {data && <>
        <Alert type="info" message="仅展示 Dashboard 升级范围内的 mgr、mon、crash、osd、mds、rgw、rbd-mirror、cephfs-mirror、iscsi、nfs。其他类型及缺少类型的库存不纳入此版本表；可在服务页检查全部守护进程。" />
        <ResourceMetaBar observedAt={data.observedAt} stale={data.stale} staleReason={data.staleReason} />
        {data.stale && <Alert type="warning" message="守护进程库存已过期，请重新采集后判断版本状态。" />}
        <Alert type="info" message="镜像名称、ID 和摘要来自守护进程库存。相同标签不保证镜像内容相同；未报告摘要不代表镜像一致，也不代表升级已完成。" />
        <Input.Search aria-label="搜索守护进程版本" placeholder="按名称、主机、版本或镜像信息搜索" value={search} onChange={(event) => setSearch(event.target.value)} allowClear />
        <DataTable rowKeyCandidates={['natural_key']} data={upgradeDaemonRows(data.items, search)} columns={[{ key: 'name', title: '守护进程' }, { key: 'hostname', title: '主机' }, { key: 'version', title: '版本' }, { key: 'image', title: '容器镜像' }, { key: 'image_id', title: '镜像 ID', ellipsis: false }, { key: 'image_digests', title: '镜像摘要', ellipsis: false }, { key: 'freshness', title: '采集状态' }]} />
      </>}
    </Space>
  </Card>
}

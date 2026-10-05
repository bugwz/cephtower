import { Alert, Button, Select, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import type { ApiRecord } from '../../api/client'
import { useClusterContext } from '../../state/ClusterContext'
import { ExternalListPage, type ExternalListPageDefinition } from '../ExternalListPage'
import { rgwBucketConfigurationReadOptions } from './rgwBucketConfiguration'

export function RgwBucketConfigurationPanel({ row, clusterId, definition }: { row: ApiRecord; clusterId?: number; definition: ExternalListPageDefinition }) {
  const { selectedClusterId } = useClusterContext()
  const id = row.natural_key
  if (!clusterId || clusterId !== selectedClusterId || row.stale === true || typeof id !== 'string' || !/^[A-Za-z0-9_-]+$/.test(id)) {
    return <Alert type="warning" message="Bucket 身份或当前集群范围不可用，请刷新库存" />
  }
  return <RgwBucketConfigurationView key={`${clusterId}:${id}`} bucketId={id} definition={definition} />
}

export function RgwBucketConfigurationView({ bucketId, definition }: { bucketId: string; definition: ExternalListPageDefinition }) {
  const [kind, setKind] = useState('policy')
  const [active, setActive] = useState(false)
  const currentKind = useRef(kind)
  currentKind.current = kind
  const mounted = useRef(true)
  useEffect(() => { mounted.current = true; return () => { mounted.current = false } }, [])
  const current = () => mounted.current && currentKind.current === kind
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Alert type="info" message={`当前 Bucket ID：${bucketId}`} description="选择配置后显式读取，使用当前集群配置的 S3 端点与凭据。不会自动读取所有配置。切换配置会关闭旧编辑表单；写入沿用权限检查、风险确认及回读核验。" />
    <Select aria-label="Bucket 配置类型" style={{ width: 320, maxWidth: '100%' }} value={kind} options={rgwBucketConfigurationReadOptions} onChange={value => {
      if (!current() || !rgwBucketConfigurationReadOptions.some(option => option.value === value)) return
      currentKind.current = value
      setKind(value); setActive(false)
    }} />
    <Button disabled={active} onClick={() => { if (current()) setActive(true) }}>读取此 Bucket 配置</Button>
    {active && <ExternalListPage key={kind} embedded definition={{
      ...definition,
      body: { bucket_id: bucketId, kind },
      filterFields: [],
      createAction: undefined,
    }} />}
  </Space>
}

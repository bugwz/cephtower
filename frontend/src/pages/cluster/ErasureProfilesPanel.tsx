import { Card } from 'antd'
import type { ApiRecord } from '../../api/client'
import { ResourceListPage, type ResourceListPageDefinition } from '../ResourceListPage'

export const erasureProfileColumns = [
  { key: 'name', title: '配置名称' },
  { key: 'plugin', title: '插件' },
  { key: 'k', title: '数据分片 k' },
  { key: 'm', title: '编码分片 m' },
  { key: 'technique', title: '编码算法' },
  { key: 'crush-root', title: 'CRUSH 根节点' },
  { key: 'crush-failure-domain', title: '故障域' },
  { key: 'crush-device-class', title: '设备类别' }
]

export function erasureProfileDeleteBlocked(row: ApiRecord) {
  if (row.stale !== false) return '配置库存过期或状态未知，请重新采集'
  if (typeof row.name !== 'string' || !row.name.trim()) return '配置名称不可用'
  return undefined
}

const definition: ResourceListPageDefinition = {
  title: '纠删码配置',
  path: '/erasure/code/profiles',
  rowKeyCandidates: ['natural_key', 'name'],
  columns: erasureProfileColumns,
  deleteAction: {
    title: '删除纠删码配置', path: '/erasure/code/profile', action: 'erasure_code_profile.delete', resourceKind: 'erasure_code_profile', risk: 'high',
    confirmation: (row) => `确认删除配置 ${row.name}？此操作不可撤销。Ceph 会拒绝删除仍被存储池使用的配置；本操作不会迁移存储池或删除数据。`,
    successMessage: '纠删码配置删除已核验', disabledWhen: erasureProfileDeleteBlocked,
    buildBody: (row, clusterId) => ({ cluster_id: clusterId, name: row.name }),
    resourceKey: (row) => `erasure-code-profile/${row.name}`
  }
}

export function ErasureProfilesPanel() {
  return <Card style={{ marginTop: 16 }}><ResourceListPage definition={definition} embedded /></Card>
}

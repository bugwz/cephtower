import { Card } from 'antd'
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

const definition: ResourceListPageDefinition = {
  title: '纠删码配置',
  path: '/erasure/code/profiles',
  rowKeyCandidates: ['natural_key', 'name'],
  columns: erasureProfileColumns
}

export function ErasureProfilesPanel() {
  return <Card style={{ marginTop: 16 }}><ResourceListPage definition={definition} embedded /></Card>
}

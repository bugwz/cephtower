import { Alert, Table } from 'antd'
import type { ApiRecord } from '../../api/client'

export function RgwZonePoolReferences({ row }: { row: ApiRecord }) {
  const refs = row.pool_references
  const valid = Array.isArray(refs) && refs.every(ref => ref && typeof ref === 'object' && typeof ref.field === 'string' && typeof ref.pool === 'string' && !!ref.pool && typeof ref.namespace === 'string' && typeof ref.raw === 'string') && new Set(refs.map(ref => ref.field)).size === refs.length
  if (!valid) return <p>池引用明细未采集或格式异常，请重新采集 Zone 配置</p>
  return <div>
    <p>根据原生 Zone 配置精确解析，不按名称包含关系匹配。仅表示配置引用，不证明池存在、未被其他 Zone/业务共享或可删除；空命名空间与整个池不同。云分层目标、动态创建的池和其他配置不在此清单的完整性声明范围内。</p>
    {row.pool_references_complete !== true && <Alert type="warning" message="部分原生配置缺失或无法解析，以下不是完整引用清单" />}
    {Array.isArray(row.pool_reference_issues) && row.pool_reference_issues.every(issue => typeof issue === 'string') && row.pool_reference_issues.length > 0 && <details><summary>查看解析问题</summary><pre>{row.pool_reference_issues.join('\n')}</pre></details>}
    <Table size="small" rowKey="field" dataSource={refs} pagination={refs.length > 10 ? { pageSize: 10 } : false} scroll={{ x: 700 }} locale={{ emptyText: row.pool_references_complete === true ? '已识别字段中无非空池引用' : '未提取出有效引用，不代表没有引用' }} columns={[
      { title: '原生字段 / 放置规则 / 存储类别', dataIndex: 'field' },
      { title: '精确池名', dataIndex: 'pool' },
      { title: '命名空间', dataIndex: 'namespace', render: (value: string) => value === '' ? '默认命名空间（空）' : value },
      { title: '原生转义表示', dataIndex: 'raw' }
    ]} />
  </div>
}

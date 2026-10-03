import { Table } from 'antd'

export function bucketCorsRows(value: unknown) {
  if (!Array.isArray(value)) return undefined
  if (value.some(rule => !rule || typeof rule !== 'object' || typeof rule.id !== 'string'
    || ['allowed_origins', 'allowed_methods', 'allowed_headers', 'expose_headers'].some(key => !Array.isArray(rule[key]) || rule[key].some((item: unknown) => typeof item !== 'string'))
    || (rule.max_age_seconds !== null && (!Number.isInteger(rule.max_age_seconds) || rule.max_age_seconds < 0 || rule.max_age_seconds > 4294967294)))) return undefined
  return value.map((rule, index) => ({ ...rule, index }))
}

export function RgwBucketCorsRules({ value, configured }: { value: unknown; configured: unknown }) {
  const rules = bucketCorsRows(value)
  if (!rules || typeof configured !== 'boolean' || (configured && rules.length === 0) || (!configured && rules.length !== 0)) return <span>CORS 数据不可用</span>
  if (!configured) return <span>未配置 CORS</span>
  const list = (values: string[]) => values.length ? <span style={{ whiteSpace: 'pre-wrap' }}>{values.map(item => JSON.stringify(item)).join('\n')}</span> : '（空列表）'
  return <Table size="small" rowKey="index" dataSource={rules} pagination={rules.length > 5 ? { pageSize: 5 } : false} scroll={{ x: 700 }} columns={[
    { title: '顺序', dataIndex: 'index', render: (index: number) => index + 1 },
    { title: '规则 ID', dataIndex: 'id', render: (id: string) => JSON.stringify(id) },
    { title: '允许来源', dataIndex: 'allowed_origins', render: list },
    { title: '允许方法', dataIndex: 'allowed_methods', render: list },
    { title: '允许头', dataIndex: 'allowed_headers', render: list },
    { title: '暴露头（原顺序）', dataIndex: 'expose_headers', render: list },
    { title: '缓存秒数', dataIndex: 'max_age_seconds', render: (value: number | null) => value === null ? '未设置' : `${value} 秒` }
  ]} />
}

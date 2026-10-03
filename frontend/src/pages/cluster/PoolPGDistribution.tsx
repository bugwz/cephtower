import { Alert, Card, Space, Tag, Typography } from 'antd'
import { AppTable } from '../../components/AppTable'
import { poolPGDistribution } from '../overview/pgCategory'

export function PoolPGDistribution({ value }: { value: unknown }) {
  const distribution = poolPGDistribution(value)
  return <Card className="page-surface-card" title="PG 状态分布（采集时）">
    <Space direction="vertical" style={{ width: '100%' }}>
      <Typography.Text type="secondary">来源：ceph pg dump pgs_brief，按当前池 ID 汇总。仅表示采集快照，不代表实时状态；占比分母为本次返回的池内 PG 总数。</Typography.Text>
      {!distribution ? <Alert type="warning" message="PG 状态分布未采集或计数无效，不能据此判断健康。" /> : <>
        <Space wrap>
          <Typography.Text>返回 PG 总数：{distribution.total}</Typography.Text>
          {distribution.categories.map((category) => <Tag key={category.key} color={category.color}>{category.label}：{category.count}</Tag>)}
        </Space>
        <AppTable size="small" rowKey="name" dataSource={distribution.rows} pagination={false} columns={[
          { title: '原始状态', dataIndex: 'name' },
          { title: '分类', key: 'category', filters: distribution.categories.map((category) => ({ text: category.label, value: category.key })), onFilter: (value, row) => row.category.key === value, render: (_, row) => <Tag color={row.category.color}>{row.category.label}</Tag> },
          { title: 'PG 数量', dataIndex: 'count', sorter: (left, right) => left.count - right.count },
          { title: '池内占比', key: 'percentage', render: (_, row) => distribution.total ? `${(row.count / distribution.total * 100).toFixed(1)}%` : '—' }
        ]} />
      </>}
    </Space>
  </Card>
}

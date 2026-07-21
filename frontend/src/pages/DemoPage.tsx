import { ApiOutlined, CheckCircleOutlined, ClockCircleOutlined, DeploymentUnitOutlined } from '@ant-design/icons'
import { Button, Card, Progress, Table, Tag, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { findNavPage, findNavSection, type PageKey } from '../navigation'
import { Page } from '../components/Page'

const { Text } = Typography

type DemoRecord = {
  key: string
  name: string
  status: '运行中' | '就绪' | '待配置'
  owner: string
  updatedAt: string
}

const columns: ColumnsType<DemoRecord> = [
  {
    title: '名称',
    dataIndex: 'name'
  },
  {
    title: '状态',
    dataIndex: 'status',
    render: (status: DemoRecord['status']) => {
      const color = status === '运行中' ? 'green' : status === '就绪' ? 'blue' : 'gold'
      return <Tag color={color}>{status}</Tag>
    }
  },
  {
    title: '归属',
    dataIndex: 'owner'
  },
  {
    title: '更新时间',
    dataIndex: 'updatedAt'
  }
]

export function DemoPage({ pageKey }: { pageKey: PageKey }) {
  const page = findNavPage(pageKey)
  const section = findNavSection(pageKey)
  const title = page?.label ?? '功能演示'
  const sectionLabel = section?.label ?? 'CephTower'
  const seed = pageKey.length + title.length
  const completion = 56 + (seed % 33)
  const records = buildDemoRecords(title, sectionLabel)

  return (
    <Page title={title} description={`${sectionLabel} / ${title} · demo 页面`}>
      <div className="demo-actionbar">
        <Button type="primary" icon={<CheckCircleOutlined />}>
          新建
        </Button>
        <Button icon={<ClockCircleOutlined />}>刷新</Button>
      </div>

      <div className="metrics-grid demo-metrics">
        <Card>
          <Text type="secondary" className="metric-label">
            资源数量
          </Text>
          <div className="demo-stat-value">{8 + (seed % 17)}</div>
          <Text type="secondary">{sectionLabel} 当前 demo 资源</Text>
        </Card>
        <Card>
          <Text type="secondary" className="metric-label">
            可用率
          </Text>
          <div className="demo-stat-value">{completion}%</div>
          <Progress percent={completion} showInfo={false} strokeColor="#23933f" />
        </Card>
        <Card>
          <Text type="secondary" className="metric-label">
            接入状态
          </Text>
          <div className="demo-inline-status">
            <DeploymentUnitOutlined />
            <Text strong>已接入</Text>
          </div>
          <Text type="secondary">等待真实 API 对接</Text>
        </Card>
        <Card>
          <Text type="secondary" className="metric-label">
            API 范围
          </Text>
          <div className="demo-inline-status">
            <ApiOutlined />
            <Text strong>/api/v1</Text>
          </div>
          <Text type="secondary">按模块补充请求封装</Text>
        </Card>
      </div>

      <Card
        title={
          <div>
            <Text strong>{title}列表</Text>
            <Text type="secondary" className="card-subtitle">
              demo 数据，用于验证导航与页面结构
            </Text>
          </div>
        }
      >
        <Table columns={columns} dataSource={records} pagination={false} size="middle" />
      </Card>
    </Page>
  )
}

function buildDemoRecords(title: string, sectionLabel: string): DemoRecord[] {
  return ['主实例', '备用实例', '策略配置'].map((suffix, index) => ({
    key: `${title}-${suffix}`,
    name: `${title}-${suffix}`,
    status: index === 0 ? '运行中' : index === 1 ? '就绪' : '待配置',
    owner: sectionLabel,
    updatedAt: `2026-07-${21 - index} 10:${20 + index * 7}`
  }))
}

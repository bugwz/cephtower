import {
  AppstoreOutlined,
  BarChartOutlined,
  CloudServerOutlined,
  DatabaseOutlined
} from '@ant-design/icons'
import { Alert, Card, ConfigProvider, Layout, Menu, Spin, Statistic, theme, Typography } from 'antd'
import { useEffect, useState } from 'react'
import { getClusterSummary, type ClusterSummary } from './api/cluster'

const { Content, Sider } = Layout
const { Paragraph, Title, Text } = Typography

const navItems = [
  { key: 'overview', icon: <AppstoreOutlined />, label: '集群概览' },
  { key: 'pools', icon: <DatabaseOutlined />, label: '存储池' },
  { key: 'osd', icon: <CloudServerOutlined />, label: 'OSD' },
  { key: 'monitoring', icon: <BarChartOutlined />, label: '监控' }
]

export default function App() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [summary, setSummary] = useState<ClusterSummary | null>(null)

  useEffect(() => {
    let ignore = false

    async function loadSummary() {
      try {
        const data = await getClusterSummary()
        if (!ignore) {
          setSummary(data)
        }
      } catch (err) {
        if (!ignore) {
          setError(err instanceof Error ? err.message : '请求集群摘要失败')
        }
      } finally {
        if (!ignore) {
          setLoading(false)
        }
      }
    }

    loadSummary()

    return () => {
      ignore = true
    }
  }, [])

  return (
    <ConfigProvider
      theme={{
        algorithm: theme.defaultAlgorithm,
        token: {
          colorPrimary: '#0f766e',
          borderRadius: 8,
          fontFamily:
            'Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif'
        }
      }}
    >
      <Layout className="app-shell">
        <Sider className="app-sidebar" width={248} breakpoint="md" collapsedWidth={0}>
          <div className="brand">
            <img className="brand-mark" src="/ceph-tower-logo.svg" alt="CephTower logo" />
            <div>
              <Text strong>CephTower</Text>
              <Text type="secondary" className="brand-subtitle">
                Ceph 管理控制台
              </Text>
            </div>
          </div>
          <Menu mode="inline" selectedKeys={['overview']} items={navItems} />
        </Sider>

        <Content className="app-content">
          <header className="page-header">
            <div>
              <Title level={1}>集群概览</Title>
              <Paragraph>通过 Ceph Manager Dashboard API 汇总集群运行状态。</Paragraph>
            </div>
          </header>

          {loading ? (
            <Card className="state-card">
              <Spin tip="正在加载..." />
            </Card>
          ) : error ? (
            <Alert type="error" message="加载失败" description={error} showIcon />
          ) : (
            <div className="metrics-grid">
              <Card>
                <Statistic title="健康状态" value={summary?.health_status ?? 'unknown'} />
              </Card>
              <Card>
                <Statistic title="Ceph 版本" value={summary?.version || '未配置'} />
              </Card>
            </div>
          )}
        </Content>
      </Layout>
    </ConfigProvider>
  )
}

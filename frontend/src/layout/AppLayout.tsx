import {
  AppstoreOutlined,
  BarChartOutlined,
  BellOutlined,
  CloudServerOutlined,
  CodeOutlined,
  DatabaseOutlined,
  FileTextOutlined,
  HddOutlined,
  SearchOutlined,
  SettingOutlined
} from '@ant-design/icons'
import { Avatar, Badge, Button, Input, Layout, Menu, Space, Typography } from 'antd'
import type { MenuProps } from 'antd'
import type { PageKey } from '../pages'

const { Content, Header, Sider } = Layout
const { Text } = Typography

const navItems: MenuProps['items'] = [
  {
    key: 'monitor-group',
    type: 'group',
    label: '监控',
    children: [
      { key: 'overview', icon: <AppstoreOutlined />, label: '集群总览' },
      { key: 'cluster', icon: <CloudServerOutlined />, label: navLabel('主机 / OSD', '6') },
      { key: 'services', icon: <CodeOutlined />, label: navLabel('守护进程', '48') }
    ]
  },
  {
    key: 'storage-group',
    type: 'group',
    label: '存储',
    children: [
      { key: 'storage', icon: <DatabaseOutlined />, label: navLabel('存储管理', '8') },
      { key: 'configuration', icon: <SettingOutlined />, label: '配置中心' },
      { key: 'logs', icon: <FileTextOutlined />, label: '运行日志' }
    ]
  },
  {
    key: 'later-group',
    type: 'group',
    label: '更多（后续）',
    children: [
      { key: 'monitoring', icon: <BarChartOutlined />, label: '监控入口', disabled: true },
      { key: 'hardware', icon: <HddOutlined />, label: '硬件资产', disabled: true }
    ]
  }
]

const pageTitles: Record<PageKey, string> = {
  overview: '集群 / 总览',
  cluster: '集群 / 资源',
  services: '集群 / 服务',
  storage: '存储 / 管理',
  configuration: '系统 / 配置',
  logs: '系统 / 日志'
}

interface AppLayoutProps {
  activePage: PageKey
  onPageChange: (page: PageKey) => void
  children: React.ReactNode
}

export function AppLayout({ activePage, onPageChange, children }: AppLayoutProps) {
  return (
    <Layout className="app-shell">
      <Sider className="app-sidebar" width={252} breakpoint="lg" collapsedWidth={0}>
        <div className="brand">
          <img className="brand-mark" src="/ceph-tower-logo.svg" alt="CephTower logo" />
          <div>
            <Text strong>CephTower</Text>
            <Text type="secondary" className="brand-subtitle">
              集群运维控制台
            </Text>
          </div>
        </div>
        <div className="cluster-card">
          <Text type="secondary" className="cluster-card-label">
            当前集群
          </Text>
          <Text strong className="cluster-card-name">
            prod-ceph-east-01
          </Text>
          <div className="cluster-card-meta">
            <span className="status-pill warning">HEALTH_WARN</span>
            <Text type="secondary">v20.2.2</Text>
          </div>
        </div>
        <Menu
          className="sidebar-menu"
          mode="inline"
          selectedKeys={[activePage]}
          items={navItems}
          onClick={({ key }) => onPageChange(key as PageKey)}
        />
        <div className="sidebar-footer">
          <div>
            <Text>Mgr</Text>
            <Text strong>mgr.a · node-01</Text>
          </div>
          <div>
            <Text>MON</Text>
            <Text strong>已连接</Text>
          </div>
          <div>
            <Text>API</Text>
            <Text strong>/api/v1</Text>
          </div>
        </div>
      </Sider>
      <Layout className="main-shell">
        <Header className="topbar">
          <Text strong className="breadcrumb-text">
            {pageTitles[activePage]}
          </Text>
          <Space size={14} className="topbar-tools">
            <Input
              className="global-search"
              prefix={<SearchOutlined />}
              suffix={<span className="shortcut-key">/</span>}
              placeholder="搜索主机、池、OSD..."
            />
            <Button className="icon-button" icon={<AppstoreOutlined />} />
            <Badge dot offset={[-4, 4]}>
              <Button className="icon-button" icon={<BellOutlined />} />
            </Badge>
            <Button className="user-button">
              <Avatar size={32}>AD</Avatar>
              <span>admin</span>
            </Button>
          </Space>
        </Header>
        <Content className="app-content">{children}</Content>
      </Layout>
    </Layout>
  )
}

function navLabel(label: string, count: string) {
  return (
    <span className="nav-label">
      <span>{label}</span>
      <span className="nav-count">{count}</span>
    </span>
  )
}

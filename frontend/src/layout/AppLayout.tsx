import {
  AppstoreOutlined,
  BarChartOutlined,
  BellOutlined,
  CloudServerOutlined,
  CodeOutlined,
  DatabaseOutlined,
  FileTextOutlined,
  HddOutlined,
  LogoutOutlined,
  SearchOutlined,
  SettingOutlined,
  TeamOutlined
} from '@ant-design/icons'
import { Avatar, Badge, Button, Input, Layout, Menu, Space, Typography } from 'antd'
import type { MenuProps } from 'antd'
import type { UserAccount } from '../api/auth'
import type { PageKey } from '../pages'

const { Content, Header, Sider } = Layout
const { Text } = Typography

const pageTitles: Record<PageKey, string> = {
  overview: '集群 / 总览',
  cluster: '集群 / 资源',
  services: '集群 / 服务',
  storage: '存储 / 管理',
  configuration: '系统 / 配置',
  logs: '系统 / 日志',
  users: '系统设置 / 用户管理'
}

interface AppLayoutProps {
  activePage: PageKey
  onPageChange: (page: PageKey) => void
  user: UserAccount
  onLogout: () => void
  children: React.ReactNode
}

export function AppLayout({ activePage, onPageChange, user, onLogout, children }: AppLayoutProps) {
  const initials = user.display_name?.slice(0, 2).toUpperCase() || user.username.slice(0, 2).toUpperCase()
  const navItems = buildNavItems(user)

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
              <Avatar size={32}>{initials}</Avatar>
              <span>{user.username}</span>
              <span className="role-chip">{user.role === 'admin' ? '管理员' : '普通用户'}</span>
            </Button>
            <Button className="icon-button" icon={<LogoutOutlined />} onClick={onLogout} title="退出登录" />
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

function buildNavItems(user: UserAccount): MenuProps['items'] {
  const isAdmin = user.role === 'admin'
  const canReadCluster = isAdmin || user.permissions.includes('cluster:read')
  const canReadStorage = isAdmin || user.permissions.includes('storage:read')
  const canReadSystem = isAdmin || user.permissions.includes('system:read')

  return [
    {
      key: 'monitor-group',
      type: 'group',
      label: '监控',
      children: [
        { key: 'overview', icon: <AppstoreOutlined />, label: '集群总览', disabled: !canReadCluster },
        { key: 'cluster', icon: <CloudServerOutlined />, label: navLabel('主机 / OSD', '6'), disabled: !canReadCluster },
        { key: 'services', icon: <CodeOutlined />, label: navLabel('守护进程', '48'), disabled: !canReadCluster }
      ]
    },
    {
      key: 'storage-group',
      type: 'group',
      label: '存储',
      children: [
        { key: 'storage', icon: <DatabaseOutlined />, label: navLabel('存储管理', '8'), disabled: !canReadStorage },
        { key: 'logs', icon: <FileTextOutlined />, label: '运行日志', disabled: !canReadSystem }
      ]
    },
    {
      key: 'system-group',
      type: 'group',
      label: '系统设置',
      children: [
        { key: 'configuration', icon: <SettingOutlined />, label: '配置中心', disabled: !canReadSystem },
        { key: 'users', icon: <TeamOutlined />, label: '用户管理', disabled: !isAdmin }
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
}

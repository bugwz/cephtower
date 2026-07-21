import {
  AppstoreOutlined,
  ApartmentOutlined,
  ApiOutlined,
  BarChartOutlined,
  BellOutlined,
  BranchesOutlined,
  CloudServerOutlined,
  CloudSyncOutlined,
  ClusterOutlined,
  ControlOutlined,
  DatabaseOutlined,
  DownOutlined,
  DoubleLeftOutlined,
  DoubleRightOutlined,
  FolderOutlined,
  FundOutlined,
  GatewayOutlined,
  GithubOutlined,
  GlobalOutlined,
  GoldOutlined,
  HddOutlined,
  HistoryOutlined,
  InboxOutlined,
  LineChartOutlined,
  LinkOutlined,
  LogoutOutlined,
  MailOutlined,
  PartitionOutlined,
  ProfileOutlined,
  ReconciliationOutlined,
  SearchOutlined,
  SettingOutlined,
  SafetyCertificateOutlined,
  SlidersOutlined,
  StopOutlined,
  TeamOutlined
} from '@ant-design/icons'
import { Badge, Button, Dropdown, Input, Layout, Menu, Space, Typography } from 'antd'
import type { MenuProps } from 'antd'
import { useEffect, useRef, useState } from 'react'
import type { FocusEvent, MouseEvent, ReactNode } from 'react'
import type { UserAccount } from '../api/auth'
import { findNavPage, findNavSection, NAV_SECTIONS, type NavIcon, type PageKey } from '../navigation'

const { Content, Header, Sider } = Layout
const { Text } = Typography

let suppressCollapsedFlyout = false
const SIDEBAR_COLLAPSED_STORAGE_KEY = 'cephtower.sidebarCollapsed'
const APP_VERSION = '0.1.0'
const GITHUB_REPOSITORY_URL = 'https://github.com/bugwz/cephtower'

interface AppLayoutProps {
  activePage: PageKey
  onPageChange: (page: PageKey) => void
  user: UserAccount
  onLogout: () => void
  children: React.ReactNode
}

export function AppLayout({ activePage, onPageChange, user, onLogout, children }: AppLayoutProps) {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(readStoredSidebarCollapsed)
  const [openFlyoutKey, setOpenFlyoutKey] = useState<string | null>(null)
  const collapsedMenuRef = useRef<HTMLElement | null>(null)
  const collapsedHoverArmed = useRef(!sidebarCollapsed)
  const hasPointerIntent = useRef(false)
  const displayName = user.display_name || user.username
  const roleLabel = user.role === 'admin' ? '管理员' : '普通用户'
  const permissionSummary = user.role === 'admin' ? '全部权限' : `${user.permissions.length} 项权限`
  const lastLoginLabel = formatDateTime(user.last_login_at)
  const navSections = buildNavSections(user)
  const navItems = buildNavItems(navSections)
  const defaultOpenKeys = getDefaultOpenKeys(navSections, activePage)
  const activeSection = findNavSection(activePage)
  const activeNavPage = findNavPage(activePage)
  const pageTitle = activeSection && activeNavPage ? `${activeSection.label} / ${activeNavPage.label}` : 'CephTower'
  const userDropdownItems: MenuProps['items'] = [
    {
      key: 'account',
      className: 'user-dropdown-account-item',
      label: (
        <div className="user-dropdown-card">
          <div className="user-dropdown-head">
            <img className="user-logo user-logo-dropdown" src="/admin-user-logo.svg" alt="" aria-hidden="true" />
            <div className="user-dropdown-title">
              <Text strong>{displayName}</Text>
              <Text type="secondary">@{user.username}</Text>
            </div>
            <span className="role-chip">{roleLabel}</span>
          </div>
          <div className="user-dropdown-meta">
            <span>
              <MailOutlined />
              {user.email || '未设置邮箱'}
            </span>
            <span>
              <SafetyCertificateOutlined />
              {permissionSummary}
            </span>
            <span>最近登录：{lastLoginLabel}</span>
          </div>
        </div>
      )
    },
    { type: 'divider' },
    {
      key: 'logout',
      danger: true,
      icon: <LogoutOutlined />,
      label: '退出登录'
    }
  ]

  useEffect(() => {
    function handlePointerMove(event: PointerEvent) {
      if (!didPointerActuallyMove(event.movementX, event.movementY)) {
        return
      }

      hasPointerIntent.current = true
      const collapsedMenu = collapsedMenuRef.current
      if (!collapsedMenu || !(event.target instanceof Node) || !collapsedMenu.contains(event.target)) {
        collapsedHoverArmed.current = true
      }
    }

    window.addEventListener('pointermove', handlePointerMove, { passive: true })
    return () => window.removeEventListener('pointermove', handlePointerMove)
  }, [])

  function handleCollapsedPageChange(event: MouseEvent<HTMLButtonElement>, page: PageKey) {
    event.currentTarget.blur()
    suppressCollapsedFlyout = true
    setOpenFlyoutKey(null)
    onPageChange(page)
  }

  function openCollapsedFlyout(section: NavSection) {
    if (collapsedHoverArmed.current && !suppressCollapsedFlyout && !section.children.every((item) => item.disabled)) {
      setOpenFlyoutKey(section.key)
    }
  }

  function handleCollapsedItemMouseEnter(event: MouseEvent<HTMLDivElement>, section: NavSection) {
    if (!hasPointerIntent.current) {
      return
    }

    openCollapsedFlyout(section)
  }

  function handleCollapsedItemMouseMove(event: MouseEvent<HTMLDivElement>, section: NavSection) {
    if (didPointerActuallyMove(event.movementX, event.movementY)) {
      hasPointerIntent.current = true
      openCollapsedFlyout(section)
    }
  }

  function toggleCollapsedFlyout(section: NavSection) {
    if (section.children.every((item) => item.disabled)) {
      return
    }

    collapsedHoverArmed.current = true
    suppressCollapsedFlyout = false
    setOpenFlyoutKey((currentKey) => (currentKey === section.key ? null : section.key))
  }

  function handleCollapsedItemBlur(event: FocusEvent<HTMLDivElement>) {
    if (!event.currentTarget.contains(event.relatedTarget)) {
      closeCollapsedFlyout()
    }
  }

  function closeCollapsedFlyout() {
    collapsedHoverArmed.current = true
    suppressCollapsedFlyout = false
    setOpenFlyoutKey(null)
  }

  function handleSidebarCollapsedToggle() {
    closeCollapsedFlyout()
    collapsedHoverArmed.current = true
    setSidebarCollapsed((collapsed) => {
      const nextCollapsed = !collapsed
      storeSidebarCollapsed(nextCollapsed)
      return nextCollapsed
    })
  }

  return (
    <Layout className="app-shell">
      <Sider
        className={`app-sidebar${sidebarCollapsed ? ' app-sidebar-collapsed' : ''}`}
        width={224}
        collapsedWidth={78}
        collapsed={sidebarCollapsed}
        trigger={null}
      >
        <Button
          className="sidebar-collapse-button"
          icon={sidebarCollapsed ? <DoubleRightOutlined /> : <DoubleLeftOutlined />}
          onClick={handleSidebarCollapsedToggle}
          title={sidebarCollapsed ? '展开导航栏' : '折叠导航栏'}
        />
        <div className="sidebar-reveal">
          <div className="brand">
            <img className="brand-mark" src="/ceph-tower-logo.svg" alt="CephTower logo" />
            <div className="brand-copy">
              <Text strong>CephTower</Text>
              <Text type="secondary" className="brand-subtitle">
                集群运维控制台
              </Text>
            </div>
          </div>
          <div className="sidebar-nav-stack">
            {!sidebarCollapsed ? (
            <Menu
                className="sidebar-menu"
                mode="inline"
                defaultOpenKeys={defaultOpenKeys}
                selectedKeys={[activePage]}
                items={navItems}
                onClick={({ key }) => onPageChange(key as PageKey)}
              />
            ) : null}
            <nav ref={collapsedMenuRef} className="collapsed-sidebar-menu" aria-label="折叠菜单" aria-hidden={!sidebarCollapsed}>
              {navSections.map((section) => (
                <div
                  key={section.key}
                  className={`collapsed-nav-item${section.children.some((item) => item.key === activePage) ? ' collapsed-nav-item-active' : ''}${openFlyoutKey === section.key ? ' collapsed-nav-item-open' : ''}`}
                  onMouseEnter={(event) => handleCollapsedItemMouseEnter(event, section)}
                  onMouseMove={(event) => handleCollapsedItemMouseMove(event, section)}
                  onMouseLeave={closeCollapsedFlyout}
                  onBlur={handleCollapsedItemBlur}
                >
                  <button
                    type="button"
                    className="collapsed-nav-button"
                    disabled={section.children.every((item) => item.disabled)}
                    tabIndex={sidebarCollapsed ? 0 : -1}
                    title={section.label}
                    aria-haspopup="menu"
                    aria-expanded={openFlyoutKey === section.key}
                    onClick={() => toggleCollapsedFlyout(section)}
                  >
                    {section.icon}
                  </button>
                  <div className="collapsed-nav-flyout" role="menu" aria-label={section.label}>
                    <div className="collapsed-nav-flyout-title">{section.label}</div>
                    <div className="collapsed-nav-flyout-list">
                      {section.children.map((item) => (
                        <button
                          key={item.key}
                          type="button"
                          className={`collapsed-nav-flyout-option${activePage === item.key ? ' collapsed-nav-flyout-option-active' : ''}`}
                          disabled={item.disabled}
                          role="menuitem"
                          tabIndex={sidebarCollapsed ? 0 : -1}
                          onClick={(event) => handleCollapsedPageChange(event, item.key as PageKey)}
                        >
                          {item.icon}
                          <span>{item.label}</span>
                        </button>
                      ))}
                    </div>
                  </div>
                </div>
              ))}
            </nav>
          </div>
          <div className="sidebar-footer">
            <div className="sidebar-version">
              <Text strong>v{APP_VERSION}</Text>
            </div>
            <Button
              className="sidebar-github-button"
              href={GITHUB_REPOSITORY_URL}
              target="_blank"
              rel="noreferrer"
              icon={<GithubOutlined />}
              title="打开 GitHub 仓库"
              aria-label="打开 GitHub 仓库"
            />
          </div>
        </div>
      </Sider>
      <Layout className="main-shell">
        <Header className="topbar">
          <Text strong className="breadcrumb-text">
            {pageTitle}
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
            <Dropdown
              menu={{
                items: userDropdownItems,
                onClick: ({ key }) => {
                  if (key === 'logout') {
                    onLogout()
                  }
                }
              }}
              placement="bottomRight"
              trigger={['click']}
            >
              <Button className="user-button" aria-label="打开用户菜单">
                <img className="user-logo user-logo-topbar" src="/admin-user-logo.svg" alt="" aria-hidden="true" />
                <span className="user-button-copy">
                  <span className="user-button-name">{displayName}</span>
                  <span className="user-button-role">{roleLabel}</span>
                </span>
                <DownOutlined className="user-button-caret" />
              </Button>
            </Dropdown>
          </Space>
        </Header>
        <Content className="app-content">{children}</Content>
      </Layout>
    </Layout>
  )
}

type NavChild = {
  key: PageKey
  icon: ReactNode
  label: string
  disabled?: boolean
}

type NavSection = {
  key: string
  icon: ReactNode
  label: string
  children: NavChild[]
}

function buildNavItems(sections: NavSection[]): MenuProps['items'] {
  return sections.map((section) => ({
    key: section.key,
    icon: section.icon,
    label: section.label,
    children: section.children.map((item) => ({
      key: item.key,
      icon: item.icon,
      label: item.label,
      disabled: item.disabled
    }))
  })) satisfies MenuProps['items']
}

function getDefaultOpenKeys(sections: NavSection[], activePage: PageKey) {
  const activeSection = sections.find((section) => section.children.some((item) => item.key === activePage))
  return activeSection ? [activeSection.key] : []
}

function didPointerActuallyMove(movementX: number, movementY: number) {
  return Math.abs(movementX) + Math.abs(movementY) > 0
}

function readStoredSidebarCollapsed() {
  try {
    return localStorage.getItem(SIDEBAR_COLLAPSED_STORAGE_KEY) === 'true'
  } catch {
    return false
  }
}

function storeSidebarCollapsed(collapsed: boolean) {
  try {
    localStorage.setItem(SIDEBAR_COLLAPSED_STORAGE_KEY, String(collapsed))
  } catch {
    // Ignore storage errors so navigation remains usable in restricted contexts.
  }
}

function formatDateTime(value?: string) {
  if (!value) {
    return '暂无记录'
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return '暂无记录'
  }

  return date.toLocaleString()
}

function buildNavSections(user: UserAccount): NavSection[] {
  const isAdmin = user.role === 'admin'
  const canReadCluster = isAdmin || user.permissions.includes('cluster:read')
  const canReadStorage = isAdmin || user.permissions.includes('storage:read')
  const canReadSystem = isAdmin || user.permissions.includes('system:read')

  return NAV_SECTIONS.map((section) => ({
    key: section.key,
    icon: renderNavIcon(section.icon),
    label: section.label,
    children: section.children.map((item) => ({
      key: item.key,
      icon: renderNavIcon(item.icon),
      label: item.label,
      disabled:
        (item.permission === 'cluster' && !canReadCluster) ||
        (item.permission === 'storage' && !canReadStorage) ||
        (item.permission === 'system' && !canReadSystem)
    }))
  }))
}

function renderNavIcon(icon: NavIcon) {
  switch (icon) {
    case 'overview':
      return <AppstoreOutlined />
    case 'cluster':
      return <ClusterOutlined />
    case 'host':
      return <CloudServerOutlined />
    case 'mon':
      return <FundOutlined />
    case 'mgr':
      return <ControlOutlined />
    case 'osd':
      return <HddOutlined />
    case 'mds':
      return <ApartmentOutlined />
    case 'block':
      return <DatabaseOutlined />
    case 'pool':
      return <GoldOutlined />
    case 'rbd':
      return <InboxOutlined />
    case 'sync':
      return <CloudSyncOutlined />
    case 'iscsi':
      return <LinkOutlined />
    case 'nvme':
      return <ApiOutlined />
    case 'file':
      return <FolderOutlined />
    case 'cephfs':
      return <PartitionOutlined />
    case 'nfs':
      return <GlobalOutlined />
    case 'smb':
      return <ProfileOutlined />
    case 'object':
      return <CloudServerOutlined />
    case 'bucket':
      return <InboxOutlined />
    case 'gateway':
      return <GatewayOutlined />
    case 'site':
      return <BranchesOutlined />
    case 'user':
      return <TeamOutlined />
    case 'monitor':
      return <BarChartOutlined />
    case 'metrics':
      return <LineChartOutlined />
    case 'logs':
      return <HistoryOutlined />
    case 'alert':
      return <BellOutlined />
    case 'rule':
      return <ReconciliationOutlined />
    case 'silence':
      return <StopOutlined />
    case 'system':
      return <SettingOutlined />
    case 'config':
      return <SlidersOutlined />
    case 'data':
      return <DatabaseOutlined />
  }
}

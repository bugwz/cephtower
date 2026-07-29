import { ApiOutlined, ClusterOutlined, SafetyCertificateOutlined, TeamOutlined } from '@ant-design/icons'
import { Alert, Card, Descriptions, Progress, Space, Statistic, Table, Tag, Typography } from 'antd'
import { useCallback } from 'react'
import { listClusterCapabilities, type ClusterCapability } from '../../api/cluster'
import { listCredentials, listEndpoints, type CredentialView, type EndpointView } from '../../api/endpoint'
import { listRoleBindings, listRoles, type RoleBindingView, type RoleView } from '../../api/rbac'
import { listUsers, setupStatus, type SetupStatus, type UserAccount } from '../../api/auth'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'
import { useClusterContext } from '../../state/ClusterContext'

const { Text } = Typography

interface SystemInfoData {
  setup: SetupStatus
  users: UserAccount[]
  roles: RoleView[]
  capabilities: ClusterCapability[]
  endpoints: EndpointView[]
  credentials: CredentialView[]
  bindings: RoleBindingView[]
}

export function SystemInfoPage() {
  const { selectedCluster, selectedClusterId } = useClusterContext()
  const loader = useCallback(async (): Promise<SystemInfoData> => {
    const [setup, users, roles] = await Promise.all([
      setupStatus(),
      listUsers(),
      listRoles()
    ])
    if (!selectedClusterId) {
      return { setup, users, roles, capabilities: [], endpoints: [], credentials: [], bindings: [] }
    }
    const [capabilities, endpoints, credentials, bindings] = await Promise.all([
      listClusterCapabilities(selectedClusterId),
      listEndpoints(selectedClusterId),
      listCredentials(selectedClusterId),
      listRoleBindings(selectedClusterId)
    ])
    return { setup, users, roles, capabilities, endpoints, credentials, bindings }
  }, [selectedClusterId])
  const { data, loading, error } = useResource(loader)

  const supportedCapabilities = data?.capabilities.filter((item) => item.supported).length ?? 0
  const capabilityPercent = data?.capabilities.length
    ? Math.round((supportedCapabilities / data.capabilities.length) * 100)
    : 0

  return (
    <Page title="系统信息" loading={loading} error={error}>
      <Space direction="vertical" size={16} className="page-stack">
        {!selectedClusterId ? (
          <Alert type="info" showIcon message="请选择集群" description="系统全局信息可以浏览，集群能力、Endpoint、授权和操作状态需要先选择一个集群。" />
        ) : null}

        <div className="metrics-grid">
          <Card>
            <Statistic prefix={<ClusterOutlined />} title="当前集群" value={selectedCluster?.name ?? '未选择'} />
            <Text type="secondary">{selectedCluster?.monitor_addresses || '暂无 MON 地址'}</Text>
          </Card>
          <Card>
            <Statistic prefix={<SafetyCertificateOutlined />} title="能力支持率" value={capabilityPercent} suffix="%" />
            <Progress percent={capabilityPercent} size="small" showInfo={false} />
          </Card>
          <Card>
            <Statistic prefix={<ApiOutlined />} title="Endpoint" value={data?.endpoints.length ?? 0} />
            <Text type="secondary">credential {data?.credentials.length ?? 0}</Text>
          </Card>
          <Card>
            <Statistic prefix={<TeamOutlined />} title="用户 / 角色" value={`${data?.users.length ?? 0} / ${data?.roles.length ?? 0}`} />
            <Text type="secondary">cluster bindings {data?.bindings.length ?? 0}</Text>
          </Card>
        </div>

        <Card className="page-surface-card" title="系统状态">
          <Descriptions size="small" column={{ xs: 1, sm: 2, lg: 3 }}>
            <Descriptions.Item label="初始化状态">{data?.setup.initialized ? '已初始化' : '需要初始化'}</Descriptions.Item>
            <Descriptions.Item label="数据库引擎">{data?.setup.database?.engine ?? '-'}</Descriptions.Item>
            <Descriptions.Item label="SQLite 文件">{data?.setup.database?.sqlite.name ?? '-'}</Descriptions.Item>
            <Descriptions.Item label="集群 ID">{selectedClusterId ?? '-'}</Descriptions.Item>
            <Descriptions.Item label="FSID">{selectedCluster?.fsid || '-'}</Descriptions.Item>
            <Descriptions.Item label="Client">{selectedCluster?.client_username || '-'}</Descriptions.Item>
          </Descriptions>
        </Card>

        <div className="content-grid two-columns system-info-grid">
          <Card title="集群能力">
            <Table<ClusterCapability>
              size="small"
              rowKey="name"
              dataSource={data?.capabilities ?? []}
              pagination={{ pageSize: 6, showSizeChanger: false }}
              columns={[
                { title: '能力', dataIndex: 'name' },
                { title: '状态', dataIndex: 'supported', width: 90, render: (supported) => <Tag color={supported ? 'success' : 'warning'}>{supported ? '支持' : '不可用'}</Tag> },
                { title: '原因', dataIndex: 'reason', ellipsis: true, render: (value) => value || '-' }
              ]}
            />
          </Card>
          <Card title="Endpoint">
            <Table<EndpointView>
              size="small"
              rowKey="endpoint_id"
              dataSource={data?.endpoints ?? []}
              pagination={{ pageSize: 6, showSizeChanger: false }}
              columns={[
                { title: 'Kind', dataIndex: 'kind' },
                { title: 'URL', dataIndex: 'url', ellipsis: true },
                { title: '状态', dataIndex: 'enabled', width: 90, render: (enabled) => <Tag color={enabled ? 'success' : 'default'}>{enabled ? '启用' : '停用'}</Tag> }
              ]}
            />
          </Card>
        </div>

        <div className="content-grid system-info-grid">
          <Card title="集群授权">
            <Table<RoleBindingView>
              size="small"
              rowKey="role_binding_id"
              dataSource={data?.bindings ?? []}
              pagination={{ pageSize: 6, showSizeChanger: false }}
              columns={[
                { title: '用户', dataIndex: 'username' },
                { title: '角色', dataIndex: 'role' },
                { title: '创建时间', dataIndex: 'created_at', render: formatTime }
              ]}
            />
          </Card>
        </div>
      </Space>
    </Page>
  )
}

function formatTime(value?: string | null) {
  if (!value) {
    return '-'
  }
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

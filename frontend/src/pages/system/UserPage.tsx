import { PlusOutlined, ReloadOutlined, SaveOutlined, TeamOutlined, UserAddOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Popconfirm, Select, Space, Table, Tag, Typography } from 'antd'
import { useCallback, useEffect, useState } from 'react'
import { createUser, listUsers, type UserAccount, type UserRole } from '../../api/auth'
import {
  createRole,
  createRoleBinding,
  deleteRoleBinding,
  listRoleBindings,
  listRoles,
  type RoleBindingView,
  type RoleView
} from '../../api/rbac'
import { DraggableModal } from '../../components/DraggableModal'
import { Page } from '../../components/Page'
import { TableAction } from '../../components/TableActions'
import { useClusterContext } from '../../state/ClusterContext'
import { message } from '../../utils/appMessage'

const { Text } = Typography

export type UserPageView = 'users' | 'roles' | 'bindings'

const roleOptions: Array<{ label: string; value: UserRole }> = [
  { label: 'Cluster Admin', value: 'cluster-admin' },
  { label: 'Security Admin', value: 'security-admin' },
  { label: 'Storage Admin', value: 'storage-admin' },
  { label: 'Operator', value: 'operator' },
  { label: 'Viewer', value: 'viewer' }
]

export function UserPage({ view = 'users' }: { view?: UserPageView }) {
  const { selectedClusterId } = useClusterContext()
  const [users, setUsers] = useState<UserAccount[]>([])
  const [roles, setRoles] = useState<RoleView[]>([])
  const [bindings, setBindings] = useState<RoleBindingView[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [createUserOpen, setCreateUserOpen] = useState(false)
  const [createRoleOpen, setCreateRoleOpen] = useState(false)
  const [bindingOpen, setBindingOpen] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [userForm] = Form.useForm()
  const [roleForm] = Form.useForm()
  const [bindingForm] = Form.useForm()

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [nextUsers, nextRoles, nextBindings] = await Promise.all([
        listUsers(),
        listRoles(),
        selectedClusterId ? listRoleBindings(selectedClusterId) : Promise.resolve([])
      ])
      setUsers(nextUsers)
      setRoles(nextRoles)
      setBindings(nextBindings)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载用户与授权失败')
    } finally {
      setLoading(false)
    }
  }, [selectedClusterId])

  useEffect(() => {
    load()
  }, [load])

  async function submitUser(values: {
    username: string
    display_name: string
    email?: string
    role: UserRole
    password: string
  }) {
    if (submitting) {
      return
    }
    setSubmitting(true)
    try {
      const user = await createUser({ ...values, enabled: true })
      setUsers((current) => [...current, user])
      setCreateUserOpen(false)
      userForm.resetFields()
      message.success('用户已创建')
    } finally {
      setSubmitting(false)
    }
  }

  async function submitRole(values: { name: string; description?: string }) {
    if (submitting) {
      return
    }
    setSubmitting(true)
    try {
      const role = await createRole(values)
      setRoles((current) => [...current, role])
      setCreateRoleOpen(false)
      roleForm.resetFields()
      message.success('角色已创建')
    } finally {
      setSubmitting(false)
    }
  }

  async function submitBinding(values: { user_id: number; role: string }) {
    if (!selectedClusterId || submitting) {
      return
    }
    setSubmitting(true)
    try {
      const binding = await createRoleBinding({
        clusterId: selectedClusterId,
        userId: values.user_id,
        role: values.role
      })
      setBindings((current) => [...current, binding])
      setBindingOpen(false)
      bindingForm.resetFields()
      message.success('集群授权已创建')
    } finally {
      setSubmitting(false)
    }
  }

  async function removeBinding(row: RoleBindingView) {
    if (!selectedClusterId) {
      return
    }
    await deleteRoleBinding(selectedClusterId, row.role_binding_id)
    setBindings((current) => current.filter((item) => item.role_binding_id !== row.role_binding_id))
    message.success('集群授权已删除')
  }

  function renderUserManagementView() {
    if (view === 'roles') {
      return (
        <Space direction="vertical" size={16} className="page-stack">
          <Space>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateRoleOpen(true)}>新建角色</Button>
          </Space>
          <Table<RoleView>
            size="middle"
            rowKey="id"
            dataSource={roles}
            pagination={{ pageSize: 8, showSizeChanger: false }}
            columns={[
              { title: '角色', dataIndex: 'name' },
              { title: '描述', dataIndex: 'description', render: (value) => value || '-' },
              { title: '创建时间', dataIndex: 'created_at', render: formatTime }
            ]}
          />
        </Space>
      )
    }

    if (view === 'bindings') {
      return (
        <Space direction="vertical" size={16} className="page-stack">
          <Space>
            <Button type="primary" icon={<TeamOutlined />} disabled={!selectedClusterId} onClick={() => setBindingOpen(true)}>新建授权</Button>
          </Space>
          <Table<RoleBindingView>
            size="middle"
            rowKey="role_binding_id"
            dataSource={bindings}
            pagination={{ pageSize: 8, showSizeChanger: false }}
            columns={[
              { title: '用户', dataIndex: 'username' },
              { title: 'User ID', dataIndex: 'user_id' },
              { title: '角色', dataIndex: 'role', render: (role) => <Tag color="geekblue">{role}</Tag> },
              { title: '创建时间', dataIndex: 'created_at', render: formatTime },
              {
                title: '操作',
                width: 70,
                render: (_, row) => (
                  <Popconfirm title="删除集群授权" okText="删除" cancelText="取消" onConfirm={() => removeBinding(row)}>
                    <TableAction danger>删除</TableAction>
                  </Popconfirm>
                )
              }
            ]}
          />
        </Space>
      )
    }

    return (
      <Space direction="vertical" size={16} className="page-stack">
        <Space>
          <Button type="primary" icon={<UserAddOutlined />} onClick={() => setCreateUserOpen(true)}>新建用户</Button>
        </Space>
        <Table<UserAccount>
          size="middle"
          rowKey="id"
          dataSource={users}
          pagination={{ pageSize: 8, showSizeChanger: false }}
          scroll={{ x: 980 }}
          columns={[
            {
              title: '用户',
              key: 'user',
              render: (_, user) => (
                <Space direction="vertical" size={0}>
                  <Text strong>{user.display_name || user.username}</Text>
                  <Text type="secondary">{user.username}</Text>
                </Space>
              )
            },
            { title: '邮箱', dataIndex: 'email', render: (value) => value || '-' },
            { title: '默认角色', dataIndex: 'role', render: (role) => <Tag color="blue">{role}</Tag> },
            { title: '状态', dataIndex: 'enabled', render: (enabled) => <Tag color={enabled ? 'success' : 'default'}>{enabled ? '启用' : '停用'}</Tag> },
            { title: '最近登录', dataIndex: 'last_login_at', render: formatTime },
            { title: '创建时间', dataIndex: 'created_at', render: formatTime }
          ]}
        />
      </Space>
    )
  }

  return (
    <Page title="用户管理" loading={loading} error={error}>
      <Card
        className="page-surface-card"
        title="用户、角色与集群授权"
        extra={<Button icon={<ReloadOutlined />} loading={loading} onClick={load}>刷新</Button>}
      >
        {renderUserManagementView()}
      </Card>

      <DraggableModal
        title="新建用户"
        open={createUserOpen}
        onCancel={() => setCreateUserOpen(false)}
        onOk={() => userForm.submit()}
        okText="创建"
        confirmLoading={submitting}
        okButtonProps={{ icon: <SaveOutlined /> }}
        destroyOnClose
      >
        <Form form={userForm} layout="vertical" initialValues={{ role: 'viewer' }} onFinish={submitUser}>
          <Form.Item name="username" label="用户名" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="display_name" label="显示名" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ type: 'email', message: '请输入有效邮箱地址' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="role" label="默认角色" rules={[{ required: true }]}>
            <Select options={roleOptions} />
          </Form.Item>
          <Form.Item name="password" label="初始密码" rules={[{ required: true, min: 8 }]}>
            <Input.Password />
          </Form.Item>
        </Form>
      </DraggableModal>

      <DraggableModal
        title="新建角色"
        open={createRoleOpen}
        onCancel={() => setCreateRoleOpen(false)}
        onOk={() => roleForm.submit()}
        okText="创建"
        confirmLoading={submitting}
        okButtonProps={{ icon: <SaveOutlined /> }}
        destroyOnClose
      >
        <Form form={roleForm} layout="vertical" onFinish={submitRole}>
          <Form.Item name="name" label="角色名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={4} />
          </Form.Item>
        </Form>
      </DraggableModal>

      <DraggableModal
        title="新建集群授权"
        open={bindingOpen}
        onCancel={() => setBindingOpen(false)}
        onOk={() => bindingForm.submit()}
        okText="创建"
        confirmLoading={submitting}
        okButtonProps={{ icon: <SaveOutlined /> }}
        destroyOnClose
      >
        <Form form={bindingForm} layout="vertical" onFinish={submitBinding}>
          <Form.Item name="user_id" label="用户" rules={[{ required: true }]}>
            <Select options={users.map((user) => ({ label: `${user.display_name || user.username} (${user.username})`, value: user.id }))} />
          </Form.Item>
          <Form.Item name="role" label="角色" rules={[{ required: true }]}>
            <Select options={roles.map((role) => ({ label: role.name, value: role.name }))} />
          </Form.Item>
        </Form>
      </DraggableModal>
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

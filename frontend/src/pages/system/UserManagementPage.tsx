import {
  KeyOutlined,
  PlusOutlined,
  ReloadOutlined,
  SaveOutlined,
  UserAddOutlined
} from '@ant-design/icons'
import { Button, Card, Form, Input, Select, Space, Switch, Table, Tag, Typography, message } from 'antd'
import { useCallback, useEffect, useState } from 'react'
import { createUser, listUsers, updateUser, type UserAccount, type UserRole } from '../../api/auth'
import { DraggableModal } from '../../components/DraggableModal'
import { Page } from '../../components/Page'

const { Text } = Typography

const permissionOptions = [
  { label: '集群读取', value: 'cluster:read' },
  { label: '存储读取', value: 'storage:read' },
  { label: '系统读取', value: 'system:read' },
  { label: '用户管理', value: 'user:manage' }
]

export function UserManagementPage() {
  const [users, setUsers] = useState<UserAccount[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [passwordTarget, setPasswordTarget] = useState<UserAccount | null>(null)
  const [createSubmitting, setCreateSubmitting] = useState(false)
  const [passwordSubmitting, setPasswordSubmitting] = useState(false)
  const [pendingUserUpdates, setPendingUserUpdates] = useState<Record<string, boolean>>({})
  const [createForm] = Form.useForm()
  const [passwordForm] = Form.useForm()

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setUsers(await listUsers())
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载用户失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  async function patchUser(id: number, payload: Parameters<typeof updateUser>[1], pendingKey = `${id}:update`) {
    if (pendingUserUpdates[pendingKey]) {
      return
    }
    setPendingUserUpdates((current) => ({ ...current, [pendingKey]: true }))
    try {
      const updated = await updateUser(id, payload)
      setUsers((current) => current.map((user) => (user.id === updated.id ? updated : user)))
      message.success('用户已更新')
    } finally {
      setPendingUserUpdates((current) => {
        const next = { ...current }
        delete next[pendingKey]
        return next
      })
    }
  }

  async function submitCreate(values: {
    username: string
    display_name: string
    email?: string
    role: UserRole
    permissions: string[]
    password: string
    enabled: boolean
  }) {
    if (createSubmitting) {
      return
    }
    setCreateSubmitting(true)
    try {
      const user = await createUser({ ...values, enabled: values.enabled ?? true })
      setUsers((current) => [...current, user])
      setCreateOpen(false)
      createForm.resetFields()
      message.success('用户已创建')
    } finally {
      setCreateSubmitting(false)
    }
  }

  async function resetPassword(values: { password: string }) {
    if (!passwordTarget) {
      return
    }
    if (passwordSubmitting) {
      return
    }
    setPasswordSubmitting(true)
    try {
      await patchUser(passwordTarget.id, { password: values.password }, `${passwordTarget.id}:password`)
      setPasswordTarget(null)
      passwordForm.resetFields()
    } finally {
      setPasswordSubmitting(false)
    }
  }

  return (
    <Page
      title="用户管理"
      loading={loading}
      error={error}
    >
      <Card
        className="page-surface-card"
        title="系统用户"
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} loading={loading} onClick={load}>
              刷新
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>
              新建用户
            </Button>
          </Space>
        }
      >
        <Table
          size="middle"
          rowKey="id"
          dataSource={users}
          pagination={{ pageSize: 8, showSizeChanger: false }}
          scroll={{ x: true }}
          columns={[
            {
              title: '用户',
              key: 'user',
              render: (_, user) => (
                <div className="user-cell">
                  <Text strong>{user.display_name}</Text>
                  <Text type="secondary">{user.username}</Text>
                </div>
              )
            },
            {
              title: '邮箱',
              dataIndex: 'email',
              render: (email: string, user) => (
                <Input
                  className="email-input"
                  value={email}
                  placeholder="未绑定邮箱"
                  disabled={Boolean(pendingUserUpdates[`${user.id}:email`])}
                  onChange={(event) =>
                    setUsers((current) =>
                      current.map((item) => (item.id === user.id ? { ...item, email: event.target.value } : item))
                    )
                  }
                  onBlur={(event) => patchUser(user.id, { email: event.target.value }, `${user.id}:email`)}
                />
              )
            },
            {
              title: '角色',
              dataIndex: 'role',
              render: (role: UserRole, user) => (
                <Select
                  className="role-select"
                  value={role}
                  options={[
                    { label: '管理员', value: 'admin' },
                    { label: '普通用户', value: 'user' }
                  ]}
                  loading={Boolean(pendingUserUpdates[`${user.id}:role`])}
                  disabled={Boolean(pendingUserUpdates[`${user.id}:role`])}
                  onChange={(value) => patchUser(user.id, { role: value }, `${user.id}:role`)}
                />
              )
            },
            {
              title: '访问权限',
              dataIndex: 'permissions',
              render: (permissions: string[], user) => (
                <Select
                  mode="multiple"
                  className="permission-select"
                  value={permissions}
                  options={permissionOptions}
                  maxTagCount="responsive"
                  loading={Boolean(pendingUserUpdates[`${user.id}:permissions`])}
                  disabled={Boolean(pendingUserUpdates[`${user.id}:permissions`])}
                  onChange={(value) => patchUser(user.id, { permissions: value }, `${user.id}:permissions`)}
                />
              )
            },
            {
              title: '状态',
              dataIndex: 'enabled',
              render: (enabled: boolean, user) => (
                <Space>
                  <Switch
                    checked={enabled}
                    loading={Boolean(pendingUserUpdates[`${user.id}:enabled`])}
                    onChange={(value) => patchUser(user.id, { enabled: value }, `${user.id}:enabled`)}
                  />
                  <Tag color={enabled ? 'success' : 'default'}>{enabled ? '启用' : '禁用'}</Tag>
                </Space>
              )
            },
            {
              title: '最近登录',
              dataIndex: 'last_login_at',
              render: (value?: string) => value ? new Date(value).toLocaleString() : '—'
            },
            {
              title: '操作',
              key: 'actions',
              render: (_, user) => (
                <Button
                  icon={<KeyOutlined />}
                  loading={Boolean(pendingUserUpdates[`${user.id}:password`])}
                  onClick={() => setPasswordTarget(user)}
                >
                  重设密码
                </Button>
              )
            }
          ]}
        />
      </Card>

      <DraggableModal
        title="新建用户"
        open={createOpen}
        onCancel={() => {
          if (!createSubmitting) {
            setCreateOpen(false)
          }
        }}
        onOk={() => createForm.submit()}
        okText="创建"
        confirmLoading={createSubmitting}
        okButtonProps={{ icon: <UserAddOutlined />, loading: createSubmitting }}
        cancelButtonProps={{ disabled: createSubmitting }}
      >
        <Form
          form={createForm}
          layout="vertical"
          initialValues={{
            role: 'user',
            permissions: ['cluster:read', 'storage:read'],
            enabled: true
          }}
          onFinish={submitCreate}
        >
          <Form.Item name="username" label="用户名" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="display_name" label="显示名" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ type: 'email', message: '请输入有效邮箱地址' }]}>
            <Input placeholder="用于忘记密码验证码" />
          </Form.Item>
          <Form.Item name="role" label="角色" rules={[{ required: true }]}>
            <Select
              options={[
                { label: '管理员', value: 'admin' },
                { label: '普通用户', value: 'user' }
              ]}
            />
          </Form.Item>
          <Form.Item name="permissions" label="访问权限" rules={[{ required: true }]}>
            <Select mode="multiple" options={permissionOptions} />
          </Form.Item>
          <Form.Item name="password" label="初始密码" rules={[{ required: true, min: 8 }]}>
            <Input.Password />
          </Form.Item>
          <Form.Item name="enabled" label="启用账号" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </DraggableModal>

      <DraggableModal
        title={`重设密码${passwordTarget ? `：${passwordTarget.username}` : ''}`}
        open={Boolean(passwordTarget)}
        onCancel={() => {
          if (!passwordSubmitting) {
            setPasswordTarget(null)
          }
        }}
        onOk={() => passwordForm.submit()}
        okText="保存"
        confirmLoading={passwordSubmitting}
        okButtonProps={{ icon: <SaveOutlined />, loading: passwordSubmitting }}
        cancelButtonProps={{ disabled: passwordSubmitting }}
      >
        <Form form={passwordForm} layout="vertical" onFinish={resetPassword}>
          <Form.Item name="password" label="新密码" rules={[{ required: true, min: 8 }]}>
            <Input.Password />
          </Form.Item>
        </Form>
      </DraggableModal>
    </Page>
  )
}

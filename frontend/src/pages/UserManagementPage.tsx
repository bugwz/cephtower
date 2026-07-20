import {
  KeyOutlined,
  PlusOutlined,
  ReloadOutlined,
  SaveOutlined,
  UserAddOutlined
} from '@ant-design/icons'
import { Button, Card, Form, Input, Modal, Select, Space, Switch, Table, Tag, Typography, message } from 'antd'
import { useCallback, useEffect, useState } from 'react'
import { createUser, listUsers, updateUser, type UserAccount, type UserRole } from '../api/auth'
import { Page } from '../components/Page'

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

  async function patchUser(id: number, payload: Parameters<typeof updateUser>[1]) {
    const updated = await updateUser(id, payload)
    setUsers((current) => current.map((user) => (user.id === updated.id ? updated : user)))
    message.success('用户已更新')
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
    const user = await createUser({ ...values, enabled: values.enabled ?? true })
    setUsers((current) => [...current, user])
    setCreateOpen(false)
    createForm.resetFields()
    message.success('用户已创建')
  }

  async function resetPassword(values: { password: string }) {
    if (!passwordTarget) {
      return
    }
    await patchUser(passwordTarget.id, { password: values.password })
    setPasswordTarget(null)
    passwordForm.resetFields()
  }

  return (
    <Page
      title="用户管理"
      description="管理 CephTower 本地用户、角色、访问权限和账号启停。"
      loading={loading}
      error={error}
      onRefresh={load}
    >
      <Card
        title="系统用户"
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={load}>
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
                  onChange={(event) =>
                    setUsers((current) =>
                      current.map((item) => (item.id === user.id ? { ...item, email: event.target.value } : item))
                    )
                  }
                  onBlur={(event) => patchUser(user.id, { email: event.target.value })}
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
                  onChange={(value) => patchUser(user.id, { role: value })}
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
                  onChange={(value) => patchUser(user.id, { permissions: value })}
                />
              )
            },
            {
              title: '状态',
              dataIndex: 'enabled',
              render: (enabled: boolean, user) => (
                <Space>
                  <Switch checked={enabled} onChange={(value) => patchUser(user.id, { enabled: value })} />
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
                <Button icon={<KeyOutlined />} onClick={() => setPasswordTarget(user)}>
                  重设密码
                </Button>
              )
            }
          ]}
        />
      </Card>

      <Modal
        title="新建用户"
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={() => createForm.submit()}
        okText="创建"
        okButtonProps={{ icon: <UserAddOutlined /> }}
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
      </Modal>

      <Modal
        title={`重设密码${passwordTarget ? `：${passwordTarget.username}` : ''}`}
        open={Boolean(passwordTarget)}
        onCancel={() => setPasswordTarget(null)}
        onOk={() => passwordForm.submit()}
        okText="保存"
        okButtonProps={{ icon: <SaveOutlined /> }}
      >
        <Form form={passwordForm} layout="vertical" onFinish={resetPassword}>
          <Form.Item name="password" label="新密码" rules={[{ required: true, min: 8 }]}>
            <Input.Password />
          </Form.Item>
        </Form>
      </Modal>
    </Page>
  )
}

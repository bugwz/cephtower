import { Alert, Button, Input, Space } from 'antd'
import type { ApiRecord } from '../../api/client'

export function smbUsersInitialValues(row?: ApiRecord) {
  return {
    users: Array.isArray(row?.user_names) && row.user_names.every((name) => typeof name === 'string') ? JSON.stringify(row.user_names.map((name) => ({ name, password: '' }))) : undefined,
    groups: Array.isArray(row?.group_names) && row.group_names.every((name) => typeof name === 'string') ? row.group_names.join('\n') : undefined,
    linked_to_cluster: typeof row?.linked_to_cluster === 'string' ? row.linked_to_cluster : undefined
  }
}

export function SMBUsersEditor({ value, onChange, id }: { value?: string; onChange?: (value: string) => void; id?: string }) {
  let users: Array<{ name: string; password: string }>
  try {
    users = value ? JSON.parse(value) : []
    if (!Array.isArray(users) || users.some((user) => !user || typeof user.name !== 'string' || typeof user.password !== 'string')) throw new Error('invalid')
  } catch {
    return <Alert type="error" message="用户列表格式异常，请重新打开表单" />
  }
  const save = (next: typeof users) => onChange?.(JSON.stringify(next))
  const update = (index: number, field: 'name' | 'password', text: string) => save(users.map((user, i) => i === index ? { ...user, [field]: text } : user))
  return <Space id={id} direction="vertical" style={{ width: '100%' }}>
    {users.map((user, index) => <Space key={index}>
      <Input aria-label={`用户 ${index + 1} 名称`} placeholder="用户名" value={user.name} onChange={(event) => update(index, 'name', event.target.value)} />
      <Input.Password aria-label={`用户 ${index + 1} 密码`} placeholder="密码" autoComplete="new-password" value={user.password} onChange={(event) => update(index, 'password', event.target.value)} />
      <Button danger onClick={() => save(users.filter((_, i) => i !== index))}>删除用户 {index + 1}</Button>
    </Space>)}
    <Button onClick={() => save([...users, { name: '', password: '' }])}>新增用户</Button>
  </Space>
}

export function smbUsersBody(value: unknown) {
  if (typeof value !== 'string') throw new Error('请提供完整用户列表')
  const users: unknown = JSON.parse(value)
  if (!Array.isArray(users) || users.some((user) => !user || typeof user.name !== 'string' || !user.name.trim() || typeof user.password !== 'string' || !user.password)) throw new Error('请填写每个用户的名称和密码')
  if (new Set(users.map((user) => user.name)).size !== users.length) throw new Error('用户名不能重复')
  return users.map((user) => ({ name: user.name, password: user.password }))
}

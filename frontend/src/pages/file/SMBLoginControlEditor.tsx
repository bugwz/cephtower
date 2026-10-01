import { Alert, Button, Input, Select, Space } from 'antd'
import type { ApiRecord } from '../../api/client'

type Entry = { name: string; category: string; access: string }
export function SMBLoginControlEditor({ value, onChange }: { value?: string; onChange?: (value: string) => void }) {
  let entries: Entry[]
  try {
    entries = value ? JSON.parse(value) : []
    if (!Array.isArray(entries) || entries.some((entry) => !entry || typeof entry.name !== 'string')) throw new Error('invalid')
  } catch { return <Alert type="error" message="登录控制格式异常，请刷新后重试" /> }
  const save = (next: Entry[]) => onChange?.(JSON.stringify(next))
  const update = (index: number, patch: Partial<Entry>) => save(entries.map((entry, i) => i === index ? { ...entry, ...patch } : entry))
  return <Space direction="vertical">
    {entries.map((entry, index) => <Space key={index} wrap>
      <Input aria-label={`规则 ${index + 1} 名称`} value={entry.name} placeholder="用户或组名" onChange={(event) => update(index, { name: event.target.value })} />
      <Select aria-label={`规则 ${index + 1} 类型`} value={entry.category} options={[{ label: '用户', value: 'user' }, { label: '组', value: 'group' }]} onChange={(category) => update(index, { category })} />
      <Select aria-label={`规则 ${index + 1} 权限`} value={entry.access} options={[{ label: '拒绝', value: 'none' }, { label: '只读', value: 'read' }, { label: '读写', value: 'read-write' }, { label: '管理员', value: 'admin' }]} onChange={(access) => update(index, { access })} />
      <Button danger onClick={() => save(entries.filter((_, i) => i !== index))}>删除规则 {index + 1}</Button>
    </Space>)}
    <Button onClick={() => save([...entries, { name: '', category: 'user', access: 'read' }])}>新增登录规则</Button>
  </Space>
}

export function smbLoginControlBody(values: ApiRecord) {
  if (values.replace_login_control !== true) return {}
  const entries: unknown = JSON.parse(typeof values.login_control === 'string' ? values.login_control : '[]')
  if (!Array.isArray(entries) || entries.some((entry) => !entry || typeof entry.name !== 'string' || !entry.name || /[\s\x00]/.test(entry.name) || [...entry.name].length > 128 || !['user', 'group'].includes(entry.category) || !['none', 'read', 'read-write', 'admin'].includes(entry.access))) throw new Error('请填写有效的登录控制规则')
  if (new Set(entries.map((entry) => `${entry.category}\0${entry.name}`)).size !== entries.length) throw new Error('不能重复配置同一个用户或组')
  const restricted = values.restrict_access === true
  if (restricted && !entries.some((entry) => entry.access !== 'none')) throw new Error('限制访问时至少需要一个允许访问的用户或组')
  return { login_control: entries, restrict_access: restricted }
}

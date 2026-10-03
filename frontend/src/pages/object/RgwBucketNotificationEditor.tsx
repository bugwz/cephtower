import { Alert, Button, Input, Select, Space } from 'antd'
import { notificationEditableRule, notificationEvents, notificationRuleShape, type NotificationDraft, type NotificationRule } from './rgwBucketNotificationForm'

export function RgwBucketNotificationEditor({ value, onChange, disabled = false }: { value?: NotificationDraft; onChange?: (value: NotificationDraft) => void; disabled?: boolean }) {
  if (!value || !notificationRuleShape(value.rule) || !Array.isArray(value.existing)) return <Alert type="error" message="通知草稿不可用，请重新打开表单" />
  const change = (next: NotificationDraft) => { if (!disabled) onChange?.(next) }
  const update = (patch: Partial<NotificationRule>) => change({ ...value, rule: { ...value.rule, ...patch } })
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Select aria-label="通知写入模式" disabled={disabled} value={value.mode || undefined} placeholder="选择创建或编辑" style={{ width: '100%' }} options={[{ value: 'create', label: '创建新通知' }, { value: 'edit', label: '编辑既有通知' }]} onChange={mode => change({ ...value, mode, selected: '', rule: { id: '', topic: '', events: [], filters: [] } })} />
    {value.mode === 'edit' ? <Select aria-label="选择通知 ID" disabled={disabled} value={value.selected || undefined} placeholder="选择唯一通知 ID" style={{ width: '100%' }} options={value.existing.filter(rule => rule.id && value.existing.filter(other => other.id === rule.id).length === 1).map(rule => ({ value: rule.id, label: JSON.stringify(rule.id) }))} onChange={id => { const rule = value.existing.find(rule => rule.id === id); if (rule) change({ ...value, selected: id, rule: notificationEditableRule(rule) }) }} /> : null}
    <Input aria-label="通知 ID" placeholder="通知 ID（保留空格）" disabled={disabled || value.mode !== 'create'} value={value.rule.id} onChange={e => update({ id: e.target.value })} />
    <Input aria-label="目标 Topic ARN" placeholder="完整 Topic ARN（可从 Topic 列表复制）" disabled={disabled} value={value.rule.topic} onChange={e => update({ topic: e.target.value })} />
    <Select aria-label="通知事件" mode="multiple" disabled={disabled} value={value.rule.events} style={{ width: '100%' }} options={notificationEvents.map(event => ({ value: event, label: event }))} onChange={events => update({ events })} />
    <Alert type="info" message="空事件列表默认订阅创建和删除；此版本不支持 ObjectRestore。S3Key 空值不会形成过滤；元数据和标签空值会保留。重复名称拒绝提交，不会静默去重。" />
    {value.rule.filters.map((filter, index) => <Space key={index} wrap>
      <Select aria-label={`过滤 ${index + 1} 类型`} disabled={disabled} value={filter.kind} options={['S3Key', 'S3Metadata', 'S3Tags'].map(kind => ({ value: kind, label: kind }))} onChange={kind => update({ filters: value.rule.filters.map((f, i) => i === index ? { ...f, kind, name: kind === 'S3Key' ? 'prefix' : '' } : f) })} />
      {filter.kind === 'S3Key' ? <Select aria-label={`过滤 ${index + 1} 名称`} disabled={disabled} value={filter.name} options={['prefix', 'suffix', 'regex'].map(name => ({ value: name, label: name }))} onChange={name => update({ filters: value.rule.filters.map((f, i) => i === index ? { ...f, name } : f) })} /> : <Input aria-label={`过滤 ${index + 1} 名称`} placeholder="名称" disabled={disabled} value={filter.name} onChange={e => update({ filters: value.rule.filters.map((f, i) => i === index ? { ...f, name: e.target.value } : f) })} />}
      <Input.TextArea aria-label={`过滤 ${index + 1} 值`} placeholder="值（允许为空）" autoSize disabled={disabled} value={filter.value} onChange={e => update({ filters: value.rule.filters.map((f, i) => i === index ? { ...f, value: e.target.value } : f) })} />
      <Button danger htmlType="button" disabled={disabled} onClick={() => update({ filters: value.rule.filters.filter((_, i) => i !== index) })}>移除</Button>
    </Space>)}
    <Button htmlType="button" disabled={disabled} onClick={() => update({ filters: [...value.rule.filters, { kind: 'S3Key', name: 'prefix', value: '' }] })}>添加过滤条件</Button>
  </Space>
}

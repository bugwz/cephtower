import { Alert, Button, Checkbox, Input, Select, Space } from 'antd'
import { lifecycleActions, lifecycleFields, lifecycleDraft, lifecycleDocument, type LifecycleDraft, type LifecycleRule, type LifecycleSelector } from './rgwBucketLifecycleForm'

export function RgwBucketLifecycleEditor({ value, onChange, disabled = false, id, storageClassOptions }: { value?: unknown; onChange?: (draft: LifecycleDraft) => void; disabled?: boolean; id?: string; storageClassOptions?:Array<{value:string;label:string}> }) {
  let draft: LifecycleDraft
  try { draft = lifecycleDraft(value) } catch { return <Alert type="error" message="规则不可用，请重新打开表单" /> }
  let warning: string | undefined
  try { lifecycleDocument(draft) } catch (error) { warning = error instanceof Error ? error.message : '配置无效' }
  const change = (rules: LifecycleRule[]) => { if (!disabled) onChange?.({ rules }) }
  const update = (index: number, patch: Partial<LifecycleRule>) => change(draft.rules.map((rule, i) => i === index ? { ...rule, ...patch } : rule))
  return <Space id={id} direction="vertical" style={{ width: '100%' }}>
    <Alert type="warning" message="提交会整体替换规则。过期规则可能永久删除对象或历史版本；请先备份原文。新增规则默认禁用。" />
    <div style={{ maxHeight: 520, overflowY: 'auto', width: '100%' }}>
      {draft.rules.map((rule, index) => {
        const selector = rule.selector
        const filter = (patch: Partial<LifecycleSelector>) => update(index, { selector: { ...selector, ...patch } })
        return <fieldset key={index} disabled={disabled} style={{ marginBottom: 16 }}>
          <legend>规则 {index + 1}</legend>
          <div>规则 ID（留空由 Ceph 生成）</div>
          <Input.TextArea aria-label={`规则 ${index + 1} ID`} disabled={disabled} value={rule.id} onChange={event => update(index, { id: event.target.value })} />
          <Select aria-label={`规则 ${index + 1} 状态`} style={{ width: '100%' }} disabled={disabled} value={rule.status} options={[{ value: 'Enabled', label: '启用' }, { value: 'Disabled', label: '禁用' }]} onChange={status => update(index, { status })} />
          <div>过滤器模式（切换模式保留条件，不会静默丢弃）</div>
          <Select aria-label={`规则 ${index + 1} 过滤器模式`} style={{ width: '100%' }} disabled={disabled} value={selector.kind} options={[{ value: 'Filter', label: 'Filter' }, { value: 'Prefix', label: '旧式 Prefix' }]} onChange={kind => filter({ kind })} />
          <Checkbox disabled={disabled} checked={selector.and} onChange={event => filter({ and: event.target.checked })}>使用 And 包装</Checkbox>
          {(['prefix', 'object_size_greater_than', 'object_size_less_than'] as const).map((key, fieldIndex) => <div key={key}>
            <Checkbox aria-label={`规则 ${index + 1} 设置 ${key}`} disabled={disabled} checked={selector[key] !== null} onChange={event => filter({ [key]: event.target.checked ? '' : null })}>{['设置前缀', '设置对象字节数下界（严格大于）', '设置对象字节数上界（严格小于）'][fieldIndex]}</Checkbox>
            {selector[key] !== null ? <Input.TextArea aria-label={`规则 ${index + 1} ${key}`} disabled={disabled} value={selector[key]!} onChange={event => filter({ [key]: event.target.value })} /> : null}
          </div>)}
          <Checkbox disabled={disabled} checked={selector.archive_zone} onChange={event => filter({ archive_zone: event.target.checked })}>ArchiveZone 条件</Checkbox>
          <div>标签条件（逐条匹配）</div>
          {selector.tags.map((tag, position) => <Space key={position} style={{ width: '100%' }}>
            {(['key', 'value'] as const).map(key => <Input.TextArea key={key} aria-label={`规则 ${index + 1} 标签 ${position + 1} ${key}`} disabled={disabled} value={tag[key]} onChange={event => filter({ tags: selector.tags.map((entry, i) => i === position ? { ...entry, [key]: event.target.value } : entry) })} />)}
            <Button htmlType="button" disabled={disabled} onClick={() => filter({ tags: selector.tags.filter((_, i) => i !== position) })}>移除标签</Button>
          </Space>)}
          <Button htmlType="button" disabled={disabled} onClick={() => filter({ tags: [...selector.tags, { key: '', value: '' }] })}>添加标签</Button>
          {rule.actions.map((action, position) => <fieldset key={position} disabled={disabled} style={{ marginTop: 10 }}>
            <legend>{lifecycleActions[action.type].label} {position + 1}</legend>
            <div>勾选要设置的字段。过期时间三选一，转换时间 Days/Date 二选一。</div>
            {lifecycleActions[action.type].fields.map(key => {
              const present = Object.prototype.hasOwnProperty.call(action.fields, key)
              const fields = (next: Record<string, string>) => update(index, { actions: rule.actions.map((entry, i) => i === position ? { ...entry, fields: next } : entry) })
              return <div key={key}>
                <Checkbox aria-label={`规则 ${index + 1} 动作 ${position + 1} 设置 ${key}`} disabled={disabled} checked={present} onChange={event => { const next = { ...action.fields }; if (event.target.checked) next[key] = key === 'ExpiredObjectDeleteMarker' ? 'false' : ''; else delete next[key]; fields(next) }}>{lifecycleFields[key]}（{key}）</Checkbox>
                {present&&key==='StorageClass'&&storageClassOptions&&<Select<string> aria-label={`规则 ${index + 1} 动作 ${position + 1} 库存转换类候选`} placeholder="选择此桶目标中的库存候选（不会自动选择）" style={{width:'100%'}} disabled={disabled||storageClassOptions.length===0} value={undefined} options={storageClassOptions} showSearch optionFilterProp="label" onChange={item=>{if(storageClassOptions.some(option=>option.value===item))fields({...action.fields,StorageClass:item})}}/>}
                {present ? key === 'ExpiredObjectDeleteMarker'
                  ? <Select aria-label={`规则 ${index + 1} 动作 ${position + 1} ${key}`} disabled={disabled} value={action.fields[key]} options={[{ value: 'true', label: 'true' }, { value: 'false', label: 'false' }]} onChange={item => fields({ ...action.fields, [key]: item })} />
                  : <Input.TextArea aria-label={`规则 ${index + 1} 动作 ${position + 1} ${key}`} disabled={disabled} value={action.fields[key]} onChange={event => fields({ ...action.fields, [key]: event.target.value })} placeholder={key === 'Date' ? '例如 2030-01-01T00:00:00Z（UTC 午夜）' : undefined} /> : null}
              </div>
            })}
            <Button htmlType="button" disabled={disabled} onClick={() => update(index, { actions: rule.actions.filter((_, i) => i !== position) })}>移除动作</Button>
          </fieldset>)}
          <div>添加动作</div>
          <Space wrap>{Object.entries(lifecycleActions).map(([type, definition]) => <Button key={type} htmlType="button" disabled={disabled} onClick={() => update(index, { actions: [...rule.actions, { type, fields: {} }] })}>{definition.label}</Button>)}</Space>
          <div><Button htmlType="button" danger disabled={disabled} onClick={() => change(draft.rules.filter((_, i) => i !== index))}>移除规则</Button></div>
        </fieldset>
      })}
    </div>
    <Button htmlType="button" disabled={disabled} onClick={() => change([...draft.rules, { id: '', status: 'Disabled', selector: { kind: 'Filter', and: false, prefix: null, tags: [], object_size_greater_than: null, object_size_less_than: null, archive_zone: false }, actions: [] }])}>添加规则</Button>
    {warning ? <Alert type="warning" message={warning} /> : null}
  </Space>
}

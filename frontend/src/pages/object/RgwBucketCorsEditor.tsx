import { Alert, Button, Input, InputNumber, Select, Space } from 'antd'
import { corsDraft, corsDocument, corsMethods, type CorsDraft, type CorsRule } from './rgwBucketCorsForm'

export function RgwBucketCorsEditor({ value, onChange, disabled = false, id }: { value?: unknown; onChange?: (draft: CorsDraft) => void; disabled?: boolean; id?: string }) {
  let draft: CorsDraft
  try { draft = corsDraft(value) } catch { return <Alert type="error" message="规则不可用，请重新打开表单" /> }
  let warning: string | undefined
  try { corsDocument(draft) } catch (error) { warning = error instanceof Error ? error.message : '配置无效' }
  const change = (rules: CorsRule[]) => { if (!disabled) onChange?.({ rules }) }
  const update = (index: number, patch: Partial<CorsRule>) => change(draft.rules.map((rule, i) => i === index ? { ...rule, ...patch } : rule))
  const move = (index: number, direction: number) => {
    const next = index + direction
    if (next < 0 || next >= draft.rules.length) return
    const rules = [...draft.rules]; [rules[index], rules[next]] = [rules[next], rules[index]]; change(rules)
  }
  return <Space id={id} direction="vertical" style={{ width: '100%' }}>
    <div style={{ maxHeight: 480, overflowY: 'auto', width: '100%' }}>
      {draft.rules.map((rule, index) => <fieldset key={index} disabled={disabled} style={{ marginBottom: 12 }}>
        <legend>规则 {index + 1}</legend>
        <Input.TextArea aria-label={`规则 ${index + 1} ID`} value={rule.id} disabled={disabled} onChange={event => update(index, { id: event.target.value })} placeholder="规则 ID（可为空）" />
        <Select aria-label={`规则 ${index + 1} 方法`} mode="multiple" style={{ width: '100%' }} disabled={disabled} value={rule.allowed_methods} options={corsMethods.map(method => ({ label: method, value: method }))} onChange={methods => update(index, { allowed_methods: methods })} placeholder="允许方法" />
        {(['allowed_origins', 'allowed_headers', 'expose_headers'] as const).map((key, fieldIndex) => <div key={key}>
          <div>{['允许来源', '允许头', '暴露头（保持顺序及重复项）'][fieldIndex]}</div>
          {rule[key].map((item, position) => <Space key={position} style={{ width: '100%' }}>
            <Input.TextArea aria-label={`规则 ${index + 1} ${key} ${position + 1}`} disabled={disabled} value={item} onChange={event => update(index, { [key]: rule[key].map((text, i) => i === position ? event.target.value : text) })} />
            <Button htmlType="button" disabled={disabled} onClick={() => update(index, { [key]: rule[key].filter((_, i) => i !== position) })}>移除条目</Button>
          </Space>)}
          <Button htmlType="button" disabled={disabled} onClick={() => update(index, { [key]: [...rule[key], ''] })}>添加{['来源', '允许头', '暴露头'][fieldIndex]}</Button>
        </div>)}
        <div>缓存秒数（清空表示未设置，0 表示零秒）</div>
        <InputNumber aria-label={`规则 ${index + 1} 缓存秒数`} disabled={disabled} value={rule.max_age_seconds} min={0} max={4294967294} precision={0} onChange={age => update(index, { max_age_seconds: age })} />
        <Space>
          <Button htmlType="button" disabled={disabled || index === 0} onClick={() => move(index, -1)}>上移</Button>
          <Button htmlType="button" disabled={disabled || index === draft.rules.length - 1} onClick={() => move(index, 1)}>下移</Button>
          <Button htmlType="button" danger disabled={disabled} onClick={() => change(draft.rules.filter((_, i) => i !== index))}>移除规则</Button>
        </Space>
      </fieldset>)}
    </div>
    <Button htmlType="button" disabled={disabled} onClick={() => change([...draft.rules, { id: '', allowed_origins: [''], allowed_methods: [], allowed_headers: [], expose_headers: [], max_age_seconds: null }])}>添加规则</Button>
    {warning ? <Alert type="warning" message={warning} /> : null}
  </Space>
}

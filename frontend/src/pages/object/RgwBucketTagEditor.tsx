import { Alert, Button, Input, Space, Typography } from 'antd'
import { bucketTagFormDocument, bucketTagFormEntries, type BucketTagEntry, type BucketTagFormValue } from './rgwBucketTagForm'

export function RgwBucketTagEditor({ value, onChange, disabled = false, id }: { value?: unknown; onChange?: (value: BucketTagFormValue) => void; disabled?: boolean; id?: string }) {
  const entries = bucketTagFormEntries(value)
  if (!entries) return <Alert type="error" message="标签条目不可用，请关闭并重新打开表单" />
  let validation: string | undefined
  try { bucketTagFormDocument(value) } catch (error) { validation = error instanceof Error ? error.message : '标签无效' }
  const change = (next: BucketTagEntry[]) => { if (!disabled) onChange?.({ entries: next }) }
  return <Space id={id} direction="vertical" style={{ width: '100%' }}>
    <Typography.Text type="secondary">{entries.length}/50 条；键 1–128 字节，值最多 256 字节（UTF-8）。允许空值和重复键，空白属于标签内容。</Typography.Text>
    <div style={{ maxHeight: 360, overflowY: 'auto', width: '100%' }}>
      {entries.map((entry, index) => <Space key={index} align="start" style={{ width: '100%', marginBottom: 8 }}>
        <Input.TextArea aria-label={`第 ${index + 1} 条标签键`} value={entry.key} disabled={disabled} autoSize={{ minRows: 1, maxRows: 3 }} placeholder="标签键" spellCheck={false}
          onChange={event => change(entries.map((item, position) => position === index ? { ...item, key: event.target.value } : item))} />
        <Input.TextArea aria-label={`第 ${index + 1} 条标签值`} value={entry.value} disabled={disabled} autoSize={{ minRows: 1, maxRows: 3 }} placeholder="标签值（允许为空）" spellCheck={false}
          onChange={event => change(entries.map((item, position) => position === index ? { ...item, value: event.target.value } : item))} />
        <Button htmlType="button" danger disabled={disabled} onClick={() => change(entries.filter((_, position) => position !== index))}>移除第 {index + 1} 条</Button>
      </Space>)}
    </div>
    <Space>
      <Button htmlType="button" disabled={disabled || entries.length >= 50} onClick={() => { if (entries.length < 50) change([...entries, { key: '', value: '' }]) }}>添加标签</Button>
      <Button htmlType="button" danger disabled={disabled || entries.length === 0} onClick={() => change([])}>清空表单条目</Button>
    </Space>
    {entries.length === 0 ? <Typography.Text>提交后将设置空标签集合；如需删除标签属性，请使用列表中的删除操作。</Typography.Text> : null}
    {validation ? <Alert type="warning" showIcon message={validation} /> : null}
  </Space>
}

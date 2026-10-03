import { Button, Input, Space, Typography } from 'antd'
import { message } from '../../utils/appMessage'
import { generateRGWCredential, type RGWCredentialKind } from './rgwCredentialGenerator'

export function RgwGeneratedCredentialInput({ kind, value, onChange, id }: { kind: RGWCredentialKind; value?: string; onChange?: (value: string) => void; id?: string }) {
  const generate = () => {
    try {
      const generated = generateRGWCredential(kind)
      onChange?.(generated)
    } catch {
      message.error('无法使用安全随机数生成凭据，原值未修改。请使用受支持的浏览器或自行提供安全保存的凭据。')
    }
  }
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Input.Password id={id} autoComplete="new-password" spellCheck={false} value={value ?? ''} onChange={event => onChange?.(event.target.value)} />
    <Button htmlType="button" onClick={generate}>{value ? '重新生成并替换' : '安全随机生成'}</Button>
    <Typography.Text type="secondary">仅在当前表单中生成；请查看并安全保存后再提交。关闭表单后不能找回，生成不会执行集群操作。</Typography.Text>
  </Space>
}

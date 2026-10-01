import { Alert, Button, Card, Input, Select, Space, Typography } from 'antd'
import { nfsAccessOptions, nfsSquashOptions } from './nfsExportFields'

type ClientRule = { addresses: string[]; access_type?: string | null; squash?: string | null; [key: string]: unknown }

export function NFSClientsEditor({ value, onChange, id }: { value?: string; onChange?: (value: string) => void; id?: string }) {
  let clients: ClientRule[]
  try {
    const parsed: unknown = value ? JSON.parse(value) : []
    if (!Array.isArray(parsed) || parsed.some((row) => !row || typeof row !== 'object' || !Array.isArray(row.addresses) || row.addresses.some((address: unknown) => typeof address !== 'string'))) throw new Error('invalid')
    clients = parsed
  } catch {
    return <Alert type="error" showIcon message="客户端规则格式异常，请刷新导出数据后重试；未修改原有规则。" />
  }
  const save = (rules: ClientRule[]) => onChange?.(JSON.stringify(rules))
  const update = (index: number, patch: Partial<ClientRule>) => save(clients.map((row, i) => i === index ? { ...row, ...patch } : row))
  const options = (base: Array<{ label: string; value: string }>, current?: string | null) => [
    { label: '继承导出设置', value: '' }, ...base,
    ...(current && !base.some((option) => option.value === current) ? [{ label: `当前值：${current}`, value: current }] : [])
  ]
  return <Space id={id} direction="vertical" style={{ width: '100%' }}>
    <Typography.Text type="secondary">规则按顺序保留；无客户端规则时使用导出级配置。删除全部规则可能改变访问范围。</Typography.Text>
    {clients.map((client, index) => <Card key={index} size="small" title={`客户端规则 ${index + 1}`} extra={<Button danger size="small" onClick={() => save(clients.filter((_, i) => i !== index))}>删除规则 {index + 1}</Button>}>
      <Space direction="vertical" style={{ width: '100%' }}>
        <Typography.Text>客户端地址（逗号分隔，支持 IP、CIDR、主机名）</Typography.Text>
        <Input aria-label={`规则 ${index + 1} 地址`} value={client.addresses.join(',')} placeholder="10.0.0.0/8,host.example.com" onChange={(event) => update(index, { addresses: event.target.value.split(',') })} />
        <Typography.Text>访问类型</Typography.Text>
        <Select aria-label={`规则 ${index + 1} 访问类型`} style={{ width: '100%' }} value={client.access_type ?? ''} options={options(nfsAccessOptions, client.access_type)} onChange={(access_type) => update(index, { access_type })} />
        <Typography.Text>身份映射</Typography.Text>
        <Select aria-label={`规则 ${index + 1} 身份映射`} style={{ width: '100%' }} value={client.squash ?? ''} options={options(nfsSquashOptions, client.squash)} onChange={(squash) => update(index, { squash })} />
      </Space>
    </Card>)}
    {!clients.length && <Typography.Text>未配置客户端规则</Typography.Text>}
    <Button onClick={() => save([...clients, { addresses: [''], access_type: '', squash: '' }])}>新增客户端规则</Button>
  </Space>
}

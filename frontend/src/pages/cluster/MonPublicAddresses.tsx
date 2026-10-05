export function MonPublicAddresses({ value }: { value: unknown }) {
  if (!Array.isArray(value)) return <span>地址列表未返回或格式无效</span>
  if (value.length === 0) return <span>原生地址列表为空</span>
  if (!value.every(item => item !== null && typeof item === 'object' && typeof item.type === 'string' && item.type.length > 0 && typeof item.addr === 'string' && item.addr.length > 0)) {
    return <span>地址列表格式无效</span>
  }
  return <ul style={{ margin: 0, paddingInlineStart: 18 }}>{value.map((item, index) => <li key={index}><code>{item.type}: {item.addr}</code></li>)}</ul>
}

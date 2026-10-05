export function MonCounterValue({ value, unit }: { value: unknown; unit: unknown }) {
  const text = typeof value === 'string' && value !== '' ? value : typeof value === 'number' && Number.isFinite(value) ? String(value) : undefined
  if (text === undefined) return <span>未返回或格式无效</span>
  return <span>{text}{typeof unit === 'string' && unit !== '' ? ` ${unit}` : ''}</span>
}

export function monCounterType(value: unknown) {
  if (value === 'counter') return '采样间隔速率'
  if (value === 'gauge') return '采集值'
  if (value === 'histogram') return '直方图'
  return '类型未知'
}

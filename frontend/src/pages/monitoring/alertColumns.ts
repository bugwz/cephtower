import type { FieldColumn } from '../../components/DataTable'

export function alertField(value: unknown, key: string): string {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return '未提供'
  const field = (value as Record<string, unknown>)[key]
  return typeof field === 'string' ? field : '未提供'
}

export const alertColumns: FieldColumn[] = [
  { key: 'alertname', title: '告警名称', render: (_, row) => alertField(row.labels, 'alertname') },
  { key: 'summary', title: '摘要', render: (_, row) => alertField(row.annotations, 'summary') },
  { key: 'severity', title: '严重程度', render: (_, row) => alertField(row.labels, 'severity') },
  { key: 'status', title: '状态', render: (value) => alertField(value, 'state') },
  { key: 'fingerprint', title: '指纹' },
  { key: 'startsAt', title: '开始时间' },
  { key: 'endsAt', title: '结束时间' },
  { key: 'generatorURL', title: '告警来源 URL' }
]

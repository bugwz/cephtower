import type { ExternalListColumn } from '../ExternalListPage'

export function alertField(value: unknown, key: string): string {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return '未提供'
  const field = (value as Record<string, unknown>)[key]
  return typeof field === 'string' ? field : '未提供'
}

export function alertFilter(parent: string, key: string, options: string[], defaultValues?: string[]): NonNullable<ExternalListColumn['localFilter']> {
  return { options, defaultValues, value: (row) => {
    const container = row[parent]
    if (!container || typeof container !== 'object' || Array.isArray(container)) return undefined
    const value = (container as Record<string, unknown>)[key]
    return typeof value === 'string' ? value : undefined
  } }
}

export const alertColumns: ExternalListColumn[] = [
  { key: 'alertname', title: '告警名称', render: (_, row) => alertField(row.labels, 'alertname') },
  { key: 'summary', title: '摘要', render: (_, row) => alertField(row.annotations, 'summary') },
  { key: 'severity', title: '严重程度', render: (_, row) => alertField(row.labels, 'severity'), localFilter: alertFilter('labels', 'severity', ['critical', 'warning', 'info']) },
  { key: 'status', title: '状态', render: (value) => alertField(value, 'state'), localFilter: alertFilter('status', 'state', ['active', 'suppressed', 'unprocessed'], ['active']) },
  { key: 'fingerprint', title: '指纹' },
  { key: 'startsAt', title: '开始时间' },
  { key: 'endsAt', title: '结束时间' },
  { key: 'generatorURL', title: '告警来源 URL' }
]

export const alertRuleColumns: ExternalListColumn[] = [
  { key: 'name', title: '名称' },
  { key: 'severity', title: '严重程度', render: (_, row) => alertField(row.labels, 'severity'), localFilter: alertFilter('labels', 'severity', ['critical', 'warning', 'info']) },
  { key: 'group', title: '规则组' },
  { key: 'file', title: '规则文件' },
  { key: 'state', title: '状态' },
  { key: 'health', title: '健康状态' },
  { key: 'duration', title: '持续时间（秒）' },
  { key: 'query', title: '表达式' },
  { key: 'summary', title: '摘要', render: (_, row) => alertField(row.annotations, 'summary') }
]

export const silenceColumns: ExternalListColumn[] = [
  { key: 'id', title: 'ID' },
  { key: 'status', title: '状态', render: (value) => alertField(value, 'state'), localFilter: alertFilter('status', 'state', ['active', 'pending', 'expired']) },
  { key: 'matchers', title: 'Matchers' },
  { key: 'startsAt', title: '开始时间' },
  { key: 'updatedAt', title: '更新时间' },
  { key: 'endsAt', title: '结束时间' },
  { key: 'createdBy', title: '创建人' },
  { key: 'comment', title: '说明' }
]

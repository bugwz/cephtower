import { Descriptions, Table, Typography } from 'antd'
import { textValue, type ApiRecord } from '../api/client'
import { formatDateTime, isDateTimeField } from '../utils/time'

const { Paragraph, Text } = Typography

interface RecordDetailProps {
  record?: ApiRecord | null
  title?: string
  preferredKeys?: string[]
}

const defaultPreferredKeys = [
  'name',
  'natural_key',
  'status',
  'kind',
  'hostname',
  'service_name',
  'daemon_name',
  'pool_name',
  'image_spec',
  'resource_version',
  'observed_at',
  'source',
  'stale'
]

export function RecordDetail({ record, preferredKeys = defaultPreferredKeys }: RecordDetailProps) {
  const rows = detailRows(record, preferredKeys)

  if (!record) {
    return <Text type="secondary">暂无详情</Text>
  }

  return (
    <div className="record-detail">
      <Descriptions size="small" column={1} bordered>
        {rows.map(([key, value]) => (
          <Descriptions.Item key={key} label={key}>
            {renderDetailValue(value, key)}
          </Descriptions.Item>
        ))}
      </Descriptions>
      <Paragraph className="record-detail-payload" copyable>
        {formatJSON(record)}
      </Paragraph>
    </div>
  )
}

function detailRows(record: ApiRecord | null | undefined, preferredKeys: string[]) {
  if (!record) {
    return []
  }

  const keys = [
    ...preferredKeys.filter((key) => key in record),
    ...Object.keys(record).filter((key) => !preferredKeys.includes(key))
  ]
  return keys.map((key) => [key, record[key]] as const)
}

function renderDetailValue(value: unknown, key: string) {
  if (isDateTimeField(key)) {
    return formatDateTime(value)
  }

  if (typeof value === 'boolean') {
    return value ? '是' : '否'
  }
  if (Array.isArray(value) && value.length && value.every((item) => item !== null && typeof item === 'object' && !Array.isArray(item))) {
    const records = value as ApiRecord[]
    const keys = Array.from(new Set(records.flatMap((item) => Object.keys(item))))
    return <Table<ApiRecord>
      size="small"
      dataSource={records}
      rowKey={(_record, index) => String(index)}
      columns={keys.map((field) => ({ title: field, key: field, dataIndex: field, render: (cell: unknown) => isDateTimeField(field) ? formatDateTime(cell) : cell !== null && typeof cell === 'object' ? <pre style={{ margin: 0, maxWidth: 400, whiteSpace: 'pre-wrap' }}>{JSON.stringify(cell, null, 2)}</pre> : typeof cell === 'boolean' ? (cell ? '是' : '否') : textValue(cell, '-') }))}
      pagination={records.length > 10 ? { pageSize: 10, showSizeChanger: false } : false}
      scroll={{ x: 'max-content' }}
    />
  }
  if (value !== null && typeof value === 'object' && !Array.isArray(value)) {
    return <pre style={{ margin: 0, maxHeight: 400, overflow: 'auto', whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{JSON.stringify(value, null, 2)}</pre>
  }
  if (Array.isArray(value)) {
    return value.length ? value.map((item) => textValue(item, '')).filter(Boolean).join(', ') : '-'
  }
  return textValue(value, '-')
}

function formatJSON(record: ApiRecord) {
  try {
    return JSON.stringify(record, null, 2)
  } catch {
    return textValue(record)
  }
}

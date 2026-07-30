import { ReloadOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, InputNumber, Space, Table, Tag } from 'antd'
import { useCallback, useState } from 'react'
import { listAuditEvents, type AuditEventView } from '../../api/audit'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'
import { useClusterContext } from '../../state/ClusterContext'

export function AuditPage() {
  const { selectedClusterId } = useClusterContext()
  const [form] = Form.useForm<AuditSearchForm>()
  const [filters, setFilters] = useState<AuditSearchForm>({ limit: 100 })
  const [refreshing, setRefreshing] = useState(false)
  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return []
    }
    return listAuditEvents({
      clusterId: selectedClusterId,
      username: clean(filters.username),
      action: clean(filters.action),
      resourceKind: clean(filters.resourceKind),
      resourceKey: clean(filters.resourceKey),
      userId: filters.userId,
      limit: filters.limit ?? 100
    })
  }, [filters, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)

  async function reload() {
    setRefreshing(true)
    try {
      await refresh()
    } finally {
      setRefreshing(false)
    }
  }

  async function search(values: AuditSearchForm) {
    setFilters({
      username: clean(values.username),
      action: clean(values.action),
      resourceKind: clean(values.resourceKind),
      resourceKey: clean(values.resourceKey),
      userId: values.userId,
      limit: values.limit ?? 100
    })
  }

  async function reset() {
    form.resetFields()
    setFilters({ limit: 100 })
  }

  return (
    <Page title="审计事件" loading={loading} error={error}>
      <Card
        className="page-surface-card"
        title="审计事件"
        extra={<Button icon={<ReloadOutlined />} loading={refreshing} onClick={reload}>刷新</Button>}
      >
        <Space direction="vertical" size={16} className="page-stack">
          <Form
            form={form}
            layout="inline"
            initialValues={{ limit: 100 }}
            onFinish={search}
            className="audit-filter-form"
          >
            <Form.Item name="username" label="用户">
              <Input allowClear placeholder="actor username" />
            </Form.Item>
            <Form.Item name="action" label="Action">
              <Input allowClear placeholder="pool.create" />
            </Form.Item>
            <Form.Item name="resourceKind" label="资源类型">
              <Input allowClear placeholder="pool" />
            </Form.Item>
            <Form.Item name="resourceKey" label="资源 Key">
              <Input allowClear placeholder="natural key" />
            </Form.Item>
            <Form.Item name="userId" label="用户 ID">
              <InputNumber min={1} precision={0} />
            </Form.Item>
            <Form.Item name="limit" label="数量">
              <InputNumber min={1} max={500} precision={0} />
            </Form.Item>
            <Form.Item>
              <Space>
                <Button type="primary" htmlType="submit">查询</Button>
                <Button onClick={reset}>重置</Button>
              </Space>
            </Form.Item>
          </Form>
          <Table<AuditEventView>
            size="middle"
            rowKey="audit_event_id"
            dataSource={data ?? []}
            pagination={{ pageSize: 10, showSizeChanger: false }}
            scroll={{ x: 1480 }}
            columns={[
              { title: '时间', dataIndex: 'occurred_at', width: 180, render: formatTime },
              { title: '类型', dataIndex: 'event_type', width: 120 },
              { title: '用户', dataIndex: 'actor_username', width: 130 },
              { title: 'Action', dataIndex: 'action', width: 190, ellipsis: true },
              { title: '资源', width: 300, render: (_, row) => `${row.resource_kind ?? '-'} / ${row.resource_key ?? '-'}` },
              { title: '结果', dataIndex: 'outcome', width: 120, render: (value) => <Tag color={value === 'succeeded' || value === 'accepted' ? 'success' : value === 'failed' ? 'error' : 'default'}>{value}</Tag> },
              { title: 'HTTP', dataIndex: 'http_status', width: 90, render: (value) => value ?? '-' },
              { title: '错误码', dataIndex: 'error_code', width: 170, ellipsis: true, render: (value) => value || '-' },
              { title: '风险', dataIndex: 'risk', width: 90, render: (value) => value || '-' },
              { title: 'Request ID', dataIndex: 'request_id', width: 180, ellipsis: true },
              { title: 'Hash', dataIndex: 'event_hash', width: 220, ellipsis: true }
            ]}
          />
        </Space>
      </Card>
    </Page>
  )
}

interface AuditSearchForm {
  username?: string
  action?: string
  resourceKind?: string
  resourceKey?: string
  userId?: number
  limit?: number
}

function clean(value?: string) {
  const text = value?.trim()
  return text || undefined
}

function formatTime(value?: string | null) {
  if (!value) {
    return '-'
  }
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

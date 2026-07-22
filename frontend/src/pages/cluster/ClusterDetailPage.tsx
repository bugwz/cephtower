import { ArrowLeftOutlined, ReloadOutlined } from '@ant-design/icons'
import { Button, Card, Descriptions, Space, Table, Tag, Typography } from 'antd'
import { useCallback } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { getClusterDetail, type CephResourceSnapshot } from '../../api/clusters'
import { textValue } from '../../api/client'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'

const { Paragraph, Text } = Typography

export function ClusterDetailPage() {
  const navigate = useNavigate()
  const { id = '' } = useParams()
  const loader = useCallback(() => getClusterDetail(id), [id])
  const { data, loading, error, refresh } = useResource(loader)
  const cluster = data?.cluster

  return (
    <Page title="集群详情" loading={loading} error={error}>
      <Space direction="vertical" size={16} className="page-stack">
        <Card
          className="page-surface-card"
          title={cluster?.name ?? '集群详情'}
          extra={
            <Space>
              <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/cluster/clusters')}>
                返回
              </Button>
              <Button icon={<ReloadOutlined />} onClick={refresh}>
                刷新
              </Button>
            </Space>
          }
        >
          <Descriptions size="small" column={{ xs: 1, sm: 2, lg: 3 }} bordered>
            <Descriptions.Item label="集群 ID">{cluster?.id}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={cluster?.enabled ? 'success' : 'default'}>{cluster?.enabled ? '启用' : '禁用'}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="FSID">{cluster?.fsid || '未发现'}</Descriptions.Item>
            <Descriptions.Item label="Dashboard URL">{cluster?.dashboard.base_url || '未发现'}</Descriptions.Item>
            <Descriptions.Item label="Dashboard 用户">{cluster?.dashboard.username || '-'}</Descriptions.Item>
            <Descriptions.Item label="Dashboard 密码">
              <Tag color={cluster?.dashboard.password_set ? 'gold' : 'default'}>{cluster?.dashboard.password_set ? '已保存' : '未保存'}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Ceph 命令">{cluster?.command.bin || 'ceph'}</Descriptions.Item>
            <Descriptions.Item label="Ceph 用户">{cluster?.command.name || 'client.admin'}</Descriptions.Item>
            <Descriptions.Item label="Keyring">
              <Tag color={cluster?.command.keyring_content_set ? 'gold' : 'default'}>{cluster?.command.keyring_content_set ? '已保存' : '未保存'}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="更新时间">{cluster?.updated_at ? new Date(cluster.updated_at).toLocaleString() : '-'}</Descriptions.Item>
          </Descriptions>
        </Card>

        <Card className="page-surface-card" title="资源快照">
          <Table
            size="middle"
            rowKey={(row) => `${row.category}:${row.resource_key}`}
            dataSource={data?.snapshots ?? []}
            pagination={{ pageSize: 8, showSizeChanger: false }}
            expandable={{
              expandedRowRender: (row) => (
                <Paragraph className="snapshot-payload" copyable>
                  {formatSnapshotPayload(row.payload)}
                </Paragraph>
              )
            }}
            columns={[
              { title: '类别', dataIndex: 'category' },
              { title: '资源键', dataIndex: 'resource_key' },
              {
                title: '同步状态',
                key: 'sync_status',
                render: (_, row) => row.last_error ? <Tag color="error">失败</Tag> : <Tag color="success">成功</Tag>
              },
              {
                title: '最后同步',
                dataIndex: 'last_synced_at',
                render: (value: string) => value ? new Date(value).toLocaleString() : '-'
              },
              {
                title: '错误信息',
                dataIndex: 'last_error',
                render: (value: string) => value ? <Text type="danger">{value}</Text> : <Text type="secondary">-</Text>
              },
              {
                title: '数据预览',
                dataIndex: 'payload',
                render: (value: unknown) => textValue(previewSnapshotPayload(value))
              }
            ]}
          />
        </Card>
      </Space>
    </Page>
  )
}

function previewSnapshotPayload(payload: unknown) {
  if (Array.isArray(payload)) {
    return `${payload.length} 条记录`
  }
  if (payload && typeof payload === 'object') {
    return Object.keys(payload).slice(0, 6).join(', ')
  }
  return payload
}

function formatSnapshotPayload(payload: CephResourceSnapshot['payload']) {
  try {
    return JSON.stringify(payload, null, 2)
  } catch {
    return textValue(payload)
  }
}

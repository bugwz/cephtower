import { Card, Tabs, Timeline, Typography } from 'antd'
import { useCallback } from 'react'
import { asArray, isRecord, textValue } from '../api/client'
import { listLogs } from '../api/resources'
import { DataTable } from '../components/DataTable'
import { Page } from '../components/Page'
import { useResource } from '../hooks'

const { Text } = Typography

export function LogsPage() {
  const loader = useCallback(() => listLogs(), [])
  const { data, loading, error, refresh } = useResource(loader)
  const audit = asArray(data?.audit_log ?? data?.audit ?? [])
  const cluster = asArray(data?.clog ?? data?.cluster_log ?? data?.cluster ?? [])
  const rawRows = asArray(data)

  return (
    <Page
      title="运行日志"
      description="集中查看 Ceph Dashboard API 返回的集群日志与审计日志。"
      loading={loading}
      error={error}
      onRefresh={refresh}
    >
      <Tabs
        items={[
          {
            key: 'cluster',
            label: '集群日志',
            children: (
              <Card>
                {cluster.length ? <LogTimeline rows={cluster} /> : <Text type="secondary">暂无集群日志</Text>}
              </Card>
            )
          },
          {
            key: 'audit',
            label: '审计日志',
            children: (
              <Card>
                {audit.length ? (
                  <DataTable
                    data={audit}
                    rowKeyCandidates={['stamp', 'time', 'message']}
                    columns={[
                      { key: 'stamp', title: '时间' },
                      { key: 'priority', title: '级别' },
                      { key: 'channel', title: '通道' },
                      { key: 'message', title: '内容' },
                      { key: 'name', title: '来源' }
                    ]}
                  />
                ) : (
                  <Text type="secondary">暂无审计日志</Text>
                )}
              </Card>
            )
          },
          {
            key: 'raw',
            label: '原始返回',
            children: (
              <Card>
                <DataTable
                  data={rawRows}
                  columns={[
                    { key: 'audit_log', title: 'Audit' },
                    { key: 'clog', title: 'Cluster' },
                    { key: 'logs', title: 'Logs' }
                  ]}
                />
              </Card>
            )
          }
        ]}
      />
    </Page>
  )
}

function LogTimeline({ rows }: { rows: Record<string, unknown>[] }) {
  return (
    <Timeline
      items={rows.slice(0, 30).map((row, index) => ({
        color: textValue(row.priority).toLowerCase().includes('err') ? 'red' : 'blue',
        children: (
          <div className="timeline-row">
            <Text strong>{textValue(row.stamp ?? row.time, `#${index + 1}`)}</Text>
            <Text>{textValue(row.message ?? row.msg ?? (isRecord(row) ? row : undefined))}</Text>
          </div>
        )
      }))}
    />
  )
}


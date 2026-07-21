import { Card, Tabs } from 'antd'
import { useCallback } from 'react'
import { listDaemons, listServices } from '../../api/resources'
import { DataTable } from '../../components/DataTable'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'

export function ServiceManagementPage() {
  const loader = useCallback(async () => {
    const [services, daemons] = await Promise.all([listServices(), listDaemons()])
    return { services, daemons }
  }, [])
  const { data, loading, error, refresh } = useResource(loader)

  return (
    <Page
      title="服务与守护进程"
      description="对应 Ceph Dashboard 的 Services 和 Daemons，用于查看编排服务和 daemon 状态。"
      loading={loading}
      error={error}
      onRefresh={refresh}
    >
      <Tabs
        items={[
          {
            key: 'services',
            label: '服务',
            children: (
              <Card>
                <DataTable
                  data={data?.services ?? []}
                  rowKeyCandidates={['service_name', 'service_id', 'name']}
                  columns={[
                    { key: 'service_name', title: '服务名' },
                    { key: 'service_type', title: '类型' },
                    { key: 'placement', title: '放置策略' },
                    { key: 'status', title: '状态' },
                    { key: 'running', title: '运行数' },
                    { key: 'size', title: '目标数' }
                  ]}
                />
              </Card>
            )
          },
          {
            key: 'daemons',
            label: '守护进程',
            children: (
              <Card>
                <DataTable
                  data={data?.daemons ?? []}
                  rowKeyCandidates={['daemon_name', 'name', 'hostname']}
                  columns={[
                    { key: 'daemon_name', title: 'Daemon' },
                    { key: 'daemon_type', title: '类型' },
                    { key: 'hostname', title: '主机' },
                    { key: 'status_desc', title: '状态' },
                    { key: 'version', title: '版本' },
                    { key: 'container_image_name', title: '镜像' }
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

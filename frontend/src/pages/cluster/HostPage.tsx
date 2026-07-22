import { Card, Tabs, Tag } from 'antd'
import { useCallback } from 'react'
import { listHosts, listOSDFlags, listOSDs } from '../../api/resource'
import { DataTable } from '../../components/DataTable'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'

export function HostPage() {
  const loader = useCallback(async () => {
    const [hosts, osds, flags] = await Promise.all([listHosts(), listOSDs(), listOSDFlags()])
    return { hosts, osds, flags }
  }, [])
  const { data, loading, error, refresh } = useResource(loader)

  return (
    <Page
      title="主机管理"
      loading={loading}
      error={error}
    >
      <Card className="page-surface-card" title="主机管理">
        <Tabs
          items={[
            {
              key: 'hosts',
              label: '主机',
              children: (
                <div className="embedded-panel">
                <DataTable
                  data={data?.hosts ?? []}
                  rowKeyCandidates={['hostname', 'addr']}
                  columns={[
                    { key: 'hostname', title: '主机名' },
                    { key: 'addr', title: '地址' },
                    { key: 'status', title: '状态' },
                    { key: 'ceph_version', title: 'Ceph 版本' },
                    { key: 'labels', title: '标签' },
                    { key: 'service_instances', title: '服务实例' }
                  ]}
                />
                </div>
              )
            },
            {
              key: 'osds',
              label: 'OSD',
              children: (
                <div className="embedded-panel">
                <DataTable
                  data={data?.osds ?? []}
                  rowKeyCandidates={['id', 'osd', 'service_id', 'name']}
                  columns={[
                    { key: 'id', title: 'ID' },
                    { key: 'hostname', title: '主机' },
                    { key: 'state', title: '状态' },
                    { key: 'up', title: 'Up' },
                    { key: 'in', title: 'In' },
                    { key: 'device_class', title: '设备类型' },
                    { key: 'stats', title: '容量/统计' }
                  ]}
                />
                </div>
              )
            },
            {
              key: 'flags',
              label: 'OSD Flags',
              children: (
                <div className="embedded-panel">
                {(data?.flags ?? []).length ? (
                  data?.flags.map((flag) => <Tag key={flag}>{flag}</Tag>)
                ) : (
                  <span className="muted">未设置 OSD flags</span>
                )}
                </div>
              )
            }
          ]}
        />
      </Card>
    </Page>
  )
}

import { Card, Descriptions, Tabs, Tag, Typography } from 'antd'
import { useCallback } from 'react'
import { textValue } from '../../api/client'
import {
  getBlockMirroringSummary,
  listBlockImages,
  listFilesystems,
  listObjectBuckets,
  listObjectGateways,
  listObjectUsers,
  listPools
} from '../../api/resources'
import { DataTable } from '../../components/DataTable'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'

const { Text } = Typography

export function StorageManagementPage() {
  const loader = useCallback(async () => {
    const [pools, images, mirroring, filesystems, gateways, users, buckets] = await Promise.all([
      listPools(),
      listBlockImages(),
      getBlockMirroringSummary(),
      listFilesystems(),
      listObjectGateways(),
      listObjectUsers(),
      listObjectBuckets()
    ])
    return { pools, images, mirroring, filesystems, gateways, users, buckets }
  }, [])
  const { data, loading, error, refresh } = useResource(loader)

  return (
    <Page
      title="存储管理"
      description="把 Ceph 的 Pools、Block、CephFS 和 Object Gateway 合并为 CephTower 的存储视图。"
      loading={loading}
      error={error}
      onRefresh={refresh}
    >
      <Tabs
        items={[
          {
            key: 'pools',
            label: '存储池',
            children: (
              <Card>
                <DataTable
                  data={data?.pools ?? []}
                  rowKeyCandidates={['pool_name', 'pool', 'name']}
                  columns={[
                    { key: 'pool_name', title: '名称' },
                    { key: 'type', title: '类型' },
                    { key: 'pg_num', title: 'PG' },
                    { key: 'pg_placement_num', title: 'PGP' },
                    { key: 'size', title: '副本数' },
                    { key: 'application_metadata', title: '应用' }
                  ]}
                />
              </Card>
            )
          },
          {
            key: 'block',
            label: '块设备',
            children: (
              <div className="content-stack">
                <Card title="RBD 镜像">
                  <DataTable
                    data={data?.images ?? []}
                    rowKeyCandidates={['id', 'name', 'image']}
                    columns={[
                      { key: 'name', title: '名称' },
                      { key: 'pool_name', title: '存储池' },
                      { key: 'size', title: '大小' },
                      { key: 'features_name', title: '特性' },
                      { key: 'num_objs', title: '对象数' },
                      { key: 'namespace', title: '命名空间' }
                    ]}
                  />
                </Card>
                <Card title="镜像同步">
                  <Descriptions size="small" column={3}>
                    {Object.entries(data?.mirroring ?? {}).map(([key, value]) => (
                      <Descriptions.Item key={key} label={key}>
                        <Tag>{textValue(value)}</Tag>
                      </Descriptions.Item>
                    ))}
                  </Descriptions>
                </Card>
              </div>
            )
          },
          {
            key: 'filesystems',
            label: '文件系统',
            children: (
              <Card>
                <DataTable
                  data={data?.filesystems ?? []}
                  rowKeyCandidates={['id', 'name', 'mdsmap']}
                  columns={[
                    { key: 'id', title: 'ID' },
                    { key: 'name', title: '名称' },
                    { key: 'metadata_pool', title: '元数据池' },
                    { key: 'data_pools', title: '数据池' },
                    { key: 'standby_count_wanted', title: 'Standby' },
                    { key: 'mdsmap', title: 'MDS Map' }
                  ]}
                />
              </Card>
            )
          },
          {
            key: 'object',
            label: '对象网关',
            children: (
              <div className="content-stack">
                <Card title="RGW Daemons">
                  <DataTable
                    data={data?.gateways ?? []}
                    rowKeyCandidates={['id', 'name', 'hostname']}
                    columns={[
                      { key: 'id', title: 'ID' },
                      { key: 'hostname', title: '主机' },
                      { key: 'version', title: '版本' },
                      { key: 'server_hostname', title: '服务主机' },
                      { key: 'zonegroup_name', title: 'Zonegroup' }
                    ]}
                  />
                </Card>
                <div className="content-grid two-columns">
                  <Card title="用户">
                    {(data?.users ?? []).length ? (
                      <DataTable
                        data={data?.users ?? []}
                        rowKeyCandidates={['uid', 'user_id', 'id']}
                        columns={[
                          { key: 'uid', title: 'UID' },
                          { key: 'display_name', title: '显示名' },
                          { key: 'email', title: '邮箱' },
                          { key: 'suspended', title: 'Suspended' }
                        ]}
                      />
                    ) : (
                      <Text type="secondary">暂无用户数据</Text>
                    )}
                  </Card>
                  <Card title="Buckets">
                    {(data?.buckets ?? []).length ? (
                      <DataTable
                        data={data?.buckets ?? []}
                        rowKeyCandidates={['bucket', 'name', 'id']}
                        columns={[
                          { key: 'bucket', title: 'Bucket' },
                          { key: 'owner', title: 'Owner' },
                          { key: 'usage', title: 'Usage' },
                          { key: 'num_shards', title: 'Shards' }
                        ]}
                      />
                    ) : (
                      <Text type="secondary">暂无 bucket 数据</Text>
                    )}
                  </Card>
                </div>
              </div>
            )
          }
        ]}
      />
    </Page>
  )
}

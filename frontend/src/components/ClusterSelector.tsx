import { ClusterOutlined, ReloadOutlined } from '@ant-design/icons'
import { Button, Select, Space, Tooltip, Typography } from 'antd'
import { useClusterContext } from '../state/ClusterContext'

const { Text } = Typography

export function ClusterSelector() {
  const { clusters, selectedClusterId, loading, error, refreshClusters, selectCluster } = useClusterContext()

  return (
    <Space size={8} className="cluster-selector">
      <ClusterOutlined className="cluster-selector-icon" />
      <Select
        className="cluster-selector-select"
        size="middle"
        loading={loading}
        value={selectedClusterId}
        placeholder={loading ? '加载集群...' : '选择集群'}
        options={clusters.map((cluster) => ({
          value: cluster.id,
          label: cluster.name || `Cluster ${cluster.id}`
        }))}
        notFoundContent={error ? <Text type="danger">{error}</Text> : '暂无集群'}
        onChange={(value) => selectCluster(value)}
      />
      <Tooltip title="刷新集群列表">
        <Button
          className="icon-button"
          icon={<ReloadOutlined />}
          loading={loading}
          onClick={() => refreshClusters()}
          aria-label="刷新集群列表"
        />
      </Tooltip>
    </Space>
  )
}


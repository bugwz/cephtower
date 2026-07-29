import { Select, Typography } from 'antd'
import { useClusterContext } from '../state/ClusterContext'

const { Text } = Typography

export function ClusterSelector() {
  const { clusters, selectedClusterId, loading, error, selectCluster } = useClusterContext()

  return (
    <Select
      className="cluster-selector"
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
  )
}

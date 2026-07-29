import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import { listClusters, type CephCluster } from '../api/cluster'
import { currentClusterId, setStoredClusterId } from '../api/resource'

interface ClusterContextValue {
  clusters: CephCluster[]
  selectedCluster?: CephCluster
  selectedClusterId?: number
  loading: boolean
  error: string
  refreshClusters: () => Promise<void>
  selectCluster: (clusterId: number | undefined) => void
}

const ClusterContext = createContext<ClusterContextValue | null>(null)

export function ClusterProvider({ enabled, children }: { enabled: boolean; children: React.ReactNode }) {
  const [clusters, setClusters] = useState<CephCluster[]>([])
  const [selectedClusterId, setSelectedClusterId] = useState<number | undefined>(() => currentClusterId())
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const refreshClusters = useCallback(async () => {
    if (!enabled) {
      setClusters([])
      setSelectedClusterId(undefined)
      setStoredClusterId(undefined)
      return
    }
    setLoading(true)
    setError('')
    try {
      const rows = await listClusters()
      setClusters(rows)
      setSelectedClusterId((current) => {
        const stillExists = current && rows.some((cluster) => cluster.id === current)
        const next = stillExists ? current : rows[0]?.id
        setStoredClusterId(next)
        return next
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载集群失败')
    } finally {
      setLoading(false)
    }
  }, [enabled])

  useEffect(() => {
    refreshClusters()
  }, [refreshClusters])

  const selectedCluster = useMemo(
    () => clusters.find((cluster) => cluster.id === selectedClusterId),
    [clusters, selectedClusterId]
  )

  const value = useMemo<ClusterContextValue>(() => ({
    clusters,
    selectedCluster,
    selectedClusterId,
    loading,
    error,
    refreshClusters,
    selectCluster: (clusterId) => {
      setSelectedClusterId(clusterId)
      setStoredClusterId(clusterId)
    }
  }), [clusters, error, loading, refreshClusters, selectedCluster, selectedClusterId])

  return <ClusterContext.Provider value={value}>{children}</ClusterContext.Provider>
}

export function useClusterContext() {
  const value = useContext(ClusterContext)
  if (!value) {
    throw new Error('useClusterContext must be used inside ClusterProvider')
  }
  return value
}


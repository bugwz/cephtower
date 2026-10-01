import { useCallback, useEffect, useRef, useState } from 'react'
import { useClusterContext } from './state/ClusterContext'

interface ResourceRefreshOptions {
  showLoading?: boolean
}

export function useResource<T>(loader: () => Promise<T>) {
  const { selectedClusterId } = useClusterContext()
  const scopeRef = useRef({ loader, clusterId: selectedClusterId, sequence: 0, active: true })
  if (scopeRef.current.loader !== loader || scopeRef.current.clusterId !== selectedClusterId) {
    scopeRef.current = { loader, clusterId: selectedClusterId, sequence: 0, active: true }
  }
  const scope = scopeRef.current
  const [state, setState] = useState<{ scope: typeof scope, data: T | null, loading: boolean, error: string }>({ scope, data: null, loading: true, error: '' })

  const load = useCallback(async ({ showLoading = true }: ResourceRefreshOptions = {}) => {
    if (scopeRef.current !== scope || !scope.active) return
    const sequence = ++scope.sequence
    const current = () => scopeRef.current === scope && scope.active && scope.sequence === sequence
    setState((previous) => ({ scope, data: previous.scope === scope ? previous.data : null, loading: showLoading || (previous.scope === scope && previous.loading), error: '' }))
    try {
      const data = await loader()
      if (current()) setState({ scope, data, loading: false, error: '' })
    } catch (err) {
      if (current()) setState((previous) => ({ scope, data: previous.scope === scope ? previous.data : null, loading: false, error: err instanceof Error ? err.message : '请求失败' }))
    }
  }, [loader, scope])

  useEffect(() => {
    scope.active = true
    void load()
    return () => {
      scope.active = false
      scope.sequence += 1
    }
  }, [load, scope])

  // Hide the previous resource immediately, before the new effect has run.
  return state.scope === scope
    ? { data: state.data, loading: state.loading, error: state.error, refresh: load }
    : { data: null, loading: true, error: '', refresh: load }
}

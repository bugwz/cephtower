import { useEffect, useMemo, useState } from 'react'
import { listClusterCapabilities } from '../api/cluster'
import { listEndpoints } from '../api/endpoint'

export interface FeatureRequirements {
  requiredCapabilities?: string[]
  requiredEndpoints?: string[]
}

export interface FeatureRequirementStatus {
  loading: boolean
  error: string
  blocked: boolean
  reasons: string[]
}

export function useFeatureRequirements(clusterId: number | null | undefined, requirements: FeatureRequirements): FeatureRequirementStatus {
  const capabilityKey = (requirements.requiredCapabilities ?? []).join('\u0000')
  const endpointKey = (requirements.requiredEndpoints ?? []).join('\u0000')
  const requiredCapabilities = useMemo(() => normalize(requirements.requiredCapabilities), [capabilityKey])
  const requiredEndpoints = useMemo(() => normalize(requirements.requiredEndpoints), [endpointKey])
  const [status, setStatus] = useState<FeatureRequirementStatus>({
    loading: false,
    error: '',
    blocked: false,
    reasons: []
  })

  useEffect(() => {
    let ignore = false

    async function load() {
      if (!clusterId || (requiredCapabilities.length === 0 && requiredEndpoints.length === 0)) {
        setStatus({ loading: false, error: '', blocked: false, reasons: [] })
        return
      }

      setStatus((current) => ({ ...current, loading: true, error: '' }))
      try {
        const [capabilities, endpoints] = await Promise.all([
          requiredCapabilities.length ? listClusterCapabilities(clusterId) : Promise.resolve([]),
          requiredEndpoints.length ? listEndpoints(clusterId) : Promise.resolve([])
        ])

        const capabilityByName = new Map(capabilities.map((item) => [item.name.toLowerCase(), item]))
        const enabledEndpoints = new Set(
          endpoints
            .filter((item) => item.enabled !== false)
            .map((item) => item.kind.toLowerCase())
        )
        const reasons = [
          ...requiredCapabilities.flatMap((name) => {
            const capability = capabilityByName.get(name)
            if (capability?.supported) {
              return []
            }
            return [`缺少 capability: ${name}${capability?.reason ? ` (${capability.reason})` : ''}`]
          }),
          ...requiredEndpoints
            .filter((kind) => !enabledEndpoints.has(kind))
            .map((kind) => `未配置或未启用 endpoint: ${kind}`)
        ]

        if (!ignore) {
          setStatus({
            loading: false,
            error: '',
            blocked: reasons.length > 0,
            reasons
          })
        }
      } catch (err) {
        if (!ignore) {
          setStatus({
            loading: false,
            error: err instanceof Error ? err.message : '功能依赖检查失败',
            blocked: true,
            reasons: []
          })
        }
      }
    }

    void load()

    return () => {
      ignore = true
    }
  }, [clusterId, requiredCapabilities, requiredEndpoints])

  return status
}

function normalize(values?: string[]) {
  return Array.from(new Set((values ?? []).map((item) => item.trim().toLowerCase()).filter(Boolean)))
}

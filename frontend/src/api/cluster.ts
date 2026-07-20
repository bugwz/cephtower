export interface ClusterSummary {
  health_status: string
  version?: string
}

const apiBaseUrl = '/api'

export async function getClusterSummary(): Promise<ClusterSummary> {
  const response = await fetch(`${apiBaseUrl}/v1/cluster/summary`)
  if (!response.ok) {
    throw new Error(`Request failed: ${response.status}`)
  }

  return response.json()
}

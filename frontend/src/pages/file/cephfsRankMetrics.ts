import { performancePoints, type PerformanceSample } from './cephfsPerformanceSeries'

export interface RankIdentity { rank: string; name: string; gid: string; state: string }
export type NamedSample = PerformanceSample & { name: string; error?: string }

export function rankSample(row: RankIdentity, samples: NamedSample[]) {
  return row.name && row.gid ? samples.find((sample) => sample.name === row.name && sample.gid === row.gid && !sample.error) : undefined
}

export function rankActivity(row: RankIdentity, samples: NamedSample[], history: Record<string, PerformanceSample[]>): number | undefined {
  const sample = rankSample(row, samples)
  if (!sample || !['active', 'standby-replay'].includes(row.state)) return undefined
  const counter = row.state === 'standby-replay' ? 'mds_log.replay' : 'mds_server.handle_client_request'
  const samplesForDaemon = history[row.name] ?? []
  if (samplesForDaemon[samplesForDaemon.length - 1]?.gid !== row.gid) return undefined
  const series = performancePoints(samplesForDaemon, counter, true)
  const latest = series[series.length - 1]
  return latest && latest.time === new Date(sample.observed_at).getTime() ? latest.value : undefined
}

export function rankClientCount(ranks: RankIdentity[], samples: NamedSample[]): string | undefined {
  let clients: string | undefined
  for (const row of ranks) {
    if (row.rank.endsWith('-s')) continue
    const value = rankSample(row, samples)?.counters.find((counter) => counter.name === 'mds_sessions.session_count')?.value
    if (value != null && (row.rank === '0' || clients === undefined || clients === '0')) clients = value
  }
  return clients
}

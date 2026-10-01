export interface PerformanceCounter { name: string; value: string | null }
export interface PerformanceSample { gid: string; counters: PerformanceCounter[]; observed_at: string }
export interface PerformancePoint { time: number; value: number }

export function performancePoints(samples: PerformanceSample[], counter: string, rate: boolean): Array<PerformancePoint | null> {
  return samples.map((sample, index) => {
    const raw = sample.counters.find((item) => item.name === counter)?.value
    const time = new Date(sample.observed_at).getTime()
    if (raw == null || !/^\d+$/.test(raw) || !Number.isFinite(time)) return null
    if (!rate) {
      const value = Number(raw)
      return Number.isSafeInteger(value) ? { time, value } : null
    }
    const previous = samples[index - 1]
    const before = previous?.counters.find((item) => item.name === counter)?.value
    if (!previous || before == null || !/^\d+$/.test(before) || previous.gid !== sample.gid) return null
    const seconds = (time - new Date(previous.observed_at).getTime()) / 1000
    const delta = BigInt(raw) - BigInt(before)
    if (!Number.isFinite(seconds) || seconds <= 0 || delta < 0n || delta > BigInt(Number.MAX_SAFE_INTEGER)) return null
    return { time, value: Number(delta) / seconds }
  })
}

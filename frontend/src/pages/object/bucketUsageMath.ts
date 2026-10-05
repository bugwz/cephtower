// The inputs are validated decimal counters from native rgw.main statistics.
export function averageBucketObjectBytes(data: { object_count: string; size_actual_bytes: string }): string {
  const count = BigInt(data.object_count)
  if (count === 0n) return '不适用（对象数为 0）'
  const hundredths = BigInt(data.size_actual_bytes) * 100n / count
  return `${hundredths / 100n}.${String(hundredths % 100n).padStart(2, '0')} B/对象`
}

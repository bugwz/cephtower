import type { ApiRecord } from '../../api/client'

// Native PERFCOUNTER_U64 | PERFCOUNTER_COUNTER; averages/histograms are not scalars.
export function daemonPerfRate(current: ApiRecord, previous: ApiRecord | undefined, elapsedMs: number): string {
  if (current.type !== 10) return '不适用（非 uint64 累计计数器）'
  if (!previous) return '无基线值'
  if (previous.name !== current.name || previous.type !== current.type || previous.units !== current.units) return '定义已变化'
  if (!Number.isSafeInteger(elapsedMs) || elapsedMs <= 0) return '需读取较新的快照'
  const unsigned = (value: unknown): value is string => typeof value === 'string' && /^(0|[1-9][0-9]*)$/.test(value)
  if (!unsigned(current.raw_value) || !unsigned(previous.raw_value)) return '值缺失或格式无效'
  const now = BigInt(current.raw_value), before = BigInt(previous.raw_value)
  const max = (1n << 64n) - 1n
  if (now > max || before > max) return '值超出 uint64 范围'
  if (now < before) return '计数器回退，可能已重置'
  // Fixed-point division retains integer precision even above Number.MAX_SAFE_INTEGER.
  const scaled = (now - before) * 1_000_000_000n / BigInt(elapsedMs)
  return `${scaled / 1_000_000n}.${(scaled % 1_000_000n).toString().padStart(6, '0')} /s`
}

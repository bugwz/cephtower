import type { ApiRecord } from '../../api/client'

// Mirrors the numeric syntax in common/strtol.cc, without converting integers to Number.
function integerValue(value: string, type: string): bigint | undefined {
  const match = type === 'size'
    ? /^([+-]?\d+)(B|[KMGTPE](?:i?B|i)?)?$/.exec(value)
    : /^\s*([+-]?\d+)([KMGTPE])?$/.exec(value)
  if (!match) return undefined
  const coefficient = BigInt(match[1])
  // Ceph parses the coefficient using strict_strtoll before applying the unit.
  if (coefficient < -(2n ** 63n) || coefficient > 2n ** 63n - 1n) return undefined
  const unit = match[2]?.[0]
  const power = unit && unit !== 'B' ? BigInt('KMGTPE'.indexOf(unit) + 1) : 0n
  return coefficient * (type === 'size' ? 1024n : 1000n) ** power
}

function integerBound(value: unknown): bigint | undefined {
  if (typeof value !== 'string' || !/^[+-]?\d+$/.test(value)) return undefined
  return BigInt(value)
}

export function configurationValueError(help: ApiRecord | null, name: unknown, value: unknown): string | undefined {
  if (!help || help.name !== name) return undefined
  const type = help.type
  if (!['int', 'uint', 'size', 'float'].includes(String(type))) return undefined
  if (typeof value !== 'string' || value === '') return '该数值类型不能写入空值'
  if (type === 'float') {
    // Other strtod spellings remain Ceph-validated rather than coerced with Number.
    if (/^\s*[+-]?(?:0x(?:[\da-f]+(?:\.[\da-f]*)?|\.[\da-f]+)(?:p[+-]?\d+)?|inf(?:inity)?|nan(?:\([\w]*\))?)$/i.test(value)) return undefined
    if (!/^\s*[+-]?(?:\d+(?:\.\d*)?|\.\d+)(?:[eE][+-]?\d+)?$/.test(value)) return '请输入浮点数，可使用小数或科学计数法'
    const number = Number(value)
    if (!Number.isFinite(number)) return '浮点值超出可表示范围'
    for (const key of ['min', 'max'] as const) {
      const raw = help[key]
      if (typeof raw !== 'string' || raw === '' || !Number.isFinite(Number(raw))) continue
      if (key === 'min' ? number < Number(raw) : number > Number(raw)) return `配置值必须${key === 'min' ? '不小于' : '不大于'} ${raw}`
    }
    return undefined
  }
  const number = integerValue(value, String(type))
  if (number === undefined) return type === 'size' ? '请输入整数容量，可使用 B、K、KiB、M、MiB 等单位' : '请输入整数，可使用 K、M、G、T、P、E 十进制单位'
  const minimum = type === 'int' ? -(2n ** 63n) : 0n
  const maximum = type === 'int' ? 2n ** 63n - 1n : 2n ** 64n - 1n
  if (number < minimum || number > maximum) return '配置值超出该整数类型的范围'
  for (const key of ['min', 'max'] as const) {
    const bound = integerBound(help[key])
    if (bound !== undefined && (key === 'min' ? number < bound : number > bound)) return `配置值必须${key === 'min' ? '不小于' : '不大于'} ${help[key]}`
  }
  return undefined
}

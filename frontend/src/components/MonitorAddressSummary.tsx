import { Typography } from 'antd'
import { message } from '../utils/appMessage'

const { Text } = Typography

interface MonitorAddressSummaryProps {
  value: unknown
  maxVisible?: number
}

export function MonitorAddressSummary({ value, maxVisible }: MonitorAddressSummaryProps) {
  const addresses = groupMonitorAddresses(parseMonitorAddresses(value))
  const visibleAddresses = typeof maxVisible === 'number' ? addresses.slice(0, Math.max(0, maxVisible)) : addresses
  const hiddenCount = addresses.length - visibleAddresses.length

  if (!addresses.length) {
    return <Text type="secondary">未配置</Text>
  }

  return (
    <div className="mon-address-summary">
      {visibleAddresses.map((address) => (
        <button key={address} type="button" className="mon-address-chip" onClick={() => copyMonitorAddress(address)}>
          {address}
        </button>
      ))}
      {hiddenCount > 0 ? (
        <span className="mon-address-more" title={addresses.join('\n')}>
          更多 {hiddenCount} 个
        </span>
      ) : null}
    </div>
  )
}

function parseMonitorAddresses(value: unknown): string[] {
  if (Array.isArray(value)) {
    return unique(value.flatMap((item) => parseMonitorAddresses(item)))
  }
  if (typeof value !== 'string') {
    return []
  }
  return unique(
    Array.from(value.matchAll(/(?:^|[\s,;\[])(?:v\d+:)?([^\s,;\[\]]+?:\d+)(?:\/\d+)?(?=$|[\s,;\]])/gi))
      .map((match) => match[1])
      .map((address) => address.trim())
  )
}

function unique(values: string[]) {
  return Array.from(new Set(values.filter(Boolean)))
}

function groupMonitorAddresses(addresses: string[]) {
  const hostOrder = new Map<string, number>()
  addresses.forEach((address) => {
    const host = monitorAddressHost(address)
    if (!hostOrder.has(host)) {
      hostOrder.set(host, hostOrder.size)
    }
  })
  return [...addresses].sort((left, right) => {
    const leftHost = hostOrder.get(monitorAddressHost(left)) ?? 0
    const rightHost = hostOrder.get(monitorAddressHost(right)) ?? 0
    return leftHost - rightHost
  })
}

function monitorAddressHost(address: string) {
  const bracketed = address.match(/^\[([^\]]+)\]:\d+$/)
  if (bracketed) {
    return bracketed[1]
  }
  return address.replace(/:\d+$/, '')
}

async function copyMonitorAddress(address: string) {
  try {
    await writeClipboardText(address)
    message.success('复制成功')
  } catch (err) {
    message.error(err instanceof Error ? err.message : '复制 MON 地址失败')
  }
}

async function writeClipboardText(value: string) {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(value)
    return
  }

  const textarea = document.createElement('textarea')
  textarea.value = value
  textarea.setAttribute('readonly', 'true')
  textarea.style.position = 'fixed'
  textarea.style.left = '-9999px'
  document.body.appendChild(textarea)
  textarea.select()
  const copied = document.execCommand('copy')
  document.body.removeChild(textarea)
  if (!copied) {
    throw new Error('复制 MON 地址失败')
  }
}

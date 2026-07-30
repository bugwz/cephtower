import { Popover, Tag, Typography } from 'antd'

const { Text } = Typography

interface MonitorAddressSummaryProps {
  value: unknown
}

export function MonitorAddressSummary({ value }: MonitorAddressSummaryProps) {
  const addresses = parseMonitorAddresses(value)

  if (!addresses.length) {
    return <Text type="secondary">未配置</Text>
  }

  const visibleAddresses = addresses.slice(0, 2)
  const hiddenCount = addresses.length - visibleAddresses.length

  return (
    <Popover
      placement="topLeft"
      overlayClassName="mon-address-popover"
      content={
        <div className="mon-address-tooltip">
          {addresses.map((address) => (
            <span key={address} className="mon-address-tooltip-item">
              {address}
            </span>
          ))}
        </div>
      }
    >
      <div className="mon-address-summary">
        <Tag color="green">{addresses.length} 个 MON</Tag>
        <div className="mon-address-chips">
          {visibleAddresses.map((address) => (
            <span key={address} className="mon-address-chip">
              {address}
            </span>
          ))}
          {hiddenCount > 0 && <span className="mon-address-more">+{hiddenCount}</span>}
        </div>
      </div>
    </Popover>
  )
}

function parseMonitorAddresses(value: unknown) {
  if (Array.isArray(value)) {
    return value.map((item) => String(item).trim()).filter(Boolean)
  }
  if (typeof value !== 'string') {
    return []
  }
  return value
    .split(/[\s,;]+/)
    .map((address) => address.trim())
    .filter(Boolean)
}

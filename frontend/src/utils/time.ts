const dateTimeFields = new Set([
  'created_at',
  'updated_at',
  'observed_at',
  'occurred_at',
  'last_login_at',
  'last_seen_at',
  'discovered_at'
])

export function formatDateTime(value: unknown, fallback = '-') {
  if (typeof value !== 'string' || !value) {
    return fallback
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  const year = date.getFullYear()
  const month = padDatePart(date.getMonth() + 1)
  const day = padDatePart(date.getDate())
  const hours = padDatePart(date.getHours())
  const minutes = padDatePart(date.getMinutes())
  const seconds = padDatePart(date.getSeconds())
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

export function renderDateTime(value: unknown) {
  return formatDateTime(value)
}

export function isDateTimeField(key: string) {
  return dateTimeFields.has(key)
}

function padDatePart(value: number) {
  return String(value).padStart(2, '0')
}

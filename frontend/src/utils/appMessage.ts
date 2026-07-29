import { message as antdMessage } from 'antd'

const maxMessageLength = 32
const defaultDuration = 3

antdMessage.config({
  duration: defaultDuration,
  maxCount: 3,
  top: 24
})

function normalizeContent(content: unknown) {
  const text = String(content ?? '').replace(/\s+/g, ' ').trim()
  const fallback = '操作已处理'
  const messageText = text || fallback

  if (messageText.length <= maxMessageLength) {
    return messageText
  }

  return `${messageText.slice(0, maxMessageLength - 3)}...`
}

export const message = {
  success(content: unknown, duration = defaultDuration) {
    return antdMessage.success(normalizeContent(content), duration)
  },
  error(content: unknown, duration = defaultDuration) {
    return antdMessage.error(normalizeContent(content), duration)
  },
  warning(content: unknown, duration = defaultDuration) {
    return antdMessage.warning(normalizeContent(content), duration)
  },
  info(content: unknown, duration = defaultDuration) {
    return antdMessage.info(normalizeContent(content), duration)
  }
}

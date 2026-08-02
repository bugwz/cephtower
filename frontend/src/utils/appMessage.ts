import { message as antdMessage, notification } from 'antd'

const maxMessageLength = 32
const defaultDuration = 2.6
const messageTop = 18

antdMessage.config({
  duration: defaultDuration,
  maxCount: 3,
  top: messageTop
})

notification.config({
  duration: defaultDuration,
  maxCount: 3,
  placement: 'top',
  top: messageTop
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

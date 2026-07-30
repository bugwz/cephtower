import { useEffect } from 'react'
import { subscribeApiErrors, type ApiErrorDetail } from '../api/client'
import { message } from '../utils/appMessage'

interface ApiErrorNotifierProps {
  onAuthenticationRequired?: (detail: ApiErrorDetail) => void
}

export function ApiErrorNotifier({ onAuthenticationRequired }: ApiErrorNotifierProps) {
  useEffect(() => {
    return subscribeApiErrors((detail) => {
      if (detail.requiresAuthentication) {
        onAuthenticationRequired?.(detail)
        message.warning('登录状态已过期，请重新登录')
        return
      }

      message.error(detail.message || '后端接口调用失败')
    })
  }, [onAuthenticationRequired])

  return null
}

import { useEffect } from 'react'
import { subscribeApiErrors } from '../api/client'
import { message } from '../utils/appMessage'

export function ApiErrorNotifier() {
  useEffect(() => {
    return subscribeApiErrors((detail) => {
      message.error(detail.message || '后端接口调用失败')
    })
  }, [])

  return null
}

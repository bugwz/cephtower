import { notification } from 'antd'
import { useEffect } from 'react'
import { subscribeApiErrors } from '../api/client'

export function ApiErrorNotifier() {
  const [api, contextHolder] = notification.useNotification()

  useEffect(() => {
    return subscribeApiErrors((detail) => {
      api.error({
        className: 'api-error-notification',
        message: '后端接口调用失败',
        description: (
          <div>
            <div>{detail.message}</div>
            {detail.path && <div className="api-error-path">接口：{detail.path}</div>}
          </div>
        ),
        duration: 6,
        placement: 'topRight'
      })
    })
  }, [api])

  return contextHolder
}

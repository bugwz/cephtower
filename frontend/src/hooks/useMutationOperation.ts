import { useState } from 'react'
import { message } from '../utils/appMessage'

const defaultSuccessMessage = '操作执行成功'

export function useMutationOperation() {
  const [loading, setLoading] = useState(false)

  async function run<T>(executor: () => Promise<T>, successMessage: string | false = defaultSuccessMessage) {
    if (loading) {
      throw new Error('已有操作正在执行')
    }
    setLoading(true)
    try {
      const result = await executor()
      if (successMessage) {
        message.success(successMessage)
      }
      return result
    } finally {
      setLoading(false)
    }
  }

  return { run, loading }
}

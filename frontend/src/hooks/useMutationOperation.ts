import { useEffect, useRef, useState } from 'react'
import { message } from '../utils/appMessage'

const defaultSuccessMessage = '操作执行成功'

export function useMutationOperation() {
  const [loading, setLoading] = useState(false)
  const running = useRef(false)
  const lifetime = useRef({ active: true, generation: 0 })
  useEffect(() => {
    lifetime.current.active = true
    return () => { lifetime.current.active = false; lifetime.current.generation += 1 }
  }, [])

  async function run<T>(executor: () => Promise<T>, successMessage: string | false = defaultSuccessMessage) {
    if (!lifetime.current.active) throw new Error('操作页面已关闭，请重新打开后再操作')
    if (running.current) {
      throw new Error('已有操作正在执行')
    }
    running.current = true
    const generation = lifetime.current.generation
    setLoading(true)
    try {
      const result = await executor()
      if (successMessage && lifetime.current.active && lifetime.current.generation === generation) {
        message.success(successMessage)
      }
      return result
    } finally {
      running.current = false
      if (lifetime.current.active) setLoading(false)
    }
  }

  return { run, loading }
}

import { useCallback, useEffect, useState } from 'react'

export function useResource<T>(loader: () => Promise<T>) {
  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setData(await loader())
    } catch (err) {
      setError(err instanceof Error ? err.message : '请求失败')
    } finally {
      setLoading(false)
    }
  }, [loader])

  useEffect(() => {
    let ignore = false

    async function run() {
      setLoading(true)
      setError('')
      try {
        const payload = await loader()
        if (!ignore) {
          setData(payload)
        }
      } catch (err) {
        if (!ignore) {
          setError(err instanceof Error ? err.message : '请求失败')
        }
      } finally {
        if (!ignore) {
          setLoading(false)
        }
      }
    }

    run()

    return () => {
      ignore = true
    }
  }, [loader])

  return { data, loading, error, refresh: load }
}


export function startAutoRefresh(refresh: () => Promise<unknown>, delay: number, clock = { setTimeout, clearTimeout }) {
  let stopped = false
  let timer: ReturnType<typeof setTimeout>
  const schedule = () => {
    if (stopped) return
    timer = clock.setTimeout(() => {
      if (stopped) return
      // The resource loader owns error presentation. Never overlap polling requests.
      void Promise.resolve().then(() => stopped ? undefined : refresh()).catch(() => {}).finally(schedule)
    }, delay)
  }
  schedule()
  return () => { stopped = true; clock.clearTimeout(timer) }
}

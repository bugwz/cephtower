import { Descriptions } from 'antd'
import { rgwRateLimitDetails } from './rgwRateLimitDetails'

export function RgwRateLimit({ value }: { value: unknown }) {
  const details = rgwRateLimitDetails(value)
  const labels = ['读操作（次/分钟）', '写操作（次/分钟）', '读取（bytes/分钟）', '写入（bytes/分钟）']
  return <Descriptions size="small" column={1} items={[
    { key: 'state', label: '限流状态', children: details.state },
    ...(details.limits?.map((limit, index) => ({ key: String(index), label: labels[index], children: limit })) ?? [])
  ]} />
}

import { Descriptions } from 'antd'
import { rgwQuotaDetails } from './rgwQuotaDetails'

export function RgwQuota({ value }: { value: unknown }) {
  const quota = rgwQuotaDetails(value)
  return <Descriptions size="small" column={1} items={[
    { key: 'state', label: '配额状态', children: quota.state },
    ...(quota.size === undefined ? [] : [
      { key: 'size', label: '容量上限（bytes）', children: quota.size },
      { key: 'objects', label: '对象数量上限', children: quota.objects }
    ])
  ]} />
}

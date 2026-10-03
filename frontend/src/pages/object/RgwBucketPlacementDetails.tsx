import { Descriptions } from 'antd'
import { rgwBucketPlacement } from './rgwBucketPlacement'

export function RgwBucketPlacementDetails({ value }: { value: unknown }) {
  const pools = rgwBucketPlacement(value)
  return <div>
    <Descriptions size="small" column={1} items={[
      { key: 'data', label: '显式数据池', children: pools.data },
      { key: 'extra', label: '显式额外数据池', children: pools.extra },
      { key: 'index', label: '显式索引池', children: pools.index }
    ]} />
    <small>仅显示 explicit_placement 原值；未显式指定不代表没有存储池，也不代表已解析放置规则的实际目标。</small>
  </div>
}

import { Tabs } from 'antd'
import type { ApiRecord } from '../../api/client'
import { RgwUserIdentityDetails, RgwUserPlacementDetails } from './RgwUserIdentityDetails'

export function RgwUserDetails({ row }: { row: ApiRecord }) {
  return <Tabs items={[
    { key: 'identity', label: '身份与归属', children: <RgwUserIdentityDetails row={row} /> },
    { key: 'placement', label: '用户放置配置', children: <RgwUserPlacementDetails row={row} /> }
  ]} />
}

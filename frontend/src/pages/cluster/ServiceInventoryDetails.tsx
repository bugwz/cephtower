import { Alert, Descriptions } from 'antd'
import type { ApiRecord } from '../../api/client'

const fields = [
  ['type', '服务类型'], ['unmanaged', '非托管'], ['networks', '绑定网段'],
  ['ports', '端口'], ['container_image_name', '容器镜像'], ['container_image_id', '镜像 ID'],
  ['service_url', '服务地址'], ['virtual_ip', '虚拟 IP'], ['ceph_created_at', 'Ceph 创建时间'],
  ['last_refresh', 'Ceph 最近刷新'], ['observed_at', '库存采集时间'], ['events', '服务事件']
] as const

export function ServiceInventoryDetails({ row }: { row: ApiRecord }) {
  return <details>
    <summary>展开服务部署信息与事件</summary>
    {row.stale !== false && <Alert type="warning" showIcon message="库存已过期或新鲜度未知，下方不是当前部署状态的确认" />}
    <Descriptions bordered size="small" column={1} items={fields.map(([key, label]) => {
      const value = row[key]
      let text: string
      if (value === null || value === undefined) text = '未返回'
      else if (typeof value === 'boolean') text = value ? '是' : '否'
      else if (Array.isArray(value)) text = value.length ? JSON.stringify(value, null, 2) : '本次未报告条目'
      else if (typeof value === 'object') text = JSON.stringify(value, null, 2)
      else text = String(value)
      return { key, label, children: <div style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{text}</div> }
    })} />
  </details>
}

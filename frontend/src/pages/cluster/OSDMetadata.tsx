import { Alert, Descriptions, Space } from 'antd'
import type { ApiRecord } from '../../api/client'
import { RecordDetail } from '../../components/RecordDetail'

export function OSDMetadata({ data }: { data: ApiRecord }) {
  const fields = [
    ['hostname', '主机'], ['ceph_version', 'Ceph 版本'], ['osd_objectstore', '对象存储后端'],
    ['os', '操作系统'], ['kernel_version', '内核版本'], ['arch', '架构'], ['cpu', 'CPU'],
    ['mem_total_kb', '主机内存（原生 KB）']
  ]
  const flags = [['bluefs', 'BlueFS'], ['bluefs_dedicated_db', '独立 DB 设备'], ['bluefs_dedicated_wal', '独立 WAL 设备']]
  const text = (value: unknown) => typeof value === 'string' && value.trim() !== '' ? value : '未返回或格式无效'
  const flag = (value: unknown) => value === '1' ? '是' : value === '0' ? '否' : '未返回或格式无效'
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Alert type="info" message="来自 ceph osd metadata 的守护进程上报元数据，不是实时主机探测；内存字段是主机总量，不是 OSD 实时用量。" />
    {Object.keys(data).length === 0 && <Alert type="warning" message="本次未返回 OSD 元数据" />}
    <Descriptions bordered size="small" column={{ xs: 1, sm: 2 }}>
      {fields.map(([key, label]) => <Descriptions.Item key={key} label={label}>{text(data[key])}</Descriptions.Item>)}
      {flags.map(([key, label]) => <Descriptions.Item key={key} label={label}>{flag(data[key])}</Descriptions.Item>)}
    </Descriptions>
    <details><summary>全部原始元数据</summary><RecordDetail record={data} /></details>
  </Space>
}

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
  const deviceGroups = [['bluestore_bdev_', 'BlueStore 数据设备'], ['bluefs_db_', 'BlueFS DB 设备'], ['bluefs_wal_', 'BlueFS WAL 设备']]
  const deviceFields = [
    ['devices', '底层设备名'], ['dev_node', '设备节点'], ['partition_path', '分区路径'], ['path', '文件路径'],
    ['size', '容量（字节原值）'], ['block_size', '块大小（字节）'], ['optimal_io_size', '最优 I/O 大小（字节）'],
    ['driver', '驱动'], ['type', '介质类型'], ['access_mode', '访问模式'], ['model', '型号'], ['serial', '序列号']
  ]
  const text = (value: unknown) => typeof value === 'string' && value.trim() !== '' ? value : '未返回或格式无效'
  const flag = (value: unknown) => value === '1' ? '是' : value === '0' ? '否' : '未返回或格式无效'
  return <Space direction="vertical" style={{ width: '100%' }}>
    <Alert type="info" message="来自 ceph osd metadata 的守护进程上报元数据，不是实时主机探测；内存字段是主机总量，不是 OSD 实时用量。" />
    {Object.keys(data).length === 0 && <Alert type="warning" message="本次未返回 OSD 元数据" />}
    <Descriptions bordered size="small" column={{ xs: 1, sm: 2 }}>
      {fields.map(([key, label]) => <Descriptions.Item key={key} label={label}>{text(data[key])}</Descriptions.Item>)}
      {flags.map(([key, label]) => <Descriptions.Item key={key} label={label}>{flag(data[key])}</Descriptions.Item>)}
    </Descriptions>
    {deviceGroups.map(([prefix, title]) => <details key={prefix}>
      <summary>{title}</summary>
      {Object.keys(data).some(key => key.startsWith(prefix)) ? <Descriptions bordered size="small" column={{ xs: 1, sm: 2 }}>
        {deviceFields.map(([key, label]) => <Descriptions.Item key={prefix + key} label={label}>{text(data[prefix + key])}</Descriptions.Item>)}
        <Descriptions.Item label="旋转设备">{flag(data[prefix + 'rotational'])}</Descriptions.Item>
        <Descriptions.Item label="支持 Discard">{flag(data[prefix + 'support_discard'])}</Descriptions.Item>
      </Descriptions> : <Alert type="info" message="未返回该角色的独立设备元数据；不能据此判断设备不存在或与其他角色共用。" />}
    </details>)}
    <details><summary>全部原始元数据</summary><RecordDetail record={data} /></details>
  </Space>
}

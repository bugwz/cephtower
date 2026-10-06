import { Alert, Button, Descriptions } from 'antd'
import { useState } from 'react'
import { type ApiRecord } from '../../api/client'
import { DraggableModal } from '../../components/DraggableModal'

export const removalBooleanFields = [
  ['started', '流程已开始'], ['draining', '正在排空'], ['stopped', '已停止'],
  ['replace', '替换 OSD'], ['replace_block', '替换 Block'], ['replace_db', '替换 DB'],
  ['replace_wal', '替换 WAL'], ['force', '强制移除'], ['zap', '清盘（Zap）']
]
export const removalTimeFields = [
  ['process_started_at', '流程开始'], ['drain_started_at', '排空开始'],
  ['drain_stopped_at', '排空停止'], ['drain_done_at', '排空完成']
]
export function removalBoolean(value: unknown): string {
  return value === true ? '是' : value === false ? '否' : '未知'
}
export function removalWeight(value: unknown): string {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0 ? String(value) : '未采集或格式无效'
}
export function removalTimestamp(value: unknown): string {
  return value === null ? '尚无时间记录' : typeof value === 'string' && value.trim() ? value : '未采集或格式无效'
}

export function OSDRemovalDetails({ clusterId, record }: { clusterId: number; record: ApiRecord }) {
  const [open, setOpen] = useState(false)
  return <>
    <Button onClick={() => setOpen(true)}>队列详情</Button>
    <DraggableModal open={open} title={`集群 ${clusterId} / OSD ${record.osd_id} 移除队列`} footer={null} onCancel={() => setOpen(false)} width={820}>
      <Alert type="info" message="展示原生队列快照，不代表当前操作已经完成。时间保留原生时区和精度；空时间仅表示尚无记录。" />
      <Descriptions bordered column={2}>
        <Descriptions.Item label="主机">{typeof record.hostname === 'string' ? record.hostname : '未知'}</Descriptions.Item>
        <Descriptions.Item label="原始权重">{removalWeight(record.original_weight)}</Descriptions.Item>
        {removalBooleanFields.map(([key, label]) => <Descriptions.Item key={key} label={label}>{removalBoolean(record[key])}</Descriptions.Item>)}
        {removalTimeFields.map(([key, label]) => <Descriptions.Item key={key} label={label}>{removalTimestamp(record[key])}</Descriptions.Item>)}
      </Descriptions>
    </DraggableModal>
  </>
}

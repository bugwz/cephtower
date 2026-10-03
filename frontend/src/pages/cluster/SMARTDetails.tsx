import { Alert, Space, Typography } from 'antd'
import { isRecord, type ApiRecord } from '../../api/client'
import { DataTable } from '../../components/DataTable'
import { RecordDetail } from '../../components/RecordDetail'

function ataRows(value: unknown): ApiRecord[] | null {
  if (!isRecord(value) || !Array.isArray(value.table) || !value.table.every(isRecord)) return null
  return value.table.map((row, index) => ({ ...row, row_key: String(index), raw_value: isRecord(row.raw) ? row.raw.value : null }))
}

function scsiRows(value: unknown): ApiRecord[] | null {
  if (!isRecord(value) || !Object.values(value).every(isRecord)) return null
  return Object.entries(value).map(([operation, counters]) => ({ ...(counters as ApiRecord), operation }))
}

export function SMARTDetails({ data }: { data: unknown }) {
  if (!isRecord(data)) return <Alert type="warning" message="原生 SMART 详情不是有效对象" />
  const ata = ataRows(data.ata_smart_attributes)
  const scsi = scsiRows(data.scsi_error_counter_log)
  return <Space direction="vertical" style={{ width: '100%', minWidth: 600 }}>
    <Typography.Text type="secondary">数值按后端精确文本展示。缺失字段不补零，属性阈值由设备报告，不自动推断剩余寿命。</Typography.Text>
    {data.ata_smart_attributes !== undefined && <>
      <Typography.Title level={5}>ATA 属性</Typography.Title>
      {ata === null ? <Alert type="warning" message="ATA 属性表格式无效，请查看原生数据" /> : <DataTable data={ata} rowKeyCandidates={['row_key']} columns={[
        { key: 'id', title: 'ID' }, { key: 'name', title: '属性名' }, { key: 'raw_value', title: '原始值' },
        { key: 'thresh', title: '阈值' }, { key: 'value', title: '当前值' }, { key: 'when_failed', title: '失败时机' }, { key: 'worst', title: '最差值' },
      ]} />}
    </>}
    {data.scsi_error_counter_log !== undefined && <>
      <Typography.Title level={5}>SCSI 错误计数</Typography.Title>
      {scsi === null ? <Alert type="warning" message="SCSI 错误计数格式无效，请查看原生数据" /> : <DataTable data={scsi} rowKeyCandidates={['operation']} columns={[
        { key: 'operation', title: '操作' }, { key: 'correction_algorithm_invocations', title: '纠错算法调用' },
        { key: 'errors_corrected_by_eccdelayed', title: '延迟 ECC 修正' }, { key: 'errors_corrected_by_eccfast', title: '快速 ECC 修正' },
        { key: 'errors_corrected_by_rereads_rewrites', title: '重读 / 重写修正' }, { key: 'gigabytes_processed', title: '处理量（GB）' },
        { key: 'total_errors_corrected', title: '已修正总数' }, { key: 'total_uncorrected_errors', title: '未修正总数' },
      ]} />}
    </>}
    {data.scsi_grown_defect_list !== undefined && <Typography.Text>SCSI 增长缺陷数：{String(data.scsi_grown_defect_list ?? '未返回')}</Typography.Text>}
    {data.nvme_smart_health_information_log !== undefined && <>
      <Typography.Title level={5}>NVMe 健康日志</Typography.Title>
      {isRecord(data.nvme_smart_health_information_log) ? <RecordDetail record={data.nvme_smart_health_information_log} /> : <Alert type="warning" message="NVMe 健康日志格式无效，请查看原生数据" />}
    </>}
    <details><summary>全部原生字段</summary><pre style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{JSON.stringify(data, null, 2)}</pre></details>
  </Space>
}

import { Alert, Button, Card, Descriptions, Space, Tag, Typography } from 'antd'
import { useCallback } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'
import { useClusterContext } from '../../state/ClusterContext'
import { TelemetryReportPanel } from './TelemetryReportPanel'
import { TelemetryControls } from './TelemetryControls'
import { TelemetryChannels } from './TelemetryChannels'
import { TelemetrySettings } from './TelemetrySettings'

export function telemetryStatusValue(value: unknown): string {
  if (value == null) return '未提供'
  if (typeof value === 'boolean') return value ? '开启' : '关闭'
  if (value === '') return '空字符串'
  return typeof value === 'string' || typeof value === 'number' ? String(value) : '未知'
}

export function TelemetryPage() {
  const { selectedClusterId } = useClusterContext()
  const loader = useCallback(async () => selectedClusterId ? request<{ status: ApiRecord, observed_at: string }>('/manager/telemetry/status', jsonInit('GET', { cluster_id: selectedClusterId })) : null, [selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const status = data?.status
  return <Page title="遥测状态" loading={loading} error={error}>
    <Card title="Ceph Telemetry" extra={<Button disabled={!selectedClusterId || loading} onClick={() => refresh()}>读取最新状态</Button>}>
      <Space direction="vertical" style={{ width: '100%' }}>
        <Alert type="info" showIcon message="状态读取执行 ceph telemetry status，不会启用遥测或发送报告。遥测启用后，Ceph 可能按配置周期向外部地址上传集群信息。" />
        {!selectedClusterId && <Alert type="info" message="请先选择集群" />}
        {selectedClusterId && <TelemetrySettings key={selectedClusterId} disabled={loading || Boolean(error)} onClose={refresh} />}
        {status && <>
          <Typography.Text>读取时间：{data.observed_at}</Typography.Text>
          {error && <Alert type="warning" message="本次读取失败，以下为上次成功读取的状态。" />}
          <Tag color={status.enabled === true ? 'warning' : 'default'}>{status.enabled === true ? '遥测已开启' : status.enabled === false ? '遥测已关闭' : '状态未知'}</Tag>
          {selectedClusterId && typeof status.enabled === 'boolean' && <TelemetryControls key={`${selectedClusterId}:${status.enabled}`} clusterId={selectedClusterId} enabled={status.enabled} disabled={loading || Boolean(error)} onComplete={refresh} />}
          <Descriptions bordered size="small" column={1}>
            {([['url', '集群报告地址'], ['device_url', '设备报告地址'], ['interval', '上传间隔（小时）'], ['last_opt_revision', '最后同意的报告修订版'], ['last_upload', '最近上传（Ceph 原始时间）'], ['channel_basic', '基本集群信息'], ['channel_ident', '身份及联系信息'], ['channel_crash', '崩溃信息'], ['channel_device', '设备健康信息'], ['channel_perf', '性能信息'], ['leaderboard', '排行榜'], ['leaderboard_description', '排行榜描述'], ['description', '描述'], ['organization', '组织'], ['contact', '联系人'], ['proxy', '代理']] as const).map(([key, label]) => <Descriptions.Item key={key} label={label}>{key === 'last_upload' && (status[key] === null || status[key] === 0) ? '尚无上传记录' : telemetryStatusValue(status[key])}</Descriptions.Item>)}
          </Descriptions>
          {selectedClusterId && <TelemetryChannels key={`${selectedClusterId}:${data.observed_at}`} clusterId={selectedClusterId} status={status} disabled={loading || Boolean(error)} onComplete={refresh} />}
        </>}
      </Space>
    </Card>
    {selectedClusterId && <TelemetryReportPanel key={selectedClusterId} clusterId={selectedClusterId} />}
  </Page>
}

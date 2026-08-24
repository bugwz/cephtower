import { ArrowLeftOutlined, DeleteOutlined, ReloadOutlined, TagOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Col, Descriptions, Form, Input, Modal, Row, Select, Space, Spin, Statistic, Tabs, Tag, Tooltip, Typography } from 'antd'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { isRecord, numberValue, textValue, type ApiRecord } from '../../api/client'
import { queryMetric, type MetricResponse } from '../../api/external'
import { getHostDeviceInfo, getHostSMART, getOptionalResource, listDaemons, listHostDevices, listResource, mutateResource, refreshResource } from '../../api/resource'
import type { ResourceDTO } from '../../api/types'
import { DataTable } from '../../components/DataTable'
import { DraggableModal } from '../../components/DraggableModal'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'
import { useFeatureRequirements } from '../../hooks/useFeatureRequirements'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useClusterContext } from '../../state/ClusterContext'
import { message } from '../../utils/appMessage'
import { formatDateTime } from '../../utils/time'
import { formatBytes, hostName, normalizeHostRow, serviceInstanceRows } from './HostPage'

const { Text } = Typography
const twoColumnDescriptions = { xs: 1, sm: 2, md: 2, lg: 2, xl: 2, xxl: 2 }

interface HostLabelFormValues {
  label?: string
  action?: 'add' | 'rm'
}

export function HostDetailPage() {
  const navigate = useNavigate()
  const { name = '' } = useParams()
  const { selectedClusterId } = useClusterContext()
  const decodedName = name
  const loader = useCallback(async () => {
    if (!selectedClusterId || !decodedName) {
      return { host: null, daemons: [], devices: [], deviceInfo: [], smart: {} }
    }
    const [hostPayload, daemons, devices, deviceInfo, smart] = await Promise.all([
      getOptionalResource('/host', selectedClusterId, { host: decodedName }),
      listDaemons(),
      listHostDevices(decodedName).then(async (items) => {
        if (items.length > 0) {
          return items
        }
        const payload = await listResource('/devices', selectedClusterId)
        return payload.items.filter((device) => textValue(device.hostname ?? device.host, '') === decodedName)
      }),
      getHostDeviceInfo(decodedName, selectedClusterId).catch(() => []),
      getHostSMART(decodedName, selectedClusterId).catch(() => ({}))
    ])
    const hostRecord = hostPayload ? resourceToRecord(hostPayload.item) : { hostname: decodedName }
    const host = normalizeHostRow(hostRecord, daemons, devices)
    const name = textValue(host.hostname ?? decodedName, '')
    return {
      host,
      daemons: daemons
        .filter((daemon) => textValue(daemon.hostname ?? daemon.host, '') === name)
        .map(normalizeDaemonRow),
      devices,
      deviceInfo,
      smart
    }
  }, [decodedName, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const host = data?.host
  const [labelForm] = Form.useForm<HostLabelFormValues>()
  const [labelModalOpen, setLabelModalOpen] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [pendingAction, setPendingAction] = useState('')
  const [refreshing, setRefreshing] = useState(false)
  const operationMutation = useMutationOperation()

  async function refreshHostDetail() {
    if (!selectedClusterId || !decodedName || refreshing) {
      return
    }
    setRefreshing(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kinds: ['host', 'daemon', 'device'] }), '刷新成功')
      await refresh()
    } finally {
      setRefreshing(false)
    }
  }

  function openLabelModal() {
    if (!host) {
      return
    }
    labelForm.resetFields()
    labelForm.setFieldsValue({
      action: 'add'
    })
    setLabelModalOpen(true)
  }

  async function submitHostLabel(values: HostLabelFormValues) {
    if (!selectedClusterId || !host || submitting) {
      return
    }
    const name = hostName(host)
    const label = values.label?.trim()
    if (!name) {
      message.error('无法识别主机名')
      return
    }
    if (!label) {
      message.error('请输入标签名称')
      return
    }
    setSubmitting(true)
    try {
      await operationMutation.run(() => mutateResource('/host', 'PATCH', {
        cluster_id: selectedClusterId,
        host: name,
        labels_add: values.action === 'rm' ? [] : [label],
        labels_remove: values.action === 'rm' ? [label] : []
      }, { ifMatch: Number(host.resource_version ?? 0) }), false)
      setLabelModalOpen(false)
      message.success('主机标签更新成功')
      void refresh({ showLoading: false })
    } finally {
      setSubmitting(false)
    }
  }

  async function runHostAction(action: string) {
    if (!selectedClusterId || !host || pendingAction) {
      return
    }
    const name = hostName(host)
    if (!name) {
      message.error('无法识别主机名')
      return
    }
    const pendingKey = `${name}:${action}`
    setPendingAction(pendingKey)
    try {
      await operationMutation.run(() => mutateResource('/host/action', 'POST', {
        cluster_id: selectedClusterId,
        host: name,
        action
      }), `主机 ${action} 执行成功`)
      await refresh()
    } finally {
      setPendingAction('')
    }
  }

  async function deleteHost() {
    if (!selectedClusterId || !host) {
      message.error('请先选择集群')
      return
    }
    const name = hostName(host)
    if (!name) {
      message.error('无法识别主机名')
      return
    }
    const generation = Number(host.resource_version ?? 0)
    Modal.confirm({
      title: `删除主机 ${name}`,
      content: '该操作为高风险操作，确认后将直接执行删除操作。',
      okText: '提交删除',
      okType: 'danger',
      cancelText: '取消',
      async onOk() {
        await operationMutation.run(() => mutateResource('/host', 'DELETE', {
          cluster_id: selectedClusterId,
          host: name
        }, { ifMatch: generation }), false)
        window.setTimeout(() => {
          message.success('主机删除执行成功')
          navigate('/cluster/host')
        })
      }
    })
  }

  return (
    <Page title="主机详情" loading={loading} error={error}>
      <Space direction="vertical" size={16} className="page-stack">
        <Card
          className="page-surface-card"
          title="基础信息"
          extra={
            <Space className="host-detail-actions">
              <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/cluster/host')}>返回</Button>
              <Button icon={<ReloadOutlined />} loading={refreshing || loading} onClick={refreshHostDetail}>刷新</Button>
              <Button danger icon={<DeleteOutlined />} disabled={!host} onClick={deleteHost}>删除</Button>
            </Space>
          }
        >
          {host ? (
            <Descriptions className="host-detail-descriptions" size="small" column={twoColumnDescriptions} bordered>
              <Descriptions.Item label="主机名">{textValue(host.hostname, decodedName)}</Descriptions.Item>
              <Descriptions.Item label="IP 地址">{textValue(host.address_display ?? host.address ?? host.addr)}</Descriptions.Item>
              <Descriptions.Item label="平台">{textValue(host.platform_display)}</Descriptions.Item>
              <Descriptions.Item label="架构">{textValue(host.architecture_display)}</Descriptions.Item>
              <Descriptions.Item label="系统">{textValue(host.system_display)}</Descriptions.Item>
              <Descriptions.Item label="内核">{textValue(host.kernel_display)}</Descriptions.Item>
              <Descriptions.Item label="CPU">{textValue(host.cpu_display)}</Descriptions.Item>
              <Descriptions.Item label="硬盘">{textValue(host.disk_count_display)} 块 / {textValue(host.storage_display)}</Descriptions.Item>
              <Descriptions.Item label="内存">{textValue(host.memory_display)}</Descriptions.Item>
              <Descriptions.Item label="标签" span={2}>{renderHostLabels(host.labels)}</Descriptions.Item>
              <Descriptions.Item label="服务" span={2}>{renderServiceInstances(serviceRows(host, data?.daemons ?? []))}</Descriptions.Item>
              <Descriptions.Item label="创建时间">{formatDateTime(host.created_at)}</Descriptions.Item>
              <Descriptions.Item label="更新时间">{formatDateTime(host.updated_at)}</Descriptions.Item>
            </Descriptions>
          ) : (
            <Text type="secondary">暂无主机详情</Text>
          )}
        </Card>

        <Card className="page-surface-card" title="主机操作">
          <Space wrap>
            <Button icon={<TagOutlined />} disabled={!host} onClick={openLabelModal}>标签</Button>
            <Button loading={pendingAction.endsWith(':maintenance_enter')} disabled={Boolean(pendingAction)} onClick={() => runHostAction('maintenance_enter')}>维护</Button>
            <Button loading={pendingAction.endsWith(':maintenance_exit')} disabled={Boolean(pendingAction)} onClick={() => runHostAction('maintenance_exit')}>退出维护</Button>
            <Button loading={pendingAction.endsWith(':drain')} disabled={Boolean(pendingAction)} onClick={() => runHostAction('drain')}>Drain</Button>
            <Button loading={pendingAction.endsWith(':stop_drain')} disabled={Boolean(pendingAction)} onClick={() => runHostAction('stop_drain')}>停止 Drain</Button>
            <Button loading={pendingAction.endsWith(':rescan')} disabled={Boolean(pendingAction)} onClick={() => runHostAction('rescan')}>Rescan</Button>
          </Space>
        </Card>

        <HostDetailTabs
          hostname={host ? hostName(host) : decodedName}
          address={host ? textValue(host.address_display ?? host.address ?? host.addr, '') : ''}
          clusterId={selectedClusterId}
          deviceInfo={data?.deviceInfo ?? []}
          devices={data?.devices ?? []}
          daemons={data?.daemons ?? []}
          smart={data?.smart ?? {}}
        />
      </Space>

      <DraggableModal
        title="标签管理"
        open={labelModalOpen}
        onCancel={() => setLabelModalOpen(false)}
        onOk={() => labelForm.submit()}
        okText="提交"
        confirmLoading={submitting}
        destroyOnClose
      >
        <Form form={labelForm} layout="vertical" onFinish={submitHostLabel}>
          <Form.Item name="label" label="标签名称" rules={[{ required: true, message: '请输入标签名称' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="action" label="操作" rules={[{ required: true, message: '请选择标签操作' }]}>
            <Select
              options={[
                { label: '添加', value: 'add' },
                { label: '移除', value: 'rm' }
              ]}
            />
          </Form.Item>
        </Form>
      </DraggableModal>
    </Page>
  )
}

function HostDetailTabs({
  hostname,
  address,
  clusterId,
  deviceInfo,
  devices,
  daemons,
  smart
}: {
  hostname: string
  address: string
  clusterId?: number
  deviceInfo: ApiRecord[]
  devices: ApiRecord[]
  daemons: ApiRecord[]
  smart: ApiRecord
}) {
  const deviceInfoRows = useMemo(() => deviceInfo.map(normalizeCephDeviceRow), [deviceInfo])
  const physicalDiskRows = useMemo(
    () => devices.map((device) => normalizeInventoryDeviceRow(device, deviceInfoRows)),
    [deviceInfoRows, devices]
  )
  const healthRows = useMemo(() => normalizeSMARTData(smart), [smart])

  return (
    <Card className="page-surface-card host-detail-tabs-card" styles={{ body: { paddingTop: 0 } }}>
      <Tabs
        className="host-detail-tabs"
        defaultActiveKey="devices"
        destroyInactiveTabPane
        items={[
          {
            key: 'devices',
            label: '设备信息',
            children: <HostDeviceInfoTable devices={deviceInfoRows} />
          },
          {
            key: 'physical-disks',
            label: '物理硬盘',
            children: <HostPhysicalDiskTable devices={physicalDiskRows} />
          },
          {
            key: 'daemons',
            label: '守护进程',
            children: <HostDaemonTable daemons={daemons} />
          },
          {
            key: 'performance',
            label: '性能详细信息',
            children: <HostPerformancePanel hostname={hostname} address={address} clusterId={clusterId} />
          },
          {
            key: 'health',
            label: '设备健康状态',
            children: <HostDeviceHealthPanel devices={healthRows} />
          }
        ]}
      />
    </Card>
  )
}

function HostDeviceInfoTable({ devices }: { devices: ApiRecord[] }) {
  return (
    <DataTable
      data={devices}
      rowKeyCandidates={['device_id', 'path', 'natural_key']}
      columns={[
        { key: 'device_id_display', title: '设备 ID' },
        { key: 'health_display', title: '健康状态', render: (value) => renderDeviceHealth(value) },
        { key: 'life_expectancy_display', title: '预计寿命' },
        { key: 'name_display', title: '设备名称' },
        { key: 'daemons_display', title: '守护进程' }
      ]}
    />
  )
}

function HostPhysicalDiskTable({ devices }: { devices: ApiRecord[] }) {
  return (
    <DataTable
      data={devices}
      rowKeyCandidates={['device_id', 'path', 'natural_key']}
      columns={[
        { key: 'path_display', title: '设备路径' },
        { key: 'type_display', title: '类型', render: (value) => <Tag>{textValue(value)}</Tag> },
        { key: 'availability_display', title: '可用性', render: (_, row) => renderDeviceAvailability(row) },
        { key: 'vendor_display', title: '供应商' },
        { key: 'model_display', title: '型号' },
        { key: 'size_display', title: '容量' },
        { key: 'osd_display', title: 'OSD' }
      ]}
    />
  )
}

function HostDaemonTable({ daemons }: { daemons: ApiRecord[] }) {
  return (
    <DataTable
      data={daemons}
      rowKeyCandidates={['name', 'daemon_name', 'daemon_display']}
      columns={[
        { key: 'daemon_display', title: '名称' },
        { key: 'type_display', title: '类型' },
        { key: 'status_display', title: '状态', render: (value) => renderDaemonStatus(value) },
        { key: 'last_refresh_display', title: '最近刷新' },
        { key: 'version_display', title: '版本' },
        { key: 'cpu_usage_display', title: 'CPU 使用率' },
        { key: 'memory_usage_display', title: '内存使用量' },
        { key: 'image_display', title: '镜像', render: (value, row) => renderImageInfo(value, row) }
      ]}
    />
  )
}

const hostPerformanceMetrics = [
  {
    key: 'cpu',
    title: 'CPU 使用率',
    metricId: 'host_cpu_usage',
    format: (value: number) => `${value.toFixed(1)}%`
  },
  {
    key: 'memory',
    title: '内存使用率',
    metricId: 'host_memory_usage',
    format: (value: number) => `${value.toFixed(1)}%`
  },
  {
    key: 'disk-read',
    title: '磁盘读取',
    metricId: 'host_disk_read_bytes',
    format: formatRate
  },
  {
    key: 'disk-write',
    title: '磁盘写入',
    metricId: 'host_disk_write_bytes',
    format: formatRate
  },
  {
    key: 'network-receive',
    title: '网络接收',
    metricId: 'host_network_receive',
    format: formatRate
  },
  {
    key: 'network-transmit',
    title: '网络发送',
    metricId: 'host_network_transmit',
    format: formatRate
  }
]

function HostPerformancePanel({ hostname, address, clusterId }: { hostname: string; address: string; clusterId?: number }) {
  const featureStatus = useFeatureRequirements(clusterId, { requiredEndpoints: ['prometheus'] })
  const [loading, setLoading] = useState(false)
  const [values, setValues] = useState<Record<string, number | undefined>>({})
  const [queryError, setQueryError] = useState('')

  useEffect(() => {
    if (!clusterId || featureStatus.loading || featureStatus.blocked || featureStatus.error) {
      return
    }
    let cancelled = false
    setLoading(true)
    setQueryError('')
    Promise.allSettled(hostPerformanceMetrics.map((metric) => queryMetric(
      clusterId,
      { metricId: metric.metricId },
      { suppressErrorNotification: true }
    )))
      .then((results) => {
        if (cancelled) {
          return
        }
        const nextValues: Record<string, number | undefined> = {}
        results.forEach((result, index) => {
          if (result.status === 'fulfilled') {
            nextValues[hostPerformanceMetrics[index].key] = metricValueForHost(result.value, hostname, address)
          }
        })
        setValues(nextValues)
        if (results.every((result) => result.status === 'rejected')) {
          setQueryError('主机性能指标查询失败')
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false)
        }
      })
    return () => {
      cancelled = true
    }
  }, [address, clusterId, featureStatus.blocked, featureStatus.error, featureStatus.loading, hostname])

  if (featureStatus.loading) {
    return <div className="host-tab-loading"><Spin /></div>
  }
  if (featureStatus.error || featureStatus.reasons.length) {
    return <Alert type="info" showIcon message="暂无主机性能数据" description={featureStatus.error || featureStatus.reasons.join('；')} />
  }
  if (queryError) {
    return <Alert type="warning" showIcon message={queryError} />
  }

  return (
    <Spin spinning={loading}>
      <Row gutter={[16, 16]} className="host-performance-grid">
        {hostPerformanceMetrics.map((metric) => {
          const value = values[metric.key]
          return (
            <Col key={metric.key} xs={24} sm={12} xl={8}>
              <Card size="small" className="host-performance-card">
                <Statistic title={metric.title} value={value === undefined ? '—' : metric.format(value)} />
              </Card>
            </Col>
          )
        })}
      </Row>
    </Spin>
  )
}

function HostDeviceHealthPanel({ devices }: { devices: ApiRecord[] }) {
  if (!devices.length) {
    return <Alert type="info" showIcon message="暂无可用的 SMART 数据" />
  }
  return (
    <DataTable
      data={devices}
      rowKeyCandidates={['device_id', 'path', 'natural_key']}
      columns={[
        { key: 'name_display', title: '设备' },
        { key: 'serial_display', title: '序列号' },
        { key: 'health_display', title: 'SMART 状态', render: (value) => renderDeviceHealth(value) },
        { key: 'temperature_display', title: '温度' },
        { key: 'power_on_hours_display', title: '通电时间' },
        { key: 'wear_level_display', title: '磨损程度' }
      ]}
    />
  )
}

function normalizeCephDeviceRow(row: ApiRecord): ApiRecord {
  const locations = Array.isArray(row.location) ? row.location.filter(isRecord) : []
  const deviceNames = locations
    .map((location) => textValue(location.dev ?? location.path ?? location.name, ''))
    .filter(Boolean)
  const daemons = stringArray(row.daemons)
  return {
    ...row,
    device_id: textValue(row.devid ?? row.device_id ?? row.id, ''),
    device_id_display: textValue(row.devid ?? row.device_id ?? row.id, ''),
    health_display: textValue(row.state ?? row.health, 'unknown'),
    life_expectancy_display: formatLifeExpectancy(row),
    name_display: deviceNames.length ? deviceNames.join(', ') : '-',
    daemons_display: daemons.length ? daemons : []
  }
}

function normalizeInventoryDeviceRow(row: ApiRecord, deviceInfo: ApiRecord[]): ApiRecord {
  const sysAPI = isRecord(row.sys_api) ? row.sys_api : {}
  const lsmData = isRecord(row.lsm_data) ? row.lsm_data : {}
  const lvs = Array.isArray(row.lvs) ? row.lvs.filter(isRecord) : []
  const osdIDs = stringArray(row.osd_ids)
  lvs.forEach((lv) => {
    const osdID = textValue(lv.osd_id, '')
    if (osdID && !osdIDs.includes(osdID)) {
      osdIDs.push(osdID)
    }
  })
  const path = textValue(row.path ?? sysAPI.path ?? row.name, '')
  const deviceName = path.split('/').filter(Boolean).pop() ?? path
  const linkedDevice = deviceInfo.find((device) => textValue(device.name_display, '').split(', ').includes(deviceName))
  stringArray(linkedDevice?.daemons_display).forEach((daemon) => {
    if (daemon.startsWith('osd.') && !osdIDs.includes(daemon)) {
      osdIDs.push(daemon)
    }
  })
  const type = textValue(row.device_type ?? row.human_readable_type ?? row.type, '')
    || (row.rotational === true || sysAPI.rotational === '1' ? 'HDD' : 'SSD')
  const size = numberValue(row.size_bytes ?? row.size ?? sysAPI.size)
  return {
    ...row,
    device_id: textValue(row.device_id ?? row.devid ?? row.id, ''),
    path_display: path,
    name_display: path,
    type_display: type.toUpperCase(),
    availability_display: row.available === true ? 'available' : 'unavailable',
    rejected_reasons_display: stringArray(row.rejected_reasons),
    vendor_display: textValue(row.vendor ?? sysAPI.vendor, ''),
    model_display: textValue(row.model ?? sysAPI.model, ''),
    serial_display: textValue(row.serial ?? lsmData.serialNum, ''),
    size_display: size ? formatBytes(size) : '-',
    osd_display: osdIDs.map((id) => id.startsWith('osd.') ? id : `osd.${id}`),
    health_display: textValue(lsmData.health, '')
  }
}

function normalizeSMARTData(value: ApiRecord): ApiRecord[] {
  return Object.entries(value).flatMap(([deviceID, raw]) => {
    if (!isRecord(raw) || raw.error) {
      return []
    }
    const smartStatus = isRecord(raw.smart_status) ? raw.smart_status : {}
    const device = isRecord(raw.device) ? raw.device : {}
    const temperature = isRecord(raw.temperature) ? raw.temperature : {}
    const powerOnTime = isRecord(raw.power_on_time) ? raw.power_on_time : {}
    const nvmeHealth = isRecord(raw.nvme_smart_health_information_log) ? raw.nvme_smart_health_information_log : {}
    const passed = smartStatus.passed
    return [{
      ...raw,
      device_id: deviceID,
      name_display: textValue(device.name ?? raw.dev ?? deviceID, deviceID),
      serial_display: textValue(raw.serial_number ?? raw.serial ?? deviceID, ''),
      health_display: typeof passed === 'boolean' ? (passed ? 'good' : 'bad') : textValue(raw.health ?? raw.state, 'unknown'),
      temperature_display: formatTemperature(numberValue(temperature.current ?? raw.temperature ?? nvmeHealth.temperature)),
      power_on_hours_display: formatHours(numberValue(powerOnTime.hours ?? raw.power_on_hours)),
      wear_level_display: formatWear(numberValue(raw.percentage_used ?? nvmeHealth.percentage_used ?? raw.ssd_life_left))
    }]
  })
}

function renderDeviceHealth(value: unknown) {
  const status = textValue(value, 'unknown').toLowerCase()
  const mapping: Record<string, { label: string; color: string }> = {
    good: { label: '良好', color: 'success' },
    passed: { label: '良好', color: 'success' },
    warning: { label: '警告', color: 'warning' },
    bad: { label: '异常', color: 'error' },
    failed: { label: '异常', color: 'error' },
    stale: { label: '数据过期', color: 'processing' },
    unknown: { label: '未知', color: 'default' }
  }
  const item = mapping[status] ?? { label: textValue(value), color: 'default' }
  return <Tag color={item.color}>{item.label}</Tag>
}

function renderDeviceAvailability(row: ApiRecord) {
  if (row.availability_display === 'available') {
    return <Tag color="success">可用</Tag>
  }
  const reasons = stringArray(row.rejected_reasons_display)
  return <Tooltip title={reasons.join('；') || undefined}><Tag>不可用</Tag></Tooltip>
}

function formatLifeExpectancy(row: ApiRecord) {
  const min = textValue(row.life_expectancy_min, '')
  const max = textValue(row.life_expectancy_max, '')
  const values = [min, max].filter((value) => value && value !== '0.000000')
  if (!values.length) {
    return 'n/a'
  }
  return values.map((value) => formatDateTime(value)).join(' ～ ')
}

function stringArray(value: unknown) {
  if (!Array.isArray(value)) {
    return []
  }
  return value.map((item) => textValue(item, '')).filter(Boolean)
}

function formatTemperature(value?: number) {
  return value === undefined ? '-' : `${value} °C`
}

function formatHours(value?: number) {
  return value === undefined ? '-' : `${value.toLocaleString()} 小时`
}

function formatWear(value?: number) {
  return value === undefined ? '-' : `${value}%`
}

function metricValueForHost(payload: MetricResponse, hostname: string, address: string) {
  const matching = payload.series.filter((series) => {
    const metric = isRecord(series.metric) ? series.metric : {}
    return Object.values(metric).some((value) => {
      const label = textValue(value, '')
      return Boolean(hostname && label.includes(hostname)) || Boolean(address && label.includes(address))
    })
  })
  const series = matching[0] ?? (payload.series.length === 1 ? payload.series[0] : undefined)
  if (!series) {
    return undefined
  }
  const values = Array.isArray(series.values) ? series.values : undefined
  const latest = values?.[values.length - 1] ?? (Array.isArray(series.value) ? series.value : undefined)
  return Array.isArray(latest) ? numberValue(latest[1] ?? latest[0]) : undefined
}

function formatRate(value: number) {
  if (value <= 0) {
    return '0 B/s'
  }
  return `${formatBytes(value)}/s`
}

function resourceToRecord(item: ResourceDTO): ApiRecord {
  const data = isRecord(item.data) ? item.data : {}
  return {
    ...data,
    kind: item.kind,
    natural_key: item.natural_key,
    name: item.name ?? data.name,
    status: item.status ?? data.status,
    resource_version: item.resource_version,
    source: item.source,
    observed_at: item.observed_at,
    created_at: item.created_at,
    updated_at: item.updated_at,
    stale: item.stale
  }
}

function serviceRows(host: ApiRecord, daemons: ApiRecord[]) {
  const rows = serviceInstanceRows(host.service_instances)
  if (rows.length) {
    return rows
  }
  const counts = new Map<string, number>()
  daemons.forEach((daemon) => {
    const type = textValue(daemon.daemon_type ?? daemon.type ?? daemon.service_type, '')
    if (type) {
      counts.set(type, (counts.get(type) ?? 0) + 1)
    }
  })
  return Array.from(counts.entries()).map(([type, count]) => ({ type, count }))
}

function renderHostLabels(value: unknown) {
  if (!Array.isArray(value)) {
    return textValue(value)
  }

  const labels = value.map((item) => textValue(item, '')).filter(Boolean)
  if (!labels.length) {
    return '-'
  }

  return (
    <Space wrap size={[6, 6]}>
      {labels.map((label) => <Tag key={label} color="default">{label}</Tag>)}
    </Space>
  )
}

function renderServiceInstances(rows: ApiRecord[]) {
  if (!rows.length) {
    return '-'
  }

  const services = rows
    .map((row) => {
      const type = textValue(row.type, '')
      const count = textValue(row.count, '')
      return type && count ? `${type}(${count})` : ''
    })
    .filter(Boolean)
  if (!services.length) {
    return '-'
  }

  return (
    <Space wrap size={[6, 6]}>
      {services.map((service) => <Tag key={service} color="default">{service}</Tag>)}
    </Space>
  )
}

function normalizeDaemonRow(row: ApiRecord): ApiRecord {
  const image = textValue(row.container_image ?? row.container_image_name, '')
  const cpuUsage = percentageValue(row.cpu_percentage ?? row.cpu_usage)
  const memoryUsage = numberValue(row.memory_usage ?? row.memory_bytes ?? row.memory)
  return {
    ...row,
    daemon_display: textValue(row.name ?? row.daemon_name, ''),
    type_display: textValue(row.type ?? row.daemon_type ?? row.service_type, ''),
    status_display: textValue(row.status ?? row.status_desc, ''),
    last_refresh_display: formatDateTime(row.last_refresh),
    version_display: textValue(row.version, ''),
    cpu_usage_display: cpuUsage === undefined ? '-' : `${cpuUsage.toFixed(1)}%`,
    memory_usage_display: memoryUsage === undefined ? '-' : formatBytes(memoryUsage),
    image_display: shortImageName(image),
    image_full: image
  }
}

function percentageValue(value: unknown) {
  if (typeof value === 'string') {
    return numberValue(value.trim().replace(/%$/, ''))
  }
  return numberValue(value)
}

function shortImageName(image: string) {
  if (!image) {
    return ''
  }

  const withoutRegistry = image.split('/').pop() ?? image
  const [nameAndTag, digest] = withoutRegistry.split('@')
  if (digest) {
    return `${nameAndTag}@${shortDigest(digest)}`
  }

  return nameAndTag
}

function shortDigest(digest: string) {
  const [algorithm, value] = digest.split(':')
  if (!algorithm || !value) {
    return digest.slice(0, 16)
  }

  return `${algorithm}:${value.slice(0, 12)}`
}

function renderDaemonStatus(value: unknown) {
  const status = textValue(value)
  if (status === '—') {
    return status
  }

  return <Tag color={daemonStatusColor(status)}>{daemonStatusText(status)}</Tag>
}

function daemonStatusColor(status: string) {
  const normalized = status.toLowerCase()
  if (normalized.includes('running') || normalized.includes('ok')) {
    return 'success'
  }
  if (normalized.includes('error') || normalized.includes('failed') || normalized.includes('stopped')) {
    return 'error'
  }
  if (normalized.includes('starting') || normalized.includes('pending')) {
    return 'processing'
  }
  return 'default'
}

function daemonStatusText(status: string) {
  const normalized = status.toLowerCase()
  if (normalized.includes('running') || normalized.includes('ok')) {
    return '运行中'
  }
  if (normalized.includes('starting') || normalized.includes('pending')) {
    return '启动中'
  }
  if (normalized.includes('stopping')) {
    return '停止中'
  }
  if (normalized.includes('stopped')) {
    return '已停止'
  }
  if (normalized.includes('error') || normalized.includes('failed')) {
    return '异常'
  }
  return status
}

function renderImageInfo(value: unknown, row: ApiRecord) {
  const shortValue = textValue(value)
  const fullValue = textValue(row.image_full, '')
  if (shortValue === '—' || !fullValue) {
    return shortValue
  }

  return (
    <Tooltip title={fullValue}>
      <span
        role="button"
        tabIndex={0}
        style={{ cursor: 'pointer' }}
        onClick={() => copyImage(fullValue)}
        onKeyDown={(event) => {
          if (event.key === 'Enter' || event.key === ' ') {
            event.preventDefault()
            copyImage(fullValue)
          }
        }}
      >
        {shortValue}
      </span>
    </Tooltip>
  )
}

async function copyImage(value: string) {
  if (await writeClipboardText(value)) {
    message.success('镜像已复制')
    return
  }

  message.error('镜像复制失败')
}

async function writeClipboardText(value: string) {
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(value)
      return true
    }
  } catch {
    // Fall through to the selection-based copy path.
  }

  const textarea = document.createElement('textarea')
  textarea.value = value
  textarea.setAttribute('readonly', 'true')
  textarea.style.position = 'fixed'
  textarea.style.left = '-9999px'
  textarea.style.top = '-9999px'
  document.body.appendChild(textarea)
  textarea.select()
  textarea.setSelectionRange(0, textarea.value.length)
  try {
    return document.execCommand('copy')
  } finally {
    document.body.removeChild(textarea)
  }
}

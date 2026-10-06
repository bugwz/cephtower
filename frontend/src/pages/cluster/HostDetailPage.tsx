import { ArrowLeftOutlined, DeleteOutlined, ReloadOutlined, TagOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Col, Descriptions, Form, Input, Modal, Row, Select, Space, Spin, Statistic, Tabs, Tag, Tooltip, Typography } from 'antd'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { isRecord, numberValue, textValue, type ApiRecord } from '../../api/client'
import { queryMetric, type MetricResponse } from '../../api/external'
import { getHostDeviceInfo, getHostSMART, getOptionalResource, listAllResources, mutateResource, refreshResource } from '../../api/resource'
import type { ResourceDTO } from '../../api/types'
import { DataTable } from '../../components/DataTable'
import { HostHardware } from './HostHardware'
import { HostNativeSummary } from './HostNativeSummary'
import { hostInitialLocation } from './hostLocation'
import { HostAddressEditor } from './HostAddressEditor'
import { SMARTDetails } from './SMARTDetails'
import { DaemonPerf } from './DaemonPerf'
import { DaemonRuntimeDetails } from './ServiceDaemons'
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

export function hostDeleteVersion(host: ApiRecord | null | undefined): string | null {
  if (!host || host.stale !== false) return null
  const raw = host.resource_version
  const version = typeof raw === 'string' ? raw : typeof raw === 'number' && Number.isSafeInteger(raw) ? String(raw) : ''
  return /^[1-9][0-9]*$/.test(version) && BigInt(version) <= 18446744073709551615n ? version : null
}

export function HostDetailPage() {
  const { name = '' } = useParams()
  const { selectedClusterId } = useClusterContext()
  return <HostDetailContent key={JSON.stringify([selectedClusterId, name])} name={name} selectedClusterId={selectedClusterId} />
}

function HostDetailContent({ name, selectedClusterId }: { name: string; selectedClusterId?: number }) {
  const navigate = useNavigate()
  const active = useRef(true)
  const deleteConfirmation = useRef<{ destroy: () => void } | null>(null)
  useEffect(() => {
    active.current = true
    return () => {
      active.current = false
      deleteConfirmation.current?.destroy()
    }
  }, [])
  const decodedName = name
  const loader = useCallback(async () => {
    if (!selectedClusterId || !decodedName) {
      return { host: null, daemons: [], devices: [], deviceInfo: [], smart: {}, deviceInfoError: '', smartError: '', inventoryWarnings: [] as string[] }
    }
    const [hostPayload, daemonInventory, deviceInventory, deviceInfo, smart] = await Promise.all([
      getOptionalResource('/host', selectedClusterId, { host: decodedName }),
      listAllResources('/daemons', selectedClusterId),
      listAllResources('/devices', selectedClusterId),
      getHostDeviceInfo(decodedName, selectedClusterId).then((value) => ({ value, error: '' })).catch((err) => ({ value: [] as ApiRecord[], error: err instanceof Error ? err.message : '设备信息读取失败' })),
      getHostSMART(decodedName, selectedClusterId).then((value) => ({ value, error: '' })).catch((err) => ({ value: {} as ApiRecord, error: err instanceof Error ? err.message : 'SMART 信息读取失败' }))
    ])
    const daemons = daemonInventory.items
    const devices = deviceInventory.items.filter(device => textValue(device.hostname ?? device.host, '') === decodedName)
    const inventoryWarnings: string[] = []
    for (const [label, inventory] of [['守护进程', daemonInventory], ['设备', deviceInventory]] as const) {
      if (inventory.stale !== false) inventoryWarnings.push(`${label}库存已过期或新鲜度未知：${inventory.staleReason || '请刷新后核对，当前列表不代表实时状态'}`)
    }
    const host = hostPayload ? normalizeHostRow(resourceToRecord(hostPayload.item), daemons, devices) : null
    const name = textValue(host?.hostname ?? decodedName, '')
    return {
      host,
      inventoryWarnings,
      daemons: daemons
        .filter((daemon) => textValue(daemon.hostname ?? daemon.host, '') === name)
        .map(normalizeDaemonRow),
      devices,
      deviceInfo: deviceInfo.value,
      deviceInfoError: deviceInfo.error,
      smart: smart.value,
      smartError: smart.error
    }
  }, [decodedName, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const host = data?.host
  const [labelForm] = Form.useForm<HostLabelFormValues>()
  const [labelModalOpen, setLabelModalOpen] = useState(false)
  const [labelVersion, setLabelVersion] = useState<string | null>(null)
  const labelRunning = useRef(false)
  const [submitting, setSubmitting] = useState(false)
  const [pendingAction, setPendingAction] = useState('')
  const actionRunning = useRef(false)
  const [refreshing, setRefreshing] = useState(false)
  const operationMutation = useMutationOperation()

  async function refreshHostDetail() {
    if (!selectedClusterId || !decodedName || refreshing) {
      return
    }
    setRefreshing(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kinds: ['host', 'daemon', 'device'] }), false)
      if (!active.current) return
      message.success('刷新成功')
      await refresh()
    } finally {
      setRefreshing(false)
    }
  }

  function openLabelModal() {
    if (!host || !active.current || loading || error || actionRunning.current || labelRunning.current) {
      return
    }
    const version = hostDeleteVersion(host)
    if (!version) { message.error('主机库存过期或版本无效，请刷新后编辑标签'); return }
    setLabelVersion(version)
    labelForm.resetFields()
    labelForm.setFieldsValue({
      action: 'add'
    })
    setLabelModalOpen(true)
  }

  async function submitHostLabel(values: HostLabelFormValues) {
    if (!selectedClusterId || !host || !active.current || submitting || labelRunning.current || actionRunning.current || loading || error || !labelVersion || !hostDeleteVersion(host)) {
      return
    }
    const name = hostName(host)
    const label = values.label?.trim()
    if (!name) {
      message.error('无法识别主机名')
      return
    }
    if (!label || !['add', 'rm'].includes(values.action ?? '')) {
      message.error('请输入标签名称')
      return
    }
    labelRunning.current = true
    setSubmitting(true)
    try {
      await operationMutation.run(() => mutateResource('/host', 'PATCH', {
        cluster_id: selectedClusterId,
        host: name,
        labels_add: values.action === 'rm' ? [] : [label],
        labels_remove: values.action === 'rm' ? [label] : []
      }, { ifMatch: labelVersion }), false)
      if (!active.current) return
      setLabelModalOpen(false)
      message.success('主机标签更新成功')
      void refresh({ showLoading: false })
    } finally {
      labelRunning.current = false
      if (active.current) { setSubmitting(false); setLabelVersion(null) }
    }
  }

  async function runHostAction(action: string) {
    if (!selectedClusterId || !host || !active.current || pendingAction || actionRunning.current || loading || error) {
      return
    }
    const name = hostName(host)
    if (!name) {
      message.error('无法识别主机名')
      return
    }
    const version = hostDeleteVersion(host)
    const warnings: Record<string, string> = {
      maintenance_enter: '进入维护将停止该主机上的 Ceph 守护进程，可能影响服务可用性。本操作不强制绕过安全检查。',
      maintenance_exit: '退出维护将恢复该主机的守护进程管理，可能启动服务并触发恢复。',
      drain: 'Drain 将排空该主机上的守护进程并调度 OSD 移除，可能触发大量数据迁移。命令接受不代表排空已完成。',
      stop_drain: '停止 Drain 不会恢复已经移除的守护进程，也不能撤销已经发生的数据迁移。请另行核对 OSD 移除队列。',
      rescan: '重新扫描主机设备，发现结果可能需要等待库存刷新；不会清空磁盘。'
    }
    if (!version || !Object.keys(warnings).includes(action)) {
      message.error('主机库存过期、版本无效或操作不支持，请刷新后再操作')
      return
    }
    let submitted = false
    deleteConfirmation.current?.destroy()
    deleteConfirmation.current = Modal.confirm({
      title: `集群 ${selectedClusterId} / 主机 ${name} / ${action}`,
      content: warnings[action], okText: '确认执行', okType: 'danger', cancelText: '取消',
      async onOk() {
        if (!active.current || submitted || actionRunning.current) return
        submitted = true
        actionRunning.current = true
        setPendingAction(`${name}:${action}`)
        try {
          await operationMutation.run(() => mutateResource('/host/action', 'POST', {
            cluster_id: selectedClusterId, host: name, action
          }, { ifMatch: version }), false)
          if (!active.current) return
          message.success(`主机 ${action} 命令已接受，请核对原生状态和队列`)
          await refresh()
        } finally {
          actionRunning.current = false
          if (active.current) setPendingAction('')
        }
      }
    })
  }

  async function deleteHost() {
    if (!selectedClusterId || !host || !active.current || loading || error || pendingAction) {
      message.error('请先选择集群')
      return
    }
    const name = hostName(host)
    if (!name) {
      message.error('无法识别主机名')
      return
    }
    const generation = hostDeleteVersion(host)
    if (!generation) {
      message.error('主机库存过期或版本无效，请刷新后再删除')
      return
    }
    let submitted = false
    deleteConfirmation.current?.destroy()
    deleteConfirmation.current = Modal.confirm({
      title: `集群 ${selectedClusterId} / 删除主机 ${name}`,
      content: '将从编排器中移除该主机，不会自动迁移守护进程或清空磁盘。请先确认主机工作负载已妥善处理；失败后应刷新核对，不要直接重试。',
      okText: '提交删除',
      okType: 'danger',
      cancelText: '取消',
      async onOk() {
        if (!active.current || submitted) return
        submitted = true
        await operationMutation.run(() => mutateResource('/host', 'DELETE', {
          cluster_id: selectedClusterId,
          host: name
        }, { ifMatch: generation }), false)
        if (!active.current) return
        message.success('主机删除执行成功')
        navigate('/cluster/host')
      }
    })
  }

  return (
    <Page title="主机详情" loading={loading} error={error}>
      {data?.inventoryWarnings.map(warning => <Alert key={warning} type="warning" showIcon message={warning} />)}
      <Space direction="vertical" size={16} className="page-stack">
        <Card
          className="page-surface-card"
          title="基础信息"
          extra={
            <Space className="host-detail-actions">
              <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/cluster/host')}>返回</Button>
              <Button icon={<ReloadOutlined />} loading={refreshing || loading} onClick={refreshHostDetail}>刷新</Button>
              <Button danger icon={<DeleteOutlined />} disabled={!hostDeleteVersion(host) || loading || Boolean(error) || Boolean(pendingAction)} onClick={deleteHost}>删除</Button>
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
              <Descriptions.Item label="初始 CRUSH 位置（主机规格）" span={2}>{hostInitialLocation(host.location)}<br /><Text type="secondary">这是编排器主机规格中的初始位置，不代表当前 CRUSH 树；请在 CRUSH 拓扑中核对实际位置。</Text></Descriptions.Item>
              <Descriptions.Item label="服务" span={2}>{renderServiceInstances(serviceRows(host, data?.daemons ?? []))}</Descriptions.Item>
              <Descriptions.Item label="创建时间">{formatDateTime(host.created_at)}</Descriptions.Item>
              <Descriptions.Item label="更新时间">{formatDateTime(host.updated_at)}</Descriptions.Item>
            </Descriptions>
          ) : (
            <Alert type="info" message={loading ? '正在读取主机详情' : '当前库存中未找到该主机'} description="主机可能已删除或库存尚未同步。可刷新库存重试；独立诊断数据不代表主机仍存在。" />
          )}
        </Card>

        {host && <HostNativeSummary host={host} />}
        <Card className="page-surface-card" title="主机操作">
          <Space wrap>
            {host && selectedClusterId && <HostAddressEditor key={`${selectedClusterId}:${hostName(host)}:${host.resource_version}`} clusterId={selectedClusterId} hostname={hostName(host)} currentAddress={textValue(host.address ?? host.addr, '')} version={hostDeleteVersion(host)} blocked={loading || Boolean(error) || Boolean(pendingAction) || submitting} />}
            <Button icon={<TagOutlined />} disabled={!host} onClick={openLabelModal}>标签</Button>
            <Button loading={pendingAction.endsWith(':maintenance_enter')} disabled={!host || Boolean(pendingAction)} onClick={() => runHostAction('maintenance_enter')}>维护</Button>
            <Button loading={pendingAction.endsWith(':maintenance_exit')} disabled={!host || Boolean(pendingAction)} onClick={() => runHostAction('maintenance_exit')}>退出维护</Button>
            <Button loading={pendingAction.endsWith(':drain')} disabled={!host || Boolean(pendingAction)} onClick={() => runHostAction('drain')}>Drain</Button>
            <Button loading={pendingAction.endsWith(':stop_drain')} disabled={!host || Boolean(pendingAction)} onClick={() => runHostAction('stop_drain')}>停止 Drain</Button>
            <Button loading={pendingAction.endsWith(':rescan')} disabled={!host || Boolean(pendingAction)} onClick={() => runHostAction('rescan')}>Rescan</Button>
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
          deviceInfoError={data?.deviceInfoError ?? ''}
          smartError={data?.smartError ?? ''}
        />
      </Space>

      <DraggableModal
        title="标签管理"
        open={labelModalOpen}
        onCancel={() => setLabelModalOpen(false)}
        onOk={() => labelForm.submit()}
        okText="提交"
        confirmLoading={submitting}
        okButtonProps={{ disabled: !labelVersion || loading || Boolean(error) }}
        destroyOnClose
      >
        <Alert type="warning" message={`集群 ${selectedClusterId} / 主机 ${host ? hostName(host) : decodedName}：标签可能影响服务放置、调度及配置分发，特殊标签也有运行影响。`} />
        {!labelVersion && <Alert type="info" message="本次编辑已提交或版本不可用。请关闭弹窗，刷新主机后重新打开，避免直接重试。" />}
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
  smart,
  deviceInfoError,
  smartError
}: {
  hostname: string
  address: string
  clusterId?: number
  deviceInfo: ApiRecord[]
  devices: ApiRecord[]
  daemons: ApiRecord[]
  smart: ApiRecord
  deviceInfoError: string
  smartError: string
}) {
  const deviceInfoRows = useMemo(() => deviceInfo.map((device) => normalizeCephDeviceRow(device, hostname)), [deviceInfo, hostname])
  const physicalDiskRows = useMemo(
    () => devices.map((device) => normalizeInventoryDeviceRow(device, deviceInfoRows)),
    [deviceInfoRows, devices]
  )
  const healthRows = useMemo(() => normalizeSMARTData(smart), [smart])

  return (
    <Card className="page-surface-card host-detail-tabs-card" styles={{ body: { paddingTop: 0 } }}>
      {deviceInfoError && <Alert type="error" message={`设备信息读取失败：${deviceInfoError}`} description="物理磁盘仍可展示独立库存，但设备信息补充字段不完整；请刷新重试。" />}
      {smartError && <Alert type="error" message={`SMART 读取失败：${smartError}`} description="当前无法判断设备健康，不代表没有设备或设备健康。请刷新重试。" />}
      <Tabs
        className="host-detail-tabs"
        defaultActiveKey="devices"
        destroyInactiveTabPane
        items={[
          { key: 'hardware', label: '硬件健康', children: clusterId ? <HostHardware key={`${clusterId}:${hostname}`} clusterId={clusterId} host={hostname} /> : null },
          {
            key: 'devices',
            label: '设备信息',
            children: deviceInfoError ? <Alert type="warning" message="设备信息不可用，不能根据空结果判断没有设备" /> : <HostDeviceInfoTable devices={deviceInfoRows} />
          },
          {
            key: 'physical-disks',
            label: '物理硬盘',
            children: <HostPhysicalDiskTable devices={physicalDiskRows} />
          },
          {
            key: 'daemons',
            label: '守护进程',
            children: <HostDaemonTable key={`${clusterId}:${hostname}`} clusterId={clusterId} daemons={daemons} />
          },
          {
            key: 'performance',
            label: '性能详细信息',
            children: <HostPerformancePanel hostname={hostname} address={address} clusterId={clusterId} />
          },
          {
            key: 'health',
            label: '设备健康状态',
            children: smartError ? <Alert type="warning" message="SMART 信息不可用，健康状态未知" /> : <HostDeviceHealthPanel devices={healthRows} />
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
        { key: 'life_expectancy_stamp', title: '预测生成时间', render: (value) => formatDateTime(value) },
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
        { key: 'serial_display', title: '序列号' },
        { key: 'health_display', title: 'LSM 健康状态' },
        { key: 'rejected_reasons_display', title: '不可用原因', ellipsis: false },
        { key: 'size_display', title: '容量' },
        { key: 'osd_display', title: 'OSD' }
      ]}
    />
  )
}

function HostDaemonTable({ daemons, clusterId }: { daemons: ApiRecord[]; clusterId?: number }) {
  const [perfName, setPerfName] = useState<string | null>(null)
  const visibleName = perfName && daemons.some((row) => row.daemon_display === perfName) ? perfName : null
  return (
    <Space direction="vertical" style={{ width: '100%' }}>
    <DataTable
      data={daemons}
      rowKeyCandidates={['name', 'daemon_name', 'daemon_display']}
      columns={[
        { key: 'daemon_display', title: '名称' },
        { key: 'type_display', title: '类型' },
        { key: 'status_display', title: '状态', render: (value) => renderDaemonStatus(value) },
        { key: 'status_desc', title: '原生状态说明', ellipsis: false },
        { key: 'runtime_details', title: '运行详情', filterKey: false, ellipsis: false, render: (_, row) => <DaemonRuntimeDetails row={{ ...row, daemon_type: row.type_display }} /> },
        { key: 'last_refresh_display', title: '最近刷新' },
        { key: 'version_display', title: '版本' },
        { key: 'cpu_usage_display', title: 'CPU 使用率' },
        { key: 'memory_usage_display', title: '内存使用量' },
        { key: 'image_display', title: '镜像', render: (value, row) => renderImageInfo(value, row) },
        { key: 'performance', title: '性能', filterKey: false, render: (_, row) => <Button
          disabled={!clusterId || !hostDaemonSupportsPerf(row.daemon_display)}
          onClick={() => setPerfName(String(row.daemon_display))}>性能计数器</Button> }
      ]}
    />
    {clusterId && visibleName && <Space direction="vertical" style={{ width: '100%' }}>
      <Space><Text strong>{visibleName} 性能计数器</Text><Button onClick={() => setPerfName(null)}>关闭性能详情</Button></Space>
      <DaemonPerf key={`${clusterId}:${visibleName}`} clusterId={clusterId} name={visibleName} />
    </Space>}
    </Space>
  )
}

function hostDaemonSupportsPerf(name: unknown): name is string {
  return typeof name === 'string' && /^(mon|mgr|mds|osd)\.[A-Za-z0-9_][A-Za-z0-9_.-]*$/.test(name)
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
        { key: 'error_display', title: '读取错误', ellipsis: false },
        { key: 'smartctl_error_code', title: 'smartctl 错误码' },
        { key: 'temperature_display', title: '温度' },
        { key: 'power_on_hours_display', title: '通电时间' },
        { key: 'wear_level_display', title: '已使用寿命（percentage_used）' },
        { key: 'ssd_life_left_display', title: 'ssd_life_left（原始值）' },
        { key: 'smart_report', title: 'SMART 详情', ellipsis: false, filterKey: false, render: (value) => <details><summary>展开 ATA / SCSI / NVMe 详情</summary><SMARTDetails data={value} /></details> }
      ]}
    />
  )
}

function normalizeCephDeviceRow(row: ApiRecord, hostname: string): ApiRecord {
  const locations = Array.isArray(row.location) ? row.location.filter((location) => isRecord(location) && location.host === hostname) : []
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
    device_names: deviceNames,
    name_display: deviceNames.length ? deviceNames.join(', ') : '-',
    daemons_display: daemons.length ? daemons : []
  }
}

function normalizeInventoryDeviceRow(row: ApiRecord, deviceInfo: ApiRecord[]): ApiRecord {
  const sysAPI = isRecord(row.sys_api) ? row.sys_api : {}
  const lsmData = isRecord(row.lsm_data) ? row.lsm_data : {}
  const lvs = Array.isArray(row.lvs) ? row.lvs.filter(isRecord) : []
  const path = textValue(row.path ?? sysAPI.path ?? row.name, '')
  const deviceName = path.split('/').filter(Boolean).pop() ?? path
  const linkedDevice = deviceInfo.find((device) => stringArray(device.device_names).includes(deviceName))
  const osdIDs = inventoryOSDNames([
    ...(Array.isArray(row.osd_ids) ? row.osd_ids : []),
    ...lvs.map((lv) => lv.osd_id),
    ...stringArray(linkedDevice?.daemons_display)
  ])
  const rotational = row.rotational ?? sysAPI.rotational
  const inferredType = rotational === true || rotational === '1' || rotational === 1 ? 'HDD'
    : rotational === false || rotational === '0' || rotational === 0 ? 'SSD' : '未知'
  const type = textValue(row.device_type ?? row.human_readable_type ?? row.type, '') || inferredType
  const size = numberValue(row.size_bytes ?? row.size ?? sysAPI.size)
  return {
    ...row,
    device_id: textValue(row.device_id ?? row.devid ?? row.id, ''),
    path_display: path,
    name_display: path,
    type_display: type.toUpperCase(),
    availability_display: row.available === true ? 'available' : row.available === false ? 'unavailable' : 'unknown',
    rejected_reasons_display: stringArray(row.rejected_reasons),
    vendor_display: textValue(row.vendor ?? sysAPI.vendor, ''),
    model_display: textValue(row.model ?? sysAPI.model, ''),
    serial_display: textValue(row.serial ?? lsmData.serialNum, ''),
    size_display: size ? formatBytes(size) : '-',
    osd_display: osdIDs,
    health_display: textValue(lsmData.health, '')
  }
}

function inventoryOSDNames(values: unknown[]): string[] {
  const names = new Set<string>()
  for (const value of values) {
    if (typeof value === 'number' && (!Number.isSafeInteger(value) || value < 0)) continue
    if (typeof value !== 'string' && typeof value !== 'number') continue
    const id = String(value).replace(/^osd\./, '')
    if (!/^\d+$/.test(id)) continue
    names.add(`osd.${BigInt(id)}`)
  }
  return [...names]
}

function normalizeSMARTData(value: ApiRecord): ApiRecord[] {
  return Object.entries(value).flatMap<ApiRecord>(([deviceID, raw]) => {
    if (!isRecord(raw) || Object.prototype.hasOwnProperty.call(raw, 'error')) {
      const error = isRecord(raw) ? raw.error : '设备 SMART 响应格式无效'
      return [{
        device_id: deviceID, name_display: deviceID, health_display: 'unavailable',
        error_display: typeof error === 'string' && error.trim() ? error : '设备 SMART 读取失败，未返回有效错误说明',
        smartctl_error_code: isRecord(raw) ? raw.smartctl_error_code : undefined,
        smart_report: raw,
      }]
    }
    const smartStatus = isRecord(raw.smart_status) ? raw.smart_status : {}
    const device = isRecord(raw.device) ? raw.device : {}
    const temperature = isRecord(raw.temperature) ? raw.temperature : {}
    const powerOnTime = isRecord(raw.power_on_time) ? raw.power_on_time : {}
    const nvmeHealth = isRecord(raw.nvme_smart_health_information_log) ? raw.nvme_smart_health_information_log : {}
    const passed = smartStatus.passed
    return [{
      ...raw,
      smart_report: raw,
      device_id: deviceID,
      name_display: textValue(device.name ?? raw.dev ?? deviceID, deviceID),
      serial_display: textValue(raw.serial_number ?? raw.serial ?? deviceID, ''),
      health_display: typeof passed === 'boolean' ? (passed ? 'good' : 'bad') : textValue(raw.health ?? raw.state, 'unknown'),
      temperature_display: formatTemperature(smartMetricNumber(temperature.current ?? raw.temperature ?? nvmeHealth.temperature)),
      power_on_hours_display: formatHours(powerOnTime.hours ?? raw.power_on_hours ?? nvmeHealth.power_on_hours),
      wear_level_display: formatWear(smartMetricNumber(raw.percentage_used ?? nvmeHealth.percentage_used)),
      ssd_life_left_display: raw.ssd_life_left == null ? '-' : String(raw.ssd_life_left)
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
    unknown: { label: '未知', color: 'default' },
    unavailable: { label: '读取失败（健康未知）', color: 'error' }
  }
  const item = Object.prototype.hasOwnProperty.call(mapping, status) ? mapping[status] : { label: textValue(value), color: 'default' }
  return <Tag color={item.color}>{item.label}</Tag>
}

function renderDeviceAvailability(row: ApiRecord) {
  if (row.availability_display === 'unknown') {
    return <Tag>未知</Tag>
  }
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

function smartMetricNumber(value: unknown): number | undefined {
  if (typeof value === 'string') {
    if (!/^-?\d+(\.\d+)?$/.test(value)) return undefined
    value = Number(value)
  }
  return typeof value === 'number' && Number.isFinite(value) && Math.abs(value) <= Number.MAX_SAFE_INTEGER ? value : undefined
}

function formatTemperature(value?: number) {
  return value === undefined ? '-' : `${value} °C`
}

function formatHours(value: unknown) {
  if (typeof value === 'string' && /^\d+(\.\d+)?$/.test(value)) return `${value} 小时`
  return typeof value === 'number' && Number.isFinite(value) && Math.abs(value) <= Number.MAX_SAFE_INTEGER ? `${value.toLocaleString()} 小时` : '-'
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
    status_display: nativeDaemonStatus(row),
    last_refresh_display: formatDateTime(row.last_refresh),
    version_display: textValue(row.version, ''),
    cpu_usage_display: cpuUsage === undefined ? '-' : `${cpuUsage.toFixed(1)}%`,
    memory_usage_display: memoryUsage === undefined ? '-' : formatBytes(memoryUsage),
    image_display: shortImageName(image),
    image_full: image
  }
}

function nativeDaemonStatus(row: ApiRecord): string {
  const codes: Record<string, string> = { '-2': 'unknown', '-1': 'error', '0': 'stopped', '1': 'running', '2': 'starting' }
  if (typeof row.status === 'number') return Object.prototype.hasOwnProperty.call(codes, String(row.status)) ? codes[String(row.status)] : 'unknown'
  if (typeof row.status === 'string' && row.status.trim()) return Object.prototype.hasOwnProperty.call(codes, row.status) ? codes[row.status] : row.status
  return typeof row.status_desc === 'string' && row.status_desc.trim() ? row.status_desc : 'unknown'
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
  if (normalized === 'running' || normalized === 'ok') {
    return 'success'
  }
  if (['error', 'failed', 'stopped'].includes(normalized)) {
    return 'error'
  }
  if (normalized === 'starting' || normalized === 'pending') {
    return 'processing'
  }
  return 'default'
}

function daemonStatusText(status: string) {
  const normalized = status.toLowerCase()
  if (normalized === 'running' || normalized === 'ok') {
    return '运行中'
  }
  if (normalized === 'starting' || normalized === 'pending') {
    return '启动中'
  }
  if (normalized === 'stopping') {
    return '停止中'
  }
  if (normalized === 'stopped') {
    return '已停止'
  }
  if (normalized === 'error' || normalized === 'failed') {
    return '异常'
  }
  return normalized === 'unknown' ? '未知' : status
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

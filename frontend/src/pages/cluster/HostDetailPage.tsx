import { ArrowLeftOutlined, DeleteOutlined, ReloadOutlined, TagOutlined } from '@ant-design/icons'
import { Button, Card, Descriptions, Form, Input, Modal, Select, Space, Tag, Tooltip, Typography } from 'antd'
import { useCallback, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { isRecord, textValue, type ApiRecord } from '../../api/client'
import { getOptionalResource, listDaemons, listHostDevices, listResource, mutateResource, refreshResource } from '../../api/resource'
import type { ResourceDTO } from '../../api/types'
import { DataTable } from '../../components/DataTable'
import { DraggableModal } from '../../components/DraggableModal'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useClusterContext } from '../../state/ClusterContext'
import { message } from '../../utils/appMessage'
import { formatDateTime } from '../../utils/time'
import { hostName, normalizeHostRow, serviceInstanceRows } from './HostPage'

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
      return { host: null, daemons: [], devices: [] }
    }
    const [hostPayload, daemons, devices] = await Promise.all([
      getOptionalResource('/host', selectedClusterId, { host: decodedName }),
      listDaemons(),
      listHostDevices(decodedName).then(async (items) => {
        if (items.length > 0) {
          return items
        }
        const payload = await listResource('/host/devices')
        return payload.items.filter((device) => textValue(device.hostname ?? device.host, '') === decodedName)
      })
    ])
    const hostRecord = hostPayload ? resourceToRecord(hostPayload.item) : { hostname: decodedName }
    const host = normalizeHostRow(hostRecord, daemons, devices)
    const name = textValue(host.hostname ?? decodedName, '')
    return {
      host,
      daemons: daemons
        .filter((daemon) => textValue(daemon.hostname ?? daemon.host, '') === name)
        .map(normalizeDaemonRow),
      devices
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
    if (!name) {
      message.error('无法识别主机名')
      return
    }
    setSubmitting(true)
    try {
      await operationMutation.run(() => mutateResource('/host', 'PATCH', {
        cluster_id: selectedClusterId,
        host: name,
        label: values.label,
        action: values.action ?? 'add'
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

        <Card className="page-surface-card" title="守护进程">
          <DataTable
            data={data?.daemons ?? []}
            rowKeyCandidates={['name', 'daemon_name', 'daemon_display']}
            columns={[
              { key: 'daemon_display', title: '名称' },
              { key: 'type_display', title: '类型' },
              { key: 'status_display', title: '状态', render: (value) => renderDaemonStatus(value) },
              { key: 'version_display', title: '版本' },
              { key: 'image_display', title: '镜像', render: (value, row) => renderImageInfo(value, row) }
            ]}
          />
        </Card>
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
  return {
    ...row,
    daemon_display: textValue(row.name ?? row.daemon_name, ''),
    type_display: textValue(row.type ?? row.daemon_type ?? row.service_type, ''),
    status_display: textValue(row.status ?? row.status_desc, ''),
    version_display: textValue(row.version, ''),
    image_display: shortImageName(image),
    image_full: image
  }
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

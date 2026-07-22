import { ReloadOutlined, ThunderboltOutlined } from '@ant-design/icons'
import { Button, Card, Form, InputNumber, Modal, Space, Switch, Tabs, Tag, message } from 'antd'
import { useCallback, useState } from 'react'
import { textValue, type ApiRecord } from '../../api/client'
import {
  applyDaemonAction,
  listDaemons,
  listMgrModules,
  listMonitors,
  listOSDFlags,
  listOSDs,
  listServices,
  markOSD,
  reweightOSD,
  scrubOSD,
  setMgrModuleEnabled
} from '../../api/resource'
import { DataTable } from '../../components/DataTable'
import { draggableModalRender } from '../../components/DraggableModal'
import { Page } from '../../components/Page'
import { useResource } from '../../hooks'
import { ClusterDetailPage } from './ClusterDetailPage'
import { ClusterPage } from './ClusterPage'
import { HostPage } from './HostPage'
import { ServicePage } from './ServicePage'

export { ClusterDetailPage, ClusterPage, HostPage, ServicePage }

export function MonManagementPage() {
  const loader = useCallback(async () => {
    const [monitor, daemons] = await Promise.all([listMonitors(), listDaemons('mon')])
    return {
      monitor,
      daemons,
      inQuorum: asRecords(monitor.in_quorum),
      outQuorum: asRecords(monitor.out_quorum)
    }
  }, [])
  const { data, loading, error, refresh } = useResource(loader)

  return (
    <Page title="MON管理" loading={loading} error={error}>
      <Card className="page-surface-card" title="MON管理">
        <Tabs
          items={[
            {
              key: 'quorum',
              label: '仲裁成员',
              children: (
                <div className="embedded-panel">
                <DataTable
                  data={data?.inQuorum ?? []}
                  rowKeyCandidates={['name', 'rank', 'addr']}
                  columns={[
                    { key: 'name', title: '名称' },
                    { key: 'rank', title: 'Rank' },
                    { key: 'public_addr', title: 'Public Addr' },
                    { key: 'priority', title: '优先级' },
                    { key: 'stats', title: '会话统计', render: (value) => textValue(value) }
                  ]}
                />
                </div>
              )
            },
            {
              key: 'out',
              label: '非仲裁成员',
              children: (
                <div className="embedded-panel">
                <DataTable
                  data={data?.outQuorum ?? []}
                  rowKeyCandidates={['name', 'rank', 'addr']}
                  columns={[
                    { key: 'name', title: '名称' },
                    { key: 'rank', title: 'Rank' },
                    { key: 'public_addr', title: 'Public Addr' },
                    { key: 'addr', title: '地址' }
                  ]}
                />
                </div>
              )
            },
            {
              key: 'daemons',
              label: '守护进程',
              children: <DaemonTable data={data?.daemons ?? []} refresh={refresh} />
            }
          ]}
        />
      </Card>
    </Page>
  )
}

export function MgrManagementPage() {
  const loader = useCallback(async () => {
    const [modules, daemons] = await Promise.all([listMgrModules(), listDaemons('mgr')])
    return { modules, daemons }
  }, [])
  const { data, loading, error, refresh } = useResource(loader)
  const [pendingModule, setPendingModule] = useState('')

  async function toggleModule(row: ApiRecord, enabled: boolean) {
    const name = textValue(row.name, '')
    if (!name || pendingModule) {
      return
    }
    setPendingModule(name)
    try {
      await setMgrModuleEnabled(name, enabled)
      message.success(enabled ? 'Mgr 模块已启用' : 'Mgr 模块已停用')
      refresh()
    } finally {
      setPendingModule('')
    }
  }

  return (
    <Page title="MGR管理" loading={loading} error={error}>
      <Card className="page-surface-card" title="MGR管理">
        <Tabs
          items={[
            {
              key: 'modules',
              label: '模块',
              children: (
                <div className="embedded-panel">
                <DataTable
                  data={data?.modules ?? []}
                  rowKeyCandidates={['name']}
                  columns={[
                    { key: 'name', title: '模块' },
                    {
                      key: 'enabled',
                      title: '启用',
                      render: (value, row) => {
                        const name = textValue(row.name, '')
                        return (
                          <Switch
                            checked={Boolean(value)}
                            disabled={Boolean(row.always_on) || (Boolean(pendingModule) && pendingModule !== name)}
                            loading={pendingModule === name}
                            onChange={(checked) => toggleModule(row, checked)}
                          />
                        )
                      }
                    },
                    { key: 'always_on', title: '常驻', render: (value) => <Tag color={value ? 'processing' : 'default'}>{value ? '是' : '否'}</Tag> },
                    { key: 'options', title: '配置项', render: (value) => textValue(value) }
                  ]}
                />
                </div>
              )
            },
            {
              key: 'daemons',
              label: '守护进程',
              children: <DaemonTable data={data?.daemons ?? []} refresh={refresh} />
            }
          ]}
        />
      </Card>
    </Page>
  )
}

export function OsdManagementPage() {
  const loader = useCallback(async () => {
    const [osds, flags] = await Promise.all([listOSDs(), listOSDFlags()])
    return { osds, flags }
  }, [])
  const { data, loading, error, refresh } = useResource(loader)
  const [pendingOSDAction, setPendingOSDAction] = useState('')

  async function runOSDAction(id: string, action: 'in' | 'out' | 'scrub' | 'deep-scrub' | 'reweight') {
    if (action === 'reweight') {
      Modal.confirm({
        title: `调整 OSD ${id} 权重`,
        content: <ReweightForm osdID={id} refresh={refresh} />,
        modalRender: draggableModalRender,
        icon: null,
        okButtonProps: { style: { display: 'none' } },
        cancelText: '关闭'
      })
      return
    }

    const pendingKey = `${id}:${action}`
    if (pendingOSDAction) {
      return
    }
    setPendingOSDAction(pendingKey)
    try {
      if (action === 'scrub' || action === 'deep-scrub') {
        await scrubOSD(id, action === 'deep-scrub')
        message.success('Scrub 任务已提交')
      } else {
        await markOSD(id, action)
        message.success(`OSD 已标记为 ${action}`)
      }
      refresh()
    } finally {
      setPendingOSDAction('')
    }
  }

  return (
    <Page title="OSD管理" loading={loading} error={error}>
      <Card className="page-surface-card" title="OSD管理">
        <Space direction="vertical" size={16} className="page-stack">
        <section className="embedded-panel">
          <div className="embedded-panel-title">OSD Flags</div>
          {(data?.flags ?? []).length ? data?.flags.map((flag) => <Tag key={flag}>{flag}</Tag>) : <span className="muted">未设置 OSD flags</span>}
        </section>
        <section className="embedded-panel">
          <DataTable
            data={data?.osds ?? []}
            rowKeyCandidates={['id', 'osd', 'service_id', 'name']}
            columns={[
              { key: 'id', title: 'ID' },
              { key: 'hostname', title: '主机' },
              { key: 'state', title: '状态' },
              { key: 'up', title: 'Up' },
              { key: 'in', title: 'In' },
              { key: 'device_class', title: '设备类型' },
              { key: 'stats', title: '容量/统计' },
              {
                key: 'actions',
                title: '操作',
                render: (_, row) => {
                  const id = osdID(row)
                  return (
                    <Space>
                      <Button size="small" loading={pendingOSDAction === `${id}:in`} disabled={Boolean(pendingOSDAction) && pendingOSDAction !== `${id}:in`} onClick={() => runOSDAction(id, 'in')}>In</Button>
                      <Button size="small" loading={pendingOSDAction === `${id}:out`} disabled={Boolean(pendingOSDAction) && pendingOSDAction !== `${id}:out`} onClick={() => runOSDAction(id, 'out')}>Out</Button>
                      <Button size="small" icon={<ThunderboltOutlined />} loading={pendingOSDAction === `${id}:scrub`} disabled={Boolean(pendingOSDAction) && pendingOSDAction !== `${id}:scrub`} onClick={() => runOSDAction(id, 'scrub')}>Scrub</Button>
                      <Button size="small" disabled={Boolean(pendingOSDAction)} onClick={() => runOSDAction(id, 'reweight')}>权重</Button>
                    </Space>
                  )
                }
              }
            ]}
          />
        </section>
        </Space>
      </Card>
    </Page>
  )
}

export function MdsManagementPage() {
  const loader = useCallback(async () => {
    const [services, daemons] = await Promise.all([listServices(), listDaemons('mds')])
    return {
      services: services.filter((service) => textValue(service.service_type || service.type, '').toLowerCase() === 'mds'),
      daemons
    }
  }, [])
  const { data, loading, error, refresh } = useResource(loader)

  return (
    <Page title="MDS管理" loading={loading} error={error}>
      <Card className="page-surface-card" title="MDS管理">
        <Tabs
          items={[
            {
              key: 'services',
              label: 'MDS服务',
              children: (
                <div className="embedded-panel">
                <DataTable
                  data={data?.services ?? []}
                  rowKeyCandidates={['service_name', 'service_id', 'name']}
                  columns={[
                    { key: 'service_name', title: '服务名' },
                    { key: 'placement', title: '放置策略' },
                    { key: 'status', title: '状态' },
                    { key: 'running', title: '运行数' },
                    { key: 'size', title: '目标数' }
                  ]}
                />
                </div>
              )
            },
            {
              key: 'daemons',
              label: '守护进程',
              children: <DaemonTable data={data?.daemons ?? []} refresh={refresh} />
            }
          ]}
        />
      </Card>
    </Page>
  )
}

function DaemonTable({ data, refresh }: { data: ApiRecord[]; refresh: () => void }) {
  const [pendingDaemonAction, setPendingDaemonAction] = useState('')

  async function runAction(row: ApiRecord, action: string) {
    const name = textValue(row.daemon_name || row.name, '')
    const pendingKey = `${name}:${action}`
    if (!name || pendingDaemonAction) {
      return
    }
    setPendingDaemonAction(pendingKey)
    try {
      await applyDaemonAction(name, action, action === 'restart')
      message.success(`Daemon ${action} 已提交`)
      refresh()
    } finally {
      setPendingDaemonAction('')
    }
  }

  return (
    <div className="embedded-panel">
      <DataTable
        data={data}
        rowKeyCandidates={['daemon_name', 'name', 'hostname']}
        columns={[
          { key: 'daemon_name', title: 'Daemon' },
          { key: 'daemon_type', title: '类型' },
          { key: 'hostname', title: '主机' },
          { key: 'status_desc', title: '状态' },
          { key: 'version', title: '版本' },
          {
            key: 'actions',
            title: '操作',
            render: (_, row) => {
              const name = textValue(row.daemon_name || row.name, '')
              return (
                <Space>
                  <Button size="small" icon={<ReloadOutlined />} loading={pendingDaemonAction === `${name}:restart`} disabled={Boolean(pendingDaemonAction) && pendingDaemonAction !== `${name}:restart`} onClick={() => runAction(row, 'restart')}>重启</Button>
                  <Button size="small" loading={pendingDaemonAction === `${name}:start`} disabled={Boolean(pendingDaemonAction) && pendingDaemonAction !== `${name}:start`} onClick={() => runAction(row, 'start')}>启动</Button>
                  <Button size="small" danger loading={pendingDaemonAction === `${name}:stop`} disabled={Boolean(pendingDaemonAction) && pendingDaemonAction !== `${name}:stop`} onClick={() => runAction(row, 'stop')}>停止</Button>
                </Space>
              )
            }
          }
        ]}
      />
    </div>
  )
}

function ReweightForm({ osdID, refresh }: { osdID: string; refresh: () => void }) {
  const [submitting, setSubmitting] = useState(false)

  async function submit(values: { weight: number }) {
    if (submitting) {
      return
    }
    setSubmitting(true)
    try {
      await reweightOSD(osdID, values.weight)
      message.success('OSD 权重调整已提交')
      Modal.destroyAll()
      refresh()
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Form layout="vertical" initialValues={{ weight: 1 }} onFinish={submit}>
      <Form.Item name="weight" label="权重" rules={[{ required: true }]}>
        <InputNumber min={0} max={1} step={0.01} precision={2} />
      </Form.Item>
      <Button type="primary" htmlType="submit" loading={submitting}>
        保存
      </Button>
    </Form>
  )
}

function asRecords(value: unknown): ApiRecord[] {
  return Array.isArray(value) ? value.filter((item): item is ApiRecord => typeof item === 'object' && item !== null && !Array.isArray(item)) : []
}

function osdID(row: ApiRecord) {
  return textValue(row.id ?? row.osd ?? row.service_id ?? row.name, '')
}

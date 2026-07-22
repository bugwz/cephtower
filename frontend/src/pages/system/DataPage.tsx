import { ReloadOutlined, RetweetOutlined, SaveOutlined } from '@ant-design/icons'
import { Button, Card, InputNumber, Select, Space, Switch, Table, Tag, Typography, message } from 'antd'
import { useCallback, useEffect, useState } from 'react'
import {
  listDataFetchRuns,
  listSystemSettings,
  resetSystemConfigDefaults,
  runDataFetchModule,
  updateSystemSetting,
  type DataFetchConfig,
  type DataFetchRun,
  type SystemSetting
} from '../../api/systemConfig'
import { Page } from '../../components/Page'

const { Text } = Typography

const moduleLabels: Record<string, string> = {
  summary: '集群概览',
  health: '健康状态',
  hosts: '主机',
  osds: 'OSD',
  osd_flags: 'OSD 标志',
  daemons: 'Daemon',
  services: '服务',
  monitors: 'MON',
  managers: 'MGR',
  mds: 'MDS',
  mgr_modules: 'Mgr 模块',
  cluster_configuration: '集群配置',
  pools: '存储池',
  rbd_images: 'RBD 镜像',
  cephfs: 'CephFS',
  rgw_daemons: 'RGW Daemon',
  rgw_users: 'RGW 用户',
  rgw_buckets: 'RGW Bucket'
}

interface DataFetchConfigRow extends DataFetchConfig {
  key: string
  updated_at: string
}

export function DataPage() {
  const [settings, setSettings] = useState<DataFetchConfigRow[]>([])
  const [runs, setRuns] = useState<DataFetchRun[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [pending, setPending] = useState<Record<string, boolean>>({})

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [nextSettings, nextRuns] = await Promise.all([listSystemSettings('ceph.data_fetch.'), listDataFetchRuns(30)])
      setSettings(nextSettings.map(toDataFetchConfigRow).filter((row): row is DataFetchConfigRow => Boolean(row)))
      setRuns(nextRuns)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载数据获取设置失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  async function patchSetting(setting: DataFetchConfigRow, payload: Partial<DataFetchConfig>, key: string) {
    const pendingKey = `${setting.key}:${key}`
    if (pending[pendingKey]) {
      return
    }
    setPending((current) => ({ ...current, [pendingKey]: true }))
    try {
      const updatedValue = JSON.stringify({ ...setting, ...payload, key: undefined, updated_at: undefined })
      const updated = await updateSystemSetting(setting.key, updatedValue)
      const row = toDataFetchConfigRow(updated)
      if (row) {
        setSettings((current) => current.map((item) => (item.key === setting.key ? row : item)))
      }
      message.success('系统配置已更新')
    } finally {
      setPending((current) => {
        const next = { ...current }
        delete next[pendingKey]
        return next
      })
    }
  }

  async function runNow(setting: DataFetchConfigRow) {
    const pendingKey = `${setting.key}:run`
    if (pending[pendingKey]) {
      return
    }
    setPending((current) => ({ ...current, [pendingKey]: true }))
    try {
      const result = await runDataFetchModule(setting.module)
      message.success(result.message || '数据获取任务已启动')
      await load()
    } finally {
      setPending((current) => {
        const next = { ...current }
        delete next[pendingKey]
        return next
      })
    }
  }

  async function resetDefaults() {
    const result = await resetSystemConfigDefaults()
    message.success(result.message || '系统配置已恢复默认')
    await load()
  }

  return (
    <Page title="配置管理" loading={loading} error={error}>
      <Space direction="vertical" size={16} className="page-stack">
        <Card
          className="page-surface-card"
          title="Ceph 数据获取配置"
          extra={
            <Space>
              <Button icon={<ReloadOutlined />} loading={loading} onClick={load}>
                刷新
              </Button>
              <Button icon={<RetweetOutlined />} onClick={resetDefaults}>
                恢复默认
              </Button>
            </Space>
          }
        >
          <Table
            size="middle"
            rowKey="key"
            rowClassName="stable-row"
            dataSource={settings}
            pagination={{ pageSize: 12, showSizeChanger: false }}
            scroll={{ x: true }}
            columns={[
              { title: '集群', dataIndex: 'cluster_id', width: 90 },
              {
                title: '模块',
                dataIndex: 'module',
                width: 180,
                render: (module: string) => moduleLabels[module] || module
              },
              {
                title: '状态',
                dataIndex: 'enabled',
                width: 130,
                render: (enabled: boolean, setting) => (
                  <Space>
                    <Switch
                      checked={enabled}
                      loading={Boolean(pending[`${setting.key}:enabled`])}
                      onChange={(value) => patchSetting(setting, { enabled: value }, 'enabled')}
                    />
                    <Tag color={enabled ? 'success' : 'default'}>{enabled ? '启用' : '停用'}</Tag>
                  </Space>
                )
              },
              {
                title: '频率（秒）',
                dataIndex: 'interval_seconds',
                width: 150,
                render: (value: number, setting) => (
                  <InputNumber
                    min={10}
                    step={30}
                    value={value}
                    disabled={Boolean(pending[`${setting.key}:interval`])}
                    onChange={(next) => {
                      if (typeof next === 'number') {
                        void patchSetting(setting, { interval_seconds: next }, 'interval')
                      }
                    }}
                  />
                )
              },
              {
                title: '超时（秒）',
                dataIndex: 'timeout_seconds',
                width: 140,
                render: (value: number, setting) => (
                  <InputNumber
                    min={1}
                    step={5}
                    value={value}
                    disabled={Boolean(pending[`${setting.key}:timeout`])}
                    onChange={(next) => {
                      if (typeof next === 'number') {
                        void patchSetting(setting, { timeout_seconds: next }, 'timeout')
                      }
                    }}
                  />
                )
              },
              {
                title: '来源',
                dataIndex: 'fetch_source',
                width: 150,
                render: (value: string, setting) => (
                  <Select
                    value={value}
                    className="role-select"
                    options={[
                      { label: 'command', value: 'command' },
                      { label: 'dashboard', value: 'dashboard' },
                      { label: 'mixed', value: 'mixed' }
                    ]}
                    loading={Boolean(pending[`${setting.key}:source`])}
                    onChange={(next) => patchSetting(setting, { fetch_source: next }, 'source')}
                  />
                )
              },
              {
                title: '优先级',
                dataIndex: 'priority',
                width: 110
              },
              {
                title: '更新时间',
                dataIndex: 'updated_at',
                width: 190,
                render: (value?: string) => formatTime(value)
              },
              {
                title: '操作',
                key: 'actions',
                fixed: 'right',
                width: 120,
                render: (_, setting) => (
                  <Button
                    icon={<SaveOutlined />}
                    loading={Boolean(pending[`${setting.key}:run`])}
                    onClick={() => runNow(setting)}
                  >
                    立即获取
                  </Button>
                )
              }
            ]}
          />
        </Card>

        <Card className="page-surface-card" title="最近运行记录">
          <Table
            size="middle"
            rowKey="id"
            dataSource={runs}
            pagination={{ pageSize: 8, showSizeChanger: false }}
            columns={[
              { title: '集群', dataIndex: 'cluster_id', width: 90 },
              {
                title: '模块',
                dataIndex: 'module',
                render: (module: string) => moduleLabels[module] || module
              },
              {
                title: '状态',
                dataIndex: 'status',
                render: (status: string) => <Tag color={status === 'success' ? 'success' : status === 'running' ? 'processing' : 'error'}>{status}</Tag>
              },
              { title: '来源', dataIndex: 'source' },
              { title: '记录数', dataIndex: 'records_upserted' },
              { title: '耗时（ms）', dataIndex: 'duration_ms' },
              { title: '开始时间', dataIndex: 'started_at', render: (value: string) => formatTime(value) },
              {
                title: '错误',
                dataIndex: 'error',
                render: (value: string) => value ? <Text type="danger">{value}</Text> : <Text type="secondary">-</Text>
              }
            ]}
          />
        </Card>
      </Space>
    </Page>
  )
}

function formatTime(value?: string) {
  return value ? new Date(value).toLocaleString() : '-'
}

function toDataFetchConfigRow(setting: SystemSetting): DataFetchConfigRow | null {
  try {
    const value = JSON.parse(setting.value) as DataFetchConfig
    return {
      ...value,
      key: setting.key,
      updated_at: setting.updated_at
    }
  } catch {
    return null
  }
}

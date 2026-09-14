import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Descriptions, Drawer, Form, Input, Modal, Select, Space, Tabs, Tag, Typography } from 'antd'
import { useCallback, useEffect, useRef, useState } from 'react'
import { jsonInit, request, textValue, type ApiRecord } from '../../api/client'
import { listAllResources, mutateResource, refreshResource } from '../../api/resource'
import { AppTable } from '../../components/AppTable'
import { DraggableModal } from '../../components/DraggableModal'
import { RecordDetail } from '../../components/RecordDetail'
import { ResourceMetaBar } from '../../components/ResourceMetaBar'
import { Page } from '../../components/Page'
import { TableAction, TableActions } from '../../components/TableActions'
import { useResource } from '../../hooks'
import { useClusterContext } from '../../state/ClusterContext'
import { message } from '../../utils/appMessage'

interface ConfigurationForm { who: string; name: string; value: string }

export function ConfigurationPage({ moduleName }: { moduleName?: string } = {}) {
  const { selectedClusterId } = useClusterContext()
  const [search, setSearch] = useState('')
  const [busy, setBusy] = useState(false)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<ApiRecord | null>(null)
  const [detail, setDetail] = useState<ApiRecord | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailError, setDetailError] = useState('')
  const [form] = Form.useForm<ConfigurationForm>()
  const [helpError, setHelpError] = useState('')
  const [help, setHelp] = useState<ApiRecord | null>(null)
  const detailsRequest = useRef(0)
  const watchedName = Form.useWatch('name', form)
  const loader = useCallback(async () => {
    if (!selectedClusterId) return { options: [] as ApiRecord[], values: [] as ApiRecord[], stale: false, observedAt: null }
    const [options, values] = await Promise.all([listAllResources('/configuration/options', selectedClusterId), listAllResources('/configuration/values', selectedClusterId)])
    const prefix = moduleName ? `mgr/${moduleName}/` : ''
    return { options: options.items.filter((row) => String(row.name).startsWith(prefix)), values: values.items.filter((row) => String(row.name).startsWith(prefix)), stale: options.stale || values.stale, observedAt: values.observedAt }
  }, [selectedClusterId, moduleName])
  const { data, loading, error, refresh } = useResource(loader)
  useEffect(() => { setOpen(false); setDetailOpen(false); detailsRequest.current++ }, [selectedClusterId, moduleName])
  useEffect(() => {
    setHelp(null); setHelpError('')
    if (!open || !selectedClusterId || !watchedName) return
    const abort = new AbortController()
    void request<ApiRecord>('/configuration/option', jsonInit('GET', { cluster_id: selectedClusterId, name: watchedName }, { signal: abort.signal, suppressErrorNotification: true }))
      .then((value) => { if (!abort.signal.aborted) setHelp(value) }).catch((err) => { if (!abort.signal.aborted) setHelpError(err instanceof Error ? err.message : '元数据读取失败') })
    return () => abort.abort()
  }, [open, selectedClusterId, watchedName])
  async function collect() {
    if (!selectedClusterId) return
    await refreshResource({ clusterId: selectedClusterId, kinds: ['config_value', 'config_option'] })
    await refresh()
  }
  async function run(work: () => Promise<void>) {
    if (busy || !selectedClusterId) return
    setBusy(true)
    try { await work() } finally { setBusy(false) }
  }
  function edit(row?: ApiRecord) {
    setEditing(row?.who ? row : null)
    form.setFieldsValue({ name: textValue(row?.name, ''), who: textValue(row?.who, moduleName ? 'mgr' : 'global'), value: row?.value == null || row.value === '[REDACTED]' ? '' : String(row.value) })
    setOpen(true)
  }
  async function showDetails(name: string) {
    if (!selectedClusterId) return
    const revision = ++detailsRequest.current
    setDetail(null); setDetailError(''); setDetailOpen(true); setDetailLoading(true)
    try {
      const value = await request<ApiRecord>('/configuration/option', jsonInit('GET', { cluster_id: selectedClusterId, name }))
      if (revision === detailsRequest.current) setDetail(value)
    } catch (err) { if (revision === detailsRequest.current) setDetailError(err instanceof Error ? err.message : '读取失败') }
    finally { if (revision === detailsRequest.current) setDetailLoading(false) }
  }
  async function save(values: ConfigurationForm) {
    await run(async () => {
      await mutateResource('/configuration/value', 'PUT', { cluster_id: selectedClusterId, ...values, value: values.value ?? '' }, editing ? { ifMatch: String(editing.resource_version) } : undefined)
      setOpen(false); message.success('集群配置已保存'); await collect()
    })
  }
  function remove(row: ApiRecord) {
    Modal.confirm({ title: `删除 ${String(row.who)} 的 ${String(row.name)} 覆盖值`, content: '删除后，该作用域将继承其他适用配置或使用默认值。', okText: '删除', okType: 'danger',
      onOk: () => run(async () => {
        await mutateResource('/configuration/value', 'DELETE', { cluster_id: selectedClusterId, who: row.who, name: row.name }, { ifMatch: String(row.resource_version) })
        message.success('配置覆盖值已删除'); await collect()
      }) })
  }
  const matches = (row: ApiRecord) => `${row.name} ${row.who ?? ''} ${row.value ?? ''}`.toLowerCase().includes(search.toLowerCase())
  const blocked = busy || loading || !selectedClusterId || Boolean(error)
  const valueTable = <AppTable<ApiRecord> size="small" rowKey="natural_key" dataSource={data?.values.filter(matches) ?? []} pagination={{ defaultPageSize: 20, showSizeChanger: true }} columns={[
    { title: '作用域', dataIndex: 'who' }, { title: '配置名', dataIndex: 'name' },
    { title: '值', dataIndex: 'value', render: (value) => <Typography.Text style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{String(value ?? '')}</Typography.Text> },
    { title: '级别', dataIndex: 'level' },
    { title: '操作', render: (_, row) => <TableActions>
      <TableAction onClick={() => showDetails(String(row.name))}>说明</TableAction>
      <TableAction disabled={blocked || Boolean(row.stale)} onClick={() => edit(row)}>编辑</TableAction>
      <TableAction disabled={blocked || Boolean(row.stale)} onClick={() => remove(row)}>删除覆盖</TableAction>
    </TableActions> }
  ]} />
  return <Page title={moduleName ? `${moduleName} 模块配置` : "集群配置"} loading={loading} error={error}>
    <Card extra={<Space wrap><Button icon={<ReloadOutlined />} loading={busy} disabled={!selectedClusterId} onClick={() => run(collect)}>从集群刷新</Button><Button type="primary" icon={<PlusOutlined />} disabled={blocked} onClick={() => edit()}>设置配置</Button></Space>}>
      <ResourceMetaBar observedAt={data?.observedAt} stale={data?.stale} />
      {data?.stale && <Alert type="warning" message="部分配置数据已过期，请刷新后确认。" />}
      <Input.Search allowClear placeholder="搜索配置名、作用域或已设置的值" value={search} onChange={(event) => setSearch(event.target.value)} />
      <Tabs items={[
        { key: 'values', label: `已设置的配置（${data?.values.length ?? 0}）`, children: valueTable },
        { key: 'options', label: `全部选项（${data?.options.length ?? 0}）`, children: <AppTable<ApiRecord> size="small" rowKey="name" dataSource={data?.options.filter(matches) ?? []} pagination={{ defaultPageSize: 20, showSizeChanger: true }} columns={[
          { title: '配置选项', dataIndex: 'name' }, { title: '已配置作用域', render: (_, row) => data?.values.filter((value) => value.name === row.name).map((value) => <Tag key={String(value.natural_key)}>{String(value.who)}</Tag>) },
          { title: '操作', render: (_, row) => <TableActions><TableAction onClick={() => showDetails(String(row.name))}>详情</TableAction><TableAction disabled={blocked} onClick={() => edit(row)}>设置</TableAction></TableActions> }
        ]} /> }
      ]} />
    </Card>
    <Drawer title="配置选项详情" open={detailOpen} width="min(900px, 95vw)" onClose={() => { detailsRequest.current++; setDetailOpen(false) }}>
      {detailError && <Alert type="error" message={detailError} />}
      <Card loading={detailLoading}>{detail && <>
        <Descriptions size="small" column={1} bordered>
          <Descriptions.Item label="名称">{String(detail.name)}</Descriptions.Item>
          <Descriptions.Item label="类型 / 级别">{String(detail.type)} / {String(detail.level)}</Descriptions.Item>
          <Descriptions.Item label="说明">{String(detail.desc ?? '')}</Descriptions.Item>
          <Descriptions.Item label="详细说明">{String(detail.long_desc ?? '')}</Descriptions.Item>
          <Descriptions.Item label="默认值">{String(detail.default ?? '—')}</Descriptions.Item>
          <Descriptions.Item label="守护进程默认值">{String(detail.daemon_default ?? '—')}</Descriptions.Item>
          <Descriptions.Item label="运行时可更新">{detail.can_update_at_runtime ? '是' : '否'}</Descriptions.Item>
          <Descriptions.Item label="范围">{String(detail.min ?? '—')} ～ {String(detail.max ?? '—')}</Descriptions.Item>
        </Descriptions>
        <RecordDetail record={detail} />
      </>}</Card>
    </Drawer>
    <DraggableModal title={editing ? '编辑配置覆盖' : '设置配置覆盖'} open={open} confirmLoading={busy} onCancel={() => { if (!busy) setOpen(false) }} onOk={() => form.submit()}>
      <Form form={form} layout="vertical" onFinish={save}>
        <Form.Item name="name" label="配置选项" rules={[{ required: true }]}><Select disabled={Boolean(editing)} showSearch options={data?.options.map((row) => ({ label: String(row.name), value: String(row.name) }))} /></Form.Item>
        <Form.Item name="who" label="作用域" extra="例如 global、osd、client.rgw、osd/host:node-1 或 osd/class:ssd。" rules={[{ required: true }]}><Input disabled={Boolean(editing)} /></Form.Item>
        {helpError && <Alert type="warning" message={`无法读取配置说明：${helpError}`} />}
        {help && <Alert type={help.can_update_at_runtime ? 'info' : 'warning'} message={String(help.desc ?? '')} description={help.can_update_at_runtime ? `默认值：${String(help.default ?? '')}` : '该选项不能在运行时更新，保存后可能需要重启相关守护进程才能生效。'} />}
        <Form.Item name="value" label="配置值" rules={editing?.value === '[REDACTED]' ? [{ required: true, message: '请输入新的配置值；原值已隐藏' }] : []} extra="空字符串会写入空值；删除覆盖请使用列表中的删除操作。">
          {Array.isArray(help?.enum_values) && help.enum_values.length ? <Select options={help.enum_values.map((value) => ({ label: String(value), value: String(value) }))} /> : help?.type === 'bool' ? <Select options={[{ label: 'true', value: 'true' }, { label: 'false', value: 'false' }]} /> : <Input.TextArea rows={3} maxLength={32768} />}
        </Form.Item>
      </Form>
    </DraggableModal>
  </Page>
}

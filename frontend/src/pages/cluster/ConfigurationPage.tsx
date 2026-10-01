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
import { configurationValueError } from './configurationValue'

interface ConfigurationForm { who: string; name: string; value: string; instance?: string }
type OverrideFilter = 'all' | 'configured' | 'unconfigured'

function localizedConfigurationTarget(name: unknown, instance: unknown): string | undefined {
  if (typeof name !== 'string' || !name) return undefined
  if (instance === undefined || instance === '') return name
  if (typeof instance !== 'string' || !/^[A-Za-z0-9_][A-Za-z0-9_.-]{0,255}$/.test(instance)) return undefined
  const match = /^(mgr\/[A-Za-z][A-Za-z0-9_]*\/)([A-Za-z][A-Za-z0-9_]{0,255})$/.exec(name)
  return match ? `${match[1]}${instance}/${match[2]}` : undefined
}

function currentConfigurationHelp<T extends { scope: unknown, name: string, help: ApiRecord | null, error: string }>(snapshot: T | null, scope: unknown, name: unknown, open: boolean): T | null {
  if (!open || !snapshot || snapshot.scope !== scope || snapshot.name !== name) return null
  if (snapshot.help && snapshot.help.name !== name) return null
  return snapshot
}

function configurationMonWriteBlocked(help: ApiRecord | null, name: unknown): boolean {
  return typeof name === 'string' && help?.name === name && Array.isArray(help.flags) && help.flags.includes('no_mon_update')
}

function configurationHelpDescription(help: ApiRecord): string {
  if (configurationMonWriteBlocked(help, help.name)) return '该选项标记为 no_mon_update，不能通过 Monitor 配置库写入；请按该选项的部署配置方式管理。'
  if (help.can_update_at_runtime === false) return '该选项不能在运行时更新，保存后可能需要重启相关守护进程才能生效。'
  if (help.can_update_at_runtime !== true) return '运行时更新能力未采集，不能据此判断是否需要重启。'
  return `默认值：${String(help.default ?? '未采集')}`
}

function configurationWriteBlocked(row: ApiRecord): string | undefined {
  if (row.stale !== false) return '配置库存已过期或新鲜度未知，请先刷新'
  if (typeof row.who !== 'string' || !row.who.trim() || typeof row.name !== 'string' || !row.name.trim()) return '配置作用域或名称未采集'
  const version = typeof row.resource_version === 'number' || typeof row.resource_version === 'string' ? Number(row.resource_version) : NaN
  if (!Number.isSafeInteger(version) || version <= 0) return '配置资源版本无效，请先刷新'
  return undefined
}

async function configurationMetadataBatch(names: string[], read: (name: string) => Promise<ApiRecord>, active: () => boolean) {
  const pending = [...new Set(names)]
  const items: Record<string, ApiRecord> = Object.create(null)
  const failed: string[] = []
  await Promise.all(Array.from({ length: Math.min(4, pending.length) }, async () => {
    while (active() && pending.length) {
      const name = pending.shift()!
      try {
        const result = await read(name)
        if (result.name !== name) throw new Error('配置元数据名称不匹配')
        items[name] = result
      } catch { failed.push(name) }
    }
  }))
  return { items, failed }
}

function filterConfigurationOptions(options: ApiRecord[], values: ApiRecord[], filter: OverrideFilter, fresh: boolean): ApiRecord[] {
  if (filter === 'all') return options
  if (!fresh) return []
  const configured = new Set(values.map((row) => row.name))
  return options.filter((row) => configured.has(row.name) === (filter === 'configured'))
}

function configurationOverrides(rows: ApiRecord[], name: unknown): ApiRecord[] {
  return typeof name === 'string' && name.length > 0 ? rows.filter((row) => row.name === name) : []
}

function configurationList(value: unknown): string {
  if (!Array.isArray(value) || !value.every((item) => typeof item === 'string')) return '未采集'
  return value.length ? value.join('、') : '无'
}

function configurationRuntime(value: unknown): string {
  return value === true ? '是' : value === false ? '否' : '未采集'
}

export function ConfigurationPage({ moduleName }: { moduleName?: string } = {}) {
  const { selectedClusterId } = useClusterContext()
  const scopeRef = useRef({ clusterId: selectedClusterId, moduleName })
  if (scopeRef.current.clusterId !== selectedClusterId || scopeRef.current.moduleName !== moduleName) {
    scopeRef.current = { clusterId: selectedClusterId, moduleName }
  }
  const scope = scopeRef.current
  const running = useRef(false)
  const metadataRequest = useRef<AbortController | null>(null)
  const [metadata, setMetadata] = useState<{ scope: typeof scope, items: Record<string, ApiRecord>, failed: string[] } | null>(null)
  const [metadataLoading, setMetadataLoading] = useState(false)
  const [optionPage, setOptionPage] = useState(1)
  const [optionPageSize, setOptionPageSize] = useState(20)
  const [search, setSearch] = useState('')
  const [overrideFilter, setOverrideFilter] = useState<OverrideFilter>('all')
  const [busy, setBusy] = useState(false)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<ApiRecord | null>(null)
  const [detail, setDetail] = useState<ApiRecord | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailError, setDetailError] = useState('')
  const [form] = Form.useForm<ConfigurationForm>()
  const [helpSnapshot, setHelpSnapshot] = useState<{ scope: typeof scope, name: string, help: ApiRecord | null, error: string } | null>(null)
  const detailsRequest = useRef(0)
  const watchedName = Form.useWatch('name', form)
  const watchedInstance = Form.useWatch('instance', form)
  const localizedTarget = localizedConfigurationTarget(watchedName, watchedInstance)
  const currentHelp = currentConfigurationHelp(helpSnapshot, scope, watchedName, open)
  const help = currentHelp?.help ?? null
  const helpError = currentHelp?.error ?? ''
  const loader = useCallback(async () => {
    if (!selectedClusterId) return { options: [] as ApiRecord[], values: [] as ApiRecord[], stale: false, observedAt: null }
    const [options, values] = await Promise.all([listAllResources('/configuration/options', selectedClusterId), listAllResources('/configuration/values', selectedClusterId)])
    const prefix = moduleName ? `mgr/${moduleName}/` : ''
    return { options: options.items.filter((row) => String(row.name).startsWith(prefix)), values: values.items.filter((row) => String(row.name).startsWith(prefix)), stale: options.stale || values.stale, observedAt: values.observedAt }
  }, [selectedClusterId, moduleName])
  const { data, loading, error, refresh } = useResource(loader)
  useEffect(() => { setOpen(false); setDetailOpen(false); detailsRequest.current++ }, [selectedClusterId, moduleName])
  useEffect(() => {
    setHelpSnapshot(null)
    if (!open || !selectedClusterId || !watchedName) return
    const abort = new AbortController()
    void request<ApiRecord>('/configuration/option', jsonInit('GET', { cluster_id: selectedClusterId, name: watchedName }, { signal: abort.signal, suppressErrorNotification: true }))
      .then((value) => {
        if (value.name !== watchedName) throw new Error('配置说明与请求的选项不一致')
        if (!abort.signal.aborted && scopeRef.current === scope) setHelpSnapshot({ scope, name: watchedName, help: value, error: '' })
      }).catch((err) => {
        if (!abort.signal.aborted && scopeRef.current === scope) setHelpSnapshot({ scope, name: watchedName, help: null, error: err instanceof Error ? err.message : '元数据读取失败' })
      })
    return () => abort.abort()
  }, [open, selectedClusterId, watchedName, scope])
  async function collect() {
    if (!selectedClusterId || scopeRef.current !== scope) return
    await refreshResource({ clusterId: selectedClusterId, kinds: ['config_value', 'config_option'] })
    if (scopeRef.current !== scope) return
    await refresh()
  }
  async function refreshAfterMutation() {
    if (scopeRef.current !== scope) return
    try {
      await refreshResource({ clusterId: selectedClusterId, kinds: ['config_value', 'config_option'] })
    } catch {
      if (scopeRef.current === scope) message.warning('配置修改已执行，但重新采集失败；请刷新并核对结果，不要重复提交修改。')
    }
    if (scopeRef.current === scope) await refresh()
  }
  async function run(work: () => Promise<void>) {
    if (running.current || !selectedClusterId || scopeRef.current !== scope) return
    running.current = true
    setBusy(true)
    try { await work() } finally { running.current = false; setBusy(false) }
  }
  function edit(row?: ApiRecord, existing = false) {
    if (loading || error) return
    if (existing && row) {
      const reason = configurationWriteBlocked(row)
      if (reason) { message.error(reason); return }
    }
    setEditing(row?.who ? row : null)
    form.setFieldsValue({ name: textValue(row?.name, ''), instance: '', who: textValue(row?.who, moduleName ? 'mgr' : 'global'), value: row?.value == null || row.value === '[REDACTED]' ? '' : String(row.value) })
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
    if (loading || error) { message.error('请先成功读取当前配置库存'); return }
    if (editing) {
      const reason = configurationWriteBlocked(editing)
      if (reason) { message.error(reason); return }
      if (values.who !== editing.who || values.name !== editing.name) { message.error('编辑目标已改变，请重新打开表单'); return }
    }
    if (configurationMonWriteBlocked(help, values.name)) { message.error('该选项不允许通过 Monitor 配置库修改'); return }
    const target = localizedConfigurationTarget(values.name, values.instance)
    if (!target || (editing && target !== editing.name)) { message.error('实例配置目标无效，请核对配置名和实例名'); return }
    await run(async () => {
      await mutateResource('/configuration/value', 'PUT', { cluster_id: selectedClusterId, who: values.who, name: target, value: values.value ?? '' }, editing ? { ifMatch: String(editing.resource_version) } : undefined)
      if (scopeRef.current !== scope) return
      setOpen(false); message.success('集群配置已保存'); await refreshAfterMutation()
    })
  }
  function remove(row: ApiRecord) {
    if (loading || error) return
    const reason = configurationWriteBlocked(row)
    if (reason) { message.error(reason); return }
    Modal.confirm({ title: `删除 ${String(row.who)} 的 ${String(row.name)} 覆盖值`, content: '删除后，该作用域将继承其他适用配置或使用默认值。', okText: '删除', okType: 'danger',
      onOk: () => {
        if (scopeRef.current !== scope) throw new Error('集群或模块已切换，请重新确认配置删除')
        return run(async () => {
          await mutateResource('/configuration/value', 'DELETE', { cluster_id: selectedClusterId, who: row.who, name: row.name }, { ifMatch: String(row.resource_version) })
          if (scopeRef.current !== scope) return
          message.success('配置覆盖值已删除'); await refreshAfterMutation()
        })
      } })
  }
  const matches = (row: ApiRecord) => `${row.name} ${row.who ?? ''} ${row.value ?? ''}`.toLowerCase().includes(search.toLowerCase())
  const blocked = busy || loading || !selectedClusterId || Boolean(error)
  const overridesFresh = !loading && !error && data?.stale === false && data.values.every((row) => row.stale === false)
  const filteredOptions = filterConfigurationOptions(data?.options ?? [], data?.values ?? [], overrideFilter, overridesFresh).filter(matches)
  const currentPage = Math.min(optionPage, Math.max(1, Math.ceil(filteredOptions.length / optionPageSize)))
  const optionMetadata = metadata?.scope === scope ? metadata.items : {}
  useEffect(() => () => { metadataRequest.current?.abort() }, [scope])
  async function loadPageMetadata() {
    if (!selectedClusterId || loading || error || metadataLoading) return
    metadataRequest.current?.abort()
    const controller = new AbortController()
    metadataRequest.current = controller
    setMetadataLoading(true)
    try {
      const names = filteredOptions.slice((currentPage - 1) * optionPageSize, currentPage * optionPageSize).map((row) => String(row.name))
      const result = await configurationMetadataBatch(names, (name) => request<ApiRecord>('/configuration/option', jsonInit('GET', { cluster_id: selectedClusterId, name }, { signal: controller.signal, suppressErrorNotification: true })), () => scopeRef.current === scope && !controller.signal.aborted)
      if (scopeRef.current !== scope || controller.signal.aborted) return
      setMetadata((previous) => {
        const items = { ...(previous?.scope === scope ? previous.items : {}) }
        names.forEach((name) => { delete items[name] })
        return { scope, items: { ...items, ...result.items }, failed: result.failed }
      })
    } finally {
      if (metadataRequest.current === controller) setMetadataLoading(false)
    }
  }
  const valueTable = <AppTable<ApiRecord> size="small" rowKey="natural_key" dataSource={data?.values.filter(matches) ?? []} pagination={{ defaultPageSize: 20, showSizeChanger: true }} columns={[
    { title: '作用域', dataIndex: 'who' }, { title: '配置名', dataIndex: 'name' },
    { title: '值', dataIndex: 'value', render: (value) => <Typography.Text style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{String(value ?? '')}</Typography.Text> },
    { title: '级别', dataIndex: 'level' },
    { title: '操作', render: (_, row) => <TableActions>
      <TableAction onClick={() => showDetails(String(row.name))}>说明</TableAction>
      <TableAction disabled={blocked || Boolean(configurationWriteBlocked(row))} onClick={() => edit(row, true)}>编辑</TableAction>
      <TableAction disabled={blocked || Boolean(configurationWriteBlocked(row))} onClick={() => remove(row)}>删除覆盖</TableAction>
    </TableActions> }
  ]} />
  return <Page title={moduleName ? `${moduleName} 模块配置` : "集群配置"} loading={loading} error={error}>
    <Card extra={<Space wrap><Button icon={<ReloadOutlined />} loading={busy} disabled={!selectedClusterId} onClick={() => run(collect)}>从集群刷新</Button><Button type="primary" icon={<PlusOutlined />} disabled={blocked} onClick={() => edit()}>设置配置</Button></Space>}>
      <ResourceMetaBar observedAt={data?.observedAt} stale={data?.stale} />
      {data?.stale && <Alert type="warning" message="部分配置数据已过期，请刷新后确认。" />}
      <Input.Search allowClear placeholder="搜索配置名、作用域或已设置的值" value={search} onChange={(event) => setSearch(event.target.value)} />
      <Tabs items={[
        { key: 'values', label: `已设置的配置（${data?.values.length ?? 0}）`, children: valueTable },
        { key: 'options', label: `全部选项（${data?.options.length ?? 0}）`, children: <Space direction="vertical" className="full-width-control">
          <Space wrap><Typography.Text>显式覆盖</Typography.Text><Select aria-label="配置覆盖状态" value={overrideFilter} onChange={setOverrideFilter} style={{ minWidth: 190 }} options={[
            { label: '全部选项', value: 'all' },
            { label: '已设置覆盖（mon）', value: 'configured', disabled: !overridesFresh },
            { label: '未设置显式覆盖', value: 'unconfigured', disabled: !overridesFresh }
          ]} /><Typography.Text type="secondary">筛选结果 {filteredOptions.length} 项；未设置覆盖不代表没有有效配置。</Typography.Text></Space>
          {overrideFilter !== 'all' && !overridesFresh && <Alert type="warning" showIcon message="配置库存不可用或已过期，暂不能判定覆盖状态，请刷新或选择全部选项。" />}
          <Space wrap><Button onClick={loadPageMetadata} loading={metadataLoading} disabled={loading || Boolean(error) || !selectedClusterId || !filteredOptions.length}>加载本页说明</Button><Typography.Text type="secondary">按需读取 ceph config help，最多 4 个并发；显示最近一次手动读取的元数据。</Typography.Text></Space>
          {metadata?.scope === scope && metadata.failed.length > 0 && <Alert type="warning" showIcon message={`${metadata.failed.length} 个选项说明读取失败，可重新加载本页重试。`} />}
          <AppTable<ApiRecord> size="small" rowKey="name" dataSource={filteredOptions} pagination={{ current: currentPage, pageSize: optionPageSize, pageSizeOptions: [10, 20, 50], showSizeChanger: true, onChange: (page, size) => { setOptionPage(page); setOptionPageSize(size) } }} columns={[
          { title: '配置选项', dataIndex: 'name' }, { title: '已配置作用域', render: (_, row) => data?.values.filter((value) => value.name === row.name).map((value) => <Tag key={String(value.natural_key)}>{String(value.who)}</Tag>) },
          ...['desc', 'type', 'level', 'default'].map((key, index) => ({ title: ['说明', '类型', '级别', '默认值'][index], key, render: (_: unknown, row: ApiRecord) => String(optionMetadata[String(row.name)]?.[key] ?? '未读取') })),
          { title: '运行时可更新', render: (_, row) => configurationRuntime(optionMetadata[String(row.name)]?.can_update_at_runtime) },
          { title: '操作', render: (_, row) => <TableActions><TableAction onClick={() => showDetails(String(row.name))}>详情</TableAction><TableAction disabled={blocked} onClick={() => edit(row)}>设置</TableAction></TableActions> }
        ]} /></Space> }
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
          <Descriptions.Item label="运行时可更新">{configurationRuntime(detail.can_update_at_runtime)}</Descriptions.Item>
          <Descriptions.Item label="范围">{String(detail.min ?? '—')} ～ {String(detail.max ?? '—')}</Descriptions.Item>
          <Descriptions.Item label="标志">{configurationList(detail.flags)}</Descriptions.Item>
          <Descriptions.Item label="服务">{configurationList(detail.services)}</Descriptions.Item>
          <Descriptions.Item label="标签">{configurationList(detail.tags)}</Descriptions.Item>
          <Descriptions.Item label="可选值">{configurationList(detail.enum_values)}</Descriptions.Item>
          <Descriptions.Item label="相关选项">{configurationList(detail.see_also)}</Descriptions.Item>
        </Descriptions>
        <Typography.Title level={5}>已设置的作用域覆盖值</Typography.Title>
        <Typography.Paragraph type="secondary">来源：ceph config dump 库存；元数据来源：ceph config help。覆盖值不等同于某个守护进程最终生效的值。</Typography.Paragraph>
        <ResourceMetaBar observedAt={data?.observedAt} stale={data?.stale} />
        {(loading || error || !data || data.stale) && <Alert type="warning" showIcon message="配置库存正在读取、读取失败或已过期，不能据此确认当前覆盖值。" />}
        <AppTable<ApiRecord>
          size="small" rowKey="natural_key"
          dataSource={configurationOverrides(data?.values ?? [], detail.name)}
          pagination={{ defaultPageSize: 10, showSizeChanger: true }}
          locale={{ emptyText: loading || error || !data || data.stale ? '覆盖值未确认' : '本次采集未发现该选项的显式覆盖；请结合默认值与继承规则判断' }}
          columns={[
            { title: '作用域（含位置限制）', dataIndex: 'who' },
            { title: '值', dataIndex: 'value', render: (value) => <Typography.Text style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{value === '' ? '空字符串' : String(value ?? '未采集')}</Typography.Text> },
            { title: '级别', dataIndex: 'level' }
          ]}
        />
        <RecordDetail record={detail} />
      </>}</Card>
    </Drawer>
    <DraggableModal title={editing ? '编辑配置覆盖' : '设置配置覆盖'} open={open} confirmLoading={busy} okButtonProps={{ disabled: configurationMonWriteBlocked(help, watchedName) }} onCancel={() => { if (!busy) setOpen(false) }} onOk={() => form.submit()}>
      <Form form={form} layout="vertical" onFinish={save}>
        <Form.Item name="name" label="配置选项" rules={[{ required: true }]}><Select disabled={Boolean(editing)} showSearch onChange={() => form.setFieldValue('instance', '')} options={data?.options.map((row) => ({ label: String(row.name), value: String(row.name) }))} /></Form.Item>
        {!editing && typeof watchedName === 'string' && /^mgr\/[^/]+\/[^/]+$/.test(watchedName) && <Form.Item name="instance" label="MGR 实例名（可选）" preserve={false} extra="留空设置模块级参数；填写实例 ID 创建本地化键。仅使用 localized option 的模块会读取该键，不等同于下方 mgr.<实例> 作用域。" rules={[{ pattern: /^[A-Za-z0-9_][A-Za-z0-9_.-]{0,255}$/, message: '实例名只能含字母、数字、下划线、点和连字符，且不能以点或连字符开头' }]}><Input maxLength={256} placeholder="例如 node-a.x1（不含 mgr. 前缀）" /></Form.Item>}
        <Typography.Paragraph type="secondary">完整配置键：{localizedTarget ?? '请先选择配置并填写有效实例名'}</Typography.Paragraph>
        <Form.Item name="who" label="作用域" extra="例如 global、osd、client.rgw、osd/host:node-1 或 osd/class:ssd。" rules={[{ required: true }]}><Input disabled={Boolean(editing)} /></Form.Item>
        {helpError && <Alert type="warning" message={`无法读取配置说明：${helpError}`} />}
        {help && help.name === watchedName && <Alert type={help.can_update_at_runtime === true && !configurationMonWriteBlocked(help, watchedName) ? 'info' : 'warning'} message={String(help.desc ?? '')} description={configurationHelpDescription(help)} />}
        {help && <Typography.Paragraph type="secondary">类型：{String(help.type ?? '未采集')}；范围：{String(help.min ?? '未采集')} ～ {String(help.max ?? '未采集')}。整数单位 K/M/G 按十进制，size 容量单位按 1024 换算；最终约束由 Ceph 校验。</Typography.Paragraph>}
        <Form.Item name="value" label="配置值" dependencies={['name']} rules={[
          ...(editing?.value === '[REDACTED]' ? [{ required: true, message: '请输入新的配置值；原值已隐藏' }] : []),
          { validator: (_, value) => { const reason = configurationValueError(help, form.getFieldValue('name'), value ?? ''); return reason ? Promise.reject(new Error(reason)) : Promise.resolve() } }
        ]} extra="空字符串会写入空值（数值类型不允许）；删除覆盖请使用列表中的删除操作。">
          {Array.isArray(help?.enum_values) && help.enum_values.length ? <Select options={help.enum_values.map((value) => ({ label: String(value), value: String(value) }))} /> : help?.type === 'bool' ? <Select options={[{ label: 'true', value: 'true' }, { label: 'false', value: 'false' }]} /> : <Input.TextArea rows={3} maxLength={32768} />}
        </Form.Item>
      </Form>
    </DraggableModal>
  </Page>
}

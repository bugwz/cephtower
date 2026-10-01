import { Alert, Button, Card, Form, Input, Modal, Popconfirm, Select, Space } from 'antd'
import { useCallback, useEffect, useRef, useState } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'
import { listAllResources, listResource, mutateResource, refreshResource, type ResourceListResult } from '../../api/resource'
import { scheduleFormScope, scheduleScope } from './snapshotScheduleScope'
import { buildRetentionRules, buildScheduleInterval, retentionFrequencyOptions, retentionRuleCounts, scheduleActiveText, scheduleFrequencyOptions, scheduleIntervalText, scheduleRetentionText, scheduleToggleAction } from './snapshotScheduleText'
import { AppTable } from '../../components/AppTable'
import { RecordDetail } from '../../components/RecordDetail'
import { ResourceMetaBar } from '../../components/ResourceMetaBar'
import { useClusterContext } from '../../state/ClusterContext'

export function SnapshotScheduleStatus() {
  const { selectedClusterId } = useClusterContext()
  return <ClusterSnapshotScheduleStatus key={selectedClusterId ?? 'none'} selectedClusterId={selectedClusterId} />
}

function ClusterSnapshotScheduleStatus({ selectedClusterId }: { selectedClusterId?: number }) {
  const active = useRef(true)
  useEffect(() => {
    active.current = true
    return () => { active.current = false }
  }, [])
  const [form] = Form.useForm()
  const [createForm] = Form.useForm()
  const [creating, setCreating] = useState(false)
  const [loading, setLoading] = useState(false)
  const [rows, setRows] = useState<ApiRecord[] | null>(null)
  const scope = useRef<ApiRecord>({})
  const [retentionCounts, setRetentionCounts] = useState<Record<string, string>>({})
  const retention = buildRetentionRules(retentionCounts)
  const currentRetentionCounts = retentionRuleCounts(rows?.[0]?.retention)
  const [mutating, setMutating] = useState(false)
  const [error, setError] = useState('')
  const [moduleState, setModuleState] = useState<'loading' | 'enabled' | 'disabled' | 'unavailable'>('loading')
  const [moduleError, setModuleError] = useState('')
  const [discoveredRows, setDiscoveredRows] = useState<ApiRecord[]>([])
  const [discoveredMeta, setDiscoveredMeta] = useState<Pick<ResourceListResult, 'observedAt' | 'stale' | 'staleReason'>>()
  const [discoveredLoading, setDiscoveredLoading] = useState(false)
  const [discoveredError, setDiscoveredError] = useState('')
  const pending = useRef<AbortController | null>(null)
  function reset() { pending.current?.abort(); setRows(null); setError(''); setLoading(false); setRetentionCounts({}) }
  const loadModule = useCallback(async () => {
    if (!selectedClusterId) {
      setModuleState('unavailable')
      setModuleError('请先选择集群')
      return
    }
    setModuleState('loading')
    setModuleError('')
    try {
      const result = await listResource('/manager/modules', selectedClusterId, { filters: { name: ['snap_schedule'] } })
      if (!active.current) return
      const module = result.items.find((item) => item.name === 'snap_schedule')
      if (!module) {
        setModuleState('unavailable')
        setModuleError('集群未返回 snap_schedule 模块信息')
      } else if (module.can_run === false) {
        setModuleState('unavailable')
        setModuleError(String(module.error_string ?? 'snap_schedule 模块当前无法运行'))
      } else {
        setModuleState(module.enabled === true ? 'enabled' : 'disabled')
      }
    } catch (err) {
      if (!active.current) return
      setModuleState('unavailable')
      setModuleError(err instanceof Error ? err.message : '读取模块状态失败')
    }
  }, [selectedClusterId])
  const loadDiscovered = useCallback(async (refresh = false) => {
    if (!active.current) return
    if (!selectedClusterId) {
      setDiscoveredRows([])
      return
    }
    setDiscoveredLoading(true)
    setDiscoveredError('')
    try {
      if (refresh) await refreshResource({ clusterId: selectedClusterId, kind: 'snapshot_schedule' })
      if (!active.current) return
      const result = await listAllResources('/filesystem/snapshot/schedules', selectedClusterId)
      if (!active.current) return
      setDiscoveredRows(result.items)
      setDiscoveredMeta({ observedAt: result.observedAt, stale: result.stale, staleReason: result.staleReason })
    } catch (err) {
      if (!active.current) return
      setDiscoveredError(err instanceof Error ? err.message : '读取快照计划列表失败')
    } finally {
      if (active.current) setDiscoveredLoading(false)
    }
  }, [selectedClusterId])
  useEffect(() => {
    reset()
    setCreating(false)
    void loadModule()
    return () => pending.current?.abort()
  }, [loadModule, selectedClusterId])
  useEffect(() => {
    if (moduleState === 'enabled') void loadDiscovered()
    if (moduleState === 'disabled' || moduleState === 'unavailable') setDiscoveredRows([])
  }, [loadDiscovered, moduleState])
  async function enableModule() {
    if (!active.current || !selectedClusterId || mutating || moduleState !== 'disabled') return
    setMutating(true)
    try {
      await mutateResource('/manager/module', 'PATCH', { cluster_id:selectedClusterId, name:'snap_schedule', enabled:true })
      if (!active.current) return
      await refreshResource({ clusterId:selectedClusterId, kind:'mgr_module' })
      if (!active.current) return
      await loadModule()
    } finally { setMutating(false) }
  }
  async function query(values: ApiRecord) {
    if (!active.current || !selectedClusterId || loading) return
    scope.current = values
    pending.current?.abort()
    const abort = new AbortController(); pending.current = abort
    setLoading(true); setRows(null); setError('')
    try {
      const result = await request<{ items: ApiRecord[] }>('/filesystem/snapshot/schedule/status', jsonInit('GET', { cluster_id: selectedClusterId, ...values }, { signal: abort.signal, suppressErrorNotification: true }))
      if (!abort.signal.aborted) setRows(result.items)
    } catch (err) { if (!abort.signal.aborted) setError(err instanceof Error ? err.message : '查询失败') }
    finally { if (!abort.signal.aborted) setLoading(false) }
  }
  async function toggle(row: ApiRecord, action: string | undefined = scheduleToggleAction(row.active), target: ApiRecord = scope.current, refreshQuery = true) {
    if (!active.current || !selectedClusterId || mutating || !action) return
    setMutating(true)
    try {
      await mutateResource('/filesystem/snapshot/schedule/action','POST',{ cluster_id:selectedClusterId, ...target, schedule:row.schedule, start:row.start, action })
      await loadDiscovered()
      if (refreshQuery) await query(scope.current)
    } finally { setMutating(false) }
  }
  async function managePath(row: ApiRecord) {
    if (loading || mutating) return
    const target = scheduleScope(row)
    form.setFieldsValue(scheduleFormScope(row))
    setRetentionCounts({})
    await query(target)
  }
  async function changeRetention(action: 'add' | 'remove') {
    if (!active.current || !selectedClusterId || mutating || !retention) return
    setMutating(true)
    try {
      await mutateResource('/filesystem/snapshot/schedule/retention','POST',{ cluster_id:selectedClusterId, ...scope.current, retention, action })
      await loadDiscovered()
      await query(scope.current)
    } finally { setMutating(false) }
  }
  async function create(values: ApiRecord) {
    if (!active.current || !selectedClusterId || mutating) return
    const target = await form.validateFields()
    if (!active.current) return
    setMutating(true)
    try {
      await mutateResource('/filesystem/snapshot/schedule','POST',{ cluster_id:selectedClusterId, ...target, schedule: buildScheduleInterval(values.interval, values.frequency), ...(values.start ? { start: values.start } : {}) })
      if (!active.current) return
      setCreating(false)
      await loadDiscovered()
      await query(target)
    } finally { setMutating(false) }
  }
  return <Card title="查询路径的实时快照计划">
    {moduleState === 'loading' && <Alert type="info" showIcon message="正在检查 snap_schedule 模块状态" />}
    {moduleState === 'disabled' && <Alert type="warning" showIcon message="snap_schedule 模块尚未启用" description="启用后才能查询和管理 CephFS 快照计划。" action={<Button type="primary" loading={mutating} onClick={enableModule}>启用模块</Button>} />}
    {moduleState === 'unavailable' && <Alert type="error" showIcon message="快照计划模块不可用" description={moduleError} />}
    {moduleState === 'enabled' && <Card type="inner" title="全部已发现的快照计划" extra={<Button loading={discoveredLoading} onClick={() => loadDiscovered(true)}>刷新</Button>}>
      {discoveredError && <Alert type="error" message={discoveredError} />}
      {discoveredMeta?.stale && <Alert type="warning" showIcon message="快照计划列表已过期" description={discoveredMeta.staleReason || '请刷新列表，或进入路径管理查询实时状态。'} />}
      <ResourceMetaBar observedAt={discoveredMeta?.observedAt} stale={discoveredMeta?.stale} staleReason={discoveredMeta?.staleReason} />
      <AppTable<ApiRecord> loading={discoveredLoading} dataSource={discoveredRows} rowKey={(row) => String(row.natural_key)} expandable={{ expandedRowRender:(row) => <RecordDetail record={row} /> }} columns={[
        {title:'文件系统',dataIndex:'fs'},
        {title:'路径',dataIndex:'path'},
        {title:'子卷',dataIndex:'subvol',render:(value) => value || '—'},
        {title:'子卷组',dataIndex:'group',render:(value) => value || '默认组'},
        {title:'周期',dataIndex:'schedule',render:scheduleIntervalText},
        {title:'状态',dataIndex:'active',render:scheduleActiveText},
        {title:'开始时间（UTC）',dataIndex:'start'},
        {title:'保留策略',dataIndex:'retention',render:scheduleRetentionText},
        {title:'已创建',dataIndex:'created_count',render:(value) => value ?? '—'},
        {title:'已清理',dataIndex:'pruned_count',render:(value) => value ?? '—'},
        {title:'操作',render:(_,row) => <Space><Button disabled={mutating || loading} onClick={() => managePath(row)}>管理路径与保留策略</Button><Popconfirm title="删除这条快照计划？" description="已有快照不会因此删除。" onConfirm={() => toggle(row, 'remove', scheduleScope(row), false)}><Button danger disabled={mutating}>删除</Button></Popconfirm><Button disabled={mutating || !scheduleToggleAction(row.active)} onClick={() => toggle(row, scheduleToggleAction(row.active), scheduleScope(row), false)}>{row.active === true ? '停用' : '启用'}</Button></Space>}
      ]} />
    </Card>}
    <Form form={form} disabled={mutating} layout="inline" initialValues={{ path: '/' }} onFinish={query} onValuesChange={reset}>
      <Form.Item name="fs" label="文件系统" rules={[{ required: true }]}><Input /></Form.Item>
      <Form.Item name="path" label="路径" rules={[{ required: true }]}><Input /></Form.Item>
      <Form.Item name="subvol" label="子卷"><Input /></Form.Item>
      <Form.Item name="group" label="子卷组"><Input /></Form.Item>
      <Button htmlType="submit" loading={loading} disabled={!selectedClusterId || moduleState !== 'enabled'}>查询</Button>
    </Form>
    <Button disabled={!selectedClusterId || mutating || loading || moduleState !== 'enabled'} onClick={() => { void form.validateFields().then(() => setCreating(true)) }}>为当前路径新建计划</Button>
    <Modal title="新建快照计划" open={creating} confirmLoading={mutating} onCancel={() => { if (!mutating) setCreating(false) }} onOk={() => createForm.submit()}>
      <Form form={createForm} layout="vertical" onFinish={create} initialValues={{ interval: '1', frequency: 'd' }}>
        <Form.Item name="interval" label="每隔" rules={[{ required:true }, { pattern: /^[1-9][0-9]*$/, message: '请输入正整数' }]}><Input inputMode="numeric" /></Form.Item>
        <Form.Item name="frequency" label="周期单位" rules={[{ required:true }]}><Select options={scheduleFrequencyOptions} /></Form.Item>
        <Form.Item name="start" label="开始时间（可选）"><Input placeholder="2026-09-14T00:00:00" /></Form.Item>
      </Form>
    </Modal>
    {Boolean(rows?.length) && <Card title="路径保留策略">
      <Alert type="info" message="保留策略作用于当前路径的全部计划。n 为最近快照数量；m/h/d/w/M/y 为分钟、小时、日、周、月、年。" />
      <Space wrap>{retentionFrequencyOptions.map(({ value, label }) => <label key={value}>{label}保留数量<Input aria-label={`${label}保留数量`} inputMode="numeric" value={retentionCounts[value] ?? ''} onChange={(event) => setRetentionCounts((current) => ({ ...current, [value]: event.target.value }))} placeholder="留空不修改" disabled={mutating} status={retentionCounts[value] && !/^[1-9][0-9]*$/.test(retentionCounts[value]) ? 'error' : undefined} /></label>)}</Space>
      <p>填写正整数数量；留空的单位不参与本次操作。已存在的单位须先移除再添加新数量；移除时数量必须与当前规则一致。</p>
      <Button disabled={mutating || loading || !currentRetentionCounts || !Object.keys(currentRetentionCounts).length} onClick={() => setRetentionCounts(currentRetentionCounts ?? {})}>填入当前规则</Button>
      {retention && <p>{scheduleRetentionText(retention)}</p>}
      <Space><Button disabled={mutating || loading || !retention} onClick={() => changeRetention('add')}>添加保留规则</Button><Button disabled={mutating || loading || !retention} onClick={() => changeRetention('remove')}>移除保留规则</Button></Space>
    </Card>}
    {error && <Alert type="error" message={error} />}
    {rows?.length === 0 && <Alert type="info" message="该路径没有快照计划。查询不包含其他路径。" />}
    {rows && <AppTable<ApiRecord> dataSource={rows} rowKey={(row) => JSON.stringify([row.path,row.schedule,row.start])} expandable={{ expandedRowRender:(row) => <RecordDetail record={row} /> }} columns={[
      {title:'路径',dataIndex:'path'},
      {title:'周期',dataIndex:'schedule',render:scheduleIntervalText},
      {title:'保留策略',dataIndex:'retention',render:scheduleRetentionText},
      {title:'状态',dataIndex:'active',render:scheduleActiveText},
      {title:'开始时间（UTC）',dataIndex:'start'},
      {title:'首次快照（UTC）',dataIndex:'first',render:(value) => value ?? '—'},
      {title:'最近快照（UTC）',dataIndex:'last',render:(value) => value ?? '—'},
      {title:'已创建',dataIndex:'created_count',render:(value) => value ?? '—'},
      {title:'已清理',dataIndex:'pruned_count',render:(value) => value ?? '—'},
      {title:'最近清理（UTC）',dataIndex:'last_pruned',render:(value) => value ?? '—'},
      {title:'操作',render:(_,row) => <Space><Popconfirm title="删除这条快照计划？" description="按周期和开始时间定位；已有快照不会因此删除。" onConfirm={() => toggle(row, 'remove')}><Button danger disabled={mutating || loading}>删除计划</Button></Popconfirm><Button loading={mutating} disabled={loading || !scheduleToggleAction(row.active)} onClick={() => toggle(row)}>{row.active === true ? '停用' : '启用'}</Button></Space>}
    ]} />}

  </Card>
}

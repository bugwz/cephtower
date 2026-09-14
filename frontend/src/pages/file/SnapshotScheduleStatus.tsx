import { Alert, Button, Card, Form, Input, Modal, Popconfirm, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'
import { mutateResource } from '../../api/resource'
import { AppTable } from '../../components/AppTable'
import { RecordDetail } from '../../components/RecordDetail'
import { useClusterContext } from '../../state/ClusterContext'

export function SnapshotScheduleStatus() {
  const { selectedClusterId } = useClusterContext()
  const [form] = Form.useForm()
  const [createForm] = Form.useForm()
  const [creating, setCreating] = useState(false)
  const [loading, setLoading] = useState(false)
  const [rows, setRows] = useState<ApiRecord[] | null>(null)
  const scope = useRef<ApiRecord>({})
  const [retention, setRetention] = useState('')
  const [mutating, setMutating] = useState(false)
  const [error, setError] = useState('')
  const pending = useRef<AbortController | null>(null)
  function reset() { pending.current?.abort(); setRows(null); setError(''); setLoading(false) }
  useEffect(() => { reset(); setCreating(false); return () => pending.current?.abort() }, [selectedClusterId])
  async function query(values: ApiRecord) {
    if (!selectedClusterId || loading) return
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
  async function toggle(row: ApiRecord, action = row.active ? 'deactivate' : 'activate') {
    if (!selectedClusterId || mutating) return
    setMutating(true)
    try {
      await mutateResource('/filesystem/snapshot/schedule/action','POST',{ cluster_id:selectedClusterId, ...scope.current, schedule:row.schedule, start:row.start, action })
      await query(scope.current)
    } finally { setMutating(false) }
  }
  async function changeRetention(action: 'add' | 'remove') {
    if (!selectedClusterId || mutating) return
    setMutating(true)
    try {
      await mutateResource('/filesystem/snapshot/schedule/retention','POST',{ cluster_id:selectedClusterId, ...scope.current, retention, action })
      await query(scope.current)
    } finally { setMutating(false) }
  }
  async function create(values: ApiRecord) {
    if (!selectedClusterId || mutating) return
    const target = await form.validateFields()
    setMutating(true)
    try {
      await mutateResource('/filesystem/snapshot/schedule','POST',{ cluster_id:selectedClusterId, ...target, ...values })
      setCreating(false)
      await query(target)
    } finally { setMutating(false) }
  }
  return <Card title="查询路径的实时快照计划">
    <Form form={form} disabled={mutating} layout="inline" initialValues={{ path: '/' }} onFinish={query} onValuesChange={reset}>
      <Form.Item name="fs" label="文件系统" rules={[{ required: true }]}><Input /></Form.Item>
      <Form.Item name="path" label="路径" rules={[{ required: true }]}><Input /></Form.Item>
      <Form.Item name="subvol" label="子卷"><Input /></Form.Item>
      <Form.Item name="group" label="子卷组"><Input /></Form.Item>
      <Button htmlType="submit" loading={loading} disabled={!selectedClusterId}>查询</Button>
    </Form>
    <Button disabled={!selectedClusterId || mutating || loading} onClick={() => { void form.validateFields().then(() => setCreating(true)) }}>为当前路径新建计划</Button>
    <Modal title="新建快照计划" open={creating} confirmLoading={mutating} onCancel={() => { if (!mutating) setCreating(false) }} onOk={() => createForm.submit()}>
      <Form form={createForm} layout="vertical" onFinish={create}>
        <Form.Item name="schedule" label="周期" rules={[{ required:true }]}><Input placeholder="1h" /></Form.Item>
        <Form.Item name="start" label="开始时间（可选）"><Input placeholder="2026-09-14T00:00:00" /></Form.Item>
      </Form>
    </Modal>
    {Boolean(rows?.length) && <Card title="路径保留策略">
      <Alert type="info" message="保留策略作用于当前路径的全部计划。n 为最近快照数量；m/h/d/w/M/y 为分钟、小时、日、周、月、年。" />
      <Space><Input value={retention} onChange={(event) => setRetention(event.target.value)} placeholder="例如 24h7d4w" disabled={mutating} /><Button disabled={mutating || loading || !retention} onClick={() => changeRetention('add')}>添加保留规则</Button><Button disabled={mutating || loading || !retention} onClick={() => changeRetention('remove')}>移除保留规则</Button></Space>
    </Card>}
    {error && <Alert type="error" message={error} />}
    {rows?.length === 0 && <Alert type="info" message="该路径没有快照计划。查询不包含其他路径。" />}
    {rows && <AppTable<ApiRecord> dataSource={rows} rowKey={(row) => JSON.stringify([row.path,row.schedule,row.start])} expandable={{ expandedRowRender:(row) => <RecordDetail record={row} /> }} columns={[
      {title:'路径',dataIndex:'path'},
      {title:'周期',dataIndex:'schedule'},
      {title:'状态',dataIndex:'active',render:(value) => value ? '启用' : '停用'},
      {title:'开始时间（UTC）',dataIndex:'start'},
      {title:'首次快照（UTC）',dataIndex:'first',render:(value) => value ?? '—'},
      {title:'最近快照（UTC）',dataIndex:'last',render:(value) => value ?? '—'},
      {title:'已创建',dataIndex:'created_count',render:(value) => value ?? '—'},
      {title:'已清理',dataIndex:'pruned_count',render:(value) => value ?? '—'},
      {title:'最近清理（UTC）',dataIndex:'last_pruned',render:(value) => value ?? '—'},
      {title:'操作',render:(_,row) => <Space><Popconfirm title="删除这条快照计划？" description="按周期和开始时间定位；已有快照不会因此删除。" onConfirm={() => toggle(row, 'remove')}><Button danger disabled={mutating || loading}>删除计划</Button></Popconfirm><Button loading={mutating} disabled={loading} onClick={() => toggle(row)}>{row.active ? '停用' : '启用'}</Button></Space>}
    ]} />}

  </Card>
}

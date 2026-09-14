import { DeleteOutlined, DownloadOutlined, MinusCircleOutlined, PlusOutlined, ReloadOutlined, UploadOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Form, Input, Modal, Select, Space, Tag, Typography } from 'antd'
import { useCallback, useEffect, useState } from 'react'
import { jsonInit, request, textValue, type ApiRecord } from '../../api/client'
import { listResource, mutateResource, refreshResource } from '../../api/resource'
import { AppTable } from '../../components/AppTable'
import { DraggableModal } from '../../components/DraggableModal'
import { Page } from '../../components/Page'
import { ResourceMetaBar } from '../../components/ResourceMetaBar'
import { TableAction, TableActions } from '../../components/TableActions'
import { useResource } from '../../hooks'
import { useClusterContext } from '../../state/ClusterContext'
import { message } from '../../utils/appMessage'

const subsystems = ['mon', 'osd', 'mds', 'mgr']
const entityPattern = /^(client|mon|mgr|osd|mds)\.[A-Za-z0-9_.:@+\-]{1,256}$/
interface UserForm { entity: string; caps: Array<{ subsystem: string; value: string }> }

export function CephUsersPage() {
  const { selectedClusterId } = useClusterContext()
  const [selected, setSelected] = useState<React.Key[]>([])
  const [search, setSearch] = useState('')
  const [busy, setBusy] = useState(false)
  const [editTarget, setEditTarget] = useState<ApiRecord | null>(null)
  const [formOpen, setFormOpen] = useState(false)
  const [importOpen, setImportOpen] = useState(false)
  const [form] = Form.useForm<UserForm>()
  const [importForm] = Form.useForm<{ keyring: string }>()
  const loader = useCallback(async () => {
    if (!selectedClusterId) return { items: [] as ApiRecord[], stale: false, observedAt: null, staleReason: null }
    const result = await listResource('/ceph/users', selectedClusterId, { limit: 500 })
    const items: ApiRecord[] = [...result.items]
    let cursor = result.nextCursor
    while (cursor) {
      const page = await listResource('/ceph/users', selectedClusterId, { limit: 500, cursor })
      items.push(...page.items)
      result.stale ||= page.stale
      result.staleReason ||= page.staleReason
      cursor = page.nextCursor
    }
    return { ...result, items }
  }, [selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  useEffect(() => {
    setSelected([])
    setFormOpen(false)
    setImportOpen(false)
    importForm.resetFields()
  }, [selectedClusterId, importForm])

  async function collect() {
    if (!selectedClusterId) return
    await refreshResource({ clusterId: selectedClusterId, kind: 'ceph_user' })
    await refresh()
    setSelected([])
  }

  async function run(action: () => Promise<unknown>, success?: string) {
    if (busy || !selectedClusterId) return
    setBusy(true)
    try {
      await action()
      if (success) message.success(success)
    } finally { setBusy(false) }
  }

  function openEditor(row?: ApiRecord) {
    setEditTarget(row ?? null)
    form.setFieldsValue({ entity: row ? textValue(row.entity, '') : '', caps: row
      ? Object.entries((row.caps ?? {}) as Record<string, string>).map(([subsystem, value]) => ({ subsystem, value }))
      : [{ subsystem: 'mon', value: 'allow r' }] })
    setFormOpen(true)
  }

  async function save(values: UserForm) {
    if (!selectedClusterId) return
    const caps = Object.fromEntries(values.caps.map(({ subsystem, value }) => [subsystem, value]))
    if (Object.keys(caps).length !== values.caps.length) { message.error('同一子系统只能填写一条权限表达式'); return }
    await run(async () => {
      await mutateResource('/ceph/user', editTarget ? 'PATCH' : 'POST', { cluster_id: selectedClusterId, entity: values.entity, caps },
        editTarget ? { ifMatch: String(editTarget.resource_version) } : undefined)
      setFormOpen(false)
      message.success(editTarget ? 'CephX 权限已更新' : 'CephX 用户已创建')
      await collect()
    })
  }

  function deleteUsers(entities: string[]) {
    if (!selectedClusterId || !entities.length) return
    Modal.confirm({
      title: `删除 ${entities.length} 个 CephX 用户`,
      content: <Space direction="vertical"><Typography.Text>相关客户端或守护进程将失去使用这些身份访问集群的能力。</Typography.Text><Typography.Text code>{entities.join(', ')}</Typography.Text></Space>,
      okText: '删除', okType: 'danger',
      onOk: () => run(async () => {
        let removed = 0
        try {
          for (const entity of entities) {
            const row = data?.items.find((item) => item.entity === entity)
            await mutateResource('/ceph/user', 'DELETE', { cluster_id: selectedClusterId, entity },
              row ? { ifMatch: String(row.resource_version) } : undefined)
            removed++
          }
          message.success(`已删除 ${removed} 个 CephX 用户`)
        } catch (err) {
          if (removed) message.warning(`已删除 ${removed} 个用户，后续删除失败，请刷新后重试剩余用户`)
          throw err
        } finally { await collect() }
      })
    })
  }

  async function exportUsers(entities: string[]) {
    if (!selectedClusterId || !entities.length) return
    await run(async () => {
      const { keyring } = await request<{ keyring: string }>('/ceph/users/export', jsonInit('GET', { cluster_id: selectedClusterId, entities }))
      const url = URL.createObjectURL(new Blob([keyring], { type: 'text/plain;charset=utf-8' }))
      const anchor = document.createElement('a')
      anchor.href = url
      anchor.download = entities.length === 1 ? `${entities[0]}.keyring` : 'ceph-users.keyring'
      anchor.click()
      setTimeout(() => URL.revokeObjectURL(url), 1000)
    })
  }

  const filtered = data?.items.filter((row) => `${row.entity} ${JSON.stringify(row.caps)}`.toLowerCase().includes(search.toLowerCase())) ?? []
  const blocked = busy || loading || !selectedClusterId || Boolean(error)
  return <Page title="CephX 用户" loading={loading} error={error}>
    <Space direction="vertical" size={16} className="page-stack">
      <Alert type="info" showIcon message="CephX 用户用于客户端和守护进程访问 Ceph 集群。导出的 keyring 包含访问密钥，请妥善保存。" />
      <Card title="认证实体与权限" extra={<Space wrap>
        <Button icon={<ReloadOutlined />} loading={busy} disabled={!selectedClusterId} onClick={() => run(collect)}>从集群刷新</Button>
        <Button icon={<UploadOutlined />} disabled={blocked} onClick={() => { importForm.resetFields(); setImportOpen(true) }}>导入</Button>
        <Button type="primary" icon={<PlusOutlined />} disabled={blocked} onClick={() => openEditor()}>创建</Button>
      </Space>}>
        <Space direction="vertical" size={12} className="page-stack">
          <ResourceMetaBar observedAt={data?.observedAt} stale={data?.stale} staleReason={data?.staleReason} />
          {data?.stale && <Alert type="warning" showIcon message="采集结果已过期，请刷新后再修改。" />}
          <Space wrap>
            <Input.Search allowClear placeholder="搜索实体或权限" value={search} onChange={(event) => setSearch(event.target.value)} />
            <Button icon={<DownloadOutlined />} disabled={blocked || !selected.length || selected.length > 100} onClick={() => exportUsers(selected.map(String))}>导出所选（最多 100 个）</Button>
            <Button danger icon={<DeleteOutlined />} disabled={blocked || !selected.length || data?.stale} onClick={() => deleteUsers(selected.map(String))}>删除所选</Button>
          </Space>
          <AppTable<ApiRecord> rowKey="entity" size="small" dataSource={filtered} pagination={{ defaultPageSize: 20, showSizeChanger: true }}
            rowSelection={{ selectedRowKeys: selected, onChange: setSelected, preserveSelectedRowKeys: true }}
            columns={[
              { title: '实体', dataIndex: 'entity', sorter: (a, b) => String(a.entity).localeCompare(String(b.entity)) },
              { title: '类型', dataIndex: 'entity_type', filters: ['client', 'mon', 'mgr', 'osd', 'mds'].map((value) => ({ text: value, value })), onFilter: (value, row) => row.entity_type === value },
              { title: '权限', dataIndex: 'caps', render: (caps) => <Space direction="vertical">{Object.entries(caps ?? {}).map(([name, value]) => <span key={name}><Tag>{name}</Tag><Typography.Text code>{String(value)}</Typography.Text></span>)}</Space> },
              { title: '操作', render: (_, row) => <TableActions>
                <TableAction disabled={blocked || Boolean(row.stale)} onClick={() => openEditor(row)}>编辑</TableAction>
                <TableAction disabled={blocked} onClick={() => exportUsers([String(row.entity)])}>导出</TableAction>
                <TableAction disabled={blocked || Boolean(row.stale)} onClick={() => deleteUsers([String(row.entity)])}>删除</TableAction>
              </TableActions> }
            ]} />
        </Space>
      </Card>
    </Space>
    <DraggableModal title={editTarget ? `编辑 ${String(editTarget.entity)}` : '创建 CephX 用户'} open={formOpen} confirmLoading={busy}
      onCancel={() => { if (!busy) setFormOpen(false) }} onOk={() => form.submit()}>
      <Form form={form} layout="vertical" onFinish={save}>
        <Form.Item name="entity" label="实体名称" rules={[{ required: true }, { pattern: entityPattern, message: '例如 client.backup 或 osd.0' }]}><Input disabled={Boolean(editTarget)} placeholder="client.backup" /></Form.Item>
        {editTarget && <Alert type="warning" message="保存将替换该实体的全部权限；未保留的子系统权限将被移除。" />}
        <Form.List name="caps" rules={[{ validator: async (_, value) => { if (!value?.length) throw new Error('至少需要一条权限') } }]}>
          {(fields, { add, remove }, { errors }) => <>
            {fields.map(({ key, name, ...rest }) => <div key={key} style={{ display: 'flex', gap: 8, alignItems: 'baseline' }}>
              <Form.Item {...rest} name={[name, 'subsystem']} rules={[{ required: true }]}><Select style={{ width: 95 }} options={subsystems.map((value) => ({ value, label: value }))} /></Form.Item>
              <Form.Item style={{ flex: 1, minWidth: 0 }} {...rest} name={[name, 'value']} rules={[{ required: true, whitespace: true }, { max: 8192 }]}><Input placeholder="allow r / profile rbd pool=data" /></Form.Item>
              <Button aria-label="删除权限" icon={<MinusCircleOutlined />} onClick={() => remove(name)} />
            </div>)}
            <Form.ErrorList errors={errors} />
            <Button icon={<PlusOutlined />} disabled={fields.length >= 4} onClick={() => add()}>添加权限</Button>
          </>}
        </Form.List>
      </Form>
    </DraggableModal>
    <DraggableModal title="导入 Ceph keyring" open={importOpen} confirmLoading={busy} onCancel={() => { if (!busy) { setImportOpen(false); importForm.resetFields() } }} onOk={() => importForm.submit()}>
      <Form form={importForm} layout="vertical" onFinish={(values) => run(async () => {
        await mutateResource('/ceph/users/import', 'POST', { cluster_id: selectedClusterId, keyring: values.keyring })
        setImportOpen(false)
        importForm.resetFields()
        message.success('Keyring 已导入')
        await collect()
      })}>
        <Alert type="warning" showIcon message="导入会按 keyring 更新已有实体的密钥和权限。" />
        <Form.Item name="keyring" label="Keyring 内容" rules={[{ required: true, whitespace: true }, { max: 262144 }]}><Input.TextArea rows={10} placeholder={'[client.backup]\n  key = ...\n  caps mon = "allow r"'} autoComplete="off" /></Form.Item>
      </Form>
    </DraggableModal>
  </Page>
}

import { Alert, Button, Card, Form, Input, InputNumber, Modal, Select, Space, Tag, Typography } from 'antd'
import { useCallback, useEffect, useRef, useState } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'
import { listResource, mutateResource } from '../../api/resource'
import { AppTable } from '../../components/AppTable'
import { useClusterContext } from '../../state/ClusterContext'

interface DirectoryQuota {
  max_bytes: number
  max_files: number
}

interface DirectoryEntry {
  name: string
  path: string
  parent?: string | null
  mode?: string
  size?: number
  uid?: number
  gid?: number
  modified_at?: string
  quotas?: DirectoryQuota | null
}

interface DirectoryList {
  filesystem: string
  path: string
  items: DirectoryEntry[]
  observed_at: string
}

export function CephFSDirectoryBrowser() {
  const { selectedClusterId } = useClusterContext()
  const [queryForm] = Form.useForm()
  const [quotaForm] = Form.useForm()
  const [filesystems, setFilesystems] = useState<string[]>([])
  const [result, setResult] = useState<DirectoryList | null>(null)
  const [loading, setLoading] = useState(false)
  const [mutating, setMutating] = useState(false)
  const [error, setError] = useState('')
  const [editing, setEditing] = useState<DirectoryEntry | null>(null)
  const pending = useRef<AbortController | null>(null)

  const load = useCallback(async (filesystem: string, path: string) => {
    if (!selectedClusterId || !filesystem) return
    pending.current?.abort()
    const controller = new AbortController()
    pending.current = controller
    setLoading(true)
    setError('')
    try {
      const data = await request<DirectoryList>('/filesystem/entries', jsonInit('GET', {
        cluster_id: selectedClusterId,
        fs: filesystem,
        path: path || '/'
      }, { signal: controller.signal, suppressErrorNotification: true }))
      if (!controller.signal.aborted) {
        setResult(data)
        queryForm.setFieldsValue({ fs: data.filesystem, path: data.path })
      }
    } catch (err) {
      if (!controller.signal.aborted) setError(err instanceof Error ? err.message : '读取 CephFS 目录失败')
    } finally {
      if (!controller.signal.aborted) setLoading(false)
    }
  }, [queryForm, selectedClusterId])

  useEffect(() => {
    pending.current?.abort()
    setResult(null)
    setFilesystems([])
    setError('')
    if (!selectedClusterId) return
    let active = true
    void listResource('/filesystems', selectedClusterId).then((response) => {
      if (!active) return
      const names = response.items.map(filesystemName).filter(Boolean)
      setFilesystems(names)
      if (names.length > 0) {
        queryForm.setFieldsValue({ fs: names[0], path: '/' })
        void load(names[0], '/')
      }
    }).catch((err) => {
      if (active) setError(err instanceof Error ? err.message : '读取文件系统列表失败')
    })
    return () => { active = false; pending.current?.abort() }
  }, [load, queryForm, selectedClusterId])

  async function updateQuota(values: { max_bytes?: number; max_files?: number }) {
    if (!selectedClusterId || !editing || mutating || !result) return
    if (values.max_bytes === undefined && values.max_files === undefined) {
      setError('至少填写一个配额值；0 表示不限制。')
      return
    }
    setMutating(true)
    try {
      await mutateResource('/filesystem/entry/quota', 'PATCH', {
        cluster_id: selectedClusterId,
        fs: result.filesystem,
        path: editing.path,
        ...(values.max_bytes !== undefined ? { max_bytes: values.max_bytes } : {}),
        ...(values.max_files !== undefined ? { max_files: values.max_files } : {})
      })
      setEditing(null)
      await load(result.filesystem, result.path)
    } finally {
      setMutating(false)
    }
  }

  function openQuota(entry: DirectoryEntry) {
    setEditing(entry)
    quotaForm.setFieldsValue({
      max_bytes: entry.quotas?.max_bytes ?? 0,
      max_files: entry.quotas?.max_files ?? 0
    })
  }

  const current = result?.items[0]
  return <Card title="CephFS 目录浏览与配额">
    <Alert type="info" showIcon message="实时读取 CephFS 数据面" description="目录不会加入定时资源缓存。每次进入目录都会通过 cephfs-shell 读取直属子目录及其 max_bytes/max_files 配额；0 表示未限制。" />
    <Form form={queryForm} layout="inline" initialValues={{ path: '/' }} onFinish={(values) => load(String(values.fs), String(values.path || '/'))}>
      <Form.Item name="fs" label="文件系统" rules={[{ required: true }]}>
        <Select style={{ minWidth: 180 }} options={filesystems.map((value) => ({ label: value, value }))} />
      </Form.Item>
      <Form.Item name="path" label="绝对路径" rules={[{ required: true }, { pattern: /^\//, message: '路径必须以 / 开头' }]}>
        <Input style={{ minWidth: 320 }} />
      </Form.Item>
      <Button htmlType="submit" type="primary" loading={loading} disabled={!selectedClusterId}>读取</Button>
    </Form>
    {error && <Alert type="error" showIcon message={error} />}
    {result && <>
      <Space wrap>
        <Typography.Text>当前位置：<Typography.Text code>{result.path}</Typography.Text></Typography.Text>
        {current?.parent && <Button onClick={() => load(result.filesystem, current.parent!)} disabled={loading}>返回上级</Button>}
        <Typography.Text type="secondary">观测时间：{new Date(result.observed_at).toLocaleString()}</Typography.Text>
      </Space>
      <AppTable<DirectoryEntry>
        loading={loading}
        dataSource={result.items}
        rowKey={(row) => row.path}
        pagination={false}
        columns={[
          { title: '名称', dataIndex: 'name', render: (value, row) => row.path === result.path ? <Tag color="blue">{String(value)}（当前）</Tag> : <Button type="link" onClick={() => load(result.filesystem, row.path)}>{String(value)}</Button> },
          { title: '路径', dataIndex: 'path', render: (value) => <Typography.Text code>{String(value)}</Typography.Text> },
          { title: '权限', dataIndex: 'mode', render: (value) => value || '—' },
          { title: 'UID / GID', render: (_, row) => row.uid === undefined ? '—' : `${row.uid} / ${row.gid}` },
          { title: '修改时间', dataIndex: 'modified_at', render: (value) => value || '—' },
          { title: '容量配额', render: (_, row) => formatQuota(row.quotas?.max_bytes, formatBytes) },
          { title: '文件数配额', render: (_, row) => formatQuota(row.quotas?.max_files, (value) => String(value)) },
          { title: '操作', render: (_, row) => row.quotas ? <Button onClick={() => openQuota(row)}>设置配额</Button> : '根目录不设置配额' }
        ]}
      />
    </>}
    <Modal title={`设置目录配额：${editing?.path ?? ''}`} open={Boolean(editing)} confirmLoading={mutating} onCancel={() => !mutating && setEditing(null)} onOk={() => quotaForm.submit()}>
      <Alert type="warning" showIcon message="输入 0 可移除对应限制" />
      <Form form={quotaForm} layout="vertical" onFinish={updateQuota}>
        <Form.Item name="max_bytes" label="最大容量（字节）" rules={[{ type: 'number', min: 0 }]}><InputNumber min={0} precision={0} style={{ width: '100%' }} /></Form.Item>
        <Form.Item name="max_files" label="最大文件数" rules={[{ type: 'number', min: 0 }]}><InputNumber min={0} precision={0} style={{ width: '100%' }} /></Form.Item>
      </Form>
    </Modal>
  </Card>
}

function filesystemName(row: ApiRecord) {
  return String(row.name ?? row.fs ?? row.filesystem ?? '')
}

function formatQuota(value: number | undefined, formatter: (value: number) => string) {
  return value && value > 0 ? formatter(value) : <Tag>不限制</Tag>
}

function formatBytes(value: number) {
  if (value < 1024) return `${value} B`
  const units = ['KiB', 'MiB', 'GiB', 'TiB', 'PiB']
  let size = value
  let unit = -1
  while (size >= 1024 && unit < units.length - 1) { size /= 1024; unit += 1 }
  return `${size.toFixed(size >= 10 ? 1 : 2)} ${units[unit]}`
}

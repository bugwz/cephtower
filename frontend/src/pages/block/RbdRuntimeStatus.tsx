import { useState } from 'react'
import { Alert, Button, Collapse, Descriptions, Modal, Table, Tabs, Typography } from 'antd'
import { cacheFields, migrationFields, runtimeDetails, runtimeObject, runtimeWatchers } from './rbdRuntimeFields'

export function RbdRuntimeStatus({ value, name }: { value: unknown; name: string }) {
  const [open, setOpen] = useState(false)
  const status = runtimeObject(value)
  if (!status) return <Typography.Text type="secondary">运行状态未返回</Typography.Text>
  const watchers = runtimeWatchers(status.watchers)
  const migration = runtimeDetails(status.migration, migrationFields)
  const cache = runtimeDetails(status.persistent_cache, cacheFields)
  return <>
    <Button type="link" size="small" onClick={() => setOpen(true)}>运行详情{watchers ? `（${watchers.length} 个客户端）` : ''}</Button>
    <Modal title={`${name} · 运行状态`} open={open} footer={null} width={960} onCancel={() => setOpen(false)}>
      <Alert type="info" message="采集快照，不是实时连接状态；未返回迁移或缓存信息不代表操作已完成或缓存已停用。" />
      <Tabs items={[
        { key: 'watchers', label: '连接客户端', children: watchers ? <Table size="small" dataSource={watchers} rowKey="key" pagination={{ pageSize: 10 }} columns={[
          { title: '地址', dataIndex: 'address' }, { title: 'Client ID', dataIndex: 'client' }, { title: 'Cookie', dataIndex: 'cookie' }
        ]} locale={{ emptyText: '本次采集没有连接客户端' }} /> : <Alert type="warning" message="客户端列表未返回或格式无效" /> },
        { key: 'migration', label: '迁移', children: migration ? <Descriptions bordered size="small" column={1} items={migration} /> : <Typography.Text type="secondary">未返回迁移信息</Typography.Text> },
        { key: 'cache', label: '持久缓存', children: cache ? <Descriptions bordered size="small" column={2} items={cache} /> : <Typography.Text type="secondary">未返回持久缓存信息</Typography.Text> }
      ]} />
      <Collapse items={[{ key: 'raw', label: '原始运行状态', children: <pre style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{JSON.stringify(status, null, 2)}</pre> }]} />
    </Modal>
  </>
}

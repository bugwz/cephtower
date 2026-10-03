import { useState } from 'react'
import { Alert, Button, Input, Modal, Select, Space, Table, Typography } from 'antd'
import { filterRbdConfiguration, rbdConfigurationRows } from './rbdConfigurationRows'

export function RbdConfiguration({ value, name }: { value: unknown; name: string }) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [source, setSource] = useState('')
  const rows = rbdConfigurationRows(value)
  if (!rows) return <Typography.Text type="secondary">配置未返回或数据不完整</Typography.Text>
  const filtered = filterRbdConfiguration(rows, query, source)
  return <>
    <Button type="link" size="small" onClick={() => setOpen(true)}>查看 {rows.length} 项配置</Button>
    <Modal title={`${name} · 生效配置及来源`} open={open} onCancel={() => setOpen(false)} footer={null} width={900}>
      <Alert type="info" message="配置来自 rbd config image list 的采集快照。镜像覆盖优先于池覆盖；移除镜像覆盖后恢复继承，并非把参数设为零。客户端配置来源不等同于编译默认值。" />
      <Space wrap style={{ margin: '16px 0' }}>
        <Input.Search aria-label="搜索镜像配置" allowClear placeholder="搜索名称、值或来源" value={query} onChange={(event) => setQuery(event.target.value)} />
        <Select aria-label="配置来源" value={source} onChange={setSource} style={{ minWidth: 160 }} options={[
          { value: '', label: '全部来源' }, { value: 'config', label: '客户端配置' }, { value: 'pool', label: '池级覆盖' }, { value: 'image', label: '镜像级覆盖' }
        ]} />
      </Space>
      <Table size="small" rowKey="key" dataSource={filtered} pagination={{ pageSize: 15 }} locale={{ emptyText: rows.length ? '没有匹配的配置' : '本次采集未返回配置项' }} columns={[
        { title: '名称', dataIndex: 'name', sorter: (a, b) => a.name.localeCompare(b.name), render: (text: string) => <Typography.Text copyable>{text}</Typography.Text> },
        { title: '生效值（原值）', dataIndex: 'value', render: (text: string) => <span style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{text}</span> },
        { title: '来源', dataIndex: 'sourceLabel' }
      ]} />
    </Modal>
  </>
}

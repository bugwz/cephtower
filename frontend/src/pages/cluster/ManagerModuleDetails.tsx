import { Alert, Button, Descriptions, Input, Space, Typography } from 'antd'
import { useCallback, useState } from 'react'
import type { ApiRecord } from '../../api/client'
import { listAllResources } from '../../api/resource'
import { useResource } from '../../hooks'
import { AppTable } from '../../components/AppTable'
import { RecordDetail } from '../../components/RecordDetail'

export function managerOptionRows(value: unknown): ApiRecord[] | null {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return null
  const entries = Object.entries(value)
  if (entries.some(([name, option]) => !name || !option || typeof option !== 'object' || Array.isArray(option))) return null
  return entries.map(([name, option]) => ({ ...option, option_name: name }))
}

export function managerOptionValue(value: unknown): string {
  if (value === undefined || value === null) return '未采集'
  if (value === '') return '空字符串'
  if (typeof value === 'string' || typeof value === 'boolean') return String(value)
  if (typeof value === 'number' && Number.isFinite(value)) return String(value)
  if (Array.isArray(value) && value.every((item) => typeof item === 'string')) return value.length ? value.join('、') : '无'
  return '未采集'
}

export function managerModuleOverrides(rows: ApiRecord[], moduleName: unknown): ApiRecord[] {
  if (typeof moduleName !== 'string' || !moduleName || moduleName.includes('/')) return []
  const prefix = `mgr/${moduleName}/`
  return rows.filter((row) => typeof row.name === 'string' && row.name.startsWith(prefix) && row.name.length > prefix.length)
}

export function ManagerModuleDetails({ record, clusterId }: { record: ApiRecord, clusterId: number }) {
  const [search, setSearch] = useState('')
  const moduleName = record.name
  const loader = useCallback(async () => {
    const result = await listAllResources('/configuration/values', clusterId)
    return { ...result, items: managerModuleOverrides(result.items, moduleName) }
  }, [clusterId, moduleName])
  const { data: overrides, loading, error, refresh } = useResource(loader)
  const confirmed = !loading && !error && overrides?.stale === false && overrides.items.every((row) => row.stale === false)
  const rows = managerOptionRows(record.options)
  const query = search.trim().toLowerCase()
  const filtered = (rows ?? []).filter((row) => ['option_name', 'type', 'desc', 'long_desc', 'tags'].some((key) => managerOptionValue(row[key]).toLowerCase().includes(query)))
  return <Space direction="vertical" style={{ width: '100%' }}>
    {record.stale !== false && <Alert type="warning" showIcon message="模块库存已过期或新鲜度未知；以下为上次采集的参数定义。" />}
    <Descriptions size="small" bordered column={2}>
      <Descriptions.Item label="模块">{managerOptionValue(record.name)}</Descriptions.Item>
      <Descriptions.Item label="加载错误">{managerOptionValue(record.error_string)}</Descriptions.Item>
      {(['enabled', 'always_on', 'can_run', 'force_disabled'] as const).map((key, index) => <Descriptions.Item key={key} label={['启用', '常驻', '可运行', '强制停用'][index]}>{record[key] === true ? '是' : record[key] === false ? '否' : '未采集'}</Descriptions.Item>)}
    </Descriptions>
    <Typography.Paragraph type="secondary">来源：ceph mgr dump 的 module_options。这里展示参数定义和原始默认值，不代表当前守护进程的生效值。</Typography.Paragraph>
    {rows === null && <Alert type="warning" message="未取得有效的模块参数定义，不能据此判断该模块没有配置项。" />}
    <Input.Search allowClear placeholder="搜索参数名、类型、标签或说明" value={search} onChange={(event) => setSearch(event.target.value)} />
    <AppTable<ApiRecord> size="small" rowKey="option_name" dataSource={filtered}
      locale={{ emptyText: rows === null ? '参数定义未采集' : query ? '没有匹配的参数' : '本次采集的模块没有参数定义' }}
      columns={[
        { title: '参数', dataIndex: 'option_name', render: (value) => <Typography.Text copyable>{String(value)}</Typography.Text> },
        ...[['type', '类型'], ['level', '级别'], ['default_value', '原始默认值'], ['min', '最小值'], ['max', '最大值'], ['desc', '说明']].map(([dataIndex, title]) => ({ title, dataIndex, render: managerOptionValue }))
      ]}
      expandable={{ expandedRowRender: (row) => <Descriptions size="small" column={1} bordered>
        {([['long_desc', '详细说明'], ['enum_allowed', '可选值'], ['flags', '原始标志'], ['tags', '标签'], ['see_also', '相关选项']] as const).map(([key, label]) => <Descriptions.Item key={key} label={label}>{managerOptionValue(row[key])}</Descriptions.Item>)}
      </Descriptions> }}
    />
    <Space><Typography.Title level={5} style={{ margin: 0 }}>显式配置覆盖</Typography.Title><Button loading={loading} onClick={() => refresh()}>重新读取库存</Button></Space>
    <Typography.Paragraph type="secondary">来源：ceph config dump 库存，配置键前缀 mgr/{String(moduleName)}/。保留各作用域和本地化键，不合并推断运行时生效值。修改请使用模块列表中的“编辑配置”。</Typography.Paragraph>
    {!confirmed && <Alert type="warning" showIcon message={error ? `配置覆盖读取失败：${error}` : '配置覆盖正在读取、已过期或新鲜度未知，不能据此确认当前配置。'} />}
    <AppTable<ApiRecord> size="small" rowKey="natural_key" loading={loading} dataSource={overrides?.items ?? []}
      locale={{ emptyText: confirmed ? '本次采集未发现该模块的显式覆盖' : '配置覆盖未确认' }}
      columns={[
        { title: '完整配置键', dataIndex: 'name', render: managerOptionValue },
        { title: '作用域（含位置限制）', dataIndex: 'who', render: managerOptionValue },
        { title: '配置值', dataIndex: 'value', render: (value) => <Typography.Text style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{managerOptionValue(value)}</Typography.Text> },
        { title: '采集时间', dataIndex: 'observed_at', render: managerOptionValue }
      ]}
    />
    <details><summary>原始模块记录</summary><RecordDetail record={record} /></details>
  </Space>
}

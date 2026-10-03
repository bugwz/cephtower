import { Table } from 'antd'
import type { ApiRecord } from '../../api/client'
import { rgwUserKeyRows } from './rgwUserKeyRows'

export function RgwUserKeyTable({ value, protocol }: { value: unknown; protocol: 'S3' | 'Swift' }) {
  const rows = rgwUserKeyRows(value)
  if (!rows) return <p>{protocol} 密钥信息未返回或格式无效</p>
  return <Table size="small" rowKey="key" dataSource={rows}
    pagination={rows.length > 5 ? { pageSize: 5 } : false}
    locale={{ emptyText: `未配置 ${protocol} 密钥` }} columns={[
      { title: '所属用户 / 子用户（命令原值）', dataIndex: 'user' },
      { title: '密钥状态', dataIndex: 'state' }
    ]} />
}

export function RgwUserKeys({ row }: { row: ApiRecord }) {
  return <>
    <p>此处仅展示密钥元数据。访问密钥和秘密密钥在库存中已脱敏，无法从本页查看或恢复；密钥状态不代表用户的完整有效权限。</p>
    <h4>S3 密钥</h4>
    <RgwUserKeyTable value={row.keys} protocol="S3" />
    <h4>Swift 密钥</h4>
    <RgwUserKeyTable value={row.swift_keys} protocol="Swift" />
  </>
}

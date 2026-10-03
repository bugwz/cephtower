import { Table } from 'antd'
import { rgwPermissionRows } from './rgwPermissionRows'

export function RgwPermissions({ value, subusers = false }: { value: unknown; subusers?: boolean }) {
  const rows = rgwPermissionRows(value, subusers)
  if (!rows) return <span>{subusers ? '子用户信息不可用' : '管理权限信息不可用'}</span>
  return <Table size="small" rowKey="key" dataSource={rows}
    pagination={rows.length > 5 ? { pageSize: 5 } : false}
    locale={{ emptyText: subusers ? '未配置子用户' : '未配置管理权限' }} columns={[
      { title: subusers ? '子用户 ID' : '管理权限类型', dataIndex: 'identity' },
      { title: '权限（命令原值）', dataIndex: 'permission' }
    ]} />
}

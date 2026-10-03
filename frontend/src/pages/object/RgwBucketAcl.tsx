import { Table } from 'antd'

type Identity = { type: string; id: string; display_name: string; uri: string; email_address: string }
type ACL = { owner: { id: string; display_name: string }; grants: { grantee: Identity; permissions: string[] }[] }
export function bucketAclData(value: unknown): ACL | undefined {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return undefined
  const acl = value as ACL
  if (!acl.owner || typeof acl.owner.id !== 'string' || !acl.owner.id || typeof acl.owner.display_name !== 'string' || !Array.isArray(acl.grants)) return undefined
  if (acl.grants.some(grant => !grant || !grant.grantee || !Array.isArray(grant.permissions) || grant.permissions.some(p => typeof p !== 'string' || !p)
    || ['type', 'id', 'display_name', 'uri', 'email_address'].some(key => typeof grant.grantee[key as keyof Identity] !== 'string') || !grant.grantee.type
    || (grant.grantee.type === 'CanonicalUser' && !grant.grantee.id)
    || (grant.grantee.type === 'Group' && !grant.grantee.uri)
    || (grant.grantee.type === 'AmazonCustomerByEmail' && !grant.grantee.email_address))) return undefined
  return acl
}

export function bucketAclAudience(grantee: Identity, ownerID: string) {
  if (grantee.type === 'CanonicalUser') return grantee.id === ownerID ? '所有者' : '指定用户'
  if (grantee.type === 'AmazonCustomerByEmail') return '邮箱授权'
  if (grantee.type === 'Group') {
    if (grantee.uri === 'http://acs.amazonaws.com/groups/global/AllUsers') return '所有用户（包括匿名访问者）'
    if (grantee.uri === 'http://acs.amazonaws.com/groups/global/AuthenticatedUsers') return '所有已认证用户（不限本账户）'
    return '其他授权组'
  }
  return `未知受权者类型 ${JSON.stringify(grantee.type)}`
}

export function RgwBucketAcl({ value, configured }: { value: unknown; configured: unknown }) {
  const acl = bucketAclData(value)
  if (configured !== true || !acl) return <span>ACL 数据不可用</span>
  return <div>
    <p>所有者 ID：{JSON.stringify(acl.owner.id)}；显示名称：{JSON.stringify(acl.owner.display_name)}</p>
    <p>仅展示原生 Bucket ACL；最终访问还受 Bucket Policy 等规则影响，不代表对象 ACL 或有效权限。</p>
    <Table size="small" rowKey="index" dataSource={acl.grants.map((grant, index) => ({ ...grant, index }))} pagination={acl.grants.length > 5 ? { pageSize: 5 } : false} scroll={{ x: 600 }} locale={{ emptyText: '原生 ACL 授权列表为空（不推断有效权限）' }} columns={[
      { title: '授权对象', dataIndex: 'grantee', render: (grantee: Identity) => bucketAclAudience(grantee, acl.owner.id) },
      { title: '原始身份', dataIndex: 'grantee', render: (grantee: Identity) => <span style={{ whiteSpace: 'pre-wrap' }}>{Object.entries(grantee).map(([key, text]) => `${key}: ${JSON.stringify(text)}`).join('\n')}</span> },
      { title: '权限（原生值）', dataIndex: 'permissions', render: (permissions: string[]) => permissions.length ? permissions.map(p => JSON.stringify(p)).join(' / ') : '未返回权限项' }
    ]} />
  </div>
}

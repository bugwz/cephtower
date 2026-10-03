import type { ApiRecord } from '../../api/client'
import { rgwSubuserPermissionOptions } from './rgwUserSubuser'

export function rgwSubuserCreateInput(values: Record<string, unknown>, row?: ApiRecord) {
  const name = values.subuser
  if (typeof name !== 'string' || !/^[A-Za-z0-9_.@-]+$/.test(name) || name.startsWith('-')) throw new Error('子用户名仅支持字母、数字、下划线、点、@ 和非首位短横线；不要包含 UID 或冒号')
  if (typeof row?.uid !== 'string' || !row.uid || !Array.isArray(row.subusers)) throw new Error('请先刷新用户及子用户库存')
  const id = `${row.uid}:${name}`
  if (row.subusers.some(item => item && typeof item === 'object' && (item as ApiRecord).id === id)) throw new Error('该子用户已存在，请使用管理已有子用户入口')
  if (values.confirm_subuser !== id) throw new Error('请输入完整子用户 ID 确认')
  if (!rgwSubuserPermissionOptions.some(option => option.value === values.subuser_permission)) throw new Error('请选择子用户权限')
  if (values.key_type !== 's3' && values.key_type !== 'swift') throw new Error('请选择密钥类型')
  const secret = values.secret_key
  if (typeof secret !== 'string' || !secret || new TextEncoder().encode(secret).length > 256 || secret.trim() !== secret || /[\x00-\x1f\x7f]/.test(secret)) throw new Error('请填写无控制字符或首尾空白的 Secret Key（最多 256 字节）')
  if (values.credentials_saved !== 'saved') throw new Error('请先将凭据保存在安全位置；提交后本界面不会显示密钥')
  const common = { action: 'create', subuser: name, confirm_subuser: id, subuser_permission: values.subuser_permission, key_type: values.key_type, secret_key: secret }
  if (values.key_type === 'swift') return common
  if (typeof values.access_key !== 'string' || !/^[A-Za-z0-9]{1,128}$/.test(values.access_key)) throw new Error('S3 Access Key 必须为 1 至 128 位字母或数字')
  return { ...common, access_key: values.access_key }
}

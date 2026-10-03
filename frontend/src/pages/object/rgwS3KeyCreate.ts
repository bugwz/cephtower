import type { ApiRecord } from '../../api/client'
import { rgwSubuserOptions } from './rgwUserSubuser'

export function rgwS3KeyOwnerOptions(row?: ApiRecord) {
  if (typeof row?.uid !== 'string' || !row.uid) return []
  return [{ label: `${row.uid}（用户本身）`, value: row.uid }, ...rgwSubuserOptions(row).map(option => ({ label: `${row.uid}:${option.value}（子用户）`, value: `${row.uid}:${option.value}` }))]
}

export function rgwS3KeyCreateInput(values: Record<string, unknown>, row?: ApiRecord) {
  if (!rgwS3KeyOwnerOptions(row).some(option => option.value === values.owner)) throw new Error('请选择该用户或其已有子用户')
  if (values.confirm_owner !== values.owner) throw new Error('请输入完整凭据所属用户 ID 确认')
  if (typeof values.access_key !== 'string' || !/^[A-Za-z0-9]{1,128}$/.test(values.access_key)) throw new Error('Access Key 必须为 1 至 128 位字母或数字')
  const secret = values.secret_key
  if (typeof secret !== 'string' || !secret || new TextEncoder().encode(secret).length > 256 || secret.trim() !== secret || /[\x00-\x1f\x7f]/.test(secret)) throw new Error('请提供无控制字符或首尾空白的 Secret Key（最多 256 字节）')
  if (values.credentials_saved !== 'saved') throw new Error('请先将凭据保存在安全位置，提交后不回显')
  const common = { access_key: values.access_key, secret_key: secret, confirm_owner: values.owner }
  return values.owner === row!.uid ? common : { ...common, subuser: (values.owner as string).slice((row!.uid as string).length + 1) }
}

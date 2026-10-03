import type { ApiRecord } from '../../api/client'
import { rgwS3KeyOwnerOptions } from './rgwS3KeyCreate'

export function rgwS3KeyDeleteOptions(row?: ApiRecord) {
  if (!Array.isArray(row?.keys)) return []
  const keys = row.keys
  return rgwS3KeyOwnerOptions(row).filter(option => keys.some(key => key && typeof key === 'object' && (key as ApiRecord).user === option.value))
}

export function rgwS3KeyDeleteInput(values: Record<string, unknown>, row?: ApiRecord) {
  if (!rgwS3KeyDeleteOptions(row).some(option => option.value === values.owner)) throw new Error('请选择已有 S3 密钥的所属用户，请先刷新库存')
  if (values.confirm_owner !== values.owner) throw new Error('请输入完整凭据所属用户 ID 确认')
  if (typeof values.access_key !== 'string' || !/^[A-Za-z0-9]{1,128}$/.test(values.access_key)) throw new Error('请提供待删除的原 Access Key（1 至 128 位字母或数字）')
  const common = { access_key: values.access_key, confirm_owner: values.owner }
  return values.owner === row!.uid ? common : { ...common, subuser: (values.owner as string).slice((row!.uid as string).length + 1) }
}

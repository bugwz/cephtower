import type { ApiRecord } from '../../api/client'
import { rgwSubuserOptions } from './rgwUserSubuser'

export function rgwSwiftRotationOptions(row?: ApiRecord) {
  if (!Array.isArray(row?.swift_keys)) return []
  const keys = row.swift_keys
  return rgwSubuserOptions(row).flatMap(option => {
    const matching = keys.filter(key => key && typeof key === 'object' && (key as ApiRecord).user === option.label)
    if (matching.length !== 1 || typeof (matching[0] as ApiRecord).active !== 'boolean') return []
    const active = (matching[0] as ApiRecord).active as boolean
    return [{ ...option, label: `${option.label}（${active ? '已激活' : '已停用'}，轮换保留状态）`, active }]
  })
}

export function rgwSwiftRotationInput(values: Record<string, unknown>, row?: ApiRecord) {
  const selected = rgwSwiftRotationOptions(row).find(option => option.value === values.subuser)
  if (!selected) throw new Error('请选择已有且状态明确的 Swift 子用户密钥，请先刷新库存')
  const id = `${row!.uid}:${selected.value}`
  if (values.confirm_subuser !== id) throw new Error('请输入完整子用户 ID 确认')
  const secret = values.secret_key
  if (typeof secret !== 'string' || !secret || new TextEncoder().encode(secret).length > 256 || secret.trim() !== secret || /[\x00-\x1f\x7f]/.test(secret)) throw new Error('请填写无控制字符或首尾空白的新 Secret Key（最多 256 字节）')
  if (values.credentials_saved !== 'saved') throw new Error('请先安全保存新凭据并准备更新客户端配置')
  return { action: 'rotate-swift-key', subuser: selected.value, confirm_subuser: id, expected_key_active: selected.active, secret_key: secret }
}

export function rgwUserCreateCredentials(values: Record<string, unknown>) {
  if (values.credential_mode === 'none') return {}
  if (values.credential_mode !== 's3') throw new Error('请选择首次凭据模式')
  const { access_key, secret_key } = values
  if (typeof access_key !== 'string' || !/^[A-Za-z0-9]{1,128}$/.test(access_key)) throw new Error('请输入有效的 Access Key')
  if (typeof secret_key !== 'string' || !secret_key || secret_key.trim() !== secret_key || /[\x00-\x1f\x7f]/.test(secret_key) || new TextEncoder().encode(secret_key).length > 256) throw new Error('请输入有效的 Secret Key')
  if (values.credentials_saved !== 'saved') throw new Error('请先安全保存凭据再确认提交')
  return { access_key, secret_key }
}

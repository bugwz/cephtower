export function topicCreateInput(values: Record<string,unknown>) {
  const name=typeof values.name==='string' ? values.name : ''
  const scope=typeof values.scope==='string' ? values.scope : ''
  const zonegroup=typeof values.zonegroup==='string' ? values.zonegroup : ''
  const owner_uid=typeof values.owner_uid==='string' ? values.owner_uid : ''
  if (!/^[A-Za-z0-9_-]{1,256}$/.test(name)) throw new Error('Topic 名称仅允许字母、数字、下划线和短横线，最长 256 字符')
  if (!zonegroup || /[:\s]/.test(zonegroup) || /[:$\s]/.test(scope) || !/^[A-Za-z0-9_.$@-]{1,512}$/.test(owner_uid)) throw new Error('请填写准确 Zonegroup 名称、租户/Account 范围及凭据所属完整 UID')
  if (values.confirm_create!=='acknowledged') throw new Error('请确认凭据、命名空间和创建影响')
  if (values.persistent!=='true' && values.persistent!=='false') throw new Error('请明确选择持久化状态')
  if (values.endpoint_mode!=='none' && values.endpoint_mode!=='url') throw new Error('请选择无推送端点或指定 URL')
  const endpoint_secret=values.endpoint_mode==='none' ? '' : values.endpoint_secret
  if (typeof endpoint_secret!=='string'||(values.endpoint_mode==='url'&&!endpoint_secret)) throw new Error('请填写完整推送 URL')
  if (endpoint_secret) {
    if (!/^(https?|amqps?|kafka):\/\/(([^:\s]+):([^@\s]+)@)?([A-Za-z0-9.:-]+)(\/[\x20-\x7E]*)?$/.test(endpoint_secret)) throw new Error('推送 URL 不符合原生格式')
    try { const url=new URL(endpoint_secret);if (!url.hostname || (!!url.username!==!!url.password)) throw new Error() } catch { throw new Error('推送 URL 无效') }
  }
  const numeric:Record<string,string>={}
  for (const key of ['time_to_live','max_retries','retry_sleep_duration']) {
    const value=values[key]
    if (typeof value!=='string'||(value!=='None'&&(!/^(0|[1-9][0-9]*)$/.test(value)||Number(value)>2147483647))) throw new Error('数值必须为 0–2147483647，或 None 表示全局默认')
    numeric[key]=value
  }
  const policy=typeof values.policy==='string' ? values.policy : ''
  if (policy) {let parsed;try {parsed=JSON.parse(policy)} catch {throw new Error('Policy 必须是 JSON 对象')};if (!parsed||typeof parsed!=='object'||Array.isArray(parsed)) throw new Error('Policy 必须是 JSON 对象')}
  let options:unknown
  try {options=JSON.parse(String(values.options ?? '{}'))} catch {throw new Error('投递参数必须是 JSON 对象')}
  const allowed=['verify-ssl','use-ssl','cloudevents','ca-location','amqp-version','amqp-exchange','amqp-ack-level','http-ack-level','kafka-ack-level','mechanism','kafka-brokers']
  if (!options||typeof options!=='object'||Array.isArray(options)||Object.entries(options).some(([k,v])=>!allowed.includes(k)||typeof v!=='string')) throw new Error('投递参数仅接受已支持的非凭据字符串字段；值由后端再次校验')
  const topic_id=btoa(Array.from(new TextEncoder().encode(`${scope}:${name}`),byte=>String.fromCharCode(byte)).join('')).replace(/\+/g,'-').replace(/\//g,'_').replace(/=+$/,'')
  const body={topic_id,topic_arn:`arn:aws:sns:${zonegroup}:${scope}:${name}`,owner_uid,endpoint_secret,persistent:values.persistent==='true',opaque_data:typeof values.opaque_data==='string'?values.opaque_data:'',policy,...numeric,options}
  if (new TextEncoder().encode(JSON.stringify(body)).length>1024*1024-128) throw new Error('创建请求超过大小限制')
  return body
}
export function topicCreateConfirmation(values:Record<string,unknown>) {
  const input=topicCreateInput(values)
  return `确认创建 ${input.topic_arn}？后端会核对已配置 S3 永久密钥属于 UID ${input.owner_uid}，并核验租户或 Account；不支持临时会话或子用户密钥。请确认 HTTPS RGW 端点对应所填 Zonegroup，优先使用主 Zone。不存在检查不是原子锁，外部并发创建仍可能被原生 CreateTopic 覆盖。持久化推送可能创建队列；Policy 可能使回读被拒绝。URL 凭据不会回显，明文推送可能泄露通知和凭据，请确认目标可信。0 的 TTL/次数表示无限，0 的间隔表示无延迟，None 为全局默认。失败可能已创建，不会自动重试或删除回滚；不创建桶通知规则，不代表消息已送达。`
}

import { Alert, Button, Card, Checkbox, Input, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'

export function RgwRealmToken({ row, clusterId }: { row: ApiRecord; clusterId?: number }) {
  const scope = JSON.stringify([clusterId, row.id, row.name, row.stale])
  const current = useRef(scope)
  current.current = scope
  const sequence = useRef(0)
  const [state, setState] = useState({ scope, token: '', error: '', busy: false, acknowledged: false })
  const scoped = state.scope === scope
  const valid = !!clusterId && typeof row.id === 'string' && !!row.id && typeof row.name === 'string' && !!row.name && row.stale !== true
  useEffect(() => {
    sequence.current++
    setState({ scope, token: '', error: '', busy: false, acknowledged: false })
    return () => { sequence.current++ }
  }, [scope])
  const reset = () => { sequence.current++; setState({ scope, token: '', error: '', busy: false, acknowledged: false }) }
  async function read() {
    if (!valid || !scoped || state.busy || !state.acknowledged) return
    const ticket = ++sequence.current
    setState({ ...state, token: '', error: '', busy: true })
    try {
      const result = await request<{ token: string }>('/rgw/realm/token', jsonInit('POST', { cluster_id: clusterId, realm_id: row.id, name: row.name }, { cache: 'no-store', suppressErrorNotification: true }))
      if (current.current !== scope || sequence.current !== ticket) return
      if (typeof result?.token !== 'string' || !result.token || result.token.length > 65536 || !/^[A-Za-z0-9+/]+={0,2}$/.test(result.token)) throw new Error('invalid token')
      setState({ ...state, busy: false, token: result.token, error: '' })
    } catch {
      if (current.current === scope && sequence.current === ticket) setState({ ...state, busy: false, token: '', error: 'Token 读取失败或不可用。请检查 rgw 管理模块、本地主 Zone、端点和系统密钥，以及 Realm 身份是否已变化。' })
    }
  }
  return <Card size="small" title="Realm 引导 Token">
    <Space direction="vertical" style={{ width: '100%' }}>
      <Alert type="warning" message="包含主 Zone 系统访问密钥，请仅交给可信的远端站点管理员" description="读取现有凭据，不生成或轮换密钥，也不执行导入、部署或同步。需启用 rgw 管理模块并具有可用的本地主 Zone。关闭详情或切换集群后清空；不要粘贴到日志或工单。" />
      {!valid && <Alert type="info" message="Realm 身份或库存状态不可用，请重新采集后再试" />}
      <Checkbox checked={scoped && state.acknowledged} disabled={!valid || (scoped && state.busy)} onChange={event => setState({ scope, token: '', error: '', busy: false, acknowledged: event.target.checked })}>我了解 Token 包含系统密钥</Checkbox>
      <Space><Button disabled={!valid || !scoped || !state.acknowledged || state.busy} loading={scoped && state.busy} onClick={() => void read()}>读取敏感 Token</Button><Button onClick={reset}>清空</Button></Space>
      {scoped && state.error && <Alert type="error" message={state.error} />}
      {scoped && state.token && <Input.Password aria-label="Realm 引导 Token（敏感）" value={state.token} readOnly autoComplete="off" />}
    </Space>
  </Card>
}

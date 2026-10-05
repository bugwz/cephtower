import { Alert, Button, Card, Checkbox, Form, Input, InputNumber, Select, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { listClusters } from '../../api/cluster'
import { jsonInit, request, type ApiRecord } from '../../api/client'
import { mutateResource } from '../../api/resource'
import { useClusterContext } from '../../state/ClusterContext'
import { rgwRealmImportAction } from './rgwRealmImport'
import { transferRealm, validTransferPair } from './realmTransferWorkflow'

export function RgwRealmTransfer({ row, clusterId }: { row: ApiRecord; clusterId?: number }) {
  const { clusters, loading, error } = useClusterContext()
  const [form] = Form.useForm()
  const [busy, setBusy] = useState(false)
  const [acknowledged, setAcknowledged] = useState(false)
  const [outcome, setOutcome] = useState<'success' | 'error'>()
  const active = useRef(true)
  const sequence = useRef(0)
  const locked = useRef(false)
  const source = clusters.find(c => c.id === clusterId)
  const scope = JSON.stringify([clusterId, row.id, row.name, row.stale, loading, error, clusters.map(c => [c.id, c.fsid, c.generation, c.enabled])])
  const currentScope = useRef(scope)
  currentScope.current = scope
  useEffect(() => {
    active.current = true
    sequence.current++
    setAcknowledged(false)
    setOutcome(undefined)
    return () => { active.current = false; sequence.current++ }
  }, [scope])
  const valid = !!source?.enabled && !!source.fsid && !loading && !error && row.stale !== true && typeof row.id === 'string' && !!row.id && typeof row.name === 'string' && !!row.name
  const mode = Form.useWatch('placement_mode', form)
  const zoneMode = Form.useWatch('zone_mode', form)
  async function submit(values: ApiRecord) {
    if (!valid || !source || !acknowledged || locked.current) return
    const target = clusters.find(c => c.id === values.target_cluster)
    if (!target || !validTransferPair(source, target)) return
    locked.current = true
    setBusy(true)
    setOutcome(undefined)
    const ticket = ++sequence.current
    const current = () => active.current && currentScope.current === scope && sequence.current === ticket
    try {
      await transferRealm({ source, target, realmId: String(row.id), realmName: String(row.name), values: { ...values, confirm_import: 'acknowledged' } }, {
        current,
        clusters: listClusters,
        token: body => request('/rgw/realm/token', jsonInit('POST', body, { cache: 'no-store', suppressErrorNotification: true })),
        importZone: body => mutateResource('/rgw/realm/import', 'POST', body)
      })
      if (current()) setOutcome('success')
    } catch {
      if (current()) setOutcome('error')
    } finally {
      locked.current = false
      if (active.current) { setBusy(false); setAcknowledged(false) }
    }
  }
  return <Card size="small" title="将已有 Realm 导入其他集群">
    <Space direction="vertical" style={{ width: '100%' }}>
      <Alert type="warning" message={`源 Realm：${String(row.name ?? '')}；源集群：${source?.name ?? '不可用'}`} description="读取现有主站 Token 并提交到明确选择的目标集群，不重新初始化主站。两端需预先配置原生命令连接及 rgw 管理模块，不依赖或设置 Ceph Dashboard 凭据。关闭详情或切换范围会阻止尚未提交的导入；已提交操作不会撤销，请到目标集群操作记录查看。" />
      <Alert type="warning" message="跨集群操作不可回滚" description={rgwRealmImportAction.confirmation?.({ zone_mode: zoneMode })} />
      {!valid && <Alert type="info" message="源 Realm 或集群身份不可用，请刷新库存和集群列表" />}
      <Form form={form} layout="vertical" initialValues={{ zone_mode: 'normal', port: 80, placement_mode: 'default' }} disabled={busy || !valid} onValuesChange={() => { sequence.current++; setAcknowledged(false); setOutcome(undefined) }} onFinish={values => void submit(values)}>
        <Form.Item name="target_cluster" label="目标集群（禁止相同 FSID）" rules={[{ required: true }]}><Select options={clusters.filter(c => validTransferPair(source, c)).map(c => ({ value: c.id, label: `${c.name} — ${c.fsid}` }))} /></Form.Item>
        <Form.Item name="name" label="新的从 Zone 名称" rules={[{ required: true }, { pattern: /^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$/ }]}><Input /></Form.Item>
        <Form.Item name="zone_mode" label="Zone 类型"><Select options={[{ value: 'normal', label: '普通从 Zone' }, { value: 'archive', label: '归档从 Zone（仅从主 Zone 同步）' }]} /></Form.Item>
        <Form.Item name="port" label="RGW 前端端口（非 TLS）" rules={[{ required: true }]}><InputNumber min={1} max={65535} precision={0} /></Form.Item>
        <Form.Item name="placement_mode" label="部署位置"><Select options={[{ value: 'default', label: 'Ceph 默认位置' }, { value: 'hosts', label: '指定主机' }, { value: 'label', label: '指定标签' }]} /></Form.Item>
        {mode === 'hosts' && <Form.Item name="hosts" label="目标主机名（每行一个）" rules={[{ required: true }]}><Input.TextArea /></Form.Item>}
        {mode === 'label' && <Form.Item name="label" label="目标主机标签" rules={[{ required: true }]}><Input /></Form.Item>}
        <Form.Item name="count" label="实例数（留空使用默认）"><InputNumber min={1} precision={0} /></Form.Item>
        <Checkbox checked={acknowledged} disabled={busy || !valid} onChange={event => setAcknowledged(event.target.checked)}>已核对源 Realm、目标集群和部署参数，确认传递系统密钥并发布 Period、部署从 Zone</Checkbox>
        <div><Button htmlType="submit" danger disabled={!valid || busy || !acknowledged} loading={busy}>读取并导入目标集群</Button></div>
      </Form>
      {outcome === 'success' && <Alert type="success" message="目标 Zone、Period 和服务规格已核验，部署已提交；不代表进程就绪或复制完成" />}
      {outcome === 'error' && <Alert type="error" message="跨集群导入未确认成功。可能已部分生效，请先检查目标操作记录和两端状态；不要直接重复提交。" />}
    </Space>
  </Card>
}

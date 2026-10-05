import { Alert, Button, Card, Space, Tree } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { readRgwTopology, type RgwTopology, type RgwTopologyNode } from './rgwTopologyData'

export function RgwTopologyView({ clusterId }: { clusterId?: number }) {
  const current = useRef(clusterId)
  current.current = clusterId
  const sequence = useRef(0), locked = useRef(false)
  const [state, setState] = useState<{ clusterId?: number; busy: boolean; error: boolean; data?: RgwTopology; selected?: RgwTopologyNode }>({ clusterId, busy: false, error: false })
  useEffect(() => {
    sequence.current++
    locked.current = false
    setState({ clusterId, busy: false, error: false })
    return () => { sequence.current++ }
  }, [clusterId])
  const scoped = state.clusterId === clusterId
  async function read() {
    if (!clusterId || !scoped || locked.current) return
    locked.current = true
    const ticket = ++sequence.current
    setState({ clusterId, busy: true, error: false })
    try {
      const data = await readRgwTopology(clusterId)
      if (current.current === clusterId && ticket === sequence.current) setState({ clusterId, busy: false, error: false, data })
    } catch {
      if (current.current === clusterId && ticket === sequence.current) setState({ clusterId, busy: false, error: true })
    } finally {
      if (current.current === clusterId && ticket === sequence.current) locked.current = false
    }
  }
  return <Card title="本地 Multisite 拓扑（采集库存）" style={{ marginBottom: 16 }}>
    <Space direction="vertical" style={{ width: '100%' }}>
      <Alert type="info" message="按 Realm → Zonegroup → Zone 的原生 ID 关联" description="读取三类已采集库存，不执行 Ceph 刷新或修改。库存可能来自不同采集时间，不是原子快照，也不是已发布 Period 或实时复制状态。点击节点查看配置；未关联节点单独展示。配置操作仍在下方 Realm 列表和 Zonegroup / Zone 页面进行。" />
      <Button disabled={!clusterId || !scoped || state.busy} loading={scoped && state.busy} onClick={() => void read()}>读取本地拓扑</Button>
      {scoped && state.error && <Alert type="error" message="拓扑读取失败、数据不完整或身份重复；未将失败来源当作空列表，请检查采集状态后重试" />}
      {scoped && state.data && <>
        {state.data.stale && <Alert type="warning" message="拓扑包含过期库存，不能据此确认当前配置" />}
        {state.data.nodes.length ? <Tree treeData={state.data.nodes} selectedKeys={state.selected ? [state.selected.key] : []} onSelect={(_keys, info) => setState({ ...state, selected: info.selected ? info.node : undefined })} /> : <Alert type="info" message="本次 API 未列出拓扑节点，请结合采集状态判断；不代表已经验证没有配置" />}
        {state.selected && <pre aria-label="拓扑节点配置" style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{JSON.stringify(state.selected.details, null, 2)}</pre>}
      </>}
    </Space>
  </Card>
}

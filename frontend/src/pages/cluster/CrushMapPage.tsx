import { Alert, Button, Card, Col, Descriptions, Row, Space, Tag, Tree } from 'antd'
import { useEffect, useState } from 'react'
import { jsonInit, request, type ApiRecord } from '../../api/client'
import { useClusterContext } from '../../state/ClusterContext'
import { CrushRulesPanel } from './CrushRulesPanel'

interface CrushMap { nodes: ApiRecord[]; roots: number[] }
interface CrushTreeNode { key: string; title: string; nodeId: number; status: unknown; children: CrushTreeNode[] }

export function crushStatus(status: unknown) {
  const label = typeof status === 'string' && status.trim() ? status : '未知'
  return { label, color: ['up', 'in'].includes(label) ? 'success' : ['down', 'out', 'destroyed'].includes(label) ? 'error' : 'default' }
}

export function crushTree(data: CrushMap): CrushTreeNode[] {
  const nodes = new Map(data.nodes.map((node) => [Number(node.id), node]))
  const build = (id: number, path: Set<number>, key: string): CrushTreeNode => {
    const node = nodes.get(id)
    if (!node || path.has(id)) throw new Error('CRUSH 拓扑包含缺失节点或环路')
    const next = new Set(path); next.add(id)
    return { key, nodeId: id, status: node.status, title: `${node.name} (${node.type})`, children: (Array.isArray(node.children) ? node.children : []).map((child) => build(Number(child), next, `${key}/${child}`)) }
  }
  return data.roots.map((id) => build(id, new Set(), String(id)))
}

export function CrushMapPage() {
  const { selectedClusterId } = useClusterContext()
  const [data, setData] = useState<CrushMap | null>(null)
  const [selected, setSelected] = useState<ApiRecord | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [revision, setRevision] = useState(0)
  useEffect(() => {
    const abort = new AbortController()
    setData(null); setSelected(null); setError(''); setLoading(false)
    if (!selectedClusterId) return () => abort.abort()
    setLoading(true)
    void request<CrushMap>('/crush/map', jsonInit('GET', { cluster_id: selectedClusterId }, { signal: abort.signal, suppressErrorNotification: true }))
      .then((value) => { crushTree(value); if (!abort.signal.aborted) setData(value) })
      .catch((err) => { if (!abort.signal.aborted) setError(err instanceof Error ? err.message : 'CRUSH 拓扑读取失败') })
      .finally(() => { if (!abort.signal.aborted) setLoading(false) })
    return () => abort.abort()
  }, [selectedClusterId, revision])
  return <><Card title="CRUSH 拓扑" loading={loading} extra={<Button disabled={!selectedClusterId || loading} onClick={() => setRevision((value) => value + 1)}>刷新</Button>}>
    {!selectedClusterId && <Alert type="info" message="请先选择集群" />}
    {error && <Alert type="error" message={error} />}
    {data && <Space direction="vertical" style={{ width: '100%' }}>
      <Alert type="info" message="来自 ceph osd tree 的即时只读快照。点击节点查看权重、设备类别及状态等原生元数据；刷新会清除当前选择。" />
      <Row gutter={[24, 16]} style={{ width: '100%' }}>
        <Col xs={24} lg={12}>
          {data.nodes.length ? <Tree defaultExpandAll treeData={crushTree(data)} titleRender={(node) => <Space>{node.status !== undefined && <Tag color={crushStatus(node.status).color}>{crushStatus(node.status).label}</Tag>}<span>{node.title}</span></Space>} onSelect={(keys, info) => setSelected(keys.length ? data.nodes.find((node) => Number(node.id) === info.node.nodeId) ?? null : null)} /> : <Alert type="info" message="当前 CRUSH 树没有节点" />}
        </Col>
        <Col xs={24} lg={12}>
          {selected ? <Descriptions title={String(selected.name)} bordered column={1} items={Object.entries(selected).map(([key, value]) => ({ key, label: key, children: <span style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{value == null ? '未提供' : typeof value === 'object' ? JSON.stringify(value, null, 2) : String(value)}</span> }))} /> : <Alert type="info" message="选择节点查看详情" />}
        </Col>
      </Row>
    </Space>}
  </Card><CrushRulesPanel /></>
}

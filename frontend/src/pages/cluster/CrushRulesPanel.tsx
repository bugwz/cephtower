import { Alert, Card, Space } from 'antd'
import { DataTable } from '../../components/DataTable'
import type { ApiRecord } from '../../api/client'
import { ResourceListPage, type ResourceListPageDefinition } from '../ResourceListPage'

export function crushRuleType(value: unknown) {
  if (value === 1) return '复制 (1)'
  if (value === 3) return '纠删码 (3)'
  return typeof value === 'number' ? `类型 ${value}` : '未知'
}

export function crushRuleSteps(value: unknown): ApiRecord[] | null {
  if (!Array.isArray(value) || !value.every((step) => step !== null && typeof step === 'object' && !Array.isArray(step) && typeof step.op === 'string')) return null
  return value.map((step, index) => ({ ...step, sequence: index + 1 }))
}

const definition: ResourceListPageDefinition = {
  title: 'CRUSH 规则', path: '/crush/rules', rowKeyCandidates: ['natural_key', 'rule_id'],
  columns: [
    { key: 'rule_name', title: '规则名称' }, { key: 'rule_id', title: '规则 ID' },
    { key: 'type', title: '规则类型', render: crushRuleType },
    { key: 'min_size', title: '最小副本/分片数' }, { key: 'max_size', title: '最大副本/分片数' }
  ],
  detailContent: (row) => {
    const steps = crushRuleSteps(row.steps)
    return <Space direction="vertical" style={{ width: '100%' }}>
      <Alert type="info" message="步骤按原生命令返回顺序展示；num 为 0 等值保留原义，不转换为实际副本数。" />
      {steps ? <DataTable rowKeyCandidates={['sequence']} data={steps} columns={[{ key: 'sequence', title: '顺序' }, { key: 'op', title: '操作' }, { key: 'item', title: '节点 ID' }, { key: 'item_name', title: '节点名称' }, { key: 'type', title: '故障域' }, { key: 'num', title: '数量参数' }, { key: 'details', title: '原生步骤', render: (_value, step) => { const { sequence: _sequence, ...native } = step; return <pre style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{JSON.stringify(native, null, 2)}</pre> } }]} /> : <Alert type="warning" message="规则步骤未提供或格式异常" />}
    </Space>
  }
}

export function CrushRulesPanel() {
  return <Card style={{ marginTop: 16 }}><ResourceListPage definition={definition} embedded /></Card>
}

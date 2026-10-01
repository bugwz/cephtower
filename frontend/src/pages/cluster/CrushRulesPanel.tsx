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

export function crushRuleDeleteBlocked(row: ApiRecord) {
  if (row.stale !== false) return '规则库存过期或状态未知，请重新采集'
  if (typeof row.rule_name !== 'string' || !row.rule_name.trim()) return '规则名称不可用'
  return undefined
}

const definition: ResourceListPageDefinition = {
  title: 'CRUSH 规则', path: '/crush/rules', rowKeyCandidates: ['natural_key', 'rule_id'],
  columns: [
    { key: 'rule_name', title: '规则名称' }, { key: 'rule_id', title: '规则 ID' },
    { key: 'type', title: '规则类型', render: crushRuleType },
    { key: 'min_size', title: '最小副本/分片数' }, { key: 'max_size', title: '最大副本/分片数' }
  ],
  updateAction: {
    title: '重命名 CRUSH 规则', buttonLabel: '重命名', path: '/crush/rule', method: 'PATCH',
    successMessage: 'CRUSH 规则重命名已核验', disabledWhen: crushRuleDeleteBlocked,
    fields: [{ name: 'new_name', label: '新名称', required: true, pattern: /^[A-Za-z0-9_.-]+$/, patternMessage: '使用字母、数字、点、下划线或连字符' }],
    initialValues: (row) => ({ new_name: String(row?.rule_name ?? '') }),
    confirmation: (values, row) => `将规则 ${row?.rule_name} 重命名为 ${values.new_name}？请同步调整外部脚本中的名称引用。`,
    buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, name: row?.rule_name, new_name: values.new_name })
  },
  deleteAction: {
    title: '删除 CRUSH 规则', path: '/crush/rule', action: 'crush_rule.delete', resourceKind: 'crush_rule', risk: 'high',
    confirmation: (row) => `确认删除规则 ${row.rule_name}？此操作不可撤销。Ceph 会拒绝删除仍被存储池使用的规则；本操作不会迁移存储池或删除数据。`,
    successMessage: 'CRUSH 规则删除已核验', disabledWhen: crushRuleDeleteBlocked,
    buildBody: (row, clusterId) => ({ cluster_id: clusterId, name: row.rule_name }),
    resourceKey: (row) => `crush-rule/${row.rule_name}`
  },
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

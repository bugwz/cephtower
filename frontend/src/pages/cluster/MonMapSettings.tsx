import { Descriptions } from 'antd'
import type { ApiRecord } from '../../api/client'

export function monMapSetting(value: unknown): string {
  if (typeof value === 'boolean') return value ? '是' : '否'
  if (typeof value === 'number' && Number.isSafeInteger(value)) return String(value)
  if (typeof value === 'string') return value || '空（原生值）'
  return '未返回或格式无效'
}

export function MonMapSettings({ value }: { value?: ApiRecord | null }) {
  return <Descriptions title="monmap 配置（采集快照）" size="small" column={2} bordered items={[
    { key: 'min_mon_release', label: '最低 MON 版本编号', children: monMapSetting(value?.min_mon_release) },
    { key: 'min_mon_release_name', label: '最低 MON 版本名称', children: monMapSetting(value?.min_mon_release_name) },
    { key: 'election_strategy', label: '选举策略（原生编号）', children: monMapSetting(value?.election_strategy) },
    { key: 'stretch_mode', label: 'Stretch 模式', children: monMapSetting(value?.stretch_mode) },
    { key: 'tiebreaker_mon', label: 'Tiebreaker MON（配置）', children: monMapSetting(value?.tiebreaker_mon) },
    { key: 'disallowed_leaders', label: '禁止 Leader（原生文本）', children: monMapSetting(value?.disallowed_leaders) },
    { key: 'removed_ranks', label: '已移除 Rank（原生文本）', children: monMapSetting(value?.removed_ranks) }
  ]} />
}

import { Alert, Tabs } from 'antd'
import { bucketLifecycleRows } from './RgwBucketLifecycleRules'

export function RgwBucketLifecycleDocument({ document, rules, configured }: { document: unknown; rules: unknown; configured: unknown }) {
  if (configured === false) return <span>未配置生命周期，无配置文档</span>
  if (configured !== true || typeof document !== 'string' || !document.trim()) return <span>生命周期配置文档不可用</span>
  const parsed = bucketLifecycleRows(rules)
  // Only project the validated fields. This is a readable API representation,
  // not an XML round trip or a replacement document for S3 PUT requests.
  const structured = parsed?.length ? parsed.map(({ id, status, selector, actions }) => ({
    id, status,
    selector: {
      kind: selector.kind, and: selector.and, prefix: selector.prefix,
      tags: selector.tags.map(({ key, value }) => ({ key, value })),
      object_size_greater_than: selector.object_size_greater_than,
      object_size_less_than: selector.object_size_less_than,
      archive_zone: selector.archive_zone,
    },
    actions: actions.map(({ type, fields }) => ({ type, fields })),
  })) : undefined
  const style = { whiteSpace: 'pre-wrap', overflowWrap: 'anywhere', maxHeight: 360, overflow: 'auto' } as const
  return <Tabs defaultActiveKey="xml" items={[
    { key: 'xml', label: '原始 XML', children: <pre aria-label="生命周期原始 XML" style={style}>{document}</pre> },
    { key: 'json', label: '结构化 JSON', children: structured ? <>
      <Alert type="info" message="只读规则表示，不是 S3 写入文档" description="保留规则顺序、字符串数值、空值与布尔值。原始 XML 保持原样；此 JSON 不用于重新生成 XML 或提交配置。" />
      <pre aria-label="生命周期结构化 JSON" style={style}>{JSON.stringify(structured, null, 2)}</pre>
    </> : <Alert type="warning" message="结构化规则不可用，请查看原始 XML" /> },
  ]} />
}

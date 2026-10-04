import { Alert, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import type { ApiRecord } from '../api/client'
import type { MutationFormValues } from './ResourceListPage'

export type ResourceImportParser = (text: string, row?: ApiRecord) => MutationFormValues

export function ResourceFormImport({ parse, row, active, disabled, scope, values, apply }: {
  parse: ResourceImportParser; row?: ApiRecord; active: boolean; disabled: boolean; scope: number | undefined
  values: () => MutationFormValues; apply: (values: MutationFormValues) => void
}) {
  const request = useRef(0)
  const current = useRef({ parse, row, active, disabled, scope })
  current.current = { parse, row, active, disabled, scope }
  const [status, setStatus] = useState('')
  const [error, setError] = useState(false)
  useEffect(() => {
    request.current++
    setStatus('')
    return () => { request.current++ }
  }, [parse, row, active, disabled, scope])
  return <Space direction="vertical" style={{ width: '100%', marginBottom: 16 }}>
    <Alert type="info" message="导入 JSON / YAML 用户组资源（最大 1 MiB）" description="仅在浏览器中读取并替换当前表单，不自动提交。用户和组各最多 1000 项。文件包含密码，请妥善保存；导入后检查全部用户、组及集群绑定。" />
    <input aria-label="导入用户组资源文件" type="file" accept=".json,.yaml,.yml" disabled={disabled || !active} onChange={async event => {
      const file = event.target.files?.[0]
      event.target.value = ''
      const ticket = ++request.current
      if (!file || disabled || !active) return
      const snapshot = JSON.stringify(values())
      const valid = () => request.current === ticket && current.current.active && !current.current.disabled && current.current.parse === parse && current.current.row === row && current.current.scope === scope
      setError(false)
      setStatus('正在读取文件…')
      try {
        if (!/\.(json|ya?ml)$/i.test(file.name) || file.size > 1048576) throw new Error('invalid file')
        const text = await file.text()
        if (!valid()) return
        if (JSON.stringify(values()) !== snapshot) {
          setError(true); setStatus('表单已修改，未覆盖新输入；请重新选择文件。'); return
        }
        const imported = parse(text, row)
        apply(imported)
        setStatus('已载入表单，尚未提交。请检查后确认执行。')
      } catch {
        if (valid()) { setError(true); setStatus('导入失败：请检查文件大小、格式、资源类型、字段和编辑资源 ID。现有表单未修改。') }
      }
    }} />
    {status && <Alert type={error ? 'error' : 'info'} message={status} />}
  </Space>
}

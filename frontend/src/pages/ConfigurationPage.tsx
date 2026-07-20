import { Card, Input } from 'antd'
import { useCallback, useMemo, useState } from 'react'
import { textValue } from '../api/client'
import { listConfiguration } from '../api/resources'
import { DataTable } from '../components/DataTable'
import { Page } from '../components/Page'
import { useResource } from '../hooks'

export function ConfigurationPage() {
  const [keyword, setKeyword] = useState('')
  const loader = useCallback(() => listConfiguration(), [])
  const { data, loading, error, refresh } = useResource(loader)
  const filtered = useMemo(() => {
    const normalized = keyword.trim().toLowerCase()
    if (!normalized) {
      return data ?? []
    }

    return (data ?? []).filter((item) => JSON.stringify(item).toLowerCase().includes(normalized))
  }, [data, keyword])

  return (
    <Page
      title="配置中心"
      description="查看 Ceph 集群配置项，并为后续变更操作预留清晰入口。"
      loading={loading}
      error={error}
      onRefresh={refresh}
    >
      <Card
        title="集群配置"
        extra={
          <Input.Search
            allowClear
            className="table-search"
            placeholder="搜索配置项"
            value={keyword}
            onChange={(event) => setKeyword(event.target.value)}
          />
        }
      >
        <DataTable
          data={filtered}
          rowKeyCandidates={['name', 'section', 'daemon_default']}
          columns={[
            { key: 'name', title: '名称' },
            { key: 'section', title: 'Section' },
            { key: 'level', title: '级别' },
            { key: 'value', title: '当前值', render: (value) => textValue(value) },
            { key: 'daemon_default', title: '默认值' },
            { key: 'desc', title: '说明' }
          ]}
        />
      </Card>
    </Page>
  )
}


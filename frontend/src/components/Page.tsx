import { Alert, Card, Spin } from 'antd'

interface PageProps {
  title: string
  loading?: boolean
  error?: string
  stateVariant?: 'card' | 'inline'
  children: React.ReactNode
}

export function Page({ loading, error, stateVariant = 'card', children }: PageProps) {
  return (
    <>
      {loading ? (
        stateVariant === 'inline' ? (
          <div className="state-card state-card-inline">
            <Spin tip="正在加载..." />
          </div>
        ) : (
          <Card className="state-card">
            <Spin tip="正在加载..." />
          </Card>
        )
      ) : error ? (
        <Alert type="error" message="加载失败" description={error} showIcon />
      ) : (
        children
      )}
    </>
  )
}

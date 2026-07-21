import { Alert, Card, Spin } from 'antd'

interface PageProps {
  title: string
  loading?: boolean
  error?: string
  children: React.ReactNode
}

export function Page({ loading, error, children }: PageProps) {
  return (
    <>
      {loading ? (
        <Card className="state-card">
          <Spin tip="正在加载..." />
        </Card>
      ) : error ? (
        <Alert type="error" message="加载失败" description={error} showIcon />
      ) : (
        children
      )}
    </>
  )
}

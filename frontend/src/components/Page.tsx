import { Alert, Button, Card, Spin, Typography } from 'antd'
import { ReloadOutlined } from '@ant-design/icons'

const { Paragraph, Title } = Typography

interface PageProps {
  title: string
  description: string
  loading?: boolean
  error?: string
  onRefresh?: () => void
  children: React.ReactNode
}

export function Page({ title, description, loading, error, onRefresh, children }: PageProps) {
  return (
    <>
      <header className="page-header">
        <div>
          <Title level={1}>{title}</Title>
          <Paragraph>{description}</Paragraph>
        </div>
        {onRefresh && (
          <Button icon={<ReloadOutlined />} onClick={onRefresh}>
            刷新
          </Button>
        )}
      </header>
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


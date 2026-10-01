import { Progress, Space, Typography } from 'antd'
import { cloneProgress } from './cephfsCloneSummary'

export function CephFSCloneProgress({ report }: { report: unknown }) {
  const progress = cloneProgress(report)
  if (progress.percent === undefined && !progress.amount && !progress.files) return <>—</>
  return <Space direction="vertical" size={0} style={{ minWidth: 160 }}>
    {progress.percent !== undefined && <Progress percent={progress.percent} size="small" status="normal" format={() => `${progress.percent!.toFixed(2)}%`} />}
    {progress.amount && <Typography.Text>已克隆容量：{progress.amount}</Typography.Text>}
    {progress.files && <Typography.Text>已克隆文件：{progress.files}</Typography.Text>}
  </Space>
}

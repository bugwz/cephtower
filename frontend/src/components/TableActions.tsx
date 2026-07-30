import { Space, Typography } from 'antd'
import type { MouseEvent, ReactNode } from 'react'

const { Link } = Typography

interface TableActionProps {
  children: ReactNode
  danger?: boolean
  disabled?: boolean
  loading?: boolean
  onClick?: () => void
}

export function TableActions({ children }: { children: ReactNode }) {
  return (
    <Space size={12} className="table-actions">
      {children}
    </Space>
  )
}

export function TableAction({ children, danger, disabled, loading, onClick }: TableActionProps) {
  const blocked = disabled || loading

  function handleClick(event: MouseEvent<HTMLElement>) {
    event.preventDefault()
    if (blocked) {
      return
    }
    onClick?.()
  }

  return (
    <Link
      className={danger ? 'table-action table-action-danger' : 'table-action'}
      disabled={blocked}
      aria-busy={loading || undefined}
      onClick={handleClick}
    >
      {children}
    </Link>
  )
}

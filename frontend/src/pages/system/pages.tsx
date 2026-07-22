import { DemoPage } from '../DemoPage'
import { DataPage } from './DataPage'
import { UserPage } from './UserPage'

export function SystemInfoPage() {
  return <DemoPage pageKey="systemInfo" />
}

export function SystemUsersPage() {
  return <UserPage />
}

export function DataManagementPage() {
  return <DataPage />
}

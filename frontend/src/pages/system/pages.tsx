import { DemoPage } from '../DemoPage'
import { UserManagementPage } from './UserManagementPage'

export function SystemInfoPage() {
  return <DemoPage pageKey="systemInfo" />
}

export function SystemUsersPage() {
  return <UserManagementPage />
}

export function DataManagementPage() {
  return <DemoPage pageKey="dataManagement" />
}

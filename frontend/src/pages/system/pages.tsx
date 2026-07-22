import { DemoPage } from '../DemoPage'
import { DataFetchSettingsPage } from './DataFetchSettingsPage'
import { UserManagementPage } from './UserManagementPage'

export function SystemInfoPage() {
  return <DemoPage pageKey="systemInfo" />
}

export function SystemUsersPage() {
  return <UserManagementPage />
}

export function DataManagementPage() {
  return <DataFetchSettingsPage />
}

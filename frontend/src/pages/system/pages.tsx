import { DataPage } from './DataPage'
import { SystemInfoPage as SystemInfo } from './SystemInfoPage'
import { UserPage } from './UserPage'

export function SystemInfoPage() {
  return <SystemInfo />
}

export function SystemUsersPage() {
  return <UserPage />
}

export function DataManagementPage() {
  return <DataPage />
}

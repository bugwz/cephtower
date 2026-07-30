import { DataPage } from './DataPage'
import { SystemInfoPage as SystemInfo } from './SystemInfoPage'
import { UserPage } from './UserPage'

export function SystemInfoPage() {
  return <SystemInfo />
}

export function SystemUsersPage() {
  return <UserPage view="users" />
}

export function SystemRolesPage() {
  return <UserPage view="roles" />
}

export function SystemRoleBindingsPage() {
  return <UserPage view="bindings" />
}

export function DataManagementPage() {
  return <DataPage />
}

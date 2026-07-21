import { DemoPage } from '../DemoPage'
import { ClusterManagementPage } from './ClusterManagementPage'
import { HostManagementPage } from './HostManagementPage'
import { ServiceManagementPage } from './ServiceManagementPage'

export { ClusterManagementPage, HostManagementPage }

export function MonManagementPage() {
  return <DemoPage pageKey="monManagement" />
}

export function MgrManagementPage() {
  return <ServiceManagementPage />
}

export function OsdManagementPage() {
  return <DemoPage pageKey="osdManagement" />
}

export function MdsManagementPage() {
  return <DemoPage pageKey="mdsManagement" />
}

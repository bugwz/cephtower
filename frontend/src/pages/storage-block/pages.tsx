import { DemoPage } from '../DemoPage'
import { StorageManagementPage } from './StorageManagementPage'

export function BlockPoolsPage() {
  return <StorageManagementPage />
}

export function RbdImagesPage() {
  return <DemoPage pageKey="rbdImages" />
}

export function ImageMirroringPage() {
  return <DemoPage pageKey="imageMirroring" />
}

export function IscsiPage() {
  return <DemoPage pageKey="iscsi" />
}

export function NvmeTcpPage() {
  return <DemoPage pageKey="nvmeTcp" />
}

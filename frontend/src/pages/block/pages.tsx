import { StoragePlaceholderPage } from '../StoragePlaceholderPage'

export function BlockPoolsPage() {
  return <StoragePlaceholderPage pageKey="blockPools" />
}

export function RbdImagesPage() {
  return <StoragePlaceholderPage pageKey="rbdImages" />
}

export function ImageMirroringPage() {
  return <StoragePlaceholderPage pageKey="imageMirroring" />
}

export function IscsiPage() {
  return <StoragePlaceholderPage pageKey="iscsi" />
}

export function NvmeTcpPage() {
  return <StoragePlaceholderPage pageKey="nvmeTcp" />
}

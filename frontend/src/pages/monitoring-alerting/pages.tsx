import { DemoPage } from '../DemoPage'
import { RuntimeLogsPage as RuntimeLogsRealPage } from './RuntimeLogsPage'

export function MonitorOverviewPage() {
  return <DemoPage pageKey="monitorOverview" />
}

export function PerformanceMetricsPage() {
  return <DemoPage pageKey="performanceMetrics" />
}

export function RuntimeLogsPage() {
  return <RuntimeLogsRealPage />
}

export function AlertListPage() {
  return <DemoPage pageKey="alertList" />
}

export function AlertRulesPage() {
  return <DemoPage pageKey="alertRules" />
}

export function AlertSilencesPage() {
  return <DemoPage pageKey="alertSilences" />
}

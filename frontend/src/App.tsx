import { ConfigProvider, theme } from 'antd'
import { useMemo, useState } from 'react'
import { AppLayout } from './layout/AppLayout'
import {
  ClusterPage,
  ConfigurationPage,
  LogsPage,
  OverviewPage,
  ServicesPage,
  StoragePage,
  type PageKey
} from './pages'

export default function App() {
  const [activePage, setActivePage] = useState<PageKey>('overview')
  const page = useMemo(() => {
    switch (activePage) {
      case 'cluster':
        return <ClusterPage />
      case 'services':
        return <ServicesPage />
      case 'storage':
        return <StoragePage />
      case 'configuration':
        return <ConfigurationPage />
      case 'logs':
        return <LogsPage />
      default:
        return <OverviewPage />
    }
  }, [activePage])

  return (
    <ConfigProvider
      theme={{
        algorithm: theme.defaultAlgorithm,
        token: {
          colorPrimary: '#0f766e',
          borderRadius: 8,
          fontSize: 13,
          fontSizeHeading1: 24,
          fontSizeHeading2: 20,
          fontSizeHeading3: 17,
          controlHeight: 34,
          controlHeightLG: 38,
          controlHeightSM: 28,
          fontFamily:
            'Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif'
        },
        components: {
          Card: {
            borderRadiusLG: 8,
            headerFontSize: 14,
            headerHeight: 44
          },
          Menu: {
            itemHeight: 38,
            fontSize: 13
          },
          Table: {
            cellFontSize: 12,
            cellFontSizeMD: 12,
            headerBg: '#f7fafc'
          },
          Button: {
            fontWeight: 700
          }
        }
      }}
    >
      <AppLayout activePage={activePage} onPageChange={setActivePage}>
        {page}
      </AppLayout>
    </ConfigProvider>
  )
}

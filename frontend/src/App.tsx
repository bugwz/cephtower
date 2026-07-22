import { ConfigProvider, theme } from 'antd'
import type React from 'react'
import { useEffect, useState } from 'react'
import { Navigate, Route, Routes, useNavigate } from 'react-router-dom'
import { currentUser, hasStoredToken, logout, setupStatus, type SetupDatabaseConfig, type UserAccount } from './api/auth'
import { ApiErrorNotifier } from './components/ApiErrorNotifier'
import { AppLayout } from './layout/AppLayout'
import { NAV_PAGES, pagePaths } from './navigation'
import {
  InitializationPage,
  ClusterDetailPage,
  LoginPage,
  pageComponents,
  type PageKey
} from './pages'

export default function App() {
  const navigate = useNavigate()
  const [user, setUser] = useState<UserAccount | null>(null)
  const [checkingSession, setCheckingSession] = useState(true)
  const [setupRequired, setSetupRequired] = useState(false)
  const [setupDatabase, setSetupDatabase] = useState<SetupDatabaseConfig | undefined>()

  useEffect(() => {
    let cancelled = false
    async function bootstrap() {
      try {
        const status = await setupStatus()
        if (cancelled) {
          return
        }
        if (!status.initialized) {
          logout()
          setUser(null)
          setSetupDatabase(status.database)
          setSetupRequired(true)
          return
        }
        setSetupRequired(false)
        setSetupDatabase(undefined)
        if (!hasStoredToken()) {
          return
        }
        try {
          const account = await currentUser()
          if (!cancelled) {
            setUser(account)
          }
        } catch {
          logout()
        }
      } finally {
        if (!cancelled) {
          setCheckingSession(false)
        }
      }
    }

    bootstrap()

    return () => {
      cancelled = true
    }
  }, [])

  function handleLogin(account: UserAccount) {
    setUser(account)
    navigate(pagePaths.overview, { replace: true })
  }

  function handleLogout() {
    logout()
    setUser(null)
    navigate('/login', { replace: true })
  }

  function handleSetupComplete() {
    logout()
    setUser(null)
    setSetupRequired(false)
    setSetupDatabase(undefined)
    navigate('/login', { replace: true })
  }

  function renderAppPage(page: PageKey) {
    if (!user) {
      return <Navigate to="/login" replace />
    }

    return (
      <AppLayout activePage={page} onPageChange={(nextPage) => navigate(pagePaths[nextPage])} user={user} onLogout={handleLogout}>
        {renderPage(page)}
      </AppLayout>
    )
  }

  function renderStandaloneAppPage(activePage: PageKey, content: React.ReactNode) {
    if (!user) {
      return <Navigate to="/login" replace />
    }

    return (
      <AppLayout activePage={activePage} onPageChange={(nextPage) => navigate(pagePaths[nextPage])} user={user} onLogout={handleLogout}>
        {content}
      </AppLayout>
    )
  }

  return (
    <ConfigProvider
      theme={{
        algorithm: theme.defaultAlgorithm,
        token: {
          colorPrimary: '#43bf8f',
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
            headerBg: '#f3fbf8'
          },
          Button: {
            fontWeight: 700
          }
        }
      }}
    >
      <ApiErrorNotifier />
      {checkingSession ? (
        <div className="session-check" aria-label="正在检查系统状态" aria-busy="true" />
      ) : setupRequired ? (
        <Routes>
          <Route path="/initialize" element={<InitializationPage database={setupDatabase} onComplete={handleSetupComplete} />} />
          <Route path="*" element={<Navigate to="/initialize" replace />} />
        </Routes>
      ) : user ? (
        <Routes>
          <Route path="/" element={<Navigate to={pagePaths.overview} replace />} />
          {NAV_PAGES.map((page) => (
            <Route key={page.key} path={page.path} element={renderAppPage(page.key)} />
          ))}
          <Route path="/cluster/cluster/:id" element={renderStandaloneAppPage('clusterManagement', <ClusterDetailPage />)} />
          <Route path="/login" element={<Navigate to={pagePaths.overview} replace />} />
          <Route path="/initialize" element={<Navigate to={pagePaths.overview} replace />} />
          <Route path="/password-reset" element={<Navigate to={pagePaths.overview} replace />} />
          <Route path="*" element={<Navigate to={pagePaths.overview} replace />} />
        </Routes>
      ) : (
        <Routes>
          <Route
            path="/login"
            element={
              <LoginPage
                mode="login"
                onLogin={handleLogin}
                onForgotPassword={() => navigate('/password-reset')}
                onPasswordResetComplete={() => navigate('/login', { replace: true })}
              />
            }
          />
          <Route
            path="/password-reset"
            element={
              <LoginPage
                mode="reset"
                onLogin={handleLogin}
                onForgotPassword={() => navigate('/password-reset')}
                onPasswordResetComplete={() => navigate('/login', { replace: true })}
              />
            }
          />
          <Route path="/initialize" element={<Navigate to="/login" replace />} />
          <Route path="*" element={<Navigate to="/login" replace />} />
        </Routes>
      )}
    </ConfigProvider>
  )
}

function renderPage(page: PageKey) {
  const PageComponent = pageComponents[page]
  return <PageComponent />
}

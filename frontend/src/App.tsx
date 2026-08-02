import { ConfigProvider, theme } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import type React from 'react'
import { useCallback, useEffect, useRef, useState } from 'react'
import { Navigate, Route, Routes, useNavigate } from 'react-router-dom'
import { currentUser, hasStoredToken, logout, setupStatus, type SetupDatabaseConfig, type UserAccount } from './api/auth'
import { ApiErrorNotifier } from './components/ApiErrorNotifier'
import { AppLayout } from './layout/AppLayout'
import { NAV_PAGES, pagePaths } from './navigation'
import {
  InitializationPage,
  ClusterDetailPage,
  HostDetailPage,
  LoginPage,
  pageComponents,
  type PageKey
} from './pages'
import { ClusterProvider } from './state/ClusterContext'

export default function App() {
  const navigate = useNavigate()
  const [user, setUser] = useState<UserAccount | null>(null)
  const [checkingSession, setCheckingSession] = useState(true)
  const [setupRequired, setSetupRequired] = useState(false)
  const [setupDatabase, setSetupDatabase] = useState<SetupDatabaseConfig | undefined>()
  const handlingAuthenticationRequired = useRef(false)

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

  const handleAuthenticationRequired = useCallback(() => {
    if (handlingAuthenticationRequired.current) {
      return
    }

    handlingAuthenticationRequired.current = true
    logout()
    setUser(null)
    navigate('/login', { replace: true })
    window.setTimeout(() => {
      handlingAuthenticationRequired.current = false
    }, 0)
  }, [navigate])

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
      locale={zhCN}
      theme={{
        algorithm: theme.defaultAlgorithm,
        token: {
          colorPrimary: '#43bf8f',
          borderRadius: 8,
          fontSize: 13,
          fontSizeHeading1: 24,
          fontSizeHeading2: 20,
          fontSizeHeading3: 17,
          controlHeight: 32,
          controlHeightLG: 36,
          controlHeightSM: 26,
          fontFamily:
            '-apple-system, BlinkMacSystemFont, "PingFang SC", "Microsoft YaHei", "Noto Sans CJK SC", "Segoe UI", sans-serif'
        },
        components: {
          Card: {
            borderRadiusLG: 8,
            headerFontSize: 14,
            headerHeight: 44
          },
          Menu: {
            itemHeight: 38,
            fontSize: 13,
            itemBg: '#ffffff',
            itemHoverBg: '#f3f4f6',
            itemActiveBg: '#f3f4f6',
            itemSelectedBg: '#d5efe3'
          },
          Table: {
            cellFontSize: 12,
            cellFontSizeMD: 12,
            headerBg: '#f3fbf8'
          },
          Button: {
            defaultBg: '#fbfefd',
            defaultBorderColor: '#cfe7df',
            defaultColor: '#35515a',
            defaultHoverBg: '#f0faf5',
            defaultHoverBorderColor: '#9bdcc4',
            defaultHoverColor: '#168766',
            defaultActiveBg: '#dff5ec',
            defaultActiveBorderColor: '#7fd0b4',
            defaultActiveColor: '#0c604f',
            primaryColor: '#0c604f',
            defaultShadow: '0 3px 10px rgb(25 88 94 / 6%)',
            primaryShadow: '0 4px 12px rgb(35 132 103 / 10%)',
            dangerShadow: '0 3px 10px rgb(150 54 64 / 6%)',
            contentFontSize: 12,
            contentFontSizeSM: 11,
            fontWeight: 500,
            paddingInline: 14,
            paddingInlineSM: 9
          }
        }
      }}
    >
      <ApiErrorNotifier onAuthenticationRequired={handleAuthenticationRequired} />
      <ClusterProvider enabled={Boolean(user)}>
        {checkingSession ? (
          <div className="session-check" aria-label="正在检查系统状态" aria-busy="true" />
        ) : setupRequired ? (
          <Routes>
            <Route path="/bootstrap" element={<InitializationPage database={setupDatabase} onComplete={handleSetupComplete} />} />
            <Route path="*" element={<Navigate to="/bootstrap" replace />} />
          </Routes>
        ) : user ? (
          <Routes>
            <Route path="/" element={<Navigate to={pagePaths.overview} replace />} />
            {NAV_PAGES.map((page) => (
              <Route key={page.key} path={page.path} element={renderAppPage(page.key)} />
            ))}
            <Route path="/cluster/cluster/:name" element={renderStandaloneAppPage('clusterManagement', <ClusterDetailPage />)} />
            <Route path="/cluster/host/:name" element={renderStandaloneAppPage('hostManagement', <HostDetailPage />)} />
            <Route path="/login" element={<Navigate to={pagePaths.overview} replace />} />
            <Route path="/bootstrap" element={<Navigate to={pagePaths.overview} replace />} />
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
            <Route path="/bootstrap" element={<Navigate to="/login" replace />} />
            <Route path="*" element={<Navigate to="/login" replace />} />
          </Routes>
        )}
      </ClusterProvider>
    </ConfigProvider>
  )
}

function renderPage(page: PageKey) {
  const PageComponent = pageComponents[page]
  return <PageComponent />
}

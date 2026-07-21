import { KeyOutlined, LockOutlined, MailOutlined, UserOutlined } from '@ant-design/icons'
import { Button, Checkbox, Form, Input, Typography, message } from 'antd'
import { useEffect, useState } from 'react'
import { confirmPasswordReset, login, requestPasswordReset, type UserAccount } from '../api/auth'

const { Text, Title } = Typography

interface LoginPageProps {
  mode: 'login' | 'reset'
  onLogin: (user: UserAccount) => void
  onForgotPassword: () => void
  onPasswordResetComplete: () => void
}

export function LoginPage({ mode, onLogin, onForgotPassword, onPasswordResetComplete }: LoginPageProps) {
  const [loading, setLoading] = useState(false)
  const [resetLoading, setResetLoading] = useState(false)
  const [resetCountdown, setResetCountdown] = useState(0)
  const [resetForm] = Form.useForm()

  useEffect(() => {
    if (resetCountdown <= 0) {
      return undefined
    }

    const timer = window.setTimeout(() => {
      setResetCountdown((current) => Math.max(current - 1, 0))
    }, 1000)

    return () => window.clearTimeout(timer)
  }, [resetCountdown])

  async function submit(values: { username: string; password: string }) {
    setLoading(true)
    try {
      const response = await login(values.username, values.password)
      onLogin(response.user)
    } catch {
      // API errors are shown by the global notifier.
    } finally {
      setLoading(false)
    }
  }

  async function sendCode() {
    setResetLoading(true)
    try {
      const { account } = await resetForm.validateFields(['account'])
      const response = await requestPasswordReset(account)
      message.success(response.message)
      setResetCountdown(90)
    } catch {
      // Field validation stays inline; API errors are shown by the global notifier.
    } finally {
      setResetLoading(false)
    }
  }

  async function resetPassword(values: { account: string; code: string; new_password: string; confirm_password: string }) {
    setLoading(true)
    try {
      const response = await confirmPasswordReset({
        account: values.account,
        code: values.code,
        new_password: values.new_password
      })
      message.success(response.message)
      resetForm.resetFields()
      setResetCountdown(0)
      onPasswordResetComplete()
    } catch {
      // API errors are shown by the global notifier.
    } finally {
      setLoading(false)
    }
  }

  return (
    <main className="login-shell">
      <section className="login-panel login-visual">
        <div className="login-brand">
          <img src="/ceph-tower-logo.svg" alt="CephTower logo" />
          <span>CephTower</span>
        </div>
        <div className="login-visual-main">
          <TowerIllustration />
          <div className="login-visual-copy">
            <Title level={1}>统一集群运维入口</Title>
            <Text>面向 Ceph 资源、服务状态和系统账号的控制台。</Text>
          </div>
        </div>
      </section>
      <section className="login-panel login-form-panel">
        <div className="login-card">
          <Title level={2}>{mode === 'login' ? '登录' : '忘记密码'}</Title>
          {mode === 'login' ? (
            <Form layout="vertical" onFinish={submit} className="login-form">
              <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
                <Input prefix={<UserOutlined />} placeholder="请输入用户名" autoComplete="username" />
              </Form.Item>
              <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
                <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" autoComplete="current-password" />
              </Form.Item>
              <div className="login-options">
                <Checkbox>记住我</Checkbox>
                <Button
                  type="link"
                  className="forgot-link"
                  onClick={onForgotPassword}
                >
                  忘记密码？
                </Button>
              </div>
              <Button type="primary" htmlType="submit" loading={loading} block>
                登录
              </Button>
            </Form>
          ) : (
            <Form form={resetForm} layout="vertical" onFinish={resetPassword} className="login-form">
              <Form.Item name="account" rules={[{ required: true, message: '请输入用户名或邮箱' }]}>
                <Input prefix={<MailOutlined />} placeholder="请输入用户名或邮箱" autoComplete="email" />
              </Form.Item>
              <div className="reset-code-row">
                <Form.Item name="code" rules={[{ required: true, message: '请输入邮箱验证码' }]}>
                  <Input prefix={<KeyOutlined />} placeholder="请输入邮箱验证码" autoComplete="one-time-code" />
                </Form.Item>
                <Form.Item noStyle>
                  <Button
                    className="send-code-button"
                    icon={<MailOutlined />}
                    loading={resetLoading}
                    disabled={resetCountdown > 0}
                    onClick={sendCode}
                  >
                    {resetCountdown > 0 ? `${resetCountdown}s 后重发` : '发送验证码'}
                  </Button>
                </Form.Item>
              </div>
              <Form.Item name="new_password" rules={[{ required: true, min: 8, message: '请输入至少 8 位新密码' }]}>
                <Input.Password prefix={<LockOutlined />} placeholder="请输入新密码" autoComplete="new-password" />
              </Form.Item>
              <Form.Item
                name="confirm_password"
                dependencies={['new_password']}
                rules={[
                  { required: true, message: '请再次输入新密码' },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (!value || getFieldValue('new_password') === value) {
                        return Promise.resolve()
                      }
                      return Promise.reject(new Error('两次输入的密码不一致'))
                    }
                  })
                ]}
              >
                <Input.Password prefix={<LockOutlined />} placeholder="请确认新密码" autoComplete="new-password" />
              </Form.Item>
              <Button type="primary" htmlType="submit" loading={loading} block>
                重设密码
              </Button>
            </Form>
          )}
        </div>
      </section>
    </main>
  )
}

export function TowerIllustration() {
  return (
    <svg className="tower-illustration" viewBox="0 0 520 360" role="img" aria-label="CephTower cluster illustration">
      <defs>
        <linearGradient id="towerPanel" x1="0" x2="1" y1="0" y2="1">
          <stop offset="0" stopColor="#f7fbff" />
          <stop offset="1" stopColor="#cfe3ec" />
        </linearGradient>
        <linearGradient id="towerAccent" x1="0" x2="1">
          <stop offset="0" stopColor="#5edb91" />
          <stop offset="1" stopColor="#1b9186" />
        </linearGradient>
      </defs>
      <rect x="56" y="286" width="408" height="18" rx="9" fill="#7ea8b2" opacity="0.18" />
      <path d="M170 286h180l-26-174H196z" fill="#d9edf3" />
      <path d="M206 136h108l18 128H188z" fill="url(#towerPanel)" />
      <path d="M228 88h64l22 48H206z" fill="#c8e3ea" />
      <path d="M238 64h44l10 24h-64z" fill="#5edb91" />
      <path d="M232 168h56M226 204h72M220 240h84" stroke="#1b8178" strokeWidth="10" strokeLinecap="round" />
      <circle cx="260" cy="96" r="16" fill="#e7f7ef" />
      <circle cx="260" cy="96" r="7" fill="#1b8178" />
      <path d="M260 56v-34M226 96h-40M294 96h40" stroke="#6fb7c2" strokeWidth="8" strokeLinecap="round" />
      <circle cx="168" cy="96" r="18" fill="#e6f7ee" />
      <circle cx="352" cy="96" r="18" fill="#e6f7ee" />
      <circle cx="260" cy="22" r="18" fill="#e6f7ee" />
      <path d="M156 96h24M340 96h24M260 10v24" stroke="url(#towerAccent)" strokeWidth="7" strokeLinecap="round" />
      <rect x="78" y="246" width="90" height="58" rx="8" fill="#e9f3f7" />
      <rect x="92" y="262" width="62" height="9" rx="4" fill="#5edb91" />
      <rect x="92" y="280" width="42" height="9" rx="4" fill="#9fb2ba" />
      <rect x="352" y="224" width="94" height="80" rx="8" fill="#e9f3f7" />
      <rect x="368" y="240" width="62" height="9" rx="4" fill="#1b8178" />
      <rect x="368" y="260" width="42" height="9" rx="4" fill="#9fb2ba" />
      <rect x="368" y="280" width="54" height="9" rx="4" fill="#5edb91" />
    </svg>
  )
}

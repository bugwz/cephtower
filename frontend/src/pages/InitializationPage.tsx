import { LockOutlined, MailOutlined, UserOutlined } from '@ant-design/icons'
import { Alert, Button, Form, Input, InputNumber, Radio, Steps, Typography, message } from 'antd'
import { useMemo, useState } from 'react'
import { initializeSetup, type SetupDatabaseConfig } from '../api/auth'
import { TowerIllustration } from './LoginPage'

const { Text, Title } = Typography

interface InitializationPageProps {
  database?: SetupDatabaseConfig
  onComplete: () => void
}

interface InitializationFormValues {
  engine: 'sqlite' | 'mysql'
  sqlite_path: string
  mysql_host: string
  mysql_port: number
  mysql_username: string
  mysql_password: string
  mysql_database: string
  mysql_params: string
  admin_username: string
  admin_email: string
  admin_password: string
  admin_confirm_password: string
}

export function InitializationPage({ database, onComplete }: InitializationPageProps) {
  const [form] = Form.useForm<InitializationFormValues>()
  const [step, setStep] = useState(0)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const initialValues = useMemo(
    () => ({
      engine: database?.engine ?? 'sqlite',
      sqlite_path: database?.sqlite.path ?? 'data/cephtower.db',
      mysql_host: database?.mysql.host ?? '127.0.0.1',
      mysql_port: database?.mysql.port ?? 3306,
      mysql_username: database?.mysql.username ?? 'cephtower',
      mysql_password: database?.mysql.password ?? '',
      mysql_database: database?.mysql.database ?? 'cephtower',
      mysql_params: database?.mysql.params ?? 'charset=utf8mb4&parseTime=True&loc=Local',
      admin_username: 'admin'
    }),
    [database]
  )
  const engine = Form.useWatch('engine', form) ?? initialValues.engine

  async function nextStep() {
    setError('')
    const fields =
      step === 0
        ? engine === 'sqlite'
          ? ['engine', 'sqlite_path']
          : ['engine', 'mysql_host', 'mysql_port', 'mysql_username', 'mysql_password', 'mysql_database', 'mysql_params']
        : ['admin_username', 'admin_email', 'admin_password', 'admin_confirm_password']
    await form.validateFields(fields as Array<keyof InitializationFormValues>)
    setStep((current) => current + 1)
  }

  async function submit() {
    setLoading(true)
    setError('')
    try {
      const values = await form.validateFields()
      await initializeSetup({
        database: {
          engine: values.engine,
          sqlite: {
            path: values.sqlite_path
          },
          mysql: {
            host: values.mysql_host,
            port: values.mysql_port,
            username: values.mysql_username,
            password: values.mysql_password,
            database: values.mysql_database,
            params: values.mysql_params
          }
        },
        admin: {
          username: values.admin_username,
          email: values.admin_email,
          password: values.admin_password
        }
      })
      message.success('初始化完成，请登录')
      onComplete()
    } catch (err) {
      setError(err instanceof Error ? err.message : '初始化失败')
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
        <div className="login-card setup-card">
          <Title level={2}>系统初始化</Title>
          <Steps
            size="small"
            current={step}
            items={[{ title: '数据库' }, { title: '管理员账户' }, { title: '确认' }]}
          />
          {error && <Alert type="error" showIcon message={error} />}
          <Form form={form} layout="vertical" initialValues={initialValues} className="login-form setup-form">
            <div hidden={step !== 0}>
              <Form.Item name="engine" label="数据库类型" rules={[{ required: true, message: '请选择数据库类型' }]}>
                <Radio.Group optionType="button" buttonStyle="solid">
                  <Radio.Button value="sqlite">SQLite</Radio.Button>
                  <Radio.Button value="mysql">MySQL</Radio.Button>
                </Radio.Group>
              </Form.Item>
              {engine === 'sqlite' ? (
                <Form.Item name="sqlite_path" label="数据库路径" rules={[{ required: true, message: '请输入数据库路径' }]}>
                  <Input placeholder="data/cephtower.db" />
                </Form.Item>
              ) : (
                <>
                  <div className="setup-grid">
                    <Form.Item name="mysql_host" label="主机" rules={[{ required: true, message: '请输入主机地址' }]}>
                      <Input placeholder="127.0.0.1" />
                    </Form.Item>
                    <Form.Item name="mysql_port" label="端口" rules={[{ required: true, message: '请输入端口' }]}>
                      <InputNumber min={1} max={65535} controls={false} />
                    </Form.Item>
                  </div>
                  <div className="setup-grid">
                    <Form.Item name="mysql_username" label="用户名" rules={[{ required: true, message: '请输入数据库用户名' }]}>
                      <Input autoComplete="username" />
                    </Form.Item>
                    <Form.Item name="mysql_password" label="密码" rules={[{ required: true, message: '请输入数据库密码' }]}>
                      <Input.Password autoComplete="new-password" />
                    </Form.Item>
                  </div>
                  <Form.Item name="mysql_database" label="数据库名" rules={[{ required: true, message: '请输入数据库名' }]}>
                    <Input />
                  </Form.Item>
                  <Form.Item name="mysql_params" label="连接参数" rules={[{ required: true, message: '请输入连接参数' }]}>
                    <Input />
                  </Form.Item>
                </>
              )}
            </div>
            <div hidden={step !== 1}>
              <Form.Item name="admin_username" label="默认管理员用户名" rules={[{ required: true, message: '请输入用户名' }]}>
                <Input prefix={<UserOutlined />} autoComplete="username" />
              </Form.Item>
              <Form.Item
                name="admin_email"
                label="管理员邮箱"
                rules={[
                  { required: true, message: '请输入邮箱' },
                  { type: 'email', message: '请输入有效邮箱' }
                ]}
              >
                <Input prefix={<MailOutlined />} autoComplete="email" />
              </Form.Item>
              <Form.Item name="admin_password" label="管理员密码" rules={[{ required: true, min: 8, message: '请输入至少 8 位密码' }]}>
                <Input.Password prefix={<LockOutlined />} autoComplete="new-password" />
              </Form.Item>
              <Form.Item
                name="admin_confirm_password"
                label="确认密码"
                dependencies={['admin_password']}
                rules={[
                  { required: true, message: '请再次输入密码' },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (!value || getFieldValue('admin_password') === value) {
                        return Promise.resolve()
                      }
                      return Promise.reject(new Error('两次输入的密码不一致'))
                    }
                  })
                ]}
              >
                <Input.Password prefix={<LockOutlined />} autoComplete="new-password" />
              </Form.Item>
            </div>
            <div hidden={step !== 2} className="setup-summary">
              <div className="setup-summary-section">
                <Text strong>数据库配置</Text>
                <div className="setup-summary-list">
                  {setupSummaryItem('类型', engine === 'sqlite' ? 'SQLite' : 'MySQL')}
                  {engine === 'sqlite' ? (
                    setupSummaryItem('数据库路径', form.getFieldValue('sqlite_path'))
                  ) : (
                    <>
                      {setupSummaryItem('主机', form.getFieldValue('mysql_host'))}
                      {setupSummaryItem('端口', form.getFieldValue('mysql_port'))}
                      {setupSummaryItem('用户名', form.getFieldValue('mysql_username'))}
                      {setupSummaryItem('密码', form.getFieldValue('mysql_password'))}
                      {setupSummaryItem('数据库名', form.getFieldValue('mysql_database'))}
                      {setupSummaryItem('连接参数', form.getFieldValue('mysql_params'))}
                    </>
                  )}
                </div>
              </div>
              <div className="setup-summary-section">
                <Text strong>管理员账户</Text>
                <div className="setup-summary-list">
                  {setupSummaryItem('用户名', form.getFieldValue('admin_username'))}
                  {setupSummaryItem('邮箱', form.getFieldValue('admin_email'))}
                </div>
              </div>
            </div>
            <div className="setup-actions">
              {step > 0 && (
                <Button onClick={() => setStep((current) => current - 1)} disabled={loading}>
                  上一步
                </Button>
              )}
              {step < 2 ? (
                <Button type="primary" onClick={nextStep}>
                  下一步
                </Button>
              ) : (
                <Button type="primary" loading={loading} onClick={submit}>
                  提交初始化
                </Button>
              )}
            </div>
          </Form>
        </div>
      </section>
    </main>
  )
}

function setupSummaryItem(label: string, value: unknown) {
  return (
    <div className="setup-summary-item">
      <Text type="secondary">{label}</Text>
      <Text>{value == null || value === '' ? '-' : String(value)}</Text>
    </div>
  )
}

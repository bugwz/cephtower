import { LockOutlined, MailOutlined, UserOutlined } from '@ant-design/icons'
import { Alert, Button, Form, Input, InputNumber, Radio, Select, Steps, Typography } from 'antd'
import { useMemo, useRef, useState } from 'react'
import { initializeSetup, testSetupDatabase, type SetupDatabaseConfig, type SetupDatabaseInput } from '../api/auth'
import { message } from '../utils/appMessage'
import { friendlyAuthError, isFormValidationError } from '../utils/friendlyAuthError'
import { TowerIllustration } from './LoginPage'

const { Text, Title } = Typography

interface InitializationPageProps {
  database?: SetupDatabaseConfig
  onComplete: () => void
}

interface InitializationFormValues {
  engine: 'sqlite' | 'mysql'
  sqlite_name: string
  mysql_host: string
  mysql_port: number
  mysql_username: string
  mysql_password: string
  mysql_database: string
  mysql_params: string
  mysql_tls: 'false' | 'true' | 'skip-verify' | 'preferred'
  admin_username: string
  admin_email: string
  admin_password: string
  admin_confirm_password: string
}

export function InitializationPage({ database, onComplete }: InitializationPageProps) {
  const [form] = Form.useForm<InitializationFormValues>()
  const [step, setStep] = useState(0)
  const [loading, setLoading] = useState(false)
  const [testingDatabase, setTestingDatabase] = useState(false)
  const databaseTestRunning = useRef(false)
  const [error, setError] = useState('')
  const initialValues = useMemo(
    () => ({
      engine: database?.engine ?? 'sqlite',
      sqlite_name: database?.sqlite.name ?? 'cephtower.db',
      mysql_host: database?.mysql.host ?? '127.0.0.1',
      mysql_port: database?.mysql.port ?? 3306,
      mysql_username: database?.mysql.username ?? 'root',
      mysql_password: database?.mysql.password ?? '',
      mysql_database: database?.mysql.database ?? 'cephtower',
      mysql_params: database?.mysql.params ?? 'charset=utf8mb4&parseTime=True&loc=Local',
      mysql_tls: database?.mysql.tls ?? 'false',
      admin_username: 'admin',
      admin_email: 'admin@admin.com'
    }),
    [database]
  )
  const engine = Form.useWatch('engine', form) ?? initialValues.engine

  async function nextStep() {
    setError('')
    const fields = step === 0 ? databaseFields(engine) : ['admin_username', 'admin_email', 'admin_password', 'admin_confirm_password']
    await form.validateFields(fields as Array<keyof InitializationFormValues>)
    setStep((current) => current + 1)
  }

  async function testDatabase() {
    if (databaseTestRunning.current) {
      return
    }
    databaseTestRunning.current = true
    setTestingDatabase(true)
    setError('')
    try {
      await form.validateFields(databaseFields(engine))
      await testSetupDatabase(databaseInput(form.getFieldsValue(), engine))
      message.success('数据库连接成功')
    } catch (err) {
      if (!isFormValidationError(err)) {
        message.error(friendlyAuthError(err, 'testDatabase', { engine }))
      }
    } finally {
      databaseTestRunning.current = false
      setTestingDatabase(false)
    }
  }

  async function submit() {
    setLoading(true)
    setError('')
    let submittedEngine = engine
    try {
      const values = await form.validateFields()
      submittedEngine = values.engine
      await initializeSetup({
        database: databaseInput(values, values.engine),
        user: {
          username: values.admin_username,
          email: values.admin_email,
          password: values.admin_password
        }
      })
      message.success('初始化完成！')
      onComplete()
    } catch (err) {
      if (!isFormValidationError(err)) {
        setError(friendlyAuthError(err, 'initialize', { engine: submittedEngine }))
      }
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
                <Form.Item name="sqlite_name" label="数据库文件" rules={[{ required: true, message: '请输入数据库文件' }]}>
                  <Input placeholder="cephtower.db" />
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
                  <div className="setup-grid">
                    <Form.Item name="mysql_database" label="数据库名" rules={[{ required: true, message: '请输入数据库名' }]}>
                      <Input />
                    </Form.Item>
                    <Form.Item name="mysql_tls" label="TLS 模式" rules={[{ required: true, message: '请选择 TLS 模式' }]}>
                      <Select
                        options={[
                          { value: 'false', label: '禁用' },
                          { value: 'true', label: '启用并验证证书' },
                          { value: 'skip-verify', label: '启用但不验证证书' },
                          { value: 'preferred', label: '优先 TLS，允许回退' }
                        ]}
                      />
                    </Form.Item>
                  </div>
                  <Form.Item name="mysql_params" label="连接参数" rules={[{ required: true, message: '请输入连接参数' }]}>
                    <Input />
                  </Form.Item>
                </>
              )}
            </div>
            <div hidden={step !== 1}>
              <Form.Item name="admin_username" label="管理员用户名" rules={[{ required: true, message: '请输入用户名' }]}>
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
                    setupSummaryItem('数据库文件', form.getFieldValue('sqlite_name'))
                  ) : (
                    <>
                      {setupSummaryItem('主机', form.getFieldValue('mysql_host'))}
                      {setupSummaryItem('端口', form.getFieldValue('mysql_port'))}
                      {setupSummaryItem('用户名', form.getFieldValue('mysql_username'))}
                      {setupSummaryItem('密码', form.getFieldValue('mysql_password'))}
                      {setupSummaryItem('数据库名', form.getFieldValue('mysql_database'))}
                      {setupSummaryItem('连接参数', form.getFieldValue('mysql_params'))}
                      {setupSummaryItem('TLS 模式', mysqlTLSLabel(form.getFieldValue('mysql_tls')))}
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
              {step === 0 && (
                <Button loading={testingDatabase} onClick={testDatabase} disabled={loading || testingDatabase}>
                  检测
                </Button>
              )}
              {step < 2 ? (
                <Button type="primary" onClick={nextStep} disabled={testingDatabase}>
                  下一步
                </Button>
              ) : (
                <Button type="primary" loading={loading} onClick={submit}>
                  初始化
                </Button>
              )}
            </div>
          </Form>
        </div>
      </section>
    </main>
  )
}

function databaseFields(engine: InitializationFormValues['engine']): Array<keyof InitializationFormValues> {
  return engine === 'sqlite'
    ? ['engine', 'sqlite_name']
    : ['engine', 'mysql_host', 'mysql_port', 'mysql_username', 'mysql_password', 'mysql_database', 'mysql_params', 'mysql_tls']
}

function databaseInput(values: Partial<InitializationFormValues>, engine: InitializationFormValues['engine']): SetupDatabaseInput {
  return {
    engine,
    sqlite: { name: values.sqlite_name ?? '' },
    mysql: {
      host: values.mysql_host ?? '',
      port: values.mysql_port ?? 0,
      username: values.mysql_username ?? '',
      password: values.mysql_password ?? '',
      database: values.mysql_database ?? '',
      params: values.mysql_params ?? '',
      tls: values.mysql_tls ?? 'false'
    }
  }
}

function mysqlTLSLabel(value: InitializationFormValues['mysql_tls']) {
  return {
    false: '禁用',
    true: '启用并验证证书',
    'skip-verify': '启用但不验证证书',
    preferred: '优先 TLS，允许回退'
  }[value]
}

function setupSummaryItem(label: string, value: unknown) {
  return (
    <div className="setup-summary-item">
      <Text type="secondary">{label}</Text>
      <Text>{value == null || value === '' ? '-' : String(value)}</Text>
    </div>
  )
}

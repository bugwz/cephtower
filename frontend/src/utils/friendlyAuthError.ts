type AuthErrorContext = 'login' | 'sendResetCode' | 'resetPassword' | 'testDatabase' | 'initialize'

interface FriendlyErrorOptions {
  engine?: 'sqlite' | 'mysql'
}

export function friendlyAuthError(err: unknown, context: AuthErrorContext, options: FriendlyErrorOptions = {}) {
  if (isNetworkError(err)) {
    return '无法连接后端，请稍后重试。'
  }

  const text = err instanceof Error ? err.message.toLowerCase() : ''

  switch (context) {
    case 'login':
      if (text.includes('disabled')) {
        return '账号已禁用，请联系管理员。'
      }
      return '用户名或密码不正确。'
    case 'sendResetCode':
      if (text.includes('email')) {
        return '账号未绑定邮箱，请联系管理员。'
      }
      if (text.includes('not_ready') || text.includes('database')) {
        return '系统未初始化，暂无法发送验证码。'
      }
      return '验证码发送失败，请检查账号。'
    case 'resetPassword':
      if (text.includes('code') || text.includes('expired') || text.includes('验证码')) {
        return '验证码无效或已过期。'
      }
      if (text.includes('password')) {
        return '新密码至少需要 8 位。'
      }
      return '重置失败，请检查验证码和密码。'
    case 'testDatabase':
      return databaseErrorMessage(text, options.engine, '数据库检测未通过，请检查配置。')
    case 'initialize':
      if (text.includes('already_initialized')) {
        return '系统已初始化，无需重复操作。'
      }
      return databaseErrorMessage(text, options.engine, '初始化失败，请检查配置。')
  }
}

export function isFormValidationError(err: unknown) {
  return typeof err === 'object' && err !== null && 'errorFields' in err
}

function databaseErrorMessage(text: string, engine: FriendlyErrorOptions['engine'], fallback: string) {
  if (text.includes('already exists')) {
    return engine === 'sqlite'
      ? 'SQLite 文件已存在，请更换名称。'
      : 'MySQL 数据库已存在，请更换名称。'
  }
  if (text.includes('is a directory') || text.includes('not a directory')) {
    return 'SQLite 文件名或目录不正确。'
  }
  if (text.includes('not writable') || text.includes('cannot be created') || text.includes('permission') || text.includes('access denied')) {
    return engine === 'mysql'
      ? 'MySQL 账号缺少建库权限。'
      : 'SQLite 数据目录不可写。'
  }
  if (text.includes('ping mysql server') || text.includes('connection refused') || text.includes('connect') || text.includes('timeout')) {
    return '无法连接 MySQL，请检查连接信息。'
  }
  if (text.includes('unsupported database engine')) {
    return '请选择 SQLite 或 MySQL。'
  }
  if (text.includes('password') && text.includes('8')) {
    return '管理员密码至少需要 8 位。'
  }
  return fallback
}

function isNetworkError(err: unknown) {
  return err instanceof TypeError && /failed to fetch|networkerror|load failed/i.test(err.message)
}

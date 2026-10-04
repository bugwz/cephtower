import type { ResourceFormAction } from '../ResourceListPage'

export const rgwRealmImportAction: ResourceFormAction = {
  title: '导入 Realm Token 并部署普通从 Zone',
  buttonLabel: '导入 Realm Token',
  path: '/rgw/realm/import', method: 'POST',
  successMessage: '从 Zone 和 Period 已核验，RGW 部署已提交；不代表守护进程就绪或同步完成',
  confirmation: () => '此操作会联系 Token 中的远端地址，使用系统密钥拉取 Realm、创建普通从 Zone、发布整个 Period，并提交 RGW 部署。请核对可信来源；HTTP 端点不保证传输安全。操作非事务，失败可能部分生效，不自动重试；Ceph 内部可能重试 Period。请先备份并协调两端管理员。本入口尚不支持归档 Zone，不能用于替换已有 Zone。确认执行？',
  initialValues: { placement_mode: 'default', port: 80 },
  changedValues: changed => 'confirm_import' in changed ? {} : ({ confirm_import: undefined }),
  fields: [
    { name: 'realm_token', label: '可信主站点的 Realm Token（含系统密钥）', type: 'password', required: true },
    { name: 'name', label: '新的普通从 Zone 名称', required: true, pattern: /^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$/ },
    { name: 'port', label: 'RGW 前端端口（非 TLS）', type: 'number', min: 1, max: 65535, required: true },
    { name: 'placement_mode', label: '部署位置', type: 'select', required: true, options: [{ value: 'default', label: 'Ceph 默认位置' }, { value: 'hosts', label: '指定主机' }, { value: 'label', label: '指定标签' }] },
    { name: 'hosts', label: '主机名（每行一个）', type: 'textarea', required: true, visibleWhen: v => v.placement_mode === 'hosts' },
    { name: 'label', label: '主机标签', required: true, visibleWhen: v => v.placement_mode === 'label' },
    { name: 'count', label: '部署实例数（留空使用 Ceph 默认值）', type: 'number', min: 1 },
    { name: 'confirm_import', label: '副作用确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已备份并核对可信端点，了解发布、部署和部分生效风险' }] }
  ],
  buildBody: (v, clusterId) => {
    if (v.confirm_import !== 'acknowledged') throw new Error('请确认导入风险')
    if (typeof v.realm_token !== 'string' || !v.realm_token || v.realm_token.length > 65536 || !/^[A-Za-z0-9+/]+={0,2}$/.test(v.realm_token)) throw new Error('请输入规范的 Realm Token')
    if (typeof v.name !== 'string' || !/^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$/.test(v.name)) throw new Error('Zone 名称无效')
    if (typeof v.port !== 'number' || !Number.isInteger(v.port) || v.port < 1 || v.port > 65535) throw new Error('端口必须为 1–65535 的整数')
    const placement: Record<string, unknown> = {}
    if (v.placement_mode === 'hosts') {
      const hosts = String(v.hosts ?? '').split(/\r?\n/).map(x => x.trim()).filter(Boolean)
      if (!hosts.length || hosts.some(x => !/^[A-Za-z0-9_][A-Za-z0-9_.-]*$/.test(x)) || new Set(hosts).size !== hosts.length) throw new Error('请输入不重复的主机名，每行一个')
      placement.hosts = hosts
    } else if (v.placement_mode === 'label') {
      if (typeof v.label !== 'string' || !v.label.trim()) throw new Error('请输入部署标签')
      placement.label = v.label.trim()
    } else if (v.placement_mode !== 'default') throw new Error('请选择部署位置')
    if (v.count !== undefined && v.count !== null && v.count !== '') {
      if (typeof v.count !== 'number' || !Number.isSafeInteger(v.count) || v.count < 1) throw new Error('实例数必须为正整数')
      placement.count = v.count
    }
    return { cluster_id: clusterId, name: v.name, realm_token: v.realm_token, port: v.port, placement, confirm_import: true }
  }
}

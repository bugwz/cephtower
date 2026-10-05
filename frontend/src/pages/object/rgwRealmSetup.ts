import type { ResourceFormAction } from '../ResourceListPage'

const lines = (v: unknown) => String(v ?? '').split(/\r?\n/).map(x => x.trim()).filter(Boolean)
export const rgwRealmSetupAction: ResourceFormAction = {
  title: '初始化本地主站点并重启 RGW', buttonLabel: '初始化主站点',
  path: '/rgw/realm/setup', method: 'POST',
  successMessage: '默认拓扑及系统密钥已核验；如有所列 RGW 服务，已核验其进程重新启动并运行。无服务时未验证网关；HTTP/TLS 可用性和复制状态仍需检查',
  confirmation: () => '将创建并设为默认 Realm、主 Zonegroup、主 Zone，生成系统用户密钥并发布整个 Period，然后重启所列全部 RGW 服务，可能中断业务。已有数据不会自动迁移，已有服务的显式 Realm/Zone 绑定不会被改写，也不会部署新网关。端点中匹配编排主机名称的主机部分将替换为其登记地址，端口和路径保留；其他地址保持原样。HTTPS 证书应覆盖转换后的地址，转换不代表连通性验证。原生工具的密钥参数可能被系统特权进程或 Ceph 调试日志读取。本操作非事务，不自动重试或回滚；任何失败都可能部分生效。这里只初始化本地主站点，不配置 Ceph Dashboard 凭据或自动操作其他集群。确认已备份并核对服务范围？',
  initialValues: { zone_mode: 'normal' },
  changedValues: changed => 'confirm_setup' in changed ? {} : { confirm_setup: undefined },
  fields: [
    { name: 'name', label: '新 Realm 名称（设为默认）', required: true },
    { name: 'zonegroup', label: '新主 Zonegroup 名称（设为默认）', required: true },
    { name: 'zone', label: '新主 Zone 名称（设为默认）', required: true },
    { name: 'username', label: '新系统用户 ID（密钥由 Ceph 生成，不返回页面）', required: true },
    { name: 'zone_mode', label: '主 Zone 类型', type: 'select', required: true, options: [{ value: 'normal', label: '普通主 Zone' }, { value: 'archive', label: '归档主 Zone' }] },
    { name: 'zonegroup_endpoints', label: 'Zonegroup 访问地址（每行一个 URL；编排主机名自动转换）', type: 'textarea', required: true },
    { name: 'zone_endpoints', label: '主 Zone 访问地址（每行一个 URL；编排主机名自动转换）', type: 'textarea', required: true },
    { name: 'expected_services', label: '将重启的完整 RGW 服务名称（每行一个 rgw.*；留空仅在无服务时成立）', type: 'textarea' },
    { name: 'confirm_setup', label: '默认配置、凭据及重启确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已备份，确认改变默认拓扑和重启全部所列 RGW 服务' }] }
  ],
  buildBody: (v, clusterId) => {
    if (v.confirm_setup !== 'acknowledged') throw new Error('请确认主站点初始化风险')
    if (!['normal', 'archive'].includes(String(v.zone_mode))) throw new Error('请选择主 Zone 类型')
    for (const key of ['name', 'zonegroup', 'zone', 'username']) if (typeof v[key] !== 'string' || !/^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$/.test(v[key] as string)) throw new Error('名称必须为 1–128 个字母、数字、下划线、点或连字符，且不能以点或连字符开头')
    const endpoints = (key: string) => {
      const values = lines(v[key])
      if (!values.length || new Set(values).size !== values.length) throw new Error('请输入不重复的实际访问地址')
      for (const value of values) {
        let u: URL
        try { u = new URL(value) } catch { throw new Error('访问地址不是有效 URL') }
        if (!['http:', 'https:'].includes(u.protocol) || !u.hostname || u.username || u.password || u.search || u.hash || /[;,=\s]/.test(value)) throw new Error('访问地址须为无凭据、查询或片段的 HTTP/HTTPS URL')
      }
      return values
    }
    const services = lines(v.expected_services)
    if (new Set(services).size !== services.length || services.some(x => !/^rgw\.[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$/.test(x))) throw new Error('请填写完整且不重复的 rgw.* 服务名称')
    return { cluster_id: clusterId, name: v.name, zonegroup: v.zonegroup, zone: v.zone, username: v.username,
      zonegroup_endpoints: endpoints('zonegroup_endpoints'), zone_endpoints: endpoints('zone_endpoints'), expected_services: services, confirm_setup: true,
      ...(v.zone_mode === 'archive' ? { tier_type: 'archive' } : {}) }
  }
}

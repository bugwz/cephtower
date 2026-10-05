import type { ResourceFormAction } from '../ResourceListPage'
import { rgwRealmSetupAction } from './rgwRealmSetup'

export const rgwRealmMigrationAction: ResourceFormAction = {
  ...rgwRealmSetupAction,
  title: '将默认单站点迁移到新 Realm', buttonLabel: '迁移默认单站点', path: '/rgw/realm/migrate',
  confirmation: () => '仅用于尚无 Realm、唯一默认 Zonegroup/Zone 均名为 default 的单站点。将保留所填旧 ID 和存储池引用，重命名并挂入新默认 Realm，发布 Period、生成新系统用户密钥并重启全部所列 RGW 服务，可能中断业务；不是对象搬迁。已有服务的显式名称配置不会自动改写，请先协调维护与备份。端点中编排主机名会替换为登记地址，不代表连通性验证。此操作非事务，失败可能已部分生效，不自动回滚或重试；系统特权进程及 Ceph 调试日志可能读取原生命令密钥参数。确认执行迁移？',
  successMessage: '默认拓扑迁移及池引用已核验；有所列服务时已核验重启进程，无服务时未验证网关。不代表业务连通或复制完成',
  fields: [
    { name: 'expected_zonegroup_id', label: '旧 default Zonegroup 的准确 ID', required: true },
    { name: 'expected_zone_id', label: '旧 default Zone 的准确 ID（保留已有存储）', required: true },
    ...rgwRealmSetupAction.fields
  ],
  buildBody: (values, clusterId) => {
    for (const field of ['expected_zonegroup_id', 'expected_zone_id']) {
      if (typeof values[field] !== 'string' || !/^[A-Za-z0-9_][A-Za-z0-9_.-]*$/.test(values[field])) throw new Error('请输入准确的旧默认拓扑 ID')
    }
    if (values.zonegroup === 'default' || values.zone === 'default') throw new Error('迁移需要新的 Zonegroup 和 Zone 名称')
    return { ...rgwRealmSetupAction.buildBody(values, clusterId), expected_zonegroup_id: values.expected_zonegroup_id, expected_zone_id: values.expected_zone_id, confirm_migration: true }
  }
}

import { ArrowLeftOutlined, BulbOutlined, DeleteOutlined, PlusOutlined, PoweroffOutlined, ReloadOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Descriptions, Form, Input, InputNumber, Modal, Select, Space, Switch, Tabs, Tag, Typography } from 'antd'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { MonMapSettings } from './MonMapSettings'
import { DeviceHardwareSummary } from './DeviceHardwareSummary'
import { DaemonRuntimeDetails, ServiceDaemons } from './ServiceDaemons'
import { DaemonPerf } from './DaemonPerf'
import { ServiceInventoryDetails } from './ServiceInventoryDetails'
import { hostStorageCapacity } from './hostStorageCapacity'
import { ManagerInventory } from './ManagerInventory'
import { osdSnapshotValue } from './OSDUsage'
import { OSDScrubConfiguration } from './OSDScrubConfiguration'
import { OSDRecoveryConfiguration } from './OSDRecoveryConfiguration'
import { OSDGlobalFlags } from './OSDGlobalFlags'
import { OSDCapacityThresholds } from './OSDCapacityThresholds'
import { osdOperationalStatus } from './osdOperationalStatus'
import { OSDRemovalStop } from './OSDRemovalStop'
import { OSDRemovalDetails, removalBoolean } from './OSDRemovalDetails'
import { numberValue, textValue, type ApiRecord } from '../../api/client'
import {
  applyDaemonAction,
  getMonitorStatus,
  listMonitors,
  listResource,
  listAllResources,
  listOSDFlags,
  mutateResource,
  refreshResource,
  reweightOSD,
  scrubOSD,
  setMgrModuleEnabled
} from '../../api/resource'
import { DataTable } from '../../components/DataTable'
import { DraggableModal, draggableModalRender } from '../../components/DraggableModal'
import { Page } from '../../components/Page'
import { RecordDetail } from '../../components/RecordDetail'
import { ResourceMetaBar } from '../../components/ResourceMetaBar'
import { TableAction, TableActions } from '../../components/TableActions'
import { useResource } from '../../hooks'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { mergeResourceFilters, useResourceTableFilters } from '../../hooks/useResourceTableFilters'
import { useClusterContext } from '../../state/ClusterContext'
import { message } from '../../utils/appMessage'
import { formatDateTime } from '../../utils/time'
import { OSDInspection } from './OSDInspection'
import { ConfigurationPage } from './ConfigurationPage'
import { ManagerModuleDetails, managerOptionRows } from './ManagerModuleDetails'
import { ClusterDetailPage } from './ClusterDetailPage'
import { ClusterPage } from './ClusterPage'
import { HostDetailPage } from './HostDetailPage'
import { formatBytes, HostPage } from './HostPage'
import { MonDetailPage } from './MonDetailPage'
import { MonPublicAddresses } from './MonPublicAddresses'
import { monSessionCount } from './monSessionCount'
import { MonQuorumState } from './MonQuorumState'
import { MonSnapshotState } from './MonSnapshotState'
import { monQuorumGroups } from './monQuorumGroups'
import { monQuorumMembers, monQuorumLeader, monQuorumInteger } from './monQuorumSummary'
import { PoolDetailPage } from './PoolDetailPage'
import { PoolManagementPage } from './PoolManagementPage'
import { ServicePage } from './ServicePage'

export { ClusterDetailPage, ClusterPage, HostDetailPage, HostPage, MonDetailPage, PoolDetailPage, PoolManagementPage, ServicePage }

const { Text } = Typography
const twoColumnDescriptions = { xs: 1, sm: 2, md: 2, lg: 2, xl: 2, xxl: 2 }

type DeviceScope = 'available' | 'used' | 'unavailable' | 'unknown'

export function MonManagementPage() {
  const navigate = useNavigate()
  const { selectedClusterId } = useClusterContext()
  const [refreshingSection, setRefreshingSection] = useState<'status' | 'nodes' | null>(null)
  const operationMutation = useMutationOperation()
  const monTableFilters = useResourceTableFilters({
    path: '/monitors',
    fields: ['name', 'rank', 'address', 'status'],
    clusterId: selectedClusterId
  })
  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return { mons: [], status: null }
    }
    const [mons, status] = await Promise.all([
      listMonitors(selectedClusterId, monTableFilters.filters),
      getMonitorStatus(selectedClusterId)
    ])
    return { mons, status }
  }, [monTableFilters.filters, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)

  async function refreshMonSection(section: 'status' | 'nodes') {
    if (!selectedClusterId || refreshingSection) {
      return
    }
    setRefreshingSection(section)
    try {
      const kinds = section === 'status' ? ['mon_status'] : ['mon', 'mon_perf_counter']
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kinds }), '刷新成功')
      await refresh({ showLoading: false })
    } finally {
      setRefreshingSection(null)
    }
  }

  return (
    <Page title="MON管理" loading={loading} error={error}>
      <Space direction="vertical" size={16} className="page-stack">
        <Card
          className="page-surface-card"
          title="MON 基础信息"
          extra={
            <Button
              icon={<ReloadOutlined />}
              loading={refreshingSection === 'status'}
              disabled={Boolean(refreshingSection && refreshingSection !== 'status') || !selectedClusterId}
              onClick={() => refreshMonSection('status')}
            >
              刷新
            </Button>
          }
        >
          <MonSnapshotState observedAt={data?.status?.observed_at} stale={data?.status?.stale} />
          <Descriptions className="mon-status-descriptions" size="small" column={twoColumnDescriptions} bordered>
            <Descriptions.Item label="集群 ID">{textValue(data?.status?.fsid)}</Descriptions.Item>
            <Descriptions.Item label="仲裁成员（采集时）">{monQuorumMembers(data?.status?.quorum_names)}</Descriptions.Item>
            <Descriptions.Item label="Leader（采集时）">{monQuorumLeader(data?.status?.quorum_leader_name)}</Descriptions.Item>
            <Descriptions.Item label="选举 epoch">{monQuorumInteger(data?.status?.election_epoch)}</Descriptions.Item>
            <Descriptions.Item label="当前仲裁持续时间（采集时）">{monQuorumInteger(data?.status?.quorum_age, true)}</Descriptions.Item>
            <Descriptions.Item label="monmap 修改时间">{formatDateTime(data?.status?.modified)}</Descriptions.Item>
            <Descriptions.Item label="monmap epoch">{textValue(data?.status?.epoch)}</Descriptions.Item>
            <Descriptions.Item label="Quorum 连接特性">{textValue(data?.status?.quorum_con)}</Descriptions.Item>
            <Descriptions.Item label="Quorum MON 特性" span={2}>{textValue(data?.status?.quorum_mon)}</Descriptions.Item>
            <Descriptions.Item label="必需连接特性" span={2}>{textValue(data?.status?.required_con)}</Descriptions.Item>
            <Descriptions.Item label="必需 MON 特性" span={2}>{textValue(data?.status?.required_mon)}</Descriptions.Item>
          </Descriptions>
          <MonMapSettings value={data?.status} />
        </Card>
        <Card
          className="page-surface-card"
          title="MON 节点"
          extra={
            <Button
              icon={<ReloadOutlined />}
              loading={refreshingSection === 'nodes'}
              disabled={Boolean(refreshingSection && refreshingSection !== 'nodes') || !selectedClusterId}
              onClick={() => refreshMonSection('nodes')}
            >
              刷新
            </Button>
          }
        >
          <Text type="secondary">按采集时的仲裁状态分组；数量为当前筛选结果，不代表实时集群状态。</Text>
          {monQuorumGroups(data?.mons ?? []).map(group => (
          <section key={group.key} aria-label={group.label}>
          <Typography.Title level={5}>{group.label}（{group.rows.length}）</Typography.Title>
          <DataTable
            data={group.rows}
          filterOptions={monTableFilters.filterOptions}
          filteredValues={monTableFilters.filters}
          onFilterChange={monTableFilters.handleFilterChange}
          rowKeyCandidates={['name', 'natural_key', 'rank']}
          columns={[
            { key: 'name', title: '名称' },
            { key: 'rank', title: 'Rank' },
            { key: 'address', title: 'Public Addr' },
            { key: 'public_addresses', title: '全部 Public 地址', filterKey: false, render: (value) => <MonPublicAddresses value={value} /> },
            {
              key: 'status',
              title: '状态',
              render: (_, row) => <MonQuorumState value={row.in_quorum} />
            },
            { key: 'open_sessions', title: 'Open sessions（采集时）', render: (value) => monSessionCount(value) },
            { key: 'observed_at', title: '采集状态', filterKey: false, render: (_, row) => <MonSnapshotState observedAt={row.observed_at} stale={row.stale} /> },
            {
              key: 'actions',
              title: '操作',
              filterKey: false,
              render: (_, row) => {
                const name = textValue(row.name ?? row.natural_key, '')
                return (
                  <TableActions>
                    <TableAction disabled={!name} onClick={() => navigate(`/cluster/mon/${encodeURIComponent(name)}`)}>详情</TableAction>
                  </TableActions>
                )
              }
            }
          ]}
          />
          </section>
          ))}
        </Card>
      </Space>
    </Page>
  )
}

export function MgrManagementPage() {
  const { selectedClusterId } = useClusterContext()
  const moduleScope = useRef({ clusterId: selectedClusterId })
  if (moduleScope.current.clusterId !== selectedClusterId) moduleScope.current = { clusterId: selectedClusterId }
  const scope = moduleScope.current
  const moduleRunning = useRef(false)
  const moduleTableFilters = useResourceTableFilters({
    path: '/manager/modules',
    fields: ['name', 'enabled', 'always_on'],
    clusterId: selectedClusterId
  })
  const daemonTableFilters = useResourceTableFilters({
    path: '/daemons',
    fields: ['name', 'type', 'hostname', 'status', 'version'],
    clusterId: selectedClusterId
  })
  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return { modules: [], daemons: [], inventoryWarnings: [] }
    }
    const [modules, daemons] = await Promise.all([
      listAllResources('/manager/modules', selectedClusterId, { filters: moduleTableFilters.filters }),
      listAllResources('/daemons', selectedClusterId, {
        filters: mergeResourceFilters(daemonTableFilters.filters, { type: ['mgr'] })
      })
    ])
    return {
      modules: modules.items,
      daemons: daemons.items,
      inventoryWarnings: [
        modules.stale !== false ? `MGR 模块库存：${modules.staleReason || '已过期或新鲜度未知'}` : null,
        daemons.stale !== false ? `MGR 守护进程库存：${daemons.staleReason || '已过期或新鲜度未知'}` : null
      ].filter((value): value is string => value !== null)
    }
  }, [daemonTableFilters.filters, moduleTableFilters.filters, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const [pendingModule, setPendingModule] = useState('')
  const [moduleSelection, setModuleSelection] = useState<{ scope: typeof scope, row: ApiRecord } | null>(null)
  const moduleDetails = moduleSelection?.scope === scope ? moduleSelection.row : null
  const [configSelection, setConfigSelection] = useState<{ scope: typeof scope, name: string } | null>(null)
  const configModule = configSelection?.scope === scope ? configSelection.name : ''
  async function collectManagerInventory() {
    if (!selectedClusterId || moduleScope.current !== scope || moduleRunning.current || loading) return
    moduleRunning.current = true
    setPendingModule('__collect__')
    try {
      await refreshResource({ clusterId: selectedClusterId, kinds: ['mgr_module', 'daemon'] })
      if (moduleScope.current !== scope) return
      message.success('MGR 模块与守护进程采集完成，正在重新读取库存')
      await refresh()
    } catch (err) {
      if (moduleScope.current === scope) message.error(err instanceof Error ? err.message : 'MGR 模块与守护进程采集失败')
    } finally {
      moduleRunning.current = false
      setPendingModule('')
    }
  }
  async function toggleModule(row: ApiRecord, enabled: boolean) {
    const name = textValue(row.name, '')
    if (!selectedClusterId || moduleScope.current !== scope || moduleRunning.current || loading || error || !name || row.stale !== false || typeof row.enabled !== 'boolean' || row.enabled === enabled || row.always_on !== false || (enabled && row.can_run !== true)) {
      return
    }
    moduleRunning.current = true
    setPendingModule(name)
    try {
      await setMgrModuleEnabled(selectedClusterId, name, enabled)
      if (moduleScope.current !== scope) return
      message.success(enabled ? 'Mgr 模块启用已核验' : 'Mgr 模块停用已核验')
      try {
        await refreshResource({ clusterId: selectedClusterId, kind: 'mgr_module' })
      } catch {
        if (moduleScope.current === scope) message.warning('模块状态修改已核验，但重新采集失败；请刷新核对，不要重复提交。')
      }
      if (moduleScope.current === scope) await refresh()
    } finally {
      moduleRunning.current = false
      setPendingModule('')
    }
  }

  return (
    <Page title="MGR管理" loading={loading} error={error}>
      {data?.inventoryWarnings.map(warning => <Alert key={warning} type="warning" showIcon message={warning} />)}
      <Modal open={Boolean(configModule)} onCancel={() => setConfigSelection(null)} footer={null} width="95vw" destroyOnClose>
        {configModule && <ConfigurationPage key={`${selectedClusterId}:${configModule}`} moduleName={configModule} />}
      </Modal>
      <Modal title={`模块 ${textValue(moduleDetails?.name, '')}`} open={Boolean(moduleDetails)} onCancel={() => setModuleSelection(null)} footer={null} width="min(1200px, 95vw)" destroyOnClose>
        {moduleDetails && selectedClusterId && <ManagerModuleDetails key={`${selectedClusterId}:${String(moduleDetails.name)}`} record={moduleDetails} clusterId={selectedClusterId} />}
      </Modal>
      <Card className="page-surface-card" title="MGR管理" extra={<Button disabled={!selectedClusterId || loading || Boolean(pendingModule)} loading={pendingModule === '__collect__'} onClick={() => void collectManagerInventory()}>重新采集模块与守护进程</Button>}>
        <Tabs
          items={[
            {
              key: 'modules',
              label: '模块',
              children: (
                <div className="embedded-panel">
                <DataTable
                  data={data?.modules ?? []}
                  filterOptions={moduleTableFilters.filterOptions}
                  filteredValues={moduleTableFilters.filters}
                  onFilterChange={moduleTableFilters.handleFilterChange}
                  rowKeyCandidates={['name']}
                  columns={[
                    { key: 'name', title: '模块' },
                    {
                      key: 'enabled',
                      title: '启用',
                      render: (value, row) => {
                        const name = textValue(row.name, '')
                        return (
                          <Switch
                            checked={value === true}
                            disabled={loading || Boolean(error) || row.stale !== false || typeof value !== 'boolean' || row.always_on !== false || (!value && row.can_run !== true) || Boolean(pendingModule)}
                            loading={pendingModule === name}
                            onChange={(checked) => toggleModule(row, checked)}
                          />
                        )
                      }
                    },
                    { key: 'always_on', title: '常驻', render: (value) => <Tag color={value ? 'processing' : 'default'}>{value ? '是' : '否'}</Tag> },
                    { key: 'can_run', title: '可运行', render: (value) => <Tag color={value === false ? 'error' : 'default'}>{value === true ? '是' : value === false ? '否' : '未知'}</Tag> },
                    { key: 'error_string', title: '加载错误' },
                    { key: 'force_disabled', title: '强制停用', render: (value) => value ? '是' : '否' },
                    { key: 'options', title: '配置项', render: (value, row) => <Button type="link" onClick={() => setModuleSelection({ scope, row })}>{managerOptionRows(value)?.length ?? '未采集'} 项 · 详情</Button> },
                    { key: 'configure', title: '操作', render: (_, row) => <Button onClick={() => setConfigSelection({ scope, name: textValue(row.name, '') })}>编辑配置</Button> }
                  ]}
                />
                </div>
              )
            },
            {
              key: 'daemons',
              label: '守护进程',
              children: <DaemonTable key={selectedClusterId} clusterId={selectedClusterId} unavailable={loading || Boolean(error)} data={data?.daemons ?? []} refresh={refresh} tableFilters={daemonTableFilters} />
            },
            {
              key: 'manager-inventory', label: '原生 MGR 清单',
              children: selectedClusterId ? <ManagerInventory key={selectedClusterId} clusterId={selectedClusterId} /> : <Text>请先选择集群</Text>
            }
          ]}
        />
      </Card>
    </Page>
  )
}

export function OsdManagementPage() {
  const { selectedClusterId } = useClusterContext()
  const osdScopeRef = useRef({ clusterId: selectedClusterId })
  if (osdScopeRef.current.clusterId !== selectedClusterId) osdScopeRef.current = { clusterId: selectedClusterId }
  const osdScope = osdScopeRef.current
  useEffect(() => {
    osdScopeRef.current = osdScope
    return () => { if (osdScopeRef.current === osdScope) osdScopeRef.current = { ...osdScope } }
  }, [osdScope])
  const osdTableFilters = useResourceTableFilters({
    path: '/osds',
    fields: ['id', 'host', 'state', 'up', 'in', 'device_class'],
    clusterId: selectedClusterId
  })
  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return { osds: [], flags: [], removals: [], removalMeta: null }
    }
    const [osds, flags, removalMeta] = await Promise.all([
      listAllResources('/osds', selectedClusterId, { filters: osdTableFilters.filters }).then((payload) => payload.items),
      listOSDFlags(selectedClusterId),
      listAllResources('/osd/removals', selectedClusterId, { limit: 500 })
    ])
    return { osds, flags, removals: removalMeta.items, removalMeta }
  }, [osdTableFilters.filters, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const [pendingOSDAction, setPendingOSDAction] = useState('')
  const osdActionRunning = useRef(false)
  const [deploymentOpen, setDeploymentOpen] = useState(false)
  const [osdInspection, setOSDInspection] = useState<{ scope: typeof osdScope; row: ApiRecord } | null>(null)
  const inspectedOSD = osdInspection?.scope === osdScope ? osdInspection.row : null
  const [scrubConfigScope, setScrubConfigScope] = useState<typeof osdScope | null>(null)
  const scrubConfigOpen = scrubConfigScope === osdScope
  const [recoveryConfigScope, setRecoveryConfigScope] = useState<typeof osdScope | null>(null)
  const recoveryConfigOpen = recoveryConfigScope === osdScope
  const [refreshingOSDs, setRefreshingOSDs] = useState(false)
  const operationMutation = useMutationOperation()

  async function refreshOSDData() {
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    setRefreshingOSDs(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kinds: ['osd', 'osd_flag', 'osd_removal'] }), '刷新成功')
      await refresh()
    } finally {
      setRefreshingOSDs(false)
    }
  }

  async function runOSDAction(id: string, action: 'in' | 'out' | 'down' | 'scrub' | 'deep-scrub' | 'reweight', currentWeight?: unknown, expectedVersion?: string) {
    if (!selectedClusterId || osdScopeRef.current !== osdScope || osdActionRunning.current) return
    if (action === 'reweight') {
      if (loading || error || !expectedVersion) return
      Modal.confirm({
        title: `调整 OSD ${id} 权重`,
        content: <ReweightForm currentWeight={currentWeight} version={expectedVersion} clusterId={selectedClusterId} isCurrent={() => osdScopeRef.current === osdScope} osdID={id} refresh={refresh} />,
        modalRender: draggableModalRender,
        icon: null,
        okButtonProps: { style: { display: 'none' } },
        cancelText: '关闭'
      })
      return
    }

    const pendingKey = `${id}:${action}`
    if (pendingOSDAction) {
      return
    }
    setPendingOSDAction(pendingKey)
    osdActionRunning.current = true
    try {
      if (action === 'scrub' || action === 'deep-scrub') {
        await operationMutation.run(() => scrubOSD(selectedClusterId, id, action === 'deep-scrub'), false)
      } else {
        if (!expectedVersion) throw new Error('请重新采集并确认 OSD 版本')
        await operationMutation.run(() => mutateResource('/osd/action', 'POST', { cluster_id: selectedClusterId, osd_id: id, action }, { ifMatch: expectedVersion }), false)
      }
      if (osdScopeRef.current !== osdScope) return
      message.success(`OSD ${action} 命令已完成`)
      await refreshResource({ clusterId: selectedClusterId, kinds: ['osd', 'osd_flag', 'osd_removal'] })
      if (osdScopeRef.current === osdScope) await refresh()
    } finally {
      osdActionRunning.current = false
      setPendingOSDAction('')
    }
  }

  function confirmOSDState(row: ApiRecord, action: 'in' | 'out' | 'down') {
    if (!selectedClusterId || osdScopeRef.current !== osdScope || loading || error || row.stale !== false || osdActionRunning.current) return
    if (action === 'down' ? row.up !== true : action === 'in' ? row.in !== false : row.in !== true) return
    const id = osdID(row)
    const version = osdInventoryVersion(row.resource_version)
    if (!id || !version) { message.error('OSD 身份或版本无效，请重新采集'); return }
    Modal.confirm({
      title: `标记集群 ${selectedClusterId} 的 OSD ${id} 为 ${action}`,
      content: action === 'down'
        ? '这会修改 OSDMap 状态，可能触发故障处理和数据恢复；不会停止 OSD 进程，运行中的 OSD 可能再次被标记 Up。'
        : action === 'in' ? '将 OSD 标记 In，允许参与数据放置，可能触发数据重平衡；不会启动 OSD 进程。' : '将 OSD 标记 Out，数据可能迁移到其他 OSD 并增加恢复负载；不会停止或删除 OSD 进程。',
      okText: `确认标记 ${action}`, okType: 'danger', cancelText: '取消',
      async onOk() {
        if (osdScopeRef.current !== osdScope) throw new Error('集群已切换或页面已关闭，请重新确认')
        await runOSDAction(id, action, undefined, version)
      }
    })
  }

  async function deleteOSD(row: ApiRecord, preserveId = false) {
    if (osdScopeRef.current !== osdScope || loading || error || row.stale !== false) {
      message.error('请先成功采集当前集群的 OSD 库存，再确认删除')
      return
    }
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    const id = osdID(row)
    if (!id) {
      message.error('无法识别 OSD ID')
      return
    }
    const generation = osdInventoryVersion(row.resource_version)
    if (generation === null) { message.error('库存版本无效，请重新采集后再确认删除'); return }
    const parameters = { cluster_id: selectedClusterId, osd_id: id, zap: false, preserve_id: preserveId }
    Modal.confirm({
      title: `${preserveId ? '替换（保留 ID）' : '移除'}集群 ${selectedClusterId} 的 OSD ${id}`,
      content: preserveId
        ? '将请求编排器排空并移除 OSD，保留 ID 用于后续替换。这不是立即部署替代 OSD，也不代表替换已完成。可能触发数据迁移；不会启用强制移除或清盘。'
        : '将请求编排器排空并移除 OSD，不保留 ID 用于替换。可能触发数据迁移；不会启用强制移除或清盘。提交不代表移除已完成，请关注移除队列。',
      okText: preserveId ? '提交保留 ID 的移除' : '提交移除',
      okType: 'danger',
      cancelText: '取消',
      async onOk() {
        if (osdScopeRef.current !== osdScope) throw new Error('集群已切换或页面已关闭，请重新确认 OSD 删除')
        await operationMutation.run(() => mutateResource('/osd', 'DELETE', parameters, { ifMatch: generation }), false)
        if (osdScopeRef.current !== osdScope) return
        message.success(preserveId ? '保留 ID 的 OSD 移除请求已提交，请关注队列及后续替换' : 'OSD 移除请求已提交')
        await refreshResource({ clusterId: selectedClusterId, kinds: ['osd', 'osd_removal'] })
        if (osdScopeRef.current === osdScope) await refresh({ showLoading: false })
      }
    })
  }

  return (
    <Page title="OSD管理" loading={loading} error={error}>
      <Modal open={recoveryConfigOpen} title={`集群 ${selectedClusterId} / 恢复速度与 QoS 配置`} onCancel={() => setRecoveryConfigScope(null)} footer={null} width="90vw" destroyOnClose>
        {recoveryConfigOpen && selectedClusterId && <OSDRecoveryConfiguration key={selectedClusterId} />}
      </Modal>
      <Modal open={scrubConfigOpen} title={`集群 ${selectedClusterId} / PG Scrub 配置`} onCancel={() => setScrubConfigScope(null)} footer={null} width="90vw" destroyOnClose>
        {scrubConfigOpen && selectedClusterId && <OSDScrubConfiguration key={selectedClusterId} />}
      </Modal>
      <Modal open={Boolean(inspectedOSD)} title={`OSD ${osdID(inspectedOSD ?? {})} 详情`} onCancel={() => setOSDInspection(null)} footer={null} width="90vw" destroyOnClose>
        {inspectedOSD && selectedClusterId && <OSDInspection key={`${selectedClusterId}:${osdID(inspectedOSD)}`} clusterId={selectedClusterId} osdId={osdID(inspectedOSD)} record={inspectedOSD} />}
      </Modal>
      <Card
        className="page-surface-card"
        title="OSD管理"
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} loading={refreshingOSDs} onClick={refreshOSDData}>刷新</Button>
            <Button disabled={!selectedClusterId} onClick={() => setScrubConfigScope(osdScope)}>PG Scrub 配置</Button>
            <Button disabled={!selectedClusterId} onClick={() => setRecoveryConfigScope(osdScope)}>恢复速度 / QoS</Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setDeploymentOpen(true)}>OSD 部署</Button>
          </Space>
        }
      >
        <Space direction="vertical" size={16} className="page-stack">
        <section className="embedded-panel">
          <div className="embedded-panel-title">OSD 移除队列</div>
          <ResourceMetaBar observedAt={data?.removalMeta?.observedAt} stale={data?.removalMeta?.stale} />
          <DataTable data={data?.removals ?? []} rowKeyCandidates={['osd_id', 'natural_key']} columns={[
            { key: 'osd_id', title: 'OSD' },
            { key: 'hostname', title: '主机' },
            { key: 'drain_status', title: '排空状态' },
            { key: 'pg_count', title: '剩余 PG' },
            { key: 'replace', title: '替换', render: removalBoolean },
            { key: 'force', title: '强制', render: removalBoolean },
            { key: 'zap', title: '清盘', render: removalBoolean },
            { key: 'drain_started_at', title: '排空开始', render: (value) => formatDateTime(value) },
            { key: 'details', title: '详情', filterKey: false, render: (_, row) => selectedClusterId ? <OSDRemovalDetails key={`${selectedClusterId}:${row.osd_id}:details`} clusterId={selectedClusterId} record={row} /> : null },
            { key: 'process_started_at', title: '开始时间', render: (value) => formatDateTime(value) },
            { key: 'stop', title: '操作', filterKey: false, render: (_, row) => selectedClusterId ? <OSDRemovalStop key={`${selectedClusterId}:${row.osd_id}:${row.resource_version}`} clusterId={selectedClusterId} record={row} blocked={loading || Boolean(error) || data?.removalMeta?.stale !== false} onChanged={refresh} /> : null }
          ]} />
        </section>
        <section className="embedded-panel">
          <div className="embedded-panel-title">OSD Flags</div>
          {data?.flags == null ? <span className="muted">OSD flags 未采集或格式无效</span> : data.flags.length ? data.flags.map((flag, index) => <Tag key={`${index}-${flag}`}>{flag}</Tag>) : <span className="muted">未设置 OSD flags</span>}
          {selectedClusterId && <OSDGlobalFlags key={selectedClusterId} clusterId={selectedClusterId} />}
          {selectedClusterId && <OSDCapacityThresholds key={selectedClusterId} clusterId={selectedClusterId} />}
        </section>
        <section className="embedded-panel">
          <DataTable
            data={data?.osds ?? []}
            filterOptions={osdTableFilters.filterOptions}
            filteredValues={osdTableFilters.filters}
            onFilterChange={osdTableFilters.handleFilterChange}
            rowKeyCandidates={['id', 'osd', 'service_id', 'name']}
            columns={[
              { key: 'id', title: 'ID' },
              { key: 'host', title: '主机' },
              { key: 'state', title: '原生状态标记', render: osdStateText },
              { key: 'operational_status', title: '运行管理状态', filterKey: false, render: (_, row) => osdOperationalStatus(row, osdID(row), data?.removals, data?.removalMeta?.stale, loading || Boolean(error)) },
              { key: 'up', title: 'Up' },
              { key: 'in', title: 'In' },
              { key: 'device_class', title: '设备类型' },
              { key: 'reweight', title: 'OSD 调权系数', render: value => typeof value === 'number' ? value : '未采集' },
              { key: 'uuid', title: '实例 UUID', render: value => typeof value === 'string' ? value || '空字符串（原生）' : '未采集' },
              { key: 'primary_affinity', title: '主副本亲和度', render: value => typeof value === 'number' ? value : '未采集' },
              { key: 'capacity_total', title: '总容量（KiB）', filterKey: false, render: (_, row) => osdSnapshotValue(row, 'stats', 'kb') },
              { key: 'capacity_used', title: '已用（KiB）', filterKey: false, render: (_, row) => osdSnapshotValue(row, 'stats', 'kb_used') },
              { key: 'capacity_available', title: '可用（KiB）', filterKey: false, render: (_, row) => osdSnapshotValue(row, 'stats', 'kb_avail') },
              { key: 'utilization', title: '利用率', filterKey: false, render: (_, row) => osdSnapshotValue(row, 'stats', 'utilization') },
              { key: 'pgs', title: 'PG 数', filterKey: false, render: (_, row) => osdSnapshotValue(row, 'stats', 'pgs') },
              { key: 'commit_latency', title: '提交延迟（ms）', filterKey: false, render: (_, row) => osdSnapshotValue(row, 'perf_stats', 'commit_latency_ms') },
              { key: 'apply_latency', title: '应用延迟（ms）', filterKey: false, render: (_, row) => osdSnapshotValue(row, 'perf_stats', 'apply_latency_ms') },
              {
                key: 'actions',
                title: '操作',
                filterKey: false,
                render: (_, row) => {
                  const id = osdID(row)
                  return (
                    <TableActions>
                      <TableAction onClick={() => setOSDInspection({ scope: osdScope, row })}>详情</TableAction>
                      <TableAction disabled={Boolean(pendingOSDAction)} onClick={() => runOSDAction(id, 'deep-scrub')}>Deep scrub</TableAction>
                      <TableAction loading={pendingOSDAction === `${id}:in`} disabled={Boolean(pendingOSDAction) || row.in !== false || row.stale !== false || loading || Boolean(error)} onClick={() => confirmOSDState(row, 'in')}>In</TableAction>
                      <TableAction loading={pendingOSDAction === `${id}:out`} disabled={Boolean(pendingOSDAction) || row.in !== true || row.stale !== false || loading || Boolean(error)} onClick={() => confirmOSDState(row, 'out')}>Out</TableAction>
                      <TableAction danger disabled={Boolean(pendingOSDAction) || row.up !== true || row.stale !== false || loading || Boolean(error)} onClick={() => confirmOSDState(row, 'down')}>Down</TableAction>
                      <TableAction loading={pendingOSDAction === `${id}:scrub`} disabled={Boolean(pendingOSDAction) && pendingOSDAction !== `${id}:scrub`} onClick={() => runOSDAction(id, 'scrub')}>Scrub</TableAction>
                      <TableAction disabled={Boolean(pendingOSDAction) || row.stale !== false || loading || Boolean(error) || !osdInventoryVersion(row.resource_version)} onClick={() => runOSDAction(id, 'reweight', row.reweight, row.stale === false ? osdInventoryVersion(row.resource_version) ?? undefined : undefined)}>权重</TableAction>
                      <TableAction danger disabled={Boolean(pendingOSDAction)} onClick={() => deleteOSD(row)}>删除</TableAction>
                      <TableAction danger disabled={Boolean(pendingOSDAction)} onClick={() => deleteOSD(row, true)}>替换（保留 ID）</TableAction>
                    </TableActions>
                  )
                }
              }
            ]}
          />
        </section>
        </Space>
      </Card>
      <OSDDeploymentModal open={deploymentOpen} onClose={() => setDeploymentOpen(false)} refresh={refresh} />
    </Page>
  )
}

function OSDDeploymentModal({ open, onClose, refresh }: { open: boolean; onClose: () => void; refresh: (options?: { showLoading?: boolean }) => void }) {
  const { selectedClusterId } = useClusterContext()
  const [preview, setPreview] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [form] = Form.useForm<{
    service_id?: string
    host_pattern?: string
    all?: boolean
    paths?: string
    rotational?: boolean
    model?: string
    vendor?: string
    size?: string
  }>()
  const operationMutation = useMutationOperation()

  async function submit(values: {
    service_id?: string
    host_pattern?: string
    all?: boolean
    paths?: string
    rotational?: boolean
    model?: string
    vendor?: string
    size?: string
  }, mode: 'preview' | 'create') {
    if (!selectedClusterId || submitting) {
      return
    }
    const payload = osdDeploymentPayload(values, selectedClusterId)
    setSubmitting(true)
    try {
      if (mode === 'preview') {
        const result = await operationMutation.run(() => mutateResource('/osd/deployment/preview', 'POST', payload), false)
        setPreview(String((result.details as ApiRecord | undefined)?.preview ?? 'Ceph 未返回预览内容'))
      } else {
        Modal.confirm({
          title: '创建 OSD 部署',
          content: '该操作为高风险操作，确认后将直接执行部署。',
          okText: '提交创建',
          okType: 'danger',
          cancelText: '取消',
          async onOk() {
            await operationMutation.run(() => mutateResource('/osd/deployment', 'POST', payload), false)
            window.setTimeout(() => {
              onClose()
              message.success('OSD 部署执行成功')
              refresh({ showLoading: false })
            })
          }
        })
        return
      }
      message.success('OSD 部署预览已生成')
    } finally {
      setSubmitting(false)
    }
  }

  function finishPreview() {
    void form.validateFields().then((values) => submit(values, 'preview'))
  }

  function finishCreate() {
    void form.validateFields().then((values) => submit(values, 'create'))
  }

  return (
    <DraggableModal
      title="OSD 部署"
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="cancel" onClick={onClose}>取消</Button>,
        <Button key="preview" loading={submitting} onClick={finishPreview}>预览</Button>,
        <Button key="create" type="primary" loading={submitting} onClick={finishCreate}>创建</Button>
      ]}
      destroyOnClose
    >
      <Form form={form} layout="vertical" initialValues={{ all: false }} preserve={false} onValuesChange={() => setPreview('')}>
        <Form.Item name="service_id" label="Service ID">
          <Input />
        </Form.Item>
        <Form.Item name="host_pattern" label="Host Pattern">
          <Input />
        </Form.Item>
        <Form.Item name="paths" label="设备路径">
          <Input placeholder="/dev/sdb,/dev/sdc" />
        </Form.Item>
        <Form.Item name="all" label="所有可用设备" valuePropName="checked">
          <Switch />
        </Form.Item>
        <Form.Item name="rotational" label="Rotational" valuePropName="checked">
          <Switch />
        </Form.Item>
        <Form.Item name="model" label="Model">
          <Input />
        </Form.Item>
        <Form.Item name="vendor" label="Vendor">
          <Input />
        </Form.Item>
        <Form.Item name="size" label="Size">
          <Input />
        </Form.Item>
      </Form>
      {preview && <Card title="部署预览"><Typography.Paragraph style={{ whiteSpace: 'pre-wrap' }} copyable>{preview}</Typography.Paragraph></Card>}
    </DraggableModal>
  )
}

export function DeviceManagementPage() {
  const navigate = useNavigate()
  const { selectedClusterId } = useClusterContext()
  const [availabilityScope, setAvailabilityScope] = useState<DeviceScope | 'all'>('all')
  useEffect(() => setAvailabilityScope('all'), [selectedClusterId])
  const deviceTableFilters = useResourceTableFilters({
    path: '/devices',
    fields: ['hostname', 'path', 'device_id', 'device_type'],
    clusterId: selectedClusterId
  })
  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return {
        items: [],
        observedAt: null,
        stale: false,
        staleReason: null
      }
    }
    return listAllResources('/devices', selectedClusterId, {
      filters: deviceTableFilters.filters
    })
  }, [deviceTableFilters.filters, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const [refreshingDevices, setRefreshingDevices] = useState(false)
  const operationMutation = useMutationOperation()
  const deviceRows = useMemo(() => (data?.items ?? []).map(normalizeDeviceRow).filter(row => availabilityScope === 'all' || row.usage_state === availabilityScope), [data?.items, availabilityScope])

  async function refreshDeviceData() {
    if (!selectedClusterId) {
      message.error('请先选择集群')
      return
    }
    setRefreshingDevices(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kind: 'device' }), '刷新成功')
      await refresh()
    } finally {
      setRefreshingDevices(false)
    }
  }

  return (
    <Page title="设备管理" loading={loading} error={error}>
      <Card
        className="page-surface-card"
        title="设备管理"
        extra={<Button icon={<ReloadOutlined />} loading={refreshingDevices || loading} onClick={refreshDeviceData}>刷新</Button>}
      >
        <Select aria-label="设备可用性筛选" value={availabilityScope} onChange={setAvailabilityScope} style={{ width: 180, marginBottom: 16 }} options={[
            { value: 'all', label: '全部可用性状态' }, { value: 'available', label: '空闲可用' },
            { value: 'used', label: '已占用' }, { value: 'unavailable', label: '不可用' }, { value: 'unknown', label: '可用性未知' }
          ]} />
        <DataTable
          data={deviceRows}
          filterOptions={deviceTableFilters.filterOptions}
          filteredValues={deviceTableFilters.filters}
          onFilterChange={deviceTableFilters.handleFilterChange}
          footer={<ResourceMetaBar observedAt={data?.observedAt} stale={data?.stale} staleReason={data?.staleReason} />}
          rowKeyCandidates={['natural_key', 'device_id', 'path', 'name']}
          columns={[
            { key: 'hostname', title: '主机' },
            { key: 'path', title: '路径' },
            { key: 'device_id', title: '设备 ID' },
            { key: 'size_display', title: '容量', filterKey: false },
            { key: 'device_type', title: '类型' },
            { key: 'usage_label', title: '状态', filterKey: false, render: (_, row) => renderDeviceUsage(row) },
            { key: 'usage_notes', title: '拒绝原因 / 原生诊断', filterKey: false, ellipsis: false, render: (value) => renderDeviceNotes(value) },
            {
              key: 'actions',
              title: '操作',
              filterKey: false,
              render: (_, row) => {
                const id = deviceID(row)
                return (
                  <TableActions>
                    <TableAction disabled={!id || !deviceHost(row) || !devicePath(row)} onClick={() => navigate(deviceDetailPath(id, deviceHost(row), devicePath(row)))}>详情</TableAction>
                  </TableActions>
                )
              }
            }
          ]}
        />
      </Card>
    </Page>
  )
}

export function DeviceDetailPage() {
  const [params] = useSearchParams()
  const deviceId = params.get('device_id') ?? ''
  const hostname = params.get('hostname') ?? ''
  const path = params.get('path') ?? ''
  const { selectedClusterId } = useClusterContext()
  return <DeviceDetailContent key={JSON.stringify([selectedClusterId, deviceId, hostname, path])} deviceId={deviceId} hostname={hostname} path={path} selectedClusterId={selectedClusterId} />
}

function DeviceDetailContent({ deviceId, hostname, path, selectedClusterId }: { deviceId: string; hostname: string; path: string; selectedClusterId?: number }) {
  const navigate = useNavigate()
  const active = useRef(true)
  const zapRunning = useRef(false)
  const identifyRunning = useRef(false)
  const deviceConfirmationEpoch = useRef(0)
  const identifyConfirmation = useRef<{ destroy: () => void } | null>(null)
  const zapConfirmation = useRef<{ destroy: () => void } | null>(null)
  useEffect(() => {
    active.current = true
    return () => { active.current = false; zapConfirmation.current?.destroy(); identifyConfirmation.current?.destroy() }
  }, [])
  const decodedDeviceId = deviceId
  const loader = useCallback(async () => {
    if (!selectedClusterId || !decodedDeviceId) {
      return {
        device: null,
        observedAt: null,
        stale: false,
        staleReason: null
      }
    }
    if (!hostname || !path) throw new Error('设备主机或路径缺失，请从设备列表重新选择目标。')
    const payload = await listAllResources('/devices', selectedClusterId, { filters: { device_id: [decodedDeviceId], hostname: [hostname], path: [path] } })
    const rows = payload.items.map(normalizeDeviceRow).filter((row) => deviceID(row) === decodedDeviceId && deviceHost(row) === hostname && devicePath(row) === path)
    if (rows.length > 1) {
      throw new Error('设备 ID 对应多条库存记录，无法唯一确定主机和路径；已阻止设备操作，请核对库存。')
    }
    return {
      device: rows[0] ?? null,
      observedAt: payload.observedAt,
      stale: payload.stale,
      staleReason: payload.staleReason
    }
  }, [decodedDeviceId, hostname, path, selectedClusterId])
  const { data, loading, error, refresh } = useResource(loader)
  const device = data?.device
  useEffect(() => () => {
    deviceConfirmationEpoch.current += 1
    zapConfirmation.current?.destroy()
    identifyConfirmation.current?.destroy()
  }, [data, loading, error])
  const currentDeviceHost = device ? deviceHost(device) : ''
  const currentDevicePath = device ? devicePath(device) : ''
  const [pendingDeviceAction, setPendingDeviceAction] = useState('')
  const [refreshingDevice, setRefreshingDevice] = useState(false)
  const operationMutation = useMutationOperation()

  async function refreshDeviceDetail() {
    if (!selectedClusterId || refreshingDevice) {
      return
    }
    setRefreshingDevice(true)
    try {
      await operationMutation.run(() => refreshResource({ clusterId: selectedClusterId, kind: 'device' }), '刷新成功')
      await refresh()
    } finally {
      setRefreshingDevice(false)
    }
  }

  function confirmIdentify(state: 'on' | 'off') {
    if (!device || !active.current || loading || error || device.stale !== false || zapRunning.current || identifyRunning.current) {
      return
    }
    const isOn = state === 'on'
    const confirmationEpoch = ++deviceConfirmationEpoch.current
    let submitted = false
    identifyConfirmation.current?.destroy()
    zapConfirmation.current?.destroy()
    identifyConfirmation.current = Modal.confirm({
      title: `${isOn ? '点灯' : '关灯'}设备 ${currentDeviceHost}:${currentDevicePath}`,
      content: `确认后将对主机 ${currentDeviceHost} 的设备 ${currentDevicePath} 执行${isOn ? '点灯' : '关灯'}操作。`,
      okText: `确认${isOn ? '点灯' : '关灯'}`,
      cancelText: '取消',
      onCancel() { if (deviceConfirmationEpoch.current === confirmationEpoch) deviceConfirmationEpoch.current += 1 },
      async onOk() {
        if (deviceConfirmationEpoch.current !== confirmationEpoch || submitted || !active.current || zapRunning.current || identifyRunning.current) return
        submitted = true
        await identify(state)
      }
    })
  }

  async function identify(state: 'on' | 'off', light: 'ident' | 'fault' = 'ident') {
    if (!selectedClusterId || !device || !active.current || loading || error || device.stale !== false || zapRunning.current || identifyRunning.current) {
      return
    }
    const pendingKey = `${currentDeviceHost}:${currentDevicePath}:identify:${state}:${light}`
    if (!currentDeviceHost || !currentDevicePath || pendingDeviceAction) {
      return
    }
    identifyRunning.current = true
    setPendingDeviceAction(pendingKey)
    try {
      await operationMutation.run(() => mutateResource('/device/identify', 'POST', {
        cluster_id: selectedClusterId,
        host: currentDeviceHost,
        device_id: decodedDeviceId,
        device: currentDevicePath,
        state,
        light
      }), false)
      if (active.current) {
        message.success('灯操作命令已执行，请核对实际灯状态')
        await refresh({ showLoading: false })
      }
    } finally {
      identifyRunning.current = false
      if (active.current) setPendingDeviceAction('')
    }
  }

  async function zap() {
    if (!selectedClusterId || !device || !active.current || loading || error || pendingDeviceAction || zapRunning.current || identifyRunning.current) {
      return
    }
    const raw = device.resource_version
    const generation = typeof raw === 'string' ? raw : typeof raw === 'number' && Number.isSafeInteger(raw) ? String(raw) : ''
    if (device.stale !== false || !/^[1-9][0-9]*$/.test(generation) || BigInt(generation) > 18446744073709551615n || !currentDeviceHost || !currentDevicePath) {
      message.error('设备身份或库存版本未确认，请刷新后再操作')
      return
    }
    const parameters = { cluster_id: selectedClusterId, host: currentDeviceHost, device: currentDevicePath }
    const pendingKey = `${currentDeviceHost}:${currentDevicePath}:zap`
    const confirmationEpoch = ++deviceConfirmationEpoch.current
    let submitted = false
    zapConfirmation.current?.destroy()
    identifyConfirmation.current?.destroy()
    zapConfirmation.current = Modal.confirm({
      title: `擦除设备 ${currentDeviceHost}:${currentDevicePath}`,
      content: `该操作会清理主机 ${currentDeviceHost} 的设备 ${currentDevicePath} 数据，为高风险操作，确认后将直接执行。`,
      okText: '提交擦除',
      okType: 'danger',
      cancelText: '取消',
      onCancel() { if (deviceConfirmationEpoch.current === confirmationEpoch) deviceConfirmationEpoch.current += 1 },
      async onOk() {
        if (deviceConfirmationEpoch.current !== confirmationEpoch || !active.current || submitted || zapRunning.current || identifyRunning.current) {
          return
        }
        submitted = true
        zapRunning.current = true
        setPendingDeviceAction(pendingKey)
        try {
          await operationMutation.run(() => mutateResource('/device/zap', 'POST', parameters, { ifMatch: generation }), false)
          if (active.current) {
            message.success('擦除命令已执行，请刷新并核对设备状态')
            await refresh({ showLoading: false })
          }
        } finally {
          zapRunning.current = false
          if (active.current) setPendingDeviceAction('')
        }
      }
    })
  }

  return (
    <Page title="设备详情" loading={loading} error={error}>
      <Space direction="vertical" size={16} className="page-stack">
        <Card
          className="page-surface-card"
          title="基础信息"
          extra={
            <Space className="host-detail-actions">
              <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/cluster/device')}>返回</Button>
              <Button icon={<ReloadOutlined />} loading={refreshingDevice || loading} onClick={refreshDeviceDetail}>刷新</Button>
            </Space>
          }
        >
          {device ? (
            <Descriptions className="host-detail-descriptions" size="small" column={twoColumnDescriptions} bordered>
              <Descriptions.Item label="主机">{textValue(device.hostname)}</Descriptions.Item>
              <Descriptions.Item label="路径">{textValue(device.path)}</Descriptions.Item>
              <Descriptions.Item label="类型">{textValue(device.device_type)}</Descriptions.Item>
              <Descriptions.Item label="容量">{textValue(device.size_display)}</Descriptions.Item>
              <Descriptions.Item label="厂商">{textValue(device.vendor)}</Descriptions.Item>
              <Descriptions.Item label="设备 ID">{textValue(device.device_id ?? device.id ?? device.name)}</Descriptions.Item>
              <Descriptions.Item label="序列号">{textValue(device.serial_number ?? device.serial)}</Descriptions.Item>
              <Descriptions.Item label="状态">{renderDeviceUsage(device)}</Descriptions.Item>
              <Descriptions.Item label="说明" span={2}>{renderDeviceNotes(device.usage_notes)}</Descriptions.Item>
              <Descriptions.Item label="创建时间">{formatDateTime(device.created_at)}</Descriptions.Item>
              <Descriptions.Item label="更新时间">{formatDateTime(device.updated_at)}</Descriptions.Item>
            </Descriptions>
          ) : (
            <Text type="secondary">暂无设备详情</Text>
          )}
        </Card>

        {device && <DeviceHardwareSummary device={device} stale={data?.stale} />}

        <Card className="page-surface-card" title="设备操作">
          <Space wrap>
            <Button
              icon={<BulbOutlined />}
              loading={pendingDeviceAction === `${currentDeviceHost}:${currentDevicePath}:identify:on:ident`}
              disabled={!device || Boolean(pendingDeviceAction)}
              onClick={() => confirmIdentify('on')}
            >
              点灯
            </Button>
            <Button
              icon={<PoweroffOutlined />}
              loading={pendingDeviceAction === `${currentDeviceHost}:${currentDevicePath}:identify:off:ident`}
              disabled={!device || Boolean(pendingDeviceAction)}
              onClick={() => confirmIdentify('off')}
            >
              关灯
            </Button>
            <Button
              danger
              icon={<DeleteOutlined />}
              loading={pendingDeviceAction === `${currentDeviceHost}:${currentDevicePath}:zap`}
              disabled={!device || Boolean(pendingDeviceAction)}
              onClick={zap}
            >
              擦除
            </Button>
          </Space>
        </Card>
      </Space>
    </Page>
  )
}

export function MdsManagementPage() {
  const { selectedClusterId } = useClusterContext()
  const serviceScope = useMemo(() => ({ clusterId: selectedClusterId, active: true, refreshing: false }), [selectedClusterId])
  useEffect(() => {
    serviceScope.active = true
    return () => { serviceScope.active = false }
  }, [serviceScope])
  const [refreshingScope, setRefreshingScope] = useState<typeof serviceScope | null>(null)
  const refreshingInventory = refreshingScope === serviceScope
  const [serviceSelection, setServiceSelection] = useState<{ scope: typeof serviceScope; name: string } | null>(null)
  const visibleService = serviceSelection?.scope === serviceScope ? serviceSelection : null
  const serviceTableFilters = useResourceTableFilters({
    path: '/services',
    fields: ['name', 'type', 'running', 'size'],
    clusterId: selectedClusterId
  })
  const daemonTableFilters = useResourceTableFilters({
    path: '/daemons',
    fields: ['name', 'type', 'hostname', 'status', 'version'],
    clusterId: selectedClusterId
  })
  const loader = useCallback(async () => {
    if (!selectedClusterId) {
      return { services: [], daemons: [], inventoryWarnings: [] }
    }
    const [services, daemons] = await Promise.all([
      listAllResources('/services', selectedClusterId, {
        filters: mergeResourceFilters(serviceTableFilters.filters, { type: ['mds'] })
      }),
      listAllResources('/daemons', selectedClusterId, {
        filters: mergeResourceFilters(daemonTableFilters.filters, { type: ['mds'] })
      })
    ])
    return {
      services: services.items,
      daemons: daemons.items,
      inventoryWarnings: [
        services.stale !== false ? `MDS 服务库存：${services.staleReason || '已过期或新鲜度未知'}` : null,
        daemons.stale !== false ? `MDS 守护进程库存：${daemons.staleReason || '已过期或新鲜度未知'}` : null
      ].filter((value): value is string => value !== null)
    }
  }, [daemonTableFilters.filters, selectedClusterId, serviceTableFilters.filters])
  const { data, loading, error, refresh } = useResource(loader)

  async function collectMDSInventory() {
    if (!selectedClusterId || !serviceScope.active || serviceScope.refreshing || loading) return
    serviceScope.refreshing = true
    setRefreshingScope(serviceScope)
    try {
      await refreshResource({ clusterId: selectedClusterId, kinds: ['service', 'daemon'] })
      if (serviceScope.active) {
        message.success('服务与守护进程采集完成')
        await refresh()
      }
    } catch {
      if (serviceScope.active) message.error('采集未完成，请检查任务结果后重试')
    } finally {
      serviceScope.refreshing = false
      if (serviceScope.active) setRefreshingScope(null)
    }
  }

  return (
    <Page title="MDS管理" loading={loading} error={error}>
      <Modal title={visibleService ? `${visibleService.name} 守护进程` : '服务守护进程'} open={Boolean(visibleService)} onCancel={() => setServiceSelection(null)} footer={null} width="95vw" destroyOnClose>
        {visibleService && selectedClusterId && <ServiceDaemons key={JSON.stringify([selectedClusterId, visibleService.name])} clusterId={selectedClusterId} name={visibleService.name} />}
      </Modal>
      {data?.inventoryWarnings.map(warning => <Alert key={warning} type="warning" showIcon message={warning} />)}
      <Card className="page-surface-card" title="MDS管理" extra={<Button icon={<ReloadOutlined />} disabled={!selectedClusterId || loading} loading={refreshingInventory} onClick={() => void collectMDSInventory()}>重新采集服务与守护进程</Button>}>
        <Tabs
          items={[
            {
              key: 'services',
              label: 'MDS服务',
              children: (
                <div className="embedded-panel">
                <DataTable
                  data={data?.services ?? []}
                  filterOptions={serviceTableFilters.filterOptions}
                  filteredValues={serviceTableFilters.filters}
                  onFilterChange={serviceTableFilters.handleFilterChange}
                  rowKeyCandidates={['name']}
                  columns={[
                    { key: 'name', title: '服务名' },
                    { key: 'placement', title: '放置策略', filterKey: false },
                    { key: 'running', title: '运行数' },
                    { key: 'size', title: '目标数' },
                    { key: 'deployment_details', title: '部署详情', filterKey: false, ellipsis: false, render: (_, row) => <ServiceInventoryDetails row={row} /> },
                    { key: 'service_daemons', title: '服务实例', filterKey: false, render: (_, row) => <TableAction
                      disabled={!selectedClusterId || loading || Boolean(error) || typeof row.name !== 'string' || !/^mds\.[A-Za-z0-9_.-]{1,252}$/.test(row.name)}
                      onClick={() => { if (selectedClusterId && !loading && !error && typeof row.name === 'string' && /^mds\.[A-Za-z0-9_.-]{1,252}$/.test(row.name)) setServiceSelection({ scope: serviceScope, name: row.name }) }}>查看守护进程</TableAction> }
                  ]}
                />
                </div>
              )
            },
            {
              key: 'daemons',
              label: '守护进程',
              children: <DaemonTable key={selectedClusterId} clusterId={selectedClusterId} unavailable={loading || Boolean(error)} data={data?.daemons ?? []} refresh={refresh} tableFilters={daemonTableFilters} />
            }
          ]}
        />
      </Card>
    </Page>
  )
}

function DaemonTable({
  clusterId,
  unavailable,
  data,
  refresh,
  tableFilters
}: {
  clusterId?: number
  unavailable: boolean
  data: ApiRecord[]
  refresh: () => void
  tableFilters?: ReturnType<typeof useResourceTableFilters>
}) {
  const [pendingDaemonAction, setPendingDaemonAction] = useState('')
  const [perfSelection, setPerfSelection] = useState<{ clusterId: number; name: string } | null>(null)
  const visiblePerf = perfSelection?.clusterId === clusterId ? perfSelection : null
  const active = useRef(true)
  const running = useRef(false)
  const actionConfirmation = useRef<{ destroy: () => void } | null>(null)
  useEffect(() => {
    active.current = true
    return () => { active.current = false; actionConfirmation.current?.destroy() }
  }, [])
  useEffect(() => () => { actionConfirmation.current?.destroy(); actionConfirmation.current = null }, [data, unavailable])
  const operationMutation = useMutationOperation()

  function confirmAction(row: ApiRecord, action: 'start' | 'stop' | 'restart' | 'reconfig' | 'redeploy' | 'rotate-key') {
    if (!clusterId || !active.current || running.current || unavailable || row.stale !== false || !osdInventoryVersion(row.resource_version) || typeof row.name !== 'string' || !row.name) return
    const label = { start: '启动', stop: '停止', restart: '重启', reconfig: '重新配置', redeploy: '重新部署', 'rotate-key': '轮换密钥' }[action]
    let submitted = false
    actionConfirmation.current?.destroy()
    const confirmation = Modal.confirm({
      title: `${label} ${row.name}（集群 ${clusterId}）`,
      content: '操作可能中断依赖此守护进程的客户端或服务。不会使用强制选项绕过 Ceph 安全检查；命令接受不代表运行状态已经完成切换。',
      okText: `确认${label}`, cancelText: '取消', okType: action === 'start' ? 'primary' : 'danger',
      onCancel() { if (actionConfirmation.current === confirmation) actionConfirmation.current = null },
      async onOk() {
        if (actionConfirmation.current !== confirmation || submitted || !active.current || running.current) return
        submitted = true
        try { await runAction(row, action) }
        finally {
          confirmation.destroy()
          if (actionConfirmation.current === confirmation) actionConfirmation.current = null
        }
      }
    })
    actionConfirmation.current = confirmation
  }

  function openPerformance(row: ApiRecord) {
    if (!clusterId || !active.current || unavailable || typeof row.name !== 'string' || !/^(mgr|mds)\.[A-Za-z0-9_][A-Za-z0-9_.-]*$/.test(row.name)) return
    setPerfSelection({ clusterId, name: row.name })
  }

  async function runAction(row: ApiRecord, action: string) {
    const name = textValue(row.name, '')
    const generation = osdInventoryVersion(row.resource_version)
    const pendingKey = `${name}:${action}`
    if (!clusterId || !active.current || running.current || unavailable || row.stale !== false || !generation || !name || pendingDaemonAction) {
      return
    }
    running.current = true
    setPendingDaemonAction(pendingKey)
    try {
      await operationMutation.run(() => applyDaemonAction(name, action, clusterId, generation), false)
      if (active.current) {
        message.success(`Daemon ${action} 命令已接受，正在重新采集状态`)
      }
    } finally {
      try {
        if (active.current) {
          try { await refreshResource({ clusterId, kinds: ['service', 'daemon'] }) }
          catch { if (active.current) message.warning('操作后的重新采集失败，请核对实际状态，不要直接重复提交。') }
          if (active.current) await refresh()
        }
      } finally {
        running.current = false
        if (active.current) setPendingDaemonAction('')
      }
    }
  }

  return (
    <div className="embedded-panel">
      <Modal title={visiblePerf ? `${visiblePerf.name} 性能计数器` : '性能计数器'} open={Boolean(visiblePerf)} onCancel={() => setPerfSelection(null)} footer={null} width="95vw" destroyOnClose>
        {visiblePerf && <DaemonPerf key={JSON.stringify([visiblePerf.clusterId, visiblePerf.name])} clusterId={visiblePerf.clusterId} name={visiblePerf.name} />}
      </Modal>
      <DataTable
        data={data}
        filterOptions={tableFilters?.filterOptions}
        filteredValues={tableFilters?.filters}
        onFilterChange={tableFilters?.handleFilterChange}
        rowKeyCandidates={['name']}
        columns={[
          { key: 'name', title: 'Daemon' },
          { key: 'type', title: '类型' },
          { key: 'hostname', title: '主机' },
          { key: 'status', title: '状态' },
          { key: 'version', title: '版本' },
          { key: 'cpu_percentage', title: 'CPU 用量（原值）', filterKey: false },
          { key: 'memory_usage', title: '内存用量（精确字节）', filterKey: false },
          { key: 'container_image', title: '容器镜像', filterKey: false },
          { key: 'last_refresh', title: 'Ceph 最近刷新', filterKey: false },
          { key: 'observed_at', title: '库存快照', filterKey: false, render: (_, row) => <Space direction="vertical">
            <Text>{formatDateTime(row.observed_at)}</Text>
            {row.stale !== false && <Text type="warning">已过期或新鲜度未知</Text>}
          </Space> },
          { key: 'runtime_details', title: '运行详情与事件', filterKey: false, ellipsis: false, render: (_, row) => <DaemonRuntimeDetails row={row} daemonType={row.type} /> },
          {
            key: 'actions',
            title: '操作',
            filterKey: false,
            render: (_, row) => {
              const name = textValue(row.name, '')
              const disabled = !clusterId || unavailable || row.stale !== false || !osdInventoryVersion(row.resource_version) || Boolean(pendingDaemonAction)
              return (
                <TableActions>
                  <TableAction disabled={!clusterId || unavailable || !/^(mgr|mds)\.[A-Za-z0-9_][A-Za-z0-9_.-]*$/.test(name)} onClick={() => openPerformance(row)}>性能计数器</TableAction>
                  <TableAction loading={pendingDaemonAction === `${name}:restart`} disabled={disabled} onClick={() => confirmAction(row, 'restart')}>重启</TableAction>
                  <TableAction loading={pendingDaemonAction === `${name}:start`} disabled={disabled} onClick={() => confirmAction(row, 'start')}>启动</TableAction>
                  <TableAction danger loading={pendingDaemonAction === `${name}:stop`} disabled={disabled} onClick={() => confirmAction(row, 'stop')}>停止</TableAction>
                  <TableAction danger loading={pendingDaemonAction === `${name}:reconfig`} disabled={disabled} onClick={() => confirmAction(row, 'reconfig')}>重新配置</TableAction>
                  <TableAction danger loading={pendingDaemonAction === `${name}:redeploy`} disabled={disabled} onClick={() => confirmAction(row, 'redeploy')}>重新部署</TableAction>
                  <TableAction danger loading={pendingDaemonAction === `${name}:rotate-key`} disabled={disabled} onClick={() => confirmAction(row, 'rotate-key')}>轮换密钥</TableAction>
                </TableActions>
              )
            }
          }
        ]}
      />
    </div>
  )
}

function osdInventoryVersion(value: unknown): string | null {
  if (typeof value === 'number') return Number.isSafeInteger(value) && value > 0 ? String(value) : null
  return typeof value === 'string' && /^[1-9]\d*$/.test(value) && BigInt(value) <= 18446744073709551615n ? value : null
}

function osdStateText(value: unknown): string {
  if (!Array.isArray(value) || value.some(flag => typeof flag !== 'string' || !flag.trim())) return '未采集或格式无效'
  return value.length ? value.join('、') : '本次未返回状态标记'
}

function osdReweightInitial(value: unknown): number | undefined {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0 && value <= 1 ? value : undefined
}

function osdReweightPreview(value: unknown): string {
  if (typeof value !== 'number' || !Number.isFinite(value) || value < 0 || value > 1) return '请输入 0 到 1 之间的有限数值以查看原生权重。'
  const effective = Math.trunc(value * 65536) / 65536
  const summary = `Ceph 实际权重：${effective}（按 1/65536 精度截断）。`
  return effective === 0 ? `${summary}注意：实际权重为零，OSD 将处于 Out，可能触发数据迁移。` : summary
}

function ReweightForm({ currentWeight, version, clusterId, isCurrent, osdID, refresh }: { currentWeight?: unknown; version: string; clusterId: number; isCurrent: () => boolean; osdID: string; refresh: (options?: { showLoading?: boolean }) => void }) {
  const running = useRef(false)
  const [submitting, setSubmitting] = useState(false)
  const operationMutation = useMutationOperation()

  async function submit(values: { weight: number }) {
    if (running.current || !isCurrent()) {
      return
    }
    if (!Number.isFinite(values.weight) || values.weight < 0 || values.weight > 1) {
      message.error('OSD 权重必须是 0 到 1 之间的有限数值')
      return
    }
    running.current = true
    setSubmitting(true)
    try {
      await operationMutation.run(() => reweightOSD(clusterId, osdID, values.weight, version), false)
      if (!isCurrent()) return
      message.success('OSD 权重调整执行成功')
      await refreshResource({ clusterId, kind: 'osd' })
      if (isCurrent()) await refresh({ showLoading: false })
    } finally {
      running.current = false
      setSubmitting(false)
    }
  }

  return (
    <Form layout="vertical" initialValues={{ weight: osdReweightInitial(currentWeight) }} onFinish={submit}>
      <Typography.Paragraph>目标：集群 {clusterId} / OSD {osdID}。调权可能触发数据迁移；这是 OSD 调权系数，不是 CRUSH 容量权重。版本冲突或切换集群后请重新采集并打开此表单。</Typography.Paragraph>
      <Form.Item name="weight" label="权重" rules={[{ required: true }]}>
        <InputNumber min={0} max={1} step={0.01} />
      </Form.Item>
      <Form.Item noStyle shouldUpdate>
        {({ getFieldValue }) => <Typography.Paragraph type="warning">{osdReweightPreview(getFieldValue('weight'))}</Typography.Paragraph>}
      </Form.Item>
      <Button type="primary" htmlType="submit" loading={submitting}>
        保存
      </Button>
    </Form>
  )
}

function osdID(row: ApiRecord) {
  return textValue(row.id ?? row.osd ?? row.service_id ?? row.name, '')
}

function osdDeploymentPayload(values: {
  service_id?: string
  host_pattern?: string
  all?: boolean
  paths?: string
  rotational?: boolean
  model?: string
  vendor?: string
  size?: string
}, clusterId: number) {
  const dataDevices: ApiRecord = {}
  if (values.all) {
    dataDevices.all = true
  }
  const paths = splitCSV(values.paths)
  if (paths.length > 0) {
    dataDevices.paths = paths
  }
  if (typeof values.rotational === 'boolean') {
    dataDevices.rotational = values.rotational
  }
  if (values.model) {
    dataDevices.model = values.model
  }
  if (values.vendor) {
    dataDevices.vendor = values.vendor
  }
  if (values.size) {
    dataDevices.size = values.size
  }
  if (Object.keys(dataDevices).length === 0) {
    dataDevices.all = true
  }
  return {
    cluster_id: clusterId,
    ...(values.service_id ? { service_id: values.service_id } : {}),
    ...(values.host_pattern ? { host_pattern: values.host_pattern } : {}),
    data_devices: dataDevices
  }
}

function splitCSV(value?: string) {
  return value?.split(',').map((item) => item.trim()).filter(Boolean) ?? []
}

function deviceHost(row: ApiRecord) {
  return textValue(row.hostname ?? row.host, '')
}

function devicePath(row: ApiRecord) {
  return textValue(row.path ?? row.device ?? row.name, '')
}

function deviceID(row: ApiRecord) {
  return textValue(row.device_id ?? row.id ?? row.natural_key, '')
}

function deviceDetailPath(id: string, hostname: string, path: string) {
  return `/cluster/device/detail?${new URLSearchParams({ device_id: id, hostname, path }).toString()}`
}

function normalizeDeviceRow(row: ApiRecord): ApiRecord {
  const usage = deviceUsage(row)
  const type = textValue(row.device_type ?? row.type, '')
  return {
    ...row,
    hostname: textValue(row.hostname ?? row.host, ''),
    path: devicePath(row),
    device_type: type ? type.toUpperCase() : '-',
    usage_state: usage.state,
    usage_label: usage.label,
    usage_notes: usage.notes,
    size_display: hostStorageCapacity([{ size_bytes: row.size_bytes ?? row.size }])
  }
}

function deviceUsage(row: ApiRecord) {
  const reasons = deviceReasonValues(row.rejected_reasons ?? row.reject_reasons ?? row.reasons)
  if (row.available === true) {
    return { state: 'available' as DeviceScope, label: '空闲可用', notes: reasons.map(readableDeviceReason) }
  }
  if (row.available !== false) {
    return { state: 'unknown' as DeviceScope, label: '可用性未知', notes: reasons.map(readableDeviceReason) }
  }
  if (reasons.some(isUsedDeviceReason)) {
    return { state: 'used' as DeviceScope, label: '已占用', notes: reasons.map(readableDeviceReason) }
  }
  return { state: 'unavailable' as DeviceScope, label: '不可用', notes: reasons.map(readableDeviceReason) }
}

function deviceReasonValues(value: unknown) {
  if (Array.isArray(value)) {
    return value.map((item) => textValue(item, '')).filter(Boolean)
  }

  const text = textValue(value, '')
  if (!text) {
    return []
  }

  return text.split(',').map((item) => item.trim()).filter(Boolean)
}

function isUsedDeviceReason(reason: string) {
  const normalized = reason.toLowerCase()
  return [
    'filesystem',
    'file system',
    'lvm',
    'mounted',
    'partition',
    'bluestore',
    'osd',
    'in use',
    'being used'
  ].some((keyword) => normalized.includes(keyword))
}

function readableDeviceReason(reason: string) {
  const summary = deviceReasonSummary(reason)
  return summary === reason ? reason : `${summary}（原文：${reason}）`
}

function deviceReasonSummary(reason: string) {
  const normalized = reason.toLowerCase()
  if (normalized.includes('filesystem') || normalized.includes('file system')) {
    return '已有文件系统'
  }
  if (normalized.includes('lvm')) {
    return '已有 LVM'
  }
  if (normalized.includes('insufficient space')) {
    return 'VG 可用空间不足'
  }
  if (normalized.includes('mounted')) {
    return '已挂载'
  }
  if (normalized.includes('partition')) {
    return '已有分区'
  }
  if (normalized.includes('bluestore') || normalized.includes('osd')) {
    return '已有 OSD 数据'
  }
  if (normalized.includes('locked')) {
    return '设备被锁定'
  }
  if (normalized.includes('read-only') || normalized.includes('readonly')) {
    return '只读设备'
  }
  return reason
}

function renderDeviceUsage(row: ApiRecord) {
  const state = textValue(row.usage_state, '') as DeviceScope
  const label = textValue(row.usage_label)
  const colors: Record<DeviceScope, string> = {
    available: 'success',
    used: 'processing',
    unknown: 'warning',
    unavailable: 'default'
  }
  return <Tag color={colors[state] ?? 'default'}>{label}</Tag>
}

function renderDeviceNotes(value: unknown) {
  if (!Array.isArray(value) || value.length === 0) {
    return '-'
  }

  return value.map((item) => <Tag key={textValue(item)}>{textValue(item)}</Tag>)
}

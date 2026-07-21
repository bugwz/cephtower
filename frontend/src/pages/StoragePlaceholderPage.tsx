import {
  ApiOutlined,
  BellOutlined,
  CloudServerOutlined,
  ClusterOutlined,
  DatabaseOutlined,
  FileTextOutlined,
  HddOutlined,
  PauseCircleOutlined,
  PlusOutlined,
  ReloadOutlined,
  SettingOutlined
} from '@ant-design/icons'
import { Button, Card, Descriptions, List, Progress, Space, Steps, Table, Tabs, Tag, Typography, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import type { ReactNode } from 'react'
import { Page } from '../components/Page'
import { findNavPage, findNavSection, type PageKey } from '../navigation'

const { Text } = Typography

type Tone = 'success' | 'processing' | 'warning' | 'error' | 'default'

type Metric = {
  label: string
  value: string
  note: string
  percent?: number
}

type Row = Record<string, ReactNode> & {
  key: string
  name: ReactNode
  status?: ReactNode
}

type Section = {
  key: string
  title: string
  note: string
  rows: Row[]
  columns: ColumnsType<Row>
}

type Workflow = {
  title: string
  steps: string[]
}

type Definition = {
  title: string
  summary: string
  icon: ReactNode
  primaryAction: string
  reference: string
  metrics: Metric[]
  sections: Section[]
  workflows: Workflow[]
  capabilities: string[]
}

export function StoragePlaceholderPage({ pageKey }: { pageKey: PageKey }) {
  const definition = definitions[pageKey] ?? fallbackDefinition(pageKey)
  const section = findNavSection(pageKey)

  function showPlaceholderMessage(action: string) {
    message.info(`${action}已预留，后续接入对应后端 API 后启用`)
  }

  return (
    <Page title={definition.title}>
      <Card
        className="page-surface-card storage-placeholder-page"
        title={definition.title}
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => showPlaceholderMessage('刷新')}>
              刷新
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => showPlaceholderMessage(definition.primaryAction)}>
              {definition.primaryAction}
            </Button>
          </Space>
        }
      >
        <section className="placeholder-summary">
          <div className="placeholder-summary-icon">{definition.icon}</div>
          <div>
            <Text strong>{section?.label ?? 'CephTower'}</Text>
            <Text type="secondary">{definition.summary}</Text>
          </div>
          <Tag color="gold" icon={<ApiOutlined />}>
            不调用 Go API
          </Tag>
        </section>

        <div className="metrics-grid placeholder-metrics">
          {definition.metrics.map((metric) => (
            <Card key={metric.label}>
              <Text type="secondary" className="metric-label">
                {metric.label}
              </Text>
              <div className="placeholder-stat-value">{metric.value}</div>
              {typeof metric.percent === 'number' ? <Progress percent={metric.percent} showInfo={false} strokeColor="#43bf8f" /> : null}
              <Text type="secondary">{metric.note}</Text>
            </Card>
          ))}
        </div>

        <Tabs
          items={definition.sections.map((sectionItem) => ({
            key: sectionItem.key,
            label: sectionItem.title,
            children: (
              <section className="embedded-panel">
                <div className="embedded-panel-title">
                  <Text strong>{sectionItem.title}</Text>
                  <Text type="secondary" className="card-subtitle">
                    {sectionItem.note}
                  </Text>
                </div>
                <Table
                  size="middle"
                  rowKey="key"
                  columns={sectionItem.columns}
                  dataSource={sectionItem.rows}
                  pagination={false}
                  scroll={{ x: true }}
                />
              </section>
            )
          }))}
        />

        <div className="content-grid two-columns placeholder-lower-grid">
          <section className="embedded-panel">
            <div className="embedded-panel-title">
              <Text strong>功能流程占位</Text>
              <Text type="secondary" className="card-subtitle">
                参考 Ceph Dashboard 的创建、详情、配置与操作路径
              </Text>
            </div>
            <div className="placeholder-workflow-list">
              {definition.workflows.map((workflow) => (
                <div key={workflow.title} className="placeholder-workflow">
                  <Text strong>{workflow.title}</Text>
                  <Steps
                    size="small"
                    direction="vertical"
                    current={workflow.steps.length - 1}
                    items={workflow.steps.map((step) => ({ title: step }))}
                  />
                </div>
              ))}
            </div>
          </section>

          <section className="embedded-panel">
            <div className="embedded-panel-title">
              <Text strong>模块能力预留</Text>
              <Text type="secondary" className="card-subtitle">
                这些区域后续可逐步替换为真实接口数据和表单
              </Text>
            </div>
            <Descriptions size="small" column={1}>
              <Descriptions.Item label="参考路径">{definition.reference}</Descriptions.Item>
              <Descriptions.Item label="当前阶段">
                <Tag color="gold">纯前端展示</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="接口调用">
                <Tag>无</Tag>
              </Descriptions.Item>
            </Descriptions>
            <List
              className="placeholder-capability-list"
              size="small"
              dataSource={definition.capabilities}
              renderItem={(item) => <List.Item>{item}</List.Item>}
            />
          </section>
        </div>
      </Card>
    </Page>
  )
}

const resourceColumns: ColumnsType<Row> = [
  { title: '名称', dataIndex: 'name', width: 230, render: renderName },
  { title: '状态', dataIndex: 'status', width: 120 },
  { title: '位置/归属', dataIndex: 'scope', width: 190 },
  { title: '容量/规模', dataIndex: 'size', width: 160 },
  { title: '配置', dataIndex: 'config', width: 260 },
  { title: '操作占位', dataIndex: 'actions', width: 220 }
]

const policyColumns: ColumnsType<Row> = [
  { title: '策略/配置', dataIndex: 'name', width: 240, render: renderName },
  { title: '状态', dataIndex: 'status', width: 120 },
  { title: '匹配范围', dataIndex: 'scope', width: 220 },
  { title: '规则', dataIndex: 'config', width: 320 },
  { title: '后续操作', dataIndex: 'actions', width: 220 }
]

const monitorColumns: ColumnsType<Row> = [
  { title: '名称', dataIndex: 'name', width: 260, render: renderName },
  { title: '状态/级别', dataIndex: 'status', width: 130 },
  { title: '来源', dataIndex: 'scope', width: 180 },
  { title: '表达式/内容', dataIndex: 'config', width: 360 },
  { title: '处理占位', dataIndex: 'actions', width: 210 }
]

const definitions: Partial<Record<PageKey, Definition>> = {
  blockPools: {
    title: '块存储池',
    summary: '按 Ceph pool 与 RBD 应用标签组织展示，预留副本池、纠删码池、PG 自动伸缩和 CRUSH 规则管理。',
    icon: <DatabaseOutlined />,
    primaryAction: '新建存储池',
    reference: 'ceph/pool + ceph/block/rbd-configuration',
    metrics: metrics('RBD 池', '7', '4 个副本池，3 个纠删码池', 76),
    sections: [
      tableSection('pools', '存储池列表', '对应 Ceph Dashboard Pool 列表与 RBD application metadata。', [
        resource('rbd-replicated', 'HEALTH_OK', 'application: rbd', '384 TiB / size 3', 'pg_num 512，autoscale on，crush_rule replicated_rule', ['详情', '编辑', 'PG 调优']),
        resource('rbd-ec-data', 'HEALTH_OK', 'application: rbd', '1.2 PiB / EC 4+2', 'allow_ec_overwrites=true，compression passive', ['详情', '压缩', '配额']),
        resource('rbd-fast-meta', 'WARN', 'ssd device class', '24 TiB / size 3', 'PG 建议扩容，min_size 2', ['调整 PG', '查看告警'])
      ]),
      tableSection('rules', 'CRUSH 与应用配置', '预留 CRUSH rule、device class、pool flags 与 QoS 配置入口。', [
        policy('replicated_rule', '启用', 'host failure domain', 'chooseleaf host，device_class hdd', ['查看规则', '复制']),
        policy('ssd_rbd_rule', '启用', 'ssd failure domain', 'chooseleaf host，device_class ssd', ['查看规则', '绑定池']),
        policy('rbd_qos_default', '草稿', 'rbd application', 'iops limit / bps limit / read-write burst', ['编辑 QoS'])
      ])
    ],
    workflows: workflows(['选择池类型', '设置 PG/副本或 EC profile', '绑定 RBD application', '确认 CRUSH 与 autoscale']),
    capabilities: ['存储池详情页预留 PG 状态、使用量、应用元数据和压缩配置', '创建表单预留副本池、纠删码池、device class 与 CRUSH rule 字段', 'RBD QoS 配置后续可复用 Ceph Dashboard rbd-configuration 结构']
  },
  rbdImages: {
    title: 'RBD镜像',
    summary: '按 Ceph Dashboard RBD 模块拆分镜像、命名空间、快照、回收站、性能和配置。',
    icon: <HddOutlined />,
    primaryAction: '新建镜像',
    reference: 'ceph/block/rbd-list, rbd-details, rbd-snapshot-list, rbd-trash-list',
    metrics: metrics('镜像', '28', '含 6 个 clone 和 4 个 protected snapshot', 82),
    sections: [
      tableSection('images', '镜像列表', '预留 image 列表、特性、父镜像、命名空间与容量字段。', [
        resource('vm/prod-db-01', '可用', 'pool rbd-replicated', '2 TiB / 71%', 'layering, exclusive-lock, object-map, fast-diff', ['详情', '快照', '克隆', '扩容']),
        resource('images/golden-linux', '受保护', 'pool rbd-fast-meta', '320 GiB / 44%', 'parent none，protected snapshots 2', ['复制', '扁平化', '保护']),
        resource('backup/daily-cache', '映射中', 'pool rbd-ec-data', '8 TiB / 63%', 'krbd mapped clients 4，encryption enabled', ['取消映射', '迁移'])
      ]),
      tableSection('snapshots', '快照与回收站', '对应 snapshot list、trash move/restore/purge 的占位操作。', [
        policy('prod-db-01@daily-0721', 'protected', 'vm/prod-db-01', 'mirror snapshot，schedule daily，retention 14', ['克隆', '取消保护', '删除']),
        policy('golden-linux@v12', 'protected', 'images/golden-linux', 'template snapshot，children 12', ['查看子镜像', '复制']),
        policy('trash/rbd-old-ci', 'deferment', 'rbd-replicated', 'expires 2026-07-24，original image ci-runner-09', ['恢复', '清理'])
      ]),
      tableSection('performance', '性能与配置', '预留 RBD performance card 和 image 级配置列表。', [
        policy('prod-db-01 latency', '正常', 'image metrics', 'read 8.2k iops，write 3.1k iops，p95 4.8ms', ['查看曲线']),
        policy('rbd_cache', '继承', 'pool override', 'rbd_cache=true，rbd_cache_size=32MiB', ['编辑配置'])
      ])
    ],
    workflows: workflows(['选择存储池/命名空间', '输入容量和对象大小', '勾选 RBD 特性', '预览快照与镜像同步设置']),
    capabilities: ['镜像详情页预留 snapshots、children、configuration、performance tabs', '命名空间管理预留 create/delete 操作', '回收站预留 move、restore、purge 生命周期操作']
  },
  imageMirroring: {
    title: '镜像同步',
    summary: '按 RBD Mirroring 展示 overview、pool mode、peer、daemon、image sync health 与 bootstrap token。',
    icon: <ClusterOutlined />,
    primaryAction: '导入 Peer',
    reference: 'ceph/block/mirroring/overview, pool-list, daemon-list, image-list',
    metrics: metrics('同步池', '5', '2 个 pool 模式，3 个 image 模式', 68),
    sections: [
      tableSection('pools', 'Pool 同步模式', '预留 enable/disable、pool/image 模式切换与 peer 管理。', [
        resource('rbd-replicated', '健康', 'mode image', '18 images', 'peer remote-a，leader rbd-mirror.a，replay caught up', ['切换模式', '添加 Peer', '创建 Token']),
        resource('rbd-ec-data', '同步中', 'mode pool', '6 images', 'remote-b lag 3m，snap based mirroring', ['查看镜像', '重新同步']),
        resource('rbd-archive', '未启用', 'mode disabled', '0 images', '无 peer，等待配置', ['启用'])
      ]),
      tableSection('daemons', 'Mirror Daemons', '对应 daemon-list 中 service、hostname、health、last_seen。', [
        resource('rbd-mirror.a', 'up', 'host-01', '2.4k ops/min', 'leader，version 18.2，last heartbeat 12s', ['重启占位', '日志']),
        resource('rbd-mirror.b', 'up', 'host-02', '1.8k ops/min', 'standby，version 18.2，last heartbeat 18s', ['重启占位', '日志'])
      ])
    ],
    workflows: workflows(['启用池同步模式', '创建或导入 bootstrap token', '绑定远端 peer', '查看 image 级同步健康']),
    capabilities: ['Pool mirroring mode 编辑弹窗占位', 'Peer add/edit/delete 操作占位', 'Image sync error 和 daemon health 详情占位']
  },
  iscsi: {
    title: 'iSCSI',
    summary: '按 Ceph iSCSI 模块展示 Targets、Gateways、Portals、LUN、Initiator ACL 和 CHAP 设置。',
    icon: <CloudServerOutlined />,
    primaryAction: '新建 Target',
    reference: 'ceph/block/iscsi-target-list, iscsi-target-details, iscsi-setting',
    metrics: metrics('Targets', '6', '发布 17 个 LUN', 74),
    sections: [
      tableSection('targets', 'Targets 与 LUN', '预留 IQN、portal group、image settings 和 LUN 映射。', [
        resource('iqn.2026-07.com.cephtower:db01', '在线', 'iscsi-gw-a', '4 LUN / 12 TiB', 'CHAP enabled，ACL 6 initiators，multipath active', ['详情', '添加 LUN', 'Discovery']),
        resource('iqn.2026-07.com.cephtower:vmstore', '在线', 'iscsi-gw-b', '12 LUN / 48 TiB', 'mutual CHAP disabled，image writeback cache', ['编辑 IQN', 'ACL']),
        resource('iqn.2026-07.com.cephtower:backup', '待配置', 'iscsi-gw-a', '1 LUN / 8 TiB', '缺少 portal IP，认证未设置', ['补全 Portal'])
      ]),
      tableSection('gateways', 'Gateways 与 Portals', '对应 iSCSI gateway 状态和 portal 发现设置。', [
        resource('iscsi-gw-01', 'up', 'host-01', '2 portals', '10.10.10.21:3260，10.10.11.21:3260', ['查看', '维护']),
        resource('iscsi-gw-02', 'up', 'host-02', '2 portals', '10.10.10.22:3260，10.10.11.22:3260', ['查看', '维护'])
      ])
    ],
    workflows: workflows(['填写 Target IQN', '选择 RBD 镜像作为 LUN', '配置 Portal 和 Gateway', '设置 Initiator ACL/CHAP']),
    capabilities: ['Target 详情预留 LUN、portal、initiator、auth tabs', 'Discovery modal 和 image settings modal 占位', '全局 iSCSI settings 占位']
  },
  nvmeTcp: {
    title: 'NVMe/TCP',
    summary: '按 NVMe-oF 展示 Gateway Group、Subsystem、Namespace、Listener、Initiator 和 Host Key。',
    icon: <HddOutlined />,
    primaryAction: '新建 Subsystem',
    reference: 'ceph/block/nvmeof-tabs, nvmeof-subsystems, nvmeof-gateway',
    metrics: metrics('Subsystems', '9', '31 TiB namespace 容量', 81),
    sections: [
      tableSection('subsystems', 'Subsystems', '预留 NQN、Namespace、Listener 和 Host ACL。', [
        resource('nqn.2026-07.io.cephtower:fast-db', '在线', 'group nvme-fast', '6 namespaces', 'listeners 10.20.1.11:4420，hosts 8，ANA enabled', ['详情', '添加 Namespace', 'Host ACL']),
        resource('nqn.2026-07.io.cephtower:ml-cache', '在线', 'group nvme-cache', '3 namespaces', 'TLS host key configured，listener 10.20.1.12:4420', ['性能', '编辑 Listener']),
        resource('nqn.2026-07.io.cephtower:archive', '维护中', 'group nvme-cold', '2 namespaces', 'listener disabled，namespace expand pending', ['启用 Listener'])
      ]),
      tableSection('gateways', 'Gateway Groups', '对应 gateway group、gateway node 与 subsystem 关系视图。', [
        resource('nvme-fast', '健康', 'placement host-01 host-02', '2 gateways', 'load balancing active，service count matched', ['节点', '调度']),
        resource('nvme-cache', '健康', 'placement label:nvme', '3 gateways', 'gateway node mode active-active', ['节点', '扩容'])
      ])
    ],
    workflows: workflows(['创建 Gateway Group', '创建 Subsystem NQN', '添加 Namespace', '配置 Listener 与 Initiator']),
    capabilities: ['Subsystem 详情预留 overview、namespaces、performance tabs', 'Namespace expand modal 占位', 'Host key edit 和 initiator list 占位']
  },
  filePools: {
    title: '文件存储池',
    summary: '按 CephFS 所需 metadata/data pool 展示应用标签、PG、CRUSH 与快照调度相关预留。',
    icon: <DatabaseOutlined />,
    primaryAction: '新建文件池',
    reference: 'ceph/pool + ceph/cephfs/cephfs-form',
    metrics: metrics('CephFS 池', '8', '3 个 metadata 池，5 个 data 池', 72),
    sections: [
      tableSection('pools', 'CephFS 池', '预留元数据池、数据池和文件系统绑定关系。', [
        resource('cephfs_meta', 'HEALTH_OK', 'metadata pool', '12 TiB / size 3', 'application cephfs，fs cephfs-prod，pg_num 128', ['详情', '绑定 FS']),
        resource('cephfs_data_ssd', 'HEALTH_OK', 'data pool hot', '480 TiB / size 3', 'layout pool for hot directories，compression off', ['目录布局']),
        resource('cephfs_data_ec', 'HEALTH_OK', 'data pool cold', '2.4 PiB / EC 6+3', 'allow_ec_overwrites，used by archive subtree', ['EC 配置'])
      ]),
      tableSection('layouts', '目录布局占位', '对应 CephFS directory layout 和 subvolume pool placement。', [
        policy('/volumes/prod/hot', '启用', 'cephfs-prod', 'pool cephfs_data_ssd，stripe_unit 4MiB', ['编辑布局']),
        policy('/volumes/prod/archive', '启用', 'cephfs-prod', 'pool cephfs_data_ec，inherit quota', ['编辑布局'])
      ])
    ],
    workflows: workflows(['创建 metadata pool', '创建一个或多个 data pool', '绑定 CephFS application', '在 CephFS 中选择默认 data pool']),
    capabilities: ['文件池和 CephFS 创建表单联动占位', '目录 layout 与 subvolume group pool 配置占位', 'PG/autoscale 和 CRUSH 规则后续接入']
  },
  cephfs: {
    title: 'CephFS',
    summary: '按 CephFS 模块展示文件系统、MDS 状态、目录、客户端、子卷组、子卷和快照计划。',
    icon: <FileTextOutlined />,
    primaryAction: '新建文件系统',
    reference: 'ceph/cephfs/list, detail, directories, clients, subvolume, snapshotschedule',
    metrics: metrics('文件系统', '3', '5 个 active MDS，171 个客户端', 77),
    sections: [
      tableSection('filesystems', '文件系统与 MDS', '预留 FS 列表、rank、standby-replay 和池绑定。', [
        resource('cephfs-prod', 'HEALTH_OK', 'metadata cephfs_meta', 'data pools 2 / clients 128', 'max_mds 3，standby_count_wanted 1，rank 0-2 active', ['详情', '目录', '客户端']),
        resource('cephfs-ai', 'HEALTH_OK', 'metadata ai_meta', 'data pools 2 / clients 43', 'max_mds 2，standby replay enabled', ['详情', '子卷']),
        resource('cephfs-backup', 'WARN', 'metadata backup_meta', 'data pool backup_ec', 'standby replay missing，mds trim lag high', ['查看告警'])
      ]),
      tableSection('directories', '目录与客户端', '对应 directory browser、quota、layout、mount details 和 clients。', [
        resource('/volumes/prod/home', 'mounted', 'cephfs-prod', 'quota 120 TiB', 'layout cephfs_data_ssd，recursive bytes 86 TiB', ['配额', '布局', '授权']),
        resource('client.48291', 'active', '10.40.2.18', 'caps 231', 'root=/volumes/prod，session uptime 7d', ['驱逐', '查看 caps'])
      ]),
      tableSection('subvolumes', '子卷与快照计划', '对应 subvolume group、subvolume、snapshot 和 schedule。', [
        resource('tenants/team-a', '可用', 'group tenants', 'quota 20 TiB', 'snapshot schedule daily，retention 14', ['快照', '扩容']),
        resource('pipelines/features', '可用', 'group pipelines', 'quota 80 TiB', 'pool ai_hot，snapshot mirror disabled', ['快照', '授权']),
        resource('/volumes/prod@hourly', '启用', 'cephfs-prod', '24 retained', 'repeat hourly，start 00:00，path /volumes/prod', ['暂停', '编辑'])
      ])
    ],
    workflows: workflows(['选择 metadata/data pools', '设置 MDS 数量', '创建子卷组/子卷', '配置目录配额和快照计划']),
    capabilities: ['CephFS 详情 tabs 预留 clients、directories、subvolumes、snapshot schedules', 'Mount details 和 auth modal 占位', '目录树和配额/布局编辑占位']
  },
  nfs: {
    title: 'NFS',
    summary: '按 NFS Ganesha 展示 NFS Cluster、Export、FSAL、Pseudo Path 和客户端限制。',
    icon: <CloudServerOutlined />,
    primaryAction: '新建 Export',
    reference: 'ceph/nfs/nfs-cluster, nfs-list, nfs-details, nfs-form',
    metrics: metrics('Exports', '14', '3 个 NFS 集群', 69),
    sections: [
      tableSection('clusters', 'NFS 集群', '预留 nfs cluster placement、virtual IP 和 daemon 状态。', [
        resource('nfs-prod', '健康', 'placement host-01 host-02 host-03', '8 exports', 'virtual IP 10.30.0.50，backend cephadm service nfs.prod', ['详情', '扩容']),
        resource('nfs-archive', '健康', 'placement label:archive', '5 exports', 'virtual IP 10.30.0.60，NFSv4 only', ['详情', '维护'])
      ]),
      tableSection('exports', 'Exports', '对应 export list/form/details，包含 FSAL、pseudo、clients 和 squash。', [
        resource('/volumes/prod', '在线', 'cluster nfs-prod', 'CephFS path /volumes/prod', 'pseudo /prod，access rw，squash no_root_squash，clients 10.0.0.0/16', ['详情', '客户端']),
        resource('/archive', '在线', 'cluster nfs-archive', 'CephFS path /archive', 'pseudo /archive，access ro，root_id_squash', ['编辑', '导出']),
        resource('/s3-share', '待验证', 'cluster nfs-rgw', 'RGW bucket shared', 'FSAL RGW，user nfs-rgw，缺少 bucket policy', ['补全配置'])
      ])
    ],
    workflows: workflows(['创建 NFS Cluster', '选择 FSAL: CephFS 或 RGW', '填写 pseudo/path', '配置客户端、squash 和协议版本']),
    capabilities: ['NFS Cluster 详情和 Export 详情占位', '客户端访问控制表单占位', 'CephFS/RGW FSAL 参数占位']
  },
  smb: {
    title: 'SMB',
    summary: '按 SMB 模块展示 SMB Cluster、Share、Join Auth、Users/Groups 和 Domain 设置。',
    icon: <CloudServerOutlined />,
    primaryAction: '新建 Share',
    reference: 'ceph/smb/smb-tabs, smb-cluster-list, smb-share-list, smb-join-auth-list',
    metrics: metrics('Shares', '11', '2 个域集群，1 个 workgroup 集群', 73),
    sections: [
      tableSection('clusters', 'SMB 集群', '预留 clustering、domain/workgroup、join auth 和 daemon 状态。', [
        resource('smb-prod', '健康', 'domain corp.example', '2 daemons / 7 shares', 'join auth configured，CTDB enabled，clustered=true', ['详情', 'Domain 设置']),
        resource('smb-edge', '健康', 'workgroup CEPH', '1 daemon / 4 shares', 'local usersgroups，guest disabled', ['详情', '用户组'])
      ]),
      tableSection('shares', 'Shares', '对应 share list/form，包含路径、访问控制、browseable 和 CephFS 关联。', [
        resource('finance', '在线', 'cluster smb-prod', 'path /finance', 'readonly=false，browseable=false，AD group finance-rw', ['详情', 'ACL']),
        resource('engineering', '在线', 'cluster smb-prod', 'path /engineering', 'snapshot directory visible，dev group rw', ['编辑', '快照']),
        resource('public', '待发布', 'cluster smb-edge', 'path /public', 'guest access disabled，local group public-ro', ['发布'])
      ]),
      tableSection('auth', 'Join Auth 与用户组', '对应 join-auth 和 usersgroups 子模块。', [
        policy('corp-join-admin', '有效', 'domain corp.example', 'username svc-ceph-smb，secret saved', ['轮换密钥']),
        policy('local-users.yaml', '已导入', 'smb-edge', 'users 23，groups 5，last import 2026-07-20', ['重新导入'])
      ])
    ],
    workflows: workflows(['创建 SMB Cluster', '配置 Domain 或 Workgroup', '导入 Join Auth/用户组', '创建 Share 并绑定 CephFS 路径']),
    capabilities: ['SMB overview、cluster tabs、share list、join auth、usersgroups 占位', 'AD domain settings modal 占位', '本地用户组文件上传占位']
  },
  rgwOverview: {
    title: 'RGW总览',
    summary: '按 RGW Overview Dashboard 展示 Daemon、Zone、Usage、Sync 和服务实体创建入口。',
    icon: <DatabaseOutlined />,
    primaryAction: '新建服务实体',
    reference: 'ceph/rgw/rgw-overview-dashboard, rgw-daemon-list, create-rgw-service-entities',
    metrics: metrics('RGW Daemons', '12', '3 个 zone，平均 p95 11ms', 84),
    sections: [
      tableSection('daemons', 'RGW Daemons', '预留 daemon details、frontend、zonegroup/zone 和性能计数。', [
        resource('rgw.realm.zone-a.0', '运行中', 'host-01:8080', '18k req/min', 'zonegroup global，zone cn-east，beast frontend', ['详情', '性能']),
        resource('rgw.realm.zone-a.1', '运行中', 'host-02:8080', '16k req/min', 'version 18.2，ssl external proxy', ['详情', '日志']),
        resource('rgw.realm.zone-b.0', '维护中', 'host-04:8080', 'drain enabled', 'zone cn-north，orchestrator event pending', ['查看事件'])
      ]),
      tableSection('usage', 'Usage 与同步摘要', '对应 overview card 的 bucket、user、object、sync 统计。', [
        resource('STANDARD storage class', '增长中', 'global placement', '1.8 PiB / 2.4B objects', 'data pool rgw.buckets.data，compression off', ['查看 Bucket']),
        resource('metadata sync', 'caught up', 'realm-prod', '0 shards behind', 'last complete 35s，full sync false', ['同步详情']),
        resource('data sync cn-east -> dr', '同步中', 'realm-prod', 'lag 4m', 'recovering shards 3，incremental sync active', ['查看同步'])
      ])
    ],
    workflows: workflows(['检查 RGW 服务', '查看使用量和热点 Bucket', '确认 metadata/data sync', '进入 daemon 或 multisite 详情']),
    capabilities: ['Overview cards、daemon details、sync primary zone 占位', '创建 RGW service entities 向导占位', 'Performance counter 入口占位']
  },
  rgwUsers: {
    title: '用户管理',
    summary: '按 RGW User 与 User Accounts 展示用户、子用户、S3/Swift Key、Capability 和配额。',
    icon: <DatabaseOutlined />,
    primaryAction: '新建用户',
    reference: 'ceph/rgw/rgw-user-list, rgw-user-tabs, rgw-user-accounts',
    metrics: metrics('用户', '326', '42 个系统或服务账号', 78),
    sections: [
      tableSection('users', 'RGW 用户', '预留 user list/details 和 account 视图字段。', [
        resource('analytics-prod', '启用', 'tenant analytics', 'quota 800 TiB', 'caps buckets=*;usage=read，S3 keys 2，subusers 1', ['详情', '密钥', '能力']),
        resource('backup-service', '启用', 'tenant infra', 'quota unlimited', 'system user，placement default-placement', ['详情', '轮换 Key']),
        resource('legacy-swift', '已暂停', 'tenant legacy', 'quota 10 TiB', 'Swift keys 2，suspended=true', ['恢复', 'Swift Key'])
      ]),
      tableSection('keys', '密钥与能力', '对应 S3 key、Swift key、capability modal。', [
        policy('analytics-prod / AKIA...7F2', '有效', 'S3 key', 'created 2026-06-12，rotation due 40d', ['轮换', '禁用']),
        policy('legacy-swift:reader', '有效', 'Swift subuser', 'access read，secret saved', ['编辑子用户'])
      ])
    ],
    workflows: workflows(['填写 UID/Display Name', '设置配额和 suspended 状态', '创建 S3/Swift Key', '配置 caps 与 bucket ownership']),
    capabilities: ['User tabs: details、subusers、keys、caps、quota 占位', 'User Accounts 新模型占位', '系统用户和多租户字段占位']
  },
  bucketManagement: {
    title: 'Bucket管理',
    summary: '按 RGW Bucket 展示列表、详情、生命周期、版本、加密、通知、Topic、限速和 Tiering。',
    icon: <DatabaseOutlined />,
    primaryAction: '新建 Bucket',
    reference: 'ceph/rgw/rgw-bucket-list, bucket-details, lifecycle, notification, rate-limit',
    metrics: metrics('Buckets', '1,248', '216 个已开启版本控制', 82),
    sections: [
      tableSection('buckets', 'Bucket 列表', '预留 owner、placement、对象数、版本、加密和标签。', [
        resource('prod-logs', '启用', 'owner observability', '412 TiB / 980M objects', 'versioning enabled，lifecycle 90d archive，SSE-S3', ['详情', '生命周期', '通知']),
        resource('ml-datasets', '启用', 'owner analytics-prod', '1.2 PiB / 62M objects', 'replication to dr，storage class STANDARD+ARCHIVE', ['复制', '加密']),
        resource('tmp-ingest', '限速', 'owner data-platform', '84 TiB / 210M objects', 'rate limit 6k ops/s，MFA delete off', ['限速', '清理'])
      ]),
      tableSection('policies', '生命周期/通知/Topic', '对应 lifecycle list、notification list、topic list 和 bucket tag modal。', [
        policy('prod-logs lifecycle', '启用', 'bucket prod-logs', 'transition after 90d，expire after 730d，noncurrent 30d', ['编辑规则']),
        policy('ml-datasets replication', '同步中', 'bucket ml-datasets', 'topic dataset-events，event s3:ObjectCreated:*，target dr-zone', ['查看事件']),
        policy('tmp-ingest rate-limit', '启用', 'bucket tmp-ingest', 'max_read_ops 6000/s，max_write_bytes 4GiB/s', ['调整限速'])
      ])
    ],
    workflows: workflows(['创建 Bucket 并选择 Owner', '设置 Placement/Storage Class', '配置版本/加密/生命周期', '绑定通知 Topic 和限速']),
    capabilities: ['Bucket details tabs: usage、acl、lifecycle、notification、encryption 占位', 'Topic、notification、tag、tiering 表单占位', 'Bucket rate limit 详情占位']
  },
  gatewayManagement: {
    title: '网关管理',
    summary: '按 RGW Daemon 与服务实体展示网关实例、服务编排、zone 归属和 frontend 配置。',
    icon: <CloudServerOutlined />,
    primaryAction: '部署网关',
    reference: 'ceph/rgw/rgw-daemon-list, rgw-daemon-details, create-rgw-service-entities',
    metrics: metrics('网关实例', '12', '服务规格 rgw.realm.zone', 88),
    sections: [
      tableSection('gateways', '网关实例', '预留 daemon list/details、hostname、version、status 和 perf counter。', [
        resource('rgw-a-01', '运行中', 'host-01 / zone-a', '18k req/min', 'frontend beast port=8080，version 18.2.2', ['详情', '性能', '日志']),
        resource('rgw-a-02', '运行中', 'host-02 / zone-a', '16k req/min', 'ssl=false，behind load balancer', ['详情', '维护']),
        resource('rgw-b-01', '重启中', 'host-04 / zone-b', 'pending', 'orchestrator redeploy event，drain active', ['查看事件'])
      ]),
      tableSection('services', '服务实体', '对应 create rgw service entities 的 realm/zone/service 配置。', [
        policy('rgw.realm.zone-a', '健康', 'placement 4 hosts', 'ports 8080，networks 10.20.0.0/16，ssl external', ['编辑规格']),
        policy('rgw.realm.zone-b', '健康', 'placement 3 hosts', 'multisite secondary，sync from zone-a', ['扩容'])
      ])
    ],
    workflows: workflows(['选择 Realm/Zone', '设置 placement 和端口', '部署 RGW service', '查看 daemon details 和 perf counter']),
    capabilities: ['Daemon 详情和 Performance Counter 占位', 'RGW service entities 创建向导占位', 'Zone/Zonegroup 映射字段占位']
  },
  multisite: {
    title: '多站点',
    summary: '按 RGW Multisite 展示 Realm、Zonegroup、Zone、Period、Sync Policy、Metadata/Data Sync 和导入导出。',
    icon: <ClusterOutlined />,
    primaryAction: '新建 Realm',
    reference: 'ceph/rgw/rgw-multisite-tabs, sync-policy, import/export, wizard',
    metrics: metrics('Realm', '2', '跨 5 个 zone 复制', 71),
    sections: [
      tableSection('topology', 'Realm / Zonegroup / Zone', '预留 multisite topology、master zone 和 period 状态。', [
        resource('realm-prod', '健康', 'master zone cn-east', '3 zonegroups / 5 zones', 'period committed，metadata sync caught up', ['详情', '导出 Realm']),
        resource('global / cn-east', 'primary', 'zonegroup global', 'data sync source', 'read/write zone，endpoints 3', ['Zone 详情']),
        resource('global / dr-west', 'secondary', 'zonegroup global', 'lag 4m', 'data sync incremental，recovering shards 3', ['同步详情'])
      ]),
      tableSection('sync-policy', '同步策略', '对应 sync flow、sync pipe、policy details 和 wizard。', [
        policy('logs-bidirectional', '启用', 'bucket prod-logs', 'flow symmetrical，pipe all-zones，groups enabled', ['编辑 Flow', '编辑 Pipe']),
        policy('datasets-to-dr', '启用', 'bucket ml-datasets', 'direction cn-east -> dr-west，storage class remap ARCHIVE', ['查看策略']),
        policy('metadata full sync', '完成', 'realm-prod', 'full sync false，incremental true，behind shards 0', ['重新同步'])
      ])
    ],
    workflows: workflows(['创建或导入 Realm', '配置 Zonegroup/Zone', '提交 Period', '设置 Sync Policy 并检查同步状态']),
    capabilities: ['Multisite wizard/import/export 占位', 'Sync Policy flow/pipe modal 占位', 'Metadata/Data sync 详情占位']
  },
  objectStorageConfig: {
    title: '对象存储配置',
    summary: '按 RGW Configuration 展示配置项、Storage Class、Rate Limit、Topic、Notification 和系统用户。',
    icon: <SettingOutlined />,
    primaryAction: '新增配置',
    reference: 'ceph/rgw/rgw-configuration-page, storage-class, topic, rate-limit',
    metrics: metrics('配置项', '46', '含 7 个存储类别与 12 个 Topic', 66),
    sections: [
      tableSection('config', 'RGW 配置项', '预留 config details/modal，支持 service/zone/global scope。', [
        policy('rgw_frontends', '已设置', 'service rgw.realm.zone-a', 'beast port=8080 tcp_nodelay=1，restart no', ['编辑']),
        policy('rgw_gc_max_objs', '已设置', 'global', '32，garbage collection batch size', ['编辑']),
        policy('rgw_lc_debug_interval', '默认值', 'zone cn-east', 'unset，using cluster default', ['覆盖配置'])
      ]),
      tableSection('storage-class', '存储类别与 Topic', '对应 storage-class list、topic list、notification form。', [
        resource('STANDARD', '启用', 'default-placement', '1.8 PiB', 'data pool rgw.buckets.data，compression off', ['详情', '编辑']),
        resource('ARCHIVE', '启用', 'archive-placement', '6.2 PiB', 'data pool rgw.archive.ec，EC 8+3', ['详情', '编辑']),
        resource('dataset-events', '启用', 'SNS topic', '3 subscriptions', 'push endpoint kafka://events，opaque data configured', ['详情', '测试通知'])
      ])
    ],
    workflows: workflows(['选择配置 scope', '编辑 RGW 配置项', '维护 Storage Class', '创建 Topic 并绑定 Bucket Notification']),
    capabilities: ['RGW config details/modal 占位', 'Storage class details/form 占位', 'Topic 和 notification form 占位', 'Rate limit details 占位']
  },
  monitorOverview: {
    title: '监控总览',
    summary: '按 Ceph Dashboard 监控入口展示健康摘要、Prometheus/Alertmanager/Grafana 状态和关键面板。',
    icon: <BellOutlined />,
    primaryAction: '添加面板',
    reference: 'dashboard overview + shared alert-panel + grafana component',
    metrics: metrics('健康评分', '92%', 'HEALTH_WARN 3 项', 92),
    sections: [
      monitorSection('health', '健康摘要', '预留 health、manager module、Prometheus 和 Alertmanager 状态。', [
        monitorRow('Cluster health', 'HEALTH_WARN', 'ceph health', '2 pgs not deep-scrubbed in time；1 osd nearfull', ['查看详情', '关联告警']),
        monitorRow('Prometheus', '已配置', 'mgr/prometheus', 'scrape interval 15s，targets 128/128 up', ['打开指标']),
        monitorRow('Alertmanager', '已配置', 'alertmanager', 'active alerts 7，silences 5', ['告警列表']),
        monitorRow('Grafana', '已配置', 'grafana', 'dashboards: cluster, osd, rgw, cephfs', ['打开面板'])
      ]),
      monitorSection('panels', '关键面板', '预留 Grafana iframe/card 和 performance-card 风格入口。', [
        monitorRow('Cluster Capacity', '正常', 'Grafana', 'used 68%，available 2.4 PiB，nearfull osd 1', ['查看']),
        monitorRow('Client Throughput', '正常', 'PromQL', 'sum(rate(ceph_client_io_bytes[5m])) = 18.6GiB/s', ['查看']),
        monitorRow('RGW Latency', '关注', 'PromQL', 'histogram_quantile(0.95, rgw_req_latency) = 82ms', ['创建告警'])
      ])
    ],
    workflows: workflows(['读取健康摘要', '确认 Prometheus/Alertmanager/Grafana 配置', '展示关键面板', '跳转告警与指标详情']),
    capabilities: ['健康卡片和 warning panel 占位', 'Grafana 组件嵌入占位', 'Prometheus 配置状态占位']
  },
  performanceMetrics: {
    title: '性能指标',
    summary: '按 Performance Card 与 Prometheus 查询展示集群、OSD、MDS、RGW、RBD 指标占位。',
    icon: <BellOutlined />,
    primaryAction: '新建指标视图',
    reference: 'shared/performance-card + prometheus service + grafana component',
    metrics: metrics('吞吐', '18.6 GiB/s', '读写合计占位', 83),
    sections: [
      monitorSection('cluster', '集群指标', '预留 PromQL、图表、时间范围与刷新周期。', [
        monitorRow('Client IO bytes', '正常', 'PromQL', 'sum(rate(ceph_client_io_bytes[5m])) by (type)', ['图表', '复制 PromQL']),
        monitorRow('Client IOPS', '正常', 'PromQL', 'sum(rate(ceph_client_io_ops[5m])) = 812k ops/s', ['图表']),
        monitorRow('Recovery traffic', '偏高', 'PromQL', 'sum(rate(ceph_recovery_bytes[5m])) = 2.1GiB/s', ['关联事件'])
      ]),
      monitorSection('services', '服务指标', '按模块预留 RGW/MDS/RBD/OSD 性能入口。', [
        monitorRow('OSD commit latency', '正常', 'OSD', 'p95 commit 8ms，apply 10ms', ['OSD 详情']),
        monitorRow('MDS requests', '正常', 'MDS', 'handle_client_request 43k ops/s', ['MDS 详情']),
        monitorRow('RBD image latency', '正常', 'RBD', 'prod-db-01 p95 4.8ms', ['镜像详情'])
      ])
    ],
    workflows: workflows(['选择指标域', '选择时间范围', '生成 PromQL 查询', '渲染图表并关联资源详情']),
    capabilities: ['Prometheus query/card 占位', 'Grafana dashboard link 占位', '服务维度过滤器占位']
  },
  runtimeLogs: {
    title: '运行日志',
    summary: '按 Cluster Logs、Audit Logs、任务事件和通知侧边栏展示日志功能占位。',
    icon: <FileTextOutlined />,
    primaryAction: '导出日志',
    reference: 'ceph/cluster/logs + shared notifications-sidebar',
    metrics: metrics('日志事件', '128', '近 24 小时静态样例', 58),
    sections: [
      monitorSection('cluster', '集群日志', '预留 clog、priority、channel、daemon 和时间线。', [
        monitorRow('2026-07-21 10:24:18', 'WRN', 'cluster', '2 pgs not deep-scrubbed in time', ['查看 PG', '静默告警']),
        monitorRow('2026-07-21 10:11:03', 'INF', 'osd.12', 'recovery starts after osd.12 marked in', ['OSD 详情']),
        monitorRow('2026-07-21 09:58:44', 'INF', 'mgr.x', 'dashboard module status refreshed', ['Mgr 详情'])
      ]),
      monitorSection('audit', '审计日志与任务', '预留 audit log、task queue、finished task 和异常详情。', [
        monitorRow('user.admin', 'POST', 'audit', 'preview create RBD image form；no backend request sent', ['查看请求']),
        monitorRow('task.1024', 'running', 'orchestrator', 'redeploy rgw.realm.zone-b.0 progress 60%', ['任务详情']),
        monitorRow('user.ops', 'PUT', 'audit', 'preview update bucket lifecycle rule', ['查看差异'])
      ])
    ],
    workflows: workflows(['加载集群日志', '过滤 priority/channel/daemon', '查看审计和任务事件', '导出或跳转资源详情']),
    capabilities: ['Cluster log timeline 占位', 'Audit log 表格占位', 'Task/notification sidebar 占位']
  },
  alertList: {
    title: '告警列表',
    summary: '按 Alertmanager active alerts 展示分组告警、子告警、标签、注解和静默入口。',
    icon: <BellOutlined />,
    primaryAction: '刷新告警',
    reference: 'ceph/cluster/prometheus/active-alert-list + prometheus-alert service',
    metrics: metrics('当前告警', '7', 'critical 1 / warning 6', 64),
    sections: [
      monitorSection('active', '当前告警', '预留 Alertmanager group alerts、fingerprint、labels 和 annotations。', [
        monitorRow('CephHealthWarning', 'warning', 'cluster=prod', 'HEALTH_WARN for 10m；alert_count 3；summary from annotations', ['详情', '创建静默']),
        monitorRow('OSDNearFull', 'critical', 'osd=23', 'ceph_osd_stat_bytes_used / ceph_osd_stat_bytes > 0.85', ['OSD 详情', '静默']),
        monitorRow('RGWHighLatency', 'warning', 'service=rgw-a', 'p95 latency above 80ms for 15m', ['RGW 详情'])
      ]),
      monitorSection('resolved', '最近恢复', '预留 resolved alert 和 notification 变化记录。', [
        monitorRow('MDSBehindOnTrimming', 'resolved', 'fs=cephfs-prod', 'resolved 26m ago，was active for 1h 12m', ['查看历史'])
      ])
    ],
    workflows: workflows(['拉取 grouped alerts', '展开子告警与标签', '查看注解/链接', '创建 Silence 或跳转资源']),
    capabilities: ['Grouped alerts 展开占位', 'Alert labels/annotations 详情占位', '和 Silence 表单联动占位']
  },
  alertRules: {
    title: '告警规则',
    summary: '按 Prometheus rules 展示 alerting/recording rule group、表达式、for、health 和当前 alerts。',
    icon: <SettingOutlined />,
    primaryAction: '新建规则',
    reference: 'ceph/cluster/prometheus/rules-list + shared/models/prometheus-alerts',
    metrics: metrics('规则组', '18', '142 条启用规则', 79),
    sections: [
      monitorSection('alerting', 'Alerting Rules', '预留 Prometheus rule group、query、duration、severity 和 active alerts。', [
        monitorRow('CephHealthWarning', '启用', 'group ceph-health', 'expr ceph_health_status > 0；for 10m；severity warning', ['测试表达式', '编辑']),
        monitorRow('OSDNearFull', '启用', 'group ceph-osd', 'expr used_bytes / total_bytes > 0.85；for 5m', ['查看告警']),
        monitorRow('RGWHighLatency', '草稿', 'group rgw', 'expr histogram_quantile(0.95, rgw_req_latency_bucket) > 0.08', ['完善规则'])
      ]),
      monitorSection('recording', 'Recording Rules', '预留 recording rule 和指标复用。', [
        monitorRow('cluster:client_io_bytes:rate5m', '启用', 'group ceph-recording', 'sum(rate(ceph_client_io_bytes[5m])) by (cluster,type)', ['复制 PromQL']),
        monitorRow('osd:commit_latency:p95', '启用', 'group osd-recording', 'histogram_quantile(0.95, rate(ceph_osd_op_latency_bucket[5m]))', ['复制 PromQL'])
      ])
    ],
    workflows: workflows(['选择 alerting 或 recording', '编辑 PromQL 表达式', '设置 for/labels/annotations', '测试并启用规则']),
    capabilities: ['Rule group 表格占位', 'PromQL 测试面板占位', '规则与 active alerts 关联占位']
  },
  alertSilences: {
    title: '告警静默',
    summary: '按 Alertmanager Silence 展示匹配器、创建者、有效期、注释和匹配告警预览。',
    icon: <PauseCircleOutlined />,
    primaryAction: '新建静默',
    reference: 'ceph/cluster/prometheus/silence-list, silence-form, silence-matcher service',
    metrics: metrics('静默策略', '5', '2 条将在 24 小时内过期', 54),
    sections: [
      monitorSection('active', '生效中', '预留 matchers、startsAt、endsAt、createdBy 和 comment。', [
        monitorRow('maintenance-osd-23', 'active', 'osd=23,severity=critical', 'createdBy ops；ends 2026-07-22 02:00；matches 1 active alert', ['编辑', '过期']),
        monitorRow('rgw-dr-test', 'active', 'service=rgw-b', 'createdBy sre；ends 2026-07-21 18:00；matches 2 alerts', ['编辑', '过期'])
      ]),
      monitorSection('matcher-preview', '匹配器预览', '对应 prometheus-silence-matcher 对规则和当前告警的匹配提示。', [
        monitorRow('alertname=CephHealthWarning', '匹配', 'rules 1 / alerts 3', 'matcher matches 1 currently defined rule with 3 active alerts', ['查看规则']),
        monitorRow('instance=host-09', '无活动告警', 'rules 2 / alerts 0', 'matcher matches 2 rules with no active alerts', ['继续创建'])
      ])
    ],
    workflows: workflows(['填写 matcher', '预览匹配规则和活动告警', '设置开始/结束时间', '保存 Silence 并跟踪过期']),
    capabilities: ['Silence list/form 占位', 'Matcher 预览服务占位', 'Silence 与 active alerts 关联占位']
  }
}

function tableSection(key: string, title: string, note: string, rows: Row[]): Section {
  return { key, title, note, rows, columns: key.includes('policy') || key.includes('rules') || key.includes('config') ? policyColumns : resourceColumns }
}

function monitorSection(key: string, title: string, note: string, rows: Row[]): Section {
  return { key, title, note, rows, columns: monitorColumns }
}

function resource(name: string, statusText: string, scope: ReactNode, size: ReactNode, config: ReactNode, actions: string[]): Row {
  return {
    key: `${name}-${statusText}`,
    name,
    status: status(statusText),
    scope,
    size,
    config,
    actions: actionTags(actions)
  }
}

function policy(name: string, statusText: string, scope: ReactNode, config: ReactNode, actions: string[]): Row {
  return {
    key: `${name}-${statusText}`,
    name,
    status: status(statusText),
    scope,
    config,
    actions: actionTags(actions)
  }
}

function monitorRow(name: string, statusText: string, scope: ReactNode, config: ReactNode, actions: string[]): Row {
  return {
    key: `${name}-${statusText}`,
    name,
    status: status(statusText),
    scope,
    config,
    actions: actionTags(actions)
  }
}

function status(value: string) {
  const lower = value.toLowerCase()
  let tone: Tone = 'default'
  if (lower.includes('ok') || lower.includes('健康') || lower.includes('正常') || lower.includes('启用') || lower.includes('有效') || lower.includes('在线') || lower.includes('up') || lower.includes('完成')) {
    tone = 'success'
  } else if (lower.includes('warn') || lower.includes('待') || lower.includes('偏高') || lower.includes('关注') || lower.includes('草稿') || lower.includes('限速')) {
    tone = 'warning'
  } else if (lower.includes('critical') || lower.includes('error')) {
    tone = 'error'
  } else if (lower.includes('同步') || lower.includes('运行') || lower.includes('active') || lower.includes('post') || lower.includes('put')) {
    tone = 'processing'
  }
  return <Tag color={tone}>{value}</Tag>
}

function actionTags(actions: string[]) {
  return actions.map((action) => <Tag key={action}>{action}</Tag>)
}

function renderName(value: ReactNode) {
  return (
    <div className="placeholder-name-cell">
      <Text strong>{value}</Text>
      <Text type="secondary">功能占位</Text>
    </div>
  )
}

function metrics(firstLabel: string, firstValue: string, firstNote: string, percent: number): Metric[] {
  return [
    { label: firstLabel, value: firstValue, note: firstNote, percent },
    { label: '健康状态', value: percent > 70 ? 'HEALTH_OK' : 'HEALTH_WARN', note: '按参考模块预留状态摘要' },
    { label: '配置覆盖', value: `${Math.max(4, Math.round(percent / 8))} 项`, note: '表单与详情字段已预留' },
    { label: 'API 状态', value: '未接入', note: '当前页面不请求 Go API' }
  ]
}

function workflows(steps: string[]): Workflow[] {
  return [
    {
      title: '主流程',
      steps
    }
  ]
}

function fallbackDefinition(pageKey: PageKey): Definition {
  const page = findNavPage(pageKey)
  return {
    title: page?.label ?? '资源管理',
    summary: '该页面已预留前端功能区，等待进一步细化。',
    icon: <SettingOutlined />,
    primaryAction: '新建',
    reference: 'docs/references/ceph',
    metrics: metrics('资源', '0', '等待配置', 0),
    sections: [tableSection('resources', '资源', '暂无细化数据。', [])],
    workflows: workflows(['查看列表', '进入详情', '编辑配置', '提交操作']),
    capabilities: ['列表占位', '详情占位', '表单占位']
  }
}

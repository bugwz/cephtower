import {
  ApiOutlined,
  DatabaseOutlined,
  DeploymentUnitOutlined,
  HddOutlined,
  LineChartOutlined,
  ReloadOutlined,
  SafetyCertificateOutlined,
  ThunderboltOutlined
} from '@ant-design/icons'
import { Button, Card, Progress, Segmented, Space, Tag, Typography } from 'antd'
import type { CSSProperties, ReactNode } from 'react'
import { useMemo, useState } from 'react'
import { Page } from '../../components/Page'

const { Text } = Typography

const dashboardData = {
  health: 'HEALTH_WARN',
  healthText: '集群可用，存在 2 个容量与 PG 风险',
  version: 'ceph version 19.2.1 squid',
  mgr: 'mgr.a@ceph-mgr-01',
  monConnection: 'connected',
  updatedAt: '2026-07-21 13:40:12',
  capacity: {
    total: '1.28 PiB',
    used: '793 TiB',
    available: '487 TiB',
    usedPercent: 62,
    nearfullRatio: 85,
    fullRatio: 95
  },
  inventory: {
    hosts: 18,
    osds: 144,
    osdsUp: 142,
    osdsIn: 143,
    mons: 5,
    mgrs: 2,
    pools: 32,
    pgs: 4096
  },
  clientIo: {
    readIops: '18.6K',
    writeIops: '7.4K',
    readThroughput: '8.9 GiB/s',
    writeThroughput: '3.2 GiB/s',
    readLatency: '2.8 ms',
    writeLatency: '4.1 ms',
    recovery: '624 MiB/s'
  },
  objects: {
    total: '2.84B',
    degraded: '12.4K',
    misplaced: '48.1K',
    pgsPerOsd: 28.4
  },
  rbdMirroring: {
    daemons: 4,
    warnings: 1,
    errors: 0,
    replayLag: '18 s'
  }
}

const poolUsage = [
  { name: 'rbd', app: 'rbd', used: 73, throughput: '4.2 GiB/s', iops: '11.8K' },
  { name: 'cephfs_data', app: 'cephfs', used: 58, throughput: '2.1 GiB/s', iops: '6.4K' },
  { name: '.rgw.buckets.data', app: 'rgw', used: 44, throughput: '860 MiB/s', iops: '3.9K' },
  { name: '.mgr', app: 'mgr', used: 17, throughput: '24 MiB/s', iops: '220' }
]

const pgStates = [
  { label: 'active+clean', value: 3920, color: '#43bf8f' },
  { label: 'degraded', value: 72, color: '#f5a623' },
  { label: 'remapped', value: 64, color: '#5470f1' },
  { label: 'backfilling', value: 40, color: '#7b61ff' }
]

const serviceHealth = [
  { name: 'MON quorum', value: '5 / 5', status: 'ok' },
  { name: 'MGR active/standby', value: '1 / 1', status: 'ok' },
  { name: 'OSD up/in', value: '142 / 143', status: 'warn' },
  { name: 'MDS active', value: '2 / 2', status: 'ok' },
  { name: 'RGW gateway', value: '6 / 6', status: 'ok' },
  { name: 'RBD mirror', value: '1 warning', status: 'warn' }
]

const healthChecks = [
  { name: 'OSD_NEARFULL', summary: '2 OSDs are above nearfull ratio', severity: 'warn', count: 2 },
  { name: 'PG_DEGRADED', summary: '72 placement groups include degraded objects', severity: 'warn', count: 72 },
  { name: 'POOL_APP_NOT_ENABLED', summary: 'all production pools have applications enabled', severity: 'ok', count: 0 }
]

const ioTrends = {
  read: {
    label: '读吞吐',
    unit: 'GiB/s',
    values: [4.2, 4.9, 4.5, 5.7, 6.1, 5.5, 7.3, 6.6, 8.1, 7.8, 8.9, 8.5]
  },
  write: {
    label: '写吞吐',
    unit: 'GiB/s',
    values: [1.4, 1.9, 1.7, 2.3, 2.8, 2.4, 3.1, 2.7, 3.4, 3.0, 3.2, 3.1]
  },
  iops: {
    label: 'IOPS',
    unit: 'K',
    values: [8.2, 9.5, 8.9, 11.6, 13.2, 12.4, 15.1, 14.2, 17.8, 16.9, 18.6, 17.5]
  },
  latency: {
    label: '延迟',
    unit: 'ms',
    values: [4.9, 4.4, 4.7, 3.9, 3.5, 3.8, 3.2, 3.6, 2.9, 3.1, 2.8, 3.0]
  }
} as const

type IoTrendKey = keyof typeof ioTrends

export function OverviewPage() {
  const capacityRingStyle = { '--used-percent': `${dashboardData.capacity.usedPercent}%` } as CSSProperties
  const [activeTrend, setActiveTrend] = useState<IoTrendKey>('read')
  const activeTrendData = ioTrends[activeTrend]

  return (
    <Page title="总览">
      <Card
        className="page-surface-card overview-surface-card"
        title="集群总览"
        extra={
          <>
          <Tag className="dashboard-version">{dashboardData.version}</Tag>
          <Button icon={<ReloadOutlined />}>刷新</Button>
          </>
        }
      >
        <div className="overview-dashboard">

        <section className="dashboard-hero">
          <Card className="dashboard-health-card">
            <div className="dashboard-health-head">
              <div>
                <Text type="secondary">集群健康</Text>
                <h2>{dashboardData.health}</h2>
              </div>
              <span className="dashboard-health-pulse" />
            </div>
            <Text>{dashboardData.healthText}</Text>
            <div className="dashboard-health-meta">
              <span>Mgr {dashboardData.mgr}</span>
              <span>MON {dashboardData.monConnection}</span>
              <span>{dashboardData.updatedAt}</span>
            </div>
          </Card>

          <MetricCard
            icon={<DatabaseOutlined />}
            label="原始容量"
            value={dashboardData.capacity.total}
            detail={`已用 ${dashboardData.capacity.used} / 可用 ${dashboardData.capacity.available}`}
            tone="green"
          />
          <MetricCard
            icon={<HddOutlined />}
            label="OSD"
            value={`${dashboardData.inventory.osdsUp}/${dashboardData.inventory.osds}`}
            detail={`in ${dashboardData.inventory.osdsIn} · hosts ${dashboardData.inventory.hosts}`}
            tone="blue"
          />
          <MetricCard
            icon={<DeploymentUnitOutlined />}
            label="Pool / PG"
            value={`${dashboardData.inventory.pools} / ${dashboardData.inventory.pgs}`}
            detail={`PG per OSD ${dashboardData.objects.pgsPerOsd}`}
            tone="purple"
          />
          <MetricCard
            icon={<LineChartOutlined />}
            label="客户端吞吐"
            value={dashboardData.clientIo.readThroughput}
            detail={`写入 ${dashboardData.clientIo.writeThroughput}`}
            tone="cyan"
          />
          <MetricCard
            icon={<ThunderboltOutlined />}
            label="IOPS"
            value={dashboardData.clientIo.readIops}
            detail={`write ${dashboardData.clientIo.writeIops}`}
            tone="amber"
          />
          <MetricCard
            icon={<ApiOutlined />}
            label="延迟"
            value={dashboardData.clientIo.readLatency}
            detail={`commit ${dashboardData.clientIo.writeLatency}`}
            tone="rose"
          />
          <MetricCard
            icon={<SafetyCertificateOutlined />}
            label="RBD 镜像"
            value={`${dashboardData.rbdMirroring.daemons} daemons`}
            detail={`warn ${dashboardData.rbdMirroring.warnings} · lag ${dashboardData.rbdMirroring.replayLag}`}
            tone="green"
          />
        </section>

        <section className="dashboard-main-grid">
          <Card className="dashboard-capacity-card" title="容量水位">
            <div className="capacity-ring" style={capacityRingStyle}>
              <div>
                <strong>{dashboardData.capacity.usedPercent}%</strong>
                <span>used</span>
              </div>
            </div>
            <div className="capacity-details">
              <Progress percent={dashboardData.capacity.usedPercent} showInfo={false} strokeColor="#43bf8f" />
              <div className="capacity-thresholds">
                <span>nearfull {dashboardData.capacity.nearfullRatio}%</span>
                <span>full {dashboardData.capacity.fullRatio}%</span>
              </div>
              <dl>
                <div>
                  <dt>对象</dt>
                  <dd>{dashboardData.objects.total}</dd>
                </div>
                <div>
                  <dt>降级对象</dt>
                  <dd>{dashboardData.objects.degraded}</dd>
                </div>
                <div>
                  <dt>错位对象</dt>
                  <dd>{dashboardData.objects.misplaced}</dd>
                </div>
              </dl>
            </div>
          </Card>

          <Card
            className="dashboard-io-card"
            title="客户端 IO 趋势"
            extra={
              <Segmented
                size="small"
                value={activeTrend}
                options={(Object.keys(ioTrends) as IoTrendKey[]).map((key) => ({
                  label: ioTrends[key].label,
                  value: key
                }))}
                onChange={(value) => setActiveTrend(value as IoTrendKey)}
              />
            }
          >
            <Sparkline values={activeTrendData.values} label={activeTrendData.label} unit={activeTrendData.unit} />
            <div className="io-summary">
              <div>
                <Text type="secondary">Read</Text>
                <strong>{dashboardData.clientIo.readThroughput}</strong>
              </div>
              <div>
                <Text type="secondary">Write</Text>
                <strong>{dashboardData.clientIo.writeThroughput}</strong>
              </div>
              <div>
                <Text type="secondary">Recovery</Text>
                <strong>{dashboardData.clientIo.recovery}</strong>
              </div>
            </div>
          </Card>

          <Card title="PG 状态分布">
            <div className="pg-distribution">
              {pgStates.map((state) => (
                <div className="pg-state-row" key={state.label}>
                  <span>
                    <i style={{ backgroundColor: state.color }} />
                    {state.label}
                  </span>
                  <strong>{state.value}</strong>
                  <Progress
                    percent={Math.round((state.value / dashboardData.inventory.pgs) * 100)}
                    showInfo={false}
                    strokeColor={state.color}
                  />
                </div>
              ))}
            </div>
          </Card>

          <Card title="服务健康">
            <div className="service-health-grid">
              {serviceHealth.map((service) => (
                <div className={`service-health-item service-health-${service.status}`} key={service.name}>
                  <Text type="secondary">{service.name}</Text>
                  <strong>{service.value}</strong>
                </div>
              ))}
            </div>
          </Card>
        </section>

        <section className="dashboard-lower-grid">
          <Card title="Top Pool 利用率">
            <div className="pool-table">
              {poolUsage.map((pool) => (
                <div className="pool-row" key={pool.name}>
                  <div>
                    <Text strong>{pool.name}</Text>
                    <Text type="secondary">{pool.app}</Text>
                  </div>
                  <Progress percent={pool.used} showInfo={false} strokeColor="#168766" />
                  <Text>{pool.throughput}</Text>
                  <Text>{pool.iops}</Text>
                </div>
              ))}
            </div>
          </Card>

          <Card title="健康检查">
            <div className="health-check-list dashboard-check-list">
              {healthChecks.map((check) => (
                <div className="health-check-row" key={check.name}>
                  <span className={check.severity === 'ok' ? 'check-icon ok' : 'check-icon warning'}>
                    {check.severity === 'ok' ? '✓' : '!'}
                  </span>
                  <div>
                    <Text strong>{check.name}</Text>
                    <Text type="secondary">{check.summary}</Text>
                  </div>
                  <Text type="secondary">{check.count || 'OK'}</Text>
                </div>
              ))}
            </div>
          </Card>

          <Card title="近期任务">
            <div className="task-list">
              <TaskRow name="pool/create" detail="创建 replicated pool cephfs_metadata" progress={100} status="done" />
              <TaskRow name="osd/reweight" detail="osd.37 reweight 0.92" progress={100} status="done" />
              <TaskRow name="rbd/snap/create" detail="vm-images/base-ubuntu-24.04" progress={68} status="running" />
            </div>
          </Card>
        </section>
        </div>
      </Card>
    </Page>
  )
}

function MetricCard({ icon, label, value, detail, tone }: { icon: ReactNode; label: string; value: string; detail: string; tone: string }) {
  return (
    <Card className="dashboard-metric-card">
      <div className={`dashboard-metric-icon metric-tone-${tone}`}>{icon}</div>
      <div>
        <Text type="secondary">{label}</Text>
        <strong>{value}</strong>
        <span>{detail}</span>
      </div>
    </Card>
  )
}

function Sparkline({ values, label, unit }: { values: readonly number[]; label: string; unit: string }) {
  const [activeIndex, setActiveIndex] = useState(values.length - 1)
  const width = 520
  const height = 180
  const max = Math.max(...values)
  const min = Math.min(...values)
  const range = max - min || 1
  const chartPoints = useMemo(
    () =>
      values.map((value, index) => {
        const x = (index / (values.length - 1)) * width
        const y = height - ((value - min) / range) * (height - 24) - 12
        return { value, x, y }
      }),
    [values, min, range]
  )
  const points = chartPoints
    .map((value, index) => {
      return `${value.x},${value.y}`
    })
    .join(' ')
  const areaPoints = `0,${height} ${points} ${width},${height}`
  const activePoint = chartPoints[Math.min(activeIndex, chartPoints.length - 1)]

  return (
    <div className="dashboard-sparkline-wrap">
      <svg className="dashboard-sparkline" viewBox={`0 0 ${width} ${height}`} role="img" aria-label={`${label}趋势图`}>
        <polygon points={areaPoints} />
        <polyline points={points} />
        <line className="dashboard-sparkline-cursor" x1={activePoint.x} x2={activePoint.x} y1="12" y2={height - 8} />
        {chartPoints.map((point, index) => (
          <circle
            key={`${point.value}-${index}`}
            className={index === activeIndex ? 'active' : undefined}
            cx={point.x}
            cy={point.y}
            r={index === activeIndex ? 6 : 4}
            tabIndex={0}
            role="button"
            aria-label={`${label} 第 ${index + 1} 个采样 ${formatTrendValue(point.value, unit)}`}
            onFocus={() => setActiveIndex(index)}
            onMouseEnter={() => setActiveIndex(index)}
          />
        ))}
      </svg>
      <div className="dashboard-sparkline-readout" style={{ left: `${(activePoint.x / width) * 100}%` }}>
        <span>{label}</span>
        <strong>{formatTrendValue(activePoint.value, unit)}</strong>
      </div>
    </div>
  )
}

function formatTrendValue(value: number, unit: string) {
  return `${value.toFixed(value >= 10 ? 1 : 2).replace(/\.0$/, '')} ${unit}`
}

function TaskRow({ name, detail, progress, status }: { name: string; detail: string; progress: number; status: 'done' | 'running' }) {
  return (
    <div className="task-row">
      <div>
        <Space size={8}>
          <Text strong>{name}</Text>
          <Tag color={status === 'done' ? 'success' : 'processing'}>{status === 'done' ? '完成' : '执行中'}</Tag>
        </Space>
        <Text type="secondary">{detail}</Text>
      </div>
      <Progress percent={progress} size="small" />
    </div>
  )
}

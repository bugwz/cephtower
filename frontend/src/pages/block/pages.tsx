import { ResourceListPage, type ResourceListPageDefinition } from '../ResourceListPage'
import { ExternalListPage, type ExternalListPageDefinition } from '../ExternalListPage'
import { PoolPage } from './PoolPage'

export function BlockPoolsPage() {
  return <PoolPage />
}

export function RbdImagesPage() {
  return <ResourceListPage definition={resourceDefinitions.rbdImages} />
}

export function RbdSnapshotsPage() {
  return <ResourceListPage definition={resourceDefinitions.rbdSnapshots} />
}

export function RbdNamespacesPage() {
  return <ResourceListPage definition={resourceDefinitions.rbdNamespaces} />
}

export function RbdTrashPage() {
  return <ResourceListPage definition={resourceDefinitions.rbdTrash} />
}

export function RbdGroupsPage() {
  return <ResourceListPage definition={resourceDefinitions.rbdGroups} />
}

export function ImageMirroringPage() {
  return <ResourceListPage definition={resourceDefinitions.imageMirroring} />
}

export function IscsiPage() {
  return <ExternalListPage definition={externalDefinitions.iscsi} />
}

export function NvmeGatewayPage() {
  return <ExternalListPage definition={externalDefinitions.nvmeGateway} />
}

export function NvmeTcpPage() {
  return <ExternalListPage definition={externalDefinitions.nvmeTcp} />
}

export function NvmeNamespacesPage() {
  return <ExternalListPage definition={externalDefinitions.nvmeNamespaces} />
}

export function NvmeListenersPage() {
  return <ExternalListPage definition={externalDefinitions.nvmeListeners} />
}

export function NvmeHostsPage() {
  return <ExternalListPage definition={externalDefinitions.nvmeHosts} />
}

export function NvmeConnectionsPage() {
  return <ExternalListPage definition={externalDefinitions.nvmeConnections} />
}

const resourceDefinitions: Record<'blockPools' | 'rbdImages' | 'rbdSnapshots' | 'rbdNamespaces' | 'rbdTrash' | 'rbdGroups' | 'imageMirroring', ResourceListPageDefinition> = {
  blockPools: {
    title: '存储池',
    path: '/pools',
    rowKeyCandidates: ['natural_key', 'pool_name', 'name'],
    columns: [
      { key: 'name', title: '名称' },
      { key: 'status', title: '状态' },
      { key: 'type', title: '类型' },
      { key: 'size', title: '副本/大小' },
      { key: 'min_size', title: '最小副本' },
      { key: 'pg_num', title: 'PG' },
      { key: 'application_metadata', title: '应用' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  rbdImages: {
    title: 'RBD 镜像',
    path: '/rbd/images',
    requiredCapabilities: ['rbd'],
    rowKeyCandidates: ['natural_key', 'image_spec', 'name'],
    createAction: {
      title: '新建 RBD 镜像',
      buttonLabel: '新建镜像',
      path: '/rbd/image',
      method: 'POST',
      successMessage: 'RBD 镜像创建执行成功',
      fields: [
        { name: 'image_spec', label: '镜像规格', required: true, placeholder: 'pool/image 或 pool/namespace/image' },
        { name: 'size', label: '容量 bytes', type: 'number', required: true, min: 1 },
        { name:'data_pool',label:'数据池（可选）' },
        { name:'object_size',label:'对象大小 bytes（可选）',type:'select',options:Array.from({length:14},(_,index)=>{const size=2**(index+12);return {label:String(size),value:size}}) },
        { name:'stripe_unit',label:'条带单元 bytes（可选）',type:'number',min:1 },
        { name:'stripe_count',label:'条带数量（可选）',type:'number',min:1 }
      ],
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        image_spec: String(values.image_spec ?? ''),
        size: Number(values.size),
        ...(values.data_pool ? {data_pool:String(values.data_pool)} : {}),
        ...(values.object_size != null ? {object_size:Number(values.object_size)} : {}),
        ...(values.stripe_unit != null ? {stripe_unit:Number(values.stripe_unit)} : {}),
        ...(values.stripe_count != null ? {stripe_count:Number(values.stripe_count)} : {})
      })
    },
    updateAction: {
      title: '更新 RBD 镜像',
      path: '/rbd/image',
      method: 'PATCH',
      successMessage: 'RBD 镜像更新执行成功',
      confirmation:(values,row)=>values.allow_shrink ? `允许将 ${String(row?.image_path ?? imageSpec(row))} 调整为 ${String(values.size)} bytes？缩容会丢失被截断区域的数据，请先缩小文件系统并停止相关写入。` : undefined,
      fields: [
        { name: 'size', label: '新容量 bytes', type: 'number', required: true, min: 1 },
        { name:'allow_shrink',label:'允许缩容（会丢失被截断部分的数据）',type:'boolean' }
      ],
      buildBody: (values, clusterId, row) => ({
        cluster_id: clusterId,
        image_spec: imageSpec(row),
        size: Number(values.size),
        allow_shrink: Boolean(values.allow_shrink)
      })
    },
    extraActions: [{
      title: 'RBD 镜像操作', buttonLabel: '更多操作', path: '/rbd/image/action', method: 'POST',
      successMessage: 'RBD 镜像操作执行成功',
      confirmation: (values,row) => values.action === 'move-to-trash' ? `将 ${imageSpec(row)} 移入回收站，镜像将从当前列表移除。` : undefined,
      fields: [
        { name:'action',label:'操作',type:'select',required:true,options:[
          {label:'扁平化克隆镜像',value:'flatten'}, {label:'回收零填充空间',value:'sparsify'},
          {label:'重命名镜像',value:'rename'}, {label:'复制镜像',value:'copy'}, {label:'深度复制（包含快照）',value:'deep-copy'}, {label:'移入回收站',value:'move-to-trash'}
        ] },
        {name:'destination',required:true,visibleWhen:(values)=>['copy','deep-copy','rename'].includes(String(values.action)),label:'目标（复制填完整路径，重命名只填新名称）',placeholder:'pool/namespace/image'},
        {name:'expires_at',visibleWhen:(values)=>values.action==='move-to-trash',label:'回收站到期时间（移入时可选，含时区）',placeholder:'2026-12-31T23:59:59+08:00'}
      ],
      initialValues:{action:'flatten'},
      buildBody:(values,clusterId,row) => ({cluster_id:clusterId,image_spec:imageSpec(row),action:String(values.action),...(['copy','deep-copy','rename'].includes(String(values.action)) ? {destination:String(values.destination ?? '')} : {}),...(values.action==='move-to-trash' && values.expires_at ? {expires_at:String(values.expires_at)} : {})})
    }, {
      title:'创建镜像快照',buttonLabel:'创建快照',path:'/rbd/image/snapshot',method:'POST',successMessage:'镜像快照已创建',
      fields:[{name:'name',label:'快照名称',required:true}],
      buildBody:(values,clusterId,row)=>({cluster_id:clusterId,image_spec:imageSpec(row),name:String(values.name ?? '')})
    }, {
      title:'清理未保护快照',buttonLabel:'清理快照',path:'/rbd/image/action',method:'POST',successMessage:'快照清理请求已执行',
      fields:[],
      confirmation:(_values,row)=>`删除镜像 ${String(row?.image_path ?? imageSpec(row))} 的所有未保护快照？删除后无法使用这些快照恢复数据。`,
      buildBody:(_values,clusterId,row)=>({cluster_id:clusterId,image_spec:imageSpec(row),action:'snapshot-purge'})
    }, {
      title:'镜像配置',buttonLabel:'配置',path:'/rbd/image/action',method:'POST',successMessage:'镜像配置已更新',
      fields:[{name:'action',label:'操作',type:'select',required:true,options:[{label:'设置覆盖值',value:'config-set'},{label:'移除覆盖值，恢复继承',value:'config-remove'}]},
        {name:'config_name',label:'配置名称',required:true,placeholder:'rbd_qos_iops_limit'},
        {name:'config_value',required:true,visibleWhen:(values)=>values.action==='config-set',label:'配置值（设置时必填）'}],
      buildBody:(values,clusterId,row)=>({cluster_id:clusterId,image_spec:imageSpec(row),action:String(values.action),config_name:String(values.config_name),...(values.action==='config-set'?{config_value:String(values.config_value ?? '')}:{})})
    }, {
      title:'修改镜像特性',buttonLabel:'特性',path:'/rbd/image/action',method:'POST',successMessage:'镜像特性已更新',
      fields:[{name:'action',label:'操作',type:'select',required:true,options:[{label:'启用',value:'feature-enable'},{label:'停用',value:'feature-disable'}]},
        {name:'feature',label:'特性（deep-flatten 仅可停用）',type:'select',required:true,options:['exclusive-lock','object-map','fast-diff','journaling','deep-flatten'].map(value=>({label:value,value}))}],
      buildBody:(values,clusterId,row)=>({cluster_id:clusterId,image_spec:imageSpec(row),action:String(values.action),feature:String(values.feature)})
    }, {
      title:'镜像同步操作',buttonLabel:'同步操作',path:'/rbd/image/action',method:'POST',
      successMessage:'镜像同步请求已执行',
      confirmation:(values,row) => {
        const effects:Record<string,string> = {
          'mirror-disable':'停用同步，镜像将不再复制更新',
          'mirror-demote':'降级为非主镜像，请先停止对该镜像的写入',
          'mirror-resync':'重新同步，本地数据将被主镜像的数据覆盖'
        }
        const effect = effects[String(values.action)]
        return effect ? `确认对 ${String(row?.image_path ?? imageSpec(row))} ${effect}？` : undefined
      },
      fields:[{name:'action',label:'操作',type:'select',required:true,options:[
        {label:'启用日志同步',value:'mirror-enable-journal'},
        {label:'启用快照同步',value:'mirror-enable-snapshot'},
        {label:'停用同步',value:'mirror-disable'},
        {label:'提升为主镜像',value:'mirror-promote'},
        {label:'降级为非主镜像',value:'mirror-demote'},
        {label:'从主镜像重新同步',value:'mirror-resync'},
        {label:'创建同步快照',value:'mirror-snapshot'}
      ]}],
      buildBody:(values,clusterId,row) => ({cluster_id:clusterId,image_spec:imageSpec(row),action:String(values.action)})
    }],
    deleteAction: {
      title: '删除 RBD 镜像',
      path: '/rbd/image',
      action: 'rbd_image.delete',
      resourceKind: 'rbd_image',
      successMessage: 'RBD 镜像删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, image_spec: imageSpec(row) }),
      resourceKey: (row) => `rbd/image/${imageSpec(row)}`
    },
    columns: [
      { key: 'name', title: '名称' },
      { key: 'image_path', title: '镜像' },
      { key: 'pool', title: 'Pool' },
      { key: 'namespace', title: '命名空间' },
      { key: 'size_bytes', title: '容量（bytes）' },
      { key: 'features', title: '特性' },
      { key: 'parent', title: '父镜像' },
      { key: 'details', title: '镜像详情' },
      { key: 'runtime_status', title: '连接客户端 / 迁移 / 缓存状态' },
      { key: 'configuration', title: '生效配置及来源' },
      { key: 'status', title: '状态' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  rbdSnapshots: {
    title: 'RBD 快照',
    path: '/rbd/image/snapshots',
    requiredCapabilities: ['rbd'],
    rowKeyCandidates: ['natural_key', 'snapshot_spec', 'name'],
    createAction: {
      title: '新建 RBD 快照',
      buttonLabel: '新建快照',
      path: '/rbd/image/snapshot',
      method: 'POST',
      successMessage: 'RBD 快照创建执行成功',
      fields: [
        { name: 'image_spec', label: '镜像规格', required: true, placeholder: 'pool/image 或 pool/namespace/image' },
        { name: 'name', label: '快照名称', required: true }
      ],
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        image_spec: encodeImageSpec(String(values.image_spec ?? '')),
        name: String(values.name ?? '')
      })
    },
    updateAction: {
      title: '更新 RBD 快照',
      path: '/rbd/image/snapshot/action',
      method: 'POST',
      successMessage: 'RBD 快照更新执行成功',
      confirmation: (values, row) => values.action === 'rollback' ? `将镜像 ${imageSpec(row)} 回滚到快照 ${snapshotName(row)}。快照之后的镜像修改将丢失，请确认相关应用已停止写入。` : undefined,
      fields: [
        {
          name: 'action',
          label: '操作',
          type: 'select',
          required: true,
          options: [
            { label: '保护', value: 'protect' },
            { label: '取消保护', value: 'unprotect' },
            { label: '克隆为新镜像', value: 'clone' },
            { label: '回滚镜像到此快照', value: 'rollback' }
          ]
        },
        { name: 'destination', required:true,visibleWhen:(values)=>values.action==='clone', label: '克隆目标（选择克隆时必填）', placeholder: 'pool/image 或 pool/namespace/image' }
      ],
      initialValues: { action: 'protect' },
      buildBody: (values, clusterId, row) => ({
        cluster_id: clusterId,
        image_spec: imageSpec(row),
        snap: snapshotName(row),
        action: String(values.action ?? 'protect'),
        ...(values.action === 'clone' ? { destination: String(values.destination ?? '') } : {})
      })
    },
    extraActions:[{
      title:'重命名快照',buttonLabel:'重命名',path:'/rbd/image/snapshot',method:'PATCH',successMessage:'快照已重命名',
      fields:[{name:'name',label:'新快照名称',required:true}],
      buildBody:(values,clusterId,row)=>({cluster_id:clusterId,image_spec:imageSpec(row),snap:snapshotName(row),name:String(values.name ?? '')})
    }],
    deleteAction: {
      title: '删除 RBD 快照',
      path: '/rbd/image/snapshot',
      action: 'rbd_snapshot.delete',
      resourceKind: 'rbd_snapshot',
      successMessage: 'RBD 快照删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, image_spec: imageSpec(row), snap: snapshotName(row) }),
      resourceKey: (row) => `rbd/image/${imageSpec(row)}/snapshot/${snapshotName(row)}`
    },
    columns: [
      { key: 'name', title: '快照' },
      { key: 'image_path', title: '镜像' },
      { key: 'pool_name', title: 'Pool' },
      { key: 'namespace', title: '命名空间' },
      { key: 'size', title: '容量' },
      { key: 'protected', title: '保护' },
      { key: 'timestamp', title: '时间' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  rbdNamespaces: {
    title: 'RBD 命名空间',
    path: '/rbd/namespaces',
    requiredCapabilities: ['rbd'],
    rowKeyCandidates: ['natural_key', 'namespace', 'name'],
    createAction: {
      title: '新建 RBD 命名空间',
      buttonLabel: '新建命名空间',
      path: '/rbd/namespace',
      method: 'POST',
      successMessage: 'RBD 命名空间创建执行成功',
      fields: [
        { name: 'pool', label: 'Pool', required: true },
        { name: 'name', label: '命名空间', required: true }
      ],
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        pool: String(values.pool ?? ''),
        name: String(values.name ?? '')
      })
    },
    deleteAction: {
      title: '删除 RBD 命名空间',
      path: '/rbd/namespace',
      action: 'rbd_namespace.delete',
      resourceKind: 'rbd_namespace',
      successMessage: 'RBD 命名空间删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, pool: namespacePool(row), namespace: namespaceName(row) }),
      resourceKey: (row) => `rbd/namespace/${namespacePool(row)}/${namespaceName(row)}`
    },
    columns: [
      { key: 'pool', title: 'Pool' },
      { key: 'namespace', title: '命名空间' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  rbdTrash: {
    title: 'RBD 回收站',
    path: '/rbd/trash',
    requiredCapabilities: ['rbd'],
    rowKeyCandidates: ['natural_key', 'image_id', 'id', 'name'],
    createAction: {
      title: '恢复 RBD 回收站镜像',
      buttonLabel: '恢复镜像',
      path: '/rbd/trash/restore',
      method: 'POST',
      successMessage: 'RBD 回收站恢复执行成功',
      fields: [
        { name: 'pool', label: 'Pool / 命名空间', required: true },
        { name: 'image_id', label: 'Image ID', required: true },
        { name: 'name', label: '恢复后名称', required: true }
      ],
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        pool: String(values.pool ?? ''),
        image_id: String(values.image_id ?? ''),
        name: String(values.name ?? '')
      })
    },
    extraActions: [{
      title:'清理已过期回收站镜像',buttonLabel:'清理本池/命名空间',path:'/rbd/trash/purge',method:'POST',successMessage:'已过期镜像清理请求已执行',
      fields:[{name:'expired_before',label:'到期截止时间（可选，含时区）',placeholder:'2026-12-31T23:59:59+08:00'}],
      confirmation:(values,row)=>`清理 ${trashPool(row)} 中 ${values.expired_before ? String(values.expired_before) : '当前时间'} 之前到期的回收站镜像？此操作永久删除数据。`,
      buildBody:(values,clusterId,row)=>({cluster_id:clusterId,pool:trashPool(row),...(values.expired_before?{expired_before:String(values.expired_before)}:{})})
    },{
      title:'恢复回收站镜像',buttonLabel:'恢复',path:'/rbd/trash/restore',method:'POST',successMessage:'镜像已恢复',
      fields:[{name:'name',label:'恢复后名称',required:true}],
      initialValues:(row) => ({name:String(row?.name ?? '')}),
      buildBody:(values,clusterId,row) => ({cluster_id:clusterId,pool:trashPool(row),image_id:trashImageID(row),name:String(values.name ?? '')})
    }],
    deleteAction: {
      title: '删除回收站镜像',
      path: '/rbd/trash',
      action: 'rbd_trash.delete',
      resourceKind: 'rbd_trash',
      successMessage: 'RBD 回收站删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, image_id: encodePair(trashPool(row), trashImageID(row)) }),
      resourceKey: (row) => `rbd/trash/${encodePair(trashPool(row), trashImageID(row))}`
    },
    columns: [
      { key: 'pool', title: 'Pool' },
      { key: 'namespace', title: '命名空间' },
      { key: 'pool_name', title: 'Pool 名称' },
      { key: 'image_id', title: 'Image ID' },
      { key: 'id', title: 'ID' },
      { key: 'name', title: '名称' },
      { key: 'deleted_at', title: '删除时间' },
      { key: 'status', title: '过期状态' },
      { key: 'source', title: '来源' },
      { key: 'parent', title: '父镜像' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  rbdGroups: {
    title: 'RBD 镜像组',
    path: '/rbd/groups',
    requiredCapabilities: ['rbd'],
    rowKeyCandidates: ['natural_key', 'name'],
    createAction: {
      title: '新建 RBD 镜像组',
      buttonLabel: '新建镜像组',
      path: '/rbd/group',
      method: 'POST',
      successMessage: 'RBD 镜像组创建执行成功',
      fields: [
        { name: 'pool', label: 'Pool / 命名空间', required: true },
        { name: 'name', label: '镜像组名称', required: true }
      ],
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        pool: String(values.pool ?? ''),
        name: String(values.name ?? '')
      })
    },
    extraActions:[{
      title:'重命名或删除镜像组',buttonLabel:'组管理',path:'/rbd/group/action',method:'POST',successMessage:'镜像组操作已完成',
      fields:[{name:'action',label:'操作',type:'select',required:true,options:[{label:'重命名',value:'rename'},{label:'删除组',value:'remove'}]},
        {name:'name',label:'新组名',required:true,visibleWhen:(values)=>values.action==='rename'}],
      confirmation:(values,row)=>values.action==='remove'?`删除镜像组 ${String(row?.group_spec ?? row?.natural_key)}？`:undefined,
      buildBody:(values,clusterId,row)=>({cluster_id:clusterId,group_spec:String(row?.group_spec ?? row?.natural_key),action:String(values.action),...(values.action==='rename'?{name:String(values.name ?? '')}:{})})
    },{
      title:'镜像组成员',buttonLabel:'管理成员',path:'/rbd/group/member',method:'POST',successMessage:'镜像组成员已更新',
      fields:[{name:'action',label:'操作',type:'select',required:true,options:[{label:'添加镜像',value:'add'},{label:'移出镜像',value:'remove'}]},{name:'image',label:'镜像完整路径',required:true,placeholder:'pool/namespace/image'}],
      initialValues:{action:'add'},
      buildBody:(values,clusterId,row) => ({cluster_id:clusterId,group_spec:String(row?.group_spec ?? row?.natural_key),image:String(values.image ?? ''),action:String(values.action)})
    },{
      title:'镜像组快照操作',buttonLabel:'组快照',path:'/rbd/group/snapshot',method:'POST',successMessage:'镜像组快照操作已完成',
      fields:[{name:'action',label:'操作',type:'select',required:true,options:[{label:'创建',value:'create'},{label:'删除',value:'remove'},{label:'回滚',value:'rollback'},{label:'重命名',value:'rename'}]},{name:'name',label:'快照名称',required:true},{name:'new_name',label:'新快照名',required:true,visibleWhen:(values)=>values.action==='rename'}],
      initialValues:{action:'create'},
      confirmation:(values,row) => values.action === 'rollback' ? `将组 ${String(row?.group_spec)} 回滚到 ${String(values.name)}，组内镜像在此快照后的修改将丢失。请确认相关应用已停止写入。` : values.action === 'remove' ? `删除组 ${String(row?.group_spec)} 的快照 ${String(values.name)}？` : undefined,
      buildBody:(values,clusterId,row) => ({cluster_id:clusterId,group_spec:String(row?.group_spec ?? row?.natural_key),name:String(values.name ?? ''),action:String(values.action),...(values.action==='rename'?{new_name:String(values.new_name ?? '')}:{})})
    }],
    columns: [
      { key: 'pool', title: 'Pool' },
      { key: 'namespace', title: '命名空间' },
      { key: 'group_id', title: '组 ID' },
      { key: 'name', title: '名称' },
      { key: 'images', title: '镜像' },
      { key: 'snapshots', title: '快照' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  imageMirroring: {
    title: 'RBD 镜像同步',
    path: '/rbd/mirroring',
    requiredCapabilities: ['rbd'],
    rowKeyCandidates: ['natural_key', 'name'],
    createAction:{
      title:'设置池同步模式',buttonLabel:'设置同步模式',path:'/rbd/mirroring',method:'PATCH',successMessage:'池同步模式已更新',
      confirmation:(values) => values.mode === 'disabled' ? `停用池 ${String(values.pool)} 的 RBD 同步？` : undefined,
      fields:[{name:'pool',label:'Pool',required:true},{name:'mode',label:'同步模式',type:'select',required:true,options:[{label:'禁用',value:'disabled'},{label:'按镜像启用',value:'image'},{label:'全池启用',value:'pool'}]}],
      buildBody:(values,clusterId) => ({cluster_id:clusterId,pool:String(values.pool),mode:String(values.mode)})
    },
    extraActions:[{
      title:'添加远端站点',buttonLabel:'添加 Peer',path:'/rbd/mirroring/peer',method:'POST',successMessage:'远端站点已添加',
      fields:[{name:'remote_cluster',label:'远端集群名称（需已有连接配置）',required:true},
        {name:'remote_client',label:'远端客户端',required:true,placeholder:'client.rbd-mirror'},
        {name:'direction',label:'同步方向',type:'select',required:true,options:[{label:'仅接收',value:'rx-only'},{label:'双向',value:'rx-tx'}]}],
      initialValues:{direction:'rx-tx'},
      buildBody:(values,clusterId,row)=>({cluster_id:clusterId,pool:String(row?.pool ?? row?.name),action:'add',remote_cluster:String(values.remote_cluster),remote_client:String(values.remote_client),direction:String(values.direction)})
    },{
      title:'编辑远端站点',buttonLabel:'编辑 Peer',path:'/rbd/mirroring/peer',method:'POST',successMessage:'远端站点已更新',
      fields:[{name:'uuid',label:'Peer UUID（见 Peers 列）',required:true},
        {name:'field',label:'配置项',type:'select',required:true,options:[{label:'站点名称',value:'site-name'},{label:'客户端名称',value:'client'},{label:'远端 Monitor 地址',value:'mon-host'},{label:'同步方向',value:'direction'}]},
        {name:'value',label:'新值（方向：rx-only / tx-only / rx-tx）',required:true}],
      buildBody:(values,clusterId,row)=>({cluster_id:clusterId,pool:String(row?.pool ?? row?.name),action:'set',uuid:String(values.uuid),field:String(values.field),value:String(values.value)})
    },{
      title:'删除远端站点',buttonLabel:'删除 Peer',path:'/rbd/mirroring/peer',method:'POST',successMessage:'远端站点已删除',
      fields:[{name:'uuid',label:'Peer UUID（见 Peers 列）',required:true}],
      confirmation:(values,row)=>`删除池 ${String(row?.pool ?? row?.name)} 的远端站点 ${String(values.uuid)}，停止与该站点同步？`,
      buildBody:(values,clusterId,row)=>({cluster_id:clusterId,pool:String(row?.pool ?? row?.name),action:'remove',uuid:String(values.uuid)})
    }],
    columns: [
      { key: 'name', title: '名称' },
      { key: 'summary', title: '同步健康与镜像状态' },
      { key: 'mode', title: '模式' },
      { key: 'site_name', title: '本站点名称' },
      { key: 'mirror_uuid', title: '同步 UUID' },
      { key: 'remote_namespace', title: '远端命名空间' },
      { key: 'peers', title: 'Peers' },
      { key: 'daemons', title: '同步守护进程' },
      { key: 'images', title: '镜像同步详情' },
      { key: 'resource_version', title: '版本' }
    ]
  }
}

function imageSpec(row?: Record<string, unknown>) {
  return String(row?.image_spec ?? row?.natural_key ?? row?.name ?? '').trim()
}

function snapshotName(row?: Record<string, unknown>) {
  return String(row?.snap ?? row?.snapshot ?? row?.name ?? '').trim()
}

function namespacePool(row?: Record<string, unknown>) {
  return String(row?.pool ?? row?.pool_name ?? '').trim()
}

function namespaceName(row?: Record<string, unknown>) {
  return String(row?.namespace ?? row?.name ?? '').trim()
}

function trashPool(row?: Record<string, unknown>) {
  const pool = String(row?.pool ?? row?.pool_name ?? '').trim()
  const namespace = String(row?.namespace ?? '').trim()
  return namespace ? `${pool}/${namespace}` : pool
}

function trashImageID(row?: Record<string, unknown>) {
  return String(row?.image_id ?? row?.id ?? row?.name ?? '').trim()
}

function encodePair(left: string, right: string) {
  const bytes = new TextEncoder().encode(`${left}\u0000${right}`)
  let binary = ''
  bytes.forEach((byte) => {
    binary += String.fromCharCode(byte)
  })
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '')
}

const externalDefinitions: Record<'iscsi' | 'nvmeGateway' | 'nvmeTcp' | 'nvmeNamespaces' | 'nvmeListeners' | 'nvmeHosts' | 'nvmeConnections', ExternalListPageDefinition> = {
  iscsi: {
    title: 'iSCSI Targets',
    path: '/iscsi/targets',
    requiredEndpoints: ['iscsi'],
    rowKeyCandidates: ['iqn', 'target_iqn', 'name'],
    createAction: {
      title: '新建 iSCSI Target',
      buttonLabel: '新建 Target',
      path: '/iscsi/target',
      method: 'POST',
      successMessage: 'iSCSI Target 创建执行成功',
      fields: [
        { name: 'iqn', label: 'IQN', required: true, placeholder: 'iqn.2026-07.local.cephtower:target' },
        { name: 'portals_json', label: 'Portals JSON', type: 'textarea', placeholder: '[{"host":"host-01","ip":"10.0.0.10"}]' },
        { name: 'disks_json', label: 'Disks JSON', type: 'textarea', placeholder: '[{"pool":"rbd","image":"demo"}]' },
        { name: 'clients_json', label: 'Clients JSON', type: 'textarea', placeholder: '[{"iqn":"iqn.client","luns":["rbd/demo"]}]' }
      ],
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        iqn: String(values.iqn ?? ''),
        portals: parseJSONArray(values.portals_json),
        disks: parseJSONArray(values.disks_json),
        clients: parseJSONArray(values.clients_json)
      })
    },
    deleteAction: {
      title: '删除 iSCSI Target',
      path: '/iscsi/target',
      action: 'iscsi_target.delete',
      resourceKind: 'iscsi_target',
      successMessage: 'iSCSI Target 删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, iqn: targetIQN(row) }),
      resourceKey: (row) => `iscsi/target/${targetIQN(row)}`
    },
    columns: [
      { key: 'iqn', title: 'IQN' },
      { key: 'target_iqn', title: 'Target' },
      { key: 'portals', title: 'Portals' },
      { key: 'disks', title: 'LUN' },
      { key: 'clients', title: 'Initiators' },
      { key: 'status', title: '状态' }
    ]
  },
  nvmeGateway: {
    title: 'NVMe-oF Gateway',
    path: '/nvmeof/gateway',
    requiredEndpoints: ['nvmeof'],
    rowKeyCandidates: ['name', 'hostname', 'version'],
    columns: [
      { key: 'name', title: '名称' },
      { key: 'hostname', title: '主机' },
      { key: 'version', title: '版本' },
      { key: 'status', title: '状态' },
      { key: 'gateway_state', title: 'Gateway State' }
    ]
  },
  nvmeTcp: {
    title: 'NVMe-oF Subsystems',
    path: '/nvmeof/subsystems',
    requiredEndpoints: ['nvmeof'],
    rowKeyCandidates: ['nqn', 'subsystem_nqn', 'name'],
    createAction: {
      title: '新建 NVMe-oF Subsystem',
      buttonLabel: '新建 Subsystem',
      path: '/nvmeof/subsystem',
      method: 'POST',
      successMessage: 'NVMe-oF Subsystem 创建执行成功',
      fields: [
        { name: 'subsystem_nqn', label: 'Subsystem NQN', required: true, placeholder: 'nqn.2026-07.io.cephtower:subsystem' },
        { name: 'serial_number', label: '序列号' },
        { name: 'model_number', label: '型号' },
        { name: 'max_namespaces', label: '最大 Namespace', type: 'number', min: 1 },
        { name: 'enable_ha', label: '启用 HA', type: 'boolean' },
        { name: 'no_group_append', label: '禁用 group append', type: 'boolean' }
      ],
      initialValues: { enable_ha: false, no_group_append: false },
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        subsystem_nqn: String(values.subsystem_nqn ?? ''),
        ...(values.serial_number ? { serial_number: String(values.serial_number) } : {}),
        ...(values.model_number ? { model_number: String(values.model_number) } : {}),
        ...(values.max_namespaces ? { max_namespaces: Number(values.max_namespaces) } : {}),
        enable_ha: Boolean(values.enable_ha),
        no_group_append: Boolean(values.no_group_append)
      })
    },
    deleteAction: {
      title: '删除 NVMe-oF Subsystem',
      path: '/nvmeof/subsystem',
      action: 'nvmeof_subsystem.delete',
      resourceKind: 'nvmeof_subsystem',
      successMessage: 'NVMe-oF Subsystem 删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, subsystem_nqn: subsystemNQN(row) }),
      resourceKey: (row) => `nvmeof/subsystem/${subsystemNQN(row)}`
    },
    columns: [
      { key: 'nqn', title: 'NQN' },
      { key: 'subsystem_nqn', title: 'Subsystem' },
      { key: 'namespaces', title: 'Namespaces' },
      { key: 'listeners', title: 'Listeners' },
      { key: 'hosts', title: 'Hosts' },
      { key: 'status', title: '状态' }
    ]
  },
  nvmeNamespaces: {
    title: 'NVMe-oF Namespaces',
    path: '/nvmeof/subsystem/namespaces',
    requiredEndpoints: ['nvmeof'],
    rowKeyCandidates: ['nsid', 'namespace_id', 'uuid'],
    filterFields: [
      { name: 'subsystem_nqn', label: 'Subsystem NQN', required: true }
    ],
    createAction: {
      title: '新建 Namespace',
      buttonLabel: '新建 Namespace',
      path: '/nvmeof/subsystem/namespace',
      method: 'POST',
      successMessage: 'Namespace 创建执行成功',
      fields: [
        { name: 'subsystem_nqn', label: 'Subsystem NQN', required: true },
        { name: 'rbd_pool_name', label: 'RBD Pool', required: true },
        { name: 'rbd_image_name', label: 'RBD Image', required: true },
        { name: 'block_size', label: 'Block Size', type: 'number', min: 1 },
        { name: 'create_image', label: '创建镜像', type: 'boolean' },
        { name: 'size', label: '镜像大小 bytes', type: 'number', min: 1 },
        { name: 'force', label: 'Force', type: 'boolean' },
        { name: 'no_auto_visible', label: 'No Auto Visible', type: 'boolean' },
        { name: 'disable_auto_resize', label: 'Disable Auto Resize', type: 'boolean' }
      ],
      initialValues: { create_image: false, force: false, no_auto_visible: false, disable_auto_resize: false },
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        subsystem_nqn: String(values.subsystem_nqn ?? ''),
        rbd_pool_name: String(values.rbd_pool_name ?? ''),
        rbd_image_name: String(values.rbd_image_name ?? ''),
        ...(values.block_size ? { block_size: Number(values.block_size) } : {}),
        create_image: Boolean(values.create_image),
        ...(values.size ? { size: Number(values.size) } : {}),
        force: Boolean(values.force),
        no_auto_visible: Boolean(values.no_auto_visible),
        disable_auto_resize: Boolean(values.disable_auto_resize)
      })
    },
    updateAction: {
      title: '扩容 Namespace',
      path: '/nvmeof/subsystem/namespace',
      method: 'PATCH',
      successMessage: 'Namespace 扩容执行成功',
      fields: [
        { name: 'new_size', label: '新大小 bytes', type: 'number', required: true, min: 1 }
      ],
      buildBody: (values, clusterId, row) => ({
        cluster_id: clusterId,
        subsystem_nqn: nvmeSubsystem(row),
        nsid: nvmeNamespaceID(row),
        new_size: Number(values.new_size)
      })
    },
    deleteAction: {
      title: '删除 Namespace',
      path: '/nvmeof/subsystem/namespace',
      action: 'nvmeof_namespace.delete',
      resourceKind: 'nvmeof_namespace',
      successMessage: 'Namespace 删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, subsystem_nqn: nvmeSubsystem(row), nsid: nvmeNamespaceID(row), force: false }),
      resourceKey: (row) => `nvmeof/subsystem/${nvmeSubsystem(row)}/namespace/${nvmeNamespaceID(row)}`
    },
    columns: [
      { key: 'subsystem_nqn', title: 'Subsystem' },
      { key: 'nsid', title: 'NSID' },
      { key: 'rbd_pool_name', title: 'RBD Pool' },
      { key: 'rbd_image_name', title: 'RBD Image' },
      { key: 'size', title: '大小' },
      { key: 'uuid', title: 'UUID' },
      { key: 'status', title: '状态' }
    ]
  },
  nvmeListeners: {
    title: 'NVMe-oF Listeners',
    path: '/nvmeof/subsystem/listeners',
    requiredEndpoints: ['nvmeof'],
    rowKeyCandidates: ['listener_id', 'traddr', 'host_name'],
    filterFields: [
      { name: 'subsystem_nqn', label: 'Subsystem NQN', required: true }
    ],
    createAction: {
      title: '新建 Listener',
      buttonLabel: '新建 Listener',
      path: '/nvmeof/subsystem/listener',
      method: 'POST',
      successMessage: 'Listener 创建执行成功',
      fields: [
        { name: 'subsystem_nqn', label: 'Subsystem NQN', required: true },
        { name: 'host_name', label: 'Host Name', required: true },
        { name: 'traddr', label: 'Transport Address', required: true },
        { name: 'adrfam', label: 'Address Family', type: 'number' },
        { name: 'trsvcid', label: 'Transport Service ID', type: 'number' },
        { name: 'secure', label: 'Secure', type: 'boolean' },
        { name: 'verify_host_name', label: 'Verify Host Name', type: 'boolean' },
        { name: 'force', label: 'Force', type: 'boolean' }
      ],
      initialValues: { secure: false, verify_host_name: false, force: false },
      buildBody: (values, clusterId) => listenerBody(values, clusterId)
    },
    deleteAction: {
      title: '删除 Listener',
      path: '/nvmeof/subsystem/listener',
      action: 'nvmeof_listener.delete',
      resourceKind: 'nvmeof_listener',
      successMessage: 'Listener 删除执行成功',
      buildBody: (row, clusterId) => ({
        cluster_id: clusterId,
        subsystem_nqn: nvmeSubsystem(row),
        host_name: text(row?.host_name),
        traddr: text(row?.traddr),
        ...(row.trsvcid ? { trsvcid: Number(row.trsvcid) } : {}),
        force: false
      }),
      resourceKey: (row) => `nvmeof/subsystem/${nvmeSubsystem(row)}/listener/${text(row.listener_id ?? row.traddr)}`
    },
    columns: [
      { key: 'subsystem_nqn', title: 'Subsystem' },
      { key: 'host_name', title: 'Host' },
      { key: 'traddr', title: '地址' },
      { key: 'trsvcid', title: '端口' },
      { key: 'secure', title: 'Secure' },
      { key: 'status', title: '状态' }
    ]
  },
  nvmeHosts: {
    title: 'NVMe-oF Hosts',
    path: '/nvmeof/subsystem/hosts',
    requiredEndpoints: ['nvmeof'],
    rowKeyCandidates: ['host_nqn', 'nqn'],
    filterFields: [
      { name: 'subsystem_nqn', label: 'Subsystem NQN', required: true }
    ],
    createAction: {
      title: '新增 Host',
      buttonLabel: '新增 Host',
      path: '/nvmeof/subsystem/host',
      method: 'POST',
      successMessage: 'Host 添加执行成功',
      fields: [
        { name: 'subsystem_nqn', label: 'Subsystem NQN', required: true },
        { name: 'host_nqn', label: 'Host NQN', required: true },
        { name: 'psk', label: 'PSK' },
        { name: 'dhchap_key', label: 'DH-HMAC-CHAP Key' }
      ],
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        subsystem_nqn: String(values.subsystem_nqn ?? ''),
        host_nqn: String(values.host_nqn ?? ''),
        ...(values.psk ? { psk: String(values.psk) } : {}),
        ...(values.dhchap_key ? { dhchap_key: String(values.dhchap_key) } : {})
      })
    },
    deleteAction: {
      title: '删除 Host',
      path: '/nvmeof/subsystem/host',
      action: 'nvmeof_host.delete',
      resourceKind: 'nvmeof_host',
      successMessage: 'Host 删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, subsystem_nqn: nvmeSubsystem(row), host_nqn: nvmeHostNQN(row) }),
      resourceKey: (row) => `nvmeof/subsystem/${nvmeSubsystem(row)}/host/${nvmeHostNQN(row)}`
    },
    columns: [
      { key: 'subsystem_nqn', title: 'Subsystem' },
      { key: 'host_nqn', title: 'Host NQN' },
      { key: 'nqn', title: 'NQN' },
      { key: 'status', title: '状态' }
    ]
  },
  nvmeConnections: {
    title: 'NVMe-oF Connections',
    path: '/nvmeof/subsystem/connections',
    requiredEndpoints: ['nvmeof'],
    rowKeyCandidates: ['connection_id', 'host_nqn', 'traddr'],
    filterFields: [
      { name: 'subsystem_nqn', label: 'Subsystem NQN', required: true }
    ],
    columns: [
      { key: 'subsystem_nqn', title: 'Subsystem' },
      { key: 'host_nqn', title: 'Host NQN' },
      { key: 'traddr', title: '地址' },
      { key: 'trsvcid', title: '端口' },
      { key: 'nqn', title: 'NQN' },
      { key: 'status', title: '状态' }
    ]
  }
}

function targetIQN(row?: Record<string, unknown>) {
  return String(row?.iqn ?? row?.target_iqn ?? row?.name ?? '').trim()
}

function subsystemNQN(row?: Record<string, unknown>) {
  return String(row?.nqn ?? row?.subsystem_nqn ?? row?.name ?? '').trim()
}

function parseJSONArray(value: unknown) {
  if (!value) {
    return []
  }
  const parsed = JSON.parse(String(value))
  if (!Array.isArray(parsed)) {
    throw new Error('JSON 字段必须是数组')
  }
  return parsed
}

function nvmeSubsystem(row?: Record<string, unknown>) {
  return text(row?.subsystem_nqn ?? row?.nqn)
}

function nvmeNamespaceID(row?: Record<string, unknown>) {
  return text(row?.nsid ?? row?.namespace_id ?? row?.id)
}

function nvmeHostNQN(row?: Record<string, unknown>) {
  return text(row?.host_nqn ?? row?.nqn)
}

function listenerBody(values: Record<string, unknown>, clusterId: number) {
  return {
    cluster_id: clusterId,
    subsystem_nqn: String(values.subsystem_nqn ?? ''),
    host_name: String(values.host_name ?? ''),
    traddr: String(values.traddr ?? ''),
    ...(values.adrfam ? { adrfam: Number(values.adrfam) } : {}),
    ...(values.trsvcid ? { trsvcid: Number(values.trsvcid) } : {}),
    secure: Boolean(values.secure),
    verify_host_name: Boolean(values.verify_host_name),
    force: Boolean(values.force)
  }
}

function text(value: unknown) {
  return value === null || value === undefined ? '' : String(value).trim()
}

function encodeImageSpec(value: string) {
  const bytes = new TextEncoder().encode(value.trim())
  let binary = ''
  bytes.forEach((byte) => { binary += String.fromCharCode(byte) })
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '')
}

import { Alert, Button, Card, Input, Select, Space, Table, Tag } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { useClusterContext } from '../../state/ClusterContext'
import { renderPlacementClassDetails } from './RgwPlacementClasses'
import { filterStorageClasses, readStorageClassInventory, type StorageClassInventory, type StorageClassInventoryRow } from './rgwStorageClassInventory'

export function RgwStorageClassesPage() {
  const {selectedClusterId}=useClusterContext()
  return <RgwStorageClassesView key={selectedClusterId??'none'} clusterId={selectedClusterId}/>
}

export function RgwStorageClassesView({clusterId}:{clusterId?:number}) {
  const current=useRef(clusterId),sequence=useRef(0),abort=useRef<AbortController>()
  current.current=clusterId
  const [state,setState]=useState<{clusterId?:number;busy:boolean;error?:boolean;data?:StorageClassInventory}>({clusterId,busy:false})
  const [query,setQuery]=useState(''),[group,setGroup]=useState<string>(),[type,setType]=useState<string>()
  async function read() {
    if(!clusterId||current.current!==clusterId)return
    abort.current?.abort();const controller=new AbortController();abort.current=controller
    const ticket=++sequence.current
    setState({clusterId,busy:true})
    try {
      const data=await readStorageClassInventory(clusterId,controller.signal)
      if(!controller.signal.aborted&&current.current===clusterId&&ticket===sequence.current)setState({clusterId,busy:false,data})
    } catch {
      if(!controller.signal.aborted&&current.current===clusterId&&ticket===sequence.current)setState({clusterId,busy:false,error:true})
    }
  }
  useEffect(()=>{
    setQuery('');setGroup(undefined);setType(undefined);setState({clusterId,busy:false})
    void read()
    return()=>{abort.current?.abort();sequence.current++}
  },[clusterId])
  const scoped=state.clusterId===clusterId,data=scoped?state.data:undefined,busy=scoped&&state.busy
  const rows=data?filterStorageClasses(data.rows,query,group,type):[]
  const groups=[...new Map(data?.rows.map(row=>[row.groupId,{value:row.groupId,label:`${row.groupName} (${row.groupId})`}])).values()]
  const field=(row:StorageClassInventoryRow,name:string)=>row.fields.find(f=>f.name===name)?.value??(row.type==='本地'?'不适用':'未返回或不可用')
  return <Card title="RGW 存储类" extra={<Button disabled={!clusterId||busy} loading={busy} onClick={()=>void read()}>重新读取库存</Button>}>
    <Space direction="vertical" style={{width:'100%'}}>
      <Alert type="info" message="跨 Zonegroup 的类声明与分层配置" description="读取 zonegroup list/get 已采集的全部分页库存，包含 STANDARD、普通本地类和云分层；同名类按组 ID 与放置目标区分。库存不是实时或原子快照，不证明各 Zone 池已配置、远端可达或 Period 已发布。"/>
      <Space wrap><Link to="/object/multisite/zonegroups">前往 ZoneGroups 创建、编辑或删除存储类</Link><Link to="/object/multisite/zones">前往 Zones 管理本地池映射</Link></Space>
      {!clusterId&&<Alert type="info" message="请先选择集群"/>}
      {scoped&&state.error&&<Alert type="error" message="读取失败、分页不完整或组身份异常；未展示部分结果，请检查采集状态后重试"/>}
      {data&&<>
        {data.stale&&<Alert type="warning" message="库存包含过期数据，不能据此确认当前配置；重新读取不会触发 Ceph 采集"/>}
        {data.issues.length>0&&<Alert type="warning" message="部分存储类配置异常" description={data.issues.join('；')}/>}
        <p>本次读取 {data.groups} 个组、{data.rows.length} 条可展示的类记录；当前筛选 {rows.length} 条。请按下表组名/ID 在管理页选择对应组。</p>
        <Space wrap>
          <Input aria-label="搜索存储类" placeholder="搜索组、目标、类、区域或端点" value={query} onChange={event=>setQuery(event.target.value)} allowClear style={{width:300}}/>
          <Select aria-label="筛选 Zonegroup" placeholder="全部 Zonegroup" value={group} onChange={setGroup} allowClear options={groups} style={{minWidth:230}}/>
          <Select aria-label="筛选存储类类型" placeholder="全部类型" value={type} onChange={setType} allowClear options={[...new Set(data.rows.map(row=>row.type))].map(value=>({value,label:value}))} style={{minWidth:180}}/>
        </Space>
        <Table key={JSON.stringify([query,group,type])} size="small" rowKey="key" dataSource={rows} pagination={{pageSize:20,showSizeChanger:true}} scroll={{x:1400}}
          locale={{emptyText:data.rows.length?'没有匹配筛选的存储类':'本次库存没有可展示记录，不代表已验证没有配置'}}
          columns={[
            {title:'存储类',dataIndex:'storageClass'},{title:'类型',dataIndex:'type'},{title:'Zonegroup',dataIndex:'groupName'},{title:'组 ID',dataIndex:'groupId'},
            {title:'放置目标',dataIndex:'placement'},{title:'Realm ID',dataIndex:'realm'},{title:'声明状态',dataIndex:'declared'},
            {title:'目标区域',render:(_value:unknown,row:StorageClassInventoryRow)=>field(row,'目标区域')},{title:'目标端点',render:(_value:unknown,row:StorageClassInventoryRow)=>field(row,'目标端点')},
            {title:'库存状态',render:(_value:unknown,row:StorageClassInventoryRow)=><Tag color={row.stale?'warning':'default'}>{row.stale?'过期':'采集快照'}</Tag>},{title:'采集时间',dataIndex:'observed'}
          ]} expandable={{expandedRowRender:renderPlacementClassDetails}}/>
      </>}
    </Space>
  </Card>
}

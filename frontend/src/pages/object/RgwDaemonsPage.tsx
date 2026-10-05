import { Alert, Button, Card, Input, Space, Table } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { jsonInit, request } from '../../api/client'
import { useClusterContext } from '../../state/ClusterContext'
import { RgwDaemonStatus } from './RgwDaemonStatus'

const fields=['service_map_id','id','hostname','version','realm_name','realm_id','zonegroup_name','zonegroup_id','zone_name','zone_id'] as const
type Daemon=Record<typeof fields[number],string>&{metadata:Record<string,string>;listeners:Array<{frontend:string;tls:boolean;port:number}>;listeners_complete:boolean}
export function daemonRegistrationData(value:unknown):{items:Daemon[];observed_at:string} {
  const data=value as {items:unknown[];observed_at:string;source:string}
  if(!data||data.source!=='service_map'||typeof data.observed_at!=='string'||!data.observed_at||!Array.isArray(data.items))throw new Error('Invalid service map')
  const seen=new Set<string>()
  const items=data.items.map(value=>{
    const row=value as Daemon
    if(!row||fields.some(key=>typeof row[key]!=='string')||!row.id||!row.service_map_id||seen.has(row.service_map_id))throw new Error('Invalid registration')
    seen.add(row.service_map_id)
    if(!row.metadata||typeof row.metadata!=='object'||Array.isArray(row.metadata)||Object.values(row.metadata).some(value=>typeof value!=='string'))throw new Error('Invalid metadata')
    if(typeof row.listeners_complete!=='boolean'||!Array.isArray(row.listeners)||row.listeners.some(item=>!item||typeof item.frontend!=='string'||!/^frontend_config#\d+$/.test(item.frontend)||typeof item.tls!=='boolean'||!Number.isInteger(item.port)||item.port<1||item.port>65535))throw new Error('Invalid listener ports')
    const metadata=Object.fromEntries(['ceph_version_short','ceph_release','os','kernel_version','kernel_description','arch','pod_name','container_name','container_image','container_hostname','pod_namespace','num_handles','service_unique_id'].filter(key=>Object.prototype.hasOwnProperty.call(row.metadata,key)).map(key=>[key,row.metadata[key]]))
    return {...Object.fromEntries(fields.map(key=>[key,row[key]])),metadata,listeners:row.listeners.map(({frontend,tls,port})=>({frontend,tls,port})),listeners_complete:row.listeners_complete} as Daemon
  })
  return {items,observed_at:data.observed_at}
}
export function RgwDaemonsPage(){const {selectedClusterId}=useClusterContext();return <RgwDaemonsView key={selectedClusterId??'none'} clusterId={selectedClusterId}/>}
export function RgwDaemonsView({clusterId}:{clusterId?:number}) {
  const [filter,setFilter]=useState(''),[state,setState]=useState<{clusterId?:number;busy:boolean;data?:ReturnType<typeof daemonRegistrationData>;error?:boolean}>({busy:false})
  const current=useRef(clusterId),mounted=useRef(true),sequence=useRef(0),abort=useRef<AbortController>()
  current.current=clusterId
  useEffect(()=>{mounted.current=true;setFilter('');setState({clusterId,busy:false});return()=>{mounted.current=false;abort.current?.abort();sequence.current++}},[clusterId])
  async function read(){
    if(!clusterId||!mounted.current||current.current!==clusterId)return
    abort.current?.abort();const controller=new AbortController();abort.current=controller;const ticket=++sequence.current
    setState({clusterId,busy:true})
    try{
      const result=await request<unknown>('/rgw/daemons',jsonInit('GET',{cluster_id:clusterId},{signal:controller.signal,cache:'no-store',suppressErrorNotification:true}))
      if(mounted.current&&current.current===clusterId&&ticket===sequence.current&&!controller.signal.aborted)setState({clusterId,busy:false,data:daemonRegistrationData(result)})
    }catch{if(mounted.current&&current.current===clusterId&&ticket===sequence.current&&!controller.signal.aborted)setState({clusterId,busy:false,error:true})}
  }
  const scoped=state.clusterId===clusterId?state:undefined
  const labels=['服务映射 ID','RGW ID','主机','版本','Realm','Realm ID','Zonegroup','Zonegroup ID','Zone','Zone ID']
  const rows=scoped?.data?.items.filter(row=>fields.some(key=>row[key].toLowerCase().includes(filter.toLowerCase())))
  const readSequence=sequence.current
  return <Card title="RGW 守护进程注册"><Space direction="vertical" style={{width:'100%'}}>
    <Alert type="info" message="来自 ceph service dump 的注册信息，不是实时健康探测" description="同一 RGW ID 的不同服务映射记录分别保留；未注册不等于进程不存在。端口从已注册 Beast frontend 配置提取，不证明监听成功、TLS 可用或网络可达；未知配置会明确标注。"/>
    <Space wrap><Button disabled={!clusterId||scoped?.busy} loading={scoped?.busy} onClick={()=>void read()}>读取注册列表</Button><Input aria-label="搜索 RGW 注册" value={filter} onChange={event=>setFilter(event.target.value)} placeholder="ID、主机、版本或多站点归属"/></Space>
    {!clusterId&&<Alert type="info" message="请先选择集群"/>}
    {scoped?.error&&<Alert type="error" message="读取失败或注册数据不完整，未展示部分结果"/>}
    {scoped?.data&&<><span>读取时间：{scoped.data.observed_at}</span><Table rowKey="service_map_id" size="small" dataSource={rows} pagination={{pageSize:20}} scroll={{x:1300}} locale={{emptyText:scoped.data.items.length?'没有匹配的注册记录':'服务映射中没有 RGW 注册记录'}}
      expandable={{expandedRowRender:row=><RgwDaemonStatus key={`${clusterId}/${readSequence}/${row.service_map_id}`} clusterId={clusterId!} serviceMapId={row.service_map_id} metadata={row.metadata} isCurrent={()=>mounted.current&&current.current===clusterId&&sequence.current===readSequence}/>}}
      columns={[...fields.map((key,i)=>({title:labels[i],dataIndex:key,render:(value:string)=>value||'—',sorter:(a:Daemon,b:Daemon)=>a[key].localeCompare(b[key])})),{title:'注册配置端口',render:(_value:unknown,row:Daemon)=><span style={{whiteSpace:'pre-wrap'}}>{row.listeners.map(item=>`${item.frontend}: ${item.tls?'HTTPS':'HTTP'} ${item.port}`).join('\n')}{!row.listeners_complete?'\n部分或全部配置无法解析':''}</span>}]}/></>}
  </Space></Card>
}

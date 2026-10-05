import { Alert, Button, Card, Descriptions, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { jsonInit, request } from '../../api/client'
import { useClusterContext } from '../../state/ClusterContext'

type Counts={realm_count:number;zonegroup_count:number;zone_count:number;started_at:string;observed_at:string;source:'radosgw-admin'}
export function topologyCountData(value:unknown):Counts {
  const data=value as Counts
  if(!data||data.source!=='radosgw-admin'||['realm_count','zonegroup_count','zone_count'].some(key=>!Number.isSafeInteger(data[key as keyof Counts])||Number(data[key as keyof Counts])<0)||typeof data.started_at!=='string'||typeof data.observed_at!=='string'||!Number.isFinite(Date.parse(data.started_at))||!Number.isFinite(Date.parse(data.observed_at))||Date.parse(data.started_at)>Date.parse(data.observed_at))throw new Error('Invalid counts')
  return {realm_count:data.realm_count,zonegroup_count:data.zonegroup_count,zone_count:data.zone_count,started_at:data.started_at,observed_at:data.observed_at,source:'radosgw-admin'}
}
export function RgwTopologyCounts(){const {selectedClusterId}=useClusterContext();return <RgwTopologyCountsView key={selectedClusterId??'none'} clusterId={selectedClusterId}/>}
export function RgwTopologyCountsView({clusterId}:{clusterId?:number}) {
  const current=useRef(clusterId),mounted=useRef(true),sequence=useRef(0),abort=useRef<AbortController>()
  current.current=clusterId
  const [state,setState]=useState<{clusterId?:number;busy:boolean;data?:Counts;error?:boolean}>({busy:false})
  useEffect(()=>{mounted.current=true;setState({clusterId,busy:false});return()=>{mounted.current=false;abort.current?.abort();sequence.current++}},[clusterId])
  async function read(){
    if(!clusterId||!mounted.current||current.current!==clusterId)return
    abort.current?.abort();const controller=new AbortController();abort.current=controller;const ticket=++sequence.current
    setState({clusterId,busy:true})
    try{
      const value=await request<unknown>('/rgw/topology/counts',jsonInit('GET',{cluster_id:clusterId},{signal:controller.signal,cache:'no-store',suppressErrorNotification:true}))
      if(mounted.current&&current.current===clusterId&&ticket===sequence.current&&!controller.signal.aborted)setState({clusterId,busy:false,data:topologyCountData(value)})
    }catch{if(mounted.current&&current.current===clusterId&&ticket===sequence.current&&!controller.signal.aborted)setState({clusterId,busy:false,error:true})}
  }
  const scoped=state.clusterId===clusterId?state:undefined
  return <Card title="本地多站点配置数量"><Space direction="vertical" style={{width:'100%'}}>
    <Alert type="info" message="来自 radosgw-admin 完整配置列表" description="统计本集群可读取的 Realm、Zonegroup、Zone 配置，不是在线站点或远端拓扑总数。三次读取不是原子快照，任一失败时不展示部分计数。"/>
    <Button disabled={!clusterId||scoped?.busy} loading={scoped?.busy} onClick={()=>void read()}>读取配置数量</Button>
    {!clusterId&&<Alert type="info" message="请先选择集群"/>}
    {scoped?.error&&<Alert type="error" message="配置统计读取失败或响应无效，未展示旧值或部分计数"/>}
    {scoped?.data&&<Descriptions column={1} items={[
      {key:'realm',label:'Realm',children:String(scoped.data.realm_count)},
      {key:'group',label:'Zonegroup',children:String(scoped.data.zonegroup_count)},
      {key:'zone',label:'Zone',children:String(scoped.data.zone_count)},
      {key:'start',label:'读取开始',children:scoped.data.started_at},
      {key:'end',label:'读取完成',children:scoped.data.observed_at},
    ]}/>}
  </Space></Card>
}

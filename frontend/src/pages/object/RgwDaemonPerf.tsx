import { Alert, Button, Space, Table } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { jsonInit, request } from '../../api/client'
import { RgwDaemonHistory } from './RgwDaemonHistory'

type Sample={metric:Record<string,string>;value:[number,string]}
type Snapshot={service_map_id:string;source:'prometheus';result_type:'vector';available:boolean;observed_at:string;series:Sample[]}
export function daemonPerfData(value:unknown,id:string):Snapshot {
  const data=value as Snapshot
  if(!data||data.service_map_id!==id||data.source!=='prometheus'||data.result_type!=='vector'||typeof data.observed_at!=='string'||!data.observed_at||!Array.isArray(data.series)||data.available!==(data.series.length>0))throw new Error('Invalid snapshot')
  const series=data.series.map(row=>{
    if(!row||!row.metric||typeof row.metric!=='object'||Array.isArray(row.metric)||Object.values(row.metric).some(v=>typeof v!=='string')||row.metric.instance_id!==id||!row.metric.__name__?.startsWith('ceph_rgw_')||!Array.isArray(row.value)||row.value.length!==2||typeof row.value[0]!=='number'||!Number.isFinite(row.value[0])||typeof row.value[1]!=='string')throw new Error('Invalid sample')
    return {metric:Object.fromEntries(Object.entries(row.metric)),value:[row.value[0],row.value[1]] as [number,string]}
  })
  return {service_map_id:id,source:'prometheus',result_type:'vector',available:data.available,observed_at:data.observed_at,series}
}
export function RgwDaemonPerf({clusterId,serviceMapId,isCurrent}:{clusterId:number;serviceMapId:string;isCurrent:()=>boolean}) {
  const scope=JSON.stringify([clusterId,serviceMapId]),current=useRef(scope),mounted=useRef(true),sequence=useRef(0),abort=useRef<AbortController>()
  current.current=scope
  const [state,setState]=useState<{scope:string;busy:boolean;data?:Snapshot;error?:boolean}>({scope,busy:false})
  useEffect(()=>{mounted.current=true;setState({scope,busy:false});return()=>{mounted.current=false;abort.current?.abort();sequence.current++}},[scope])
  async function read(){
    if(!mounted.current||current.current!==scope||!isCurrent())return
    abort.current?.abort();const controller=new AbortController();abort.current=controller;const ticket=++sequence.current
    setState({scope,busy:true})
    try{
      const value=await request<unknown>('/rgw/daemon/perf',jsonInit('GET',{cluster_id:clusterId,service_map_id:serviceMapId},{signal:controller.signal,cache:'no-store',suppressErrorNotification:true}))
      if(mounted.current&&current.current===scope&&isCurrent()&&ticket===sequence.current&&!controller.signal.aborted)setState({scope,busy:false,data:daemonPerfData(value,serviceMapId)})
    }catch{if(mounted.current&&current.current===scope&&isCurrent()&&ticket===sequence.current&&!controller.signal.aborted)setState({scope,busy:false,error:true})}
  }
  const scoped=state.scope===scope?state:undefined
  return <Space direction="vertical" style={{width:'100%'}}>
    <Alert type="info" message={`服务映射 ${serviceMapId} 的 Prometheus 性能快照`} description="需要配置集群 Prometheus 并导出对应 instance_id 的 ceph_rgw_* 指标。数值为采集样本，不是 Dashboard 内部历史速率或健康结论；多个采集来源分别保留，不相加。单位取决于指标定义，时间为 Unix 秒。"/>
    <Button loading={scoped?.busy} disabled={scoped?.busy} onClick={()=>void read()}>读取性能快照</Button>
    {scoped?.error&&<Alert type="error" message="性能快照读取失败或响应无效，请检查监控端点与实例标签后重试"/>}
    {scoped?.data&&<><span>读取时间：{scoped.data.observed_at}</span>{!scoped.data.available?<Alert type="warning" message="没有匹配的导出样本，性能数据不可用（不表示零值）"/>:<Table size="small" rowKey="key" pagination={{pageSize:20}} scroll={{x:1000}} dataSource={scoped.data.series.map((row,key)=>({...row,key}))} columns={[
      {title:'指标',render:(_value:unknown,row:Sample)=>row.metric.__name__},
      {title:'采集标签',render:(_value:unknown,row:Sample)=><pre style={{whiteSpace:'pre-wrap',overflowWrap:'anywhere',margin:0}}>{JSON.stringify(row.metric,null,2)}</pre>},
      {title:'样本时间（Unix 秒）',render:(_value:unknown,row:Sample)=>String(row.value[0])},
      {title:'原始数值',render:(_value:unknown,row:Sample)=>row.value[1]},
    ]}/>}</>}
    <RgwDaemonHistory clusterId={clusterId} serviceMapId={serviceMapId} isCurrent={isCurrent}/>
  </Space>
}

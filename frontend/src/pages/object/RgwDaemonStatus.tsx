import { Alert, Button, Descriptions, Space, Table } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { jsonInit, request } from '../../api/client'
import { RgwDaemonPerf } from './RgwDaemonPerf'

type Status={service_map_id:string;status_stamp:string;last_beacon:string;observed_at:string;status:Record<string,string>}
export function daemonStatusData(value:unknown,id:string):Status {
  const data=value as Status
  if(!data||data.service_map_id!==id||typeof data.status_stamp!=='string'||typeof data.last_beacon!=='string'||typeof data.observed_at!=='string'||!data.observed_at||!data.status||typeof data.status!=='object'||Array.isArray(data.status)||Object.values(data.status).some(value=>typeof value!=='string'))throw new Error('Invalid status')
  return {service_map_id:id,status_stamp:data.status_stamp,last_beacon:data.last_beacon,observed_at:data.observed_at,status:Object.fromEntries(Object.entries(data.status))}
}
export function RgwDaemonStatus({clusterId,serviceMapId,isCurrent}:{clusterId:number;serviceMapId:string;isCurrent:()=>boolean}) {
  const scope=JSON.stringify([clusterId,serviceMapId]),current=useRef(scope),mounted=useRef(true),sequence=useRef(0),abort=useRef<AbortController>()
  current.current=scope
  const [state,setState]=useState<{scope:string;busy:boolean;data?:Status;error?:boolean}>({scope,busy:false})
  useEffect(()=>{mounted.current=true;setState({scope,busy:false});return()=>{mounted.current=false;abort.current?.abort();sequence.current++}},[scope])
  async function read(){
    if(!mounted.current||current.current!==scope||!isCurrent())return
    abort.current?.abort();const controller=new AbortController();abort.current=controller;const ticket=++sequence.current
    setState({scope,busy:true})
    try{
      const value=await request<unknown>('/rgw/daemon/status',jsonInit('GET',{cluster_id:clusterId,service_map_id:serviceMapId},{signal:controller.signal,cache:'no-store',suppressErrorNotification:true}))
      if(mounted.current&&current.current===scope&&isCurrent()&&ticket===sequence.current&&!controller.signal.aborted)setState({scope,busy:false,data:daemonStatusData(value,serviceMapId)})
    }catch{if(mounted.current&&current.current===scope&&isCurrent()&&ticket===sequence.current&&!controller.signal.aborted)setState({scope,busy:false,error:true})}
  }
  const scoped=state.scope===scope?state:undefined
  return <Space direction="vertical" style={{width:'100%'}}>
    <Alert type="info" message={`服务映射 ${serviceMapId} 的原生状态`} description="状态时间与最后 beacon 为 Ceph 原始报告，不据此推断存活或健康。状态内容已由后端脱敏；JSON 状态按文本展示以保留大整数精度。"/>
    <Button loading={scoped?.busy} disabled={scoped?.busy} onClick={()=>void read()}>读取服务状态</Button>
    {scoped?.error&&<Alert type="error" message="状态不可用、记录已消失或响应无效，请刷新注册列表后重试"/>}
    {scoped?.data&&<><Descriptions column={1} items={[{key:'stamp',label:'状态时间',children:scoped.data.status_stamp||'未返回'},{key:'beacon',label:'最后 beacon',children:scoped.data.last_beacon||'未返回'},{key:'read',label:'读取时间',children:scoped.data.observed_at}]}/><Table size="small" rowKey="name" pagination={false} locale={{emptyText:'此记录未上报状态字段（不代表健康）'}} dataSource={Object.entries(scoped.data.status).map(([name,value])=>({name,value}))} columns={[{title:'状态字段',dataIndex:'name'},{title:'值',render:(_value:unknown,row:{value:string})=><pre style={{whiteSpace:'pre-wrap',overflowWrap:'anywhere',margin:0}}>{row.value===''?'""':row.value}</pre>}]}/></>}
    <RgwDaemonPerf clusterId={clusterId} serviceMapId={serviceMapId} isCurrent={isCurrent}/>
  </Space>
}

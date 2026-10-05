import { Alert, Button, Space, Table } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { jsonInit, request } from '../../api/client'

type Series={metric:Record<string,string>;values:Array<[number,string]>}
type History={service_map_id:string;source:'prometheus';result_type:'matrix';available:boolean;observed_at:string;start:string;end:string;step_seconds:60;series:Series[]}
export function daemonHistoryData(value:unknown,id:string):History {
  const data=value as History
  if(!data||data.service_map_id!==id||data.source!=='prometheus'||data.result_type!=='matrix'||typeof data.observed_at!=='string'||!data.observed_at||typeof data.start!=='string'||typeof data.end!=='string'||data.step_seconds!==60||!Array.isArray(data.series)||data.available!==(data.series.length>0))throw new Error('Invalid history')
  const start=Date.parse(data.start)/1000,end=Date.parse(data.end)/1000
  if(!Number.isFinite(start)||!Number.isFinite(end)||end-start!==3600)throw new Error('Invalid window')
  const series=data.series.map(row=>{
    if(!row||!row.metric||typeof row.metric!=='object'||Array.isArray(row.metric)||Object.values(row.metric).some(v=>typeof v!=='string')||row.metric.instance_id!==id||!row.metric.__name__?.startsWith('ceph_rgw_')||!Array.isArray(row.values)||!row.values.length||row.values.length>61)throw new Error('Invalid series')
    let previous=-Infinity
    const values=row.values.map(sample=>{
      if(!Array.isArray(sample)||sample.length!==2||typeof sample[0]!=='number'||!Number.isFinite(sample[0])||sample[0]<start||sample[0]>end||sample[0]<=previous||typeof sample[1]!=='string')throw new Error('Invalid sample')
      previous=sample[0]
      return [sample[0],sample[1]] as [number,string]
    })
    return {metric:Object.fromEntries(Object.entries(row.metric)),values}
  })
  return {service_map_id:id,source:'prometheus',result_type:'matrix',available:data.available,observed_at:data.observed_at,start:data.start,end:data.end,step_seconds:60,series}
}
export function RgwDaemonHistory({clusterId,serviceMapId,isCurrent}:{clusterId:number;serviceMapId:string;isCurrent:()=>boolean}) {
  const scope=JSON.stringify([clusterId,serviceMapId]),current=useRef(scope),mounted=useRef(true),sequence=useRef(0),abort=useRef<AbortController>()
  current.current=scope
  const [state,setState]=useState<{scope:string;busy:boolean;data?:History;error?:boolean}>({scope,busy:false})
  useEffect(()=>{mounted.current=true;setState({scope,busy:false});return()=>{mounted.current=false;abort.current?.abort();sequence.current++}},[scope])
  async function read(){
    if(!mounted.current||current.current!==scope||!isCurrent())return
    abort.current?.abort();const controller=new AbortController();abort.current=controller;const ticket=++sequence.current
    setState({scope,busy:true})
    try{
      const value=await request<unknown>('/rgw/daemon/perf/history',jsonInit('GET',{cluster_id:clusterId,service_map_id:serviceMapId},{signal:controller.signal,cache:'no-store',suppressErrorNotification:true}))
      if(mounted.current&&current.current===scope&&isCurrent()&&ticket===sequence.current&&!controller.signal.aborted)setState({scope,busy:false,data:daemonHistoryData(value,serviceMapId)})
    }catch{if(mounted.current&&current.current===scope&&isCurrent()&&ticket===sequence.current&&!controller.signal.aborted)setState({scope,busy:false,error:true})}
  }
  const scoped=state.scope===scope?state:undefined
  return <Space direction="vertical" style={{width:'100%'}}>
    <Alert type="info" message="最近一小时的导出指标历史" description="固定 60 秒步长。时间为 Prometheus 评估时间，不是原始抓取时间；可能重复使用最近样本。展开序列查看原始值，不计算速率、不合并来源、不补齐缺失点。"/>
    <Button loading={scoped?.busy} disabled={scoped?.busy} onClick={()=>void read()}>读取一小时历史</Button>
    {scoped?.error&&<Alert type="error" message="历史数据读取失败或响应无效，请检查监控配置后重试"/>}
    {scoped?.data&&<><span>窗口：{scoped.data.start} — {scoped.data.end}；读取时间：{scoped.data.observed_at}</span>{!scoped.data.available?<Alert type="warning" message="此实例没有匹配的历史序列，不表示零值"/>:<Table size="small" rowKey="key" pagination={{pageSize:10}} dataSource={scoped.data.series.map((row,key)=>({...row,key}))} columns={[
      {title:'指标',render:(_value:unknown,row:Series)=>row.metric.__name__},
      {title:'采集标签',render:(_value:unknown,row:Series)=><pre style={{whiteSpace:'pre-wrap',overflowWrap:'anywhere',margin:0}}>{JSON.stringify(row.metric,null,2)}</pre>},
      {title:'评估点数',render:(_value:unknown,row:Series)=>row.values.length},
    ]} expandable={{expandedRowRender:row=><Table size="small" rowKey="timestamp" pagination={{pageSize:15}} dataSource={row.values.map(([timestamp,value])=>({timestamp,value}))} columns={[{title:'评估时间（Unix 秒）',dataIndex:'timestamp'},{title:'原始数值',dataIndex:'value'}]}/>}}/>}</>}
  </Space>
}

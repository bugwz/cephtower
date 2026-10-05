import { Alert, Button, Card, Space, Table } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { jsonInit, request } from '../../api/client'
import { useClusterContext } from '../../state/ClusterContext'

const metrics=[
  {id:'rgw_request_rate',name:'请求率',unit:'请求/s'},
  {id:'rgw_get_bytes_rate',name:'GET 带宽',unit:'B/s'},
  {id:'rgw_put_bytes_rate',name:'PUT 带宽',unit:'B/s'},
  {id:'rgw_get_latency_ms',name:'GET 平均延迟',unit:'ms'},
  {id:'rgw_put_latency_ms',name:'PUT 平均延迟',unit:'ms'},
]
type Value={value?:string;timestamp?:number;status:string}
export function overviewMetricValue(value:unknown):Value {
  const data=value as {result_type:string;series:Array<{metric:Record<string,string>;value:[number,string]}>}
  if(!data||data.result_type!=='vector'||!Array.isArray(data.series)||data.series.length>1)throw new Error('Invalid aggregate')
  if(!data.series.length)return {status:'无样本'}
  const row=data.series[0]
  if(!row?.metric||typeof row.metric!=='object'||Array.isArray(row.metric)||Object.keys(row.metric).length!==0||!Array.isArray(row.value)||row.value.length!==2||typeof row.value[0]!=='number'||!Number.isFinite(row.value[0])||typeof row.value[1]!=='string')throw new Error('Invalid sample')
  const [timestamp,text]=row.value
  if(['NaN','+Inf','-Inf'].includes(text))return {status:'不可计算',value:text,timestamp}
  if(!/^[+-]?(?:\d+(?:\.\d*)?|\.\d+)(?:[eE][+-]?\d+)?$/.test(text)||!Number.isFinite(Number(text))||Number(text)<0)throw new Error('Invalid value')
  return {status:'已读取',value:text,timestamp}
}
export function RgwOverviewMetrics(){const {selectedClusterId}=useClusterContext();return <RgwOverviewMetricsView key={selectedClusterId??'none'} clusterId={selectedClusterId}/>}
export function RgwOverviewMetricsView({clusterId}:{clusterId?:number}) {
  const current=useRef(clusterId),mounted=useRef(true),sequence=useRef(0),abort=useRef<AbortController>()
  current.current=clusterId
  const [state,setState]=useState<{clusterId?:number;busy:boolean;time?:string;rows?:Array<typeof metrics[number]&Value>}>({busy:false})
  useEffect(()=>{mounted.current=true;setState({clusterId,busy:false});return()=>{mounted.current=false;abort.current?.abort();sequence.current++}},[clusterId])
  async function read(){
    if(!clusterId||!mounted.current||current.current!==clusterId)return
    abort.current?.abort();const controller=new AbortController();abort.current=controller;const ticket=++sequence.current,time=new Date().toISOString()
    setState({clusterId,busy:true})
    const rows=await Promise.all(metrics.map(async metric=>{
      try{
        const params=new URLSearchParams({metric_id:metric.id,time})
        const result=await request<unknown>(`/metric/query?${params}`,jsonInit('GET',{cluster_id:clusterId},{signal:controller.signal,cache:'no-store',suppressErrorNotification:true}))
        return {...metric,...overviewMetricValue(result)}
      }catch{return {...metric,status:'读取失败或响应无效'}}
    }))
    if(mounted.current&&current.current===clusterId&&ticket===sequence.current&&!controller.signal.aborted)setState({clusterId,busy:false,time,rows})
  }
  const scoped=state.clusterId===clusterId?state:undefined
  return <Card title="RGW 性能概览"><Space direction="vertical" style={{width:'100%'}}>
    <Alert type="info" message="Prometheus 一分钟速率窗口" description="聚合端点内所有匹配 RGW 序列；端点必须限定当前集群并避免重复采集。平均延迟按操作计数加权；无样本或不可计算不表示零值，也不表示健康。"/>
    <Button disabled={!clusterId||scoped?.busy} loading={scoped?.busy} onClick={()=>void read()}>读取性能概览</Button>
    {!clusterId&&<Alert type="info" message="请先选择集群"/>}
    {scoped?.rows&&<><span>查询时间：{scoped.time}</span><Table size="small" rowKey="id" pagination={false} dataSource={scoped.rows} columns={[{title:'指标',dataIndex:'name'},{title:'值',dataIndex:'value',render:(value?:string)=>value??'—'},{title:'单位',dataIndex:'unit'},{title:'状态',dataIndex:'status'},{title:'评估时间（Unix 秒）',dataIndex:'timestamp',render:(value?:number)=>value===undefined?'—':String(value)}]}/></>}
  </Space></Card>
}

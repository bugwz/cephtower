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
type Point={time:number;raw:string;value?:number}
export function overviewHistoryPoints(value:unknown,start:number,end:number):Point[] {
  const data=value as {result_type:string;series:Array<{metric:Record<string,string>;values:Array<[number,string]>}>}
  if(!data||data.result_type!=='matrix'||!Array.isArray(data.series)||data.series.length>1)throw new Error('Invalid history')
  if(!data.series.length)return []
  const row=data.series[0]
  if(!row||!Array.isArray(row.values)||row.values.length>61)throw new Error('Invalid samples')
  let previous=-Infinity
  return row.values.map(sample=>{
    const result=overviewMetricValue({result_type:'vector',series:[{metric:row.metric,value:sample}]})
    const time=result.timestamp!
    if(time<start||time>end||time<=previous)throw new Error('Invalid timestamp')
    previous=time
    return {time,raw:result.value!,value:result.status==='已读取'?Number(result.value):undefined}
  })
}
export function overviewHistoryPath(points:Point[],start:number,end:number):string {
  const max=Math.max(1,...points.map(point=>point.value??0))
  let previous:number|undefined
  return points.map(point=>{
    if(point.value===undefined){previous=undefined;return ''}
    const command=previous===undefined||point.time-previous>60?'M':'L'
    previous=point.time
    return `${command}${40+(point.time-start)/(end-start)*700},${110-point.value/max*80}`
  }).filter(Boolean).join(' ')
}
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
  const [state,setState]=useState<{clusterId?:number;busy:boolean;time?:string;start?:number;end?:number;history?:boolean;rows?:Array<typeof metrics[number]&Value&{points?:Point[]}>}>({busy:false})
  useEffect(()=>{mounted.current=true;setState({clusterId,busy:false});return()=>{mounted.current=false;abort.current?.abort();sequence.current++}},[clusterId])
  async function read(history=false){
    if(!clusterId||!mounted.current||current.current!==clusterId)return
    abort.current?.abort();const controller=new AbortController();abort.current=controller;const ticket=++sequence.current,time=new Date().toISOString()
    setState({clusterId,busy:true})
    const end=Math.floor(Date.parse(time)/1000),start=end-3600
    const rows=await Promise.all(metrics.map(async metric=>{
      try{
        const params=new URLSearchParams(history?{metric_id:metric.id,start:new Date(start*1000).toISOString(),end:new Date(end*1000).toISOString(),step:'60s'}:{metric_id:metric.id,time})
        const result=await request<unknown>(`/metric/${history?'range':'query'}?${params}`,jsonInit('GET',{cluster_id:clusterId},{signal:controller.signal,cache:'no-store',suppressErrorNotification:true}))
        if(history){const points=overviewHistoryPoints(result,start,end);return {...metric,points,status:points.some(point=>point.value!==undefined)?'已读取':'无可绘制样本'}}
        return {...metric,...overviewMetricValue(result)}
      }catch{return {...metric,status:'读取失败或响应无效'}}
    }))
    if(mounted.current&&current.current===clusterId&&ticket===sequence.current&&!controller.signal.aborted)setState({clusterId,busy:false,time,rows,history,start,end})
  }
  const scoped=state.clusterId===clusterId?state:undefined
  return <Card title="RGW 性能概览"><Space direction="vertical" style={{width:'100%'}}>
    <Alert type="info" message="Prometheus 一分钟速率窗口" description="聚合端点内所有匹配 RGW 序列；端点必须限定当前集群并避免重复采集。平均延迟按操作计数加权；无样本或不可计算不表示零值，也不表示健康。"/>
    <Button disabled={!clusterId||scoped?.busy} loading={scoped?.busy} onClick={()=>void read()}>读取性能概览</Button>
    <Button disabled={!clusterId||scoped?.busy} loading={scoped?.busy} onClick={()=>void read(true)}>读取一小时趋势</Button>
    {!clusterId&&<Alert type="info" message="请先选择集群"/>}
    {scoped?.rows&&<><span>查询时间：{scoped.time}</span>{scoped.history?<>
      <Alert type="info" message="60 秒评估步长；缺失点和不可计算值处断线，不补零。图形使用浮点近似，各图独立纵轴；展开查看原始数值。"/>
      {scoped.rows.map(row=><Card key={row.id} size="small" title={`${row.name} · ${row.unit}`}>
        <span>{row.status}</span>
        {row.points?.some(point=>point.value!==undefined)&&<svg viewBox="0 0 800 150" width="100%" role="img" aria-label={`${row.name}一小时趋势`}>
          <line x1="40" y1="110" x2="740" y2="110" stroke="currentColor"/>
          <path d={overviewHistoryPath(row.points,scoped.start!,scoped.end!)} fill="none" stroke="#1677ff" strokeWidth="2"/>
          {row.points.filter(point=>point.value!==undefined).map(point=><circle key={point.time} cx={40+(point.time-scoped.start!)/3600*700} cy={110-point.value!/Math.max(1,...row.points!.map(p=>p.value??0))*80} r="2" fill="#1677ff"><title>{point.time}: {point.raw} {row.unit}</title></circle>)}
          <text x="40" y="135" fill="currentColor" fontSize="12">{new Date(scoped.start!*1000).toISOString()}</text><text x="740" y="135" textAnchor="end" fill="currentColor" fontSize="12">{new Date(scoped.end!*1000).toISOString()}</text>
          <text x="40" y="20" fill="currentColor" fontSize="12">纵轴上限：{Math.max(1,...row.points.map(point=>point.value??0))} {row.unit}</text>
        </svg>}
        {row.points&&<details><summary>原始评估样本（{row.points.length}）</summary><Table size="small" rowKey="time" pagination={{pageSize:15}} dataSource={row.points} columns={[{title:'评估时间（Unix 秒）',dataIndex:'time'},{title:'原始数值',dataIndex:'raw'}]}/></details>}
      </Card>)}
    </>:<Table size="small" rowKey="id" pagination={false} dataSource={scoped.rows} columns={[{title:'指标',dataIndex:'name'},{title:'值',dataIndex:'value',render:(value?:string)=>value??'—'},{title:'单位',dataIndex:'unit'},{title:'状态',dataIndex:'status'},{title:'评估时间（Unix 秒）',dataIndex:'timestamp',render:(value?:number)=>value===undefined?'—':String(value)}]}/>}</>}
  </Space></Card>
}

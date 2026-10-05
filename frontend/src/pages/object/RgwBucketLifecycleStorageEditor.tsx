import { Alert, Button, Form, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { useClusterContext } from '../../state/ClusterContext'
import { RgwBucketLifecycleEditor } from './RgwBucketLifecycleEditor'
import type { LifecycleDraft } from './rgwBucketLifecycleForm'
import { readLifecycleStorageClasses, type LifecycleStorageClasses } from './rgwLifecycleStorageClasses'

export function RgwBucketLifecycleStorageEditor(props:{value?:unknown;onChange?:(value:LifecycleDraft)=>void;disabled?:boolean;id?:string}) {
  const {selectedClusterId}=useClusterContext()
  const bucketId=Form.useWatch('bucket_id')
  return <RgwLifecycleStorageEditorScope {...props} clusterId={selectedClusterId} bucketId={typeof bucketId==='string'?bucketId:undefined}/>
}

export function RgwLifecycleStorageEditorScope({clusterId,bucketId,...props}:{clusterId?:number;bucketId?:string;value?:unknown;onChange?:(value:LifecycleDraft)=>void;disabled?:boolean;id?:string}) {
  const scope=JSON.stringify([clusterId,bucketId]),current=useRef(scope),sequence=useRef(0),abort=useRef<AbortController>()
  current.current=scope
  const [state,setState]=useState<{scope:string;busy:boolean;error?:boolean;data?:LifecycleStorageClasses}>({scope,busy:false})
  useEffect(()=>{abort.current?.abort();sequence.current++;setState({scope,busy:false});return()=>{abort.current?.abort();sequence.current++}},[scope])
  const scoped=state.scope===scope,data=scoped?state.data:undefined
  async function read() {
    if(!clusterId||!bucketId||props.disabled||current.current!==scope)return
    abort.current?.abort();const controller=new AbortController();abort.current=controller
    const ticket=++sequence.current;setState({scope,busy:true})
    try {
      const data=await readLifecycleStorageClasses(clusterId,bucketId,controller.signal)
      if(!controller.signal.aborted&&current.current===scope&&sequence.current===ticket)setState({scope,busy:false,data})
    } catch {
      if(!controller.signal.aborted&&current.current===scope&&sequence.current===ticket)setState({scope,busy:false,error:true})
    }
  }
  return <Space direction="vertical" style={{width:'100%'}}>
    <Alert type="info" message="转换类库存候选" description="按桶原生 Zonegroup ID 和放置目标关联已声明的非 STANDARD 类。不使用其他组的同名类，不自动填写。库存关联不证明池、远端或生命周期执行可用，也不是实时原子校验；手工输入用于保留或明确填写原生类名，提交前需核对配置。"/>
    <Button htmlType="button" disabled={props.disabled||!clusterId||!bucketId||!scoped||state.busy} loading={scoped&&state.busy} onClick={()=>void read()}>读取此桶的转换类候选</Button>
    {scoped&&state.error&&<Alert type="warning" message="候选读取失败、库存过期或桶放置身份无法确认；未回填任何类名，现有规则保持不变"/>}
    {data&&<Alert type={data.options.length?'info':'warning'} message={`组 ${data.groupId} / 目标 ${data.placement}；桶采集时间 ${data.observed}`} description={data.options.length?`可选 ${data.options.length} 个已声明类。选择候选只修改当前动作的 StorageClass，不更改组或桶放置。`:'该目标没有可确认的非 STANDARD 转换候选；不能据此认为任意手填值有效。'}/>}
    <RgwBucketLifecycleEditor {...props} onChange={draft=>{if(current.current===scope&&!props.disabled)props.onChange?.(draft)}} storageClassOptions={data?.options}/>
  </Space>
}

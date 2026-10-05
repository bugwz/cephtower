import { Alert, Button, Space, Table } from 'antd'
import { useEffect, useRef, useState } from 'react'
import type { ApiRecord } from '../../api/client'
import { localClassMappings, readLocalClassZones } from './rgwLocalClassMappings'

export function RgwLocalClassDetails({row,clusterId}:{row:ApiRecord;clusterId?:number}) {
  const scope=JSON.stringify([clusterId,row]),current=useRef(scope),sequence=useRef(0),locked=useRef(false)
  current.current=scope
  const abort=useRef<AbortController>(),mounted=useRef(true)
  const [state,setState]=useState<{scope:string;busy:boolean;error?:boolean;data?:ReturnType<typeof localClassMappings>;stale?:boolean}>({scope,busy:false})
  useEffect(()=>{mounted.current=true;abort.current?.abort();sequence.current++;locked.current=false;setState({scope,busy:false});return()=>{mounted.current=false;abort.current?.abort();sequence.current++}},[scope])
  const scoped=state.scope===scope
  async function read() {
    if(!clusterId||!scoped||current.current!==scope||!mounted.current||locked.current)return
    locked.current=true;const ticket=++sequence.current
    abort.current?.abort();const controller=new AbortController();abort.current=controller
    setState({scope,busy:true})
    try {
      const zones=await readLocalClassZones(clusterId,controller.signal)
      if(controller.signal.aborted||current.current!==scope||sequence.current!==ticket)return
      const data=localClassMappings(row,zones.rows)
      setState({scope,busy:false,data,stale:zones.stale||row.stale===true})
    } catch {
      if(!controller.signal.aborted&&current.current===scope&&sequence.current===ticket)setState({scope,busy:false,error:true})
    } finally {if(current.current===scope&&sequence.current===ticket)locked.current=false}
  }
  return <Space direction="vertical" style={{width:'100%'}}>
    <Alert type="info" message="本地存储类的成员 Zone 池映射" description="按组成员 ID、Realm、目标和类名关联所有成员，不选取任意第一个匹配项。读取已采集库存，不执行原生命令刷新；不同采集时间不是原子快照，匹配不证明数据可访问。云分层不使用此本地池关联。"/>
    <Button disabled={!clusterId||!scoped||state.busy} loading={scoped&&state.busy} onClick={()=>void read()}>读取成员 Zone 池映射</Button>
    {scoped&&state.error&&<Alert type="error" message="读取失败、分页不完整或身份异常，未将失败视为空配置；请重试并检查采集状态"/>}
    {scoped&&state.data&&<>
      {state.stale&&<Alert type="warning" message="组或 Zone 库存过期，不能据此确认当前配置"/>}
      {state.data.issues.length>0&&<Alert type="warning" message={state.data.issues.join('；')}/>}
      <Table size="small" rowKey="key" dataSource={state.data.rows} pagination={{pageSize:10}} scroll={{x:1000}} locale={{emptyText:'本次库存无可关联的本地类与成员组合，不代表已验证没有配置'}} columns={[
        {title:'放置目标',dataIndex:'placement'},{title:'存储类',dataIndex:'storageClass'},{title:'成员 Zone',dataIndex:'zone'},{title:'Zone ID',dataIndex:'zoneId'},
        {title:'关联状态',dataIndex:'status'},{title:'数据池',dataIndex:'pool'},{title:'压缩',dataIndex:'compression'},{title:'Zone 采集时间',dataIndex:'observed'}
      ]}/>
    </>}
  </Space>
}

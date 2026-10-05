import { Alert, Button, Checkbox, Input, Select, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { mutateResource } from '../../api/resource'
import type { Configuration } from './RgwEncryptionPage'

export function RgwEncryptionEditor({configuration,clusterId,labels,isCurrent}:{configuration:Configuration;clusterId:number;labels:Record<string,string>;isCurrent:()=>boolean}) {
  const [values,setValues]=useState<Record<string,string|boolean>>({})
  const [confirmed,setConfirmed]=useState(false),[saved,setSaved]=useState(false)
  const [status,setStatus]=useState<'editing'|'running'|'done'|'error'>('editing')
  const mounted=useRef(true),submitted=useRef(false),revision=useRef(0)
  // Effect replay does not change the draft; only edits and confirmations advance it.
  useEffect(()=>{mounted.current=true;return()=>{mounted.current=false}},[])
  const ticket=revision.current
  const editable=()=>mounted.current&&isCurrent()&&!submitted.current&&ticket===revision.current
  function change(name:string,value:string|boolean|undefined) {
    if(!editable())return
    const next={...values};if(value===undefined)delete next[name];else next[name]=value
    revision.current++;setValues(next);setConfirmed(false);setSaved(false)
  }
  async function submit() {
    if(!editable()||!confirmed||!Object.keys(values).length||('password' in values&&!saved))return
    submitted.current=true;setStatus('running')
    const patch={...values};setValues({});setConfirmed(false);setSaved(false)
    try {
      await mutateResource('/rgw/encryption/configuration','PATCH',{
        cluster_id:clusterId,entity:configuration.entity,encryption_type:configuration.encryption_type,
        provider:configuration.provider,expected_backend:configuration.backend,values:patch,
        confirm_disruption:true,confirm_credentials_saved:saved
      })
      if(mounted.current&&isCurrent())setStatus('done')
    } catch {if(mounted.current&&isCurrent())setStatus('error')}
  }
  if(status!=='editing')return <Alert type={status==='done'?'success':status==='error'?'error':'info'} message={status==='done'?'配置写入及回读核验完成；请重新读取配置':'running'===status?'正在提交并等待任务结果，请勿重复写入':'操作失败或结果未知，可能已有部分配置生效；请先核对任务和配置，不要直接重试'} description="任务不会自动重试。离开页面不取消已提交任务；配置核验不代表 RGW 进程已加载或密钥服务连接成功。"/>
  return <Space direction="vertical" style={{width:'100%'}}>
    <Alert type="warning" message="编辑提供商配置（高风险）" description="仅写入勾选字段；未选字段保持不变，不切换 Monitor 后端选择。空字符串会建立显式空值覆盖，不是删除覆盖。错误配置可能导致对象无法加解密；操作非事务，失败可能部分生效。密码不回填，所有文件字段均为 RGW 主机路径；命令参数可能被本机进程检查工具看到。"/>
    {configuration.fields.map(field=>{
      const selected=Object.prototype.hasOwnProperty.call(values,field.name)
      const options=field.name==='auth'?['token','agent']:field.name==='secret_engine'?(configuration.encryption_type==='s3'?['transit']:['transit','kv-v2']):undefined
      return <Space key={field.name} wrap>
        <Checkbox aria-label={`修改 ${field.name}`} checked={selected} onChange={event=>change(field.name,event.target.checked?(field.name==='verify_ssl'?false:options?.[0]??''):undefined)}>{labels[field.name]}</Checkbox>
        {selected&&(field.name==='verify_ssl'?<Select aria-label={field.name} value={String(values[field.name])} options={[{value:'true',label:'true'},{value:'false',label:'false'}]} onChange={value=>change(field.name,value==='true')}/>:options?<Select aria-label={field.name} value={String(values[field.name])} options={options.map(value=>({value,label:value}))} onChange={value=>change(field.name,value)}/>:field.name==='password'?<Input.Password aria-label={field.name} autoComplete="new-password" value={String(values[field.name])} onChange={event=>change(field.name,event.target.value)} maxLength={4096}/>:<Input aria-label={field.name} value={String(values[field.name])} placeholder="空字符串将显式写入" onChange={event=>change(field.name,event.target.value)} maxLength={4096}/>)}
      </Space>
    })}
    <Checkbox checked={confirmed} onChange={event=>{if(editable()){revision.current++;setConfirmed(event.target.checked)}}}>我已备份配置，确认修改 {configuration.entity} 的 {configuration.encryption_type}/{configuration.provider}，理解加解密中断及部分生效风险</Checkbox>
    {'password' in values&&<Checkbox checked={saved} onChange={event=>{if(editable()){revision.current++;setSaved(event.target.checked)}}}>我已安全保存凭据，确认写入或清空密码</Checkbox>}
    <Button danger type="primary" disabled={!confirmed||!Object.keys(values).length||('password' in values&&!saved)} onClick={()=>void submit()}>提交配置补丁</Button>
  </Space>
}

import { Alert, Button, Card, Descriptions, Input, Select, Space, Table } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { jsonInit, request } from '../../api/client'
import { useClusterContext } from '../../state/ClusterContext'

const labels:Record<string,string>={addr:'服务地址',auth:'认证方式',prefix:'路径前缀',secret_engine:'Secret Engine',namespace:'命名空间',token_file:'Token 文件路径',ssl_cacert:'CA 证书路径',ssl_clientcert:'客户端证书路径',ssl_clientkey:'客户端私钥路径',verify_ssl:'校验 TLS 证书',username:'用户名',password:'密码',client_cert:'客户端证书路径',client_key:'客户端私钥路径',ca_path:'CA 路径',kms_key_template:'KMS Key 模板',s3_key_template:'S3 Key 模板',key_template:'SSE-S3 Key 模板'}
type Configuration={entity:string;encryption_type:string;provider:string;backend:string;observed_at:string;fields:Array<{name:string;option:string;value:string;redacted:boolean}>}
export function encryptionConfiguration(value:unknown,entity:string,profile:string):Configuration {
  if(!['kms/vault','kms/kmip','s3/vault'].includes(profile))throw new Error('Invalid encryption profile')
  const data=value as Configuration,[encryptionType,provider]=profile.split('/')
  if(!data||typeof data!=='object'||data.entity!==entity||data.encryption_type!==encryptionType||data.provider!==provider||typeof data.backend!=='string'||typeof data.observed_at!=='string'||!data.observed_at||!Array.isArray(data.fields))throw new Error('Invalid configuration identity')
  const names=provider==='kmip'?['addr','username','password','client_cert','client_key','ca_path','kms_key_template','s3_key_template']:['addr','auth','prefix','secret_engine','namespace','token_file','ssl_cacert','ssl_clientcert','ssl_clientkey','verify_ssl']
  if(encryptionType==='s3')names.push('key_template')
  const prefix=encryptionType==='s3'?'rgw_crypt_sse_s3_vault_':`rgw_crypt_${provider}_`
  if(data.fields.length!==names.length||new Set(data.fields.map(field=>field?.name)).size!==names.length||data.fields.some(field=>!field||!names.includes(field.name)||field.option!==(field.name==='key_template'?'rgw_crypt_sse_s3_key_template':prefix+field.name)||typeof field.value!=='string'||typeof field.redacted!=='boolean'||field.name==='password'&&(!field.redacted||field.value!=='[REDACTED]')))throw new Error('Incomplete configuration')
  return {entity:data.entity,encryption_type:data.encryption_type,provider:data.provider,backend:data.backend,observed_at:data.observed_at,fields:data.fields.map(({name,option,value,redacted})=>({name,option,value,redacted}))}
}

export function RgwEncryptionPage() {
  const {selectedClusterId}=useClusterContext()
  return <RgwEncryptionView key={selectedClusterId??'none'} clusterId={selectedClusterId}/>
}

export function RgwEncryptionView({clusterId}:{clusterId?:number}) {
  const [entity,setEntity]=useState(''),[profile,setProfile]=useState('kms/vault')
  const scope=JSON.stringify([clusterId,entity,profile]),current=useRef(scope),abort=useRef<AbortController>(),sequence=useRef(0)
  current.current=scope
  const mounted=useRef(true)
  const [state,setState]=useState<{scope:string;busy:boolean;data?:Configuration;error?:boolean}>({scope,busy:false})
  useEffect(()=>{mounted.current=true;abort.current?.abort();sequence.current++;setState({scope,busy:false});return()=>{mounted.current=false;abort.current?.abort();sequence.current++}},[scope])
  const data=state.scope===scope?state.data:undefined,busy=state.scope===scope&&state.busy
  const valid=!!clusterId&&/^client\.rgw\.[A-Za-z0-9][A-Za-z0-9_.-]{0,255}$/.test(entity)
  async function read() {
    if(!valid||current.current!==scope||!mounted.current)return
    abort.current?.abort();const controller=new AbortController();abort.current=controller;const ticket=++sequence.current
    setState({scope,busy:true})
    try {
      const [encryption_type,provider]=profile.split('/')
      const response=await request<unknown>('/rgw/encryption/configuration',jsonInit('GET',{cluster_id:clusterId,entity,encryption_type,provider},{signal:controller.signal,cache:'no-store',suppressErrorNotification:true}))
      if(!controller.signal.aborted&&current.current===scope&&ticket===sequence.current)setState({scope,busy:false,data:encryptionConfiguration(response,entity,profile)})
    } catch {if(!controller.signal.aborted&&current.current===scope&&ticket===sequence.current)setState({scope,busy:false,error:true})}
  }
  return <Card title="RGW 服务端加密配置"><Space direction="vertical" style={{width:'100%'}}>
    <Alert type="info" message="读取 Monitor 解析后的配置，不是 Bucket 默认加密策略" description="按明确的 client.rgw.<id> 实体执行 config get，包含继承值和原生默认值。配置实体可以不存在运行中的进程；结果不证明 RGW 已加载配置、证书文件存在或密钥服务可达。多项读取不是原子快照。"/>
    <Space wrap>
      <Input aria-label="RGW 配置实体" placeholder="client.rgw.<id>" value={entity} onChange={event=>setEntity(event.target.value)} style={{width:360}}/>
      <Select aria-label="加密提供商配置" value={profile} onChange={setProfile} options={[{value:'kms/vault',label:'SSE-KMS / Vault'},{value:'kms/kmip',label:'SSE-KMS / KMIP'},{value:'s3/vault',label:'SSE-S3 / Vault'}]} style={{width:220}}/>
      <Button disabled={!valid||busy} loading={busy} onClick={()=>void read()}>读取配置</Button>
    </Space>
    {!clusterId&&<Alert type="info" message="请先选择集群"/>}
    {state.scope===scope&&state.error&&<Alert type="error" message="配置读取失败、字段不完整或后端选择发生变化；未展示部分结果，请检查权限与实体后重试"/>}
    {data&&<>
      <Descriptions column={1} items={[{key:'entity',label:'配置实体',children:data.entity},{key:'backend',label:'Monitor 后端选择值',children:JSON.stringify(data.backend)},{key:'time',label:'读取时间',children:data.observed_at}]}/>
      <Alert type={data.backend===data.provider?'info':'warning'} message={data.backend===data.provider?'Monitor 后端选择与所查看提供商一致（不代表运行时验证）':'所查看提供商不是 Monitor 当前选择值；以下仅为该提供商配置'}/>
      <Table size="small" rowKey="name" dataSource={data.fields} pagination={false} scroll={{x:800}} columns={[{title:'配置项',render:(_value:unknown,row:Configuration['fields'][number])=>labels[row.name]},{title:'原生选项',dataIndex:'option'},{title:'值',render:(_value:unknown,row:Configuration['fields'][number])=><span style={{whiteSpace:'pre-wrap'}}>{row.redacted?(row.value==='[REDACTED]'?'已隐藏':`${JSON.stringify(row.value)}（敏感部分已隐藏）`):JSON.stringify(row.value)}</span>}]}/>
      <Alert type="info" message="文件字段是 RGW 主机上的路径，不是文件内容。密码始终隐藏，空字符串按原值展示；此页面尚不提供修改操作。"/>
    </>}
  </Space></Card>
}

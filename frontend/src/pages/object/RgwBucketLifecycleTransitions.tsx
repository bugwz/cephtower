import { Table } from 'antd'
import { bucketLifecycleRows, lifecycleSelectorText } from './RgwBucketLifecycleRules'

export function bucketLifecycleTransitions(value:unknown) {
  const rules=bucketLifecycleRows(value)
  if(!rules)return undefined
  return rules.flatMap(rule=>rule.actions.flatMap((action,index)=>{
    if(action.type!=='Transition'&&action.type!=='NoncurrentVersionTransition')return []
    const fields=action.fields,current=action.type==='Transition'
    const has=(key:string)=>Object.prototype.hasOwnProperty.call(fields,key)
    const valid=current
      ? has('Days')!==has('Date')&&Object.keys(fields).every(key=>['Days','Date','StorageClass'].includes(key))
      : has('NoncurrentDays')&&Object.keys(fields).every(key=>['NoncurrentDays','StorageClass'].includes(key))
    return [{key:JSON.stringify([rule.index,index]),ruleIndex:rule.index+1,id:rule.id,status:rule.status,
      version:current?'当前版本':'非当前版本',storageClass:has('StorageClass')?JSON.stringify(fields.StorageClass):'未返回',
      timing:valid?(current?(has('Days')?`创建后天数：${JSON.stringify(fields.Days)}`:`UTC 日期：${JSON.stringify(fields.Date)}`):`成为非当前版本后的天数：${JSON.stringify(fields.NoncurrentDays)}`):'转换字段组合异常，请查看完整规则',
      selector:lifecycleSelectorText(rule.selector)}]
  }))
}

export function RgwBucketLifecycleTransitions({value,configured}:{value:unknown;configured:unknown}) {
  const rules=bucketLifecycleRows(value),rows=bucketLifecycleTransitions(value)
  if(!rules||!rows||typeof configured!=='boolean'||configured!==!!rules.length)return <span>生命周期分层数据不可用</span>
  if(!configured)return <span>未配置生命周期</span>
  if(!rows.length)return <span>已配置生命周期，但没有当前或非当前版本转换动作</span>
  type Row=typeof rows[number]
  return <Table size="small" rowKey="key" dataSource={rows} pagination={{pageSize:10,showSizeChanger:true}} scroll={{x:950}}
    locale={{emptyText:'没有匹配筛选的转换动作'}} columns={[
      {title:'规则序号',dataIndex:'ruleIndex'},
      {title:'规则 ID',dataIndex:'id',render:(id:string)=>JSON.stringify(id)},
      {title:'状态',dataIndex:'status',render:(status:string)=>status==='Enabled'?'启用':'禁用',filters:[{text:'启用',value:'Enabled'},{text:'禁用',value:'Disabled'}],onFilter:(value,row:Row)=>row.status===value},
      {title:'版本范围',dataIndex:'version',filters:['当前版本','非当前版本'].map(value=>({text:value,value})),onFilter:(value,row:Row)=>row.version===value},
      {title:'目标存储类（原生值）',dataIndex:'storageClass',filterSearch:true,filters:[...new Set(rows.map(row=>row.storageClass))].map(value=>({text:value,value})),onFilter:(value,row:Row)=>row.storageClass===value},
      {title:'转换时间条件',dataIndex:'timing'},
      {title:'过滤条件',dataIndex:'selector',render:(value:string)=><span style={{whiteSpace:'pre-wrap'}}>{value}</span>}
    ]}/>
}

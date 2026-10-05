import { Alert, Descriptions, Table } from 'antd'
import { groupPlacementClassRows, zonePlacementClassRows, type PlacementClassRow } from './rgwPlacementClassData'

export function renderPlacementClassDetails(row:PlacementClassRow) {
  return <div>
    <Descriptions column={1} size="small" items={row.fields.map(f=>({key:f.name,label:f.name,children:f.value}))}/>
    {row.acl&&<section aria-label="云分层 ACL 映射">
      <h4>ACL 身份映射</h4>
      <p>按原生类型分组排序，来源身份映射到远端目标身份；不是当前对象的访问权限列表，不证明远端身份存在或授权生效。URI 仅作为文本展示。</p>
      {row.acl.issues.length>0&&<Alert type="warning" message={row.acl.issues.join('；')}/>}
      <Table size="small" rowKey="key" dataSource={row.acl.rows} pagination={row.acl.rows.length>8?{pageSize:8}:false} scroll={{x:600}}
        locale={{emptyText:row.acl.available&&!row.acl.issues.length?'本次配置中没有显式 ACL 映射':'暂无可可靠展示的 ACL 映射'}}
        columns={[{title:'类型',dataIndex:'type',filters:[...new Set(row.acl.rows.map(a=>a.type))].map(type=>({text:type,value:type})),onFilter:(value,item)=>item.type===value},{title:'原生映射键',dataIndex:'mappingKey'},{title:'来源身份',dataIndex:'source'},{title:'远端目标身份',dataIndex:'destination'}]}/>
    </section>}
  </div>
}

export function RgwPlacementClasses({value,zone=false}:{value:unknown;zone?:boolean}) {
  const {rows,issues}=zone?zonePlacementClassRows(value):groupPlacementClassRows(value)
  const field=(row:PlacementClassRow,name:string)=>row.fields.find(f=>f.name===name)?.value??(['本地','本地映射'].includes(row.type)?'不适用':'未返回或不可用')
  return <div>
    <p>{zone?'Zone 池映射快照，不证明组已声明此类、池存在或数据可访问。':'Zonegroup 类声明与分层配置快照，包含 STANDARD；不证明各 Zone 已配置池或远端目标可达。'} 仅展示已识别字段，凭据与未知嵌套字段不展示；请在组/Zone 操作菜单中管理配置。</p>
    {issues.length>0&&<Alert type="warning" message="部分配置无法确定，不能将缺失视为空配置" description={issues.join('；')}/>}
    <Table size="small" rowKey="key" dataSource={rows} pagination={rows.length>8?{pageSize:8}:false} scroll={{x:700}}
      locale={{emptyText:issues.length?'暂无可可靠展示的记录':'本次配置中无存储类记录'}}
      columns={[
        {title:'放置目标',dataIndex:'placement'}, {title:'存储类',dataIndex:'storageClass'}, {title:'类型',dataIndex:'type'},
        {title:'声明状态',dataIndex:'declared'},
        ...(zone?[{title:'数据池',render:(_value:unknown,row:PlacementClassRow)=>field(row,'数据池（原生引用）')},{title:'压缩',render:(_value:unknown,row:PlacementClassRow)=>field(row,'压缩配置')}]:[{title:'目标区域',render:(_value:unknown,row:PlacementClassRow)=>field(row,'目标区域')},{title:'目标端点',render:(_value:unknown,row:PlacementClassRow)=>field(row,'目标端点')}])
      ]}
      expandable={{expandedRowRender:renderPlacementClassDetails}}
    />
  </div>
}

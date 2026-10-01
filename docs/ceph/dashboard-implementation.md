# Ceph Dashboard 功能迁移与验证记录

本任务目标是以本地参考源码为依据，实现 CephTower 的展示与操作。当前仍在实施，
**不是全部功能已完成的验收报告**。参考目录不作为运行时依赖；不连接或部署测试集群。

## 如何追踪调用链

池编辑入口和提交同时拒绝过期/未知库存、未知池类型、缺失名称或有效资源版本，
加载中及读取失败时不允许编辑。PATCH 不再以 0 补造 If-Match 版本；nodelete 仅限制删除，
不错误阻断编辑。离线测试覆盖上述阻断条件，后端仍负责最终版本和权限校验。

池列表增加库存时效警告，同时检查列表汇总和各池 stale 标志；任一过期即提示历史结果，
缺失或异常标志提示时效未知。明确容量、PG 状态及配置并非实时值，保留采集时间和原因。
离线测试覆盖全新鲜、空列表、汇总过期、混合行及未知标志。

自动伸缩建议补齐 logical_used、raw_used_rate、actual_capacity_ratio、capacity_ratio。
按原生计算保留浮点值：逻辑用量由实际原始占用除以冗余开销得到；计算容量占比还会
考虑目标字节数，不与实际容量占比混淆。前端分别标明字节、倍数和比值，缺失值不补零。

新增独立自动伸缩建议区块：参考 Dashboard 池属性之外，以 Ceph pg_autoscaler/module.py
的 osd pool autoscale-status --format json 获取建议 PG 数、would_adjust、目标容量、
CRUSH 子树容量、目标/有效比例、偏置和 bulk 标志，按 pool_id 关联至 autoscale_status。
此区块明确为采集时建议，不表示已执行调整，不触发写操作。可选命令不可用时为 null；
零值和 false 保留，缺失值不补默认。离线测试覆盖命令参数、异常响应、跨池隔离和 API 序列化。

池详情展示当前/目标 PG 和 PGP 数量。osd pool ls detail 原生输出的 pg_num_target、
pg_placement_num_target 经采集映射为库存 API 的 pg_num_target、pgp_num_target；
缺失目标不以当前值补齐。与参考池详情一致区分调整中的当前值和目标值。
离线测试覆盖四个数量不同及缺失目标的序列化，前端验证四字段绑定。

池类型按原生 osd_types.h 的 TYPE_REPLICATED=1、TYPE_ERASURE=3 映射；缺失或其他值
保留 unknown，不再默认为副本池。前端列表/详情显示未知，编辑入口及表单初始化拒绝
未知类型，防止按副本池参数编辑未识别的池。离线测试覆盖原生映射及前端未知值处理。

池详情概览补齐参考表单的压缩算法、最小/最大 Blob 大小、所需压缩比。
数据来自现有 osd pool ls detail 的 options 采集，经 Pool 库存字段透传；不补默认值。
所需压缩比明确标记配置阈值，保留原始 0..1 数值，不冒充实测压缩率。
离线测试验证字段绑定及比例的零值、正常值、缺失和异常值。

池详情严格区分配额零值和未采集：原生 quota_max_bytes/quota_max_objects 为零时
显示无限制，缺失或异常值显示未采集；不再使用旧字段别名补值。压缩模式缺失不再
冒充 none。离线展示测试覆盖零值、正常配额及无效值。

池重命名沿用参考 pool.py 的 osd pool rename 调用，并以 osd pool ls --format json
回读核验新名称存在、旧名称消失。未知、重复或异常名称列表不视为成功；命令已接受
但无法核验时禁止自动重试。离线执行链测试覆盖成功及旧名称残留等失败结果，未连接集群。

列表和详情的数据保护展示不再把缺失 size 补成三副本；仅使用采集的正整数副本数。
纠删码池不将分片总数当作副本数，未知池类型及缺失副本数分别明确提示。
离线测试覆盖实际两副本、纠删码及异常/缺失值。

池详情增加累计读写操作次数，使用 df detail 的 rd/wr，经库存 API 映射为
read_operations/write_operations。与累计字节数分别展示，不转换或标注为 IOPS；
复用整数显示校验，保留零值并将缺失值标记为未采集。

存储池详情增加库存时效警告：stale=true 明确提示历史采集结果，缺失或异常时效
标志提示未知；只有明确 stale=false 才不展示警告。采集时间仍保留在概览中，
离线测试覆盖过期、新鲜和未知三类状态，防止把历史 PG/容量统计误当实时结果。

池删除入口识别采集的 nodelete 标志并展示删除保护原因，不打开确认框、不提交操作。
该提示不替代 Ceph 的实时保护检查，也不会自动清除保护标志。离线测试覆盖混合标志
及无 nodelete 情况。

池删除执行后使用 osd pool ls --format json 核验目标名称确已消失，拒绝目标仍在、
异常 JSON 或未知结果。命令已接受但回读未确认时标记不可自动重试，避免盲目重复
高风险删除。离线测试覆盖成功缺席和异常/仍存在结果，未执行真实删除。

存储池列表补齐参考删除入口，接入已有 DELETE /pool 高风险操作链路（原生命令
osd pool rm）。确认框明确全部数据永久丢失；过期/未知库存、缺失版本禁止提交，
发送 If-Match，确认前切换集群会拒绝旧操作。不会自动启用 mon_allow_pool_delete。
离线测试覆盖阻断条件、版本和集群参数及切换后拒绝提交；未执行真实删除。

池读写字节数接通 df detail 的 rd_bytes/wr_bytes（原生已将 KiB 计数转换为字节），
以 read_bytes/write_bytes 进入库存 API。列表和详情标为累计读取量/累计写入量，
不冒充吞吐率；保留零值和未知值。离线库存序列化测试覆盖非零读取与零写入。

池详情增加压缩数据实际占用及原始大小，复用 df detail 的 compress_bytes_used
（data_compressed_allocated）和 compress_under_bytes（data_compressed_original）。
两者不是全部池容量，也不把实际分配空间误称为纯压缩后数据长度；未知值保持未采集。

存储池详情手动采集增加集群及池名称作用域检查，切换资源后旧采集完成不再调用
旧详情加载器，也不显示误导性的刷新成功提示。原集群已提交的采集仍继续执行。
离线延迟回调测试覆盖当前资源和已切换资源两种完成路径。

存储池详情概览复用列表统计格式，展示使用率、用户数据量、实际占用、最大可用、
对象数量及纠删码配置。移除详情页残留的固定 active+clean 和默认 PG 自动伸缩
状态，PG 状态直接读取原生计数。离线检查覆盖详情绑定与禁止虚构健康状态。

存储池对象数量从 df detail 的 pools[].stats.objects 经库存 API 进入列表新增列。
保留真实零值，缺失、非法或超出前端安全整数范围的值显示未采集，不从容量推算。

存储池列表新增用户数据量（stored）、实际占用（bytes_used）和最大可用（max_avail）
三列，直接读取已接通的 df detail 库存字段，不以副本数反推容量。按二进制单位显示，
原生零值显示 0 B，缺失/非法值显示未采集；离线测试覆盖字段绑定、单位及零值。

存储池容量接入 ceph df detail --format json 的 pools[].stats，按数值池 ID 关联。
原生 percent_used 为 0–1 比例，转换为 API used_percent 的 0–100 百分比；同时保留
stored、bytes_used、max_avail 字节值。缺失使用率展示未采集，而不是虚构 0%；
重复池 ID、非法比例或负容量不发布统计。容量字段可由 API 和资源详情读取。

池的 crush_rule 从原生 pg_pool_t::dump 到库存 API 保持整数 ID，不再转成字符串。
这使 CRUSH 规则引用视图能够按 rule_id 正确匹配，且避免数字规则名称与 ID 混淆。
缺失 ID 保持 null，零值作为有效规则 ID；离线测试核对库存 JSON 的数值类型。

池库存增加原生 erasure_code_profile 字段：从 osd pool ls detail 的 pg_pool_t::dump
输出经采集模型、库存 JSON 和现有 API 到达前端，供配置详情和引用匹配使用。
编辑表单不再将缺失名称补成 default；不可修改的名称缺失时显示未采集，创建时
仍要求显式选择配置。离线用例覆盖非默认配置透传、缺失值及非法字段类型。

修正存储池主列表固定显示 active+clean 的错误。参考 CephService.get_pool_list_with_stats
和 pool-list 的状态计数格式，界面仅格式化有效 pg_status 计数；缺失或异常明确显示
未采集，不再根据 pg_num 编造健康状态。后端 storage 采集增加一次
`ceph pg dump pgs_brief --format json`，按 pgid 的池 ID 汇总原生 state，写入 Pool
库存载荷的 pg_status，由现有列表 API 提供给前端。命令失败、缺失字段、非法 PG ID
或重复 PG 会舍弃本次全部状态计数；不回退为健康。离线测试覆盖原生命令参数、
多池隔离、混合/降级状态、异常输出及库存 JSON 序列化，尚未做真实集群验证。

存储池主列表改为读取全部筛选后分页，再由表格分页展示，与参考 PoolService.getList
的完整列表语义一致。跨页保留集群及筛选条件，合并过期状态和最早采集时间；后续页
失败时不将第一页当成完整结果。离线测试覆盖第二页数据、状态合并及失败传播。

存储池压缩比例说明修正为压缩后/原始大小的上限，而非比例下限。参考 pool-form
的 helperText 及 BlueStore::_do_alloc_write 的 result_len <= want_len 判定；说明同时
保留分配单元对齐和压缩头开销的限制，避免把目标比例承诺为最终磁盘节省比例。

CRUSH 规则详情及复制池表单新增存储池引用视图，按原生数值 rule_id 或已解析名称
匹配 pool.crush_rule，不使用数组位置。复用全分页、集群隔离及过期提示的引用查询；
结果仅反映采集库存，不能替代原生命令的依赖检查。离线匹配测试及构建通过。

存储池复制池表单增加所选 CRUSH 规则的可展开步骤说明，复用规则库存详情组件。
按名称读取当前集群已采集规则，原生步骤顺序、num=0 和额外步骤字段保持原义；
新建规则尚未采集时明确提示，不根据表单推测步骤。前端检查和构建通过，未做真实
浏览器或集群验证。

纠删码详情及存储池表单增加“使用此配置的存储池”，对应参考 Used by pools。
读取当前集群全部分页池库存，按原生 erasure_code_profile 精确匹配，独立于页面筛选。
过期/未知库存标注历史引用，不将空结果解释为可安全删除；切换配置忽略旧请求结果。
离线匹配和过期状态测试通过，未做真实浏览器或集群验证。

存储池表单选择纠删码配置后提供可展开详情，对应参考 pool-form 的配置说明入口。
直接使用当前集群已采集的 profile 原生字段，复用中文参数说明；刚创建但尚未采集
的配置显示等待详情，不用创建表单默认值伪装回读结果。前端检查与构建通过，未做
真实浏览器或集群验证。

存储池创建、编辑和手动采集完成后同样核对集群作用域，避免切换集群后的旧回调
关闭当前表单、显示误导性成功提示或重新加载旧集群。已提交的原集群编辑序列继续
执行，界面隔离不等于取消后端操作。离线测试覆盖创建和编辑的迟到完成回调。

存储池页面的 CRUSH/纠删码创建回调增加集群作用域标记；切换集群（包括切回原集群）
后旧操作完成不再写入候选、修改表单或触发旧加载器刷新。已提交后端操作不被取消，
可在操作记录查看结果。离线延迟回调测试覆盖两种创建链路，尚无真实浏览器验证。

切换到 LRC 插件不再把已选局部性强制改成 host；提交时核对非空局部性属于当前
根节点及设备类别，并允许清空可选局部性。防止没有 host 层级的拓扑发送虚构选项。
离线测试覆盖仅有 OSD 的根节点、合法局部性和省略参数；未做真实集群验证。

纠删码配置详情新增中文参数说明，覆盖参考模型中的 LRC l、SHEC c、CLAY d/scalar_mds、
Jerasure packetsize 及 CRUSH 局部性/故障域数量等全部已知字段。保留原生字符串和零值，
缺失参数显示未提供而非套用插件默认值；通用详情仍保留其他原生字段。
离线字段覆盖和零值测试通过，尚未做真实浏览器验证。

纠删码创建不再预填固定 /usr/lib64 路径，也不复制其他配置的 directory。该字段
保留为可选显式输入，留空不提交。参考 OSDMonitor::normalize_profile 使用集群
erasure_code_dir 加载插件，界面不能将任意已有配置字段当作集群目录配置来源。
前端回归检查和构建通过；动态插件信息读取仍待接入，未做真实集群验证。

CRUSH 规则和纠删码配置的写接口 OpenAPI 改为直接引用运行时请求契约，避免通用
字段并集掩盖重命名必填的 name/new_name，以及创建时的放置参数要求。生成文件
同步更新，测试逐个比对五个写接口的契约并检查缺失目标名称被拒绝。

CRUSH 规则面板新增重命名表单及确认提示，过期库存禁用并携带资源版本。修正旧 API
名称冲突：name 定位原规则，new_name 指定目标，原生命令 rename 后通过 ls 核验。
成功后重新采集规则库存。此入口扩展已有原生命令能力，不宣称参考 Dashboard 有同名
操作；未做真实浏览器或集群验证。

CRUSH 重命名后改用 rule ls 核验名称集合，确认新名称存在、旧名称消失（同名请求
只检查名称存在）。空、重复、错误类型或尾随 JSON 不再报告成功，失败提示先检查规则。
此项完善已有后端原生命令链路，尚未新增界面重命名入口；离线测试通过。

CRUSH 拓扑增加可暂停的 5 秒只读轮询，每次请求完成后再计时以避免重叠。自动刷新
同步已选节点详情；节点消失或读取失败时清除旧详情，失败不继续展示旧拓扑。切换集群、
手动刷新或卸载取消请求及计时器，忽略迟到响应。下方规则/配置库存仍独立刷新。
离线测试覆盖轮询时序及取消，未做真实浏览器或集群验证。

CRUSH rule dump 采集现在要求原生对象数组和唯一的 rule_id/rule_name，不再用数组
位置伪造缺失名称，也不把 null 或对象响应当作空库存。异常整批标记不可用，保留原
观测；合法空数组仍可清空库存。离线测试覆盖乱序 ID、重复身份、缺失字段与类型错误。

纠删码配置面板接入高风险删除确认，使用 DELETE /erasure/code/profile 与资源版本
前置条件；禁用过期/未知库存和缺失名称。后端执行 profile rm，再用 ls 验证消失，
前端成功后重新采集配置库存。原生 OSDMonitor 会拒绝删除被存储池使用的配置，不自动
解除引用或删除池；离线前端测试通过，未实际删除集群资源。

CRUSH 页面新增纠删码配置库存面板，展示名称、插件、k/m、算法、根节点、故障域和
设备类别；使用原生 profile ls/get 采集及 GET /erasure/code/profiles，沿用通用列表
分页、刷新、过期提示和详情。字段依据参考 erasure-code-profile 模型及表单，独立
库存面板是本项目的组织方式；创建仍在存储池表单，面板删除操作尚未接入。
前端字段和挂载检查、构建通过，未做真实浏览器或集群验证。

纠删码配置采集不再把 get 失败降级为只有名称的成功资源。原生 ls 必须返回非空值的
唯一名称数组，get 必须返回包含 plugin 的字符串键值表；任一详情失败则丢弃本批
配置观测并标记该类库存不可用，避免不完整快照替换已有数据。空数组仍是有效空库存。
新增离线测试覆盖失败详情、缺失插件、类型错误和重复名称，尚无真实集群验证。

存储池放置候选现在检查分页汇总和各行的 stale 状态，过期或未知库存显示加载错误。
纠删码配置缺少 plugin（例如详情采集失败仅保留名称）也不能作为完整候选使用。
页面刷新同时采集 pool、crush_rule 和 erasure_code_profile，提供恢复最新候选的入口。
离线测试覆盖新鲜空库存、过期汇总、过期/未知行及不完整配置，未做真实集群验证。

存储池表单保留 CRUSH 原生 rule_id/rule_name/type 元数据，按 ID 而非数组位置
解析已有池的规则，缺失映射保留原 ID，不臆造默认规则。复制池候选仅显示 type=1
的原生规则及当前页面成功创建的复制规则。参考 pool.py 的 ID 名称映射和类型筛选；
离线测试覆盖乱序、稀疏 ID、未知 ID 及数字名称，尚无真实集群验证。

存储池创建不再凭空加入 replicated_rule/default 放置配置候选；CRUSH 规则及纠删码
配置请求失败会显示加载错误，不再转为空列表成功。创建必须显式选择已读取或当前页面
成功创建的候选，并在提交前检查候选存在。离线测试覆盖无默认资源、有效候选和空选择；
规则类型过滤、库存新鲜度及真实浏览器验证仍需进一步核对。

纠删码配置删除在 `osd erasure-code-profile rm` 后读取原生 `ls --format json`，
严格校验名称数组并确认目标不存在。目标残留、空值、类型错误或尾随 JSON 均返回
不可自动重试的验证失败，避免仅凭退出码误报删除成功；与 CRUSH 删除共用名称缺席校验。
离线测试覆盖空列表、保留其他配置、残留目标及异常响应，未实际删除集群配置。

纠删码配置创建增加参数回读验证：参考 OSDMonitor 的 profile set/get 实现，
写入后解析原生字符串键值表，逐项核对实际发送的插件、编码和 CRUSH 放置参数，
允许 Ceph 补充默认字段。空、异常、尾随 JSON 或参数不一致不再报告成功，提示先检查
配置且不建议自动重试。仅通过离线命令与响应测试，未验证真实集群。

SMB 共享补充原生 `max_connections`：依据 `smb/resources.py` 的非负整数校验、
`smb/handler.py` 的 Samba 配置生成及 `doc/mgr/smb.rst` 的字段说明，提供创建、
编辑、列表展示和 API 整数契约。0 表示不限制连接数；创建省略使用原生默认值，
更新省略保留原配置。通过 `smb apply` 写入，再用 `smb show` 验证上限。
此项是原生能力扩展，并非已发现的 Dashboard 表单字段；仅完成离线测试，未做真实连接压测。

- 参考界面：`docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/ceph`。
- 参考 HTTP 服务：同目录的 `shared/api`；[完整静态索引](dashboard-source-index.md)。
- 参考接口实现：`dashboard/controllers`，底层 `dashboard/services`。
- Ceph CLI 契约：`src/mon/MonCommands.h`、各 mgr 模块的 CLI 装饰器与实现。
- 当前采集：`backend/internal/integration/ceph/collector*.go`。
- 当前读链路：reconciler → resource store → `handler.ReadResource` → `frontend/src/api/resource.ts`。
- 当前写链路：严格请求契约 → `handler.MutateResource` → `service/mutation` → executor。
- executor 接收参数数组，不拼 shell；现有集群凭证、权限、审计、超时体系继续使用。
- `integration/ceph/command` 有包装函数不等于 API 和前端已接通，必须逐层核对。

## 模块分析与后续核对范围

新增“集群升级”状态入口，对照参考 upgrade-progress 和 UpgradeStatusInterface，
使用现有 `orch upgrade status --format json` 采集、upgrade 库存及 GET /upgrade，
展示目标镜像、进行/暂停状态、范围、原生进度文本、已完成服务和消息。
页面支持重新读取与触发采集，展示采集时间和过期标记；未知状态不推断为未升级。
暂停/恢复/停止通过 POST /upgrade/action 接入异步操作及资源版本校验，执行后按
原生 in_progress/is_paused 回读核验，再采集状态。过期或不匹配的状态禁用按钮，
确认框明确停止不回滚已升级守护进程；打开确认框不会先执行暂停。
升级前检查允许输入目标版本，经 POST /upgrade/check → orch upgrade check
--ceph-version 返回结构化报告，展示目标镜像/版本/摘要、待升级守护进程及当前镜像、
已匹配列表和非 Ceph 镜像列表。修改版本或切换集群清除报告，不自动启动升级。
版本不兼容的原生文本即使退出码为 0 也作为检查失败。启动入口要求报告对应当前版本，
库存状态未过期且明确未在升级；确认服务影响后携带资源版本提交，后端再次检查兼容性，
回读运行状态及目标镜像/摘要。提交后清除启动资格并重新采集；成功只表示启动，不表示升级完成。
候选版本通过 GET /upgrade/versions → orch upgrade ls 按需加载，展示镜像与仓库，
支持选择或手动输入；保留原生命令顺序，不把候选列表当作兼容性证明。
升级检查和启动也支持自定义容器镜像，通过 --image 传入，和 version 二选一。
切换目标类型清除旧报告；启动前复查和启动命令使用同一镜像，仍核验回读目标。
升级页增加守护进程版本表，复用 orch ps --refresh → daemon 库存 → GET /daemons，
读取全部分页，按名称/主机/版本搜索，展示镜像、未知值、采集时间及过期状态。
页面重新采集同时刷新 upgrade 和 daemon 库存。
升级状态支持可暂停的自动读取，每次请求完成后等待 10 秒读取库存；不自动触发采集或写操作。
读取失败保留旧状态并标为过期，禁用升级控制；切换集群或离开页面取消请求和定时器。
升级页嵌入精简集群日志面板，复用 GET /logs → ceph log last 100 debug cluster，
每次读取完成后等待 10 秒刷新，可暂停或手动刷新；切换集群取消旧请求。
仅展示 MON 缓冲区近期日志，自动刷新失败会保留旧记录并提示，不代表长期归档。
尚未完成浏览器和真实集群验证。

新增 CRUSH 拓扑页面，对照参考 crushmap 组件及 CrushRuleUi.info：
GET /crush/map → ceph osd tree --format json，展示可展开节点和完整节点元数据。
后端从父子关系计算根节点，校验重复 ID、缺失引用与环路；前端支持刷新和节点选择，
空树与请求错误分开展示。只读即时查询，不修改 CRUSH 规则或权重；尚未完成浏览器及真实集群验证。
节点状态按参考语义使用 up/in 正常、down/out/destroyed 异常标签，其余状态保持中性；
树与元数据采用响应式双栏布局，节点通过显式 ID 选择，取消选择清除详情。
存储池表单依赖的 CRUSH 规则、纠删码配置读取全部分页；节点选项和故障域计数直接读取
GET /crush/map 原生树，包含空桶，不再从 OSD 库存反推路径或伪造 default 根选项。
计数严格遍历所选节点子树并筛选设备类别，default 不作为所有根的通配符；
拓扑请求失败显示错误，没有有效节点时禁止打开规则/纠删码配置创建表单。
复制规则和纠删码表单的设备类别分别按所选子树过滤；失效类别/故障域自动清除，
提交前再次检查节点、故障域及设备类别的有效性，空子树和零计数故障域不能提交。
复制规则创建后核验 rule dump 的名称、复制类型、take 根/设备类别、choose 故障域与 emit 步骤。
原生命令遇到同名规则可返回成功但不改变配置，因此回读不匹配时报告 post_check_failed，
要求检查现有规则，不将退出码为零直接视为配置成功。
CRUSH 页面增加规则库存列表与步骤详情，复用 osd crush rule dump → GET /crush/rules，
展示规则名、ID、类型、min_size/max_size，步骤保持顺序和原生参数（包含 0 与负节点 ID）。
复用资源列表的分页、筛选、重新采集及时间元数据；未知类型保留编号，缺失步骤独立提示。
规则列表支持确认后删除，携带资源版本，过期/未知库存禁用。后端执行 rule rm 后严格检查
rule ls 中目标已不存在，再报告成功；Ceph 原生拒绝删除被存储池使用的规则，不强制绕过。
删除成功后重新采集规则库存。

下表中的命令族用于定位实现，不是可直接执行的完整命令，也不保证所有 Ceph 版本支持。
“已有”仅表示本次开始时能找到相关代码，尚需逐项检查参数、响应与操作覆盖。

| 参考模块 | 数据获取方式 / 命令族 | 当前情况及待补齐范围 |
| --- | --- | --- |
| dashboard / health | `status`、`df detail`、`health detail`、`pg dump`、`osd df`、`osd dump` | 已有总览、PG 分布、对象统计、IOPS、恢复速率、scrub 状态、健康详情和静默 |
| cluster / host | `orch host ls --detail`、`orch ps`、`orch device ls`、`device ls-by-host` | 已有主机详情、SSH、SMART、服务操作；核对硬件和维护/排空全部参数 |
| cluster / monitor | `mon dump`、`quorum_status`、`tell mon.* perf dump` | 已有详情和性能计数；核对历史速率与时钟偏移 |
| cluster / mgr | `mgr dump`、`mgr module ls`、`config get/set` | 已有模块开关；模块配置项编辑与验证待核对 |
| cluster / osd | `osd dump/tree/df`、`osd metadata`、`tell osd.*`、`orch osd rm` | 已有状态、权重、scrub、部署/移除；详情、个体 flags、设备 class、histogram 待补 |
| pool / CRUSH / EC | `osd pool ls detail`、`osd pool get/set`、`osd crush rule`、`osd erasure-code-profile` | 已有管理表单；核对统计、配额、压缩、全部编辑参数与只读状态 |
| cluster / configuration | `config dump/help/set/rm` | 已补集群配置入口、全部选项浏览、按需元数据详情、设置和删除覆盖；修复范围键与字符串值语义 |
| cluster / ceph-user | `auth ls/add/caps/del/import/export` | CephX 实体与 CephTower 登录用户不同；已补独立列表、增删改、keyring 导入及所选实体导出，包含缓存脱敏与 API 链路测试 |
| cluster / logs | `log last` / 外部日志源 | 已接通 log last 直接读取、频道/级别/条数筛选、轮询与下载；移除仅发心跳的假流接口 |
| cluster / upgrade | `orch upgrade check/status/start/pause/resume/stop` | 已有状态和操作，版本选择及 daemon 范围参数待核对 |
| block / rbd | `rbd ls/info/create/resize/rm/feature`、`rbd snap`、`rbd trash`、`rbd namespace` | 已有基础管理；克隆/扁平化、配置、快照保护、任务、回收站完整操作需逐项补齐 |
| block / mirroring | `rbd mirror pool/image`、`rbd mirror snapshot schedule` | 已有池状态、模式、peer、bootstrap token、镜像操作与镜像级调度；池/集群级调度管理待补齐 |
| block / iSCSI | ceph-iscsi REST API | 需要网关 endpoint，不能把所有操作替换成普通 `ceph` CLI；当前已有外部客户端 |
| block / NVMe-oF | 网关 gRPC | 当前已有 gRPC 客户端；子系统、namespace、listener、host、连接与 QoS 待完整对照 |
| cephfs / filesystem | `fs dump/status/get/set`、`fs volume`、`tell mds.* client ls/evict/perf dump` | 已有文件系统详情、客户端、授权、卷重命名、MDS 计数器会话趋势、Rank/备用 MDS 及池容量；其余完整字段待核对 |
| cephfs / subvolume | `fs subvolumegroup`、`fs subvolume`、`fs subvolume snapshot`、`fs clone` | 已有组范围采集、clone 状态/进度/失败展示、快照克隆、进行中任务取消和快照可见性；metadata 与其余完整参数需核对 |
| cephfs / snapshot schedule | `fs snap-schedule` | 已有全路径发现、精确状态、创建、删除、激活/停用、retention 和模块启用 |
| cephfs / directory | libcephfs 或 CephFS 数据面客户端 | 已有实时目录浏览、目录元数据、双维度配额、目录增删/重命名/移动及目录快照列表/创建/删除 |
| nfs | `nfs cluster`、`nfs export` | 已有基础管理；完整 export 属性、CephFS/RGW FSAL 与 ingress 待核对 |
| smb | `smb show/apply/rm` 与模块资源定义 | 已有部分管理；域加入、用户组、资源校验与配置语义需核对 |
| rgw / user / account / role | `radosgw-admin user/account/role` 与 RGW Admin Ops | 当前存在基础页面；配额、subuser、caps、rate limit、角色策略等需逐项扩展 |
| rgw / bucket | RGW Admin Ops、S3 API、`radosgw-admin bucket` | 当前有桶及策略入口；生命周期、版本、锁、加密、tag、notification、replication 待核对 |
| rgw / multisite | `radosgw-admin realm/zonegroup/zone/period` | 已有部分管理；同步策略、placement/storage class、bootstrap/import 待核对 |
| monitoring / prometheus | Prometheus HTTP API | 时间序列/规则由 Prometheus 提供；本次修正规则页错误依赖 Alertmanager 的声明 |
| monitoring / alertmanager | Alertmanager HTTP API | 已有 alerts/silences；需核对规则分组、matcher 和错误展示 |
| grafana | Grafana HTTP API | 已有看板列表；嵌入配置与无 endpoint 行为需核对 |
| auth / role / multi-cluster | CephTower 自有认证、RBAC、集群管理 | 不直接复制 Dashboard 内部 JWT 用户库；根据用户可操作功能对照现有实现 |
| telemetry / settings / feedback | mgr 模块命令及外部服务 | 尚需区分 Ceph 集群功能和仅服务于原 Dashboard 的产品设置 |

## 已实现：总览及健康链路

参考依据：

- `dashboard/controllers/health.py` 的 `basic_health`、`client_perf` 和 `pg_info`。
- `dashboard/services/ceph_service.py` 的 `get_client_perf`、`get_pg_info`。
- `src/mon/MonCommands.h` 中 `health`、`health mute`、`health unmute`。
- `src/mon/HealthMonitor.cc` 和 `health_check.h` 的 checks/mutes JSON 格式。

实现：

1. `ceph health detail --format json` 单独采集健康详情，避免 `status` 的摘要造成详情丢失。
2. 将 `checks` 与 `mutes` 合并；已恢复但仍有 sticky 静默的记录保留，支持用户取消。
3. 详情命令失败将 `health_check` 标记为不可用，交给 reconciler 保留过期缓存。
4. `POST /api/v1/health/mute` 支持可选 `ttl` 和 `sticky`，严格检查时长和布尔类型。
5. `DELETE /api/v1/health/mute` 沿用现有取消操作；成功后前端刷新健康采集。
6. 展示 PG 状态数量/占比、读写 IOPS、读写吞吐、MGR/MDS active/standby。
7. 健康表格展示展开详情、当前是否触发、静默到期和持续静默状态。
8. `alert/rules` 由 Prometheus 提供，前端依赖声明同步修正。

### 对象健康、恢复和 scrub 指标

- `ceph status --format json` 提供池数量、对象数、恢复吞吐和 PG 状态；保留缺失字段为
  未知，不把旧版本或暂不可用的数据显示成零。
- `ceph pg dump summary --format json` 的 `pg_map.pg_stats_sum.stat_sum` 提供对象副本、
  degraded、misplaced 和 unfound 计数，和 Dashboard `get_pg_info()` 使用同一组字段。
- `ceph osd df --format json` 的逐 OSD `pgs` 计算平均 PG/OSD，忽略 CRUSH 根和主机节点。
- `ceph osd dump --format json` 提供 `noscrub`、`nodeep-scrub` 标志；结合 PG 状态，按
  Dashboard 语义输出 disabled、active 或 inactive。命令失败时输出未知，不伪造状态。
- 总览页新增对象健康副本比例、异常副本明细、恢复吞吐、scrub 状态、池数量和
  PG/OSD 密度。对象健康计算与参考前端一致：副本总数扣除三类异常副本。
- 附加命令是可选采集；失败不会覆盖已有总览，也不会让核心 `status`/`df` 数据失效。

## 已修正：CephFS 快照计划的文件系统参数

参考 `src/pybind/mgr/snap_schedule/module.py` 的 `snap_schedule_add`，位置参数顺序为
`path, snap_schedule, start, fs`。原实现把文件系统名作为第三个参数，实际传入了 start。
现在使用 `fs snap-schedule add <path> <schedule> --fs <filesystem>`，后置检查使用
`fs snap-schedule status <path> --fs <filesystem> --format json`，避免依赖默认文件系统。
测试明确核对写命令和后置检查的参数顺序。

该模块的 `list --recursive` JSON 只包含 schedule/retention 汇总，不能直接当成前端所需的
完整逐路径详情列表。实现与 Dashboard 控制器保持一致：先执行
`fs snap-schedule list / --recursive=true --fs=<filesystem>`，从逐行短格式发现并去重路径，
再对每个路径执行 `fs snap-schedule status <path> --fs=<filesystem> --format=json`。这样缓存的
每条记录保留 path、schedule、start、active、retention、创建/清理计数及子卷作用域。

无计划时 plain list 会返回 `ENOENT`；采集器使用同命令的 JSON 模式确认 `{}` 后才把结果视为
权威空列表，其他命令失败或非法结构会保留上次有效数据。新增
`GET /api/v1/filesystem/snapshot/schedules` 返回缓存列表；前端同时提供全部计划表和原有的指定
路径实时查询，全部计划可直接按其文件系统、路径、周期和开始时间执行启用、停用或删除。

## 已实现：CephX 认证实体管理

参考 `dashboard/controllers/ceph_users.py` 的 `user_list`、`user_create`、`user_edit`、
`user_delete`、`export`，以及 `src/mon/MonCommands.h` 的 auth 命令定义。

| API | Ceph 命令及语义 |
| --- | --- |
| `GET /api/v1/ceph/users` | `auth ls --format json` 定期采集后的实体、类型、caps；列表不包含 key |
| `POST /api/v1/ceph/user` | `auth add <entity> <subsystem> <caps> ...`；由 Ceph 生成密钥 |
| `PATCH /api/v1/ceph/user` | `auth caps <entity> <subsystem> <caps> ...`；替换全部权限 |
| `DELETE /api/v1/ceph/user` | `auth rm <entity>`；参考旧名称 auth del 已在源码标为 deprecated |
| `POST /api/v1/ceph/users/import` | `auth import -i -`，keyring 从 stdin 传入 |
| `GET /api/v1/ceph/users/export` | 对每个显式选择的实体执行 `auth export <entity>`，返回 keyring |

- 新增 `ceph_auth` 采集模块及 `ceph_ceph_user` 专用实体表，使用现有数据库迁移机制。
- 表中只保存实体、类型和权限，不保存 Ceph 返回的密钥；导出结果不进入资源缓存。
- 导入请求使用 `keyring` 敏感字段，审计自动脱敏；错误响应不回显导入/导出命令的原始输出。
- 导出限定显式选中的 1–100 个实体，响应设置 `Cache-Control: no-store`。
- 新增集群管理 → CephX 用户入口，支持搜索、类型过滤、分页、权限表单、批量删除与下载导出。
- 编辑会提示权限整体替换；删除确认说明影响；写操作完成后重新采集认证实体。
- caps 按单个参数传递，支持空格和带引号的路径；拒绝选项注入、NUL 和非法子系统。
- 无集群 API 集成测试实际经过路由、请求契约、命令执行模拟、资源入库、列表、导入、导出及审计。

## 已实现：CephFS 访问授权发现

CephFS Dashboard 的授权写入最终调用 `fs authorize`，但 Ceph 没有单独的授权列表命令。
CephTower 复用 `ceph auth ls --format json` 的 MDS caps，在不增加 CLI 调用的前提下把每条
`allow` 授权拆成 `cephfs_authorization` 资源。解析遵循 `MDSAuthCaps.cc` 的 cap 语法，支持：

- 逗号或分号分隔的多条授权，以及包含逗号的 `gids`；
- `fsname`、带引号或无引号的 `path`、`uid`、`gids`、`network` 和 `root_squash`；
- `r`、`rw`、`p`、`s`、`f` 扩展及 `*`/`all` 通配权限；
- 没有 `fsname` 的全局 MDS cap，使用 `*` 明确展示其作用域；
- 原始 MDS cap 与结构化字段同时展示，且认证密钥不会进入资源缓存。

授权创建成功后，`cephfs_authorization` 归属 `ceph_auth` 采集模块，立即重新执行同一条
`auth ls` 命令并更新授权列表；定时采集也会把已撤销或被修改的授权标记为过期。

## 已实现：CephFS 客户端会话

参考 Dashboard `CephFSClients` 与客户端表格，拓扑采集对每个文件系统执行
`tell mds.<fsname>:0 session ls --format json`。采集兼容 Ceph 版本中顶层数组及
`sessions` 对象两种 JSON 外形，保留完整 session 指标，并把 `client_metadata` 规范化为
类型、版本、主机和挂载根路径。列表展示与 Dashboard 对齐的 ID、类型、状态、版本、主机和
根路径，详情仍可查看 lease、cap、请求和存活时间等原始字段。

客户端驱逐限定到所属文件系统的 rank 0：
`tell mds.<fsname>:0 client evict id=<client_id>`，不再向 `mds.*` 广播可能与其他文件系统
同号客户端冲突的命令。后置检查复用同一 rank 的 `session ls`，成功后立即刷新
`cephfs_client` 缓存；没有活动 MDS 或返回结构无效时保留最后一次有效会话列表。

## 已修正：CephFS Shell 与目录配额

`cephfs-shell` 的入口参数只支持 `--config`、`--fs`、`--batch` 和 `--test`，不能复用
`ceph`/`rbd` 的 `--conf --name --keyring` 参数前缀。executor 现在通过 `CEPH_CONF` 和
`CEPH_ARGS` 向 libcephfs 传递临时配置、客户端实体及 keyring，并校验客户端实体名称，避免
把认证参数当成 shell 子命令参数或注入环境参数。

目录配额写入明确使用所选文件系统：
`cephfs-shell --fs <filesystem> setxattr <path> ceph.quota.max_bytes <bytes>`，随后在同一
文件系统执行 `getxattr` 并核对返回值。路径支持空格和引号，传给 shell 前按单一 token
转义；拒绝 NUL、换行和逗号，其中逗号在 cephfs-shell 中是多命令分隔符。这样写操作不再
依赖默认文件系统，也不会把“命令退出成功但值未生效”当成成功。

## 已实现：CephFS 目录浏览与双维度配额

参考 Dashboard `CephFS.ls_dir()`、`get_directory()` 和 `get_quotas()`，目录页改为按需读取，
不再查询一个从未被采集器填充的 `cephfs_entry` 缓存。`GET /api/v1/filesystem/entries`
接收文件系统和绝对路径，执行：

1. `cephfs-shell --fs <filesystem> ls -la <path>`，解析直属目录的权限、大小、UID、GID、
   修改时间和名称；文件不会混入目录树。
2. 对当前非根目录及每个直属子目录执行
   `cephfs-shell --fs <filesystem> quota get <path>`，读取 `max_bytes` 与 `max_files`。
3. cephfs-shell 用退出码 9 表示扩展属性未设置；仅该退出码会按 Dashboard 语义转换为
   0（不限制），其他退出、超时和不可解析输出均返回错误，不伪装成空目录。

executor 为非交互调用写入临时 `cephfs-shell.conf` 并关闭颜色，确保 `ls -l` 输出不会夹带
ANSI 转义。配额读取最多 8 路并发，且单层目录最多接受 500 个子目录，以限制一次请求触发
的数据面调用数量；路径中的
逗号、换行和 NUL 被拒绝，含空格或引号的路径作为一个 shell token 转义。

目录页支持选择文件系统、输入路径、进入子目录、返回上级，并实时显示两种配额。配额写入
支持同时或分别设置 `max_bytes`、`max_files`，0 表示不限制；双字段写入分成两个明确的
`setxattr` 步骤，最后在同一个指定文件系统回读两个属性并逐值核对。能力探测改用
cephfs-shell 实际支持的 `--help`，不再调用不存在的 `--version`。

## 已实现：CephFS Rank 与备用 MDS

对照 Dashboard `fs_status()`、`_find_standby_replays()` 与 Rank 表，新增实时
`GET /api/v1/filesystem/mds`：`ceph fs dump --format json` 按 `mdsmap.fs_name` 选择 FS，
从 `in`、`up`、`info` 关联活跃 Rank，未分配的 in Rank 标为 failed，standby-replay
独立标为 `<rank>-s`；保留 laggy 标记和字符串 GID。全局 standbys 单独展示，明确不是
该文件系统独占资源。核验重复 Rank、GID/名称、info 键与 GID 一致性及 up→info 关联，
不把畸形 FS Map 当作空结果。FS 不存在返回 404。

随后 `ceph mds metadata --format json` 读取实例 Ceph 版本；元数据失败不隐藏有效拓扑，
单独提示，缺失版本显示 —。性能采样新增 `mds_mem.dn/dir/cap`、
`mds_sessions.session_count` 和 `mds_log.replay`，Rank 表按名称及 GID 匹配已有样本，
展示 Dentries/Inodes/Dirs/Caps、采样时间及活动速率。active 使用客户端请求速率，
standby-replay 使用日志回放速率，其他状态与缺失样本不伪造活动。客户端数优先 Rank 0，
缺失或为零时参考其他 Rank，但不求和；保持字符串精度。拓扑和性能分别采样、展示时间，
不宣称为跨命令原子快照。离线测试覆盖 failed/replay/laggy、备用实例、版本局部失败、
拓扑身份异常、路由及前端 GID/速率/客户端计数逻辑；尚未实机验证。

## 已实现：CephFS 存储池容量

对照 `CephFS.fs_status()` 的 pools 表与 `CephfsDetailComponent` 容量展示，详情页新增
元数据池/数据池实时容量表。`GET /api/v1/filesystem/pools` 接收 `cluster_id`、`fs`，
执行 `ceph fs get <fs> --format json` 读取 `metadata_pool` 与 `data_pools`，再执行
`ceph df detail --format json`，按 pool ID 精确匹配，只返回该文件系统关联的池。
池名来自原生 df 条目，匹配缺失的池仍保留 ID 并标记错误，不把其他池的值用于替代。
逻辑存储量使用 `stats.stored`，可用量使用 `stats.max_avail`，估算容量为两者之和；
`stats.bytes_used` 另列为物理占用，不混入参考 Dashboard 的逻辑使用率口径。
容量是当前集群空间、放置约束下的估算，不能视为预留空间；共享池的值也不是此 FS 独占量。

API 以十进制字符串输出容量，后端 big.Int 加总和前端 BigInt 百分比计算避免大整数失真。
页面显示二进制容量单位、精确字节 tooltip、用途和逻辑使用率，支持独立实时读取及详情页
刷新联动。缺失/非法统计显示 null 或池级错误，容量为零时不计算无意义的百分比。
离线测试覆盖多个数据池、不同于逻辑量的物理占用、超 uint64 总和、零值、缺失统计、
重复 ID、原生命令失败、路由调用与前端格式/使用率边界；尚未经过真实集群验证。

## 已实现：CephFS MDS 性能计数器

参考 Dashboard `CephFS._mds_counters()` 的 11 项非标签计数器与 `CephfsChartComponent`
的 inode/request 趋势，新增实时 `GET /api/v1/filesystem/performance`。
先执行 `ceph fs get <fs> --format json`，核验 `mdsmap.fs_name` 并从 `info` 定位该文件系统
的全部 MDS（包含 standby-replay，不引入无归属的全局 standby）；再逐实例执行
`ceph tell mds.<name> perf dump --format json`。`ceph_context.cc` 注册的 `perf dump`
返回非标签指标，与参考 mgr `get_unlabeled_counter` 的指标来源对应。此读取链不依赖
volumes 模块或外部 Prometheus。MDS 名称/GID 校验、排序和最多 256 实例限制防止不受控目标。
每个实例返回身份、rank、state、采样时间和指标：
`mds_server.handle_client_request`、`mds_log.ev`、`mds_cache.num_strays`、
`mds.exported`、`mds.exported_inodes`、`mds.imported`、`mds.imported_inodes`、
`mds.inodes`、`mds.caps`、`mds.subtrees`、`mds_mem.ino`。
uint64 值与 GID 以字符串提供，缺失指标为 null；错误单独挂在实例上，不伪造 0 或吞掉错误。

详情页每轮完成后间隔 10 秒采样，手动采样可用，最多保留 60 次当前页面会话样本。
展示全部当前值与 inode 数/请求速率图；速率由累计差值除以真实采样秒数得到，使用 BigInt
先求差再转换，实例 GID 变化、负差、缺失值、读取失败与无效时间都断开趋势。
原生命令不提供 ceph-mgr 已保存历史，因此不将页面会话趋势冒充 mgr 历史；超过安全整数
范围的 gauge 保留精确表值，不绘制失真的图点。空关联 MDS 与局部失败均有明确提示。
离线测试覆盖命令作用域、大整数、部分失败、缺失项、畸形返回、实时路由和前端速率边界；
尚未进行真实集群验证。

## 已实现：CephFS 文件系统卷重命名

对照 Dashboard `CephFSUi.rename()`、CephFS 编辑表单和 volumes 模块的
`rename_fs_volume()`，文件系统列表新增独立“重命名卷”操作。
`PUT /api/v1/filesystem` 要求 `cluster_id`、`fs`、`new_name`、`confirmed: true`，
执行 `ceph fs volume rename <old> <new> --yes-i-really-mean-it`。任务风险级别为 high，
界面确认提醒维护窗口、客户端 CephX 重新授权、可能的存储池与 MDS 服务变更。
只调用完整的 volumes 操作，不使用仅改变 FS Map 名称的 `fs rename` 替代。
原生模块负责调整 MDS 服务和存储池，多个数据池时可能仅返回未重命名全部数据池的提示；
保留脱敏后的原生命令输出与警告到任务详情，不将名称核验冒充全部底层对象的验证。
后置 `fs volume ls --format json` 必须同时验证新名称存在、旧名称消失；空、畸形或重复
对象不能证明成功。成功后刷新 filesystem 和 pool 缓存；MDS 服务需重新观测。
拒绝无确认、相同名称和不合法名称；离线测试覆盖原生命令及回读失败，尚无真实集群验证。

## 已实现：CephFS 目录生命周期与快照

### 目录创建与删除

对照 Dashboard `mk_dirs()` 与 `rm_dir()`，目录页可在当前位置创建相对路径，或删除直属
空目录。`POST /api/v1/filesystem/entry` 调用
`cephfs-shell --fs <fs> mkdir -p -m 0755 <absolute-path>`，允许同时创建缺失父目录。
`DELETE /api/v1/filesystem/entry` 调用 `cephfs-shell --fs <fs> rmdir <absolute-path>`，
由 Ceph 检查目录是否为空。两项写操作都回读父目录的 `ls -la`，逐名称验证存在或消失，
并复用异步任务和审计。根目录、快照目录、通配符、逗号和换行路径会被拒绝，避免误操作。
测试覆盖含空格的目录、非法作用域、无法解析的后置列表以及路由到任务执行的完整调用链。

### 目录重命名与移动

目录行新增目标绝对路径表单，`PATCH /api/v1/filesystem/entry` 接收 `fs`、`path`、
`destination`。对照 Dashboard `rename_path()` 与原生 shell `do_mv()`，执行
`cephfs-shell --fs <fs> mv <source> <destination>`，最终调用 LibCephFS `rename`。
执行前读取源父目录，确认源是目录；执行后分别读取目标父目录和源父目录，验证目标目录
存在、源目录消失。未能读取或解析列表会让任务失败，而不是将命令接受当作验证成功。
拒绝根目录、快照路径、源和目标相同、目标位于源子树，以及通配符/逗号/换行。
保留 Ceph 原生替换语义：可能替换空目标目录，非空目标由 Ceph 拒绝，界面明确提示。
这不是事务性的目录身份验证，其他客户端并发修改仍可能造成竞态；尚未进行真实集群验证。
离线测试覆盖跨父目录含空格命令、作用域校验、路由任务及前置/双后置调用链。

### 目录快照

目录浏览页新增当前目录的快照表、创建表单和删除确认。参考 Dashboard
`CephFS.ls_snapshots()`、`mk_snapshot()`、`rm_snapshot()`，数据面调用链为：

- `GET /api/v1/filesystem/entry/snapshots` → `cephfs-shell --fs <fs> ls -la <path>/.snap`。
- `POST /api/v1/filesystem/entry/snapshot` → `cephfs-shell --fs <fs> snap create <name> <path>`。
- `DELETE /api/v1/filesystem/entry/snapshot` → `cephfs-shell --fs <fs> snap delete <name> <path>`。

创建和删除通过同目录快照列表核验目标名称存在或消失；不可解析的列表不能证明删除成功。
列表忽略普通文件、`.`、`..` 和以下划线开头的内部快照，与参考 Dashboard 的过滤一致。
展示 `ls -l` 提供的修改时间，不将其冒充 libcephfs 的 `st_ctime` 创建时间。

`client_snapdir` 是客户端虚拟目录名称。executor 给临时 cephfs-shell 客户端设置
`--client_snapdir=.snap`，使列表与 `snap` 内部 `conf_get('client_snapdir')` 的路径一致，
不会修改集群配置。写请求明确限定集群、文件系统、目录和名称，复用任务执行与审计链路。
无集群测试覆盖目录快照发现、命令参数、名称校验、后置核验和 API 路由。

## 已实现：集群配置与运行日志

配置依据：`dashboard/controllers/cluster_configuration.py`、`src/mon/ConfigMonitor.cc`、
`src/common/options.cc` 和 `MonCommands.h`。新增集群配置页面，从已有 `config ls`/`config dump`
缓存浏览选项和已设置的作用域，选中后调用 `GET /api/v1/configuration/option` →
`ceph config help <name> --format json`，展示类型、级别、说明、默认值、范围、枚举、
标签及运行时可更新性。避免每次刷新为数千个选项逐一启动命令。

- 修复配置值只允许“标识符”的错误限制；支持空串、空格、负数及多行 JSON 内容。
- 使用 `--value=<value>` 明确传值；空值/多行值在 `--` 后使用 `--value` 与独立参数。已直接运行参考
  `ceph_argparse.validate` 核对空值、负数、空格和以 `--` 开头的值，避免 CLI 解析歧义。
- 设置、删除、资源版本检查和入库键均保留 `osd/host:node-a`、`osd/class:ssd` 等作用域。
- 后置检查使用 `config dump`，不把带 mask 的作用域误用为单一 daemon 身份查询。
- 重新采集后展示 Ceph 实际返回的值，本地已提交配置不再覆盖发现值。
- 密码类选项的 name/value 结构在审计及资源存储中脱敏；命令参数标为敏感。

日志依据：`dashboard/controllers/logs.py` 的 `load_buffer`、`src/mon/LogMonitor.cc` 和
`src/common/LogEntry.cc`。`GET /api/v1/logs` 接收 `cluster_id`、`channel`、`level`、`limit`，
调用 `ceph log last <limit> <level> <channel> --format json`，返回 `items` 和 `observed_at`。
支持 cluster/audit/cephadm/全部频道、5 个最低级别，限制 1–500 条；最新日志在前。
序列号以字符串输出，避免浏览器丢失大整数精度。

前端支持搜索、暂停/启用 10 秒轮询、取消过期请求和下载当前结果；读取失败不会伪装成
空日志。移除原先只发送心跳、从不提供日志的 `/logs/stream`，以真实轮询替代。
该视图只提供 MON 缓冲区记录，不宣称具备长期日志归档。

新增测试覆盖：命令参数、JSON 错误/空数组、日志顺序和大序列号、配置 metadata、
配置值与 mask 作用域、凭证脱敏，以及路由到命令模拟和入库的完整链路。

## 本轮检查结果

- `make test-backend` 通过（全量 Go 测试及 OpenAPI 一致性检查）。
- `make test-frontend` 通过（TypeScript 类型检查与 Vite 生产构建）。
- `git diff --check` 通过。
- 未提交 Git commit；未修改用户已有的实验环境配置变更。

## 验证边界

无真实集群，fixture 是合成/仓库测试数据，不能当作实机测试记录。
测试覆盖详细健康 JSON、告警恢复后的 sticky 记录、失败保留缓存标志、
命令参数、非法 TTL/类型，以及请求契约。完整检查使用 `make test-backend` 和
`make test-frontend`；OpenAPI 由仓库生成器更新。

仍需实机验证：Ceph 权限与版本差异、静默过期语义、网络故障、规模性能，以及各外部
网关/Prometheus 的实际协议。没有因缺乏集群而在生产路径返回模拟成功。

### MGR module state and schemas

- Reference: `src/pybind/mgr/dashboard/controllers/mgr_modules.py` and
  `src/mon/MgrMonitor.cc` in the Ceph reference tree.
- Combine `ceph mgr module ls --format json` activation state with
  `ceph mgr dump --format json` available-module metadata. Preserve enabled,
  always-on, force-disabled, can-run, load-error and module-option schema fields.
- Exclude the reference dashboard's internal selftest module. Missing command
  output is unavailable data, rather than an empty authoritative module list.
- Module UI now displays load failures and schema details; activation triggers
  collection before rereading the list. The module configuration action opens
  the shared editor filtered to `mgr/<module>/` options, with default scope `mgr`.
  It supports setting and removing overrides and inspecting option metadata.
  `ConfigMonitor.cc` includes manager options in both config ls and config help.
  Writes follow `PyModuleConfig::set_config`: config set/rm with scope mgr.
  Configuration resource identities encode scope and full name together so
  slash-containing option names cannot be confused with scope masks.
- Verified backend tests, OpenAPI consistency and frontend production build.
  Fixtures cover disabled-module objects and force-disabled always-on modules;
  no live Ceph cluster was used.

### OSD on-demand inspection

- Reference `dashboard/controllers/osd.py` get/histogram and `OSD.cc` command handler.
- Added GET `/osd/inspection` with explicit metadata/histogram section selection.
  Executes `ceph osd metadata <id> --format json` or
  `ceph tell osd.<id> perf histogram dump --format json` through the native executor.
- OSD details expose cached state, live metadata and histogram axes/bucket values.
  Histogram rendering now uses a scrollable two-dimensional heat table, with
  inclusive Ceph-provided bucket ranges, axis names, cumulative counts, logarithmic
  color intensity and expandable raw data. Shape mismatches are not plotted.
  Axis ordering and open-ended buckets follow common/perf_histogram.h and .cc.
  Each live section reports errors and supports rereading; no offline success data.
- Exposed the existing deep-scrub operation in the OSD list.
- Backend tests verify command arguments, read-only execution, invalid targets,
  malformed responses and offline failures. Backend/OpenAPI checks and frontend
  build passed; real-cluster and browser visual validation remain outstanding.

- OSD in/out/scrub, reweight and removal now explicitly recollect native resources
  before reloading cached rows. Removal messaging reflects an asynchronous request,
  not completion of physical removal. Frontend build and diff checks passed.

### OSD deployment preview output

- The dry-run command output is returned in action details and rendered in the
  deployment form. Editing inputs clears the preceding preview.
- Preview executes orch apply osd --dry-run, is marked read-only in executor
  metadata, skips the unrelated daemon post-check, and does not persist a
  configured deployment resource. Output uses the existing text redactor.
- A service test verifies exact arguments, one command only, output propagation,
  secret redaction and failure propagation. Backend/OpenAPI checks passed.

- OSD management now shows removal queue status, remaining PGs, host, replacement
  and force flags, and process start time from orch osd rm status. The table exposes
  observation age/staleness and is included in explicit OSD refresh requests.

- OSD, removal queue and MGR module lists now consume all API cursor pages.
  Shared pagination also serves the configuration editor, aggregates stale metadata
  and rejects repeated cursors instead of looping or silently truncating results.

- Snapshot schedule creation supports optional start, subvol and group, matching
  snap_schedule/module.py. Post-check retains subvolume/group scope. Command tests
  verify exact scope, start placement and rejection of control-line input.

- Snapshot schedule status now uses a direct scoped command. Its UI supports
  activate/deactivate/remove of an exact period/start and path-level retention
  add/remove. Retention expressions reject duplicate units and malformed input;
  command tests cover operations and invalid expressions. Full-path discovery
  remains incomplete. The obsolete cached list endpoint and UI have been removed;
  creation now lives alongside direct scoped querying and actions.

- The schedule page reads the collected `snap_schedule` manager-module state before
  enabling query or mutation controls. A disabled runnable module can be enabled through
  the existing strict manager-module API; unavailable modules surface `can_run` errors
  instead of failing the first schedule command.

- Schedule responses require path, period, start and boolean active state before
  being exposed to actionable UI. Tests reject null rows and ambiguous identities.

### Snapshot schedule discovery boundary

The reference CLI recursive JSON list returns one requested path and flattened
schedule/retention arrays; it does not emit each matching path. Plain output uses
`Schedule.__str__` with unescaped space/newline-separated paths, so arbitrary
filesystem names cannot be recovered reliably. Full-path discovery still needs a
structured data-plane enumeration approach. Scoped status remains authoritative.
Schedule mutations no longer write the obsolete per-filesystem configured cache,
which could collapse multiple paths and schedules into one resource key.

- RBD snapshot UI now exposes clone and rollback via the existing native action
  endpoint. Rollback confirms the target and data-loss effect. Form mutations
  recollect RBD state. Tests verify source/destination namespace preservation,
  exact commands/post-checks and rejection of a missing clone destination.

- RBD group member add/remove and group snapshot creation now connect UI to
  native commands, with qualified-group command tests. Namespace discovery covers
  group membership and snapshots; image and trash scopes are also preserved.
  Group snapshot create/delete/rollback now use explicit group and snapshot targets; broader mirroring remains outstanding.

### RBD pool mirroring collection and configuration

- Mirroring API returns one resource per pool, matching the list UI and mutation identity.
- Pool mode is read with `rbd mirror pool info`; enabled pools additionally use
  `rbd mirror pool status --verbose --format json` for health summaries, daemon records,
  and per-image status, following `src/tools/rbd/action/MirrorPool.cc` output fields.
- The UI exposes disabled/image/pool configuration and confirms disabling mirroring.
  Successful changes trigger native collection before refreshing the table.
- Fixture tests cover two independent pool identities and status fields. No live cluster
  execution has been performed. Pool- and cluster-level schedule administration remains unfinished.

### RBD image mirroring actions

- Image collection reuses `rbd info <pool[/namespace]/image> --format json` and
  exposes its `mirroring.mode`, `state`, `global_id`, and `primary` fields. The
  image table now shows the synchronization mode, state, primary/secondary role,
  and the next scheduled mirror snapshot without adding another per-image command.
- Image rows expose journal/snapshot enable, disable, promote, demote, resync, and
  mirror snapshot creation through the existing strict image-action API.
- Commands follow `src/tools/rbd/action/MirrorImage.cc`, with full pool/namespace/image
  targets. Postchecks use mirror image status, or ordinary image info after disable.
- Demote, disable and resync show operation-specific confirmation. The native CLI
  enforces mirror mode and primary-state prerequisites and returns errors to the UI.
- Promotion optionally maps to `rbd mirror image promote <spec> --force`; the UI
  requires an explicit switch and warns about dual-primary and data-conflict risk.
  Disable similarly maps to `rbd mirror image disable <spec> --force`, matching
  `MirrorImage.cc` support for disabling a non-primary image. Both paths require an
  explicit switch and operation-specific high-risk confirmation.
- Strict request validation rejects non-boolean force values and force on other actions.
- Namespace-aware command tests cover all seven operations, forced promotion, and
  forced disable. Fixture tests cover enabled and absent mirroring metadata.
  Live-cluster validation remains pending.

### RBD mirror snapshot schedules

- RBD image rows now show the effective snapshot schedule, its image/namespace/pool/
  cluster inheritance source, all interval/start-time pairs, and the next scheduled run.
- Collection executes one recursive `rbd mirror snapshot schedule list --format json`
  and one global `status --format json` per storage refresh, then applies the most
  specific schedule to every image without per-image command fan-out.
- Image actions add a schedule, remove one interval/start pair, or remove all schedules
  defined directly on that image. They map to native `schedule add/remove --image=...`
  commands and verify the exact image-level state through a recursive list readback.
- Interval and start-time inputs are strictly validated. Fixture tests cover exact
  command boundaries, normalized timezone-aware start times, readback failures,
  inheritance precedence, next-run merging, malformed native data, and namespaces.
  Pool- and cluster-level schedule administration and live-cluster validation remain.

### RBD mirroring peers

- Added `POST /rbd/mirroring/peer` with strict add/remove actions, per-pool UI forms,
  explicit peer UUID deletion confirmation, and refresh after native collection.
- Commands map to `rbd mirror pool peer add/remove` in `MirrorPool.cc`; add uses
  explicit remote cluster/client and rx-only/rx-tx direction. Postcheck reads pool info.
- Direct peer add still requires remote connection configuration already present on the
  execution host. Peer editing and bootstrap token exchange are available separately.
- Command tests verify argument boundaries, postchecks, missing targets and invalid
  directions. These are fixture checks, not real cluster validation.

### RBD peer editing

- Peer forms now update site name, client name, remote monitor addresses and direction
  via `rbd mirror pool peer set <pool> <uuid> <field> <value>`.
- The API validates field names and accepts rx-only/tx-only/rx-tx for updates, matching
  Ceph's separate add/set direction rules. Monitor address strings remain one argument.
- Tests cover supported fields, invalid direction, missing UUID and rejected key-file
  parameters. Bootstrap token exchange is implemented as a separate secret-aware flow.

### RBD mirroring bootstrap tokens

- Pool rows expose token creation and import forms following Ceph Dashboard's bootstrap
  workflow. Disabled pools are enabled in image mode before either operation.
- Creation maps to `rbd mirror pool peer bootstrap create <pool> --site-name=<site>`;
  import maps to `rbd mirror pool peer bootstrap import <pool> -` with explicit
  `--site-name=<site>` and `--direction=<rx-only|rx-tx>` options. The token is passed
  only through stdin.
- The two APIs execute synchronously instead of entering the durable operation queue so
  bootstrap secrets are never persisted in operation parameters. Requests and responses
  use `Cache-Control: no-store`; audit request redaction removes the token value.
- Token shape, pool/site names, and direction are validated before execution. Fixture
  tests cover automatic pool enablement, exact arguments, stdin-only secret handling,
  postchecks, invalid input, malformed native output, and secret-free errors. No live
  cluster execution has been performed.

### Mirroring identity fields and malformed responses

- Pool tables display site_name, mirror_uuid and remote_namespace from native pool info.
- Collector validates the native mode field (disabled/image/pool/init-only); null,
  missing, non-string and unrecognized modes mark mirroring unavailable instead of
  producing empty authoritative state. Tests verify cache-preservation signaling.
- Peer monitor addresses and credentials are not returned by ordinary pool info;
  obtaining those requires a dedicated secret-aware detail path and remains pending.

### RBD group rename and removal

- Group management now exposes rename/remove through `POST /rbd/group/action`.
- Native commands follow `src/tools/rbd/action/Group.cc`. Rename accepts a new group
  name and keeps the source pool and namespace; removal requires an explicit group.
- Postchecks list groups with explicit pool/namespace options. UI confirms removal
  and recollects after successful operations.
- Tests exercise default/named namespaces and reject cross-scope rename, empty path
  components and missing targets. Real cluster behavior remains unverified.

### RBD group snapshot rename

- Group snapshot forms support rename through the existing strict snapshot action API.
- Native command is `rbd group snap rename <pool/namespace/group@old> <new-name>`,
  matching `execute_group_snap_rename` in `Group.cc`. Destination paths are rejected;
  the operation remains within the source group. Postcheck lists that group's snapshots.
- Tests cover namespace-qualified source, plain destination, missing names and invalid
  destination paths. Live cluster verification is still outstanding.

### RBD image rename

- Image actions expose rename. New names stay within the source pool/namespace, as
  required by `src/tools/rbd/action/Rename.cc`; paths and snapshot suffixes are rejected.
- The command uses complete source/destination specs and checks destination image info.
  Successful mutations trigger recollection before rendering the image list.
- Tests cover a namespace-qualified image, destination postcheck and invalid new names.
  No live cluster validation has been performed.

### Snapshot rename UI and group request contract correction

- Ordinary RBD snapshot rows now expose rename through the existing PATCH snapshot API,
  preserving encoded image identity and the original snapshot name.
- Review found the group action request contract had not been registered despite passing
  command tests. Registered it, regenerated OpenAPI, and added explicit registration and
  request validation tests. Prior command-level checks did not establish API completeness.

### Mutation registration verification

- Replaced the self-referential request-contract coverage test with an AST scan of
  actual handler MutateResource registrations. Every registered action must have a
  request contract; missing registrations such as the earlier group-action omission
  now fail the backend test suite. Dynamic action registrations require explicit coverage.
- This verifies handler-to-contract wiring, not execution against a live Ceph cluster.

### RBD trash purge and detail fields

- Trash rows expose confirmed purge of expired images in their pool/namespace.
  Native purge and postcheck use explicit pool/namespace flags; malformed scopes fail.
- Collection uses `trash ls --long` and displays native deleted_at, status, source and
  parent fields, matching `Trash.cc`. Ordinary short listing omits these fields.
- Namespace-aware purge tests pass; real cluster validation remains outstanding.

### RBD trash deferment

- Move-to-trash accepts optional expires_at in the image action API and form.
- RFC3339 timestamps with timezone are normalized to UTC for `trash mv --expires-at`,
  matching `Trash.cc` expiration handling. Omission preserves the native default.
- Tests cover timezone conversion and invalid date inputs. No live cluster verification.

### RBD trash purge cutoff

- Purge accepts an optional RFC3339 expired_before value, normalized to UTC and passed
  to native `trash purge --expired-before`. The confirmation displays the selected cutoff.
- Tests verify namespace preservation, timezone conversion and invalid timestamp rejection.
  Pool usage threshold and scheduled purge remain unfinished.

### RBD image info collection

- Default and named namespace images now fetch `rbd info <spec> --format json`.
  Features, parent identity, size, object count/size, format, creation time, data
  pool, block prefix, order, and stripe layout are exposed as typed table fields.
- Source fields follow `src/tools/rbd/action/Info.cc`; details retain striping,
  timestamps and other native attributes. Failed info reads mark image collection
  unavailable instead of treating missing details as authoritative.
- Images with `fast-diff` additionally run `rbd du <spec> --format json`. The head
  row supplies current-image usage, while summing snapshot and head rows supplies
  total usage including snapshots, matching the two values shown by Dashboard.
  Images without `fast-diff` omit usage rather than triggering an expensive scan.
- Fixture coverage verifies the exact usage command, snapshot aggregation, 64-bit
  capacity fields, features, parent snapshot, and layout retention. Collection adds
  one usage request only for eligible images; large-cluster performance is unverified.

### RBD feature mutation

- Image forms enable/disable exclusive-lock, object-map, fast-diff and journaling;
  deep-flatten is disable-only, matching Dashboard RbdService allowed feature sets.
- Native commands use `rbd feature enable/disable <spec> <feature>` from Feature.cc,
  with image info postcheck and recollection. Ceph enforces feature dependencies.
- Tests cover namespace-qualified commands, unknown features and rejected enable of
  deep-flatten. Multi-feature dependency ordering and live validation remain pending.

### RBD explicit shrink

- Resize forms expose allow_shrink with confirmation describing truncated-data loss.
- Backend adds `--allow-shrink` only for explicit true; sizes retain byte suffixes.
  Native `Resize.cc` rejects shrinking without this flag.
- Tests check both flag paths and namespaced image identity. Filesystem preparation
  remains the operator's responsibility; no live image resize was performed.

### RBD create layout options

- Create forms/API accept data_pool, stripe_unit (bytes), and stripe_count.
- Native create passes --data-pool, --stripe-unit with byte suffix, and --stripe-count,
  following ArgumentTypes.cc. Numeric layout fields must be positive integers.
- Command tests cover full namespace identity, units, and zero-count rejection.
  Ceph validates layout compatibility; no live image creation was performed.

### RBD object size

- Image creation supports object_size, sent to native --object-size in bytes.
- UI offers powers of two from 4 KiB to 32 MiB; backend validates the same range and
  shape. ArgumentTypes.cc documents the native range. Boundary tests cover valid
  endpoints, undersize, oversize and non-power-of-two values.
- No live cluster creation has been performed.

### Native RBD long-list correction

- List.cc emits image (not name) for long-list records and mixes image and snapshot
  entries in the same array. Corrected the wire field and skip records with snapshot.
- Default and named namespace images now share the scoped `ls --long --pool` collector.
  This removes the earlier duplicate default-namespace list/enrichment path and lets one
  image usage result feed both image and snapshot observations.
- Updated fixture data to match native output, including an image plus its snapshot.
  Namespace tests verify exactly one list and one eligible usage command per scope.
- Earlier fixtures did not represent the actual native long-list schema; their passing
  results were insufficient evidence for live compatibility. Backend checks now use the
  corrected shape. Real cluster validation is still outstanding.

### RBD snapshot purge

- Image rows expose confirmed purge of unprotected snapshots via native `snap purge`.
  The API action preserves full image scope and postchecks with snap ls.
- Snapshot protected/timestamp fields come from `rbd snap ls <spec> --format json`,
  following `Snap.cc`; string protection values are normalized to booleans for API and
  UI action state.
- For fast-diff images, the already collected `rbd du` rows are joined by snapshot name
  and exposed as `used_bytes`/`disk_usage`, without another command per snapshot. The
  action form defaults to protect or unprotect according to the observed state.
- Non-snapshot-mirrored images query `rbd children <image@snapshot> --format json`,
  matching `Children.cc`; child pool, namespace, and image identities are displayed.
  Snapshot-mirrored images skip this fan-out, matching Dashboard behavior.
- Delete is disabled with an explanatory tooltip while the snapshot is protected or
  still has child images. The submit path repeats the guard before issuing any request.
- Snapshot actions now include independent copy alongside clone. Copy maps to
  `rbd copy <pool[/namespace]/image@snapshot> <destination>` from `Copy.cc`; clone
  continues to map to `rbd clone`. Both validate full destination scope and verify the
  created image with `rbd info`.
- Namespace fixtures cover protected state, timestamp, usage, children, identity, and
  command deduplication. Command tests cover namespaced copy, clone, and purge. Live
  validation remains pending.

### RBD runtime status

- Image collection reads `rbd status <spec> --format json` and displays watchers,
  migration and persistent_cache data through runtime_status, matching Status.cc.
- Empty responses mark collection unavailable. Fixture tests retain all three sections.
- This adds another request per image; large-cluster collection cost and live behavior
  remain unverified. Status reflects the last collection, not a continuous live stream.

### RBD observed-state ownership

- RBD mutations no longer synthesize resource configurations from request bodies.
  Native collection owns displayed resource state; audit/task records retain operations.
- RBD DTOs ignore configured overlays so request names/targets cannot override native
  identities or leak destination fields into observed resources.
- Regression tests verify DTO ownership and that RBD mutations do not write synthetic
  state. UI already recollects after successful operations. Live validation is pending.

### RBD mutation scope correction

- PATCH image rename preserves namespace; image deletion postcheck lists the same
  pool and namespace instead of default namespace only.
- Encoded image specs require pool/image or pool/namespace/image, without empty
  components, extra levels or snapshot suffixes. Tests cover malformed specs and scope.
- Backend checks pass; real cluster mutation validation remains outstanding.

### RBD malformed list protection

- Null image, snapshot, trash and group lists mark their resource kind unavailable;
  valid empty arrays remain authoritative empty results. Missing snapshot names and
  trash IDs also mark the collection incomplete instead of silently omitting records.
- Backend regression checks exercise null list handling. Live failure-mode validation
  remains pending.

### RBD group identity details

- Groups collect native `group info` and expose group_id in the table. Missing IDs
  mark group collection unavailable.
- Group fixtures now match Group.cc: member state is numeric and snapshot list uses
  snapshot for the name field. Native member namespace is retained in raw details.
- Backend/frontend checks pass; live group inspection remains unverified.

### RBD namespace native shape correction

- Namespace.cc returns an array of objects with name, not an array of strings.
  Corrected collection and namespace fixtures; invalid names mark collection unavailable.
- Namespace rows now carry an explicit namespace field. Removed redundant/empty table
  columns. Namespaced image/snapshot/trash/group tests use the corrected native shape.
- Earlier namespace fixtures were inaccurate; real cluster validation remains pending.

### RBD effective image configuration

- Image collection reads `config image list <spec> --format json`; configuration rows
  preserve name/value/source from Config.cc and are displayed in image details.
- Fixture coverage verifies image-level QoS value/source retention. Null replies mark
  collection unavailable. Per-image configuration editing remains unfinished.
- Additional per-image collection cost and real-cluster behavior remain unverified.

### RBD image configuration editing

- Image forms set/remove non-secret RBD configuration overrides via native config image
  set/remove. Postcheck lists effective configuration, followed by recollection.
- Names require rbd_ prefix; set values are nonempty single-line strings. Ceph validates
  supported options and values. Tests cover namespace scope and missing set values.
- Secret configuration and empty-value overrides are not supported by this form.
  Real-cluster validation remains pending.

### Resource action menus

- Extended row operations are grouped in a click-open menu; detail/edit/delete remain
  directly accessible. Column width accounts for visible controls. Capability blocking
  applies to the menu and each operation.
- Added mirroring to the manual-refresh resource mapping. RBD mutation-triggered
  recollection already covered mirroring.
- Frontend typecheck/build passes; browser layout and keyboard interaction are unverified.

### Nested resource details

- Complex list cells open a field-focused drawer instead of flattening objects into
  long table strings. Object arrays render as paginated field tables in RecordDetail;
  nested objects use formatted text with bounded scrolling.
- Removed the 12-field detail cutoff so collected attributes are all accessible.
- Frontend build/typecheck passes; browser layout and interaction are unverified.

### Operation-specific form inputs

- Resource forms support conditional fields. RBD copy/rename targets, trash expiry,
  clone destinations, configuration values and group rename fields appear only for
  their relevant operation. Required inputs validate before submission.
- Unmounted fields do not preserve values, preventing stale input across action changes.
- Frontend typecheck/build passes; browser interaction remains unverified.

### Create snapshots from image rows

- Image action menus create snapshots using the selected encoded image identity, so
  pool and namespace need not be typed again. Existing native snapshot API is reused.
- Corrected the snapshot path in the manual-refresh mapping.
- Frontend typecheck/build passes; browser interaction remains unverified.

### Ceph time fields in details

- Detail views recognize deletion/snapshot/create/access/modify timestamps and last_update,
  including fields in nested tables.
- Explicitly zoned ISO timestamps use browser local formatting. Native ctime-style
  values without timezone remain verbatim rather than receiving an assumed timezone.
- Frontend typecheck/build passes; browser timezone scenarios remain unverified.

### RBD watcher identifier precision

- Native Status.cc emits unsigned client/cookie identifiers. Collector converts these
  fields to decimal strings before persistence/API serialization to avoid browser rounding.
- Tests retain 9007199254740993 and uint64 maximum 18446744073709551615 exactly.
  This verifies fixture serialization behavior, not a live watcher connection.

### RBD destination path validation

- Creation, copy, deep-copy and clone now validate complete pool/image or
  pool/namespace/image paths with the same rules as encoded source identities.
- Regression tests reject bare names, empty components, extra levels and snapshot
  suffixes for copy/clone destinations. Real cluster operations remain unverified.

### RGW user suspension correction

- Replaced nonexistent user modify --suspended with native user suspend/enable.
  Combined edits run modify then the state subcommand and user info postcheck.
- UI recognizes native numeric suspended state and permits max_buckets=-1 (unlimited).
- Tests cover suspend/enable alone and combined with edits. Source evidence is
  radosgw-admin.cc command registration and RGW user JSON encoding. No live RGW test.

### RGW editable email clearing

- User edits send empty email explicitly; backend uses --email= to retain empty argument
  semantics. Missing email remains unchanged for API callers. Multiline values fail.
- Native ceph_argparse confirms binary system flags accept true/false values and named
  arguments support equals syntax. Tests cover clearing and invalid email lines.
- Backend/frontend checks pass; live RGW validation is pending.

### RGW user creation requirements

- Creation requires display_name, matching RGW user validation, and supports optional
  max_buckets including -1 for unlimited. Invalid limits fail before native execution.
- Tests cover missing display name and bucket limits. Backend/frontend checks pass;
  actual RGW creation remains unverified.

### RGW observed user state

- User mutations trigger native user recollection; request bodies no longer synthesize
  user resource state or override observed fields. UID is the full native list identifier.
- Tests verify native suspended/email values take precedence and no synthetic create
  state is written. Backend/frontend checks pass; live RGW validation remains pending.

### RGW user storage statistics

- User collection reads native user stats without sync/reset mutation flags. Display
  retains stats and last_stats_sync/update. Account users are labeled account scope,
  following radosgw-admin.cc owner selection.
- Fixture tests cover account scope and full UID. Real RGW accounting freshness and
  collection performance remain unverified.

### RGW incomplete detail handling

- User/account/role detail failures and empty responses no longer emit placeholder
  resources. Their failure mappings mark collection unavailable and preserve old state.
- Null list replies are also unavailable. A regression test verifies null user details
  produce no placeholder and no nil-map write.
- Backend checks pass; live failure-mode validation remains pending.

### RGW account detail fields

- Maps native account id/name to account_id/account_name used by the UI.
- Account tables expose quota, bucket_quota and user/role/group/bucket/access-key limits,
  following RGWAccountInfo JSON encoding in rgw_common.cc.
- Fixture tests cover identity mapping and quota retention; live account validation pending.

### RGW 角色原生数据格式与创建约束

- 对照 `src/rgw/radosgw-admin/radosgw-admin.cc` 的角色列表实现，直接采集完整对象数组，避免按字符串名称列表解析时丢失全部角色。
- 页面显示原生 `RoleName`、`Path`、`Arn`、`AssumeRolePolicyDocument`、`MaxSessionDuration`、`CreateDate` 字段。
- 创建角色的信任策略改为必填 JSON 对象，API 契约与表单同步；Ceph 负责最终 IAM 策略语义校验。
- 增加原生对象数组采集回归测试。当前角色操作仍只有创建，删除、策略管理、租户和账户范围尚待实现；未进行真实集群验证。

### RGW 角色删除操作

- 新增 `DELETE /api/v1/rgw/role`，使用必填 `name` 构造 `radosgw-admin role delete --role-name <name>`，后续执行 `role list` 检查；对应原生 `OPT::ROLE_DELETE` 分支。
- 角色列表新增删除确认操作，创建和删除后主动采集 `rgw_role`。角色资源不再写入或合并请求参数作为展示状态。
- 命令参数及缺失名称回归测试、`make test-backend`（含 OpenAPI 一致性检查）、`make test-frontend` 均通过。
- 本节取代上节“角色操作只有创建”的状态描述。角色策略管理、租户/账户范围仍未完成；未进行真实集群验证。

### RGW 角色编辑

- 新增 `PATCH /api/v1/rgw/role` 和角色列表编辑表单，维护信任策略 JSON 与最大会话时长（3600–43200 秒，匹配 `rgw_role.h`）。
- 命令链：`role-trust-policy modify --role-name ... --assume-role-policy-doc ...` → `role update --role-name ... --max-session-duration ...` → `role get`。参考原生 `ROLE_TRUST_POLICY_MODIFY`、`ROLE_UPDATE` 分支及 Dashboard `RGWRoleEndpoints.role_update`。
- 两次写入不具备事务性，第二步失败可能留下第一步的修改。命令链及输入验证测试、后端测试、OpenAPI 检查、前端构建通过，未进行真实集群验证。
- 尚待角色内联/托管权限策略管理及租户、账户范围实现。

### RGW 角色内联权限策略

- 新增 `POST /api/v1/rgw/role/policy`，按 `action=put/delete` 调用 `role-policy put/delete --role-name ... --policy-name ...`；写入通过 `--perm-policy-doc` 提交 JSON 对象，随后 `role get` 检查。
- 角色行操作提供新增/替换与删除策略表单及确认，操作后重新采集；列表展示原生 `PermissionPolicies`（PolicyName/PolicyValue）和 `ManagedPermissionPolicies`。
- 对照 `radosgw-admin.cc` 的 `ROLE_POLICY_PUT/DELETE` 与 `rgw_role.cc` 序列化字段实现。命令参数及非法 JSON 回归测试、后端测试/OpenAPI 检查、前端构建均通过。
- 尚未进行实机验证；托管策略挂载/解除及租户、账户范围尚待实现。

### 角色字段独立更新与托管策略前置条件

- 角色更新 API 允许单独提交信任策略或会话时长，拒绝空更新，并在最后一次写入后读取角色。页面仅在信任策略文本改变时提交该字段，避免单改时长时覆盖策略。
- 独立字段更新与最终读取回归测试、后端测试/OpenAPI 检查、前端构建通过。
- 核对原生 `ROLE_POLICY_ATTACH` 分支发现其要求账户角色；当前角色采集只覆盖默认租户，因此本轮未接入托管策略挂载按钮。账户范围采集和操作标识仍是后续必要工作，不能视为已完成。

### 账户角色范围

- 采集账户详情后执行 `role list --account-id <id>`，角色采用 `AccountId/RoleName` 标识，默认租户角色仍使用角色名；拒绝采集结果中账户 ID 与请求不一致的角色。
- 角色创建表单增加账户 ID，列表展示账户；创建、编辑、删除、内联权限策略 API 与原生命令均传递账户范围，后续读取也使用同一账户。
- 账户列表或详情获取失败会同时将角色采集标记为不完整，避免把未采集的账户角色当作已删除。
- 账户身份隔离和写入/读取范围回归测试、前后端检查通过。未进行实机验证；非默认租户角色枚举及托管策略挂载仍待实现。

### 角色创建参数与托管策略源码核对

- 创建角色支持 `description`、`max_session_duration`，分别传给 `--description`、`--max-session-duration`；会话时长验证为 3600–43200 整数秒，未提交时使用 Ceph 默认值。页面新增输入并展示原生 Description。
- 参数和边界回归测试、后端测试/OpenAPI 检查、前端构建均通过，未实机验证。
- 托管策略 attach/detach 分支调用 `get_role(role_name, tenant, account_id)` 后 `load_by_id`；`RGWRole` 名称构造函数不设置 `info.id`，而 RadosRole::load_by_id 直接按 info.id 读取。参考源码存在定位不一致，需要进一步验证或寻找可用服务接口，本轮未将此命令接入可操作按钮。

### RGW 账户删除

- 对照 Dashboard `rgw_iam.py` 账户 DELETE 操作及原生 `ACCOUNT_RM`，新增 `DELETE /api/v1/rgw/account` → `account rm --account-id` → `account list`。
- 页面提供删除确认，创建/删除后重新采集账户和角色；账户展示仅采用采集结果，不再叠加请求参数。账户非空限制交由 `rgw::account::remove` 检查，没有引入级联清理。
- 命令及必填参数测试、后端测试/OpenAPI 检查、前端构建通过，尚未实机验证。账户修改、限额和配额操作仍需继续补齐。

### RGW 账户编辑与数量上限

- 对照 Dashboard `rgw_iam.py::set` 和原生 `rgw::account::modify`，新增 `PATCH /api/v1/rgw/account`，执行 `account modify` 后 `account get`。
- 页面支持账户名称、邮箱、max_users/max_roles/max_groups/max_buckets/max_access_keys；数量限制验证为 -1 或非负 32 位整数。
- 原生 modify 不允许更换租户，且忽略空名称和邮箱，故不提供租户编辑并拒绝清空名称/邮箱，避免虚报更新成功。
- 参数、零值、负值与小数验证测试，后端测试/OpenAPI 检查、前端构建均通过；未实机验证。账户容量/对象配额操作仍待实现。

### RGW 账户统计展示

- 对照 `rgw_account.cc::stats`，采集 `account stats --account-id <id> --format json`，保留 stats、last_synced、last_updated 于 storage_stats。
- 账户列表增加容量与对象统计详情入口；未触发 sync/reset，命令失败或缺失 stats 标记账户采集不完整。
- 账户原生数据回归测试已包含统计返回，后端测试/OpenAPI 检查、前端构建及 diff 检查通过。未进行真实集群验证；配额设置仍待实现。

### 账户与默认 Bucket 配额设置

- 新增 `PUT /api/v1/rgw/account/quota` 和两种范围的配额表单，使用 `quota enable/disable --account-id --quota-scope account/bucket --max-size --max-objects`，随后读取账户；原生命令同次写入限额和启用状态。
- 容量以字节提交并由 Ceph 向上取整至 KiB；API -1 表示无限制，容量命令用 -1024 避免参考源码中 unsigned rgw_rounded_kb(-1) 溢出为零，最终原生 clamp 得到 -1。对象上限直接使用 -1。
- 回归测试覆盖范围、单位、非法负值、小数及超出安全整数的值。后端测试/OpenAPI 检查与前端构建通过；尚未实机验证。

### RGW 用户配额

- 新增 `PUT /api/v1/rgw/user/quota`，支持 user/bucket 范围、启停及容量/对象上限，用户列表增加两类配额操作与字段展示。
- 启用用 `quota enable --uid` 同时保存限额；停用因原生 set_quota_info 的 disable 分支忽略限额，先 `quota set` 再 `quota disable`，最后 `user info`，两次写入不具备事务性。
- 保留 tenant$user 标识并验证输入；测试覆盖租户 UID、无限容量、零对象限额及停用命令链。前端构建通过，未实机验证。

### RGW 用户管理权限

- 对照 Dashboard 用户 capability 表单与原生 caps add/rm，新增 `POST /api/v1/rgw/user/caps` 及用户行操作，支持单个类型的 read/write/read,write/* 权限添加、移除。
- 命令使用 tenant$user 原始 UID 和独立 caps 参数，随后 user info 并刷新采集；拒绝在类型中拼接多个权限表达式，Ceph 最终校验支持的类型。
- 命令及类型验证回归测试、后端测试/OpenAPI 检查通过。未进行真实集群验证。

### RGW 用户限流读取

- 对照 `show_user_ratelimit` 与 `RGWRateLimitInfo::dump`，采集 `ratelimit get --uid <uid> --ratelimit-scope user --format json`，解析 user_ratelimit 对象并展示为 rate_limit。
- 保留 enabled、max_read_ops/max_write_ops/max_read_bytes/max_write_bytes；命令失败或返回对象缺失标记用户采集不完整。
- 原生嵌套结构回归测试、后端测试/OpenAPI 检查、前端构建通过，未实机验证。限流修改和全局限流仍待实现。

### RGW 用户限流修改

- 新增 `PUT /api/v1/rgw/user/ratelimit` 及用户操作表单，提交 enabled 和四项读写请求/字节限额。页面注明每 RGW 每分钟计量，0 表示无限制。
- 对照 set_ratelimit_info，先 ratelimit set 保存限额，再 enable/disable，最后 get；保留租户 UID，拒绝负数、非整数及超出安全整数范围的值。两次写入不具备事务性。
- 启停命令链、参数回归测试、后端测试/OpenAPI 检查、前端构建均通过，未实机验证。全局和 Bucket 限流仍待补齐。

### Bucket 限流读取

- 对照 show_bucket_ratelimit，采集 `ratelimit get --bucket ... --ratelimit-scope bucket`，有 tenant 时显式传入；解析 bucket_ratelimit 并在 Bucket 列表展示。
- Bucket 详情失败或空对象不再生成占位数据，详情与限流失败均标记 Bucket 采集不完整。
- 原生嵌套返回回归测试、后端测试/OpenAPI 检查、前端构建通过，未实机验证。Bucket 限流修改及全局限流仍待实现。

### Bucket 限流修改

- 新增 `PUT /api/v1/rgw/bucket/ratelimit`，从编码 Bucket ID 解析租户与名称，使用 ratelimit set → enable/disable → get，三步传递相同的 bucket/tenant/scope。
- Bucket 行操作支持四项限额和启停，完成后重新采集；原生命令使用 rgw_admin 能力，不要求 S3 端点，不将请求体持久化为 Bucket 展示状态。
- 租户命令链回归测试、后端测试/OpenAPI 检查、前端构建均通过，未实机验证；两次写入不具备事务性，全局限流仍待实现。

### Bucket 全局列表租户标识修正

- 原生全局 bucket list 遍历 bucket metadata 键，键为 `[tenant/]bucket`。采集现在拆分名称和租户，为 bucket stats 与 ratelimit get 分别传递 --bucket/--tenant。
- 校验统计返回的 bucket、tenant 与请求一致，再构造编码资源标识，避免重复编码租户前缀或误用其他租户数据。
- 租户 Bucket fixture 改为原生列表格式，并校验展示名与编码标识；后端测试（含 OpenAPI 检查）通过。本次未改前端，未进行真实集群验证。

### Bucket 统计字段展示

- 对照原生 bucket_stats 输出，列表展示 tenant、versioning、num_shards、placement_rule、zonegroup、bucket_quota、object_lock_enabled、mfa_enabled、reshard_status、id、creation_time、mtime。
- 原生 Bucket ID 使用 id 字段，与操作使用的编码 natural_key 分开；版本控制编辑读取已采集的 suspended 状态。
- 前端构建通过；字段来自现有 bucket stats 采集，无后端变更，未实机验证。

### 单个 Bucket 配额设置

- 新增 `PUT /api/v1/rgw/bucket/quota` 与行操作表单，编码标识解析为 bucket/tenant，调用 quota enable 或 set→disable，随后 bucket stats 并刷新采集。
- 对照 set_quota_info，容量为字节并向上取整到 KiB，负值 -1 表示无限制；停用命令忽略限额，因此先单独保存，两次写入不具备事务性。
- 租户范围及启停命令链测试、后端测试/OpenAPI 检查、前端构建通过。未进行真实集群验证。

### 多步操作最终读取修复

- 共用执行器现在使用命令链最后一个已声明的后置检查（含其二进制），不再忽略 followup.check；空 RBD/RGW 检查不再被转换成单独的 --format json。
- 实际 Service.Execute 回归测试验证限流依次执行 set、enable、get，且读取标记为非写入；后端测试/OpenAPI 检查通过。
- 此修复保证声明的读取被执行，不代表已比较读取值和请求值。尤其参考源码 Bucket quota 外层忽略 set_bucket_quota 返回值，仍需进一步增加结果字段一致性校验；未实机验证。

### Bucket 配额读取值校验

- Bucket 配额操作现在比较最终 bucket stats 的 tenant、bucket、bucket_quota.enabled/max_size/max_objects 与请求，容量按 KiB 向上取整。
- 写入返回成功但结果未变化、读取字段缺失、错误租户或 JSON 无效均返回 post_check_failed，补上参考原生配额调用可能忽略写入错误的缺口。
- Service.Execute 模拟执行器回归测试及后端测试/OpenAPI 检查通过。未实机验证；该结果一致性校验目前针对单个 Bucket 配额。

### 默认 Realm 全局限流读取

- RGW 总览新增 `global ratelimit get --format json` 采集，展示默认 Realm 的 user_ratelimit、bucket_ratelimit、anonymous_ratelimit。任一范围对象缺失标记状态采集不完整。
- 总览移除不对应 RGWStatus 的用户字段，显示 realms 与 global_rate_limit；限额单位为每 RGW 每分钟。
- 三类原生返回结构测试、后端测试/OpenAPI 检查及前端构建通过，未实机验证。其他 Realm 的限流选择及全局写入操作仍待实现。

### Realm 原生详情

- Realm 名称列表后逐项执行 `realm get --rgw-realm <name>`，保留 id、name、current_period、epoch；页面增加 Epoch 展示。
- 返回身份不匹配或读取失败标记 Realm 采集不完整，不生成仅含名称的占位详情。多站点列表返回 null 也标记不完整。
- 原生详情及错误身份回归测试、后端测试/OpenAPI 检查和前端构建通过，未实机验证。Zonegroup、Zone 完整详情仍待补齐。

### Zonegroup 原生详情

依据 `src/rgw/radosgw-admin/radosgw-admin.cc` 的 ZONEGROUP_GET 和
`src/rgw/rgw_zone.cc` 的 RGWZoneGroup::dump，枚举后调用
`radosgw-admin zonegroup get --rgw-zonegroup <name> --format json`。
详情通过现有资源 API 提供，页面展示 Realm、主 Zone、成员、端点、放置目标、
主机名、同步策略和启用特性。名称不匹配、缺少 ID 或空返回均视为采集失败，
不生成只有名称的详情记录，并保留该资源类型的采集不完整状态。
新增离线测试覆盖嵌套详情及异常返回；尚无真实集群验证，Zonegroup 编辑操作待补齐。

### Zone 原生详情

参考 Dashboard 的 `rgw-multisite-zone-form.component.ts::getZonePlacementData`，
结合原生 ZONE_GET 和 RGWZoneParams::dump，枚举后逐项执行
`radosgw-admin zone get --rgw-zone <name> --format json`。
通过现有资源 API 提供 Realm、放置池、存储类别及内部池配置，页面增加对应列；
完整详情保留其它原生字段。Zone 端点属于 Zonegroup 的成员 Zone，已在该页展示。
名称或 ID 异常时标记 Zone 采集不完整。系统密钥由现有持久化前递归脱敏处理；
离线回归覆盖详情、异常身份、空数据以及凭据脱敏后保留放置池。
Zone 编辑与真实集群验证仍待完成。

### Realm 编辑与默认 Realm

对照 RgwMultisite::edit_realm 接入 PATCH /rgw/realm：名称变更执行
`realm rename --rgw-realm <name> --realm-new-name <new_name>`，
选中默认时随后执行 `realm default --rgw-realm <new_name>`，最后按新名称读取详情。
无名称变化且未选择默认会拒绝空操作。多个写命令不具有事务性。
页面提交后重新采集 Realm 和 RGW 状态，默认标记来自 realm list 的 default_info，
不再用请求体覆盖 Realm 的真实采集数据。离线测试覆盖命令顺序、目标名称、
布尔校验和默认标记；未执行真实集群验证。

### 创建默认 Realm

Realm 创建表单增加“设为默认 Realm”，POST /rgw/realm 接受可选布尔字段 default。
依据参考 RgwMultisite::create_realm，选中时在 realm create 命令添加 --default；
未选中不传此开关，保留 Ceph 原生默认选择行为。拒绝非布尔参数，
执行后沿用 Realm/RGW 状态刷新。离线测试覆盖省略、false、true 和非法参数。

### Zonegroup 创建参数

依据 RgwMultisite::create_zonegroup，创建表单/API 支持 Realm、逗号分隔端点、
master 和 default；后端分别传入 --rgw-realm、--endpoints、--master、--default。
省略选项时使用 Ceph 原生行为，布尔值 false 不传对应开关。创建后重新采集
Zonegroup，停止以请求数据覆盖原生详情。离线测试覆盖完整参数和非法类型；
真实集群验证、Zonegroup 编辑仍待完成。

### Zonegroup 基础编辑

对照 RgwMultisite::edit_zonegroup/modify_zonegroup，新增 PATCH /rgw/zonegroup
及编辑表单，支持重命名、替换非空端点、设为主/默认 Zonegroup。写入后执行
period update --commit，并显式使用 Realm ID 限定目标，最后按新名称读取详情。
无 Realm 的 Zonegroup 编辑、清空端点、成员和放置目标编辑仍待补齐。
多个命令不具有事务性；测试覆盖顺序、Period 范围、读回目标及无效输入，未实机验证。

### 无 Realm 的 Zonegroup 编辑及 Realm 参数修正

原生 ZONEGROUP_MODIFY 将 --realm-id 用于重新绑定 Realm，并非查询范围限制。
编辑命令移除此参数，Realm ID 仅用于 Period 提交；没有 Realm 时跳过 Period，
直接读取更新后的 Zonegroup。增加独立 Zonegroup 命令回归，页面允许空 Realm ID。

### Zonegroup 成员管理

参考 add_or_remove_zone，在 Zonegroup 编辑 API 和表单增加 add_zones/remove_zones。
依次执行 zonegroup add/remove --rgw-zonegroup <新名称> --rgw-zone <成员>，
成员变更后再提交所属 Realm 的 Period，并读回详情。拒绝重复、交叉和非法成员名。
离线测试覆盖重命名后的成员目标、命令顺序及参数异常；未实机验证。

### 多站点默认状态展示

原生 REALM_LIST、ZONEGROUP_LIST、ZONE_LIST 均输出 default_info ID。采集器按
详情 ID 比较后通过 API 提供 is_default，Zonegroup 与 Zone 页面展示“当前上下文默认”。
命令未指定 Realm 上下文，因此此标记代表该次命令解析出的默认对象，
不声称展示每个 Realm 各自的默认对象。缺失或非法 default_info 保持未知，
不伪装成 false；空字符串明确表示没有匹配对象。离线测试覆盖三类资源的
匹配、不匹配、空、缺失及非法字段。

### Zone 创建拓扑参数

依据 RgwMultisite::create_zone，Zone 创建表单/API 支持 Zonegroup、端点及
master/default 开关，映射为 --rgw-zonegroup、--endpoints、--master、--default。
创建后刷新 Zone 和 Zonegroup 原生数据，停止持久化请求覆盖层。离线测试覆盖
完整命令和非法参数。系统凭据、同步配置、分层参数与真实集群验证仍待补齐。

### Zone 创建同步来源

对照 create_zone 和原生参数解析，创建 API/表单增加 sync_from_all 与 sync_from。
布尔选项显式生成 --sync-from-all=true/false，指定来源生成 --sync-from <列表>。
省略时使用 Ceph 默认行为，false 不会被当作省略。离线测试覆盖两个布尔值、
逗号分隔来源及非法类型；未实机验证。

### 创建归档 Zone

参考 rgw-multisite-zone-form 的 archive_zone 开关，Zone 创建表单新增“归档 Zone”，
选中时提交 tier_type=archive，原生命令传入 --tier-type archive。未选中不传参数，
沿用普通 Zone 的原生行为。API 限定目前表单支持的 archive 值；离线测试覆盖
命令及非法类型/取值，未实机验证。

### Zone 系统用户凭据

Zone 创建增加密码输入框，成对接受 access_key/secret_key，映射参考 create_zone
的 --access-key/--secret 参数。参数标记为敏感，API 字段为 writeOnly，沿用
审计及原生详情递归脱敏；不保存请求覆盖层。测试覆盖敏感参数位置和不完整凭据。
未实机验证。

### Zone 的拓扑和同步详情

RGWZone::dump 中的端点、归档类型、同步来源、只读标记及支持特性属于
Zonegroup 成员记录，而非 ZoneParams。采集后按原生 Zone ID 关联这些记录，
通过 zonegroup_memberships 提供给 Zone 页面详情表，保留每个 Zonegroup 的名称、
ID、Realm 和主 Zone 标记，避免同名误关联或多组信息互相覆盖。
离线测试覆盖名称变化、同名不同 ID、多组成员及 false 值保留；未实机验证。

### Zone 重命名

新增 PATCH /rgw/zone 与重命名表单，执行 zone rename --rgw-zone <旧名称>
--zone-new-name <新名称>，可指定 --rgw-zonegroup 更新成员名称；原生代码未指定时
尝试默认组，读取组失败或没有该成员仍可能返回成功。页面在唯一成员组时预填，
随后刷新 Zone/Zonegroup。提供 Realm ID 时提交该 Realm Period，最后按新名称读回。
离线测试覆盖组与 Period 参数、无 Realm 和空操作；其它 Zone 编辑仍待补齐。

### Zone 端点编辑

对照 modify_zone，PATCH /rgw/zone 和编辑表单支持替换非空端点列表。
指定 Zonegroup 后执行 zone modify --rgw-zone <新名称> --endpoints <列表>
--rgw-zonegroup <组>；仅修改端点不触发重命名，组合操作先重命名再修改，
最后提交提供的 Realm Period 并读回。要求显式选择组避免更新默认组中的错误成员。
测试覆盖两种执行路径和缺少组名；清空端点及其它 Zone 编辑参数仍待补齐。

### Zone 同步来源编辑

对照 modify_zone，编辑 API/表单支持 sync_from_all 和 sync_from，明确区分
保持原值、开启、关闭。显式开启全部同步且提供来源时，按参考实现使用
--sync-from-rm；其余使用 --sync-from。编辑要求指定 Zonegroup，沿用 Period
提交和刷新流程。测试覆盖布尔参数及两种来源命令；未实机验证。

### Zone 主/默认选项编辑

对照 modify_zone，Zone 编辑 API 和表单支持设为主 Zone、设为默认 Zone，
分别生成 --master、--default。false 表示不修改该选项，不表示取消主/默认身份。
要求指定 Zonegroup，支持单独更新和重命名后更新，沿用 Period 提交。
离线测试覆盖两种执行路径及布尔校验；未实机验证。

### Zone 归档类型编辑

参考 modify_zone 和同步模块注册，空 tier_type 映射默认 rgw 模块，archive
映射归档模块。编辑表单区分保持原值、普通、归档；普通显式生成 --tier-type=，
归档生成 --tier-type=archive，避免把空值误认为省略。测试覆盖类型切换和
非法类型，沿用指定 Zonegroup 与 Period 流程；未实机验证。

### Zone 只读编辑

原生 ZONE_MODIFY 的 is_read_only_set 区分省略与显式 false。编辑 API/表单
增加 read_only，生成 --read-only=true/false，支持保持原值、开启、关闭。
只读状态仍从 Zonegroup 成员详情读取；沿用指定组和 Period 提交流程。
离线测试覆盖开启、关闭及非法类型，未实机验证。

### Zone 系统凭据编辑

Zone 编辑增加成对系统凭据更新，映射 modify_zone 的 --access-key/--secret。
创建与编辑共用参数校验和敏感位置标记，组合重命名时敏感标记附在 modify
后续命令上；凭据省略则保持原值，不回填已有密钥。测试覆盖单独更新、重命名
组合及执行前拒绝不完整凭据，沿用审计脱敏；未实机验证。

### Zonegroup 更新后的关联刷新

Zone 的 zonegroup_memberships 由同次采集中的 Zonegroup 详情派生。后端
ReconcileKinds 只持久化请求类型，因此 Zonegroup 操作和手动刷新必须同时
请求 rgw_zonegroup、rgw_zone，否则 Zone 页面可能保留旧成员配置。
现已补齐这两条刷新入口；真实集群交互验证仍待完成。

### Zone 编辑的成员读回校验

指定 Zonegroup 的 Zone 编辑在原有 Zone 读回之后，再执行 zonegroup get，
确认组名、新成员名和成员 ID，并比较本次请求的端点、只读、归档类型、
全来源同步以及设为主 Zone 的结果。读回不匹配返回 post_check_failed。
避免原生命令仅返回成功却未修改目标成员时误报成功；同步来源集合、默认
选择和系统密钥值尚未做结果比较。离线测试覆盖字段不匹配与异常返回。

### Zone 同步来源结果校验

原生 zonegroup 成员更新对 sync_from 执行集合插入或删除，不替换整组。
读回校验现按本次操作检查指定来源全部存在（追加）或全部不存在（删除），
允许保留其它来源，拒绝缺失、null 和非法类型的列表。离线测试覆盖追加、
删除和异常返回；默认选择及密钥值的结果比较仍待补齐。

### Zone 凭据读回验证

Zone 创建或编辑携带系统凭据时，校验 zone get 返回的名称、ID 和 system_key
中的 Access/Secret Key 是否与请求一致。仅在内存比较，不将返回密钥写入结果
或错误消息；不匹配统一返回 post_check_failed。离线测试覆盖创建、重命名、
旧密钥和缺失字段；未实机验证。

### Zone 写操作身份校验

Zone 创建和编辑统一验证 zone get 返回的名称及非空原生 ID，不再仅在请求
携带凭据时验证身份。没有凭据的普通更新也会拒绝空返回、缺少 ID 或错误对象，
携带凭据时继续比较密钥。离线测试覆盖有效无凭据详情及各类错误身份。

### CephFS 子卷克隆状态与取消

对照 volumes 模块的 `fs subvolume info`、`fs clone status` 和
`fs clone cancel`，采集默认组及所有命名组中的子卷。以 info 的 type=clone 识别克隆，
不以 source 是否存在作为身份判断；来源缺失或为 N/A 仍查询克隆状态，保留快照的
已删除克隆不查询目录克隆状态。向页面提供 state、source、progress_report 和 failure。
子卷与快照缓存键包含组名，更新、删除、快照创建/删除也保留组作用域，避免同名
资源碰撞或操作错误对象。

CephFS 快照页新增克隆操作，支持源组、目标组和 pool layout，分别映射
`--group_name`、`--target_group_name` 和 `--pool_layout`。快照列表逐项调用
`fs subvolume snapshot info`，展示创建时间、数据池和待处理克隆标记。子卷页仅对 pending 或
in-progress 状态开放取消操作，调用 `fs clone cancel` 后以 `fs clone status`
读回。离线测试覆盖命名组采集、状态字段和精确命令参数；未在真实集群验证。

原生进度使用 `percentage cloned`、`amount cloned`、`files cloned`，来自 volumes 的
stats_util 和 `_get_clone_progress_report`。页面展示百分比进度条、原生容量比及文件比，
缺失报告不假定为 0% 或完成。来源按文件系统/组/子卷@快照展示，N/A 明确标记不可用；
错误展示原生 error_msg 与 errno。离线夹具改用真实字段，测试覆盖采集→API、类型识别、
缺失来源和页面组件输出。取消条件仍只使用成功读取的 clone_state。

### CephFS 子卷挂载命令

对照参考子卷 Attach 与 cephfs-mount-details，已就绪子卷的详情展示可复制的内核、
FUSE 和 NFS 挂载模板。FSID 使用集群 API 的发现数据（原生 status/discovery），
路径使用 `fs subvolume info` 经缓存/API 提供的 path，不推导子卷目录，也不复用服务端
客户端用户名或密钥。缺失/无效 FSID、路径或错误集群身份时不生成命令；仅 complete
状态展示入口。FUSE 按本地文档使用当前 `--client_fs` 参数，不复制旧兼容选项。
客户端名称与挂载目录可编辑，shell 参数统一引用，离线测试验证特殊路径不会被 shell
解释为命令。客户端仍需准备对应集群 ceph.conf/keyring 和本地目录；本功能不执行
任何挂载。NFS 模板明确要求已有导出并手动替换主机、端口、导出路径，不假设子卷
内部路径就是 NFS export。测试覆盖命令参数、缺失信息、集群身份及组件输出。

### CephFS 快照克隆依赖

对照快照列表的 Pending Clones 列与 volumes `snapshot_info/get_pending_clones`，
快照详情展示 `pending_clones` 的目标子卷和 target_group，省略目标组表示 `_nogroup`；
列表把 has_pending_clones 的 yes/no 显示为中文状态，并展示 orphan_clones_count。
原生 info 已由 `fs subvolume snapshot info` 采集，不新增额外命令。未知、无依赖、
目标列表缺失和孤儿记录分别显示，不能把未知当作无依赖；有待处理目标或孤儿记录
时禁用删除并提示刷新和检查集群，不自动取消克隆或清理索引。原生 snapshot rm 仍
负责执行时的依赖检查，缓存提示不能替代它。离线 API 测试核对组范围和完整原生
依赖数据，前端测试覆盖默认组、命名组、未知数据及组件输出；未连接真实集群。

### CephFS 子卷高级删除

对照参考界面删除确认框与 `fs subvolume rm` 原生参数，子卷页新增高级删除，
可选择 `--retain-snapshots` 保留已有快照，并可用 `--force` 删除失败或已取消的
克隆。请求仍携带资源版本前置条件并进行二次危险操作确认；普通删除入口保持简洁。
离线测试覆盖命名组、两个开关的精确顺序及作用域读回；未在真实集群验证。

### CephFS 组范围读取与版本检查修正

再次逐层检查发现，采集器的子卷和快照 natural_key 已包含组名，但通用接口的
详情查找、If-Match 检查与操作锁仍省略组名。现统一使用
`filesystem/group/subvolume[/snapshot]` 的缓存键；未指定组时明确使用 `_nogroup`。
命令参数仍携带请求中的原生 group，不以另一个组的同名缓存资源进行版本校验。

快照列表不再在未选子卷时错误设置 `/` 父键，集群范围页面能显示全部快照；
选择文件系统、子卷与组时使用精确父键，其他范围通过原生 fs/group/subvolume 字段过滤。
快照表格增加组名列，以区分不同组的同名快照。

离线 API 测试同时建立默认组、team-a、team-b 中的同名子卷和快照，单独提升 team-b
资源版本，验证详情读取、全量/范围列表、错误版本拒绝、异步执行时再次检查版本、
组范围锁以及 resize、snapshot rm、clone cancel 的准确 argv。未连接真实集群。

### CephFS 子卷完整创建参数

子卷创建改用原生文档中的命名参数，不再把内部保留名 `_nogroup` 作为显式组传入。
默认组会省略 `--group_name`，UID/GID 留空时继承父组，不再隐式固定为 0。
表单和 API 新增 earmark、Unicode normalization 与 case-sensitive 选项，并映射为
`--earmark`、`--normalization` 和 `--casesensitive`。离线测试覆盖默认组、命名组、
可选所有者以及完整参数组合；未在真实集群验证。

### CephFS 子卷摘要字段

参考子卷列表使用 `fs subvolume info` 的 bytes_pcent 和 created_at 展示配额使用率
与创建时间。子卷表格展示 bytes_pcent，并将原生 created_at 规范为 ceph_created_at，
避免资源信封的同名缓存记录创建时间覆盖它。子卷组及快照使用同样的字段规范，
表格统一格式化原生时间，原生 info 不返回时间时显示缺失，不用缓存时间代替。
无需新增命令或推导近似值；离线测试验证字段隔离，未连接真实集群。

### CephFS 资源展示与筛选字段一致性

子卷组的文件系统范围由采集器写入 filesystem，移除不存在的 fs 空列。通用资源
列表现在与 FieldColumn 定义一致：筛选选项请求、表格列筛选键及列表 API 请求使用
filterKey 指定的原生字段，filterKey=false 的列不请求选项也不显示筛选器。快照孤儿
记录列使用原生省略含义计算显示 0，因此禁用此展示列的筛选，避免把显示 0 当作
后端真实存在的字段值。测试验证字段去重、原生字段映射、禁用筛选及实际 API 参数；
不改变采集命令或缓存字段结构。

### CephFS 子卷就绪状态

子卷列表展示 `fs subvolume info` 的 state、type 和 pool_namespace，区分普通子卷、
克隆子卷和隔离命名空间。资源 status 改为原生 state；缺失时为 unknown，不再固定
为 available。参考 volumes 的 V2 info 在 snapshot-retained 状态只返回类型、特性和
状态：子卷目录已删除但快照保留。因此页面显示明确警告，只有 complete 状态开放
配额编辑和目录快照可见性入口，不对保留快照的子卷发起目录可见性读取。保留快照
仍可在快照页面管理，删除和克隆取消仍按各自规则处理。后端原生命令继续验证操作
结果；页面就绪限制不替代后端权限、版本校验或原生命令的实际状态检查。
离线测试覆盖原生采集→数据库→API 的 complete、snapshot-retained、unknown 状态、
克隆状态与命名空间，以及全部状态的页面就绪规则。未连接真实集群。

### CephFS 配额使用率和权限展示

子卷与子卷组列表根据原生 info 的 bytes_quota、bytes_used、bytes_pcent 展示二进制
容量及使用率进度条。`infinite` 显示“无限制”，不再向用户显示 `undefined`；缺失的
使用率保持不可用，不用缺失信息推导百分比。超过 100% 时保留原生百分比并告警，
进度条宽度限制为 100%。整数容量超过 JavaScript 安全精度时只接受精确数字字符串，
不展示已经失真的数字。原生 mode 展示 rwx 和八进制权限，并保留 setuid、setgid、
sticky 位以及 0000 无权限状态。数据仍由 `fs subvolume info` 和
`fs subvolumegroup info` 经现有采集/API 链路提供；离线测试覆盖格式和组件输出。

### CephFS 子卷配额更新

子卷编辑支持取消配额限制和禁止缩容。取消限制生成原生大小 `inf`，保护开关生成
`--no_shrink`，并继续保留命名组作用域。表单从 bytes_quota 预填当前状态，API 使用
独立的精确请求结构；离线测试覆盖无限配额、缩容保护和缺失大小校验。

组/子卷配额编辑使用十进制文本输入，更新 API 的 size 为正整数字符串，不接受
旧数值契约。前后端范围统一为 1..9223372036854775807 字节，命令 argv 原样保留
精确值。采集器在 json.Number 尚未失真时将 bytes_quota/bytes_used 转为字符串，
缓存与 API 继续保持精度。取消配额时隐藏缩容保护并发送 no_shrink=false，后端也
不生成无意义的无限配额保护参数，与原生 `_resize` 一致。保护文案为“不允许配额
低于已用空间”，而非禁止一切配额缩小。离线测试覆盖大整数采集→API、异步操作→
准确 argv、数值契约拒绝、上下界、异常值和表单预填；未连接真实集群。

子卷 resize 在命令成功后必须以同一文件系统、子卷和组范围的 `fs subvolume info`
确认 bytes_quota 精确等于请求值，取消限制时必须为原生 infinite。空响应、旧配额、
失真整数、错误类型、尾随 JSON 或读回失败均为 post_check_failed，不显示成功。
该错误明确提示修改可能已生效，需刷新实际状态；不自动回滚。离线执行测试覆盖
有限/无限配额、精确大整数、失败场景及命令组范围。

### CephFS 子卷组编辑

参考 `cephfs-subvolumegroup-form.component.ts` 的编辑流程通过 create API 更新已有组的
数据池、UID/GID、权限及配额。本地 volumes 模块 `create_subvolume_group` 对已有组
调用 `set_group_attrs`；未提供 UID/GID 时会设为 0，未提供 pool 时重新继承祖先布局，
mode 也有默认值。因此属性更新要求四个字段同时提供，不把省略字段当作“不修改”。

本项目 `PATCH /filesystem/subvolume/group` 新增可选的 pool、uid、gid、mode，
并支持 unlimited 和 no_shrink。配额使用 `fs subvolumegroup resize`，取消限制生成
`inf`；保护开关实际禁止配额低于已用空间，不是禁止一切缩小。属性更新用
`fs subvolumegroup create` 的命名参数，省略 size，避免覆盖刚设置的配额；
属性操作前以 info 确认组存在，最后核对 bytes_quota、data_pool、uid、gid 和权限位。

表单从当前 info 字段预填，mode 去除目录类型位后显示八进制；缺失或超出浏览器安全
整数范围的配额不自动舍入预填。用户可仅更新配额，或显式选择同时更新属性。
数据池布局更新不迁移既有子卷，配额与属性分步执行，不承诺原子事务；确认框提示
后续步骤失败可能已有部分修改生效。离线测试覆盖命令顺序、更新前存在性检查、
逐字段读回、权限位、部分失败、严格 API 契约及版本检查；未在真实集群验证。

### CephFS 文件系统启用状态与创建时间

参考 `cephfs-list.component.ts` 显示 mdsmap.enabled 与 mdsmap.created。
本地 `MDSMap::dump` 在 `fs dump --format json` 输出这两个原生字段，现由采集器
保留为 enabled 与 created，经 reconciler、资源缓存及文件系统 GET/list 接口提供给
列表和详情页。enabled 使用可空布尔，不根据 MDS 数量或缓存 status 推测；false
明确显示未启用，缺失显示未知。created 保留原生字符串，与 CephTower 缓存记录的
created_at 分开，缺失时不填充当前时间。

离线测试走真实 NativeProvider → reconciler → store → API 链路，验证 enabled 的
true/false/缺失、Ceph 创建时间原值与本地缓存时间分离，并拒绝字符串伪布尔响应。
前端单元检查确认缺失和非布尔值不会误显示已启用；未在真实集群验证。

### CephFS 扩展访问权限

对照参考授权弹窗，`fs authorize` 表单新增 quota、snapshot 和 root squash。
读写权限可组合为 `rwp`、`rws` 或 `rwps`，root squash 作为独立 capability 参数；
只读模式忽略仅适用于写权限的 p/s 选择。离线测试覆盖完整组合和只读约束。

### CephFS 子卷快照可见性

对照参考 `cephfs-subvolume-form`、`cephfs-subvolume.service.ts` 和 controller 的
snapshot-visibility GET/PUT，子卷详情新增实时读取与允许快照浏览开关。
本项目接口为 `GET/PUT /filesystem/subvolume/snapshot/visibility`，请求明确携带
cluster_id、fs、subvolume、group；PUT 的 visible 必须为 JSON boolean。
调用 `ceph fs subvolume snapshot_visibility get/set`，默认组省略 --group_name，
命名组始终保留；设置后读回 0/1 并核对目标值，失败不报告成功。

本地参考 Dashboard 有此功能，但本地 volumes 模块未包含该 CLI 定义，因此另外
核对了[上游模块源码](https://github.com/ceph/ceph/blob/main/src/pybind/mgr/volumes/module.py)
及[原生命令文档](https://docs.ceph.com/en/latest/cephfs/fs-volumes/#controlling-subvolume-snapshot-visibility)。
开关只在客户端启用 client_respect_subvolume_snapshot_visibility 后生效，目前仅支持
FUSE/libcephfs；不会自动修改客户端配置，也不代表删除快照。页面明确说明限制，旧集群
命令不支持、输出异常或命令失败时展示未知及错误，不默认当作开启。
离线测试覆盖默认/命名组、布尔契约、非法作用域、失败读回和异步 API 全链路；
未在真实 Ceph 集群验证。

## CephFS 快照克隆目标组选择补齐

快照克隆表单的目标子卷组使用源快照所在文件系统的组列表，包含默认
`_nogroup`，并通过 `listAllResources` 读取全部分页。数据来自现有
`/filesystem/subvolume/groups` 接口的 `fs` 范围过滤及 CephFS 子卷组采集链路。
参考 Dashboard 的 `cephfs-subvolume-snapshots-list.component.ts` 同样为克隆
表单提供子卷组列表。通用表单选项加载支持当前行上下文，并在关闭、切换操作或
切换集群后忽略旧请求结果；加载失败会提示重试。

验证：前端测试覆盖源文件系统范围、默认组去重、缺少文件系统及请求失败，
`make test-frontend` 通过。尚未进行真实 Ceph 集群验证。

## CephFS 快照创建资源联动选择

创建快照表单按文件系统 → 子卷组 → 子卷选择资源，组列表包含 `_nogroup`，
子卷列表仅提供原生状态为 `complete` 的资源。切换上游选择会级联清空下游值；
提交时检查依赖选项的当前成员资格。依赖变化触发选项重载，旧请求结果被忽略。

沿用现有命令 `ceph fs subvolumegroup ls <fs> --format json` 和
`ceph fs subvolume ls <fs> [--group_name <group>] --format json` 的采集结果。
`/filesystem/subvolume/groups` 按文件系统过滤，`/filesystem/subvolumes` 同时按
`fs` 和 `group` 过滤；两者读取全部分页。参考快照列表组件的组、子卷选择流程。

验证包含默认组与命名组的请求范围、未选组时不请求、未就绪子卷排除、逆序定义
下的级联清空；`make test-frontend` 通过。尚未进行真实集群验证。

## CephFS 创建资源与克隆的数据池范围

子卷创建使用文件系统范围内的子卷组下拉选择；子卷组与子卷创建的数据池随
文件系统联动，快照克隆的数据池使用源快照的文件系统范围。参考子卷表单的
`this.pools.filter((pool) => pool.type === 'data')`：不能仅凭池的 CephFS 应用标签
判断它是否可供当前文件系统使用。

当前实现读取已有 `ceph fs dump --format json` 采集的 `data_pools`，与
`ceph osd pool ls detail --format json` 采集的池 `id` 匹配；文件系统与池列表均
读取全部分页。排除元数据池和其他文件系统的数据池，缺失映射时提供空选项。
离线测试覆盖池 ID 为零、多数据池、元数据池排除、跨文件系统排除和无文件系统
时不请求；`make test-frontend` 通过。尚未进行真实集群验证。

## CephFS 创建配额精确传递

子卷与子卷组创建接口的 `size` 使用十进制字符串，前端使用文本输入并保持
原值直到 Ceph argv。支持 `0`（不限制）及 1–9223372036854775807 字节，缺省
仍为不限制；拒绝数字类型、非规范整数字符串及越界值。OpenAPI 已同步替换
原数值契约。参考 volumes 模块两种创建命令的 `size` 均为 `CephInt`；子卷
零配额省略 `--size`，子卷组使用位置参数 `0`。

前端、请求契约和命令构造测试覆盖超过 JavaScript 安全整数的配额及边界，
`make test-backend`（含 OpenAPI 检查）和 `make test-frontend` 通过；未实测集群。

## 子卷组创建大小写敏感开关

参考 volumes 模块 `fs subvolumegroup create` 的 `casesensitive` 为可选
`CephBool`，控制器将显式 false 与未指定分别传递。修复创建表单关闭开关后
后端遗漏参数的问题：true/false 分别发送 `--casesensitive=true/false`，未指定
保持省略。命令构造测试覆盖三种状态及非布尔值拒绝；`make test-backend`
通过。未进行真实 Ceph 集群验证。

## 资源表单集群作用域

通用资源表单（包含 CephFS 创建、编辑和克隆）记录打开时的集群，切换集群
关闭表单并清空资源行与详情。提交前检查作用域；等待确认期间切换集群后即使
切回原集群，也不会继续旧提交。已经发出的请求保持原集群目标，返回结果不再
写入新集群的表单。`make test-frontend` 通过；未进行真实集群验证。

## 快照计划列表进入路径管理

已发现计划列表读取全部分页，并增加“管理路径与保留策略”入口，直接填入
计划的文件系统、路径、子卷和子卷组后查询实时状态。切换到普通路径会显式
清空之前的子卷与组，避免表单合并保留旧范围；列表增加子卷组展示以区分同名子卷。
参考快照计划列表通过选中行进入编辑/保留策略管理的流程。

复用现有实时状态接口及保留策略 API，后端使用
`ceph fs snap-schedule retention add/remove <path> <retention> --fs=<fs>`，
并携带子卷、组范围。前端测试覆盖路径范围传递与清空，`make test-frontend`
通过；未进行真实集群验证。

## 快照计划页面按集群隔离

快照计划页面以集群 ID 作为组件身份，切换集群重新初始化表单、模块状态、
已发现列表、查询范围与保留策略。旧页面卸载后忽略模块和列表响应，停止后续
刷新查询；异步表单校验结束后也检查页面是否仍有效，避免离开后发送创建请求。
已发出的操作仍针对原集群执行。`make test-frontend` 通过；未实测集群。

## 快照周期与保留规则可读展示

参考 Dashboard 快照计划服务的 `parseScheduleCopy` 和 `parseRetentionCopy`，
已发现列表与实时查询列表均显示中文周期、各时间粒度的保留数量及最近快照数。
周期保留原始表达式；区分分钟 `m` 与月份 `M`。策略兼容原生对象和规则字符串
两种数据形式，缺失数据为未知，空策略为未设置；异常数值不伪装成有效规则。
测试覆盖单位、组合规则、空/未知数据和大整数精度，`make test-frontend` 通过。

## 快照计划启停状态判定

参考 snap_schedule 的 Schedule 对象将 `active` 转为布尔值。列表与实时查询
严格识别 true/false，缺失、字符串或其他异常值显示未知并禁止推断启停动作。
删除与路径管理仍可使用。测试覆盖布尔状态及异常输入，`make test-frontend`
通过；未进行真实集群验证。

## 快照计划创建周期选择

创建表单采用正整数字符串间隔与单位下拉框，默认每日一次，生成原生命令所需的
`schedule` 表达式。小时、日、周、月、年可选，月与年注明 30/365 天。
参考前端 RepeatFrequency 枚举将年写为 `Y`，但同版本原生
`snap_schedule/fs/schedule.py` 的 repeat 解析只接受 `y`，本实现遵循原生命令。
分钟级周期仅用于开发测试，不列入创建选项；已有分钟级计划仍可展示。
测试覆盖周期构造、非法间隔和单位，`make test-frontend` 通过；未实测集群。

## 快照计划 API 周期校验

创建与启停/删除计划的命令构造验证正整数间隔及原生 `m/h/d/w/M/y` 单位，
拒绝零间隔、小数、复合周期及错误大小写单位。已有分钟级计划仍可管理；
原生支持的带前导零正整数表达式原样传递。命令测试覆盖所有单位与无效输入，
`make test-backend`（含 OpenAPI 检查）通过；未实测集群。

## 快照计划采集时效展示

已发现计划列表展示列表 API 的 observedAt、stale 和 staleReason；多页数据使用
listAllResources 汇总的采集状态。过期列表显示警告及刷新/实时路径查询提示，
加载失败保留旧列表的采集信息。`make test-frontend` 通过；未实测集群。

## 快照保留策略数量表单

保留策略按分钟、小时、日、周、月、年和最近快照分别填写正整数数量，支持
多单位组合及提交前可读预览。留空单位不参与添加或移除，非法数量禁止提交；
切换查询范围清空待编辑规则。生成原生 retention 字符串传给已有添加/移除 API，
不经过浮点数转换。测试覆盖组合、空值、大小写单位及精确大整数，
`make test-frontend` 通过；未实测集群。

## 使用现有保留规则填写移除表单

原生 Schedule.rm_retention 要求单位和数量均匹配，add_retention 拒绝覆盖已有
单位。页面明确说明先移除再添加，并提供“填入当前规则”，使用当前路径实时
查询的 retention；缺失、重复单位、未知单位或不安全数值不自动填入。用户可以
清空不参与移除的单位。测试覆盖对象/字符串规则解析与异常数据，
`make test-frontend` 通过；未实测集群。

## 快照计划子卷组范围校验

原生 snap_schedule 使用子卷与组共同解析路径；仅提供组不能定位子卷。创建与
计划动作现在同保留策略一样拒绝“有组、无子卷”的请求，查询表单也在提交前
提示。命令构造测试覆盖缺失子卷拒绝与完整范围接受，前后端检查通过；未实测集群。

## 快照计划文件系统选择

路径查询与创建计划使用当前集群文件系统下拉框，读取已有 `/filesystems`
接口全部分页，支持名称搜索及失败重试。切换文件系统重置路径为根目录，并
清空子卷、组与待编辑保留规则；来自已发现计划的范围仍可直接填入查询。
复用 fs dump 采集链路，`make test-frontend` 通过；未实测集群。

## 保留策略写入读回验证

保留策略添加/移除后，使用已有带路径及子卷范围的 snap-schedule status 命令
验证所有返回计划的 retention。添加规则要求原生整数数量精确匹配，移除规则
要求单位消失。空结果、缺失策略、规则不一致或异常 JSON 均返回 post_check_failed，
说明变更可能已经生效。执行链测试覆盖成功与失败读回；`make test-backend`
及 OpenAPI 检查通过，未实测集群。

## 快照计划列表与实时结果同步

从已发现列表启停或删除计划后，如果下方实时结果正在查看同一文件系统、路径、
子卷和组，则重新查询该范围；其他范围保持当前查询。默认组省略与 `_nogroup`
视为相同范围，同名子卷的不同组不会混淆。测试覆盖各范围差异，
`make test-frontend` 通过；未实测集群。

## NFS 导出原生字段展示

`nfs export ls <cluster> --detailed --format json` 的原始导出数据直接经资源 API
提供。列表改用 cluster_id、access_type、fsal.name/fs_name/user_id，并补充
protocols 与 transports；参考 NFS 列表相同字段。CephFS 导出编辑初始值正确
映射 RO/RW 与文件系统。当前简化表单不支持 RGW 或其他访问类型，阻止其误用，
完整 FSAL 编辑仍待补齐。字段映射测试及 `make test-frontend` 通过；未实测集群。

## NFS 导出删除身份解析

前端编辑/删除使用采集器的 natural_key（NFS 集群与导出 ID 的不透明组合），
不再把单独数字 export_id 用作资源键。原生 `nfs export rm` 要求伪路径，删除前
通过同集群的详细导出列表按 ID 唯一定位 pseudo，再执行删除。缺失、重复、
集群不匹配或非法伪路径均停止操作。执行链测试覆盖解析与禁止误删，前后端
完整检查通过；未实测集群。

## NFS 导出更新保留完整属性

更新前使用同集群原生详细列表按不透明资源键定位导出，在完整 JSON 上修改
pseudo、path、fsal.fs_name 和显式 read_only 对应的 access_type，保留 export_id、
客户端规则、协议、传输、squash 和其余 FSAL 属性。原生 apply 通过 export_id
识别伪路径重命名，避免误创建另一个导出。前后端禁止跨 NFS 集群移动；后端
同样拒绝简化表单不支持的 FSAL 和访问类型。执行测试验证属性保留和跨集群拒绝，
前后端完整检查通过；未实测集群。

## NFS 导出更新读回验证

原生 AppliedExportResults 通过非零退出码报告 apply 失败，现有执行器处理该错误。
更新成功后进一步读取详细导出列表，按原集群和 export_id 验证 pseudo、path、
CephFS 名称及显式访问类型。缺失、原值未变或只有其他 ID 匹配时返回
post_check_failed，并说明操作可能已生效。执行链测试及 `make test-backend`
通过，未实测集群。

## NFS 导出 squash 设置

创建/编辑表单提供 root_squash、root_id_squash、all_squash、no_root_squash，
列表展示原生 squash。API 枚举与命令构造双重验证，apply JSON 携带显式选择，
更新读回比对请求策略。未指定时创建采用原生默认值，更新保留现有策略及客户端
级规则；已有其他原生别名不自动改写。参考 ganesha_conf.py 的 squash 校验及
Dashboard NFS 表单。前后端测试和 OpenAPI 检查通过，未实测集群。

## NFS 导出资源选择

创建导出从 `/nfs/clusters` 选择 NFS 集群，创建及编辑从 `/filesystems` 选择
CephFS 文件系统；读取全部分页，选项按当前 Ceph 集群隔离。复用原生
`nfs cluster ls` 与 `fs dump` 的采集数据和通用选项加载错误提示。
`make test-frontend` 通过；未实测集群。

## NFS 导出删除后验证

删除后读取同集群详细导出列表，验证原 export_id 已消失。有效空列表表示没有
剩余导出；null、错误结构、错误集群、非法 ID 或原 ID 仍存在均视为无法验证，
返回 post_check_failed 并说明删除可能已经生效。命令链与读回测试、
`make test-backend` 和 OpenAPI 检查通过；未实测集群。

## NFS 导出与客户端规则详情

参考 NFS details 的配置与客户端页，详情增加结构化 FSAL、协议、传输、安全
类型、安全标签、访问类型和 squash 展示，以及客户端地址/网段与访问规则表。
客户端未设置的属性显示继承导出配置；缺失或异常客户端列表不当作空规则，
安全标签仅识别原生布尔值。使用现有详细导出采集数据，`make test-frontend`
通过；未实测集群。

## NFS 导出路径校验

依据原生 Export.validate，创建与编辑的 pseudo 必须为绝对路径且不能是根目录；
CephFS 数据路径要求绝对路径，允许根目录。前端即时提示，后端命令构造再次
校验，测试覆盖相对路径、根伪路径及有效根数据路径。前后端检查通过；未实测集群。

## NFS 导出安全标签配置

创建与编辑提供安全标签启用/禁用选项，经 API 布尔契约传入 apply JSON，明确
保留 false。未填写时创建沿用原生默认，编辑保留现有值；更新读回核对显式请求。
依据 ganesha_conf.py 的 security_label 布尔校验，测试覆盖 true、false 与错误
字符串类型，前后端及 OpenAPI 检查通过；未实测集群。

## NFS 新建导出伪路径预检查

原生 apply 会按伪路径查找已有导出并进入更新，因此新建操作执行前读取详细
列表，拒绝已存在或无法验证的伪路径，比较时规范化目录表达式。测试确认冲突、
异常列表与跨集群结果不会触发写命令，`make test-backend` 通过。该预检查与
apply 之间仍存在外部并发窗口，尚不声称提供 Ceph 侧原子创建保证；未实测集群。

## NFS 新建导出读回验证

创建后读取详细列表，按伪路径唯一匹配具有有效 ID 的导出，检查集群、CephFS、
目录、显式访问类型、安全标签和 squash。路径比较按原生规范化语义处理；
更新读回也采用同样的路径比较。缺失或不一致返回 post_check_failed。创建执行链
测试覆盖成功、缺失、路径错误与集群不匹配，后端完整检查通过；未实测集群。

## NFS 导出协议版本配置

创建和编辑支持 NFSv3、NFSv4 或两者，通过 protocols 整数数组交给
`nfs export apply`。依据 nfs/ganesha_conf.py 校验，仅接受非空且不重复的
3/4 集合；省略字段保留原生默认或现有设置。创建和更新读回按集合检查，
顺序不同不误报，缺失或协议不一致拒绝确认成功。无真实集群验证。

## NFS 导出传输配置

参考 Dashboard NFS 表单的 TCP/UDP 选择，创建和编辑支持 TCP、UDP 或两者。
API 使用 transports 字符串数组，写入 `nfs export apply` 的原生 JSON；
只接受非空、不重复的大写 TCP/UDP，省略时使用原生默认或保留现有配置。
创建和更新读回按集合验证传输配置。测试基于原生结构，无真实集群验证。

## NFS 导出访问类型

用 access_type 枚举替换旧 read_only 布尔接口及表单开关，不保留旧参数别名。
创建和编辑可选 RO、RW、NONE，与 nfs/ganesha_conf.py 的原生访问类型一致。
NONE 表示导出默认拒绝访问，客户端规则仍可覆盖；更新保留现有客户端规则。
原生 apply JSON 与读回校验均使用 access_type，允许编辑已有 NONE 导出。
省略字段时保留原生默认或当前设置；无真实集群验证。

## NFS 导出认证方式

创建和编辑支持原生 sectype 数组：none、sys、krb5、krb5i、krb5p。
表单逗号分隔输入，留空保留设置，default 显式提交空数组恢复原生默认。
后端校验枚举与重复值，通过 export apply 写入，读回按集合比较；原生
to_dict 在空数组时省略 sectype，因此恢复默认允许读回缺失字段。
Kerberos 需预先配置服务端，本功能不创建 Kerberos 环境。未实测集群。

## NFS 客户端规则写入

创建和编辑提供客户端规则编辑区，支持地址数组、访问类型、身份映射，
不提交字段则不修改，空数组清空规则。API 字段名为 client_rules，后端转换为原生 clients，
避免与 iSCSI clients 契约混淆。空 access_type/squash 继承导出设置；原生 null
读回与空字符串按继承语义比较。后端校验结构和枚举，拒绝地址中的配置分隔符、
引号与控制字符，通过 export apply 写入并按原顺序读回验证。当前地址检查是
配置安全字符检查，不替代 DNS/IP 可达性检测。
参考 nfs-form-client、ganesha_conf.py Client；无真实集群验证。

## NFS 客户端逐行编辑

客户端表单改为规则卡片，支持逐行新增/删除、逗号分隔地址、访问类型与身份映射
选择，并明确标注继承语义及删除全部规则的访问范围风险。提交前去除地址两端
空格，未知原生身份映射值保留显示，不静默改写。异常规则结构显示错误而非清空。
沿用 client_rules API 与原生命令链路；离线验证，不声称真实集群验证。

## NFS RGW 导出

创建可选 CEPH/RGW，RGW 提交用户 ID 和桶路径（/ 表示用户根目录），CephFS
仍要求文件系统和绝对路径。编辑禁止切换 FSAL，允许调整 RGW 用户和路径；
后端原生详细列表预读保留凭据，经 export apply 交给 Ceph 管理模块处理，
表单不接收密钥。读回检查 FSAL、用户和路径，发现数据继续使用既有脱敏链路。
参考 nfs/export.py 的 _apply_export 与 _create_rgw_export_user；未实测集群。

## NFS 路径与配置字符串边界

参考 ganesha_conf.py 的 _format_val 对字符串直接加双引号而不转义，因此
JSON stdin 本身不能阻止后续 Ganesha 配置注入。旧通用标识符校验同时误拒绝
中文及带空格路径，现为路径与伪路径采用专门校验（最多 4096 字节）。创建/编辑在执行任何命令前
拒绝 cluster、pseudo、path、filesystem、rgw_user_id 中的双引号、反斜杠和
控制字符；路径保留中文、普通空格、单引号及引号内安全的分号等字符，集群、文件系统与用户标识符仍使用原有限制。
测试覆盖 CephFS/RGW 参数及执行器零调用，不依赖真实集群。

## NFS RGW 用户选择

创建/编辑 RGW 导出从当前 Ceph 集群的 `/rgw/users` 全分页结果选择用户，
使用采集器 user list/info 链路保留的完整 uid，显示名称仅用于标签。
切换存储后端会清空依赖的用户选择，CephFS 模式不请求 RGW 用户列表。
后端 NFS 用户校验支持 tenant$user，同时拒绝选项前缀及配置危险字符。
桶路径仍手工输入；用户列表依赖现有采集数据，未实测集群。

## NFS RGW 导出范围

表单区分用户根目录与单桶模式。用户模式固定路径 / 并选择用户；桶模式只提交
桶路径，不发送隐藏的旧用户值，Ceph 根据 bucket stats 的 owner 读取用户凭据。
更换桶时清除预读 FSAL 中的旧用户及密钥，触发原生拥有者重新解析；同桶更新
保留已有身份与凭据，不改变现有授权。读回必须
包含已解析用户，根目录没有用户时拒绝执行。参考 Dashboard 的 setUsers/setBucket
与 nfs/export.py；桶路径仍手工输入，无真实集群验证。

## NFS RGW 桶选择

单桶表单改为当前集群 `/rgw/buckets` 全分页选择，使用 bucket 与 tenant 字段，
不使用不透明 natural_key。选择值编码租户和桶名，避免同名桶混淆。
用户根目录模式不加载桶列表，切换范围清除旧桶选择。
沿用 bucket list/stats 采集和 NFS apply 链路，无真实集群验证。

## NFS 租户桶拥有者解析

桶表单通过 rgw_bucket_tenant 显式传递租户（空字符串表示无租户）。后端先执行
`radosgw-admin bucket stats --bucket <name> --tenant <tenant>`，严格核对返回的
桶名、租户和完整拥有者 UID，再传 user_id 给 NFS apply，绕过原生 NFS 缺失
--tenant 的自动查询。无租户不传该命令参数。创建/编辑读回核对已解析 UID，
同名桶跨租户更新不会保留旧身份。源代码 rgw_lib.h 用用户身份确定 bucket_tenant。
只支持可解析为同租户用户的拥有者；账户拥有者尚需独立处理，未实测集群。

## NFS CephFS 安全标签扩展属性

创建、编辑与详情支持 FSAL 的 sec_label_xattr，对应参考 Dashboard 的同名表单及
`nfs/ganesha_conf.py` 的 CephFSFSAL 序列化。API 可选字符串通过
`ceph nfs export apply <cluster> -i -` 写入嵌套 FSAL；省略保持已有值，空字符串
清空配置，原生详细列表省略空属性时视为清空成功，非空值必须精确读回。
仅 CephFS 可用，属性名限制为最多 255 个 ASCII 字母、数字、点、下划线和连字符，
防止原生配置未转义字符串注入；不自动启用 security_label，也不修改 CephX 身份。
已覆盖 payload、保留/清空、读回不一致、非法值与前端转换测试；未实测集群。

## NFS CephFS 挂载根路径

创建、编辑与详情提供 cmount_path，映射到原生 apply JSON 的 FSAL 字段。
创建省略时由 Ceph 默认使用 /，编辑省略时保留原值；显式 / 可恢复根目录。
后端规范化路径并按目录边界检查其包含导出路径，拒绝相对路径、越界及配置注入。
即使编辑没有提交挂载根路径，也校验已有根路径是否包含新的导出路径。
参考 nfs/export.py 的 create_export_from_dict 和 get_user_id，修改文件系统或
挂载根路径时移除旧 user_id/cephx_key，让 Ceph 生成新身份；其他 FSAL 配置保留。
原生模块可能重启 NFS 服务，操作需注意客户端影响。读回校验挂载路径；覆盖路径、
身份重建及前端转换测试。未连接实际集群验证授权和服务重启行为。

## NFS 集群端点信息

集群名称列表之外，逐集群执行 `ceph nfs cluster info <name> --format json`，
按返回对象的集群键取值，采集 VIP、入口/监控端口、入口模式和后端主机/IP/端口。
数据随现有 nfs_cluster 资源由 `/nfs/clusters` API 提供；页面新增结构化端点详情，
移除没有采集来源的放置策略列。字段缺失不猜测默认端口，查询失败仍保留集群，
以 info_available=false 明示端点信息未获取。源自 nfs/cluster.py 的
show_nfs_cluster_info；已测试命令参数、IPv6、多种缺失/异常返回，未实测集群。

端点响应进一步校验 IP、主机名类型及 1–65535 整数端口；允许原生未报告的空端口，
但拒绝错误类型、无效 IP 和不完整后端条目，避免误标为已获取。
只保留已知端点字段，不向资源 API 转发原生响应中的额外属性。

## NFS 集群创建放置策略

创建表单支持可选 nfs_placement 字符串，通过单个 `--placement=<expression>` 参数
传给 `ceph nfs cluster create`，支持原生主机、标签及数量表达式。对应 nfs/module.py
的 placement 参数，由编排器验证具体语法；本地限制空值、控制字符和超过 1024 字节。
省略时沿用 Ceph 默认策略，不将文本拆成命令参数或交给 shell。使用独立字段避免
与 service API 的对象型 placement 混淆。创建命令成功不代表守护进程已就绪，
需查看集群端点及服务状态。命令构造与非法输入已测试，无实际集群部署验证。

## NFS 集群入口部署

创建支持 ingress、virtual_ip、ingress_mode 和 nfs_port，分别映射原生 create 的
--ingress、--virtual-ip、--ingress-mode、--port。关闭入口时表单不提交隐藏 VIP/模式。
VIP 支持 IPv4/IPv6 和 CIDR；入口必须有 VIP，VIP/入口模式不能脱离入口单独设置。
端口必须是 1–65535 整数；依照 nfs/cluster.py 的派生端口逻辑，HAProxy 模式上限
55535，keepalive-only 上限 58535。省略使用原生默认值。入口模式与 orchestrator
IngressType 对齐，keepalive-only 原生强制单实例，界面明确提示。
已覆盖命令参数、依赖校验、端口边界和隐藏字段测试；未实测入口部署或故障切换。

NFS 集群创建/删除现在严格解析 `cluster ls` 读回：创建要求目标存在，删除要求目标
消失；拒绝 null、非字符串列表、重复名称及尾随 JSON。无法验证时返回
post_check_failed，提示操作可能已经生效，不把查询命令成功当成变更成功。
此检查仅确认集群存在性，不证明部署参数已生效或守护进程就绪。

创建前额外执行只读 `cluster ls`，只有能确认名称不存在时才执行创建；已有名称
或异常列表均阻止写入。参考原生 create_nfs_cluster：已有集群不会应用新的部署
参数，因此不能用创建表单更新现有配置。预查与创建不是原子操作，仍可能与外部
创建并发；当前不宣称该检查能证明所有部署参数已生效。

删除确认明确列出原生 delete_nfs_cluster 的级联影响：全部导出、配置和 NFS/入口
服务会一并移除，提醒先迁移客户端并保存配置。通用删除组件支持资源专属提示，
保留高风险确认、版本校验及原有执行链路。提示文本与接入点已回归测试。

## SMB 集群原生配置展示

根据参考 Dashboard SMBCluster.list/get 的 smb show 调用，名称采集后逐集群执行
`ceph smb show ceph.smb.cluster.<id> --format json`，核对资源类型和集群 ID。
经现有 `/smb/clusters` 资源 API 提供认证模式、期望状态、域设置/凭据引用、用户组
引用、DNS、公开地址、放置及集群模式。详情展示这些原生字段，不补造缺失值；
失败保留集群名称并标记 info_available=false。编辑不再把未知认证模式默认成 user。
不查询 join auth 或 usersgroups 密码资源，响应额外顶层字段不转发；沿用资源脱敏。
已覆盖原生命令、身份不匹配及配置读取测试，尚未真实部署 SMB 验证。

## SMB 共享原生列表

修正将 `smb share ls` 的字符串 ID 列表误当对象的问题，改用
`smb show ceph.smb.share.<cluster> --results=full --format json`，与 Dashboard 的
show 数据源一致。严格核对 resource_type、cluster_id、share_id 和重复 ID；异常
标记采集不可用，不将异常当作空列表删除已有资源。保留嵌套 cephfs 配置，前端从中
展示文件系统、路径、子卷/组及原生 readonly/browseable（未知不当成 false）。
编辑初值使用原生字段，资源定位统一使用包含集群和共享 ID 的 natural_key；更新
解析该身份并拒绝跨集群移动。已覆盖列表命令、字段和身份测试，未实测 SMB 集群。

## SMB 共享更新保留配置

原生 share create 为 create_only，不能用于编辑。更新改为先读取目标 show 资源，
严格核对身份及 CephFS 对象，只覆盖提交的文件系统和路径，保留名称、只读、
可浏览、子卷/组、provider 等配置，再以 JSON stdin 执行 `smb apply -i -`。
省略路径保持原值，显式空路径拒绝。读回逐项核对提交资源的全部字段，不一致报
post_check_failed，避免以命令接受代替更新完成。覆盖保留配置、错误身份及读回
不一致的执行链路测试；无实际 SMB 部署验证。

SMB 共享编辑支持 readonly/browseable，对照参考 smb-share-form 的布尔字段。
API 严格接受可选布尔值，省略保留；前端用显式选择区分 false 与未知，不把未采集
的设置自动覆盖为默认值。通过上述 apply 链路写入并读回验证。隐藏共享仅控制
浏览可见性，不替代访问控制，表单明确提示。创建表单的可浏览选项尚待接入。

SMB 共享创建从当前 Ceph 集群 `/smb/clusters` 全分页选择集群，创建/编辑从
`/filesystems` 选择文件系统，对齐参考表单 filesystems 数据源。编辑集群只读，
与后端禁止跨集群移动一致。沿用现有 cluster ls、fs dump 采集及表单加载错误处理，
不新增命令或兼容字段。已测试选择器请求作用域，未实测集群。

SMB 共享 ID 与客户端名称分离：创建的 `name` 参数仍表示不可变 share_id，新增
`share_name` 对应原生 `smb share create --share-name`；编辑通过 show/apply 更新
资源 `name`，不更改 ID。名称验证依据参考 smb/validation.py（1–64 个 ASCII
字母、数字、空格、点、下划线、短横线，首字符限字母数字下划线）。省略创建名称
使用原生 ID 默认值，省略编辑名称保留原值，显式空字符串 API 拒绝。界面明确区分
两个字段，编辑初值来自原生 name，回读不一致拒绝报告成功。已做命令与 fixture
测试，尚未实测客户端重连或 SMB 名称修改的集群行为。

SMB 创建支持原生 readonly：对照参考 Dashboard 表单的默认 false 和 module.py
share_create 布尔参数，API 接受严格布尔值，true 添加 `--readonly`，false 或省略
使用原生命令的 false 默认值。前端显式选择只读或允许写入，将字符串选项转换为
JSON 布尔值。命令测试覆盖 true、false、省略及非法类型，未实测 SMB 挂载写入。

SMB 创建/删除后严格解析目标集群的 `smb share ls --format json` 字符串列表，
分别确认共享 ID 存在/消失；删除身份从资源键解析，不信任额外名称参数。null、
错误结构、空 ID、重复 ID、尾随 JSON 或未达到目标状态均返回 post_check_failed，
提醒操作可能已经生效。此检查证明资源存在性，不证明 Samba 已部署或客户端可用；
创建属性的完整回读仍待完善。已覆盖执行链路与异常 fixture，无真实集群验证。

SMB 创建支持指定 CephFS 子卷：API 可选 subvolume 字符串映射原生命令
`smb share create --subvolume=<name>` 或 `<group>/<name>`，由 CephFSStorage
拆分为 subvolumegroup/subvolume，对应参考表单的两个存储字段。省略仍使用文件系统
根范围；路径在选定子卷范围内解释。拒绝空值、多层路径、点目录、NUL 和换行。
已覆盖原生命令参数测试，尚未进行实际子卷共享验证。

SMB 创建的子卷输入改为文件系统 → 子卷组 → 子卷联动选择，复用现有全分页
`/filesystem/subvolume/groups`、`/filesystem/subvolumes` API 与原生 CephFS 采集。
子卷组包含默认组 `_nogroup`，子卷过滤不可用状态；依赖变化沿用通用表单清空
下游选择，提交组合为原生命令 group/name 参数。组与子卷均未选择时使用文件系统范围。
测试覆盖请求集群/文件系统/组作用域、默认组、空依赖和参数组合。编辑存储范围
仍待完善；未实测 SMB 或浏览器交互。

修正只选择子卷组时静默丢弃组信息的问题：已选组必须补选子卷，否则提交构造
直接报错，不发送创建请求；提示清空组才能使用文件系统范围。参考 smb/fs.py
在存在组或子卷时调用 fs subvolume getpath，并无组根共享的独立解析分支。
测试同时覆盖普通组和默认组的不完整选择，避免依赖切换后扩大共享范围。

SMB 集群编辑修正为 show → apply → show：参考 module.py 的 cluster_create
明确设置 create_only，不能更新已有资源。更新预读严格检查资源类型和 cluster_id，
仅覆盖 auth_mode，保留 DNS、placement、认证引用等全部原生字段，以 JSON stdin
提交并检查 success，再逐字段校验回读。认证模式切换依赖匹配的域/用户配置，
本轮未添加这些编辑字段，不会自动删除旧认证配置；不满足原生约束会由 Ceph 拒绝。
已测试配置保留、错误身份、失败 apply 和丢失字段；未实测集群。

SMB 集群编辑接入 custom_dns 字符串数组，对照参考集群表单同名字段。界面支持
换行或逗号分隔，加载原生列表；未知字段不自动清空，编辑为空明确发送 []。
API 省略时保留，后端验证每项为 IPv4/IPv6 地址，使用已有 show/apply/show 链路
校验配置。测试覆盖双栈、清空、非法输入和回读不一致。创建 DNS 选项仍待接入，
未实测容器 DNS 生效或真实集群。

SMB 集群编辑支持已有用户组资源引用 user_group_ref，转换为原生
user_group_settings 的 resource/ref 列表。要求本地用户模式、非空、唯一且符合
原生 SMB ID 规则；省略保留。显式提交引用时移除旧 domain_settings，使切换到
本地用户模式得到一致配置，界面提示此影响；不删除域凭据资源。回读检查引用及
域配置移除。前端文本输入已有 ID，不加载密码；资源创建和引用库存选择仍待补齐。
引用是否存在由原生 apply 验证，测试使用 fixture，未实测认证切换。

SMB 集群编辑支持 domain_realm/domain_join_ref，成对提交并映射为原生
domain_settings.realm/join_sources。要求 AD 模式、非空域名与唯一有效引用；
省略两项保留原值。显式提交时移除不兼容的 user_group_settings，回读检查域配置
和本地来源移除，不删除凭据或用户组资源。前端按认证模式显示字段，从原生配置
加载初值，仅提交引用 ID。已覆盖参数转换、缺失/非法引用、模式冲突及回读测试；
凭据资源管理与库存选择仍待补齐，未实测域加入。

SMB 集群编辑支持实例数量 count，映射参考表单的 placement.count。省略保留原值，
显式设置要求正整数；只更新 count，保留 placement 中的 hosts、label、host_pattern
等约束，原 placement 结构异常时拒绝覆盖。通过 apply 后回读校验配置，而非声称
相应数量的守护进程已经部署成功。前端加载 count 初值并校验整数，测试覆盖约束
保留、非法数量及旧值回读；创建数量和其他放置字段仍待接入，未实测部署。

SMB 本地用户模式创建补齐必需的用户组资源引用：表单提交 user_group_ref 数组，
后端校验非空、唯一、原生 ID 规则，逐项生成 `--user-group-ref=<id>`，由原生
cluster_create 构造 user_group_settings。不使用用户名/密码命令参数。缺失引用
在执行前拒绝，AD 模式不能混入本地来源。已有资源是否存在由 Ceph 验证；已覆盖
原生命令和错误输入测试，AD 创建配置、凭据资源管理仍待完善，未实测创建。

SMB AD 创建接入 domain_realm/domain_join_ref：表单按模式要求填写域名与已有
凭据资源 ID，复用编辑表单的引用转换；后端生成原生 --domain-realm 和重复
--domain-join-ref 参数。AD 模式必须提供非空域名和唯一有效引用，本地模式拒绝
域字段，AD 模式拒绝本地用户组来源。测试覆盖命令、非法/空字段与混合来源。
未创建或读取凭据明文，凭据资源管理仍待实现，未实测域加入或部署。

SMB 集群创建补齐 custom_dns：表单沿用编辑时的换行/逗号解析，API 字符串数组
经共用 IPv4/IPv6 校验后逐项生成 --custom-dns 参数；省略或空列表不传参数，
保留原生命令默认行为。测试覆盖双栈、空列表和非法输入，未实测容器解析域名。

SMB 集群创建支持可选实例数量 count，校验正整数并生成
`--placement=count:<n>`，由原生 PlacementSpec.from_string 解析。前端复用编辑
数量转换；省略不传 placement，不猜测默认数量。测试覆盖命令参数、非法数量及
省略行为。其他创建放置约束仍待补齐，未实测守护进程部署数量。

SMB 创建支持 smb_hosts 主机名数组，与 count 合并为单个 --placement 参数，
对应参考表单 placement.hosts。省略保留自动放置；显式列表要求非空、唯一普通
主机名，拒绝标签、通配符、空白和纯数字，避免被 PlacementSpec 解析为其他语义。
前端以换行/逗号文本输入，库存选择仍待补齐；主机存在性由 Ceph 校验。测试覆盖
主机单独设置、与数量组合及非法语法，未实测部署。

SMB 创建主机输入改为当前集群 /hosts 全分页多选列表，使用原生 hostname，去重
并忽略缺少主机名的记录。通用资源表单增加多选模式，校验每项属于当前选项范围；
复用集群切换关闭表单和异步加载隔离。提交直接发送字符串数组，清空表示不指定
主机。测试覆盖请求集群、去重和数组转换；未进行浏览器或真实集群验证。

SMB 集群创建/删除后解析原生 `smb cluster ls --format json` 的 ID 数组，分别确认
目标存在或消失；删除以资源路径中的 ID 为准，不信任请求体名称。空值、非法类型、
重复 ID、空 ID 和尾随 JSON 均不能作为成功证据。未达目标状态返回 post_check_failed，
提示操作可能已经生效。执行链路测试覆盖正常和异常回读；这仅验证资源登记状态，
不证明 SMB 服务部署就绪，也未进行真实集群验证。

SMB 集群编辑支持显式替换部署主机：界面开关默认关闭，开启后从当前集群库存选择
非空主机列表，经 PATCH 的 smb_hosts 字段进入原生 smb apply 配置。原生 PlacementSpec
不允许 hosts 与 label 共存，因此替换时移除 label 和 host_pattern，保留 count 等其他
配置；未提交该字段则完全保留原部署配置。沿用更新前读取、更新后完整配置回读校验。
测试覆盖互斥约束替换、数量保留、非法主机及前端显式选择；未实测真实服务重新部署。

SMB 创建和编辑表单增加 CTDB clustering 模式，匹配参考枚举 default/always/never。
创建通过 `smb cluster create --clustering=...`，编辑在保留原配置的基础上通过
`smb apply` 更新 clustering 并回读校验。API 只允许三个原生枚举值，字段省略时
创建使用原生默认、编辑保留原值。测试覆盖三种模式、非法值和前端初始化/请求转换；
尚未验证真实 CTDB 服务行为。

SMB 认证模式切换在原生配置预读之后、smb apply 之前校验新模式必需的认证来源：
切换为 AD 必须同时提交域名和域加入引用，切换为本地用户必须提交用户组引用。
仅提交模式的请求不会执行变更命令；保持模式不变时仍允许保留既有认证设置。
执行链路测试确认两个方向的不完整切换仅触发只读预检查。

SMB 共享列表和编辑增加原生 comment 描述。现有原生资源采集已保留此字段，PATCH
新增可选字符串契约，通过 smb apply 修改并回读比对。省略保留已有描述，空字符串
清空描述；拒绝换行、回车和 NUL。参考资源模型支持该字段，参考共享表单尚未发现
对应输入。测试覆盖中文、清空、保留及非法值；未进行真实 Samba 客户端显示验证。

SMB 共享编辑暂未提供子卷范围替换，因此后端现在拒绝在保留已有 subvolume 或
subvolumegroup 的同时更换 filesystem，避免把旧范围解释为另一文件系统下的同名
资源或错误路径。原生 smb/fs.py 将 volume/group/subvolume 联合用于 getpath。
同文件系统内编辑及无子卷约束的文件系统更换不受影响。执行链路测试确认拒绝发生在
只读预检查之后、变更之前；完整的子卷范围编辑仍待补齐。

SMB 共享编辑现已增加显式存储范围操作：保留、替换为指定子卷、移除子卷限制。
子卷选项按当前 Ceph 集群、文件系统和子卷组联动加载。PATCH subvolume 省略表示
保留，group/name 表示替换，空字符串表示文件系统路径；替换必须明确提供 path。
后端拆分原生 cephfs.subvolume/subvolumegroup，保留 provider 和其他共享配置，支持
显式跨文件系统切换。回读验证新范围以及旧组移除；未指定新范围时仍拒绝跨文件系统
沿用旧子卷。测试覆盖执行链路、三种范围、非法输入及前端转换；未实测真实集群。

SMB 共享创建和更新路径按参考 validation.normalize_path 的 POSIX 语义规范化，
包括重复斜杠、点路径以及恰好两个前导斜杠。更新提交规范化后的配置，避免 Ceph
规范化后回读与原始输入不一致而误报失败。拒绝规范化后仍含相对父目录的路径，
以及 NUL/换行。测试覆盖创建命令、更新配置和路径边界；未验证真实文件系统路径存在性。

SMB 集群创建增加按主机标签部署，标签由当前集群 /hosts 全分页库存的 labels
汇总去重后选择。API smb_label 与 smb_hosts 互斥，通过原生 --placement 中的
label:<标签> 与 count 组合执行。拒绝空标签及空白、逗号等 placement 分隔语法。
测试覆盖库存作用域、去重、命令组合和互斥；标签编辑及真实部署验证仍待补齐。

SMB 集群编辑现支持显式替换部署标签，使用当前集群标签库存。开启标签替换会移除
已有 hosts/host_pattern，保留 count 等配置；与主机替换开关及 API smb_hosts 互斥。
未开启开关不提交标签变更。更新经原生 smb apply 并回读完整配置，测试确认旧主机
残留不能作为成功结果；覆盖执行链路、非法值和前端开关，未实测真实调度迁移。

SMB 创建集群/共享 ID 的前后端校验对齐原生 validation.check_id：1–18 个 ASCII
字母、数字或连字符，首尾必须为字母或数字。共享更新同样校验路径解码所得 ID 和
父集群 ID。客户端共享显示名称继续使用独立的 64 字符规则，不与资源 ID 混淆。
测试覆盖边界长度、字符和非法父集群；未进行真实集群验证。

通用资源表单的可选下拉框增加清除入口，包括 SMB 创建标签、可选子卷组和存储范围
操作。清除后沿用字段自身的省略/默认语义，不等同于删除远端配置。必填或只读下拉框
不提供清除按钮。控件渲染测试覆盖单选、多选、必填和只读，未进行浏览器交互验证。

SMB 集群创建增加客户端访问地址输入，API smb_public_addresses 字符串数组映射到
重复的原生 --public-addrs=IP/前缀%目标网络 参数。目标网络可省略；支持 IPv4/IPv6，
要求地址携带前缀，目标使用无主机位的网络 CIDR。空列表不指定地址。参考 module.py
按 % 拆分 address/destination，资源模型再交由 SMBClusterPublicIPSpec 校验。
测试覆盖双栈、目标网络、非法输入和表单转换；不修改主机网络，编辑及实网验证待补齐。

SMB 编辑现提供显式替换全部客户端访问地址开关，关闭时保留原始 public_addrs（包括
多目标网络结构），开启时将输入的地址列表转换为原生 address/destination 资源。
开启后留空发送空数组清空地址，界面提示可能中断访问。通过 smb apply 更新并回读
完整配置。测试覆盖保留复杂原值、双栈替换、清空、数量保留及旧地址回读拒绝；
当前输入每个地址支持一个目标网络，真实客户端连通性尚未验证。

SMB 共享创建改为单次 smb apply 提交完整资源，支持创建时配置登录控制、限制访问、
可浏览和描述，不先创建开放共享再追加权限。保留只读、显示名、CephFS 路径及子卷选择，
并显式使用原生 samba-vfs 默认 provider。创建前通过 share ls 确认 ID 不存在，创建后
通过 show 核验完整配置而非仅确认 ID 出现。原生 apply 缺少 create-only CLI 参数，
外部并发同名创建仍存在竞争窗口。旧 share create 参数测试已替换为资源载荷测试，
覆盖权限、路径归一化、子卷、显示名、已有资源拒绝及原生回读；实际客户端验证待完成。

SMB 共享编辑新增登录控制：显式开启替换后可增删用户/组规则，设置 none、read、
read-write、admin，并控制 restrict_access。未开启替换保持原规则，清空规则且关闭
限制可移除登录控制。API 要求规则和限制开关同时提交，校验名称、类型、权限和重复项。
参考 smb/resources.py、enums.py 及 handler.py 的 Samba 配置生成逻辑；限制模式要求
至少一个非拒绝项，避免原生不生成 valid users 时造成白名单误解。管理员规则和替换
访问范围有确认提示。经 smb apply 更新后核验完整配置，接受原生省略 false 限制字段。
测试覆盖保留/替换/清空、非法和仅拒绝规则、原生回读、前端增删和权限输入；共享创建时
配置规则及真实客户端权限验证仍待补齐。

SMB 用户组资源新增 PATCH /smb/usersgroup 完整替换编辑。脱敏采集额外保存 user_names，
并在列表中展示用户名，对应参考 smb-usersgroups-details 的 Username 展示。
用户名仅取自脱敏库存；空数组展示“无用户”，缺失或异常数据展示“未知”，不展示密码对象。
只有每项用户名均可解析时才提供回填列表；不保存密码。编辑回填用户名、组名和绑定，
保留的每个用户密码必须重新输入，允许删除全部用户。明确提示未列出的旧用户将被移除，
空组名清空组，空绑定解除绑定。执行前确认资源存在，禁止以更新接口创建缺失资源。
沿用加密参数队列、标准输入、脱敏输出和完整非密码元数据回读，未验证客户端登录。
测试覆盖用户名白名单、无密码回填、配置替换及异常/缺失前置状态；前后端检查通过。
实际浏览器/集群验证及其余 Dashboard 功能仍待推进。

SMB 本地用户组资源创建已接通 POST /smb/usersgroup：表单支持增删用户行及密码输入、
多组名和可选集群绑定；请求 users 映射 values.users，groups 映射 values.groups.name。
后端接受显式空用户数组，拒绝缺失/null 列表、缺失密码、重复用户名/组名及非法资源 ID。沿用加密队列、stdin、
脱敏原生输出和固定失败消息。创建前确认 ID 不存在，创建后对照用户名、组名及绑定，
不对照密码或验证客户端登录。原生 apply 的外部并发同名竞争限制仍然存在。
测试覆盖原生 payload、密码原值、已有资源拒绝、异常回读和前端编辑器/请求转换。
用户组编辑及真实浏览器、Ceph 集群验证仍待完成。

域加入凭据编辑通过 PATCH /smb/join/auth 替换账号、密码和绑定配置。界面不回填密码，
每次保存必须重新输入；绑定留空表示解除绑定，保存前提示引用集群可能受影响。
执行前严格验证目标存在，避免将缺失资源意外创建；资源 ID 不允许改名。沿用加密队列、
stdin 输入、脱敏输出、固定失败消息与元数据回读，仍不验证域登录或密码可用性。
原生 apply 缺少条件更新参数，外部并发修改仍存在竞争窗口。测试覆盖回填不含密码、
替换绑定、目标缺失/响应异常、身份不匹配和旧配置回读；真实集群验证待完成。

SMB 域加入凭据支持创建：POST /smb/join/auth 接收 name、username、write-only
password 和可选 linked_to_cluster。表单提供密码控件及当前集群内的 SMB 绑定选择。
沿用操作队列加密参数存储，密码只进入 smb apply 标准输入，不进入命令参数；强制
--password-filter-out=hidden，命令失败使用固定错误，不透传可能含密码的原始输出。
执行前脱敏列举拒绝已有 ID，执行后核验 ID、账号名及绑定信息，不回读密码，也不声称
验证了域登录。原生 apply 不提供 create-only CLI 参数，外部并发同名写入仍有竞争窗口。
密码原值保留、已有资源拒绝、失败错误屏蔽、非法输入、回读与前端请求测试已覆盖。
本地用户组资源创建、凭据编辑、浏览器及真实集群验证仍待完成。

SMB 域加入凭据和用户组资源增加删除操作：DELETE /smb/join/auth 与
DELETE /smb/usersgroup 接收集群作用域及 name，进入现有高风险异步操作链路。
参考 controllers/smb.py 删除逻辑，使用 smb apply -i - 提交资源类型、ID 和
intent: removed；Ceph staging.py 负责拒绝仍被集群引用的资源，不自动解除引用。
确认 apply 成功后以 password-filter=hidden 列举相应资源类型，严格校验响应并确认
目标消失才报告成功。前端提示密码无法从本系统恢复以及先调整引用的要求。
测试覆盖身份映射、原生 stdin、脱敏回读、残留资源、异常响应及失败 apply；没有实际
删除任何 Ceph 资源，真实集群验证及凭据创建/编辑仍待完成。

SMB 导航新增域加入凭据和用户组资源只读页面，对照参考 smb-join-auth-list 与
smb-usersgroups-list 展示资源 ID、绑定集群，以及域账号名、用户数量、组名。
采集仍强制 password-filter=hidden，仅新增 username、user_count、group_names
展示字段白名单，不持久化 auth/values 原始对象或密码。缺失统计显示未知而非零。
两个列表使用现有分页、集群作用域、SMB 能力与采集状态提示；不提供未实现的写入按钮。
测试覆盖展示字段提取、密码排除、页面列、路由注册和绑定信息；凭据增删改与真实环境
验证仍待补齐。

SMB 认证引用库存现独立采集 ceph.smb.join.auth 与 ceph.smb.usersgroups，命令为
smb show <资源类型> --results=full --password-filter=hidden --format json。
仅保存资源类型、ID、intent、linked_to_cluster 白名单元数据，不保存 auth、values
或其他账号密码字段。未绑定集群的资源也会采集；不支持脱敏参数时不降级为明文命令。
异常响应标记对应库存不可用，防止误删旧观测；成功空数组表示确实无资源。
只读接口 GET /smb/join/auths、GET /smb/usersgroups 使用现有集群作用域和 SMB 能力检查。
新增实体改变开发数据库基线校验，旧开发数据库需重建，不提供兼容迁移。
脱敏白名单、参数、非法资源、重复 ID 与空库存测试通过；前端选择和凭据管理尚未接入，
没有真实 Ceph 集群验证。

SMB 集群创建和编辑的域加入凭据、用户组引用现使用上述脱敏库存多选框，沿用参考
Dashboard 的资源引用方式，不在集群表单中读取或传输密码。候选项按当前 Ceph 集群
分页加载，仅允许未绑定或绑定目标 SMB 集群的资源；创建时修改集群 ID 会刷新范围。
编辑回填资源 ID 数组，提交时校验选项属于当前范围；认证模式在创建时必须选择。
前端测试覆盖两个 API 路径、集群作用域、绑定过滤、去重、回填及数组请求转换，
make test-frontend 通过。凭据本身的增删改与真实浏览器/集群验证仍待完成。

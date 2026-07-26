# 阿里云 Ceph 临时实验环境工具

这个 Go 工具按照配置文件的 `nodes` 列表，在阿里云 VPC 中创建一台或多台按量付费 ECS，
等待实例启动后通过公网 SSH 并行执行 `hooks/init-node.sh`。当配置至少 3 台机器时，工具会在
第一台机器完成 cephadm bootstrap、主机纳管、扩展盘 OSD 部署、配置分发和 CephFS 创建；
少于 3 台时只初始化每台机器的基础环境。

安全和费用保护：

- 每次 `RunInstances` 都设置 UTC 格式的 `AutoReleaseTime`，本地进程退出不影响到期释放；
- 数据盘设置 `DeleteWithInstance=true`，实例释放时一并释放；
- `ssh.password` 留空时使用加密随机数生成符合 ECS 要求的 10 位登录密码；
- `create` 在配置校验通过后立即创建资源；`delete` 先输出待删除节点 JSON，再交互确认，
  也可传入 `--yes` 跳过确认；
- 工具只释放状态文件中记录的实例 ID，不按名称扫描或删除其他资源；
- 网络 ID 为空时检索并复用带约定名称前缀的兼容资源，否则创建可供后续复用的专用网络；
- 托管安全组默认允许所有 IPv4 访问 TCP 22、3300、6789、6800～7568 和 8443；
- 只有本次运行创建的网络资源才有资格在手动销毁时被删除；
- 状态文件、独立的 SSH `known_hosts` 和按 hostname 保存的执行日志位于 `.state/`，该目录不会提交到 Git。

## 前置条件

1. Go 1.23 或更高版本。
2. 目标镜像支持使用自定义密码登录；VPC、交换机和安全组可以预先提供，也可以自动创建。
3. 显式指定安全组时，需自行允许当前出口 IP 访问 TCP 22，并允许节点间 Ceph 内网流量；
   自动创建或复用的托管安全组会按 `network.ssh_source_cidr` 配置规则。
4. RAM 身份至少具备 ECS 实例的创建、查询和删除权限。启用自动网络管理时，还需要
   VPC/交换机的查询、创建和删除权限，以及安全组的查询、创建、授权和删除权限。
5. 账户余额、ECS/磁盘配额和目标可用区库存充足。

AccessKey 配置在专属的 `credentials` 模块中。默认模板故意保留空值，复制到 Git 忽略的
`config.local.yaml` 后再填写：

```yaml
credentials:
  access_key_id: "你的 AccessKey ID"
  access_key_secret: "你的 AccessKey Secret"
  security_token: ""
```

长期 AccessKey 的 `security_token` 保持为空；使用 STS 临时凭证时，三个字段都必须填写且
必须属于同一组未过期凭证。工具不会回退读取环境变量、CLI profile 或 ECS RAM Role。
`validate`、`create`、`list` 和 `delete` 都会拒绝空的 AccessKey ID 或 Secret。

不要在仓库跟踪的 `config.yaml` 中填写真实密钥，也不要把 `config.local.yaml` 强制加入 Git。
优先使用最小权限 RAM 用户或短期 STS 凭证，不要使用阿里云主账号 AccessKey。

## 配置和运行

默认配置文件为带逐字段中文注释的 `config.yaml`。工具仅接受 `.yaml` 或 `.yml` 配置，
并拒绝未知字段。先复制为 Git 忽略的 `config.local.yaml`，填写 AccessKey，再调整密码、节点规格
和磁盘；网络资源 ID 可以保留或留空自动管理。空凭证的模板不能通过 `validate`。
`hooks.init_script` 和 `hooks.deploy_script` 的相对路径均相对于配置文件所在目录解析。

工具通过 `RunInstances.Password` 设置 `ssh.password`，不传递 `KeyPairName` 或 `ImageOptions`，
因此 ECS 使用默认的 root 用户和“自定义密码”登录。`ssh.user` 只允许填写 `root`，其他值会在
创建任何云资源前校验失败。密码配置为 `""` 时，工具使用加密随机数生成 10 位密码，并确保同时包含
大写字母、小写字母、数字和特殊字符；也可显式指定符合 ECS 8～30 字符规则的密码。`create`
会在创建任何云资源前打印一次密码，并在所有配置节点进入 `Running` 后再次打印。自动初始化
使用同一密码认证。为便于故障恢复，每个节点的 SSH 用户、密码、端口、公网地址、密码是否
自动生成以及日志路径都会写入权限为 `0600` 的状态文件，但密码不会出现在远程命令参数或
节点日志中。请像保护凭证一样保护 `.state/`。默认的
`network.ssh_source_cidr=0.0.0.0/0` 便于临时实验，但会暴露 SSH、Dashboard 和 Ceph 服务；
只需从固定出口管理集群时，应将其改为可信公网地址的 `/32` CIDR。

所有实际执行的子命令都会把步骤日志写到标准输出，工具步骤使用进程当前时区的 RFC3339 时间戳，
例如 `[2026-07-26T18:42:10+08:00]`。网络资源状态轮询每 3 秒输出一次，ECS 和 SSH 等待阶段
每 5 秒输出一次，避免长时间没有反馈。日志器不使用用户态缓冲区，每条完整日志通过一次标准输出
写操作立即提交，并执行 best-effort flush；远程脚本的原始标准输出和标准错误会同时显示在终端，
并追加到 `.state/log/<hostname>.log`，例如 `.state/log/ceph-node-1.log`；每次 SSH 等待或脚本运行
都有带时间的阶段边界。日志目录权限
为 `0700`，日志文件权限为 `0600`。同一份 hook 输出还会由远端 `tee` 实时追加到每台 ECS 的
`/var/log/ceph-lab/<脚本名>.log`（权限 `0600`）。脚本失败时，终端、本地节点日志和远端日志都会
记录失败行号、命令及退出码；SSH 连接失败原因也会同时写到终端和本地节点日志。Ceph 部署只在
第一台机器执行，因此完整集群部署日志位于第一台机器的 `deploy-ceph.sh.log`。`list` 和 `create`
的最终 JSON 会跟在日志之后，并包含节点的 SSH 连接信息和本地日志路径。`delete` 成功释放资源后
会删除整个本地 `.state/` 目录，包括状态文件、known_hosts 和节点日志；远端日志随 ECS 生命周期
保留。

远端 hook 每 15 秒输出一次 `[RUNNING]` 心跳，并使用随机 run ID 将退出状态原子写入远端
`/var/log/ceph-lab/status/`。如果主 SSH 输出通道中断或丢失 exit status，工具会通过新的 SSH
连接持续查询该 run ID；远端明确返回状态 0 时输出 `[RECOVERED]` 并继续流程，非零状态或超过
`lifecycle.wait_timeout` 才判定失败。初始化要求镜像预装 `openssh-clients`，脚本不会在承载
当前 hook 的 SSH 会话中升级 OpenSSH，以免 sshd 重启影响输出通道。

当前示例已按 Launch Advisor 导出的新加坡配置调整：

- 地域/可用区：`ap-southeast-1` / `ap-southeast-1b`；
- 规格：当前模板包含三台 `ecs.g7a.xlarge`，每台 2 核 × 2 线程；
- 镜像：`centos_stream_9_x64_20G_alibase_20260616.vhd`，使用 root 登录；
- 磁盘：40 GiB ESSD PL0 系统盘，每台附加两块 100 GiB ESSD PL0；
- 网络：`PayByTraffic`，公网带宽上限 100 Mbit/s；
- 元数据和安全：`http_tokens=optional`、`security_enhancement_strategy=Active`，并设置临时 root 密码。

默认配置将 VSwitch 和安全组 ID 留空，以启用托管资源复用或自动创建。100 Mbit/s 是
按流量计费的带宽上限，运行前也应确认安全组的 SSH 来源限制和预计流量费用。

有几项有意不按 Terraform 文本原样复制：工具使用 `nodes` 中的唯一名称作为主机名；每个节点
分别调用一次 `RunInstances` 并设置 `Amount=1`；`is_outdated` 是 Provider 状态字段，不是
创建参数；`Affinity`、`Tenancy`、私有 DNS 和 `UniqueSuffix` 继续使用 ECS 默认值。数据盘的
`delete_with_instance=true` 已直接映射到 ECS API。

### 自动网络管理

`network.vpc_id`、`network.v_switch_id` 和 `network.security_group_id` 可以为空。解析顺序如下：

1. 显式 ID 优先；VPC ID 为空时可从显式交换机或安全组推断，并校验三者属于同一 VPC。
2. 缺少的资源会按名称前缀检索：VPC 使用 `ceph-vpc-`，交换机使用 `ceph-switch-`，
   安全组使用 `ceph-security-group-`。当前集群的完整名称优先，其他同前缀资源按 ID 稳定选择；
   VPC 的 CIDR 必须匹配配置，交换机还必须属于选定 VPC 和目标可用区。
3. 交换机必须为 `Available`，位于配置的可用区，可用私网 IP 数不得少于 `nodes` 的节点数。
4. 没有合适资源且 `auto_create=true` 时，按配置 CIDR 创建 VPC、交换机和普通安全组。

自动创建或复用托管安全组时，默认允许所有 IPv4 地址访问实验环境：

```yaml
network:
  vpc_id: ""
  v_switch_id: ""
  security_group_id: ""
  auto_create: true
  reuse_managed_resources: true
  vpc_cidr: "172.31.0.0/16"
  v_switch_cidr: "172.31.0.0/24"
  ssh_source_cidr: "0.0.0.0/0"
  cleanup_created_resources: false
```

工具为托管安全组配置以下入站规则：

- `ssh_source_cidr` → TCP 22：SSH 初始化和管理；
- `ssh_source_cidr` → TCP 8443：Ceph Dashboard HTTPS；
- `ssh_source_cidr` → TCP 3300、6789：Ceph Monitor v2/v1；
- `ssh_source_cidr` → TCP 6800～7568：Ceph Manager、OSD 和 MDS；
- 所有节点均加入同一个普通安全组，利用其默认组内互通策略承载 Ceph 内网流量。

`0.0.0.0/0` 会把上述 TCP 端口开放给整个 IPv4 互联网，只适合短生命周期测试。工具没有添加
Windows RDP 使用的 3389 端口；ICMP 也不是当前部署流程的必要条件。显式填写已有
`security_group_id` 时，工具尊重其现有规则且不会修改。

所有网络资源 ID 和所有权标记都会逐项写入状态文件。默认的
`cleanup_created_resources=false` 会在 `delete --yes` 释放 ECS 后保留网络，以便后续运行按
名称前缀复用。设为 true 时，工具会按安全组、交换机、VPC 的依赖顺序删除本次运行创建的
资源；复用资源和显式配置资源始终不会被删除。ECS 到期自动释放也不会触发网络删除。

`cluster.ecs_endpoint` 是可选 ECS 自定义端点；VPC 自定义端点需单独使用
`cluster.vpc_endpoint`，避免把一个产品的端点错误用于另一个产品。

自动网络管理涉及的主要 RAM Action 为：

- `vpc:DescribeVpcs`、`vpc:CreateVpc`、`vpc:DeleteVpc`
- `vpc:DescribeVSwitches`、`vpc:CreateVSwitch`、`vpc:DeleteVSwitch`
- `ecs:DescribeSecurityGroups`、`ecs:CreateSecurityGroup`
- `ecs:AuthorizeSecurityGroup`、`ecs:DeleteSecurityGroup`

```bash
cd scripts/aliyun-ceph-lab
cp config.yaml config.local.yaml
chmod 600 config.local.yaml
go mod download

go run ./cmd --help
go run ./cmd create --help
go run ./cmd validate --config config.local.yaml
go run ./cmd create --config config.local.yaml
go run ./cmd list --config config.local.yaml
go run ./cmd delete --config config.local.yaml
# 或跳过交互确认；两种方式都会先输出待删除节点 JSON
go run ./cmd delete --config config.local.yaml --yes
```

也可以构建单文件命令：

```bash
go build -o bin/aliyun-ceph-lab ./cmd
```

`lifecycle.max_runtime` 接受 Go duration（例如 `90m`、`6h`），范围为 30 分钟到三年。工具会把
到期时间向上取整到分钟，以符合 ECS API 要求。`lifecycle.wait_timeout` 分别用于等待 ECS 进入
`Running` 状态、等待/执行 SSH 初始化，以及部署脚本中每一个 Ceph 就绪检查阶段。

若创建到一半失败，不会自动删除已经成功创建的实例，以便保留现场排查；实例 ID 已逐台
写入状态文件，并且每台实例仍有云端自动释放时间。此时可立即执行 `delete --yes`。

## Hook 约定

初始化 hook 会通过 root SSH 用户在 `nodes` 的每台机器上运行，可读取：

- `CEPH_LAB_CLUSTER_NAME`
- `CEPH_LAB_NODE_NAME`
- `CEPH_LAB_NODE_NAMES`：逗号分隔，顺序与配置一致
- `CEPH_LAB_PUBLIC_IPS`：逗号分隔，顺序与配置一致
- `CEPH_LAB_PRIVATE_IPS`：逗号分隔，顺序与配置一致，用于节点间通信
- `CEPH_LAB_DATA_DISK_COUNT`：当前节点在配置中的扩展盘数量
- `CEPH_LAB_SSH_PRIVATE_KEY_BASE64`、`CEPH_LAB_SSH_PUBLIC_KEY_BASE64`：本次创建动态生成

默认初始化 hook 会安装 cephadm 所需的基础软件、启用时间同步，并用上述节点列表维护
`/etc/hosts` 中带标记的实验集群区段，主机名映射使用私网 IP。工具还会在每次 `create`
时动态生成一对新的 Ed25519 SSH 密钥，通过脚本输入流发给全部节点，实现 root 用户互信；
密钥不会写入配置文件或程序代码。初始化脚本会排除系统盘，并按 `data_disks` 数量记录
可用于 OSD 的未挂载扩展盘。

部署 hook 只在第一台机器运行，可读取：

- `CEPH_LAB_CLUSTER_NAME`
- `CEPH_LAB_BOOTSTRAP_NODE_NAME`：配置中的第一个节点，也是部署 hook 的唯一执行节点
- `CEPH_LAB_NODE_NAMES`：逗号分隔，顺序与配置一致
- `CEPH_LAB_PUBLIC_IPS`：逗号分隔，顺序与配置一致
- `CEPH_LAB_PRIVATE_IPS`：逗号分隔，顺序与配置一致，用于 Ceph 和节点间 SSH
- `CEPH_LAB_DATA_DISK_COUNTS`：逗号分隔，顺序与配置一致
- `CEPH_LAB_WAIT_TIMEOUT_SECONDS`

默认部署 hook 只在节点数至少为 3 时执行。管理端到 ECS 的 SSH 和 Dashboard 外部入口使用
公网 IP；`/etc/hosts`、节点间 SSH、cephadm bootstrap Monitor 地址、orchestrator 主机地址和
配置分发全部使用私网 IP。阿里云公网 IP 通过 NAT 映射到 ECS，因此脚本在 bootstrap 使用
第一台机器的私网 IP，同时输出 `https://<第一台公网IP>:8443/` 作为 Dashboard 外部入口。
部署 hook 会校验当前主机名就是配置中的第一个节点；在该节点完成 bootstrap、注册其余节点，
并通过 orchestrator 为当前节点及其他节点部署服务。
脚本会先统一输出节点的私网 IP、公网 IP、OSD 设备和预期 OSD 数量，再按阶段部署。每个阶段会
轮询实际状态：SSH 互信、Ceph 与 orchestrator 可用、主机全部纳管、OSD 全部注册并运行、
CephFS 可访问。轮询总时限沿用
`lifecycle.wait_timeout`，避免前一步尚未就绪时过早执行下一步。

## 参考资料

- [ECS 2014-05-26 Go Tea SDK 入门](https://api.aliyun.com/api-tools/sdk/Ecs?version=2014-05-26&language=go-tea&tab=primer-doc)
- [Darabonba OpenAPI Go 客户端中文说明](https://github.com/alibabacloud-go/darabonba-openapi/blob/master/README-CN.md)
- [RunInstances API](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-runinstances)
- [ModifyInstanceAutoReleaseTime API](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-modifyinstanceautoreleasetime)
- [DeleteInstance API](https://api.aliyun.com/document/Ecs/2014-05-26/DeleteInstance)
- [VPC 2016-04-28 API 概览](https://help.aliyun.com/zh/vpc/developer-reference/api-vpc-2016-04-28-overview)
- [CreateVpc API](https://help.aliyun.com/en/vpc/developer-reference/api-vpc-2016-04-28-createvpc)
- [CreateVSwitch API](https://help.aliyun.com/en/vpc/api-createvswitch)
- [安全组 API](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-dir-security-groups/)

# 阿里云 Ceph 实验工具

这个工具用于在阿里云上创建短生命周期 Ceph 实验环境。

它会按 `config.yaml` 的 `nodes` 创建 ECS，等待实例运行后通过公网 SSH 执行初始化 hook。当节点数不少于 3 台时，第一台节点还会执行 Ceph 部署 hook：完成 cephadm bootstrap、主机纳管、OSD 部署、配置分发和 CephFS 创建。少于 3 台时只初始化基础环境。

## 快速开始

```bash
cd tools/aliyun-ceph-lab
cp config.yaml config.local.yaml
chmod 600 config.local.yaml

go mod download
go run ./cmd validate --config config.local.yaml
go run ./cmd create --config config.local.yaml
go run ./cmd list --config config.local.yaml
go run ./cmd delete --config config.local.yaml
```

跳过删除确认：

```bash
go run ./cmd delete --config config.local.yaml --yes
```

构建单文件命令：

```bash
go build -o bin/aliyun-ceph-lab ./cmd
```

## 前置条件

- Go 1.23 或更高版本。
- 阿里云账号余额、ECS 配额、磁盘配额和目标可用区库存充足。
- 目标镜像支持 root 用户通过自定义密码登录。
- RAM 身份具备 ECS 实例的创建、查询和删除权限。
- 启用自动网络管理时，还需要 VPC、VSwitch、安全组的查询、创建、授权和删除权限。
- 显式指定安全组时，需要自行放通 SSH、Ceph 节点内网通信和 Dashboard 入口。

## 模块

### cmd

命令入口，提供：

- `validate`：校验配置，不创建资源。
- `create`：创建 ECS，初始化节点，并在 3 台及以上时部署 Ceph。
- `list`：读取本地状态并输出节点连接信息。
- `delete`：按状态文件释放 ECS，可选清理本次创建的网络资源。

### config

负责读取和校验 YAML 配置。

- 只接受 `.yaml` 或 `.yml`。
- 拒绝未知字段。
- `hooks.*_script` 的相对路径按配置文件所在目录解析。
- `ssh.user` 只允许 `root`。
- `credentials.access_key_id` 和 `credentials.access_key_secret` 不能为空。

### cloud

封装阿里云 ECS/VPC API。

- 创建按量付费 ECS。
- 设置实例 `AutoReleaseTime`。
- 创建或复用 VPC、VSwitch、安全组。
- 授权托管安全组入站规则。
- 删除状态文件中记录的实例和可清理网络资源。

### lab

编排完整实验流程。

- 校验配置。
- 准备网络。
- 创建节点。
- 等待 ECS `Running`。
- 并行执行初始化 hook。
- 在第一台节点执行部署 hook。
- 写入状态和节点日志。

### remote

负责 SSH 执行。

- 使用独立的 `.state/known_hosts`。
- 将脚本输出实时写到终端和 `.state/log/<hostname>.log`。
- 远端 hook 每 15 秒输出 `[RUNNING]` 心跳。
- SSH 输出通道异常时，会重新连接查询远端 run ID 状态。

### state

维护本地状态文件。

- 记录实例 ID、IP、SSH 信息、日志路径和网络资源归属。
- 文件权限为 `0600`。
- `delete` 成功后会删除整个 `.state/`。

### logging

提供工具侧日志。

- 工具步骤使用当前时区 RFC3339 时间戳。
- 网络轮询约每 3 秒输出一次。
- ECS 和 SSH 等待阶段约每 5 秒输出一次。
- 日志写入不依赖用户态缓冲。

### hooks

包含默认远端脚本。

- `hooks/init-node.sh`：每台 ECS 都执行，安装基础依赖、配置主机名、维护 `/etc/hosts`、记录 OSD 候选盘。
- `hooks/deploy-ceph.sh`：仅第一台 ECS 执行，负责 cephadm bootstrap、主机纳管、OSD、CephFS 和 Dashboard 密码配置。

## 凭证

不要在仓库跟踪的 `config.yaml` 中填写真实密钥。复制到 Git 忽略的 `config.local.yaml` 后填写：

```yaml
credentials:
  access_key_id: "你的 AccessKey ID"
  access_key_secret: "你的 AccessKey Secret"
  security_token: ""
```

长期 AccessKey 的 `security_token` 保持为空。使用 STS 临时凭证时，三个字段都必须填写，且必须属于同一组未过期凭证。

工具不会回退读取环境变量、CLI profile 或 ECS RAM Role。优先使用最小权限 RAM 用户或短期 STS 凭证。

## 配置

默认模板对应新加坡实验环境：

- 地域/可用区：`ap-southeast-1` / `ap-southeast-1b`
- 规格：三台 `ecs.g7a.xlarge`
- 镜像：`centos_stream_9_x64_20G_alibase_20260616.vhd`
- 系统盘：40 GiB ESSD PL0
- 数据盘：每台两块 100 GiB ESSD PL0
- 网络：`PayByTraffic`，公网带宽上限 100 Mbit/s
- 元数据：`http_tokens=optional`

常用配置项：

- `lifecycle.max_runtime`：云端自动释放时间，支持 Go duration，例如 `90m`、`6h`。
- `lifecycle.wait_timeout`：等待 ECS、SSH 和 Ceph 阶段就绪的基础超时时间；节点初始化 hook 最多等待该值的 2 倍。
- `ssh.password`：为空时自动生成 10 位 ECS 合规密码；也可显式填写 8 到 30 字符密码。
- `network.access_source_cidr`：管理入口来源 CIDR，默认 `0.0.0.0/0` 只适合临时实验。
- `network.cleanup_created_resources`：为 `true` 时，删除 ECS 后继续删除本次创建的安全组、交换机和 VPC。

## 网络

`network.vpc_id`、`network.v_switch_id` 和 `network.security_group_id` 可以留空。

解析顺序：

1. 显式 ID 优先。
2. 缺少的资源按名称前缀查找并复用。
3. 找不到可复用资源且 `auto_create=true` 时，自动创建 VPC、VSwitch 和普通安全组。

托管资源名称前缀：

- VPC：`ceph-vpc-`
- VSwitch：`ceph-switch-`
- 安全组：`ceph-security-group-`

托管安全组会允许 `access_source_cidr` 访问：

- TCP 22：SSH
- TCP 36900：CephTower HTTP 服务
- TCP 8443：Ceph Dashboard
- TCP 3300、6789：Ceph Monitor
- TCP 6800-7568：Ceph Manager、OSD、MDS

所有节点加入同一个普通安全组，依赖默认组内互通承载 Ceph 内网流量。显式填写已有安全组时，工具不会修改其规则。

## 安全

- 每次创建实例都会设置 UTC `AutoReleaseTime`。
- 数据盘设置 `DeleteWithInstance=true`。
- 工具只删除状态文件中记录的实例 ID，不按名称扫描删除其他资源。
- 只有本次运行创建的网络资源，才可能在手动删除时被清理。
- `.state/` 包含状态文件、known_hosts 和节点日志，不会提交到 Git。
- `0.0.0.0/0` 会把 SSH、CephTower HTTP、Dashboard 和 Ceph 端口暴露给整个 IPv4 互联网。

## 日志

本地日志：

- 工具日志输出到终端。
- 节点日志追加到 `.state/log/<hostname>.log`。
- 日志目录权限为 `0700`，日志文件权限为 `0600`。

远端日志：

- hook 会在 ECS 后台脱离 SSH 会话执行，输出写入 `/var/log/ceph-lab/<脚本名>.log`。
- 本地节点日志会轮询并同步远端 hook 日志；短暂 SSH 断开不会中断正在运行的 hook。
- 远端日志权限为 `0600`。
- Ceph 部署只在第一台机器执行，完整部署日志在第一台机器的 `deploy-ceph.sh.log`。

`create` 的最终 JSON 会包含节点 SSH 连接信息、本地日志路径和 `ceph` 连接信息。
`ceph.cephtower_cluster_create.monitor_addresses` 可直接填写到 CephTower 新增集群
表单的 MON 地址中，支持 v1、v2 或 v2+v1 addrvec 格式。

## Hook 参数

默认情况下，`go run ./cmd create` 会按下面的参数形式调用 hook。也可以把 hook
复制到机器上手工执行：

```bash
sudo bash init-node.sh --help
sudo bash deploy-ceph.sh --help
```

### init-node.sh

每台节点都会执行，需要传入：

- `--cluster-name`：集群名，例如 `ceph-dev`。
- `--node-name`：当前节点主机名，例如 `ceph-node-1`。
- `--node-names`：按集群顺序排列的节点主机名，逗号分隔。
- `--public-ips`：按同一顺序排列的公网 IP，逗号分隔。
- `--private-ips`：按同一顺序排列的私网 IP，逗号分隔。
- `--data-disk-count`：当前节点数据盘数量。
- `--ssh-private-key-base64`：集群共享 SSH 私钥的 base64 内容。
- `--ssh-public-key-base64`：集群共享 SSH 公钥的 base64 内容。

手工执行时，所有节点需要使用同一对集群 SSH key。可以先生成一次：

```bash
ssh-keygen -t ed25519 -f ceph_lab_ed25519 -N '' -C ceph-dev
base64 -w0 ceph_lab_ed25519
base64 -w0 ceph_lab_ed25519.pub
```

macOS 上可用：

```bash
base64 -i ceph_lab_ed25519 | tr -d '\n'
base64 -i ceph_lab_ed25519.pub | tr -d '\n'
```

示例：

```bash
sudo bash init-node.sh \
  --cluster-name ceph-dev \
  --node-name ceph-node-1 \
  --node-names ceph-node-1,ceph-node-2,ceph-node-3 \
  --public-ips 8.8.8.1,8.8.8.2,8.8.8.3 \
  --private-ips 172.31.0.10,172.31.0.11,172.31.0.12 \
  --data-disk-count 2 \
  --ssh-private-key-base64 '<base64 of ceph_lab_ed25519>' \
  --ssh-public-key-base64 '<base64 of ceph_lab_ed25519.pub>'
```

复制到第二、第三台机器时，只需要改：

- `--node-name`
- `--data-disk-count`

### deploy-ceph.sh

只在第一台节点执行，需要传入：

- `--cluster-name`：集群名，例如 `ceph-dev`。
- `--bootstrap-node-name`：第一台/bootstrap 节点主机名。
- `--node-names`：按集群顺序排列的节点主机名，逗号分隔。
- `--public-ips`：按同一顺序排列的公网 IP，逗号分隔。
- `--private-ips`：按同一顺序排列的私网 IP，逗号分隔。
- `--data-disk-counts`：按同一顺序排列的每台节点数据盘数量，逗号分隔。
- `--wait-timeout-seconds`：每个就绪等待步骤的超时时间，例如 `900`。
- `--dashboard-password`：要配置并写入状态 JSON 的 Dashboard admin 密码。

部署脚本使用私网 IP 做 Ceph 内部通信，并配置 `https://<第一台公网IP>:8443/` 作为 Dashboard 外部入口。

部署脚本必须在所有节点的 `init-node.sh` 成功完成后，只在第一台节点执行。示例：

```bash
sudo bash deploy-ceph.sh \
  --cluster-name ceph-dev \
  --bootstrap-node-name ceph-node-1 \
  --node-names ceph-node-1,ceph-node-2,ceph-node-3 \
  --public-ips 8.8.8.1,8.8.8.2,8.8.8.3 \
  --private-ips 172.31.0.10,172.31.0.11,172.31.0.12 \
  --data-disk-counts 2,2,2 \
  --wait-timeout-seconds 900 \
  --dashboard-password 'CephTower#example'
```

手工执行顺序：

1. 准备三台机器，确保主机名分别匹配 `--node-names`。
2. 把 `init-node.sh` 复制到每台机器。
3. 在每台机器执行 `sudo bash init-node.sh ...`，并调整当前节点参数。
4. 把 `deploy-ceph.sh` 放到第一台机器。
5. 在第一台机器执行 `sudo bash deploy-ceph.sh ...`。

## 故障处理

- `create` 中途失败时，不会自动删除已创建实例，便于保留现场。
- 实例 ID 会逐台写入状态文件，可立即执行 `delete --yes` 清理。
- 实例仍有云端自动释放时间，本地进程退出不影响到期释放。
- SSH 连接失败、脚本失败行号、命令和退出码会同时写入终端和节点日志。
- 如果远端 hook 进程退出但没有写出状态文件，create 会尽早报错而不是等完整超时。
- 初始化要求镜像预装 `openssh-clients`；hook 不会升级当前 SSH 会话依赖的 OpenSSH。

## RAM Action

自动网络管理涉及：

- `vpc:DescribeVpcs`
- `vpc:CreateVpc`
- `vpc:DeleteVpc`
- `vpc:DescribeVSwitches`
- `vpc:CreateVSwitch`
- `vpc:DeleteVSwitch`
- `ecs:DescribeSecurityGroups`
- `ecs:CreateSecurityGroup`
- `ecs:AuthorizeSecurityGroup`
- `ecs:DeleteSecurityGroup`

ECS 创建、查询、删除还需要对应 ECS 实例相关权限。

## 参考

- [ECS 2014-05-26 Go Tea SDK 入门](https://api.aliyun.com/api-tools/sdk/Ecs?version=2014-05-26&language=go-tea&tab=primer-doc)
- [Darabonba OpenAPI Go 客户端中文说明](https://github.com/alibabacloud-go/darabonba-openapi/blob/master/README-CN.md)
- [RunInstances API](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-runinstances)
- [ModifyInstanceAutoReleaseTime API](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-modifyinstanceautoreleasetime)
- [DeleteInstance API](https://api.aliyun.com/document/Ecs/2014-05-26/DeleteInstance)
- [VPC 2016-04-28 API 概览](https://help.aliyun.com/zh/vpc/developer-reference/api-vpc-2016-04-28-overview)
- [CreateVpc API](https://help.aliyun.com/en/vpc/developer-reference/api-vpc-2016-04-28-createvpc)
- [CreateVSwitch API](https://help.aliyun.com/en/vpc/api-createvswitch)
- [安全组 API](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-dir-security-groups/)

# 部署到机器工具

`deploy-to-machine` 用于把本地 CephTower release 部署到一台远程机器，并查看最近
一次部署机器上的服务日志。

工具会执行 `make release`，连接目标机器获取系统和 CPU 架构，选择匹配的
`dist/cephtower-*-<goos>-<goarch>` 二进制文件，上传到
`/opt/cephtower/bin/cephtower`，再把本项目 `config/config.yaml` 改写
`server.dir=/opt/cephtower` 后上传到 `/opt/cephtower/config/config.yaml`，
最后在远端后台启动服务。服务进程的 stdout/stderr 也会追加到
`/opt/cephtower/log/cephtower.log`，不会再拆分写入单独的 stdout 日志文件。

部署时会先连接远端机器识别架构，再执行
`make release TARGET=<goos>/<goarch>`，只构建当前机器需要的二进制。

上传配置前，工具会先读取远端已有的
`/opt/cephtower/config/config.yaml`。如果其中存在有效的
`database.encryption_key`，本次上传会继续使用这个值；如果不存在或为空，
工具会生成新的 32 位 key。如果远端配置已经写入 `server.bootstrap`，本次上传也会
继续使用远端值，避免初始化完成后的 `false` 被本地模板中的 `true` 覆盖。使用
`--replace conf` 时，也会先读取这些旧值再替换配置文件，避免已有数据库内容无法
解密或初始化状态被重置。使用 `--replace data` 或 `--replace all` 时，新上传配置会
把 `server.bootstrap` 设置为 `true`，用于清空数据后的首次初始化。

## 使用方式

复制 `config.yaml` 为 `config.local.yaml` 后填写固定机器信息，再部署：

```bash
cd tools/deploy-to-machine
go run ./cmd deploy --config config.local.yaml
```

配置文件中不指定 `host` 时，工具会自动扫描
`tools/aliyun-ceph-lab/.state/*.json`，列出已记录的机器并让用户通过序号选择：

```bash
cd tools/deploy-to-machine
go run ./cmd deploy
```

部署成功后，工具会把本次目标机器信息记录到
`tools/deploy-to-machine/.state/last-deploy.json`。随后可以观察最近一次部署的服务
日志：

```bash
go run ./cmd watch
```

`watch` 会通过 SSH 连接最近一次部署的机器，持续输出
`/opt/cephtower/log/cephtower.log`。按 `Ctrl+C` 会结束观测。日志轮转后，`watch`
会按文件名自动跟随新的 `cephtower.log` 文件，避免一直停留在旧日志文件上。

如果需要直接观察配置文件指定的机器，也可以显式传入配置：

```bash
go run ./cmd watch --config config.local.yaml
```

部署前替换远端目录：

```bash
go run ./cmd deploy --replace bin,conf
go run ./cmd deploy --replace all
```

`--replace` 支持 `bin`、`conf`、`data`、`log` 和 `all`。

使用 `--replace all` 时，工具会先把远端已有的 `/opt/cephtower` 目录内容复制到
`/opt/cephtower.backup/<YYYYMMDDHHMMSS>/`，再重新初始化部署目录。

## 配置

复制 `config.yaml` 为 `config.local.yaml` 后填写固定机器信息。配置字段都在
YAML 顶层：

- `host`：目标机器 SSH 地址；为空时自动从 aliyun-ceph-lab 状态文件选择。
- `port`：目标机器 SSH 端口，默认 `22`。
- `user`：目标机器 SSH 用户，默认 `root`。
- `password`：目标机器 SSH 密码；自动选择 lab 机器时从状态文件读取。
- `known_hosts`：部署工具使用的 SSH known_hosts 文件。

工具不会读取或写入当前用户的 `~/.ssh/known_hosts`。默认只使用
`tools/deploy-to-machine/.state/known_hosts`；如果云主机重建后复用了相同 IP，工具会
自动刷新这个部署工具状态文件里的旧主机密钥记录，避免旧记录导致连接失败。

应用配置固定使用仓库根目录下的 `config/config.yaml`，release 产物目录固定使用
仓库根目录下的 `dist/`。

命令行提供 `deploy` 和 `watch` 两个子命令，必须显式指定子命令。

```text
Usage: deploy-to-machine <deploy|watch> [command options]

Commands:
  deploy         build, upload, configure, and start CephTower
  watch          stream the last deployed service log

Run deploy-to-machine <command> --help for command options.
```

`deploy` 子命令：

```text
Usage: deploy-to-machine deploy [--config path] [--replace list]

Options:
  --config path   deploy YAML configuration
  --replace list  remote directories to replace: bin,conf,data,log,all
```

`watch` 子命令：

```text
Usage: deploy-to-machine watch [--config path]

Options:
  --config path   deploy YAML configuration
```

工具日志输出到 stderr，格式与 `aliyun-ceph-lab` 一致：

```text
[2026-07-27T23:59:59+08:00] INFO release: running make release
```

# 部署到机器工具

`deploy-to-machine` 用于把本地 CephTower release 部署到一台远程机器。

工具会执行 `make release`，连接目标机器获取系统和 CPU 架构，选择匹配的
`dist/cephtower-*-<goos>-<goarch>` 二进制文件，上传到
`/opt/cephtower/bin/cephtower`，再把本项目 `config/config.yaml` 改写
`server.dir=/opt/cephtower` 后上传到 `/opt/cephtower/config/config.yaml`，
最后在远端后台启动服务。

部署时会先连接远端机器识别架构，再执行
`make release TARGET=<goos>/<goarch>`，只构建当前机器需要的二进制。

上传配置前，工具会先读取远端已有的
`/opt/cephtower/config/config.yaml`。如果其中存在有效的
`database.encryption_key`，本次上传会继续使用这个值；如果不存在或为空，
工具会生成新的 32 位 key。即使使用 `--replace conf` 或 `--replace all`，
也会先读取旧 key 再替换配置文件，避免已有数据库内容无法解密。

## 使用方式

复制 `config.yaml` 为 `config.local.yaml` 后填写固定机器信息，再部署：

```bash
cd tools/deploy-to-machine
go run ./cmd --config config.local.yaml
```

配置文件中不指定 `host` 时，工具会自动扫描
`tools/aliyun-ceph-lab/.state/*.json`，列出已记录的机器并让用户通过序号选择：

```bash
cd tools/deploy-to-machine
go run ./cmd
```

部署前替换远端目录：

```bash
go run ./cmd --replace bin,conf
go run ./cmd --replace all
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

应用配置固定使用仓库根目录下的 `config/config.yaml`，release 产物目录固定使用
仓库根目录下的 `dist/`。

命令行只提供 `--config` 和 `--replace`：

```text
Usage: deploy-to-machine [--config path] [--replace list]

Options:
  --config path   deploy YAML configuration
  --replace list  remote directories to replace: bin,conf,data,log,all
```

工具日志输出到 stderr，格式与 `aliyun-ceph-lab` 一致：

```text
[2026-07-27T23:59:59+08:00] INFO release: running make release
```

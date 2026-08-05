# 同步远程工作区

`sync-workspace` 用于维护 CephTower 的远程开发工作区。它根据本地 Git 工作树持续把
源码增量同步到远端 `/root/cephtower`，在每批变更同步完成后重新构建并重启远端
服务，同时把远端构建日志和应用日志镜像到本工具的 `.state/logs/` 目录。

远端构建使用 `make build`，随后运行：

```text
/root/cephtower/bin/cephtower -config /root/cephtower/app/config/config.yaml
```

工具会强制远端开发配置使用 `0.0.0.0:36900`。前端生产资源会被嵌入后端二进制，
因此 Web UI 和 API 都通过远端机器的 `36900` 端口提供。这里不使用 `make run`，
因为该目标会把 Vite 开发服务放在 `36901` 端口。

## 使用方式

复制 `config.yaml` 为 `config.local.yaml` 并填写固定机器信息，或保持 `host` 为空，
从 `tools/aliyun-ceph-lab/.state/*.json` 中交互选择机器。

```bash
cd tools/sync-workspace
go run ./cmd run
```

`run` 会执行以下流程：

1. 可选清空远端 `/root/cephtower/app`。
2. 扫描 Git 已跟踪文件和未被 `.gitignore` 忽略的新文件。
3. 根据内容摘要增量同步到 `/root/cephtower`；变更文件使用临时目录原子覆盖。
4. 在远端构建前端和后端并启动 Web 服务。
5. 持续检查本地变更；每批变更同步后重新构建并重启服务。
6. 每秒检查远端构建日志和应用日志，增量拉取追加内容并更新本地镜像文件。

按 `Ctrl+C` 会停止同步，并停止本次工具管理的远端开发服务。

工具运行日志输出到 stderr，格式与 `deploy-to-machine` 一致。源码上传、远端删除、
运行环境清理、构建、服务停止、服务启动和健康检查完成都会输出明确记录：

```text
[2026-08-05T12:00:00+08:00] INFO sync: uploaded backend/cmd/main.go -> /root/cephtower/backend/cmd/main.go
[2026-08-05T12:00:05+08:00] INFO service: previous remote service stopped
[2026-08-05T12:00:06+08:00] INFO service: started host=192.0.2.10 pid=1234 Web=http://192.0.2.10:36900
```

查看或停止最近一次目标机器上的开发服务：

```bash
go run ./cmd status
go run ./cmd stop
```

`status` 和 `stop` 在没有显式指定配置时使用
`.state/workspace.json` 中记录的最近一次目标机器。

## 同步范围

工具通过下面的 Git 命令确定候选文件：

```text
git ls-files -co --exclude-standard
```

在 Git 忽略规则之外，工具固定排除以下内容：

- `tools/`、`app/`、`docs/`
- `.git/`、`.github/`、`.agents/`、`.codex/` 和 IDE 配置
- 根目录下的说明文档、许可证和 Agent 指令文件
- `bin/`、`dist/`、Go 构建缓存和工具缓存
- `frontend/node_modules/`、`frontend/dist/`
- `backend/internal/webui/frontend/dist/`、测试数据和覆盖率文件

远端删除操作只会处理远端 source manifest 中由本工具同步过的文件。远端 `app/`、
`.sync-workspace/` 和其他排除目录不会因为源码同步而被清理。

变更文件会先上传并解包到远端 `.sync-workspace/upload-*` 临时目录，传输完成后再通过
同文件系统的原子重命名覆盖对应源码文件。同步过程中不会预先删除变更文件；只有本地
已经删除的路径才会从远端删除。临时目录在成功或失败后都会清理。

## 清空运行环境

配置中的 `clean_on_start` 可以在执行 `run` 时重置远端运行环境：

```yaml
clean_on_start: true
```

启用后，工具会先停止旧服务，再删除并重建 `/root/cephtower/app`。这会清除远端开发
环境的配置、数据库和应用日志。该操作只在 `run` 首次启动时执行一次，文件变更触发
的自动重启不会再次清理 `app/`。

## 本地状态

本地状态保存在 `tools/sync-workspace/.state/`：

```text
.state/
├── known_hosts
├── manifest.json
├── workspace.json
└── logs/
    ├── std.log
    └── cephtower.log
```

远端构建输出和服务标准输出写入 `/root/cephtower/app/log/std.log`，服务 PID 写入
`/root/cephtower/app/data/runtime/server.pid`。远端 `std.log` 和 `cephtower.log` 分别镜像
为本地 `.state/logs/` 下的同名文件，不再按机器或运行时间创建目录。每次执行 `run`
都会清理旧的本地日志目录。运行期间每秒检查一次远端文件；正常追加时只传输新增的
字节区间，并通过本地临时文件原子替换镜像。发生日志截断、轮转、同大小改写或本地
文件状态不一致时会自动全量覆盖，远端文件删除时也会删除对应的本地文件。

## 远端依赖

目标机器需要提供 SSH、GNU `tar`、GNU `find`、`curl`、`make`、Go 1.26 或更高版本，
以及 Node.js 20、npm 10 或更高版本。首次构建时 `make` 会安装缺失的前端依赖。

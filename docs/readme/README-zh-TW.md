<div align="center">

<img src="../../frontend/public/ceph-tower-logo.svg" alt="CephTower logo" width="128" height="128">

# CephTower

Ceph 叢集 Web 管理控制台

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](../../backend/go.mod)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](../../frontend/package.json)
[![Ceph](https://img.shields.io/badge/Ceph-Dashboard%20API-EF5C55)](https://docs.ceph.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](../../LICENSE)

[简体中文](../../README.md) | [**繁體中文**](README-zh-TW.md) | [English](README-en.md) | [日本語](README-ja.md) | [Français](README-fr.md) | [Deutsch](README-de.md) | [Español](README-es.md) | [Português](README-pt.md) | [Русский](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower 使用 Go 後端與 React / Ant Design 前端，透過 Ceph Dashboard API 和
Ceph 指令管理一個或多個 Ceph 叢集。後端提供版本化 REST API、持久化、背景資料
擷取及內嵌 Web UI；前端一律經由同源 `/api` 存取後端。

## 1. 目前功能與狀態

- 首次啟動精靈：選擇 SQLite 或 MySQL、測試連線並建立管理員。
- 身分驗證：12 小時 Bearer Token 工作階段、管理員/一般使用者角色、細緻的讀取與使用者管理權限；設定 SMTP 後可用郵件驗證碼重設密碼。
- 多叢集連線：儲存 MON 位址、`client.admin` 金鑰與 Dashboard 憑證，自動探索並快取主機、守護程序、服務、MON、MGR、MDS、OSD、Mgr 模組及叢集設定。
- 叢集介面：叢集連線與詳細資料、主機、MON、MGR、OSD 和 MDS 管理；支援 Mgr 模組切換、守護程序動作及 OSD in/out、reweight、scrub。
- 資料擷取：依模組設定來源、週期、逾時、重試與優先順序，也可立即執行並查看紀錄。
- 後端整合：涵蓋叢集、Pool/RBD、CephFS/NFS/SMB、RGW、iSCSI、NVMe-oF、Prometheus/Grafana，以及 Dashboard 使用者、角色與設定 API。
- 交付方式：正式環境建置會將前端嵌入 Go 執行檔，由同一 HTTP 服務提供 UI 與 API。

> [!IMPORTANT]
> 專案仍在開發中。叢集管理、使用者管理與資料擷取設定已連接真實後端；總覽與系統
> 資訊頁仍含示範資料，區塊、檔案、物件儲存及監控頁主要是工作流程預留內容。
> 後端已有整合介面不表示所有前端操作均已完成。

## 2. 專案結構

```text
CephTower/
├── backend/
│   ├── cmd/main.go              # 程序進入點
│   └── internal/
│       ├── api/v1/              # REST 路由與處理器
│       ├── service/             # 驗證、叢集、擷取、設定與初始化邏輯
│       ├── store/               # GORM、遷移與 SQLite/MySQL 儲存
│       ├── integration/ceph/    # Ceph Dashboard 與指令用戶端
│       ├── task/                # 背景工作與排程
│       └── webui/               # 內嵌前端資源
├── frontend/src/                # React 控制台、路由、頁面與 API 用戶端
├── config/config.yaml           # 含完整註解的參考設定
├── docs/                        # 架構、Ceph 資料與多語言 README
├── Makefile                     # 開發、測試與建置入口
└── README.md
```

詳細分層與生命週期請參閱 [docs/architecture.md](../architecture.md)。

## 3. 環境需求

| 工具/服務 | 最低版本 | 用途 |
|---|---:|---|
| Go | 1.26 | 後端建置與測試 |
| Node.js | 20 | 前端開發與建置 |
| npm | 10 | 前端相依套件管理 |
| C 編譯工具鏈 | 適用於作業系統的版本 | SQLite 驅動程式使用 CGO |
| Ceph | 已啟用 Dashboard API | 還需 MON 位址與具足夠權限的 keyring |
| MySQL | 選用 | 不使用預設 SQLite 時需要 |

## 4. 快速開始

在儲存庫根目錄執行：

```bash
make run
```

此命令會檢查環境、視需要安裝前端相依套件，並在缺少時從 `config/config.yaml`
建立 `app/config/config.yaml`（開發執行目錄改為 `./app`），接著啟動：

- 後端與正式 Web 入口：<http://localhost:36900>
- Vite 開發伺服器：<http://localhost:36901>（`/api` 代理至後端）

首次瀏覽會轉至 `/initialize`。設定資料庫與管理員後，再於叢集管理新增 Ceph 連線。
若要分別啟動，先執行 `make ensure-run-config`，再於兩個終端機執行：

```bash
make run-backend
make run-frontend
```

### 正式環境建置

```bash
make build
./bin/cephtower -config /path/to/config.yaml
```

執行檔位於 `bin/cephtower`。未傳入 `-config` 時預設讀取
`/opt/cephtower/config/config.yaml`；該檔案必須在啟動前存在。

## 5. 設定

所有選項與預設值請以 [config/config.yaml](../../config/config.yaml) 為準。

| 區段 | 用途 |
|---|---|
| `server` | 監聽位址、連接埠與執行目錄（預設 `0.0.0.0:36900`、`/opt/cephtower`） |
| `log` | 輸出、層級、格式、輪替與保留時間 |
| `runtime` | Ceph 設定、keyring 等執行階段檔案目錄 |
| `database` | SQLite 檔案或 MySQL 連線/TLS；啟動時自動遷移 |
| `smtp` | 選用的密碼重設郵件服務 |

Ceph 叢集憑證不寫入此 YAML，而是在初始化後經由叢集管理儲存至資料庫。請限制設定、
資料庫與執行目錄的存取權，並在正式環境採用適當的 TLS 驗證。

## 6. 常用命令

| 命令 | 用途 |
|---|---|
| `make check-env` | 檢查 Go、Node.js 與 npm 版本 |
| `make run` | 同時啟動開發後端與前端 |
| `make run-backend` | 建置並啟動後端；以 `CONFIG` 指定設定檔 |
| `make run-frontend` | 在 `36901` 啟動 Vite |
| `make build` | 建置前端並產生內嵌 UI 的 `bin/cephtower` |
| `make build-frontend` | 型別檢查、建置前端並同步內嵌資源 |
| `make test` | 執行後端測試與前端建置驗證 |
| `make test-backend` | 執行 `go test ./...` |
| `make test-frontend` | 執行前端型別檢查與 Vite 建置驗證 |
| `ruby tools/generate_ceph_dashboard_client.rb` | 從本機資料重新產生 Ceph Dashboard 用戶端 |

可用 `CONFIG=/path/to/config.yaml` 覆寫後端設定，或以 `FRONTEND_PORT=連接埠`
覆寫 `make run` 使用的前端連接埠。

## 7. API 與文件

API 前綴為 `/api/v1`。基本免驗證端點包括：

| 方法 | 路徑 | 用途 |
|---|---|---|
| `GET` | `/api/v1/healthz` | 程序存活檢查 |
| `GET` | `/api/v1/readyz` | 初始化就緒檢查 |
| `GET` | `/api/v1/setup/status` | 首次啟動狀態 |
| `POST` | `/api/v1/auth/login` | 登入並取得 Token |

除初始化、登入及密碼重設端點外，請求需要 `Authorization: Bearer <token>`。
路由位於 `backend/internal/api/v1/router/`；Ceph 整合範圍與相容性請參閱
[docs/ceph/apis/index.md](../ceph/apis/index.md)。

## 8. 開發與貢獻

- 後端修改執行 `make test-backend`；前端修改執行 `make test-frontend`。
- 請勿提交 `app/` 中的本機資料、資料庫、日誌或叢集金鑰。
- 提交訊息遵循 [docs/commit-convention.md](../commit-convention.md)。
- 歡迎 Issue 與 Pull Request；請明確標示已驗證及仍為預留的功能。

## 9. 授權條款

CephTower 採用 [MIT License](../../LICENSE)。

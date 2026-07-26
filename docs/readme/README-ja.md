<div align="center">

<img src="../../frontend/public/ceph-tower-logo.svg" alt="CephTower logo" width="128" height="128">

# CephTower

Ceph クラスター用 Web 管理コンソール

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](../../backend/go.mod)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](../../frontend/package.json)
[![Ceph](https://img.shields.io/badge/Ceph-Dashboard%20API-EF5C55)](https://docs.ceph.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](../../LICENSE)

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [English](README-en.md) | [**日本語**](README-ja.md) | [Français](README-fr.md) | [Deutsch](README-de.md) | [Español](README-es.md) | [Português](README-pt.md) | [Русский](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower は Go バックエンドと React / Ant Design フロントエンドを組み合わせ、
Ceph Dashboard API と Ceph コマンドを通じて複数の Ceph クラスターを管理します。
バックエンドはバージョン付き REST API、永続化、バックグラウンド収集、組み込み
Web UI を提供し、フロントエンドは同一オリジンの `/api` のみを使用します。

## 1. 現在の機能と状態

- 初回起動ウィザード：SQLite/MySQL の選択、接続テスト、管理者作成。
- 認証：12 時間の Bearer Token セッション、管理者/一般ユーザーロール、読み取り・ユーザー管理権限。SMTP 設定時はメールコードでパスワードをリセット可能。
- 複数クラスター接続：MON アドレス、`client.admin` キー、Dashboard 認証情報を保存し、ホスト、デーモン、サービス、MON、MGR、MDS、OSD、Mgr モジュール、クラスター設定を自動検出・キャッシュ。
- クラスター UI：接続と詳細、ホスト、MON、MGR、OSD、MDS 管理。Mgr モジュール切替、デーモン操作、OSD in/out、reweight、scrub をサポート。
- データ収集：モジュールごとに取得元、周期、タイムアウト、再試行、優先度を設定し、手動実行と履歴確認が可能。
- バックエンド統合：クラスター、Pool/RBD、CephFS/NFS/SMB、RGW、iSCSI、NVMe-oF、Prometheus/Grafana、Dashboard のユーザー、ロール、設定 API。
- 本番ビルドではフロントエンドを Go 実行ファイルに埋め込み、単一 HTTP サービスで UI と API を配信。

> [!IMPORTANT]
> 本プロジェクトは開発中です。クラスター管理、ユーザー管理、収集設定は実バックエンドに
> 接続済みです。概要とシステム情報にはデモデータが含まれ、ブロック、ファイル、
> オブジェクト、監視ページは主にワークフローのプレースホルダーです。バックエンド統合が
> 存在しても、すべてのフロントエンド操作が完成しているとは限りません。

## 2. プロジェクト構成

```text
CephTower/
├── backend/
│   ├── cmd/main.go              # プロセスのエントリーポイント
│   └── internal/
│       ├── api/v1/              # REST ルートとハンドラー
│       ├── service/             # 認証、クラスター、収集、設定、初期化
│       ├── store/               # GORM、マイグレーション、SQLite/MySQL
│       ├── integration/ceph/    # Ceph Dashboard・コマンドクライアント
│       ├── task/                # バックグラウンドジョブとスケジュール
│       └── webui/               # 組み込みフロントエンド資産
├── frontend/src/                # React コンソール、ルート、ページ、API クライアント
├── config/config.yaml           # コメント付きリファレンス設定
├── docs/                        # アーキテクチャ、Ceph 資料、多言語 README
├── Makefile                     # 開発、テスト、ビルド
└── README.md
```

詳細は [docs/architecture.md](../architecture.md) を参照してください。

## 3. 動作要件

| ツール/サービス | 最低要件 | 用途 |
|---|---:|---|
| Go | 1.26 | バックエンドのビルドとテスト |
| Node.js | 20 | フロントエンド開発とビルド |
| npm | 10 | 依存関係管理 |
| C ツールチェーン | OS 対応版 | CGO SQLite ドライバーに必要 |
| Ceph | Dashboard API 有効 | MON アドレスと十分な権限の keyring も必要 |
| MySQL | 任意 | デフォルトの SQLite を使わない場合 |

## 4. クイックスタート

リポジトリルートで実行します。

```bash
make run
```

環境確認、必要時の依存関係インストール、`config/config.yaml` からの
`app/config/config.yaml` 作成（開発用実行ディレクトリは `./app`）を行い、次を起動します。

- バックエンドと本番 Web エントリー：<http://localhost:36900>
- Vite 開発サーバー：<http://localhost:36901>（`/api` をバックエンドへプロキシ）

初回アクセスは `/initialize` に移動します。データベースと管理者を設定後、クラスター
管理で Ceph 接続を追加してください。個別起動は先に `make ensure-run-config` を実行し、
2 つのターミナルで次を実行します。

```bash
make run-backend
make run-frontend
```

### 本番ビルド

```bash
make build
./bin/cephtower -config /path/to/config.yaml
```

実行ファイルは `bin/cephtower` です。`-config` 省略時は
`/opt/cephtower/config/config.yaml` を読み、起動前にファイルが必要です。

## 5. 設定

全オプションと既定値は [config/config.yaml](../../config/config.yaml) を参照してください。

| セクション | 用途 |
|---|---|
| `server` | リッスン先、ポート、実行ディレクトリ（既定：`0.0.0.0:36900`、`/opt/cephtower`） |
| `log` | 出力、レベル、形式、ローテーション、保持期間 |
| `runtime` | Ceph 設定、keyring などの実行時ファイル |
| `database` | SQLite または MySQL 接続/TLS。起動時に自動マイグレーション |
| `smtp` | パスワードリセット用の任意メールサービス |

Ceph 認証情報は YAML ではなく、初期化後にクラスター管理からデータベースへ保存されます。
設定、DB、実行時ディレクトリへのアクセスを制限し、本番では適切な TLS 検証を使用してください。

## 6. 主なコマンド

| コマンド | 用途 |
|---|---|
| `make check-env` | Go、Node.js、npm のバージョン確認 |
| `make run` | 開発用バックエンドとフロントエンドを同時起動 |
| `make run-backend` | バックエンドをビルド・起動。`CONFIG` で設定を指定 |
| `make run-frontend` | ポート `36901` で Vite を起動 |
| `make build` | フロントエンドをビルドし、UI 内蔵の `bin/cephtower` を生成 |
| `make build-frontend` | 型検査、ビルド、組み込み資産の同期 |
| `make test` | バックエンドテストとフロントエンドビルド検証を実行 |
| `make test-backend` | `go test ./...` を実行 |
| `make test-frontend` | 型検査とフロントエンドビルドを実行 |

`CONFIG=/path/to/config.yaml` で設定を、`FRONTEND_PORT=ポート` で `make run` の
フロントエンドポートを変更できます。

## 7. API とドキュメント

API プレフィックスは `/api/v1` です。基本的な認証不要エンドポイント：

| メソッド | パス | 用途 |
|---|---|---|
| `GET` | `/api/v1/healthz` | プロセス生存確認 |
| `GET` | `/api/v1/readyz` | 初期化完了確認 |
| `GET` | `/api/v1/setup/status` | 初回起動状態 |
| `POST` | `/api/v1/auth/login` | ログインと Token 取得 |

初期化、ログイン、パスワードリセット以外は `Authorization: Bearer <token>` が必要です。
ルートは `backend/internal/api/v1/router/`、Ceph 統合範囲と互換性は
[docs/ceph/apis/index.md](../ceph/apis/index.md) を参照してください。

## 8. 開発とコントリビューション

- バックエンド変更は `make test-backend`、フロントエンド変更は `make test-frontend`。
- `app/` のローカルデータ、DB、ログ、クラスタ―キーをコミットしないでください。
- コミットは [docs/commit-convention.md](../commit-convention.md) に従います。
- Issue と Pull Request を歓迎します。検証済み機能とプレースホルダーを明記してください。

## 9. ライセンス

CephTower は [MIT License](../../LICENSE) で提供されます。

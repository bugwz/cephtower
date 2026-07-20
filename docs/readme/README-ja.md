<div align="center">

<img src="../../frontend/public/ceph-tower-logo.svg" alt="CephTower logo" width="128" height="128">

# CephTower

</div>

<div align="center">

[![Go](https://img.shields.io/badge/Go-Backend-00ADD8?logo=go)](../../backend/go.mod)
[![React](https://img.shields.io/badge/React-Frontend-61DAFB?logo=react)](../../frontend/package.json)
[![Ceph](https://img.shields.io/badge/Ceph-Dashboard%20API-EF5C55)](https://docs.ceph.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](../../LICENSE)
[![多言語](https://img.shields.io/badge/多言語-yellow)](../../README.md)

</div>

<div align="center">

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [English](README-en.md) | [**日本語**](README-ja.md) | [Français](README-fr.md) | [Deutsch](README-de.md) | [Español](README-es.md) | [Português](README-pt.md) | [Русский](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower は、Ceph Manager Dashboard API を通じて Ceph クラスターを管理するための Go バックエンドと React / Ant Design フロントエンドのプロジェクトです。

## 1. 機能

- ヘルスチェックとクラスター概要 API を備えた Go HTTP サービス。
- React、Ant Design、Vite、TypeScript による管理コンソール。
- Ceph Dashboard API の認証と今後の管理機能を分離したクライアント層。
- MIT ライセンス。

## 2. クイックスタート

```bash
make backend-dev
```

```bash
cd frontend
npm install
npm run dev
```

## 3. プロジェクト構成

```text
backend/     Go API サービス
frontend/    React Web コンソール
  public/ceph-tower-logo.svg    README とフロントエンド共通のロゴ
docs/        アーキテクチャ、多言語 README、ローカル参考資料
```

## 4. ライセンス

MIT License。詳細は [LICENSE](../../LICENSE) を参照してください。

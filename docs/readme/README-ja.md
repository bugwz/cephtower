<div align="center">

# CephTower

</div>

<div align="center">

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [English](README-en.md) | [**日本語**](README-ja.md) | [Français](README-fr.md) | [Deutsch](README-de.md) | [Español](README-es.md) | [Português](README-pt.md) | [Русский](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower は、Ceph Manager Dashboard API を通じて Ceph クラスターを管理するための Go バックエンドと Vue フロントエンドのプロジェクトです。

## 1. 機能

- ヘルスチェックとクラスター概要 API を備えた Go HTTP サービス。
- Vue 3、Vite、TypeScript による管理コンソール。
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
frontend/    Vue Web コンソール
docs/        アーキテクチャ、多言語 README、ローカル参考資料
```

## 4. ライセンス

MIT License。詳細は [LICENSE](../../LICENSE) を参照してください。

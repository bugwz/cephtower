<div align="center">

<img src="../../frontend/public/ceph-tower-logo.svg" alt="CephTower logo" width="128" height="128">

# CephTower

</div>

<div align="center">

[![Go](https://img.shields.io/badge/Go-Backend-00ADD8?logo=go)](../../backend/go.mod)
[![React](https://img.shields.io/badge/React-Frontend-61DAFB?logo=react)](../../frontend/package.json)
[![Ceph](https://img.shields.io/badge/Ceph-Dashboard%20API-EF5C55)](https://docs.ceph.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](../../LICENSE)
[![다국어](https://img.shields.io/badge/다국어-yellow)](../../README.md)

</div>

<div align="center">

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [English](README-en.md) | [日本語](README-ja.md) | [Français](README-fr.md) | [Deutsch](README-de.md) | [Español](README-es.md) | [Português](README-pt.md) | [Русский](README-ru.md) | [**한국어**](README-ko.md)

</div>

CephTower는 Ceph Manager Dashboard API를 통해 Ceph 클러스터를 관리하기 위한 Go 백엔드와 React 및 Ant Design 프런트엔드 프로젝트입니다.

## 1. 기능

- 상태 확인과 클러스터 요약 엔드포인트를 제공하는 Go HTTP API 서비스.
- React, Ant Design, Vite, TypeScript 기반 관리 콘솔.
- Ceph Dashboard API 인증과 호출을 분리한 클라이언트 계층.
- MIT 라이선스.

## 2. 빠른 시작

```bash
make backend-dev
```

```bash
cd frontend
npm install
npm run dev
```

## 3. 프로젝트 구조

```text
backend/     Go API 서비스
frontend/    React Web 콘솔
  public/ceph-tower-logo.svg    README와 frontend가 공유하는 로고
docs/        아키텍처 문서, 다국어 README 파일, 로컬 참고 자료
```

## 4. 라이선스

MIT License. [LICENSE](../../LICENSE)를 참고하세요.

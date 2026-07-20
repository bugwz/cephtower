<div align="center">

<img src="../../frontend/public/ceph-tower-logo.svg" alt="CephTower logo" width="128" height="128">

# CephTower

</div>

<div align="center">

[![Go](https://img.shields.io/badge/Go-Backend-00ADD8?logo=go)](../../backend/go.mod)
[![React](https://img.shields.io/badge/React-Frontend-61DAFB?logo=react)](../../frontend/package.json)
[![Ceph](https://img.shields.io/badge/Ceph-Dashboard%20API-EF5C55)](https://docs.ceph.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](../../LICENSE)
[![Multilíngue](https://img.shields.io/badge/Multilíngue-yellow)](../../README.md)

</div>

<div align="center">

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [English](README-en.md) | [日本語](README-ja.md) | [Français](README-fr.md) | [Deutsch](README-de.md) | [Español](README-es.md) | [**Português**](README-pt.md) | [Русский](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower é um projeto com backend em Go e frontend em React com Ant Design para administrar clusters Ceph por meio da API Ceph Manager Dashboard.

## 1. Funcionalidades

- Serviço HTTP em Go com endpoints de saúde e resumo do cluster.
- Console de administração com React, Ant Design, Vite e TypeScript.
- Camada cliente para autenticação e chamadas à API Ceph Dashboard.
- Licença MIT.

## 2. Início rápido

```bash
make backend-dev
```

```bash
cd frontend
npm install
npm run dev
```

## 3. Estrutura do projeto

```text
backend/     Serviço API Go
frontend/    Console Web React
  public/ceph-tower-logo.svg    Logo compartilhado entre README e frontend
docs/        Notas de arquitetura, README multilíngues e referências locais
```

## 4. Licença

MIT License. Consulte [LICENSE](../../LICENSE).

<div align="center">

<img src="../../frontend/public/ceph-tower-logo.svg" alt="CephTower logo" width="128" height="128">

# CephTower

Console Web de administração para clusters Ceph

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](../../backend/go.mod)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](../../frontend/package.json)
[![Ceph](https://img.shields.io/badge/Ceph-Dashboard%20API-EF5C55)](https://docs.ceph.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](../../LICENSE)

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [English](README-en.md) | [日本語](README-ja.md) | [Français](README-fr.md) | [Deutsch](README-de.md) | [Español](README-es.md) | [**Português**](README-pt.md) | [Русский](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower combina um backend Go com um frontend React / Ant Design para administrar um ou vários
clusters Ceph pela API Ceph Dashboard e por comandos Ceph. O backend fornece API REST versionada,
persistência, coleta em segundo plano e Web UI incorporada. O frontend sempre acessa o backend pelo
caminho de mesma origem `/api`.

## 1. Recursos e estado atuais

- Assistente inicial para escolher SQLite ou MySQL, testar a conexão e criar o administrador.
- Autenticação com sessões Bearer Token de 12 horas, perfis administrador/usuário, permissões granulares de leitura e gestão de usuários e redefinição por código de e-mail quando SMTP está configurado.
- Conexões multicluster com endereços MON, chave `client.admin` e credenciais Dashboard; descoberta e cache automáticos de hosts, daemons, serviços, MON, MGR, MDS, OSD, módulos Mgr e configuração.
- Interface de cluster para conexões, detalhes, hosts, MON, MGR, OSD e MDS; inclui módulos Mgr, ações de daemon e operações OSD in/out, reweight e scrub.
- Coleta configurável por módulo: fonte, intervalo, timeout, tentativas e prioridade, com execução manual e histórico.
- Integrações backend para cluster, Pool/RBD, CephFS/NFS/SMB, RGW, iSCSI, NVMe-oF, Prometheus/Grafana e usuários, perfis e configurações Dashboard.
- O build de produção incorpora o frontend no executável Go; um serviço HTTP fornece UI e API.

> [!IMPORTANT]
> O projeto está em desenvolvimento ativo. Gestão de clusters e usuários e configuração de coleta
> usam o backend real. As páginas de visão geral e informações do sistema ainda contêm dados de
> demonstração; páginas de bloco, arquivo, objeto e monitoramento são principalmente marcadores de
> fluxo. Uma integração backend não significa que toda ação frontend esteja concluída.

## 2. Estrutura do projeto

```text
CephTower/
├── backend/
│   ├── cmd/main.go              # entrada do processo
│   └── internal/
│       ├── api/v1/              # rotas e handlers REST
│       ├── service/             # autenticação, cluster, coleta, configurações e setup
│       ├── store/               # GORM, migrações e SQLite/MySQL
│       ├── integration/ceph/    # clientes Dashboard e comandos Ceph
│       ├── task/                # tarefas em segundo plano e agendamento
│       └── webui/               # recursos frontend incorporados
├── frontend/src/                # console React, rotas, páginas e clientes API
├── config/config.yaml           # configuração de referência comentada
├── docs/                        # arquitetura, referências Ceph e README traduzidos
├── Makefile                     # desenvolvimento, testes e build
└── README.md
```

Veja [docs/architecture.md](../architecture.md) para camadas e ciclo de vida.

## 3. Requisitos

| Ferramenta/serviço | Mínimo | Uso |
|---|---:|---|
| Go | 1.26 | build e testes backend |
| Node.js | 20 | desenvolvimento e build frontend |
| npm | 10 | dependências frontend |
| Toolchain C | adequada ao SO | necessária para o driver SQLite CGO |
| Ceph | API Dashboard habilitada | requer também endereços MON e keyring com privilégios suficientes |
| MySQL | opcional | somente quando SQLite padrão não for usado |

## 4. Início rápido

Na raiz do repositório:

```bash
make run
```

O comando verifica o ambiente, instala dependências quando necessário e cria
`app/config/config.yaml` a partir de `config/config.yaml` se ausente (diretório de execução `./app`). Inicia:

- Backend e entrada Web de produção: <http://localhost:36900>
- Servidor Vite: <http://localhost:36901> (`/api` é encaminhado ao backend)

A primeira visita redireciona para `/initialize`. Configure banco e administrador e adicione uma
conexão Ceph. Para iniciar separadamente, execute primeiro `make ensure-run-config` e depois, em dois terminais:

```bash
make run-backend
make run-frontend
```

### Build de produção

```bash
make build
./bin/cephtower -config /path/to/config.yaml
```

O executável fica em `bin/cephtower`. Sem `-config`, lê
`/opt/cephtower/config/config.yaml`, que deve existir antes da inicialização.

## 5. Configuração

Veja [config/config.yaml](../../config/config.yaml) para opções e padrões.

| Seção | Uso |
|---|---|
| `server` | endereço, porta e diretório de execução (padrões `0.0.0.0:36900`, `/opt/cephtower`) |
| `log` | saída, nível, formato, rotação e retenção |
| `runtime` | configuração Ceph, keyrings e outros arquivos de execução |
| `database` | SQLite ou conexão/TLS MySQL; migrações automáticas na inicialização |
| `smtp` | serviço opcional de e-mail para redefinição de senha |

Credenciais Ceph não ficam neste YAML: após a inicialização são salvas no banco pela gestão de clusters.
Restrinja acesso à configuração, banco e arquivos de execução e use validação TLS adequada em produção.

## 6. Comandos comuns

| Comando | Uso |
|---|---|
| `make check-env` | verificar versões de Go, Node.js e npm |
| `make run` | iniciar juntos backend e frontend de desenvolvimento |
| `make run-backend` | compilar/iniciar backend; `CONFIG` escolhe a configuração |
| `make run-frontend` | iniciar Vite na porta `36901` |
| `make build` | compilar frontend e criar `bin/cephtower` com UI incorporada |
| `make build-frontend` | verificar tipos, compilar e sincronizar recursos incorporados |
| `make test` | executar testes backend e validar o build frontend |
| `make test-backend` | executar `go test ./...` |
| `make test-frontend` | verificar tipos e compilar o frontend |

Use `CONFIG=/path/to/config.yaml` para a configuração backend ou `FRONTEND_PORT=porta` para a
porta frontend usada por `make run`.

## 7. API e documentação

O prefixo é `/api/v1`. Endpoints básicos sem autenticação:

| Método | Caminho | Uso |
|---|---|---|
| `GET` | `/api/v1/healthz` | atividade do processo |
| `GET` | `/api/v1/readyz` | prontidão após inicialização |
| `GET` | `/api/v1/setup/status` | estado do primeiro início |
| `POST` | `/api/v1/auth/login` | login e obtenção do Token |

Exceto setup, login e redefinição, as requisições exigem `Authorization: Bearer <token>`.
As rotas ficam em `backend/internal/api/v1/router/`. Escopo e compatibilidade Ceph:
[docs/ceph/apis/index.md](../ceph/apis/index.md).

## 8. Desenvolvimento e contribuição

- Execute `make test-backend` para backend e `make test-frontend` para frontend.
- Não envie dados, bancos, logs ou chaves de cluster locais de `app/`.
- Siga [docs/commit-convention.md](../commit-convention.md) nos commits.
- Issues e Pull Requests são bem-vindos; diferencie recursos verificados e marcadores.

## 9. Licença

CephTower é distribuído sob a [licença MIT](../../LICENSE).

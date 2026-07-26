<div align="center">

<img src="../../frontend/public/ceph-tower-logo.svg" alt="CephTower logo" width="128" height="128">

# CephTower

Ceph 클러스터용 웹 관리 콘솔

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](../../backend/go.mod)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](../../frontend/package.json)
[![Ceph](https://img.shields.io/badge/Ceph-Dashboard%20API-EF5C55)](https://docs.ceph.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](../../LICENSE)

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [English](README-en.md) | [日本語](README-ja.md) | [Français](README-fr.md) | [Deutsch](README-de.md) | [Español](README-es.md) | [Português](README-pt.md) | [Русский](README-ru.md) | [**한국어**](README-ko.md)

</div>

CephTower는 Go 백엔드와 React / Ant Design 프런트엔드를 사용하여 Ceph Dashboard API와
Ceph 명령으로 하나 이상의 Ceph 클러스터를 관리합니다. 백엔드는 버전이 지정된 REST API,
영속성, 백그라운드 수집 및 내장 Web UI를 제공하며 프런트엔드는 동일 출처 `/api`를 사용합니다.

## 1. 현재 기능 및 상태

- 최초 실행 마법사: SQLite/MySQL 선택, 연결 테스트, 관리자 계정 생성.
- 인증: 12시간 Bearer Token 세션, 관리자/일반 사용자 역할, 세분화된 읽기 및 사용자 관리 권한. SMTP 구성 시 이메일 코드로 비밀번호 재설정 가능.
- 다중 클러스터 연결: MON 주소, `client.admin` 키와 Dashboard 자격 증명을 저장하고 호스트, 데몬, 서비스, MON, MGR, MDS, OSD, Mgr 모듈 및 클러스터 설정을 자동 검색·캐시.
- 클러스터 UI: 연결/상세 정보, 호스트, MON, MGR, OSD, MDS 관리. Mgr 모듈 전환, 데몬 작업, OSD in/out, reweight, scrub 지원.
- 데이터 수집: 모듈별 소스, 주기, 시간 제한, 재시도, 우선순위를 설정하고 즉시 실행 및 기록 확인 가능.
- 백엔드 통합: 클러스터, Pool/RBD, CephFS/NFS/SMB, RGW, iSCSI, NVMe-oF, Prometheus/Grafana, Dashboard 사용자·역할·설정 API.
- 프로덕션 빌드는 프런트엔드를 Go 실행 파일에 내장하여 하나의 HTTP 서비스로 UI와 API를 제공.

> [!IMPORTANT]
> 프로젝트는 개발 중입니다. 클러스터 관리, 사용자 관리 및 수집 설정은 실제 백엔드에
> 연결되어 있습니다. 개요와 시스템 정보에는 데모 데이터가 있으며 블록, 파일, 오브젝트,
> 모니터링 페이지는 주로 워크플로 자리 표시자입니다. 백엔드 통합이 있다고 해서 모든
> 프런트엔드 작업이 완성된 것은 아닙니다.

## 2. 프로젝트 구조

```text
CephTower/
├── backend/
│   ├── cmd/main.go              # 프로세스 진입점
│   └── internal/
│       ├── api/v1/              # REST 라우트와 핸들러
│       ├── service/             # 인증, 클러스터, 수집, 설정, 초기화
│       ├── store/               # GORM, 마이그레이션, SQLite/MySQL
│       ├── integration/ceph/    # Ceph Dashboard 및 명령 클라이언트
│       ├── task/                # 백그라운드 작업과 스케줄링
│       └── webui/               # 내장 프런트엔드 자산
├── frontend/src/                # React 콘솔, 라우트, 페이지, API 클라이언트
├── config/config.yaml           # 주석이 포함된 참조 구성
├── docs/                        # 아키텍처, Ceph 자료, 다국어 README
├── Makefile                     # 개발, 테스트, 빌드 진입점
└── README.md
```

자세한 계층과 수명 주기는 [docs/architecture.md](../architecture.md)를 참조하세요.

## 3. 요구 사항

| 도구/서비스 | 최소 버전 | 용도 |
|---|---:|---|
| Go | 1.26 | 백엔드 빌드와 테스트 |
| Node.js | 20 | 프런트엔드 개발과 빌드 |
| npm | 10 | 의존성 관리 |
| C 툴체인 | OS에 맞는 버전 | CGO SQLite 드라이버에 필요 |
| Ceph | Dashboard API 활성화 | MON 주소와 충분한 권한의 keyring도 필요 |
| MySQL | 선택 사항 | 기본 SQLite를 사용하지 않을 때 필요 |

## 4. 빠른 시작

저장소 루트에서 실행합니다.

```bash
make run
```

환경을 검사하고 필요 시 프런트엔드 의존성을 설치하며, 없으면 `config/config.yaml`에서
`app/config/config.yaml`을 생성합니다(개발 런타임 디렉터리 `./app`). 이후 다음을 시작합니다.

- 백엔드 및 프로덕션 Web 진입점: <http://localhost:36900>
- Vite 개발 서버: <http://localhost:36901> (`/api`를 백엔드로 프록시)

첫 방문은 `/initialize`로 이동합니다. DB와 관리자를 설정한 후 클러스터 관리에서 Ceph
연결을 추가하세요. 별도로 실행하려면 먼저 `make ensure-run-config`를 실행하고 두 터미널에서:

```bash
make run-backend
make run-frontend
```

### 프로덕션 빌드

```bash
make build
./bin/cephtower -config /path/to/config.yaml
```

실행 파일은 `bin/cephtower`에 생성됩니다. `-config`를 생략하면
`/opt/cephtower/config/config.yaml`을 읽으며 시작 전에 파일이 존재해야 합니다.

## 5. 구성

전체 옵션과 기본값은 [config/config.yaml](../../config/config.yaml)을 참조하세요.

| 섹션 | 용도 |
|---|---|
| `server` | 수신 주소, 포트, 런타임 디렉터리(기본 `0.0.0.0:36900`, `/opt/cephtower`) |
| `log` | 출력, 수준, 형식, 순환, 보존 기간 |
| `runtime` | Ceph 구성, keyring 등 런타임 파일 디렉터리 |
| `database` | SQLite 파일 또는 MySQL 연결/TLS. 시작 시 자동 마이그레이션 |
| `smtp` | 선택적 비밀번호 재설정 메일 서비스 |

Ceph 자격 증명은 YAML이 아니라 초기화 후 클러스터 관리를 통해 DB에 저장됩니다. 구성,
DB 및 런타임 디렉터리 접근을 제한하고 프로덕션에서 적절한 TLS 검증을 사용하세요.

## 6. 주요 명령

| 명령 | 용도 |
|---|---|
| `make check-env` | Go, Node.js, npm 버전 검사 |
| `make run` | 개발 백엔드와 프런트엔드 동시 시작 |
| `make run-backend` | 백엔드 빌드 및 시작. `CONFIG`로 구성 지정 |
| `make run-frontend` | 포트 `36901`에서 Vite 시작 |
| `make build` | 프런트엔드 빌드 및 UI 내장 `bin/cephtower` 생성 |
| `make build-frontend` | 타입 검사, 빌드, 내장 자산 동기화 |
| `make test` | 백엔드 테스트와 프런트엔드 빌드 검증 실행 |
| `make test-backend` | `go test ./...` 실행 |
| `make test-frontend` | 프런트엔드 타입 검사와 빌드 실행 |
| `ruby tools/generate_ceph_dashboard_client.rb` | 로컬 자료에서 Ceph Dashboard 클라이언트 재생성 |

`CONFIG=/path/to/config.yaml`로 백엔드 구성을, `FRONTEND_PORT=포트`로 `make run`의
프런트엔드 포트를 변경할 수 있습니다.

## 7. API 및 문서

API 접두사는 `/api/v1`입니다. 기본 비인증 엔드포인트:

| 메서드 | 경로 | 용도 |
|---|---|---|
| `GET` | `/api/v1/healthz` | 프로세스 생존 확인 |
| `GET` | `/api/v1/readyz` | 초기화 준비 확인 |
| `GET` | `/api/v1/setup/status` | 최초 실행 상태 |
| `POST` | `/api/v1/auth/login` | 로그인 및 Token 발급 |

초기화, 로그인, 비밀번호 재설정 외에는 `Authorization: Bearer <token>`이 필요합니다.
라우트는 `backend/internal/api/v1/router/`에 있으며 Ceph 통합 범위와 호환성은
[docs/ceph/apis/index.md](../ceph/apis/index.md)를 참조하세요.

## 8. 개발 및 기여

- 백엔드 변경에는 `make test-backend`, 프런트엔드 변경에는 `make test-frontend`를 실행하세요.
- `app/`의 로컬 데이터, DB, 로그 또는 클러스터 키를 커밋하지 마세요.
- 커밋은 [docs/commit-convention.md](../commit-convention.md)를 따릅니다.
- Issue와 Pull Request를 환영합니다. 검증된 기능과 자리 표시자 기능을 명확히 표시하세요.

## 9. 라이선스

CephTower는 [MIT License](../../LICENSE)로 제공됩니다.

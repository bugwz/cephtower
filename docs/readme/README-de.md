<div align="center">

<img src="../../frontend/public/ceph-tower-logo.svg" alt="CephTower logo" width="128" height="128">

# CephTower

Web-Verwaltungskonsole für Ceph-Cluster

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](../../backend/go.mod)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](../../frontend/package.json)
[![Ceph](https://img.shields.io/badge/Ceph-Dashboard%20API-EF5C55)](https://docs.ceph.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](../../LICENSE)

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [English](README-en.md) | [日本語](README-ja.md) | [Français](README-fr.md) | [**Deutsch**](README-de.md) | [Español](README-es.md) | [Português](README-pt.md) | [Русский](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower kombiniert ein Go-Backend mit einem React-/Ant-Design-Frontend, um einen oder mehrere
Ceph-Cluster über die Ceph Dashboard API und Ceph-Befehle zu verwalten. Das Backend bietet eine
versionierte REST-API, Persistenz, Hintergrundsammlung und eine eingebettete Weboberfläche. Das
Frontend greift stets über den Same-Origin-Pfad `/api` auf das Backend zu.

## 1. Aktuelle Funktionen und Status

- Ersteinrichtungsassistent für SQLite/MySQL, Verbindungstest und Administratorkonto.
- Authentifizierung mit 12-stündigen Bearer-Token-Sitzungen, Administrator-/Benutzerrollen, granularen Lese- und Benutzerverwaltungsrechten sowie E-Mail-Code-Passwortreset bei konfiguriertem SMTP.
- Multi-Cluster-Verbindungen speichern MON-Adressen, `client.admin`-Schlüssel und Dashboard-Zugangsdaten; Hosts, Daemons, Dienste, MONs, MGRs, MDSs, OSDs, Mgr-Module und Clusterkonfiguration werden automatisch erkannt und zwischengespeichert.
- Clusteroberfläche für Verbindungen, Details, Hosts, MON, MGR, OSD und MDS; einschließlich Mgr-Modulschaltern, Daemon-Aktionen und OSD in/out, reweight und scrub.
- Datensammlung mit Quelle, Intervall, Timeout, Wiederholung und Priorität pro Modul sowie manueller Ausführung und Verlauf.
- Backend-Integrationen für Cluster, Pool/RBD, CephFS/NFS/SMB, RGW, iSCSI, NVMe-oF, Prometheus/Grafana sowie Dashboard-Benutzer, Rollen und Konfiguration.
- Produktionsbuilds betten das Frontend in die Go-Datei ein; ein HTTP-Dienst liefert UI und API.

> [!IMPORTANT]
> Das Projekt wird aktiv entwickelt. Cluster- und Benutzerverwaltung sowie Sammlungseinstellungen
> verwenden das echte Backend. Übersicht und Systeminformationen enthalten noch Demodaten;
> Block-, Datei-, Objekt- und Monitoringseiten sind überwiegend Workflow-Platzhalter. Eine vorhandene
> Backend-Integration bedeutet nicht, dass jede Frontend-Aktion fertiggestellt ist.

## 2. Projektstruktur

```text
CephTower/
├── backend/
│   ├── cmd/main.go              # Prozesseinstieg
│   └── internal/
│       ├── api/v1/              # REST-Routen und Handler
│       ├── service/             # Auth-, Cluster-, Sammlung-, Einstellungs- und Setup-Logik
│       ├── store/               # GORM, Migrationen und SQLite/MySQL
│       ├── integration/ceph/    # Ceph-Dashboard- und Befehlsclients
│       ├── task/                # Hintergrundaufgaben und Planung
│       └── webui/               # eingebettete Frontend-Ressourcen
├── frontend/src/                # React-Konsole, Routen, Seiten und API-Clients
├── config/config.yaml           # kommentierte Referenzkonfiguration
├── docs/                        # Architektur, Ceph-Referenzen und übersetzte READMEs
├── Makefile                     # Entwicklung, Tests und Build
└── README.md
```

Details zu Schichten und Lebenszyklus: [docs/architecture.md](../architecture.md).

## 3. Voraussetzungen

| Werkzeug/Dienst | Minimum | Zweck |
|---|---:|---|
| Go | 1.26 | Backend-Builds und Tests |
| Node.js | 20 | Frontend-Entwicklung und Builds |
| npm | 10 | Frontend-Abhängigkeiten |
| C-Toolchain | passend zum Betriebssystem | für den CGO-SQLite-Treiber |
| Ceph | Dashboard API aktiviert | zusätzlich MON-Adressen und ausreichend berechtigtes keyring |
| MySQL | optional | nur ohne die standardmäßige SQLite-Datenbank |

## 4. Schnellstart

Im Repository-Stammverzeichnis:

```bash
make run
```

Der Befehl prüft die Umgebung, installiert bei Bedarf Frontend-Abhängigkeiten und erzeugt fehlendes
`app/config/config.yaml` aus `config/config.yaml` (Entwicklungslaufzeitverzeichnis `./app`). Danach starten:

- Backend und Produktions-Webeingang: <http://localhost:36900>
- Vite-Entwicklungsserver: <http://localhost:36901> (`/api` wird zum Backend weitergeleitet)

Der erste Aufruf führt zu `/initialize`. Datenbank und Administrator einrichten und anschließend eine
Ceph-Verbindung hinzufügen. Für getrennten Start zunächst `make ensure-run-config`, dann in zwei Terminals:

```bash
make run-backend
make run-frontend
```

### Produktionsbuild

```bash
make build
./bin/cephtower -config /path/to/config.yaml
```

Die Datei liegt unter `bin/cephtower`. Ohne `-config` wird
`/opt/cephtower/config/config.yaml` gelesen; sie muss vor dem Start existieren.

## 5. Konfiguration

Alle Optionen und Standardwerte stehen in [config/config.yaml](../../config/config.yaml).

| Abschnitt | Zweck |
|---|---|
| `server` | Adresse, Port und Laufzeitverzeichnis (Standard: `0.0.0.0:36900`, `/opt/cephtower`) |
| `log` | Ausgabe, Stufe, Format, Rotation und Aufbewahrung |
| `runtime` | Ceph-Konfiguration, keyrings und weitere Laufzeitdateien |
| `database` | SQLite-Datei oder MySQL-Verbindung/TLS; Migrationen beim Start |
| `smtp` | optionaler Maildienst für Passwort-Resets |

Ceph-Zugangsdaten stehen nicht in dieser YAML-Datei, sondern werden nach der Initialisierung über die
Clusterverwaltung in der Datenbank gespeichert. Zugriffe auf Konfiguration, Datenbank und Laufzeitdateien
einschränken und in Produktion eine geeignete TLS-Prüfung verwenden.

## 6. Häufige Befehle

| Befehl | Zweck |
|---|---|
| `make check-env` | Go-, Node.js- und npm-Versionen prüfen |
| `make run` | Entwicklungs-Backend und -Frontend gemeinsam starten |
| `make run-backend` | Backend bauen/starten; Konfiguration mit `CONFIG` wählen |
| `make run-frontend` | Vite auf Port `36901` starten |
| `make build` | Frontend bauen und `bin/cephtower` mit eingebetteter UI erstellen |
| `make build-frontend` | Typprüfung, Build und Synchronisierung eingebetteter Ressourcen |
| `make test` | Backend-Tests und Frontend-Buildprüfung ausführen |
| `make test-backend` | `go test ./...` ausführen |
| `make test-frontend` | Frontend-Typprüfung und Build ausführen |

Mit `CONFIG=/path/to/config.yaml` wird die Backend-Konfiguration, mit `FRONTEND_PORT=Port` der
Frontend-Port von `make run` überschrieben.

## 7. API und Dokumentation

Das API-Präfix ist `/api/v1`. Grundlegende Endpunkte ohne Authentifizierung:

| Methode | Pfad | Zweck |
|---|---|---|
| `GET` | `/api/v1/healthz` | Prozess-Liveness |
| `GET` | `/api/v1/readyz` | Initialisierungsbereitschaft |
| `GET` | `/api/v1/setup/status` | Status der Ersteinrichtung |
| `POST` | `/api/v1/auth/login` | Anmeldung und Token-Ausgabe |

Außer Setup, Anmeldung und Passwortreset benötigen Anfragen `Authorization: Bearer <token>`.
Routen liegen in `backend/internal/api/v1/router/`; Umfang und Kompatibilität der Ceph-Integration:
[docs/ceph/apis/index.md](../ceph/apis/index.md).

## 8. Entwicklung und Beiträge

- Für Backend-Änderungen `make test-backend`, für Frontend-Änderungen `make test-frontend` ausführen.
- Lokale Daten, Datenbanken, Logs und Clusterschlüssel aus `app/` nicht committen.
- Commit-Regeln: [docs/commit-convention.md](../commit-convention.md).
- Issues und Pull Requests sind willkommen; verifizierte und Platzhalterfunktionen klar kennzeichnen.

## 9. Lizenz

CephTower steht unter der [MIT-Lizenz](../../LICENSE).

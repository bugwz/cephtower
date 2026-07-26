<div align="center">

<img src="../../frontend/public/ceph-tower-logo.svg" alt="CephTower logo" width="128" height="128">

# CephTower

Console Web d’administration de clusters Ceph

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](../../backend/go.mod)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](../../frontend/package.json)
[![Ceph](https://img.shields.io/badge/Ceph-Dashboard%20API-EF5C55)](https://docs.ceph.com/)
[![License](https://img.shields.io/badge/License-MIT-green)](../../LICENSE)

[简体中文](../../README.md) | [繁體中文](README-zh-TW.md) | [English](README-en.md) | [日本語](README-ja.md) | [**Français**](README-fr.md) | [Deutsch](README-de.md) | [Español](README-es.md) | [Português](README-pt.md) | [Русский](README-ru.md) | [한국어](README-ko.md)

</div>

CephTower associe un backend Go à un frontend React / Ant Design afin de gérer un ou plusieurs
clusters Ceph via l’API Ceph Dashboard et les commandes Ceph. Le backend fournit une API REST
versionnée, la persistance, la collecte en arrière-plan et une interface Web intégrée. Le frontend
accède toujours au backend via le chemin même origine `/api`.

## 1. Capacités et état actuels

- Assistant initial : choix de SQLite ou MySQL, test de connexion et création de l’administrateur.
- Authentification : sessions Bearer Token de 12 heures, rôles administrateur/utilisateur, droits fins de lecture et de gestion des utilisateurs ; réinitialisation par code e-mail si SMTP est configuré.
- Connexions multiclusters : stockage des adresses MON, de la clé `client.admin` et des identifiants Dashboard ; découverte et cache automatiques des hôtes, démons, services, MON, MGR, MDS, OSD, modules Mgr et de la configuration.
- Interface cluster : connexions et détails, hôtes, MON, MGR, OSD et MDS ; bascule des modules Mgr, actions sur les démons et opérations OSD in/out, reweight et scrub.
- Collecte : source, intervalle, délai, tentatives et priorité par module, exécution manuelle et historique.
- Intégrations backend : cluster, Pool/RBD, CephFS/NFS/SMB, RGW, iSCSI, NVMe-oF, Prometheus/Grafana, ainsi que les utilisateurs, rôles et paramètres Dashboard.
- Le build de production intègre le frontend dans l’exécutable Go ; un seul service HTTP fournit UI et API.

> [!IMPORTANT]
> Le projet est en développement actif. La gestion des clusters et des utilisateurs ainsi que la
> configuration de collecte utilisent le backend réel. Les pages de vue d’ensemble et d’informations
> système contiennent encore des données de démonstration ; les pages bloc, fichier, objet et supervision
> sont principalement des maquettes de workflow. Une intégration backend ne garantit pas que toutes les
> actions frontend soient terminées.

## 2. Structure du projet

```text
CephTower/
├── backend/
│   ├── cmd/main.go              # point d’entrée du processus
│   └── internal/
│       ├── api/v1/              # routes et handlers REST
│       ├── service/             # auth, cluster, collecte, paramètres et initialisation
│       ├── store/               # GORM, migrations et stockage SQLite/MySQL
│       ├── integration/ceph/    # clients Dashboard et commandes Ceph
│       ├── task/                # tâches en arrière-plan et planification
│       └── webui/               # ressources frontend intégrées
├── frontend/src/                # console React, routes, pages et clients API
├── config/config.yaml           # configuration de référence commentée
├── docs/                        # architecture, références Ceph et README traduits
├── Makefile                     # commandes de développement, test et build
└── README.md
```

Voir [docs/architecture.md](../architecture.md) pour les couches et le cycle de vie.

## 3. Prérequis

| Outil/service | Minimum | Utilisation |
|---|---:|---|
| Go | 1.26 | build et tests backend |
| Node.js | 20 | développement et build frontend |
| npm | 10 | gestion des dépendances frontend |
| Chaîne C | adaptée au système | requise par le pilote SQLite CGO |
| Ceph | API Dashboard activée | nécessite aussi des adresses MON et un keyring suffisamment privilégié |
| MySQL | facultatif | uniquement sans la base SQLite par défaut |

## 4. Démarrage rapide

Depuis la racine du dépôt :

```bash
make run
```

La commande vérifie l’environnement, installe les dépendances frontend si nécessaire et crée
`app/config/config.yaml` depuis `config/config.yaml` s’il manque (répertoire d’exécution `./app`), puis lance :

- Backend et entrée Web de production : <http://localhost:36900>
- Serveur Vite : <http://localhost:36901> (`/api` est relayé vers le backend)

La première visite redirige vers `/initialize`. Configurez la base et l’administrateur, puis ajoutez
une connexion Ceph. Pour lancer les services séparément, exécutez d’abord
`make ensure-run-config`, puis dans deux terminaux :

```bash
make run-backend
make run-frontend
```

### Build de production

```bash
make build
./bin/cephtower -config /path/to/config.yaml
```

L’exécutable se trouve dans `bin/cephtower`. Sans `-config`, il lit
`/opt/cephtower/config/config.yaml`, qui doit exister avant le démarrage.

## 5. Configuration

Consultez [config/config.yaml](../../config/config.yaml) pour toutes les options et valeurs par défaut.

| Section | Utilisation |
|---|---|
| `server` | adresse, port et répertoire d’exécution (défauts : `0.0.0.0:36900`, `/opt/cephtower`) |
| `log` | sortie, niveau, format, rotation et rétention |
| `runtime` | fichiers de configuration Ceph, keyrings et fichiers d’exécution |
| `database` | fichier SQLite ou connexion/TLS MySQL ; migrations automatiques au démarrage |
| `smtp` | service e-mail facultatif pour la réinitialisation des mots de passe |

Les identifiants Ceph ne figurent pas dans ce YAML : ils sont enregistrés dans la base via la gestion
des clusters après l’initialisation. Restreignez l’accès à la configuration, à la base et aux fichiers
d’exécution, et utilisez une validation TLS appropriée en production.

## 6. Commandes courantes

| Commande | Utilisation |
|---|---|
| `make check-env` | vérifier les versions de Go, Node.js et npm |
| `make run` | lancer ensemble backend et frontend de développement |
| `make run-backend` | construire et lancer le backend ; `CONFIG` choisit le fichier |
| `make run-frontend` | lancer Vite sur le port `36901` |
| `make build` | construire le frontend et créer `bin/cephtower` avec l’UI intégrée |
| `make build-frontend` | vérifier les types, construire et synchroniser les ressources intégrées |
| `make test` | exécuter les tests backend et valider le build frontend |
| `make test-backend` | exécuter `go test ./...` |
| `make test-frontend` | vérifier les types et construire le frontend |
| `ruby tools/generate_ceph_dashboard_client.rb` | régénérer le client Dashboard depuis les références locales |

Utilisez `CONFIG=/path/to/config.yaml` pour la configuration backend ou `FRONTEND_PORT=port`
pour changer le port frontend de `make run`.

## 7. API et documentation

Le préfixe API est `/api/v1`. Endpoints de base sans authentification :

| Méthode | Chemin | Utilisation |
|---|---|---|
| `GET` | `/api/v1/healthz` | vitalité du processus |
| `GET` | `/api/v1/readyz` | initialisation terminée |
| `GET` | `/api/v1/setup/status` | état du premier démarrage |
| `POST` | `/api/v1/auth/login` | connexion et obtention du Token |

Hors initialisation, connexion et réinitialisation, les requêtes exigent
`Authorization: Bearer <token>`. Les routes sont dans `backend/internal/api/v1/router/`.
Voir [docs/ceph/apis/index.md](../ceph/apis/index.md) pour la portée et la compatibilité Ceph.

## 8. Développement et contribution

- Exécutez `make test-backend` pour le backend et `make test-frontend` pour le frontend.
- Ne validez pas les données, bases, journaux ou clés de cluster locales de `app/`.
- Suivez [docs/commit-convention.md](../commit-convention.md) pour les commits.
- Issues et Pull Requests sont bienvenus ; distinguez clairement fonctions validées et maquettes.

## 9. Licence

CephTower est distribué sous [licence MIT](../../LICENSE).

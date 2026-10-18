# Orla Numis

> A self-hosted Progressive Web App for cataloguing a personal coin and banknote collection.

![Go](https://img.shields.io/badge/Go-1.26.6-00ADD8?logo=go)
![Vue](https://img.shields.io/badge/Vue-3-42b883?logo=vuedotjs)
![TypeScript](https://img.shields.io/badge/TypeScript-5-3178c6?logo=typescript)
![Python](https://img.shields.io/badge/Python-3.12-3776ab?logo=python)
![SQLite](https://img.shields.io/badge/SQLite-self--hosted-003b57?logo=sqlite)
![License](https://img.shields.io/badge/license-MIT-blue)

## What is this?

**Orla Numis** catalogues a personal collection of coins and banknotes on your own computer or server: records with photos and catalog references, Numista integration, sets and checklists, statistics, a wish list, backups, and optional AI tools. The interface is in English and European Portuguese.

Built for one collector, plus invited people if you choose. If you expose it outside a trusted home network, follow the hardening checklist in [`docs/deployment.md`](docs/deployment.md): TLS, reverse-proxy security headers, trusted proxy IP handling, closed registration, a private agent network, backups and monitoring.

Orla Numis is based on [Aurearia](https://github.com/briandenicola/Aurearia) by Brian DeNicola, under the MIT license; its footer says so too. Some internal names keep the first name, so that existing data stays where it is: the Docker Compose projects (`aurearia-dev`, `aurearia-qa`) and their volumes, the backup folder (`~/Aurearia-backups`, `AUREARIA_BACKUP_DIR`), and the backups made before the rename (`aurearia-….zip`, still listed and restored).

---

## ✨ Feature Highlights

### 🏛️ Collection
- **Coin gallery** with search and filters by face value, country (extinct countries included), coin type, presentation, rarity and set; grouping by country, year, rarity or presentation; grid or list.
- **Banknotes** with their own form and gallery, and a **collection overview** with the totals, the new releases and the duplicates to trade.
- **Coin details**: numismatic fields, quantity, images, structured catalog references, Numista market value, sets, activity journal and timeline.
- **Official photos** from the ECB, INCM, Monnaie de Paris and Münze Österreich, for a coin added or from its page.

### ➕ Adding coins
- From the **Numista catalog**, manually, or with **AI Assist Mode** from photos.
- **Identify Coin** at a show: NGC certificate extraction and verification, Numista matches, save to the wish list or collection.
- **Quick Capture** drafts for photos and notes taken on the go.
- **Import** from CSV, JSON or a uCoin.net export.

### 🗂️ Sets
- Lists from the **commemorative catalogue** (each euro country's €2 commemoratives and collector coins), ready-made **challenges**, sets **by filters**, free sets, goals with a wish list, and AI proposals reviewed before they are created.

### 📊 Stats
- Summary cards, distributions, top coins, acquisition flow, a heat map, timeline, mint map, **collection health** with quick fixes, and a **Time Machine** that shows the collection on any past date.

### 🎯 Wish list
- Dealer pages turned into wishlist coins, availability checks, saved search alerts with a review queue, and **Mark as purchased**.

### 🤖 AI (optional)
- Photo analysis and grade estimates, **Deep Analysis** with Numista evidence and cited proposals, a **Coin Agent** chat for dealer and auction searches, shows and price trends, and the read-only **Coin Copilot** (beta). Anthropic Claude or a self-hosted Ollama.

### 💾 Backups
- **Export ZIP** and an insurance **PDF** of your collection, and a **full backup** of the whole database and every photo, made every day by itself, with import and restore.

### 📱 Installed app
- Installable PWA with offline reading, swipe gallery, pull to refresh, camera capture and voice dictation.

### 🔐 Security
- Username and password with JWT access and refresh tokens, account lockout, automatic IP blocking, rate limits, closed registration after the first user, and private media served only to their owner.

---

## Architecture

Three services, two containers:

```
Browser ──► Vue 3 SPA  ──► Go API (Gin, GORM, SQLite) ──► Python LangGraph Agent
            (PWA)          :8080                          :8081 (stateless, SSE)
```

| Layer    | Tech                                  | Path         |
| -------- | ------------------------------------- | ------------ |
| Backend  | Go 1.26.6 (Gin), GORM, pure-Go SQLite | `src/api/`   |
| Frontend | Vue 3, TypeScript, Vite, Pinia (PWA)  | `src/web/`   |
| Agent    | Python 3.12, FastAPI, LangGraph       | `src/agent/` |

The Go API serves the REST API (`/api/*`) and the compiled Vue SPA, and proxies AI requests (including SSE streams) to the Python agent. The agent is stateless: provider, keys, prompts and context come with each request. Details in **[`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)**.

## Quick Start

With [Docker](https://docs.docker.com/get-docker/) and [Task](https://taskfile.dev/):

```sh
task init                                                # .env with random secrets
docker compose -f docker-compose.dev.yml up -d --build   # build and start both containers
```

Open `http://localhost:8080` and register: the first user becomes the admin, who can configure Numista and the AI providers in **Admin**. A walkthrough is in **[`docs/getting-started.md`](docs/getting-started.md)**; running on a server is in **[`docs/deployment.md`](docs/deployment.md)**.

For development without Docker ([Go 1.26.6](https://go.dev/dl/), [Node.js 24](https://nodejs.org/), Python 3.12 with [uv](https://docs.astral.sh/uv/)):

```sh
task up          # API (:8080) + web (:5173)
task up-all      # API + web + Python agent (:8081)
```

### Running tests

```sh
task check:go      # Go build, vet, gofmt and tests
task check:web     # lint, type-check, tests and build of the web app
task check:agent   # Ruff and pytest of the agent
task check         # all of the above plus the OpenAPI snapshot
```

The same checks run on every pull request (`.github/workflows/ci.yml`). See **[`docs/testing.md`](docs/testing.md)**.

## 📚 Documentation

| Purpose | Where to Look |
|---|---|
| **Features overview** | [`docs/features.md`](docs/features.md), and each area in [`docs/features/`](docs/features/INDEX.md) |
| **Getting started** | [`docs/getting-started.md`](docs/getting-started.md) |
| **Installation and deployment** | [`docs/deployment.md`](docs/deployment.md) |
| **Installed app (PWA)** | [`docs/pwa-guide.md`](docs/pwa-guide.md) |
| **Quick Capture** | [`docs/quick-capture.md`](docs/quick-capture.md) |
| **Architecture** | [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) |
| **API reference** | [`docs/api-reference.md`](docs/api-reference.md) and [`docs/openapi.json`](docs/openapi.json) |
| **Authentication** | [`docs/authentication.md`](docs/authentication.md) |
| **Testing** | [`docs/testing.md`](docs/testing.md) |
| **Design system** | [`docs/design-system.md`](docs/design-system.md) |
| **Security** | [`docs/security-principles.md`](docs/security-principles.md), [`docs/threat-model.md`](docs/threat-model.md), [`docs/incident-response.md`](docs/incident-response.md), [`SECURITY.md`](SECURITY.md) |
| **References** | [`docs/references.md`](docs/references.md) |
| **Cross-service contracts** | [`contracts/`](contracts/) |
| **Contributing** | [`CONTRIBUTING.md`](CONTRIBUTING.md) |

## License

[MIT](LICENSE).

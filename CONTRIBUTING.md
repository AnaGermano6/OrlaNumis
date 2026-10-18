# Contributing

Guidelines for changing Orla Numis.

## Getting Started

```bash
task init        # creates .env with random secrets
task up          # API and web dev servers (task up-all adds the Python agent)
```

| Service | URL | Source |
|---|---|---|
| API | `http://localhost:8080/api/*` | `src/api/` |
| Web | `http://localhost:5173` | `src/web/` |
| Agent | `http://localhost:8081` (internal) | `src/agent/` |
| Swagger | `http://localhost:8080/swagger/index.html` | generated |

Or run everything in Docker: `docker compose -f docker-compose.dev.yml up -d --build` (see [docs/deployment.md](docs/deployment.md)).

## Branches and Pull Requests

- `main` holds the released code.
- Work on a feature branch and open a pull request to `main`. The Quality Gate ([`.github/workflows/ci.yml`](.github/workflows/ci.yml)) runs the Go, web and agent checks on every pull request.
- Keep a pull request to one coherent change, and describe what changed and how it was verified.

## Code Architecture

Read [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) before changing the API. The key rules:

1. **Only the composition root (`main.go` + `deps.go`) imports the `database` package** — everything else uses dependency injection
2. **Handlers are thin** — parse the request, call a service or repository, return the response
3. **Services hold business logic** — HTTP-agnostic, orchestrating repositories
4. **Repositories own all database access** — every GORM query lives there
5. **Multi-step writes use transactions**
6. **Never leak internal errors to clients**

`src/api/architecture_test.go` enforces them.

## Adding a Feature

### New API endpoint

1. **Model** — add or update structs in `src/api/models/` (and the migration in `src/api/database/database.go` when a column or table goes away)
2. **Repository** — add the database methods in `src/api/repository/`
3. **Service** — business logic in `src/api/services/`
4. **Handler** — a thin HTTP handler in `src/api/handlers/`
5. **Wire** — register it in `src/api/deps.go` and the `routes_*.go` files
6. **Swagger** — add the annotations, run `task openapi` and commit the regenerated files
7. **Test** — `task check:go`

### New Vue page or component

1. Add it under `src/web/src/pages/` or `src/web/src/components/`
2. Add the route in `src/web/src/router/index.ts` for a page
3. Add the texts in both `src/web/src/i18n/locales/en` and `pt-PT`; a test fails on keys nothing uses
4. Run `task check:web`

## Code Style

### Go (API)
- Standard Go conventions (`gofmt`, `go vet`)
- Constructor injection for dependencies
- Typed errors from services (sentinel errors such as `ErrNotFound`)
- Log internal errors server-side, return generic messages to clients

### TypeScript (Web)
- Optional chaining (`?.`) and nullish coalescing (`??`) on array access
- The production build type-checks strictly — verify with `npm run type-check`
- Follow [docs/design-system.md](docs/design-system.md)
- No emojis in the UI, prompts or AI responses

## Build and Test

From the repository root:

```bash
task check:go       # build, vet, gofmt, tests
task check:web      # lint, type-check, tests, build
task check:agent    # Ruff and pytest in the locked environment
task check:openapi  # regenerates the OpenAPI artifacts and fails on drift (writes files)
task check          # all of the above
```

Dependencies are installed separately with `task setup:go`, `task setup:web`, `task setup:agent` and `task setup:openapi`. See [docs/testing.md](docs/testing.md).

## Commit Messages

Use conventional commits:

```
feat: add coin grading service
fix: handle a missing purchase price in stats
refactor: extract set rules to the service layer
```
